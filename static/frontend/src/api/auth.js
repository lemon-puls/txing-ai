import { ElMessage } from 'element-plus'
import {defaultApi} from "@/api/index.js";

// 标记是否正在刷新token
let isRefreshing = false
// 存储等待刷新token的请求
let requests = []

/**
 * 获取认证头信息
 * @returns {Object} 包含认证信息的请求头
 */
export function getAuthHeaders() {
  const token = localStorage.getItem('token')
  return {
    'Authorization': token ? `Bearer ${token}` : ''
  }
}

/**
 * 刷新token
 * @param {Function} refreshTokenApi - 刷新token的API函数
 * @returns {Promise<Object>} 包含新token的对象
 */
export async function refreshToken(refreshTokenApi) {
  const refreshToken = localStorage.getItem('refreshToken')
  if (!refreshToken) {
    logout()
    return Promise.reject(new Error('请先登录'))
  }

  if (!isRefreshing) {
    isRefreshing = true
    try {
      const refreshResponse = await refreshTokenApi(`Bearer ${refreshToken}`)

      if (refreshResponse.code === 0) {
        // 更新token
        const { accessToken: newToken, refreshToken: newRefreshToken } = refreshResponse.data
        localStorage.setItem('token', newToken)
        localStorage.setItem('refreshToken', newRefreshToken)

        // 重新发送之前失败的请求
        requests.forEach(cb => cb(newToken))
        requests = []

        return { newToken, newRefreshToken }
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
        resolve({ newToken: token })
      })
    })
  }
}

/**
 * 处理401错误
 * @param {Object} error - 错误对象
 * @param {string} path - 请求路径
 * @param {Function} refreshTokenApi - 刷新token的API函数
 * @param {Function} retryRequest - 重试请求的函数
 * @param {Function} onError - 可选的错误处理回调
 * @returns {Promise} 重试请求的结果
 */
export async function handle401Error(error, path, refreshTokenApi, retryRequest, onError) {
  if (error.status === 401) {
    const refreshTokenValue = localStorage.getItem('refreshToken')

    // 如果有刷新令牌且不是刷新token的请求
    if (refreshTokenValue && !path.includes('/user/refresh')) {
      try {
        // 直接在这里实现刷新token的逻辑，避免调用同名函数
        if (!isRefreshing) {
          isRefreshing = true
          try {
            const refreshResponse = await defaultApi.apiUserRefreshPost(`Bearer ${refreshTokenValue}`)

            if (refreshResponse.code === 0) {
              // 更新token
              const { accessToken: newToken, refreshToken: newRefreshToken } = refreshResponse.data
              localStorage.setItem('token', newToken)
              localStorage.setItem('refreshToken', newRefreshToken)

              // 重新发送之前失败的请求
              requests.forEach(cb => cb(newToken))
              requests = []

              // 重试当前请求
              return await retryRequest(newToken)
            } else {
              // 刷新失败，清除用户信息
              logout()
              ElMessage.error('请重新登录!')
              if (onError) onError(new Error('请重新登录'))
              return Promise.reject(new Error('请重新登录'))
            }
          } catch (refreshError) {
            // 刷新失败，清除用户信息
            logout()
            if (onError) onError(refreshError)
            return Promise.reject(refreshError)
          } finally {
            isRefreshing = false
          }
        } else {
          // 正在刷新token，将请求暂存
          return new Promise(resolve => {
            requests.push(token => {
              resolve(retryRequest(token))
            })
          })
        }
      } catch (err) {
        if (onError) onError(err)
        return Promise.reject(err)
      }
    } else {
      // 没有刷新令牌，执行登出
      logout()
      ElMessage.error('请先登录')
      if (onError) onError(new Error('请先登录'))
      return Promise.reject(new Error('请先登录'))
    }
  }

  // 其他错误处理
  ElMessage.error(error.message || '请求失败')
  if (onError) onError(error)
  return Promise.reject(error)
}


/**
 * 退出登录
 */
export function logout() {
  // 清除用户信息
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('user-store')
}

export default {
  getAuthHeaders,
  refreshToken,
  handle401Error,
  logout
}
