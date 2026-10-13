package model

import (
	"time"

	"github.com/google/uuid"
)

// User 用户。
type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username        string    `gorm:"size:64;uniqueIndex" json:"username"`
	Email           string    `gorm:"size:255;uniqueIndex" json:"email"`
	PasswordHash    string    `json:"-"`
	DisplayName     string    `gorm:"size:128" json:"display_name"`
	AvatarURL       string    `gorm:"size:512" json:"avatar_url"`
	ThemePreference string    `gorm:"size:16;default:system" json:"theme_preference"` // light / dark / system
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Deck 题库（≈ Git 仓库）。
type Deck struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID            uuid.UUID `gorm:"type:uuid;index" json:"owner_id"`
	Name               string    `gorm:"size:200" json:"name"`
	Description        string    `gorm:"type:text" json:"description"`
	Visibility         string    `gorm:"size:16;default:private;index" json:"visibility"` // private / public
	HeadCommitID       *string   `gorm:"size:36" json:"head_commit_id"`                   // 当前主干 commit hash
	ForkedFrom         *uuid.UUID `gorm:"type:uuid" json:"forked_from"`                    // fork 来源题库
	LikesCount         int       `gorm:"default:0" json:"likes_count"`
	ForksCount         int       `gorm:"default:0" json:"forks_count"`
	PendingMergeSource *string   `gorm:"size:36" json:"pending_merge_source"` // 待解决的合并来源 commit
	PendingMergeBase   *string   `gorm:"size:36" json:"pending_merge_base"`   // 待解决合并的 base commit
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Entry 词条（≈ Git 文件）。
// ID 是行级主键；LogicalID 是跨题库稳定的逻辑身份，fork 时保持不变，
// commit diff 与三路合并均以 LogicalID 为准。
type Entry struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DeckID     uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_entries_deck_logical" json:"deck_id"`
	LogicalID  uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_entries_deck_logical" json:"logical_id"`
	Term       string    `gorm:"size:256" json:"term"`
	Definition string    `gorm:"size:512" json:"definition"`
	Tags       []string  `gorm:"serializer:json;type:jsonb" json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// EntryChange 单条词条的变更快照（commit / merge 中使用）。
type EntryChange struct {
	ID         uuid.UUID `json:"id"`
	Term       string    `json:"term"`
	Definition string    `json:"definition"`
	Tags       []string  `json:"tags"`
}

// EntryChanges 一组变更 diff。
type EntryChanges struct {
	Added   []EntryChange `json:"added,omitempty"`
	Updated []EntryChange `json:"updated,omitempty"`
	Removed []uuid.UUID   `json:"removed,omitempty"`
}

func (c EntryChanges) Empty() bool {
	return len(c.Added) == 0 && len(c.Updated) == 0 && len(c.Removed) == 0
}

// Commit 提交（≈ git commit）。Hash 在 fork 时原样复制，用于跨题库求公共祖先。
type Commit struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	DeckID       uuid.UUID    `gorm:"type:uuid;uniqueIndex:idx_commits_deck_hash" json:"deck_id"`
	Hash         string       `gorm:"size:36;uniqueIndex:idx_commits_deck_hash" json:"hash"`
	Seq          int64        `json:"seq"` // 题库内单调递增，保证重放顺序稳定
	ParentHashes []string     `gorm:"serializer:json;type:jsonb" json:"parent_hashes"`
	AuthorID     uuid.UUID    `gorm:"type:uuid" json:"author_id"`
	Changes      EntryChanges `gorm:"serializer:json;type:jsonb" json:"changes"`
	Message      string       `gorm:"size:500" json:"message"`
	CreatedAt    time.Time    `json:"created_at"`
}

// EntrySnapshot 合并冲突时某一侧的词条快照。
type EntrySnapshot struct {
	Exists     bool     `json:"exists"`
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Tags       []string `json:"tags"`
}

// Conflict 合并冲突记录。
type Conflict struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	DeckID            uuid.UUID     `gorm:"type:uuid;index" json:"deck_id"`
	EntryID           uuid.UUID     `gorm:"type:uuid" json:"entry_id"` // LogicalID
	Base              EntrySnapshot `gorm:"serializer:json;type:jsonb" json:"base"`
	Ours              EntrySnapshot `gorm:"serializer:json;type:jsonb" json:"ours"`
	Theirs            EntrySnapshot `gorm:"serializer:json;type:jsonb" json:"theirs"`
	Status            string        `gorm:"size:16;default:open;index" json:"status"` // open / resolved
	Resolution        string        `gorm:"size:16" json:"resolution"`                 // ours / theirs / custom
	ResolvedSnap      EntrySnapshot `gorm:"serializer:json;type:jsonb" json:"resolved_snap"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// Comment 题库评论。
type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DeckID    uuid.UUID `gorm:"type:uuid;index" json:"deck_id"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Like 点赞（题库维度，deck_id+user_id 唯一）。
type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DeckID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_likes_deck_user" json:"deck_id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_likes_deck_user" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Question 一道题目（服务端持有答案映射，客户端只拿到 token）。
type Question struct {
	EntryID         uuid.UUID `json:"entry_id"`
	TermToken       string    `json:"term_token"`
	DefinitionToken string    `json:"definition_token"`
	Term            string    `json:"term"`
	Definition      string    `json:"definition"`
	Matched         bool      `json:"matched"`
}

// GameSession 一局练习记录。
type GameSession struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index" json:"user_id"`
	DeckID       uuid.UUID  `gorm:"type:uuid;index" json:"deck_id"`
	Mode         string     `gorm:"size:16" json:"mode"` // custom / campaign
	Level        int        `json:"level"`                // 闯关模式关卡号，custom 为 0
	BlockCount   int        `json:"block_count"`
	DurationMS   int64      `json:"duration_ms"`
	CorrectCount int        `json:"correct_count"`
	WrongCount   int        `json:"wrong_count"`
	CompletedAt  *time.Time `json:"completed_at"`
	Questions    []Question `gorm:"serializer:json;type:jsonb" json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
}
