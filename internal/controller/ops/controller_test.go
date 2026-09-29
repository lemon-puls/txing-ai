package ops

import "testing"

// TestStripLatinOpening 历史消息英文开场白剥离矩阵：
// 整句英文开场剥除、中文开头的模型名保持原样、纯英文消息整条剥除
func TestStripLatinOpening(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "英文开场+中文正文：剥除英文前缀",
			in:   "I'll start by checking the existing models to avoid duplicates and match the naming style.两个模型提案均已生成，请确认。",
			want: "两个模型提案均已生成，请确认。",
		},
		{
			name: "英文开场（含撇号缩写）",
			in:   "Let me generate proposals for both models now.先查询已有条目。",
			want: "先查询已有条目。",
		},
		{
			name: "中文开头不动",
			in:   "我先查询现有模型，避免重复。",
			want: "我先查询现有模型，避免重复。",
		},
		{
			name: "模型名开头（英文词不足3个）不动",
			in:   "GLM-5 与 mimo-pro 是两个不同的模型。",
			want: "GLM-5 与 mimo-pro 是两个不同的模型。",
		},
		{
			name: "两个英文词开头不动（未达阈值）",
			in:   "mimo flash pro 是小米的模型系列。",
			want: "mimo flash pro 是小米的模型系列。",
		},
		{
			name: "纯英文消息整条剥除",
			in:   "I will check the existing models first.",
			want: "",
		},
		{
			name: "无汉字且非英文句子保持原样",
			in:   "GLM-5 / mimo-v2.6",
			want: "GLM-5 / mimo-v2.6",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripLatinOpening(c.in); got != c.want {
				t.Errorf("stripLatinOpening(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
