package ops

import (
	"strings"
	"testing"
)

// TestPreviewWebsiteValidation 校验 website_preview_tool 的字段校验逻辑（无 DB 场景）
func TestPreviewWebsiteValidation(t *testing.T) {
	deps := OpsToolDeps{}

	cases := []struct {
		name       string
		req        websitePreviewRequest
		wantStatus string
		wantURL    string // 期望规范化后的 URL（空表示不校验）
		wantTags   string
	}{
		{
			name: "合法提案",
			req: websitePreviewRequest{
				Name: "GitHub", Description: "全球最大的代码托管平台",
				URL: "https://github.com", Tags: []string{"开发", "代码"},
			},
			wantStatus: "ok",
			wantURL:    "https://github.com",
			wantTags:   "开发,代码",
		},
		{
			name: "缺 scheme 自动补全并规范化路径",
			req: websitePreviewRequest{
				Name: "Gorm", Description: "Go 语言的 ORM 框架",
				URL: "gorm.io/docs/", Tags: []string{"Go"},
			},
			wantStatus: "ok",
			wantURL:    "https://gorm.io/docs",
			wantTags:   "Go",
		},
		{
			name: "名称过短",
			req: websitePreviewRequest{
				Name: "G", Description: "描述", URL: "https://example.com", Tags: []string{"标签"},
			},
			wantStatus: "invalid",
		},
		{
			name: "描述为空",
			req: websitePreviewRequest{
				Name: "Example", Description: "  ", URL: "https://example.com", Tags: []string{"标签"},
			},
			wantStatus: "invalid",
		},
		{
			name: "URL 非法",
			req: websitePreviewRequest{
				Name: "Example", Description: "描述", URL: "::::", Tags: []string{"标签"},
			},
			wantStatus: "invalid",
		},
		{
			name: "无标签",
			req: websitePreviewRequest{
				Name: "Example", Description: "描述", URL: "https://example.com", Tags: nil,
			},
			wantStatus: "invalid",
		},
		{
			name: "标签去重且最多保留 5 个",
			req: websitePreviewRequest{
				Name: "Example", Description: "描述", URL: "https://example.com",
				Tags: []string{"a", " a ", "b", "c", "d", "e", "f"},
			},
			wantStatus: "ok",
			wantTags:   "a,b,c,d,e",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := deps.previewWebsite(t.Context(), &tc.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status != tc.wantStatus {
				t.Fatalf("status = %q (%s), want %q", result.Status, result.Message, tc.wantStatus)
			}
			if tc.wantURL != "" && result.Proposal.Url != tc.wantURL {
				t.Fatalf("url = %q, want %q", result.Proposal.Url, tc.wantURL)
			}
			if tc.wantTags != "" && result.Proposal.Tags != tc.wantTags {
				t.Fatalf("tags = %q, want %q", result.Proposal.Tags, tc.wantTags)
			}
			if tc.wantStatus == "ok" && result.Proposal == nil {
				t.Fatal("proposal should not be nil when status is ok")
			}
			if tc.wantStatus == "invalid" && result.Proposal != nil {
				t.Fatal("proposal should be nil when status is invalid")
			}
		})
	}
}

// TestPreviewWebsiteProposalShape 校验提案字段与 dto.CreateWebsiteReq 的对应关系
func TestPreviewWebsiteProposalShape(t *testing.T) {
	deps := OpsToolDeps{}
	result, err := deps.previewWebsite(t.Context(), &websitePreviewRequest{
		Name:    "Eino",
		Description: strings.Repeat("描述", 50), // 恰好 100 字
		URL:     "https://www.cloudwego.io/zh/docs/eino/",
		Avatar:  "https://cdn.example.com/avatar.png",
		Tags:    []string{"Go", "LLM", "框架"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("status = %q, want ok", result.Status)
	}
	p := result.Proposal
	if p.Type != "website" {
		t.Fatalf("type = %q, want website", p.Type)
	}
	if p.Tags != "Go,LLM,框架" {
		t.Fatalf("tags = %q", p.Tags)
	}
	if p.Avatar != "https://cdn.example.com/avatar.png" {
		t.Fatalf("avatar = %q", p.Avatar)
	}
}

// TestNormalizeTags 标签清洗
func TestNormalizeTags(t *testing.T) {
	got := normalizeTags([]string{" 开发 ", "", "开发", "代码", "", " 开源 "})
	want := "开发,代码,开源"
	if strings.Join(got, ",") != want {
		t.Fatalf("normalizeTags = %v, want [%s]", got, want)
	}
}
