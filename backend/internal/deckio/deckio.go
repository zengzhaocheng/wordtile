package deckio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// WordTile 题库交换格式（Deck Exchange Format）v1。
// 外部工具只需生成一个 JSON 对象或数组即可导入，见 README「题库导入 / 导出」。
const (
	Format     = "wordtile-deck"
	Version    = 1
	MaxEntries = 5000
	MaxTermLen = 200
	MaxDefLen  = 1000
	MaxTags    = 20
	MaxTagLen  = 50
)

// EntryItem 是一条可交换的词条；tags 可省略。
type EntryItem struct {
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Tags       []string `json:"tags,omitempty"`
}

// DeckFile 是题库 JSON 文件的结构。format/version/name/description 均可省略，
// 最小可用文件形如：{"entries":[{"term":"apple","definition":"苹果"}]}
type DeckFile struct {
	Format      string      `json:"format,omitempty"`
	Version     int         `json:"version,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	ExportedAt  string      `json:"exported_at,omitempty"`
	Entries     []EntryItem `json:"entries"`
}

// NewExport 生成用于导出的完整文档。
func NewExport(name, description string, items []EntryItem) DeckFile {
	return DeckFile{
		Format:      Format,
		Version:     Version,
		Name:        name,
		Description: description,
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Entries:     items,
	}
}

// Parse 解析并校验题库文件。兼容两种输入：
//  1. 完整文档 {"format":"wordtile-deck","version":1,"name":"...","entries":[...]}
//  2. 纯数组   [{"term":"apple","definition":"苹果"}]
func Parse(data []byte) (*DeckFile, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, errors.New("文件内容为空")
	}
	var f DeckFile
	if strings.HasPrefix(trimmed, "[") {
		var items []EntryItem
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, fmt.Errorf("JSON 数组格式错误: %w", err)
		}
		f = DeckFile{Entries: items}
	} else {
		if err := json.Unmarshal(data, &f); err != nil {
			return nil, fmt.Errorf("JSON 格式错误: %w", err)
		}
	}
	if f.Format != "" && f.Format != Format {
		return nil, fmt.Errorf("不支持的 format %q（应为 %q）", f.Format, Format)
	}
	if f.Version > Version {
		return nil, fmt.Errorf("不支持的 version %d（当前最高 %d）", f.Version, Version)
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

// Validate 校验并就地规范化（去空白、去空标签、补全默认 tags）。
func (f *DeckFile) Validate() error {
	if len(f.Entries) == 0 {
		return errors.New("entries 不能为空，至少需要 1 条词条")
	}
	if len(f.Entries) > MaxEntries {
		return fmt.Errorf("词条数量 %d 超过上限 %d", len(f.Entries), MaxEntries)
	}
	for i := range f.Entries {
		e := &f.Entries[i]
		e.Term = strings.TrimSpace(e.Term)
		e.Definition = strings.TrimSpace(e.Definition)
		if e.Term == "" || e.Definition == "" {
			return fmt.Errorf("第 %d 条词条的 term / definition 不能为空", i+1)
		}
		if len([]rune(e.Term)) > MaxTermLen {
			return fmt.Errorf("第 %d 条词条的 term 超过 %d 字", i+1, MaxTermLen)
		}
		if len([]rune(e.Definition)) > MaxDefLen {
			return fmt.Errorf("第 %d 条词条的 definition 超过 %d 字", i+1, MaxDefLen)
		}
		if len(e.Tags) > MaxTags {
			return fmt.Errorf("第 %d 条词条的标签超过 %d 个", i+1, MaxTags)
		}
		tags := make([]string, 0, len(e.Tags))
		for _, tag := range e.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if len([]rune(tag)) > MaxTagLen {
				return fmt.Errorf("第 %d 条词条的标签超过 %d 字", i+1, MaxTagLen)
			}
			tags = append(tags, tag)
		}
		e.Tags = tags
	}
	f.Name = strings.TrimSpace(f.Name)
	f.Description = strings.TrimSpace(f.Description)
	return nil
}

// ParseImport 解析「导入到已有题库」请求，额外读取 mode（append / replace，默认 append）。
func ParseImport(data []byte) (*DeckFile, string, error) {
	f, err := Parse(data)
	if err != nil {
		return nil, "", err
	}
	var probe struct {
		Mode string `json:"mode"`
	}
	_ = json.Unmarshal(data, &probe) // 纯数组输入会失败，此时用默认 mode
	mode := strings.ToLower(strings.TrimSpace(probe.Mode))
	if mode == "" {
		mode = "append"
	}
	if mode != "append" && mode != "replace" {
		return nil, "", errors.New("mode 仅支持 append / replace")
	}
	return f, mode, nil
}
