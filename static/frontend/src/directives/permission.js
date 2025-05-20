import { useUserStore } from '@/stores/user'
import { hasRole, hasPermission } from '@/utils/auth'
import { ElMessage } from 'element-plus'

/**
 * 权限指令
 * Permission directive
 * 
 * 使用方式 Usage:
 * v-permission:role="['admin', 'editor']"  // 角色权限 Role permission
 * v-permission:perm="['create', 'edit']"   // 操作权限 Permission-based
 * v-permission:role.hide="['admin']"       // 无权限时隐藏元素 Hide when no permission
 * v-permission:perm.hide="['create']"      // 无权限时隐藏元素 Hide when no permission
 */
export const permission = {
  mounted(el, binding) {
    const userStore = useUserStore()
    const { arg, value, modifiers } = binding

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

    // 如果没有权限
    // If no permission
    if (!hasAuth) {
      // 检查是否需要隐藏元素
      // Check if element should be hidden
      if (modifiers.hide) {
        // 隐藏元素
        // Hide element
        el.style.display = 'none'
      } else {
        // 设置禁用样式，但不设置 disabled 属性
        // Set disabled style without setting disabled attribute
        el.style.opacity = '0.5'
        el.style.cursor = 'not-allowed'
        
        // 创建一个覆盖层来捕获点击事件
        // Create an overlay to capture click events
        const overlay = document.createElement('div')
        overlay.style.position = 'absolute'
        overlay.style.top = '0'
        overlay.style.left = '0'
        overlay.style.width = '100%'
        overlay.style.height = '100%'
        overlay.style.zIndex = '999'
        overlay.style.backgroundColor = 'transparent'
        overlay.style.cursor = 'not-allowed'
        
        // 设置鼠标悬停提示
        // Set mouseover tooltip
        overlay.title = '您没有权限执行此操作'
        
        // 添加点击事件监听器
        // Add click event listener
        overlay.onclick = (e) => {
          e.preventDefault()
          e.stopPropagation()
          
          // 显示无权限提示
          // Show no permission message
          ElMessage({
            message: '您没有权限执行此操作',
            type: 'warning',
            duration: 2000
          })
          
          return false
        }
        
        // 确保父元素是相对定位，以便正确定位覆盖层
        // Make sure parent element has relative position for proper overlay positioning
        if (getComputedStyle(el).position === 'static') {
          el.style.position = 'relative'
        }
        
        // 添加覆盖层到元素
        // Add overlay to element
        el.appendChild(overlay)
      }
    }
  }
} 