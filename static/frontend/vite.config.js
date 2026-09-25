import { fileURLToPath } from 'node:url'
import path from 'path'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  base: mode === 'production' ? '/dash/' : '/',
  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
    createSvgIconsPlugin({
      // 指定需要缓存的图标文件夹
      iconDirs: [path.resolve(__dirname, 'src/assets/icons')],
      // 指定symbolId格式
      symbolId: 'icon-[name]'
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    },
  },
  server: {
    // WSL2 下 /mnt/d（drvfs）的 inotify 不可靠，文件变更常不触发 HMR；
    // 用轮询监听保证热更新生效（ignored 限制范围，控制 CPU 开销）
    // inotify is unreliable on /mnt/d (drvfs) under WSL2 — poll for changes so HMR fires
    watch: {
      usePolling: true,
      interval: 500,
      ignored: ['**/node_modules/**', '**/dist/**', '**/.git/**']
    },
    proxy: {
      '/api': {
        // 多 worktree 并行开发时后端端口不同，可用环境变量覆盖（默认 8081）
        target: process.env.VITE_PROXY_TARGET || 'http://localhost:8081',
        changeOrigin: true,
        ws: true,
        rewrite: (path) => path,
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq) => {
            proxyReq.removeHeader('Origin')
            proxyReq.removeHeader('Referer')
            proxyReq.removeHeader('User-Agent')
          })
        }
      }
    }
  }
}))
