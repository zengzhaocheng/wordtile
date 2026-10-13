package service

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/apperr"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/repository"
)

// GameService 出题、匹配判定、结算。
type GameService struct {
	repo *repository.DB
}

func NewGameService(repo *repository.DB) *GameService { return &GameService{repo: repo} }

// Tile 发给客户端的块：token 代替真实答案映射，防作弊。
type Tile struct {
	Token string `json:"token"`
	Text  string `json:"text"`
}

// StartResponse 开局数据（词条块与释义块各自独立打乱）。
type StartResponse struct {
	SessionID   string `json:"session_id"`
	Mode        string `json:"mode"`
	Level       int    `json:"level"`
	DeckID      string `json:"deck_id"`
	DeckName    string `json:"deck_name"`
	BlockCount  int    `json:"block_count"`
	Terms       []Tile `json:"terms"`
	Definitions []Tile `json:"definitions"`
}

func shuffleTiles(t []Tile) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(t), func(i, j int) { t[i], t[j] = t[j], t[i] })
}

func (s *GameService) Start(userID, deckID uuid.UUID, mode string, level, blockCount int, tags []string) (*StartResponse, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	if mode != "custom" && mode != "campaign" {
		return nil, apperr.New(400, "mode 仅支持 custom / campaign")
	}

	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		tagSet := map[string]bool{}
		for _, t := range tags {
			tagSet[t] = true
		}
		filtered := entries[:0]
		for _, e := range entries {
			for _, t := range e.Tags {
				if tagSet[t] {
					filtered = append(filtered, e)
					break
				}
			}
		}
		entries = filtered
	}
	if len(entries) < 2 {
		return nil, apperr.New(400, "可用词条不足（至少需要 2 条）")
	}

	n := blockCount
	if n < 2 {
		n = 2
	}
	if n > 50 {
		n = 50
	}
	if n > len(entries) {
		n = len(entries)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	picked := append([]model.Entry{}, entries...)
	r.Shuffle(len(picked), func(i, j int) { picked[i], picked[j] = picked[j], picked[i] })
	picked = picked[:n]

	questions := make([]model.Question, n)
	terms := make([]Tile, n)
	defs := make([]Tile, n)
	for i, e := range picked {
		questions[i] = model.Question{
			EntryID:         e.LogicalID,
			TermToken:       uuid.NewString(),
			DefinitionToken: uuid.NewString(),
			Term:            e.Term,
			Definition:      e.Definition,
		}
		terms[i] = Tile{Token: questions[i].TermToken, Text: e.Term}
		defs[i] = Tile{Token: questions[i].DefinitionToken, Text: e.Definition}
	}

	session := &model.GameSession{
		ID: uuid.New(), UserID: userID, DeckID: deckID, Mode: mode,
		Level: level, BlockCount: n, Questions: questions,
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}
	shuffleTiles(terms)
	shuffleTiles(defs)

	return &StartResponse{
		SessionID: session.ID.String(), Mode: mode, Level: level,
		DeckID: deckID.String(), DeckName: deck.Name, BlockCount: n,
		Terms: terms, Definitions: defs,
	}, nil
}

func (s *GameService) getSession(userID, sessionID uuid.UUID) (*model.GameSession, error) {
	sess, err := s.repo.GetSession(sessionID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if sess.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return sess, nil
}

// SessionState 会话状态（刷新恢复用，不含答案映射）。
type SessionState struct {
	SessionID   string `json:"session_id"`
	Mode        string `json:"mode"`
	Level       int    `json:"level"`
	BlockCount  int    `json:"block_count"`
	DeckID      string `json:"deck_id"`
	Correct     int    `json:"correct_count"`
	Wrong       int    `json:"wrong_count"`
	Matched     int    `json:"matched_count"`
	Terms       []Tile `json:"terms"`
	Definitions []Tile `json:"definitions"`
}

func (s *GameService) GetState(userID, sessionID uuid.UUID) (*SessionState, error) {
	sess, err := s.getSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	state := &SessionState{
		SessionID: sess.ID.String(), Mode: sess.Mode, Level: sess.Level,
		BlockCount: sess.BlockCount, DeckID: sess.DeckID.String(),
		Correct: sess.CorrectCount, Wrong: sess.WrongCount,
		Terms: make([]Tile, 0, len(sess.Questions)),
		Definitions: make([]Tile, 0, len(sess.Questions)),
	}
	for _, q := range sess.Questions {
		state.Terms = append(state.Terms, Tile{Token: q.TermToken, Text: q.Term})
		state.Definitions = append(state.Definitions, Tile{Token: q.DefinitionToken, Text: q.Definition})
		if q.Matched {
			state.Matched++
		}
	}
	return state, nil
}

// MatchResult 单次匹配判定结果。
type MatchResult struct {
	Correct      bool `json:"correct"`
	CorrectCount int  `json:"correct_count"`
	WrongCount   int  `json:"wrong_count"`
	Finished     bool `json:"finished"`
}

func (s *GameService) Match(userID, sessionID uuid.UUID, termToken, definitionToken string) (*MatchResult, error) {
	sess, err := s.getSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.CompletedAt != nil {
		return nil, apperr.New(409, "该局已结束")
	}
	for i := range sess.Questions {
		q := &sess.Questions[i]
		if q.TermToken != termToken {
			continue
		}
		if q.Matched {
			return nil, apperr.New(409, "该词条已匹配")
		}
		correct := q.DefinitionToken == definitionToken
		if correct {
			q.Matched = true
			sess.CorrectCount++
		} else {
			sess.WrongCount++
		}
		if err := s.repo.SaveSession(sess); err != nil {
			return nil, err
		}
		matched := 0
		for _, qq := range sess.Questions {
			if qq.Matched {
				matched++
			}
		}
		return &MatchResult{
			Correct: correct, CorrectCount: sess.CorrectCount,
			WrongCount: sess.WrongCount, Finished: matched == len(sess.Questions),
		}, nil
	}
	return nil, apperr.New(400, "无效的词条 token")
}

// HintResponse 提示：随机返回一对尚未匹配的正确配对。
type HintResponse struct {
	TermToken       string `json:"term_token"`
	DefinitionToken string `json:"definition_token"`
}

func (s *GameService) Hint(userID, sessionID uuid.UUID) (*HintResponse, error) {
	sess, err := s.getSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	var candidates []*model.Question
	for i := range sess.Questions {
		if !sess.Questions[i].Matched {
			candidates = append(candidates, &sess.Questions[i])
		}
	}
	if len(candidates) == 0 {
		return nil, apperr.New(409, "本局已全部匹配")
	}
	q := candidates[rand.Intn(len(candidates))]
	return &HintResponse{TermToken: q.TermToken, DefinitionToken: q.DefinitionToken}, nil
}

// FinishSummary 结算数据。
type FinishSummary struct {
	SessionID   string  `json:"session_id"`
	DurationMS  int64   `json:"duration_ms"`
	CorrectCount int    `json:"correct_count"`
	WrongCount  int     `json:"wrong_count"`
	Accuracy    float64 `json:"accuracy"`
	Mode        string  `json:"mode"`
	Level       int     `json:"level"`
}

func (s *GameService) Finish(userID, sessionID uuid.UUID, durationMS int64) (*FinishSummary, error) {
	sess, err := s.getSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.CompletedAt == nil {
		if durationMS < 0 {
			durationMS = 0
		}
		sess.DurationMS = durationMS
		now := time.Now()
		sess.CompletedAt = &now
		if err := s.repo.SaveSession(sess); err != nil {
			return nil, err
		}
	}
	total := sess.CorrectCount + sess.WrongCount
	var acc float64
	if total > 0 {
		acc = float64(sess.CorrectCount) / float64(total)
	}
	return &FinishSummary{
		SessionID: sess.ID.String(), DurationMS: sess.DurationMS,
		CorrectCount: sess.CorrectCount, WrongCount: sess.WrongCount,
		Accuracy: acc, Mode: sess.Mode, Level: sess.Level,
	}, nil
}
