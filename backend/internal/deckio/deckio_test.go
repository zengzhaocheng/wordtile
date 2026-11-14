package deckio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDocument(t *testing.T) {
	data := `{
		"format":"wordtile-deck","version":1,
		"name":"  水果词汇  ","description":"测试",
		"entries":[
			{"term":" apple ","definition":" 苹果 ","tags":[" 食物 ","", "水果"]},
			{"term":"banana","definition":"香蕉"}
		]}`
	f, err := Parse([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "水果词汇", f.Name)
	assert.Len(t, f.Entries, 2)
	assert.Equal(t, "apple", f.Entries[0].Term)
	assert.Equal(t, "苹果", f.Entries[0].Definition)
	assert.Equal(t, []string{"食物", "水果"}, f.Entries[0].Tags)
	assert.Equal(t, []string{}, f.Entries[1].Tags)
}

func TestParseBareArray(t *testing.T) {
	f, err := Parse([]byte(`[{"term":"cat","definition":"猫"}]`))
	require.NoError(t, err)
	assert.Empty(t, f.Format)
	assert.Len(t, f.Entries, 1)
	assert.Equal(t, "cat", f.Entries[0].Term)
}

func TestParseMinimalDocument(t *testing.T) {
	f, err := Parse([]byte(`{"entries":[{"term":"a","definition":"b"}]}`))
	require.NoError(t, err)
	assert.Equal(t, "a", f.Entries[0].Term)
}

func TestParseRejectsInvalid(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty", ""},
		{"bad json", "{not json"},
		{"no entries", `{"name":"x"}`},
		{"empty entries", `{"entries":[]}`},
		{"missing definition", `{"entries":[{"term":"a"}]}`},
		{"blank term", `{"entries":[{"term":"  ","definition":"b"}]}`},
		{"wrong format", `{"format":"other","entries":[{"term":"a","definition":"b"}]}`},
		{"future version", `{"version":99,"entries":[{"term":"a","definition":"b"}]}`},
		{"long term", `{"entries":[{"term":"` + strings.Repeat("x", 201) + `","definition":"b"}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.body))
			assert.Error(t, err)
		})
	}
}

func TestParseRejectsTooManyEntries(t *testing.T) {
	items := make([]string, MaxEntries+1)
	for i := range items {
		items[i] = `{"term":"a","definition":"b"}`
	}
	_, err := Parse([]byte(`{"entries":[` + strings.Join(items, ",") + `]}`))
	assert.Error(t, err)
}

func TestParseImportMode(t *testing.T) {
	body := `{"mode":"replace","entries":[{"term":"a","definition":"b"}]}`
	f, mode, err := ParseImport([]byte(body))
	require.NoError(t, err)
	assert.Equal(t, "replace", mode)
	assert.Len(t, f.Entries, 1)

	f, mode, err = ParseImport([]byte(`{"entries":[{"term":"a","definition":"b"}]}`))
	require.NoError(t, err)
	assert.Equal(t, "append", mode)

	_, _, err = ParseImport([]byte(`{"mode":"merge","entries":[{"term":"a","definition":"b"}]}`))
	assert.Error(t, err)

	// 纯数组输入时 mode 默认为 append
	_, mode, err = ParseImport([]byte(`[{"term":"a","definition":"b"}]`))
	require.NoError(t, err)
	assert.Equal(t, "append", mode)
}

func TestNewExport(t *testing.T) {
	f := NewExport("我的题库", "说明", []EntryItem{{Term: "a", Definition: "b"}})
	assert.Equal(t, Format, f.Format)
	assert.Equal(t, Version, f.Version)
	assert.NotEmpty(t, f.ExportedAt)
	require.NoError(t, f.Validate())
}
