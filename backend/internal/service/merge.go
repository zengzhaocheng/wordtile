package service

import (
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/apperr"
	"github.com/wordtile/backend/internal/merge"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/repository"
)

// MergeService Git-like 题库版本与合并：
// commit 保存、三路合并（Pull / 合并回原题库）、冲突解决工作台。
type MergeService struct {
	repo *repository.DB
}

func NewMergeService(repo *repository.DB) *MergeService { return &MergeService{repo: repo} }

/* ---------- 状态重建 ---------- */

func toState(e model.Entry) merge.EntryState {
	return merge.EntryState{ID: e.LogicalID, Term: e.Term, Definition: e.Definition, Tags: e.Tags}
}

// entriesState 工作区（entries 表）当前状态。
func (s *MergeService) entriesState(deckID uuid.UUID) (merge.StateMap, error) {
	entries, err := s.repo.ListEntries(deckID)
	if err != nil {
		return nil, err
	}
	st := merge.StateMap{}
	for _, e := range entries {
		st[e.LogicalID] = toState(e)
	}
	return st, nil
}

// replayState 重放 commit 链直到 stopHash（含），得到该 commit 的词条快照。
func (s *MergeService) replayState(commits []model.Commit, stopHash string) merge.StateMap {
	st := merge.StateMap{}
	for _, c := range commits {
		var added, updated []merge.EntryState
		for _, a := range c.Changes.Added {
			added = append(added, merge.EntryState{ID: a.ID, Term: a.Term, Definition: a.Definition, Tags: a.Tags})
		}
		for _, u := range c.Changes.Updated {
			updated = append(updated, merge.EntryState{ID: u.ID, Term: u.Term, Definition: u.Definition, Tags: u.Tags})
		}
		st = merge.Apply(st, added, updated, c.Changes.Removed)
		if stopHash != "" && c.Hash == stopHash {
			break
		}
	}
	return st
}

// applyChanges 把变更集写入工作区（按 LogicalID upsert / delete）。
func (s *MergeService) applyChanges(deckID uuid.UUID, ch merge.Changes) error {
	for _, id := range ch.Removed {
		if err := s.repo.DeleteEntryByLogicalID(deckID, id); err != nil {
			return err
		}
	}
	for _, u := range append(append([]merge.EntryState{}, ch.Updated...), ch.Added...) {
		if err := s.repo.UpsertEntryByLogicalID(deckID, u.ID, u.Term, u.Definition, u.Tags); err != nil {
			return err
		}
	}
	return nil
}

func (s *MergeService) createCommit(deck *model.Deck, authorID uuid.UUID, parents []string, ch merge.Changes, message string) (string, error) {
	seq, err := s.repo.MaxCommitSeq(deck.ID)
	if err != nil {
		return "", err
	}
	cleanParents := make([]string, 0, len(parents))
	for _, p := range parents {
		if p != "" {
			cleanParents = append(cleanParents, p)
		}
	}
	hash := uuid.NewString()
	commit := &model.Commit{
		ID: uuid.New(), DeckID: deck.ID, Hash: hash, Seq: seq + 1,
		ParentHashes: cleanParents, AuthorID: authorID,
		Changes: model.EntryChanges{
			Added:   toModelChanges(ch.Added),
			Updated: toModelChanges(ch.Updated),
			Removed: ch.Removed,
		},
		Message: message,
	}
	if err := s.repo.CreateCommit(commit); err != nil {
		return "", err
	}
	deck.HeadCommitID = &hash
	if err := s.repo.UpdateDeck(deck); err != nil {
		return "", err
	}
	return hash, nil
}

func toModelChanges(states []merge.EntryState) []model.EntryChange {
	out := make([]model.EntryChange, 0, len(states))
	for _, st := range states {
		out = append(out, model.EntryChange{ID: st.ID, Term: st.Term, Definition: st.Definition, Tags: st.Tags})
	}
	return out
}

/* ---------- 对外流程 ---------- */

// Commit 把工作区相对 head 的变更固化为一个 commit。
func (s *MergeService) Commit(userID, deckID uuid.UUID, message string) (*model.Commit, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	commits, err := s.repo.ListCommits(deckID)
	if err != nil {
		return nil, err
	}
	head := ""
	if deck.HeadCommitID != nil {
		head = *deck.HeadCommitID
	}
	base := s.replayState(commits, head)
	current, err := s.entriesState(deckID)
	if err != nil {
		return nil, err
	}
	ch := merge.Diff(base, current)
	if ch.Empty() {
		return nil, apperr.New(400, "没有可提交的变更")
	}
	if message == "" {
		message = "更新词条"
	}
	var parents []string
	if head != "" {
		parents = []string{head}
	}
	if _, err := s.createCommit(deck, userID, parents, ch, message); err != nil {
		return nil, err
	}
	created, err := s.repo.ListCommits(deckID)
	if err != nil {
		return nil, err
	}
	return &created[len(created)-1], nil
}

// MergeResponse 合并结果。
type MergeResponse struct {
	Status       string           `json:"status"` // up_to_date / merged / pending_conflicts
	MergedCount  int              `json:"merged_count"`
	Conflicts    int              `json:"conflicts"`
	ConflictList []model.Conflict `json:"conflict_list,omitempty"`
}

// Merge 三路合并：把 sourceCommit（来自 sourceDeck）合并进 deck。
// sourceDeck 缺省取 deck.ForkedFrom（即 Pull 远程更新）；也可把 fork 的 head 合并回原题库。
func (s *MergeService) Merge(userID, deckID, sourceDeckID uuid.UUID, sourceHash string) (*MergeResponse, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	if sourceDeckID == uuid.Nil {
		if deck.ForkedFrom == nil {
			return nil, apperr.New(400, "缺少 source_deck_id（且该题库不是 fork）")
		}
		sourceDeckID = *deck.ForkedFrom
	}
	sourceDeck, err := s.repo.GetDeck(sourceDeckID)
	if err != nil {
		return nil, apperr.New(404, "来源题库不存在")
	}
	if sourceHash == "" && sourceDeck.HeadCommitID != nil {
		sourceHash = *sourceDeck.HeadCommitID
	}

	localHead := ""
	if deck.HeadCommitID != nil {
		localHead = *deck.HeadCommitID
	}
	if sourceHash == "" || sourceHash == localHead {
		return &MergeResponse{Status: "up_to_date"}, nil
	}

	localCommits, err := s.repo.ListCommits(deckID)
	if err != nil {
		return nil, err
	}
	sourceCommits, err := s.repo.ListCommits(sourceDeckID)
	if err != nil {
		return nil, err
	}

	// 跨题库构建 commit DAG（hash 在 fork 时保留，因此可互通）
	parents := map[string][]string{}
	for _, c := range localCommits {
		parents[c.Hash] = c.ParentHashes
	}
	for _, c := range sourceCommits {
		parents[c.Hash] = c.ParentHashes
	}

	lcaHash, hasLCA := merge.LCA(parents, localHead, sourceHash)
	base := merge.StateMap{}
	if hasLCA {
		base = s.replayState(localCommits, lcaHash)
	}

	ours, err := s.entriesState(deckID)
	if err != nil {
		return nil, err
	}
	theirs := s.replayState(sourceCommits, sourceHash)

	result := merge.ThreeWay(base, ours, theirs)

	// 自动合并部分直接写入工作区
	auto := merge.Diff(ours, result.Merged)
	if err := s.applyChanges(deckID, auto); err != nil {
		return nil, err
	}
	mergedCount := len(ours) + len(auto.Added) - len(auto.Removed)

	if len(result.Conflicts) == 0 {
		commitChanges := merge.Diff(base, result.Merged)
		if !commitChanges.Empty() {
			if _, err := s.createCommit(deck, userID, []string{localHead, sourceHash}, commitChanges,
				"Merge "+sourceDeck.Name+" ("+sourceHash[:8]+")"); err != nil {
				return nil, err
			}
		}
		return &MergeResponse{Status: "merged", MergedCount: mergedCount}, nil
	}

	// 有冲突：写入冲突表，挂起合并来源，待工作台逐条解决
	for _, cf := range result.Conflicts {
		if err := s.repo.CreateConflict(&model.Conflict{
			ID: uuid.New(), DeckID: deckID, EntryID: cf.EntryID,
			Base: toModelSnap(cf.Base), Ours: toModelSnap(cf.Ours), Theirs: toModelSnap(cf.Theirs),
			Status: "open",
		}); err != nil {
			return nil, err
		}
	}
	src, lca := sourceHash, lcaHash
	deck.PendingMergeSource = &src
	deck.PendingMergeBase = &lca
	if err := s.repo.UpdateDeck(deck); err != nil {
		return nil, err
	}
	openList, err := s.repo.ListConflicts(deckID, "open")
	if err != nil {
		return nil, err
	}
	return &MergeResponse{Status: "pending_conflicts", MergedCount: mergedCount, Conflicts: len(openList), ConflictList: openList}, nil
}

func toModelSnap(s merge.Snapshot) model.EntrySnapshot {
	return model.EntrySnapshot{Exists: s.Exists, Term: s.Term, Definition: s.Definition, Tags: s.Tags}
}

func fromModelSnap(s model.EntrySnapshot) merge.Snapshot {
	return merge.Snapshot{Exists: s.Exists, Term: s.Term, Definition: s.Definition, Tags: s.Tags}
}

// ListConflicts 冲突列表（默认 open）。
func (s *MergeService) ListConflicts(userID, deckID uuid.UUID, status string) ([]model.Conflict, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	if status == "" {
		status = "open"
	}
	return s.repo.ListConflicts(deckID, status)
}

// ResolveResult 解决冲突后的状态。
type ResolveResult struct {
	Remaining        int    `json:"remaining"`
	AutoCommitted    bool   `json:"auto_committed"`
	Status           string `json:"status"` // resolved_all / resolved_one
}

// ResolveConflict 在合并工作台解决单条冲突；全部解决后自动生成 merge commit。
func (s *MergeService) ResolveConflict(userID, deckID, conflictID uuid.UUID, resolution string, custom model.EntrySnapshot) (*ResolveResult, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if deck.OwnerID != userID {
		return nil, apperr.ErrForbidden
	}
	cf, err := s.repo.GetConflict(conflictID)
	if err != nil || cf.DeckID != deckID {
		return nil, apperr.New(404, "冲突不存在")
	}
	if cf.Status != "open" {
		return nil, apperr.New(409, "该冲突已解决")
	}

	var chosen model.EntrySnapshot
	switch resolution {
	case "ours":
		chosen = cf.Ours
	case "theirs":
		chosen = cf.Theirs
	case "custom":
		if !custom.Exists || custom.Term == "" || custom.Definition == "" {
			return nil, apperr.New(400, "自定义内容不完整")
		}
		chosen = custom
	default:
		return nil, apperr.New(400, "resolution 仅支持 ours / theirs / custom")
	}

	if chosen.Exists {
		if err := s.repo.UpsertEntryByLogicalID(deckID, cf.EntryID, chosen.Term, chosen.Definition, chosen.Tags); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.DeleteEntryByLogicalID(deckID, cf.EntryID); err != nil {
			return nil, err
		}
	}
	cf.Status = "resolved"
	cf.Resolution = resolution
	cf.ResolvedSnap = chosen
	if err := s.repo.UpdateConflict(cf); err != nil {
		return nil, err
	}

	remaining, err := s.repo.ListConflicts(deckID, "open")
	if err != nil {
		return nil, err
	}
	res := &ResolveResult{Remaining: len(remaining), Status: "resolved_one"}

	// 全部解决 → 生成 merge commit（base 取挂起的 pending base）
	if len(remaining) == 0 && deck.PendingMergeSource != nil {
		commits, err := s.repo.ListCommits(deckID)
		if err != nil {
			return nil, err
		}
		base := merge.StateMap{}
		if deck.PendingMergeBase != nil && *deck.PendingMergeBase != "" {
			base = s.replayState(commits, *deck.PendingMergeBase)
		}
		current, err := s.entriesState(deckID)
		if err != nil {
			return nil, err
		}
		ch := merge.Diff(base, current)
		localHead := ""
		if deck.HeadCommitID != nil {
			localHead = *deck.HeadCommitID
		}
		if !ch.Empty() {
			if _, err := s.createCommit(deck, userID, []string{localHead, *deck.PendingMergeSource}, ch,
				"Merge (resolved) "+(*deck.PendingMergeSource)[:8]); err != nil {
				return nil, err
			}
			res.AutoCommitted = true
		}
		deck.PendingMergeSource = nil
		deck.PendingMergeBase = nil
		if err := s.repo.UpdateDeck(deck); err != nil {
			return nil, err
		}
		res.Status = "resolved_all"
	}
	return res, nil
}
