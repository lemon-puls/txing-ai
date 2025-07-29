/**
 * 从环境变量加载分析脚本
 * // 从环境变量加载分析脚本
 */
export function loadAnalyticsScript() {
  const scriptUrl = import.meta.env.VITE_ANALYTICS_SCRIPT_URL
  const websiteId = import.meta.env.VITE_ANALYTICS_WEBSITE_ID
  
  // 开发环境下，如果没有配置分析脚本，只显示警告
  if (!scriptUrl || !websiteId) {
    if (import.meta.env.DEV) {
      console.warn('分析脚本配置不完整，跳过加载。请确保在 .env 文件中设置了 VITE_ANALYTICS_SCRIPT_URL 和 VITE_ANALYTICS_WEBSITE_ID')
    }
    return
  }
  
  try {
    // 创建脚本元素
    const script = document.createElement('script')
    script.defer = true
    script.src = scriptUrl
    script.setAttribute('data-website-id', websiteId)
    
    // 添加到文档头部
    document.head.appendChild(script)
    
    if (import.meta.env.DEV) {
      console.log('分析脚本已加载:', scriptUrl)
    }
  } catch (error) {
    console.error('加载分析脚本时出错:', error)
  }
} 