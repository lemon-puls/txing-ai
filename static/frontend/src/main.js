import './assets/main.css'
import './assets/styles/list-common.css'
import 'virtual:svg-icons-register'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import SvgIcon from '@/components/common/SvgIcon.vue'

import App from './App.vue'
import router from './router'

const app = createApp(App)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册 SVG 图标组件
app.component('SvgIcon', SvgIcon)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
