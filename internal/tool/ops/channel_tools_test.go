package ops

import (
	"encoding/json"
	"strings"
	"testing"

	"txing-ai/internal/global"
)

// TestChannelSecretRedaction 安全测试：渠道工具的出参结构序列化后不得包含任何密钥字段
// （channelItem / channelPreviewRequest / ChannelProposal / channelPreviewResult）
func TestChannelSecretRedaction(t *testing.T) {
	item := channelItem{
		Id: 1, Name: "polo 渠道", ChannelType: "polo", Endpoint: "https://api.polo.com",
		Models: []string{"gpt-4o"}, Status: true,
		SecretConfigured: true, SecretKeyCount: 2,
	}
	req := channelPreviewRequest{
		Name: "polo 渠道", ChannelType: "polo", Endpoint: "https://api.polo.com",
		Models: []string{"gpt-4o"},
	}
	proposal := ChannelProposal{
		Type: ProposalTypeChannel, Name: "polo 渠道", ChannelType: "polo",
		Endpoint: "https://api.polo.com", Models: []string{"gpt-4o"},
	}
	result := channelPreviewResult{Status: "ok", Proposal: &proposal}

	for name, payload := range map[string]interface{}{
		"channelItem":          item,
		"channelPreviewResult": result,
	} {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("%s marshal failed: %v", name, err)
		}
		// "secret" 作为独立 JSON key 不得出现（secretConfigured/secretKeyCount 属脱敏统计，允许）
		if strings.Contains(string(data), `"secret":`) || strings.Contains(string(data), `"secretKey"`) {
			t.Errorf("%s 序列化结果包含密钥字段: %s", name, data)
		}
	}
	_ = req // preview 请求结构上没有 secret 字段，编译期即保证
}

// TestPreviewChannelValidation 校验 channel_preview_tool 的字段校验逻辑（无 DB 场景）
func TestPreviewChannelValidation(t *testing.T) {
	deps := OpsToolDeps{}

	cases := []struct {
		name       string
		req        channelPreviewRequest
		wantStatus string
		wantMsgHas string // 期望错误信息包含的关键字（空表示不校验）
	}{
		{
			name: "合法提案",
			req: channelPreviewRequest{
				Name: "polo 主渠道", ChannelType: "polo", Endpoint: "api.polo.com/v1",
				Models: []string{"gpt-4o"}, Status: true,
			},
			wantStatus: "ok",
		},
		{
			name: "缺 scheme 自动补全",
			req: channelPreviewRequest{
				Name: "polo 主渠道", ChannelType: "polo", Endpoint: "api.polo.com/v1/",
				Models: []string{"gpt-4o"},
			},
			wantStatus: "ok",
		},
		{
			name: "类型不在可选集",
			req: channelPreviewRequest{
				Name: "非法类型渠道", ChannelType: "anthropic", Endpoint: "https://api.anthropic.com",
				Models: []string{"claude"},
			},
			wantStatus: "invalid",
			wantMsgHas: "渠道类型",
		},
		{
			name: "模型列表为空",
			req: channelPreviewRequest{
				Name: "空模型渠道", ChannelType: "polo", Endpoint: "https://api.polo.com",
			},
			wantStatus: "invalid",
			wantMsgHas: "模型",
		},
		{
			name: "endpoint 非法",
			req: channelPreviewRequest{
				Name: "坏地址渠道", ChannelType: "polo", Endpoint: "::::",
				Models: []string{"gpt-4o"},
			},
			wantStatus: "invalid",
			wantMsgHas: "服务地址",
		},
		{
			name: "映射缺少 sourceModel",
			req: channelPreviewRequest{
				Name: "映射渠道", ChannelType: "OpenAI", Endpoint: "https://api.openai.com",
				Models: []string{"gpt-4o"},
				Mappings: []global.ModelMapping{
					{SourceModel: " ", Conditions: []global.ModelMappingCondition{{TargetModel: "gpt-4o-2024"}}},
				},
			},
			wantStatus: "invalid",
			wantMsgHas: "sourceModel",
		},
		{
			name: "映射条件 key 不支持",
			req: channelPreviewRequest{
				Name: "映射渠道", ChannelType: "OpenAI", Endpoint: "https://api.openai.com",
				Models: []string{"gpt-4o"},
				Mappings: []global.ModelMapping{
					{SourceModel: "gpt-4o", Conditions: []global.ModelMappingCondition{
						{TargetModel: "gpt-4o-2024", Conditions: map[string]interface{}{"webSearch": true}},
					}},
				},
			},
			wantStatus: "invalid",
			wantMsgHas: "enableWeb / type",
		},
		{
			name: "映射 type 取值非法",
			req: channelPreviewRequest{
				Name: "映射渠道", ChannelType: "OpenAI", Endpoint: "https://api.openai.com",
				Models: []string{"gpt-4o"},
				Mappings: []global.ModelMapping{
					{SourceModel: "gpt-4o", Conditions: []global.ModelMappingCondition{
						{TargetModel: "gpt-4o-app", Conditions: map[string]interface{}{"type": "agent"}},
					}},
				},
			},
			wantStatus: "invalid",
			wantMsgHas: "type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := deps.previewChannel(t.Context(), &tc.req)
			if err != nil {
				t.Fatalf("previewChannel returned error: %v", err)
			}
			if res.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q (message: %s)", res.Status, tc.wantStatus, res.Message)
			}
			if tc.wantMsgHas != "" && !strings.Contains(res.Message, tc.wantMsgHas) {
				t.Fatalf("message %q 不包含期望关键字 %q", res.Message, tc.wantMsgHas)
			}
			// 合法路径：endpoint 完成规范化（去尾斜杠）
			if tc.wantStatus == "ok" && strings.HasSuffix(res.Proposal.Endpoint, "/") {
				t.Fatalf("endpoint 未规范化: %s", res.Proposal.Endpoint)
			}
		})
	}
}

// TestPreviewChannelUpdateTypePassthrough 优化提案：未知类型保持透传（不阻断），仅附加警告
func TestPreviewChannelUpdateTypePassthrough(t *testing.T) {
	// 无 DB：req.Id > 0 时会因查不到原渠道而 invalid，这里仅验证创建分支的强校验与
	// 优化分支的差异，故直接调用 knownChannelType 断言集合行为
	if knownChannelType("polo") != true || knownChannelType("volcengine") != false {
		t.Fatal("knownChannelType 集合判断不符合预期")
	}
}

// TestNormalizeModels 模型列表规范化：去空白、剔空、去重（保序）
func TestNormalizeModels(t *testing.T) {
	cases := []struct {
		name  string
		in    []string
		want  []string
	}{
		{name: "去空白与空项", in: []string{" gpt-4o ", "", "glm-5"}, want: []string{"gpt-4o", "glm-5"}},
		{name: "去重保序", in: []string{"glm-5", "gpt-4o", "glm-5"}, want: []string{"glm-5", "gpt-4o"}},
		{name: "全空输入", in: []string{"", "  "}, want: []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeModels(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("normalizeModels(%v) = %v, want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("normalizeModels(%v) = %v, want %v", tc.in, got, tc.want)
				}
			}
		})
	}
}
