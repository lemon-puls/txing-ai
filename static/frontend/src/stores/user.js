import { defineStore } from 'pinia'
import { UserRoleEnum } from "@/enume"

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: null,
    token: null,
    refreshToken: null
  }),
  getters: {
    isLoggedIn: (state) => !!state.userInfo,
    isAdmin: (state) => state.userInfo?.role === UserRoleEnum.ADMIN,
    userId: (state) => state.userInfo?.id,
    username: (state) => state.userInfo?.username,
    avatar: (state) => state.userInfo?.avatar || ''
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
    },
    logout() {
      console.log('logout')
      this.clearUserInfo()
      this.clearToken()
    }
  },
  persist: {
    key: 'user-store',
    storage: localStorage,
    paths: ['token', 'refreshToken', 'userInfo'],
  }
})
