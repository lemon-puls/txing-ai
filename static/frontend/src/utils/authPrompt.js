import { ref } from 'vue'

// 全局登录弹窗状态：登录失效（401 且无法静默续期）时由 api 层唤起，
// App.vue 统一挂载 AuthDialog，任何页面无需各自实现
// Global login-dialog state: raised by the api layer when a session expires
// (401 that can't be silently refreshed); App.vue mounts the dialog once for all pages

// 登录弹窗可见性 / login dialog visibility
export const authDialogVisible = ref(false)

// 登录成功计数器：页面 watch 它即可在登录成功后自动重载自己failed的数据
// resolved tick: pages watch this to refetch after a successful re-login
export const authResolvedTick = ref(0)

// 唤起登录弹窗（幂等：并发多个 401 也只弹一个）
// raise the login dialog (idempotent: parallel 401s open a single dialog)
export function showAuthDialog() {
  authDialogVisible.value = true
}

// 登录成功后收尾：关弹窗并广播 tick
// called on login success: close the dialog and bump the tick
export function notifyAuthResolved() {
  authDialogVisible.value = false
  authResolvedTick.value++
}

// 是否为登录类错误（页面 catch 里可据此跳过冗余的报错 toast，弹窗本身就是提示）
// auth-related rejections: pages can skip their own error toast for these
export function isAuthError(err) {
  const msg = err?.message || ''
  return msg.includes('请先登录') || msg.includes('请重新登录')
}
