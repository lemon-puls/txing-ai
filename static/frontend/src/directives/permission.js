import { useUserStore } from '@/stores/user'
import { hasRole, hasPermission } from '@/utils/auth'

/**
 * 权限指令
 * Permission directive
 * 
 * 使用方式 Usage:
 * v-permission:role="['admin', 'editor']"  // 角色权限 Role permission
 * v-permission:perm="['create', 'edit']"   // 操作权限 Operation permission
 */
export const permission = {
  mounted(el, binding) {
    const userStore = useUserStore()
    const { arg, value } = binding

    // 如果没有指定权限要求，则不进行权限控制
    // If no permission requirement is specified, no permission control
    if (!value) return

    let hasAuth = false

    if (arg === 'role') {
      // 检查角色权限
      // Check role permission
      hasAuth = hasRole(userStore.userRoles, value)
    } else if (arg === 'perm') {
      // 检查操作权限
      // Check operation permission
      hasAuth = hasPermission(userStore.userPermissions, value)
    }

    // 如果没有权限，则禁用元素
    // If no permission, disable the element
    if (!hasAuth) {
      el.disabled = true
      el.style.pointerEvents = 'none'
      el.style.opacity = '0.5'
      el.style.cursor = 'not-allowed'
    }
  }
} 