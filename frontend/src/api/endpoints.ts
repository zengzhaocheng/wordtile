import { api } from "./client";
import type {
  AuthResponse,
  CampaignStat,
  CommentItem,
  Commit,
  Conflict,
  ContributionDay,
  Deck,
  DeckDetail,
  Entry,
  FinishSummary,
  GameStartResponse,
  MarketDeck,
  MatchResponse,
  MergeResponse,
  ProgressPoint,
  Tile,
  User,
} from "../types";

/* ---------- 认证 ---------- */
export const authApi = {
  register: (body: { username: string; email: string; password: string; display_name?: string }) =>
    api.post<AuthResponse>("/auth/register", body).then((r) => r.data),
  login: (body: { identifier: string; password: string }) =>
    api.post<AuthResponse>("/auth/login", body).then((r) => r.data),
  me: () => api.get<{ user: User }>("/users/me").then((r) => r.data.user),
  updateMe: (body: Partial<Pick<User, "display_name" | "avatar_url" | "theme_preference">>) =>
    api.patch<{ user: User }>("/users/me", body).then((r) => r.data.user),
};

/* ---------- 题库 ---------- */
export const deckApi = {
  listMine: () => api.get<{ decks: Deck[] }>("/decks").then((r) => r.data.decks),
  create: (body: { name: string; description: string }) =>
    api.post<{ deck: Deck }>("/decks", body).then((r) => r.data.deck),
  detail: (id: string) => api.get<DeckDetail>(`/decks/${id}`).then((r) => r.data),
  update: (id: string, body: { name?: string; description?: string; visibility?: string }) =>
    api.patch<{ deck: Deck }>(`/decks/${id}`, body).then((r) => r.data.deck),
  remove: (id: string) => api.delete(`/decks/${id}`),
  createEntry: (deckId: string, body: { term: string; definition: string; tags: string[] }) =>
    api.post<{ entry: Entry }>(`/decks/${deckId}/entries`, body).then((r) => r.data.entry),
  updateEntry: (entryId: string, body: { term?: string; definition?: string; tags?: string[] }) =>
    api.patch<{ entry: Entry }>(`/entries/${entryId}`, body).then((r) => r.data.entry),
  deleteEntry: (entryId: string) => api.delete(`/entries/${entryId}`),
};

/* ---------- 社区 ---------- */
export const marketApi = {
  list: (search: string, sort: "likes" | "newest" | "forks") =>
    api
      .get<{ decks: MarketDeck[] }>("/marketplace/decks", { params: { search, sort } })
      .then((r) => r.data.decks),
  publish: (deckId: string) =>
    api.post<{ deck: Deck }>(`/decks/${deckId}/publish`).then((r) => r.data.deck),
  fork: (deckId: string) => api.post<{ deck: Deck }>(`/decks/${deckId}/fork`).then((r) => r.data.deck),
  like: (deckId: string) =>
    api.post<{ liked: boolean; likes_count: number }>(`/decks/${deckId}/like`).then((r) => r.data),
  comments: (deckId: string) =>
    api.get<{ comments: CommentItem[] }>(`/decks/${deckId}/comments`).then((r) => r.data.comments),
  addComment: (deckId: string, content: string) =>
    api.post<{ comment: CommentItem }>(`/decks/${deckId}/comments`, { content }).then((r) => r.data.comment),
};

/* ---------- 版本与合并 ---------- */
export const versionApi = {
  commits: (deckId: string) =>
    api.get<{ commits: Commit[] }>(`/decks/${deckId}/commits`).then((r) => r.data.commits),
  commit: (deckId: string, message: string) =>
    api.post<{ commit: Commit }>(`/decks/${deckId}/commits`, { message }).then((r) => r.data.commit),
  merge: (deckId: string, body: { source_deck_id?: string; source_commit_id?: string }) =>
    api.post<MergeResponse>(`/decks/${deckId}/merge`, body).then((r) => r.data),
  conflicts: (deckId: string, status = "open") =>
    api.get<{ conflicts: Conflict[] }>(`/decks/${deckId}/conflicts`, { params: { status } }).then((r) => r.data.conflicts),
  resolve: (
    deckId: string,
    body: { conflict_id: string; resolution: "ours" | "theirs" | "custom"; custom?: { exists: boolean; term: string; definition: string; tags: string[] } },
  ) =>
    api
      .post<{ remaining: number; auto_committed: boolean; status: string }>(`/decks/${deckId}/conflicts/resolve`, body)
      .then((r) => r.data),
};

/* ---------- 游戏 ---------- */
export const gameApi = {
  start: (body: { deck_id: string; mode: "custom" | "campaign"; block_count?: number; level?: number; tags?: string[] }) =>
    api.post<GameStartResponse>("/games/start", body).then((r) => r.data),
  match: (sessionId: string, termToken: string, definitionToken: string) =>
    api
      .post<MatchResponse>(`/games/${sessionId}/match`, { term_token: termToken, definition_token: definitionToken })
      .then((r) => r.data),
  hint: (sessionId: string) =>
    api.post<{ term_token: string; definition_token: string }>(`/games/${sessionId}/hint`).then((r) => r.data),
  finish: (sessionId: string, durationMs: number) =>
    api.post<FinishSummary>(`/games/${sessionId}/finish`, { duration_ms: durationMs }).then((r) => r.data),
  state: (sessionId: string) =>
    api
      .get<{
        session_id: string;
        mode: "custom" | "campaign";
        level: number;
        block_count: number;
        terms: Tile[];
        definitions: Tile[];
      }>(`/games/${sessionId}`)
      .then((r) => r.data),
};

/* ---------- 统计 ---------- */
export const statsApi = {
  contribution: () => api.get<{ days: ContributionDay[] }>("/stats/contribution").then((r) => r.data.days),
  deckProgress: (deckId: string) =>
    api.get<{ points: ProgressPoint[] }>(`/stats/decks/${deckId}/progress`).then((r) => r.data.points),
  campaign: () => api.get<{ levels: CampaignStat[] }>("/stats/campaign").then((r) => r.data.levels),
};
