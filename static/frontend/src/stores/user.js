import { defineStore } from 'pinia'
import { UserRoleEnum } from "@/enume"
import { useConversationStore } from './conversation'

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: null,
    token: localStorage.getItem('token') || '',
    refreshToken: localStorage.getItem('refreshToken') || '',
  }),
  getters: {
    isLoggedIn: (state) => !!state.token && !!state.userInfo,
    isAdmin: (state) => state.userInfo?.role === UserRoleEnum.ADMIN,
    userId: (state) => state.userInfo?.id || null,
    username: (state) => state.userInfo?.username,
    avatar: (state) => state.userInfo?.avatar || '',
    userRoles: (state) => state.userInfo?.role === UserRoleEnum.ADMIN ? ['admin'] : [],
  },
  actions: {
    setUserInfo(info) {
      this.userInfo = info
    },
    clearUserInfo() {
      this.userInfo = null
    },
    getToken() {
      return this.token
    },
    getRefreshToken() {
      return this.refreshToken
    },
    setToken(token, refreshToken) {
      this.token = token
      this.refreshToken = refreshToken
      localStorage.setItem('token', token)
      localStorage.setItem('refreshToken', refreshToken)
    },
    clearToken() {
      this.token = null
      this.refreshToken = null
      // localStorage.removeItem('user-store')
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
    },
    login(data) {
      this.setUserInfo({
        id: data.id,
        username: data.username,
        role: data.role,
        avatar: data.avatar,
        createdAt: data.createdAt,
        updatedAt: data.updatedAt
      })
      this.setToken(data.token, data.refreshToken)

      // 登录后清除本地会话数据，并加载服务器端会话数据
      this.syncConversationsAfterLogin()
    },
    logout() {
      this.clearUserInfo()
      this.clearToken()

      // 登出后清除会话数据，重新加载本地会话数据
      this.syncConversationsAfterLogout()
    },

    // 同步会话数据（登录后）
    syncConversationsAfterLogin() {
      const conversationStore = useConversationStore()

      // 获取当前本地会话
      const localConversations = JSON.parse(localStorage.getItem('conversations') || '[]')

      // 如果有本地会话，先保存到服务器（可以选择性实现这个功能）
      // 这里简化处理，登录后直接清空本地会话，加载服务器端会话

      // 清空会话存储
      conversationStore.clearConversations()

      // 加载服务器端会话
      conversationStore.loadConversations()
    },

    // 同步会话数据（登出后）
    syncConversationsAfterLogout() {
      const conversationStore = useConversationStore()

      // 清空会话存储
      conversationStore.clearConversations()

      // 加载本地会话
      conversationStore.loadConversations()
    }
  },
  persist: {
    key: 'user-store',
    storage: localStorage,
    paths: ['token', 'refreshToken', 'userInfo'],
  }
})
