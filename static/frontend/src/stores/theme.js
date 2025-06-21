import { defineStore } from 'pinia'
import '@/styles/dark.scss'
import '@/styles/light.scss'
import { toRgba, mix } from 'color2k'

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

    // 将 rgba 字符串转换为 RGB 值
    getRgbValues(color) {
      const rgba = toRgba(color)
      // toRgba 返回形如 "rgba(r, g, b, a)" 的字符串
      const values = rgba.match(/[\d.]+/g)
      if (!values) return [0, 0, 0]
      // 将 0-1 的值转换为 0-255
      return values.slice(0, 3).map(v => Math.round(parseFloat(v) * 255))
    },

    setPrimaryColor(color) {
      if (!color) return

      const el = document.documentElement
      el.style.setProperty('--el-color-primary', color)

      // 转换为 RGB 格式并存储
      const [r, g, b] = this.getRgbValues(color)
      el.style.setProperty('--el-color-primary-rgb', `${r}, ${g}, ${b}`)

      // 生成主题色的不同层级
      for (let i = 1; i <= 9; i++) {
        // 计算混合比例：从 90% 到 10%（越小越浅）
        const weight = i * 0.1
        // 使用 color2k 的 mix 函数混合颜色
        const lightColor = mix(color, '#ffffff', weight)
        el.style.setProperty(`--el-color-primary-light-${i}`, lightColor)
      }

      // 保存到本地存储
      localStorage.setItem('primaryColor', color)
      this.primaryColor = color
    },

    applyTheme() {
      // 应用暗色主题
      document.documentElement.classList.toggle('dark', this.isDark)

      // 设置主题色
      const el = document.documentElement
      const [r, g, b] = this.getRgbValues(this.primaryColor)

      el.style.setProperty('--el-color-primary', this.primaryColor)
      el.style.setProperty('--el-color-primary-rgb', `${r}, ${g}, ${b}`)

      // 生成主题色的不同层级
      for (let i = 1; i <= 9; i++) {
        const weight = i * 0.1
        const lightColor = mix(this.primaryColor, '#ffffff', weight)
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
        // 黑暗主题下 hover 下背景色为透明
        el.style.setProperty('--el-hover-fill', 'rgba(255, 255, 255, 0.1)')
      } else {
        el.style.removeProperty('--el-bg-color')
        el.style.removeProperty('--el-bg-color-page')
        el.style.removeProperty('--el-text-color-primary')
        el.style.removeProperty('--el-text-color-regular')
        el.style.removeProperty('--el-text-color-secondary')
        el.style.removeProperty('--el-border-color-light')
        el.style.removeProperty('--el-fill-color-light')
        el.style.removeProperty('--el-menu-bg-color')
        // hover 下背景色
        el.style.setProperty('--el-hover-fill', '#dcdbdb')
      }
    },

    // 初始化主题
    initTheme() {
      this.applyTheme()
    }
  }
})
