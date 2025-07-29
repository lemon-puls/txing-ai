/**
 * 从环境变量加载分析脚本
 * // 从环境变量加载分析脚本
 */
export function loadAnalyticsScript() {
  const scriptUrl = import.meta.env.VITE_ANALYTICS_SCRIPT_URL
  const websiteId = import.meta.env.VITE_ANALYTICS_WEBSITE_ID
  
  if (!scriptUrl || !websiteId) {
    console.warn('分析脚本配置不完整，跳过加载')
    return
  }
  
  // 创建脚本元素
  const script = document.createElement('script')
  script.defer = true
  script.src = scriptUrl
  script.setAttribute('data-website-id', websiteId)
  
  // 添加到文档头部
  document.head.appendChild(script)
} 