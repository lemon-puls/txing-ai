import apiClient from './config'

// LLM Wiki 知识库 API 封装
// 后端尚未生成 OpenAPI 客户端，这里手工包装（与 generated 客户端共用同一 apiClient，
// 自动附带 Authorization 与 401 刷新逻辑）

const call = async (path, method, { query, body } = {}) => {
  // returnType 必须传 Object：ApiClient.deserialize 对 null 直接返回 null；
  // 传 Object 时 convertToType 原样返回解析后的响应体
  const res = await apiClient.callApi(
    path, method,
    {}, query || {}, {}, {}, body || null,
    [], ['application/json'], ['application/json'], Object, null
  )
  // 兼容两种 callApi 返回形态：{data, response} 包装（当前生成版本）或直接返回响应体
  if (res && typeof res === 'object' && 'data' in res && 'response' in res) {
    return res.data ?? {}
  }
  return res
}

export default {
  // --- 源管理 ---
  createMDSource: (title, content) =>
    call('/api/admin/wiki/sources/md', 'POST', { body: { title, content } }),
  createURLSource: (title, url) =>
    call('/api/admin/wiki/sources/url', 'POST', { body: { title, url } }),
  listSources: (page = 1, pageSize = 20) =>
    call('/api/admin/wiki/sources/list', 'GET', { query: { page, pageSize } }),
  deleteSource: (id) =>
    call(`/api/admin/wiki/sources/${id}`, 'DELETE'),
  triggerIngest: (id) =>
    call(`/api/admin/wiki/sources/${id}/ingest`, 'POST'),

  // --- 草稿审核 ---
  listDrafts: (page = 1, pageSize = 50, keyword = '', pageType = '') =>
    call('/api/admin/wiki/drafts/list', 'GET', {
      query: { page, pageSize, keyword: keyword || undefined, pageType: pageType || undefined }
    }),
  updateDraft: (id, data) =>
    call(`/api/admin/wiki/drafts/${id}`, 'PUT', { body: data }),
  confirmDraft: (id) =>
    call(`/api/admin/wiki/drafts/${id}/confirm`, 'POST'),
  confirmAllDrafts: () =>
    call('/api/admin/wiki/drafts/confirm_all', 'POST'),
  deleteDraft: (id) =>
    call(`/api/admin/wiki/drafts/${id}`, 'DELETE'),

  // --- 已发布页管理 ---
  listPublished: (page = 1, pageSize = 20, keyword = '', pageType = '') =>
    call('/api/admin/wiki/pages/list', 'GET', {
      query: { page, pageSize, keyword: keyword || undefined, pageType: pageType || undefined }
    }),
  getPublished: (id) =>
    call(`/api/admin/wiki/pages/${id}`, 'GET'),
  updatePublished: (id, data) =>
    call(`/api/admin/wiki/pages/${id}`, 'PUT', { body: data }),
  offlinePage: (id) =>
    call(`/api/admin/wiki/pages/${id}/offline`, 'POST'),

  // --- 知识图谱 ---
  graph: () =>
    call('/api/admin/wiki/graph', 'GET'),

  // --- 导出（带鉴权头拉取 zip 二进制） ---
  async exportAll() {
    const authHeaders = (await import('./auth')).default.getAuthHeaders()
    const response = await fetch('/api/admin/wiki/export', { headers: authHeaders })
    if (!response.ok) throw new Error(`导出失败（HTTP ${response.status}）`)
    const blob = await response.blob()
    const disposition = response.headers.get('Content-Disposition') || ''
    const match = disposition.match(/filename="([^"]+)"/)
    return { blob, filename: match ? match[1] : 'wiki_export.zip' }
  }
}
