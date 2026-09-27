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

// PresetModelTags 模型内定标签，模型提案优先从中匹配，展示时置前排。
// 前端镜像见 static/frontend/src/views/admin/model/ModelList.vue 的 modelTags，两处需同步维护。
var PresetModelTags = []string{
	"通用",
	"联网搜索",
	"深度思考",
	"编程强化",
}

// reorderTagsFirst 将标签稳定重排：preset 中出现的标签按 preset 顺序置前，其余标签保持原相对顺序排后
func reorderTagsFirst(tags, preset []string) []string {
	rank := func(t string) int {
		for i, p := range preset {
			if p == t {
				return i
			}
		}
		return len(preset)
	}
	ordered := make([]string, len(tags))
	copy(ordered, tags)
	sort.SliceStable(ordered, func(i, j int) bool {
		return rank(ordered[i]) < rank(ordered[j])
	})
	return ordered
}

// reorderTagsPresetFirst 网站录入标签重排（内定分类标签置前）
func reorderTagsPresetFirst(tags []string) []string {
	return reorderTagsFirst(tags, PresetWebsiteTags)
}
