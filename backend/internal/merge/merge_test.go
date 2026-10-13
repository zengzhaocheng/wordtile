package merge

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func st(id uuid.UUID, term, def string, tags ...string) EntryState {
	return EntryState{ID: id, Term: term, Definition: def, Tags: tags}
}

func TestApplyAndDiffRoundTrip(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()

	base := StateMap{a: st(a, "苹果", "apple", "食物"), b: st(b, "猫", "cat", "动物")}
	ch := Diff(base, StateMap{
		a: st(a, "苹果（改）", "apple", "食物"),
		c: st(c, "新词", "new"),
	})
	assert.Len(t, ch.Updated, 1)
	assert.Len(t, ch.Added, 1)
	assert.Len(t, ch.Removed, 1)
	assert.Equal(t, b, ch.Removed[0])

	next := Apply(base, ch.Added, ch.Updated, ch.Removed)
	assert.Equal(t, "苹果（改）", next[a].Term)
	assert.Contains(t, next, c)
	assert.NotContains(t, next, b)

	// 再 diff 一次应为空
	assert.True(t, Diff(next, next).Empty())
}

func TestThreeWay_NoConflict(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	base := StateMap{a: st(a, "苹果", "apple"), b: st(b, "猫", "cat")}
	ours := StateMap{a: st(a, "苹果", "apple（本地改）"), b: st(b, "猫", "cat")} // 改 a，b 保持不变
	theirs := Apply(base, []EntryState{st(b, "猫", "cat（远程改）")}, nil, nil) // 改 b

	res := ThreeWay(base, ours, theirs)
	assert.Empty(t, res.Conflicts)
	assert.Equal(t, "apple（本地改）", res.Merged[a].Definition)
	assert.Equal(t, "cat（远程改）", res.Merged[b].Definition)
}

func TestThreeWay_SameChangeBothSides(t *testing.T) {
	a := uuid.New()
	base := StateMap{a: st(a, "苹果", "apple")}
	ours := StateMap{a: st(a, "苹果", "APPLE")}
	res := ThreeWay(base, ours, ours)
	assert.Empty(t, res.Conflicts)
	assert.Equal(t, "APPLE", res.Merged[a].Definition)
}

func TestThreeWay_ConflictWhenBothModified(t *testing.T) {
	a := uuid.New()
	base := StateMap{a: st(a, "苹果", "apple")}
	ours := StateMap{a: st(a, "苹果", "apple-local")}
	theirs := StateMap{a: st(a, "苹果", "apple-remote")}

	res := ThreeWay(base, ours, theirs)
	assert.Len(t, res.Conflicts, 1)
	assert.Equal(t, a, res.Conflicts[0].EntryID)
	assert.Equal(t, "apple-local", res.Conflicts[0].Ours.Definition)
	assert.Equal(t, "apple-remote", res.Conflicts[0].Theirs.Definition)
	assert.True(t, res.Conflicts[0].Base.Exists)
	// 冲突项暂取本地
	assert.Equal(t, "apple-local", res.Merged[a].Definition)
}

func TestThreeWay_DeleteVsModifyConflict(t *testing.T) {
	a := uuid.New()
	base := StateMap{a: st(a, "苹果", "apple")}
	ours := StateMap{}                                            // 本地删除
	theirs := StateMap{a: st(a, "苹果", "apple-remote")} // 远程修改

	res := ThreeWay(base, ours, theirs)
	assert.Len(t, res.Conflicts, 1)
	assert.False(t, res.Conflicts[0].Ours.Exists)
	assert.True(t, res.Conflicts[0].Theirs.Exists)
}

func TestThreeWay_BothDeleteNoConflict(t *testing.T) {
	a := uuid.New()
	base := StateMap{a: st(a, "苹果", "apple")}
	res := ThreeWay(base, StateMap{}, StateMap{})
	assert.Empty(t, res.Conflicts)
	assert.NotContains(t, res.Merged, a)
}

func TestThreeWay_AddOnBothSidesDifferentContent(t *testing.T) {
	// 同一逻辑 ID 双方各自添加且内容不同 → 冲突
	a := uuid.New()
	base := StateMap{}
	ours := StateMap{a: st(a, "苹果", "apple")}
	theirs := StateMap{a: st(a, "苹果", "APPL")}
	res := ThreeWay(base, ours, theirs)
	assert.Len(t, res.Conflicts, 1)
}

func TestLCA(t *testing.T) {
	parents := map[string][]string{
		"c1": {},
		"c2": {"c1"},
		"c3": {"c2"},
		"c4": {"c2"},
		"c5": {"c3"},
	}
	// c5（基于 c3）与 c4（基于 c2）的公共祖先是 c2
	got, ok := LCA(parents, "c5", "c4")
	assert.True(t, ok)
	assert.Equal(t, "c2", got)

	// 自身比较
	got, ok = LCA(parents, "c3", "c3")
	assert.True(t, ok)
	assert.Equal(t, "c3", got)

	// 无公共祖先
	_, ok = LCA(parents, "c9", "c4")
	assert.False(t, ok)
}
