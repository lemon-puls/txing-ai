// 通过 key 字符串动态获取 Element Plus 图标组件
// Get Element Plus icon component by key string
// 用于后端配置存 iconKey（如 "Platform"、"Cpu"），前端按 key 解析为组件

import * as ElIcons from '@element-plus/icons-vue'

const iconMap = {
  ...ElIcons,
  // 自定义图标（about 页面用到的 SVG）
  Github: {
    template: `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg>`
  }
}

/**
 * 通过 key 字符串解析图标组件
 * @param {string} key 图标 key（如 "Cpu"、"Platform"）
 * @returns 图标组件，若未找到则返回 null
 */
export function resolveIcon(key) {
  if (!key) return null
  return iconMap[key] || null
}

/**
 * 可供后台配置时选择的图标 key 列表
 * @returns Array<{key: string, label: string}>
 */
export const AVAILABLE_ICONS = [
  { key: 'Cpu', label: 'Cpu' },
  { key: 'Monitor', label: 'Monitor' },
  { key: 'Promotion', label: 'Promotion' },
  { key: 'Platform', label: 'Platform' },
  { key: 'Document', label: 'Document' },
  { key: 'ChatDotRound', label: 'ChatDotRound' },
  { key: 'DataLine', label: 'DataLine' },
  { key: 'ArrowRight', label: 'ArrowRight' },
  { key: 'ArrowDown', label: 'ArrowDown' },
  { key: 'CircleCheck', label: 'CircleCheck' },
  { key: 'Close', label: 'Close' },
  { key: 'Message', label: 'Message' },
  { key: 'Github', label: 'Github' },
  { key: 'Star', label: 'Star' },
  { key: 'TrophyBase', label: 'TrophyBase' },
  { key: 'MagicStick', label: 'MagicStick' },
  { key: 'DataAnalysis', label: 'DataAnalysis' },
  { key: 'Aim', label: 'Aim' },
  { key: 'TrendCharts', label: 'TrendCharts' },
  { key: 'Connection', label: 'Connection' },
  { key: 'Box', label: 'Box' },
  { key: 'Sunny', label: 'Sunny' },
  { key: 'Moon', label: 'Moon' },
  { key: 'Bell', label: 'Bell' },
  { key: 'Brush', label: 'Brush' },
  { key: 'Camera', label: 'Camera' },
  { key: 'Cellphone', label: 'Cellphone' },
  { key: 'ChatDotSquare', label: 'ChatDotSquare' },
  { key: 'Coin', label: 'Coin' },
  { key: 'Compass', label: 'Compass' },
  { key: 'CreditCard', label: 'CreditCard' },
  { key: 'Discount', label: 'Discount' },
  { key: 'Eleme', label: 'Eleme' },
  { key: 'ElemeFilled', label: 'ElemeFilled' },
  { key: 'Failed', label: 'Failed' },
  { key: 'Goods', label: 'Goods' },
  { key: 'GoodsFilled', label: 'GoodsFilled' },
  { key: 'Histogram', label: 'Histogram' },
  { key: 'HomeFilled', label: 'HomeFilled' },
  { key: 'Iphone', label: 'Iphone' },
  { key: 'Key', label: 'Key' },
  { key: 'Lightning', label: 'Lightning' },
  { key: 'Link', label: 'Link' },
  { key: 'Lock', label: 'Lock' },
  { key: 'MapLocation', label: 'MapLocation' },
  { key: 'Microphone', label: 'Microphone' },
  { key: 'Mute', label: 'Mute' },
  { key: 'Notification', label: 'Notification' },
  { key: 'Operation', label: 'Operation' },
  { key: 'Phone', label: 'Phone' },
  { key: 'Picture', label: 'Picture' },
  { key: 'PictureFilled', label: 'PictureFilled' },
  { key: 'PieChart', label: 'PieChart' },
  { key: 'Position', label: 'Position' },
  { key: 'Printer', label: 'Printer' },
  { key: 'Reading', label: 'Reading' },
  { key: 'Refresh', label: 'Refresh' },
  { key: 'School', label: 'School' },
  { key: 'Scissor', label: 'Scissor' },
  { key: 'Search', label: 'Search' },
  { key: 'Select', label: 'Select' },
  { key: 'Sell', label: 'Sell' },
  { key: 'Service', label: 'Service' },
  { key: 'Setting', label: 'Setting' },
  { key: 'Share', label: 'Share' },
  { key: 'Ship', label: 'Ship' },
  { key: 'Shop', label: 'Shop' },
  { key: 'ShoppingBag', label: 'ShoppingBag' },
  { key: 'ShoppingCart', label: 'ShoppingCart' },
  { key: 'Smoking', label: 'Smoking' },
  { key: 'Soccer', label: 'Soccer' },
  { key: 'Stamp', label: 'Stamp' },
  { key: 'Stopwatch', label: 'Stopwatch' },
  { key: 'SuccessFilled', label: 'SuccessFilled' },
  { key: 'SwitchButton', label: 'SwitchButton' },
  { key: 'TakeawayBox', label: 'TakeawayBox' },
  { key: 'Tickets', label: 'Tickets' },
  { key: 'Timer', label: 'Timer' },
  { key: 'Tools', label: 'Tools' },
  { key: 'Top', label: 'Top' },
  { key: 'TopRight', label: 'TopRight' },
  { key: 'TurnOff', label: 'TurnOff' },
  { key: 'Umbrella', label: 'Umbrella' },
  { key: 'Unlock', label: 'Unlock' },
  { key: 'Upload', label: 'Upload' },
  { key: 'UploadFilled', label: 'UploadFilled' },
  { key: 'User', label: 'User' },
  { key: 'UserFilled', label: 'UserFilled' },
  { key: 'Van', label: 'Van' },
  { key: 'VideoCamera', label: 'VideoCamera' },
  { key: 'VideoCameraFilled', label: 'VideoCameraFilled' },
  { key: 'VideoPause', label: 'VideoPause' },
  { key: 'VideoPlay', label: 'VideoPlay' },
  { key: 'View', label: 'View' },
  { key: 'Wallet', label: 'Wallet' },
  { key: 'WalletFilled', label: 'WalletFilled' },
  { key: 'Warning', label: 'Warning' },
  { key: 'WarningFilled', label: 'WarningFilled' },
  { key: 'Watch', label: 'Watch' },
  { key: 'Watermelon', label: 'Watermelon' },
  { key: 'ZoomIn', label: 'ZoomIn' },
  { key: 'ZoomOut', label: 'ZoomOut' }
]