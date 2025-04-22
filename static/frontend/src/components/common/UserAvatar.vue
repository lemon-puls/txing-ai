<template>
  <!-- 未登录状态显示登录按钮 -->
  <el-button v-if="!userStore.isLoggedIn" type="primary" link @click="showLoginDialog">
    登录
  </el-button>

  <!-- 已登录状态显示头像下拉菜单 -->
  <el-dropdown v-else trigger="click">
    <div class="user-avatar">
      <el-avatar :size="32" :src="userStore.avatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'" />
      <el-icon class="el-icon--right"><CaretBottom /></el-icon>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item @click="handleCommand('profile')">
          <el-icon><User /></el-icon>
          <span>个人中心</span>
        </el-dropdown-item>
        <el-dropdown-item @click="handleCommand('settings')">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-dropdown-item>
        <el-dropdown-item v-if="userStore.isAdmin" @click="handleCommand('admin')" divided>
          <el-icon><Monitor /></el-icon>
          <span>管理后台</span>
        </el-dropdown-item>
        <el-dropdown-item divided @click="handleCommand('logout')">
          <el-icon><SwitchButton /></el-icon>
          <span>退出登录</span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>

  <!-- 登录弹窗 -->
  <AuthDialog 
    v-model="loginDialogVisible"
    @login-success="handleLoginSuccess"
  />
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  Setting,
  SwitchButton,
  CaretBottom,
  Monitor
} from '@element-plus/icons-vue'
import AuthDialog from '../AuthDialog.vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

// 登录弹窗显示状态
const loginDialogVisible = ref(false)

// 显示登录弹窗
const showLoginDialog = () => {
  loginDialogVisible.value = true
}

const handleCommand = (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'admin':
      router.push('/admin')
      break
    case 'logout':
      // 处理登出逻辑
      userStore.logout()
      break
  }
}

// 登录成功回调
const handleLoginSuccess = () => {
  loginDialogVisible.value = false
}
</script>

<style scoped lang="scss">
.user-avatar {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  transition: all 0.3s ease;

  &:hover {
    background: rgba(var(--el-color-primary-rgb), 0.1);
  }

  .el-icon--right {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}
</style> 