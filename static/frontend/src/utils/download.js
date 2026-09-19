import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import { getAuthHeaders, refreshToken } from '@/api/auth'

/**
 * 携带认证信息下载文件；token 过期（401）时自动刷新并重试一次，
 * 避免过期后点击下载"没有反应"
 * @param {string} url 文件下载地址（相对或绝对）
 * @param {string} filename 保存的文件名
 * @returns {Promise<boolean>} 是否下载成功
 */
export async function downloadFileWithAuth(url, filename) {
  try {
    let response = await fetch(url, { headers: getAuthHeaders() })

    // token 过期：刷新令牌后重试下载
    if (response.status === 401) {
      await refreshToken(defaultApi.apiUserRefreshPost)
      response = await fetch(url, { headers: getAuthHeaders() })
    }

    if (!response.ok) {
      ElMessage.error(response.status === 401 ? '登录已过期，请重新登录' : `文件下载失败（${response.status}）`)
      return false
    }

    // 后端错误可能仍以 200 返回（utils.ErrorWithMsg 固定 200 + JSON 包装），
    // 检查响应类型：application/json 说明是错误信息而非文件，不应保存成 1KB 假文件
    const contentType = response.headers.get('content-type') || ''
    if (contentType.includes('application/json')) {
      const errBody = await response.json().catch(() => null)
      ElMessage.error(errBody?.msg || '文件下载失败')
      return false
    }

    const blob = await response.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
    return true
  } catch (error) {
    // refreshToken 失败时会自动登出并清空本地凭证
    console.error('下载失败:', error)
    ElMessage.error('登录已过期，请重新登录')
    return false
  }
}
