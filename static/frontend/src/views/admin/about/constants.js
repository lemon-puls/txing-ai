/**
 * 关于我页面配置 - 分区元信息与常量
 * About page admin - section metadata & constants
 */

// 分区元信息：锚点导航、scrollspy、统计 chips 共用
// Section metadata shared by anchor nav / scrollspy / stats chips
export const SECTIONS = [
  { id: 'section-hero', key: 'hero', label: 'Hero 顶部', icon: 'UserFilled' },
  { id: 'section-floating', key: 'floating', label: '浮动图标', icon: 'MagicStick' },
  { id: 'section-reasons', key: 'reasons', label: '为什么选择我', icon: 'Stamp' },
  { id: 'section-skills', key: 'skills', label: '核心能力', icon: 'Cpu' },
  { id: 'section-projects', key: 'projects', label: '精选作品', icon: 'FolderOpened' },
  { id: 'section-timeline', key: 'timeline', label: '成长轨迹', icon: 'Timer' },
  { id: 'section-contact', key: 'contact', label: '联系区', icon: 'ChatDotRound' }
]

export const SECTION_IDS = SECTIONS.map((s) => s.id)

// 项目封面渐变色板：必须与前台 .project-gradient-1..4 逐字一致（views/about/index.vue）
// Project gradient palette: must match front-end .project-gradient-1..4 exactly
export const PROJECT_GRADIENTS = [
  { value: 1, label: '紫蓝', from: '#667eea', to: '#764ba2' },
  { value: 2, label: '粉红', from: '#f093fb', to: '#f5576c' },
  { value: 3, label: '蓝青', from: '#4facfe', to: '#00f2fe' },
  { value: 4, label: '绿青', from: '#43e97b', to: '#38f9d7' }
]

// 历史数据可能存 5/6 或 number，归一化到 1-4
// Normalize legacy values (5/6 or number) into 1-4
export const normalizeGradient = (v) => {
  const n = Number(v)
  return PROJECT_GRADIENTS.some((g) => g.value === n) ? n : 1
}

// 前台预览地址：dev 下 BASE_URL='/'，生产为 '/dash/'（vite.config.js base）
// Front-end preview URL: works for both dev and prod base paths
export const previewUrl = `${window.location.origin}${import.meta.env.BASE_URL}#/about`
