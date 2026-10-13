-- WordTile 初始结构（与 GORM AutoMigrate 生成的结构对齐；生产环境使用 golang-migrate）
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    username varchar(64) NOT NULL UNIQUE,
    email varchar(255) NOT NULL UNIQUE,
    password_hash text NOT NULL,
    display_name varchar(128),
    avatar_url varchar(512),
    theme_preference varchar(16) DEFAULT 'system',
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS decks (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL,
    name varchar(200) NOT NULL,
    description text,
    visibility varchar(16) DEFAULT 'private',
    head_commit_id varchar(36),
    forked_from uuid,
    likes_count integer DEFAULT 0,
    forks_count integer DEFAULT 0,
    pending_merge_source varchar(36),
    pending_merge_base varchar(36),
    created_at timestamptz,
    updated_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_decks_owner ON decks(owner_id);
CREATE INDEX IF NOT EXISTS idx_decks_visibility ON decks(visibility);

CREATE TABLE IF NOT EXISTS entries (
    id uuid PRIMARY KEY,
    deck_id uuid NOT NULL,
    logical_id uuid NOT NULL,
    term varchar(256) NOT NULL,
    definition varchar(512) NOT NULL,
    tags jsonb,
    created_at timestamptz,
    updated_at timestamptz,
    CONSTRAINT idx_entries_deck_logical UNIQUE (deck_id, logical_id)
);

CREATE TABLE IF NOT EXISTS commits (
    id uuid PRIMARY KEY,
    deck_id uuid NOT NULL,
    hash varchar(36) NOT NULL,
    seq bigint NOT NULL DEFAULT 0,
    parent_hashes jsonb,
    author_id uuid NOT NULL,
    changes jsonb,
    message varchar(500),
    created_at timestamptz,
    CONSTRAINT idx_commits_deck_hash UNIQUE (deck_id, hash)
);

CREATE TABLE IF NOT EXISTS conflicts (
    id uuid PRIMARY KEY,
    deck_id uuid NOT NULL,
    entry_id uuid NOT NULL,
    base jsonb,
    ours jsonb,
    theirs jsonb,
    status varchar(16) DEFAULT 'open',
    resolution varchar(16),
    resolved_snap jsonb,
    created_at timestamptz,
    updated_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_conflicts_deck ON conflicts(deck_id);

CREATE TABLE IF NOT EXISTS comments (
    id uuid PRIMARY KEY,
    deck_id uuid NOT NULL,
    user_id uuid NOT NULL,
    content text NOT NULL,
    created_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_comments_deck ON comments(deck_id);

CREATE TABLE IF NOT EXISTS likes (
    id uuid PRIMARY KEY,
    deck_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz,
    CONSTRAINT idx_likes_deck_user UNIQUE (deck_id, user_id)
);

CREATE TABLE IF NOT EXISTS game_sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    deck_id uuid NOT NULL,
    mode varchar(16),
    level integer DEFAULT 0,
    block_count integer DEFAULT 0,
    duration_ms bigint DEFAULT 0,
    correct_count integer DEFAULT 0,
    wrong_count integer DEFAULT 0,
    completed_at timestamptz,
    questions jsonb,
    created_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_game_sessions_user ON game_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_deck ON game_sessions(deck_id);
