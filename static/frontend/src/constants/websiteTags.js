// 内定标签（起分类作用，展示时突出置前）
// 与后端 internal/tool/ops/tags.go 的 PresetWebsiteTags 保持同步，两处需一起维护
export const PRESET_WEBSITE_TAGS = ['实用工具', 'AI', 'Skill', '开源项目', '开发框架', '开发工具', '文档教程', '设计资源']

const PRESET_SET = new Set(PRESET_WEBSITE_TAGS)

// 是否内定标签
export const isPresetTag = (tag) => PRESET_SET.has(tag)

// 逗号分隔的 tags 字符串 → 数组（去空白、去空项）
export const splitTags = (tags) => (tags || '').split(',').map((t) => t.trim()).filter(Boolean)

// 内定标签在前（按内定顺序），其余标签按原相对顺序排后
export const orderTagsPresetFirst = (tags) => {
  const preset = PRESET_WEBSITE_TAGS.filter((t) => tags.includes(t))
  const custom = tags.filter((t) => !PRESET_SET.has(t))
  return [...preset, ...custom]
}
