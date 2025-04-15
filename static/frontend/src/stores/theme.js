import { defineStore } from 'pinia'
import '@/styles/dark.scss'

export const useThemeStore = defineStore('theme', {
  state: () => ({
    isDark: localStorage.getItem('theme') === 'dark',
    primaryColor: localStorage.getItem('primaryColor') || '#409EFF',
  }),

  actions: {
    toggleTheme() {
      this.isDark = !this.isDark
      localStorage.setItem('theme', this.isDark ? 'dark' : 'light')
      this.applyTheme()
    },

    setPrimaryColor(color) {
      if (!color) return
      
      const el = document.documentElement
      el.style.setProperty('--el-color-primary', color)
      
      // 转换为 RGB 格式并存储
      const rgb = this.hexToRgb(color)
      if (rgb) {
        el.style.setProperty('--el-color-primary-rgb', `${rgb.r}, ${rgb.g}, ${rgb.b}`)
      }
      
      // 生成主题色的不同层级
      const levels = [0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1]
      levels.forEach((level, index) => {
        const lightColor = this.getLightColor(color, level)
        el.style.setProperty(`--el-color-primary-light-${index + 1}`, lightColor)
      })
      
      // 保存到本地存储
      localStorage.setItem('primaryColor', color)
      this.primaryColor = color
    },

    applyTheme() {
      // 应用暗色主题
      document.documentElement.classList.toggle('dark', this.isDark)
      
      // 设置主题色
      const el = document.documentElement
      const primaryRgb = this.hexToRgb(this.primaryColor)
      
      el.style.setProperty('--el-color-primary', this.primaryColor)
      el.style.setProperty('--el-color-primary-rgb', primaryRgb)
      
      // 生成主题色的不同层级
      for (let i = 1; i <= 9; i++) {
        const lightColor = this.getLightColor(this.primaryColor, i / 10)
        el.style.setProperty(`--el-color-primary-light-${i}`, lightColor)
      }
      
      // 生成暗色主题变量
      if (this.isDark) {
        el.style.setProperty('--el-bg-color', '#141414')
        el.style.setProperty('--el-bg-color-page', '#0a0a0a')
        el.style.setProperty('--el-text-color-primary', '#ffffff')
        el.style.setProperty('--el-text-color-regular', '#e5eaf3')
        el.style.setProperty('--el-text-color-secondary', '#a3a6ad')
        el.style.setProperty('--el-border-color-light', '#363637')
        el.style.setProperty('--el-fill-color-light', '#262727')
        el.style.setProperty('--el-menu-bg-color', '#141414')
      } else {
        el.style.removeProperty('--el-bg-color')
        el.style.removeProperty('--el-bg-color-page')
        el.style.removeProperty('--el-text-color-primary')
        el.style.removeProperty('--el-text-color-regular')
        el.style.removeProperty('--el-text-color-secondary')
        el.style.removeProperty('--el-border-color-light')
        el.style.removeProperty('--el-fill-color-light')
        el.style.removeProperty('--el-menu-bg-color')
      }
    },

    // 将 hex 颜色转换为 rgb
    hexToRgb(hex) {
      const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
      return result
        ? `${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}`
        : null
    },

    // 获取主题色的浅色变体
    getLightColor(color, level) {
      const rgb = this.hexToRgb(color)
      if (!rgb) return color
      const [r, g, b] = rgb.split(',').map(Number)
      const amount = level * 100
      return `rgb(${Math.min(r + amount, 255)}, ${Math.min(g + amount, 255)}, ${Math.min(b + amount, 255)})`
    },

    // 初始化主题
    initTheme() {
      this.applyTheme()
    }
  }
}) 