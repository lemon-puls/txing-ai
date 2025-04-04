import ApiClient from './generated/src/ApiClient.js'
import { ElMessage } from 'element-plus'

const apiClient = new ApiClient()

// 根据环境设置基础路径
if (import.meta.env.MODE === 'development') {
  // 开发环境，使用代理
  apiClient.basePath = '/api'
} else {
  // 生产环境，使用相对路径
  apiClient.basePath = ''
}

// 设置默认请求头
apiClient.defaultHeaders = {
  'Content-Type': 'application/json'
}

// 标记是否正在刷新token
let isRefreshing = false
// 存储等待刷新token的请求
let requests = []

// 重写callApi方法来处理响应
// const originalCallApi = apiClient.callApi
// apiClient.callApi = async function (...args) {
//   const [path] = args
//
//   // 如果不是刷新token的请求，添加token到请求头
//   if (!path.includes('/user/refreshtoken')) {
//     const token = localStorage.getItem('token')
//     if (token) {
//       this.defaultHeaders['Authorization'] = `Bearer ${token}`
//     }
//   }
//
//   try {
//     const response = await originalCallApi.apply(this, args)
//     const data = response.data
//     console.log("请求路径：", path, data, data.code)
//
//     return response
//   } catch (error) {
//     // 处理401错误（未登录或token失效）
//     if (error.status === 401) {
//       const refreshToken = localStorage.getItem('refreshToken')
//
//       // 如果有刷新令牌且不是刷新token的请求
//       if (refreshToken && !path.includes('/user/refreshtoken')) {
//         if (!isRefreshing) {
//           isRefreshing = true
//
//           try {
//             // 尝试刷新token，将 refreshToken 放在请求头中
//             const refreshResponse = await originalCallApi.call(
//               this,
//               '/api/user/refreshtoken',
//               'POST',
//               {},
//               {},
//               { 'Authorization': `Bearer ${refreshToken}` }, // 在请求头中传递 refreshToken
//               {},
//               {}, // 不需要在 body 中传递 refreshToken
//               [],
//               ['application/json'],
//               ['application/json'],
//               null
//             )
//             console.log("刷新token完成：", refreshResponse, refreshResponse.response.body)
//             const refreshData = refreshResponse.response.body
//             if (refreshData.code === 200) {
//               // 更新token
//               const { token: newToken, refreshToken: newRefreshToken } = refreshData.data
//               localStorage.setItem('token', newToken)
//               localStorage.setItem('refreshToken', newRefreshToken)
//
//               // 重新发送之前失败的请求
//               requests.forEach(cb => cb(newToken))
//               requests = []
//
//               // 重试当前请求
//               this.defaultHeaders['Authorization'] = `Bearer ${newToken}`
//               return await originalCallApi.apply(this, args)
//             } else {
//               // 刷新失败，清除用户信息并跳转到登录页
//               handleLogout()
//               return Promise.reject(new Error('请重新登录'))
//             }
//           } catch (refreshError) {
//             handleLogout()
//             return Promise.reject(refreshError)
//           } finally {
//             isRefreshing = false
//           }
//         } else {
//           // 正在刷新token，将请求暂存
//           return new Promise(resolve => {
//             requests.push(token => {
//               this.defaultHeaders['Authorization'] = `Bearer ${token}`
//               resolve(originalCallApi.apply(this, args))
//             })
//           })
//         }
//       } else {
//         // 没有刷新令牌，执行登出
//         handleLogout()
//         return Promise.reject(new Error('请先登录'))
//       }
//     }
//
//     // 其他错误处理
//     ElMessage.error(error.message || '请求失败')
//     return Promise.reject(error)
//   }
// }

// // 处理登出逻辑
// const handleLogout = () => {
//   localStorage.removeItem('token')
//   localStorage.removeItem('refreshToken')
//   localStorage.removeItem('user')
//
//   // 获取当前路由信息
//   const currentPath = window.location.hash.slice(1) // 去掉 # 号
//
//   // 跳转到登录页，并记录当前页面路径
//   if (currentPath !== '/login') {
//     window.location.href = `/#/login?redirect=${encodeURIComponent(currentPath)}`
//   }
// }

export default apiClient
