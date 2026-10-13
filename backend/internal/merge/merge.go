// Package merge 实现 Git-like 题库版本合并的核心算法：
// 状态重放（Replay）、差异计算（Diff）、最近公共祖先（LCA）与三路合并（ThreeWay）。
// 全部为纯函数，不依赖数据库，便于单元测试。
package merge

import (
	"sort"

	"github.com/google/uuid"
)

// EntryState 词条状态（以 LogicalID 为键）。
type EntryState struct {
	ID         uuid.UUID `json:"id"`
	Term       string    `json:"term"`
	Definition string    `json:"definition"`
	Tags       []string  `json:"tags"`
}

// StateMap 一个题库在某个 commit 上的词条快照。
type StateMap map[uuid.UUID]EntryState

func (s StateMap) Clone() StateMap {
	n := make(StateMap, len(s))
	for k, v := range s {
		n[k] = EntryState{ID: v.ID, Term: v.Term, Definition: v.Definition, Tags: append([]string{}, v.Tags...)}
	}
	return n
}

// Apply 返回应用一组变更后的新状态（不修改入参）。
func Apply(state StateMap, added, updated []EntryState, removed []uuid.UUID) StateMap {
	n := state.Clone()
	for _, id := range removed {
		delete(n, id)
	}
	for _, u := range updated {
		n[u.ID] = EntryState{ID: u.ID, Term: u.Term, Definition: u.Definition, Tags: append([]string{}, u.Tags...)}
	}
	for _, a := range added {
		n[a.ID] = EntryState{ID: a.ID, Term: a.Term, Definition: a.Definition, Tags: append([]string{}, a.Tags...)}
	}
	return n
}

func equalTags(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ac, bc := append([]string{}, a...), append([]string{}, b...)
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

// Equal 比较两个词条内容是否一致（忽略标签顺序）。
func Equal(a, b EntryState) bool {
	return a.Term == b.Term && a.Definition == b.Definition && equalTags(a.Tags, b.Tags)
}

// Diff 计算 base → current 的变更集。
type Changes struct {
	Added   []EntryState `json:"added,omitempty"`
	Updated []EntryState `json:"updated,omitempty"`
	Removed []uuid.UUID  `json:"removed,omitempty"`
}

func (c Changes) Empty() bool {
	return len(c.Added) == 0 && len(c.Updated) == 0 && len(c.Removed) == 0
}

func Diff(base, current StateMap) Changes {
	var ch Changes
	for id, cur := range current {
		if b, ok := base[id]; ok {
			if !Equal(b, cur) {
				ch.Updated = append(ch.Updated, cur)
			}
		} else {
			ch.Added = append(ch.Added, cur)
		}
	}
	for id := range base {
		if _, ok := current[id]; !ok {
			ch.Removed = append(ch.Removed, id)
		}
	}
	sort.Slice(ch.Added, func(i, j int) bool { return ch.Added[i].ID.String() < ch.Added[j].ID.String() })
	sort.Slice(ch.Updated, func(i, j int) bool { return ch.Updated[i].ID.String() < ch.Updated[j].ID.String() })
	sort.Slice(ch.Removed, func(i, j int) bool { return ch.Removed[i].String() < ch.Removed[j].String() })
	return ch
}

// LCA 在 commit DAG（hash → 父 hash 列表）中求 a、b 的最近公共祖先。
func LCA(parents map[string][]string, a, b string) (string, bool) {
	if a == "" || b == "" {
		return "", false
	}
	ancestorsOfA := map[string]bool{}
	queue := []string{a}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if ancestorsOfA[cur] {
			continue
		}
		ancestorsOfA[cur] = true
		queue = append(queue, parents[cur]...)
	}
	seen := map[string]bool{}
	queue = []string{b}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if ancestorsOfA[cur] {
			return cur, true
		}
		if seen[cur] {
			continue
		}
		seen[cur] = true
		queue = append(queue, parents[cur]...)
	}
	return "", false
}

// Snapshot 冲突时某一侧的词条快照。
type Snapshot struct {
	Exists     bool     `json:"exists"`
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Tags       []string `json:"tags"`
}

func Snap(s StateMap, id uuid.UUID) Snapshot {
	if e, ok := s[id]; ok {
		return Snapshot{Exists: true, Term: e.Term, Definition: e.Definition, Tags: append([]string{}, e.Tags...)}
	}
	return Snapshot{}
}

// Conflict 一条无法自动合并的冲突。
type Conflict struct {
	EntryID uuid.UUID
	Base    Snapshot
	Ours    Snapshot
	Theirs  Snapshot
}

// Result 三路合并结果。Merged 中冲突项暂取「本地」一侧，待用户解决后覆盖。
type Result struct {
	Merged    StateMap
	Conflicts []Conflict
}

func snapEqual(s StateMap, id uuid.UUID, other StateMap) bool {
	a, aok := s[id]
	b, bok := other[id]
	if aok != bok {
		return false
	}
	if !aok {
		return true
	}
	return Equal(a, b)
}

// ThreeWay 经典三路合并：
//   - 两侧都未变 → 保留；
//   - 仅一侧变更 → 采用该侧；
//   - 两侧相同变更 → 采用；
//   - 两侧不同变更（含 删除 vs 修改）→ 标记冲突，暂取本地侧；
//   - 两侧都删除 → 删除，不算冲突。
func ThreeWay(base, ours, theirs StateMap) Result {
	ids := map[uuid.UUID]bool{}
	for id := range base {
		ids[id] = true
	}
	for id := range ours {
		ids[id] = true
	}
	for id := range theirs {
		ids[id] = true
	}

	merged := StateMap{}
	var conflicts []Conflict
	for id := range ids {
		o, ook := ours[id]
		t, tok := theirs[id]
		oChanged := !snapEqual(ours, id, base)
		tChanged := !snapEqual(theirs, id, base)

		switch {
		case !oChanged && !tChanged:
			if ook {
				merged[id] = o
			}
		case oChanged && !tChanged:
			if ook {
				merged[id] = o
			}
		case !oChanged && tChanged:
			if tok {
				merged[id] = t
			}
		default: // 双方均有变更
			if !ook && !tok {
				// 双方都删除 → 无冲突
			} else if ook && tok && Equal(o, t) {
				merged[id] = o
			} else {
				conflicts = append(conflicts, Conflict{
					EntryID: id,
					Base:    Snap(base, id),
					Ours:    Snap(ours, id),
					Theirs:  Snap(theirs, id),
				})
				if ook { // 暂取本地
					merged[id] = o
				}
			}
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].EntryID.String() < conflicts[j].EntryID.String() })
	return Result{Merged: merged, Conflicts: conflicts}
}
