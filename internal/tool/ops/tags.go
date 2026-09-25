package ops

import "sort"

// PresetWebsiteTags 内定标签（起分类作用），录入提案优先从中匹配，展示时置前排。
// 前端镜像常量见 static/frontend/src/constants/websiteTags.js，两处需同步维护。
var PresetWebsiteTags = []string{
	"实用工具",
	"AI",
	"Skill",
	"开源项目",
	"开发框架",
	"开发工具",
	"文档教程",
	"设计资源",
}

// reorderTagsPresetFirst 将标签稳定重排：内定标签按内定顺序置前，其余标签保持原相对顺序排后
func reorderTagsPresetFirst(tags []string) []string {
	rank := func(t string) int {
		for i, p := range PresetWebsiteTags {
			if p == t {
				return i
			}
		}
		return len(PresetWebsiteTags)
	}
	ordered := make([]string, len(tags))
	copy(ordered, tags)
	sort.SliceStable(ordered, func(i, j int) bool {
		return rank(ordered[i]) < rank(ordered[j])
	})
	return ordered
}
