-- 题库「归档」标记：仅从「我的题库」移除，公开版本继续留在市场。
-- 幂等，可与 0001 之后直接使用 AutoMigrate 的库共存。
ALTER TABLE decks ADD COLUMN IF NOT EXISTS archived boolean DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_decks_archived ON decks(archived);
