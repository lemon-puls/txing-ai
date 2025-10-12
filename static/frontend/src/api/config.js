import ApiClient from './generated/src/ApiClient.js'
import { defaultApi } from "@/api/index.js"
import authService from './auth.js'

const apiClient = new ApiClient()

// 根据环境设置基础路径
// if (import.meta.env.MODE === 'development') {
//   // 开发环境，使用代理
//   apiClient.basePath = ''
// } else {
//   // 生产环境，使用相对路径
//   apiClient.basePath = ''
// }

apiClient.basePath = ''

// 设置默认请求头
apiClient.defaultHeaders = {
  'Content-Type': 'application/json'
}

// 重写callApi方法来处理响应
const originalCallApi = apiClient.callApi
apiClient.callApi = async function (...args) {
  const [path] = args

  // 如果不是刷新token的请求，添加token到请求头
  if (!path.includes('/user/refresh')) {
    const authHeaders = authService.getAuthHeaders()
    if (authHeaders.Authorization) {
      this.defaultHeaders['Authorization'] = authHeaders.Authorization
    }
  }

  try {
    const response = await originalCallApi.apply(this, args)
    return response
  } catch (error) {
    // 使用公共模块处理401错误
    return await authService.handle401Error(
      error, 
      path, 
      defaultApi.apiUserRefreshPost,
      (newToken) => {
        this.defaultHeaders['Authorization'] = `Bearer ${newToken}`
        return originalCallApi.apply(this, args)
      }
    )
  }
}

export default apiClient
