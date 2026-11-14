DROP INDEX IF EXISTS idx_decks_archived;
ALTER TABLE decks DROP COLUMN IF EXISTS archived;
