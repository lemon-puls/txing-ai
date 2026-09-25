package builtin

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// 短文档不分片
func TestSplitMarkdownChunks_Short(t *testing.T) {
	content := "# 标题\n\n正文内容"
	chunks := splitMarkdownChunks(content, 100)
	if len(chunks) != 1 || chunks[0] != content {
		t.Fatalf("expect single untouched chunk, got %d: %q", len(chunks), chunks)
	}
}

// 长文档在标题边界分片，且每片不超过 maxRunes
func TestSplitMarkdownChunks_ByHeadings(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 30; i++ {
		b.WriteString("## Section\n\n")
		b.WriteString(strings.Repeat("段落内容。", 200)) // 1000 rune/节
		b.WriteString("\n\n")
	}
	content := b.String()
	chunks := splitMarkdownChunks(content, 3500)
	if len(chunks) < 3 {
		t.Fatalf("expect multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if utf8.RuneCountInString(c) > 3500 {
			t.Fatalf("chunk %d exceeds maxRunes: %d", i, utf8.RuneCountInString(c))
		}
	}
	// 无内容丢失（分片只在空白处断开/规整，比较时忽略所有空白）
	stripSpace := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			if !strings.ContainsRune(" \t\n\r", r) {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	if stripSpace(strings.Join(chunks, "\n\n")) != stripSpace(content) {
		t.Fatal("content lost across chunks")
	}
}

// 单个超长段落（无标题无空行）按字符硬切
func TestSplitMarkdownChunks_HardSplit(t *testing.T) {
	content := strings.Repeat("字", 10000)
	chunks := splitMarkdownChunks(content, 3000)
	if len(chunks) != 4 {
		t.Fatalf("expect 4 hard-split chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if utf8.RuneCountInString(c) > 3000 {
			t.Fatalf("chunk %d exceeds maxRunes: %d", i, utf8.RuneCountInString(c))
		}
	}
	if strings.Join(chunks, "") != content {
		t.Fatal("hard split lost content")
	}
}

// "C#1" 之类不是标题，不能作为分节边界
func TestIsMarkdownHeading(t *testing.T) {
	cases := map[string]bool{
		"# 标题":      true,
		"##  Sub":    true, // 多空格也算
		"###### h6":  true,
		"####### h7": false, // 7 个 # 不是合法标题
		"C#1 不是标题":   false,
		"#nospace":   false,
		"正文 # 不是行首":  false,
	}
	for line, want := range cases {
		if got := isMarkdownHeading(line); got != want {
			t.Errorf("isMarkdownHeading(%q) = %v, want %v", line, got, want)
		}
	}
}
