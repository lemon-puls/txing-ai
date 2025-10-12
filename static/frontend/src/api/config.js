import ApiClient from './generated/src/ApiClient.js'
import {ElMessage} from 'element-plus'
import {defaultApi} from "@/api/index.js";

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

// 标记是否正在刷新token
let isRefreshing = false
// 存储等待刷新token的请求
let requests = []


// 重写callApi方法来处理响应
const originalCallApi = apiClient.callApi
apiClient.callApi = async function (...args) {
  const [path] = args

  // 如果不是刷新token的请求，添加token到请求头
  if (!path.includes('/user/refresh')) {
    const token = localStorage.getItem('token')
    if (token) {
      this.defaultHeaders['Authorization'] = `Bearer ${token}`
    }
  }

  try {
    const response = await originalCallApi.apply(this, args)
    return response
  } catch (error) {
    // 处理401错误（未登录或token失效）
    if (error.status === 401) {
      const refreshToken = localStorage.getItem('refreshToken')

      // 如果有刷新令牌且不是刷新token的请求
      if (refreshToken && !path.includes('/user/refresh')) {
        if (!isRefreshing) {
          isRefreshing = true

          try {
            // 尝试刷新token，将 refreshToken 放在请求头中
            // const refreshResponse = await originalCallApi.call(
            //   this,
            //   '/user/refresh',
            //   'POST',
            //   {},
            //   {},
            //   {'Authorization': `Bearer ${refreshToken}`}, // 在请求头中传递 refreshToken
            //   {},
            //   {}, // 不需要在 body 中传递 refreshToken
            //   [],
            //   ['application/json'],
            //   ['application/json'],
            //   null
            // )'
            const refreshResponse = await defaultApi.apiUserRefreshPost(`Bearer ${refreshToken}`)
            // const refreshData = refreshResponse.response.body
            if (refreshResponse.code === 0) {
              // 更新token
              const {accessToken: newToken, refreshToken: newRefreshToken} = refreshResponse.data
              localStorage.setItem('token', newToken)
              localStorage.setItem('refreshToken', newRefreshToken)

              // 重新发送之前失败的请求
              requests.forEach(cb => cb(newToken))
              requests = []

              // 重试当前请求
              this.defaultHeaders['Authorization'] = `Bearer ${newToken}`
              return await originalCallApi.apply(this, args)
            } else {
              // 刷新失败，清除用户信息
              logout()
              ElMessage.error('请重新登录!')
              return Promise.reject(new Error('请重新登录'))
            }
          } catch (refreshError) {
            // 刷新失败，清除用户信息
            logout()
            return Promise.reject(refreshError)
          } finally {
            isRefreshing = false
          }
        } else {
          // 正在刷新token，将请求暂存
          return new Promise(resolve => {
            requests.push(token => {
              this.defaultHeaders['Authorization'] = `Bearer ${token}`
              resolve(originalCallApi.apply(this, args))
            })
          })
        }
      } else {
        // 没有刷新令牌，执行登出
        logout()
        return Promise.reject(new Error('请先登录'))
      }
    }

    // 其他错误处理
    ElMessage.error(error.message || '请求失败')
    return Promise.reject(error)
  }
}

// 退出登陆
// 由于在拦截中无法直接使用 userStore，因此这里直接清除用户信息实现退出登陆
function logout() {
  // 清除用户信息
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('user-store')
}

export default apiClient
