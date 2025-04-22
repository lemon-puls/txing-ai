import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {UserRoleEnum} from "@/enume"

export const useUserStore = defineStore('user', () => {
  // 状态
  const userInfo = ref(null)

  // 计算属性
  const isLoggedIn = computed(() => !!userInfo.value)
  const isAdmin = computed(() => userInfo.value?.role === UserRoleEnum.ADMIN)
  const userId = computed(() => userInfo.value?.id)
  const username = computed(() => userInfo.value?.username)
  const avatar = computed(() => userInfo.value?.avatar || '')

  // 方法
  const setUserInfo = (info) => {
    userInfo.value = info
  }

  const clearUserInfo = () => {
    userInfo.value = null
  }

  // Token 相关方法
  const getToken = () => localStorage.getItem('token')
  const getRefreshToken = () => localStorage.getItem('refreshToken')

  const setToken = (token, refreshToken) => {
    localStorage.setItem('token', token)
    localStorage.setItem('refreshToken', refreshToken)
  }

  const clearToken = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  // 登录
  const login = (data) => {
    setUserInfo({
      id: data.id,
      username: data.username,
      email: data.email,
      phone: data.phone,
      gender: data.gender,
      status: data.status,
      role: data.role,
      age: data.age,
      avatar: data.avatar,
      createdAt: data.createdAt,
      updatedAt: data.updatedAt
    })
    setToken(data.token, data.refreshToken)
  }

  // 登出
  const logout = () => {
    clearUserInfo()
    clearToken()
  }

  return {
    userInfo,
    isLoggedIn,
    isAdmin,
    userId,
    username,
    avatar,
    login,
    logout,
    setUserInfo,
    clearUserInfo,
    getToken,
    getRefreshToken,
    setToken,
    clearToken
  }
})
