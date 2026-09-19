package channel

import (
	"errors"
	"strings"
	"testing"

	"txing-ai/internal/domain"
	"txing-ai/internal/global"
)

// 构造一个仅当 enableWeb=false 时才匹配的 channel 映射
func newWebOnlyChannel() domain.Channel {
	return domain.Channel{
		Mappings: []global.ModelMapping{
			{
				SourceModel: "deepseek-v3-250324",
				Conditions: []global.ModelMappingCondition{
					{
						TargetModel: "deepseek-v3-offline",
						Conditions:  map[string]interface{}{"enableWeb": false},
					},
				},
			},
		},
	}
}

// TestChooseFromFilteredEmpty 回归测试：
// 过滤后没有可用的 channel 时，应返回错误而不是触发 rand.Intn(0) panic
// （panic 信息为 "invalid argument to Intn"）
func TestChooseFromFilteredEmpty(t *testing.T) {
	_, _, err := chooseFromFiltered([]domain.Channel{}, "deepseek-v3-250324", map[string]interface{}{"enableWeb": false})
	if err == nil {
		t.Fatal("expected error when filtered sequence is empty, got nil")
	}
	if !strings.Contains(err.Error(), "no available channel") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if !errors.Is(err, ErrNoAvailableChannel) {
		t.Fatalf("expected error to wrap ErrNoAvailableChannel, got: %v", err)
	}
}

// TestChooseFromFilteredMappingNotMatched 模拟真实场景：
// 渠道支持该模型，但当前请求参数（enableWeb=true）不满足所有映射条件，
// 过滤后列表为空，chooseFromFiltered 应返回错误而不是 panic
func TestChooseFromFilteredMappingNotMatched(t *testing.T) {
	ch := newWebOnlyChannel()

	// enableWeb=true 时映射不匹配，返回空字符串
	if m := ch.GetMappingModel("deepseek-v3-250324", map[string]interface{}{"enableWeb": true}); m != "" {
		t.Fatalf("expected mapping to be filtered out, got %q", m)
	}

	filtered := make([]domain.Channel, 0)
	_, _, err := chooseFromFiltered(filtered, "deepseek-v3-250324", map[string]interface{}{"enableWeb": true})
	if err == nil {
		t.Fatal("expected error when no channel matches mapping conditions, got nil")
	}
}

// TestChooseFromFilteredMatched 验证正常路径：
// 映射条件满足时能选中 channel 并返回映射后的模型
func TestChooseFromFilteredMatched(t *testing.T) {
	ch := newWebOnlyChannel()

	filtered := []domain.Channel{ch}
	got, mappingModel, err := chooseFromFiltered(filtered, "deepseek-v3-250324", map[string]interface{}{"enableWeb": false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetId() != ch.GetId() {
		t.Fatalf("expected channel id %d, got %d", ch.GetId(), got.GetId())
	}
	if mappingModel != "deepseek-v3-offline" {
		t.Fatalf("expected mapping model %q, got %q", "deepseek-v3-offline", mappingModel)
	}
}
