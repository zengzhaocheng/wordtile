export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  theme_preference: "light" | "dark" | "system";
  created_at: string;
}

export interface Deck {
  id: string;
  owner_id: string;
  name: string;
  description: string;
  visibility: "private" | "public";
  head_commit_id: string | null;
  forked_from: string | null;
  likes_count: number;
  forks_count: number;
  created_at: string;
  updated_at: string;
}

export interface MarketDeck extends Deck {
  owner_name: string;
}

export interface Entry {
  id: string;
  deck_id: string;
  logical_id: string;
  term: string;
  definition: string;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export interface EntrySnapshot {
  exists: boolean;
  term: string;
  definition: string;
  tags: string[];
}

export interface Commit {
  id: string;
  deck_id: string;
  hash: string;
  seq: number;
  parent_hashes: string[];
  author_id: string;
  changes: {
    added?: { id: string; term: string; definition: string; tags: string[] }[];
    updated?: { id: string; term: string; definition: string; tags: string[] }[];
    removed?: string[];
  };
  message: string;
  created_at: string;
}

export interface Conflict {
  id: string;
  deck_id: string;
  entry_id: string;
  base: EntrySnapshot;
  ours: EntrySnapshot;
  theirs: EntrySnapshot;
  status: "open" | "resolved";
  resolution: string | null;
}

export interface CommentItem {
  id: string;
  deck_id: string;
  user_id: string;
  content: string;
  created_at: string;
  author_name: string;
}

export interface Tile {
  token: string;
  text: string;
}

export interface GameStartResponse {
  session_id: string;
  mode: "custom" | "campaign";
  level: number;
  deck_id: string;
  deck_name: string;
  block_count: number;
  terms: Tile[];
  definitions: Tile[];
}

export interface MatchResponse {
  correct: boolean;
  correct_count: number;
  wrong_count: number;
  finished: boolean;
}

export interface FinishSummary {
  session_id: string;
  duration_ms: number;
  correct_count: number;
  wrong_count: number;
  accuracy: number;
  mode: "custom" | "campaign";
  level: number;
}

export interface ContributionDay {
  date: string;
  sessions: number;
  duration_ms: number;
  accuracy: number;
}

export interface ProgressPoint {
  completed_at: string;
  duration_ms: number;
  correct_count: number;
  wrong_count: number;
  error_rate: number;
}

export interface CampaignStat {
  level: number;
  plays: number;
  best_ms: number;
  best_wrong: number;
}

export interface DeckDetail {
  deck: Deck;
  entries: Entry[];
  origin_head?: string;
  open_conflicts: number;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface MergeResponse {
  status: "up_to_date" | "merged" | "pending_conflicts";
  merged_count: number;
  conflicts: number;
  conflict_list?: Conflict[];
}
