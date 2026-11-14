package service

import (
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/apperr"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/repository"
)

// CommunityService 在线题库社区：发布、fork、点赞、评论、市场列表。
type CommunityService struct {
	repo *repository.DB
}

func NewCommunityService(repo *repository.DB) *CommunityService { return &CommunityService{repo: repo} }

func (s *CommunityService) Marketplace(search, sort string) ([]repository.DeckWithAuthor, error) {
	return s.repo.ListPublicDecks(search, sort, 50)
}

// Publish 发布到在线市场（设为 public；首个 commit 若不存在则在此创建）。
func (s *CommunityService) Publish(userID, deckID uuid.UUID) (*model.Deck, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	if deck.Archived {
		return nil, apperr.New(400, "该题库已归档为市场公开版，不能再发布")
	}
	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, apperr.New(400, "空题库不能发布，请先添加词条")
	}
	deck.Visibility = "public"
	if err := s.repo.UpdateDeck(deck); err != nil {
		return nil, err
	}
	return deck, nil
}

// Fork 将线上题库复制为本地副本：词条沿用 LogicalID，commit 历史原样复制（hash 不变，
// 因此 fork 与源题库可互相求公共祖先、执行三路合并）。
//
// 幂等（GitHub 语义）：同一用户对同一来源只保留一个活跃 fork。重复调用返回既有副本、
// created=false，且不再增加来源题库的 ForksCount。已彻底删除或已归档的副本不算「活跃」。
func (s *CommunityService) Fork(userID, deckID uuid.UUID) (*model.Deck, bool, error) {
	origin, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, false, apperr.ErrNotFound
	}
	if origin.Visibility != "public" && origin.OwnerID != userID {
		return nil, false, apperr.ErrForbidden
	}

	existing, err := s.repo.FindActiveFork(userID, deckID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	fork := &model.Deck{
		ID:          uuid.New(),
		OwnerID:     userID,
		Name:        origin.Name + " (fork)",
		Description: origin.Description,
		Visibility:  "private",
		ForkedFrom:  &origin.ID,
	}
	// 复制词条（保留 LogicalID）
	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, false, err
	}
	for _, e := range entries {
		clone := e
		clone.ID = uuid.New()
		clone.DeckID = fork.ID
		if err := s.repo.CreateEntry(&clone); err != nil {
			return nil, false, err
		}
	}
	// 复制 commit 历史（保留 hash / seq / parents / createdAt，保证重放顺序与 LCA 匹配）
	commits, err := s.repo.ListCommits(deckID)
	if err != nil {
		return nil, false, err
	}
	for _, c := range commits {
		cc := c
		cc.ID = uuid.New()
		cc.DeckID = fork.ID
		if err := s.repo.CreateCommit(&cc); err != nil {
			return nil, false, err
		}
	}
	head := ""
	if origin.HeadCommitID != nil {
		head = *origin.HeadCommitID
	}
	if head != "" {
		fork.HeadCommitID = &head
	}
	if err := s.repo.CreateDeck(fork); err != nil {
		return nil, false, err
	}
	origin.ForksCount++
	if err := s.repo.UpdateDeck(origin); err != nil {
		return nil, false, err
	}
	return fork, true, nil
}

func (s *CommunityService) ToggleLike(userID, deckID uuid.UUID) (bool, int, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return false, 0, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return false, 0, apperr.ErrForbidden
	}
	existing, err := s.repo.GetLike(deckID, userID)
	if err != nil {
		return false, 0, err
	}
	if existing != nil {
		if err := s.repo.DeleteLike(existing.ID); err != nil {
			return false, 0, err
		}
		deck.LikesCount--
		if err := s.repo.UpdateDeck(deck); err != nil {
			return false, 0, err
		}
		return false, deck.LikesCount, nil
	}
	if err := s.repo.CreateLike(&model.Like{ID: uuid.New(), DeckID: deckID, UserID: userID}); err != nil {
		return false, 0, err
	}
	deck.LikesCount++
	if err := s.repo.UpdateDeck(deck); err != nil {
		return false, 0, err
	}
	return true, deck.LikesCount, nil
}

func (s *CommunityService) ListComments(userID, deckID uuid.UUID) ([]repository.CommentWithAuthor, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	return s.repo.ListComments(deckID)
}

func (s *CommunityService) AddComment(userID, deckID uuid.UUID, content string) (*repository.CommentWithAuthor, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.Visibility != "public" && deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	if content == "" {
		return nil, apperr.New(400, "评论内容不能为空")
	}
	cm := &model.Comment{ID: uuid.New(), DeckID: deckID, UserID: userID, Content: content}
	if err := s.repo.CreateComment(cm); err != nil {
		return nil, err
	}
	author, _ := s.repo.GetUserByID(userID)
	name := userID.String()
	if author != nil {
		name = author.DisplayName
		if name == "" {
			name = author.Username
		}
	}
	return &repository.CommentWithAuthor{Comment: *cm, AuthorName: name}, nil
}
