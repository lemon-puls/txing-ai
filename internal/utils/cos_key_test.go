package utils

import "testing"

// COSKeyFromURL 单元测试：从完整 URL 提取 key / 非 URL 原样返回
func TestCOSKeyFromURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://static.ai.txing.vip/media/1789921673288-34675.png?q-sign-algorithm=sha1&q-ak=AKID", "media/1789921673288-34675.png"},
		{"https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com/a/b/c.mp4?sign=1", "a/b/c.mp4"},
		{"http://example.com/key%20with%20space.png", "key%20with%20space.png"},
		{"media/1789921673288-34675.png", "media/1789921673288-34675.png"}, // 已是 key，原样返回
		{"", ""},
	}
	for _, c := range cases {
		if got := COSKeyFromURL(c.in); got != c.want {
			t.Errorf("COSKeyFromURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
