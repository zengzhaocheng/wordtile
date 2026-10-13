package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/wordtile/backend/internal/model"
)

// DB 封装 GORM 数据访问层（Repository）。
type DB struct {
	*gorm.DB
}

func New(db *gorm.DB) *DB { return &DB{db} }

/* ---------- Users ---------- */

func (d *DB) CreateUser(u *model.User) error { return d.Create(u).Error }

func (d *DB) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	if err := d.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	if err := d.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) GetUserByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := d.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) UpdateUser(u *model.User) error { return d.Save(u).Error }

/* ---------- Decks ---------- */

func (d *DB) CreateDeck(deck *model.Deck) error { return d.Create(deck).Error }

func (d *DB) GetDeck(id uuid.UUID) (*model.Deck, error) {
	var deck model.Deck
	if err := d.First(&deck, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &deck, nil
}

func (d *DB) UpdateDeck(deck *model.Deck) error { return d.Save(deck).Error }

func (d *DB) DeleteDeckCascade(deckID uuid.UUID) error {
	return d.Transaction(func(tx *gorm.DB) error {
		for _, m := range []any{&model.Entry{}, &model.Commit{}, &model.Conflict{}, &model.Like{}, &model.Comment{}} {
			if err := tx.Where("deck_id = ?", deckID).Delete(m).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.Deck{}, "id = ?", deckID).Error
	})
}

func (d *DB) ListDecksByOwner(ownerID uuid.UUID) ([]model.Deck, error) {
	var decks []model.Deck
	err := d.Where("owner_id = ?", ownerID).Order("updated_at DESC").Find(&decks).Error
	return decks, err
}

// DeckWithAuthor 市场列表项（带作者昵称）。
type DeckWithAuthor struct {
	model.Deck
	OwnerName string `json:"owner_name"`
}

func (d *DB) ListPublicDecks(search, sort string, limit int) ([]DeckWithAuthor, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := d.Table("decks").
		Select("decks.*, COALESCE(users.display_name, users.username) AS owner_name").
		Joins("LEFT JOIN users ON users.id = decks.owner_id").
		Where("decks.visibility = ?", "public")
	if search != "" {
		q = q.Where("decks.name ILIKE ?", "%"+search+"%")
	}
	switch sort {
	case "newest":
		q = q.Order("decks.created_at DESC")
	case "forks":
		q = q.Order("decks.forks_count DESC, decks.created_at DESC")
	default:
		q = q.Order("decks.likes_count DESC, decks.created_at DESC")
	}
	var out []DeckWithAuthor
	err := q.Limit(limit).Scan(&out).Error
	return out, err
}

/* ---------- Entries ---------- */

func (d *DB) ListEntries(deckID uuid.UUID) ([]model.Entry, error) {
	var entries []model.Entry
	err := d.Where("deck_id = ?", deckID).Order("created_at ASC, term ASC").Find(&entries).Error
	return entries, err
}

func (d *DB) GetEntry(id uuid.UUID) (*model.Entry, error) {
	var e model.Entry
	if err := d.First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (d *DB) CreateEntry(e *model.Entry) error { return d.Create(e).Error }

func (d *DB) UpdateEntry(e *model.Entry) error { return d.Save(e).Error }

func (d *DB) DeleteEntry(id uuid.UUID) error { return d.Delete(&model.Entry{}, "id = ?", id).Error }

// UpsertEntryByLogicalID 按逻辑 ID 写入词条（合并应用使用）。
func (d *DB) UpsertEntryByLogicalID(deckID, logicalID uuid.UUID, term, definition string, tags []string) error {
	var e model.Entry
	err := d.Where("deck_id = ? AND logical_id = ?", deckID, logicalID).First(&e).Error
	if err == gorm.ErrRecordNotFound {
		return d.Create(&model.Entry{
			ID: uuid.New(), DeckID: deckID, LogicalID: logicalID,
			Term: term, Definition: definition, Tags: tags,
		}).Error
	}
	if err != nil {
		return err
	}
	e.Term, e.Definition, e.Tags = term, definition, tags
	return d.Save(&e).Error
}

func (d *DB) DeleteEntryByLogicalID(deckID, logicalID uuid.UUID) error {
	return d.Where("deck_id = ? AND logical_id = ?", deckID, logicalID).Delete(&model.Entry{}).Error
}

/* ---------- Commits ---------- */

func (d *DB) CreateCommit(c *model.Commit) error { return d.Create(c).Error }

// ListCommits 按题库内单调序号重放顺序返回。
func (d *DB) ListCommits(deckID uuid.UUID) ([]model.Commit, error) {
	var commits []model.Commit
	err := d.Where("deck_id = ?", deckID).Order("seq ASC").Find(&commits).Error
	return commits, err
}

func (d *DB) MaxCommitSeq(deckID uuid.UUID) (int64, error) {
	var max int64
	err := d.Model(&model.Commit{}).Where("deck_id = ?", deckID).
		Select("COALESCE(MAX(seq), 0)").Scan(&max).Error
	return max, err
}

/* ---------- Conflicts ---------- */

func (d *DB) CreateConflict(cf *model.Conflict) error { return d.Create(cf).Error }

func (d *DB) UpdateConflict(cf *model.Conflict) error { return d.Save(cf).Error }

func (d *DB) GetConflict(id uuid.UUID) (*model.Conflict, error) {
	var cf model.Conflict
	if err := d.First(&cf, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cf, nil
}

func (d *DB) ListConflicts(deckID uuid.UUID, status string) ([]model.Conflict, error) {
	var list []model.Conflict
	err := d.Where("deck_id = ? AND status = ?", deckID, status).Order("created_at ASC").Find(&list).Error
	return list, err
}

/* ---------- Likes / Comments ---------- */

func (d *DB) GetLike(deckID, userID uuid.UUID) (*model.Like, error) {
	var l model.Like
	err := d.Where("deck_id = ? AND user_id = ?", deckID, userID).First(&l).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (d *DB) CreateLike(l *model.Like) error { return d.Create(l).Error }

func (d *DB) DeleteLike(id uuid.UUID) error { return d.Delete(&model.Like{}, "id = ?", id).Error }

// CommentWithAuthor 评论（带作者昵称）。
type CommentWithAuthor struct {
	model.Comment
	AuthorName string `json:"author_name"`
}

func (d *DB) ListComments(deckID uuid.UUID) ([]CommentWithAuthor, error) {
	var out []CommentWithAuthor
	err := d.Table("comments").
		Select("comments.*, COALESCE(users.display_name, users.username) AS author_name").
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Where("comments.deck_id = ?", deckID).
		Order("comments.created_at DESC").
		Limit(200).Scan(&out).Error
	return out, err
}

func (d *DB) CreateComment(cm *model.Comment) error { return d.Create(cm).Error }

/* ---------- GameSessions ---------- */

func (d *DB) CreateSession(s *model.GameSession) error { return d.Create(s).Error }

func (d *DB) GetSession(id uuid.UUID) (*model.GameSession, error) {
	var s model.GameSession
	if err := d.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) SaveSession(s *model.GameSession) error { return d.Save(s).Error }

func (d *DB) ListFinishedSessions(userID uuid.UUID) ([]model.GameSession, error) {
	var list []model.GameSession
	err := d.Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Order("completed_at ASC").Find(&list).Error
	return list, err
}

/* ---------- Stats 聚合 ---------- */

// ContributionRow 按日聚合行。
type ContributionRow struct {
	Date      string `json:"date"`
	Sessions  int64  `json:"sessions"`
	Duration  int64  `json:"duration_ms"`
	Correct   int64  `json:"correct"`
	Wrong     int64  `json:"wrong"`
}

func (d *DB) ContributionByDay(userID uuid.UUID) ([]ContributionRow, error) {
	var rows []ContributionRow
	err := d.Raw(`
		SELECT to_char(completed_at, 'YYYY-MM-DD') AS date,
		       COUNT(*) AS sessions,
		       COALESCE(SUM(duration_ms), 0) AS duration_ms,
		       COALESCE(SUM(correct_count), 0) AS correct,
		       COALESCE(SUM(wrong_count), 0) AS wrong
		FROM game_sessions
		WHERE user_id = ? AND completed_at IS NOT NULL
		  AND completed_at >= now() - interval '370 days'
		GROUP BY 1 ORDER BY 1`, userID).Scan(&rows).Error
	return rows, err
}

// CampaignRow 闯关模式按关卡聚合行。
type CampaignRow struct {
	Level     int   `json:"level"`
	Plays     int64 `json:"plays"`
	BestMS    int64 `json:"best_ms"`
	BestWrong int64 `json:"best_wrong"`
}

func (d *DB) CampaignByLevel(userID uuid.UUID) ([]CampaignRow, error) {
	var rows []CampaignRow
	err := d.Raw(`
		SELECT level,
		       COUNT(*) AS plays,
		       MIN(duration_ms) AS best_ms,
		       MIN(wrong_count) AS best_wrong
		FROM game_sessions
		WHERE user_id = ? AND mode = 'campaign' AND completed_at IS NOT NULL AND level > 0
		GROUP BY level ORDER BY level`, userID).Scan(&rows).Error
	return rows, err
}
