import './assets/main.css'
import './assets/styles/list-common.css'
import 'virtual:svg-icons-register'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import SvgIcon from '@/components/common/SvgIcon.vue'
import { permission } from './directives/permission'
import { loadAnalyticsScript } from '@/utils/analytics'

import App from './App.vue'
import router from './router'

// 加载分析脚本
loadAnalyticsScript()

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册 SVG 图标组件
app.component('SvgIcon', SvgIcon)

// 注册全局指令
app.directive('permission', permission)

app.use(pinia)
app.use(router)
app.use(ElementPlus)

app.mount('#app')
