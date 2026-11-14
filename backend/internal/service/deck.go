package service

import (
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/apperr"
	"github.com/wordtile/backend/internal/deckio"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/repository"
)

// DeckService 本地题库 CRUD 与词条管理。
type DeckService struct {
	repo *repository.DB
}

func NewDeckService(repo *repository.DB) *DeckService { return &DeckService{repo: repo} }

func (s *DeckService) ownedDeck(userID, deckID uuid.UUID) (*model.Deck, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	return deck, nil
}

// DeckDetail 题库详情（含词条、fork 来源信息、未解决冲突数）。
type DeckDetail struct {
	Deck          model.Deck    `json:"deck"`
	Entries       []model.Entry `json:"entries"`
	OriginHead    string        `json:"origin_head,omitempty"`
	OpenConflicts int           `json:"open_conflicts"`
}

func (s *DeckService) GetDeckDetail(userID, deckID uuid.UUID) (*DeckDetail, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, err
	}
	detail := &DeckDetail{Deck: *deck, Entries: entries}
	if deck.ForkedFrom != nil {
		if origin, err := s.repo.GetDeck(*deck.ForkedFrom); err == nil && origin.HeadCommitID != nil {
			detail.OriginHead = *origin.HeadCommitID
		}
	}
	conflicts, _ := s.repo.ListConflicts(deckID, "open")
	detail.OpenConflicts = len(conflicts)
	return detail, nil
}

func (s *DeckService) ListMyDecks(userID uuid.UUID) ([]model.Deck, error) {
	return s.repo.ListDecksByOwner(userID)
}

func (s *DeckService) CreateDeck(userID uuid.UUID, name, description string) (*model.Deck, error) {
	if name == "" {
		return nil, apperr.New(400, "题库名称不能为空")
	}
	deck := &model.Deck{
		ID: uuid.New(), OwnerID: userID, Name: name,
		Description: description, Visibility: "private",
	}
	if err := s.repo.CreateDeck(deck); err != nil {
		return nil, err
	}
	return deck, nil
}

func (s *DeckService) UpdateDeck(userID, deckID uuid.UUID, name, description, visibility *string) (*model.Deck, error) {
	deck, err := s.ownedDeck(userID, deckID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if *name == "" {
			return nil, apperr.New(400, "题库名称不能为空")
		}
		deck.Name = *name
	}
	if description != nil {
		deck.Description = *description
	}
	if visibility != nil {
		if *visibility != "private" && *visibility != "public" {
			return nil, apperr.New(400, "visibility 仅支持 private / public")
		}
		deck.Visibility = *visibility
	}
	if err := s.repo.UpdateDeck(deck); err != nil {
		return nil, err
	}
	return deck, nil
}

func (s *DeckService) DeleteDeck(userID, deckID uuid.UUID) error {
	if _, err := s.ownedDeck(userID, deckID); err != nil {
		return err
	}
	return s.repo.DeleteDeckCascade(deckID)
}

func (s *DeckService) CreateEntry(userID, deckID uuid.UUID, term, definition string, tags []string) (*model.Entry, error) {
	if _, err := s.ownedDeck(userID, deckID); err != nil {
		return nil, err
	}
	if term == "" || definition == "" {
		return nil, apperr.New(400, "词条与释义均不能为空")
	}
	e := &model.Entry{
		ID: uuid.New(), DeckID: deckID, LogicalID: uuid.New(),
		Term: term, Definition: definition, Tags: tags,
	}
	if err := s.repo.CreateEntry(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *DeckService) UpdateEntry(userID, entryID uuid.UUID, term, definition *string, tags []string, tagsSet bool) (*model.Entry, error) {
	e, err := s.repo.GetEntry(entryID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if _, err := s.ownedDeck(userID, e.DeckID); err != nil {
		return nil, err
	}
	if term != nil {
		if *term == "" {
			return nil, apperr.New(400, "词条不能为空")
		}
		e.Term = *term
	}
	if definition != nil {
		if *definition == "" {
			return nil, apperr.New(400, "释义不能为空")
		}
		e.Definition = *definition
	}
	if tagsSet {
		e.Tags = tags
	}
	if err := s.repo.UpdateEntry(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *DeckService) DeleteEntry(userID, entryID uuid.UUID) error {
	e, err := s.repo.GetEntry(entryID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if _, err := s.ownedDeck(userID, e.DeckID); err != nil {
		return err
	}
	return s.repo.DeleteEntry(entryID)
}

func (s *DeckService) ListCommits(userID, deckID uuid.UUID) ([]model.Commit, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	return s.repo.ListCommits(deckID)
}

/* ---------- JSON 导入 / 导出 ---------- */

// ExportDeck 把题库导出为可交换的 JSON 文档。公开题库任何人可导出，私有题库仅所有者。
func (s *DeckService) ExportDeck(userID, deckID uuid.UUID) (*deckio.DeckFile, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, err
	}
	items := make([]deckio.EntryItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, deckio.EntryItem{Term: e.Term, Definition: e.Definition, Tags: e.Tags})
	}
	f := deckio.NewExport(deck.Name, deck.Description, items)
	return &f, nil
}

// ImportNewDeck 从 JSON 文档创建一个全新的私有题库（含全部词条，单事务）。
func (s *DeckService) ImportNewDeck(userID uuid.UUID, f *deckio.DeckFile) (*model.Deck, error) {
	name := f.Name
	if name == "" {
		name = "导入题库"
	}
	deck := &model.Deck{
		ID: uuid.New(), OwnerID: userID, Name: name,
		Description: f.Description, Visibility: "private",
	}
	if err := s.repo.CreateDeckWithEntries(deck, toModelEntries(deck.ID, f.Entries)); err != nil {
		return nil, err
	}
	return deck, nil
}

// ImportIntoDeck 把词条导入已有题库。replace=true 会清空原词条；否则追加。
// 注意：JSON 文件里的 name / description 在导入已有题库时被忽略。
func (s *DeckService) ImportIntoDeck(userID, deckID uuid.UUID, f *deckio.DeckFile, replace bool) (int, error) {
	if _, err := s.ownedDeck(userID, deckID); err != nil {
		return 0, err
	}
	entries := toModelEntries(deckID, f.Entries)
	if err := s.repo.ImportEntries(deckID, entries, replace); err != nil {
		return 0, err
	}
	return len(entries), nil
}

// toModelEntries 为每条导入词条生成新的行 ID 与 LogicalID，保证 Git-like 合并键唯一。
func toModelEntries(deckID uuid.UUID, items []deckio.EntryItem) []model.Entry {
	out := make([]model.Entry, 0, len(items))
	for _, it := range items {
		out = append(out, model.Entry{
			ID: uuid.New(), DeckID: deckID, LogicalID: uuid.New(),
			Term: it.Term, Definition: it.Definition, Tags: it.Tags,
		})
	}
	return out
}
