<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
      <!-- Logo -->
      <div class="logo" :class="{ 'is-collapse': isCollapse }">
        <div class="logo-content">
          <el-icon class="logo-icon"><Monitor /></el-icon>
          <span class="logo-text" v-show="!isCollapse">Txing Admin</span>
        </div>
      </div>

      <!-- 导航菜单 -->
      <el-scrollbar>
        <el-menu
          class="menu"
          :collapse="isCollapse"
          :collapse-transition="false"
          :router="true"
          :default-active="activeMenu"
        >
          <el-menu-item index="/dashboard">
            <el-icon><DataBoard /></el-icon>
            <template #title>控制台</template>
          </el-menu-item>

          <el-menu-item index="/users">
            <el-icon><User /></el-icon>
            <template #title>用户管理</template>
          </el-menu-item>

          <el-menu-item index="/channels">
            <el-icon><Connection /></el-icon>
            <template #title>渠道设置</template>
          </el-menu-item>

          <el-menu-item index="/models">
            <el-icon><Cpu /></el-icon>
            <template #title>模型管理</template>
          </el-menu-item>

          <el-menu-item index="/pricing">
            <el-icon><Wallet /></el-icon>
            <template #title>价格设置</template>
          </el-menu-item>

          <el-menu-item index="/orders">
            <el-icon><List /></el-icon>
            <template #title>订单管理</template>
          </el-menu-item>

          <el-menu-item index="/system">
            <el-icon><Setting /></el-icon>
            <template #title>系统设置</template>
          </el-menu-item>

          <el-menu-item index="/logs">
            <el-icon><Document /></el-icon>
            <template #title>系统日志</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <!-- 主体区域 -->
    <el-container class="main-container">
      <!-- 顶部导航 -->
      <el-header height="60px" class="header">
        <div class="header-left">
          <el-icon
            class="toggle-sidebar"
            @click="toggleSidebar"
          >
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentRoute.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <!-- 全局搜索 -->
          <div class="search-box">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索..."
              :prefix-icon="Search"
              clearable
            />
          </div>

          <!-- 快捷操作 -->
          <div class="quick-actions">
            <el-tooltip content="全屏" placement="bottom">
              <el-icon class="action-icon" @click="toggleFullscreen"><FullScreen /></el-icon>
            </el-tooltip>
            <el-tooltip :content="isDark ? '浅色模式' : '深色模式'" placement="bottom">
              <el-icon class="action-icon" @click="toggleTheme"><Moon /></el-icon>
            </el-tooltip>
            <el-tooltip content="主题设置" placement="bottom">
              <el-icon class="action-icon" @click="showThemePanel = true"><Setting /></el-icon>
            </el-tooltip>
          </div>

          <!-- 用户信息 -->
          <el-dropdown trigger="hover" @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :src="userAvatar" class="user-avatar">
                {{ userInfo?.userName?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <div class="user-detail" v-show="!isCollapse">
                <span class="username">{{ userInfo?.userName || '管理员' }}</span>
                <span class="user-role">超级管理员</span>
              </div>
              <el-icon class="el-icon--right"><CaretBottom /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  <span>个人信息</span>
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  <span>系统设置</span>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  <span>退出登录</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>

    <!-- 主题设置抽屉 -->
    <el-drawer
      v-model="showThemePanel"
      title="主题设置"
      size="300px"
      :with-header="false"
    >
      <div class="theme-drawer">
        <div class="drawer-header">
          <h2>主题设置</h2>
          <el-icon class="close-icon" @click="showThemePanel = false"><Close /></el-icon>
        </div>
        <div class="drawer-content">
          <div class="setting-item">
            <span class="setting-label">主题色</span>
            <el-color-picker
              v-model="tempPrimaryColor"
              :predefine="[
                '#409EFF',
                '#67C23A',
                '#E6A23C',
                '#F56C6C',
                '#909399'
              ]"
              show-alpha
            />
          </div>
          <div class="setting-item">
            <span class="setting-label">深色模式</span>
            <el-switch
              v-model="isDark"
              @change="toggleTheme"
            />
          </div>
        </div>
        <div class="drawer-footer">
          <el-button @click="showThemePanel = false">取消</el-button>
          <el-button type="primary" @click="confirmThemeChange">确认</el-button>
        </div>
      </div>
    </el-drawer>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElColorPicker } from 'element-plus'
import { useThemeStore } from '@/stores/theme'
import {
  User,
  Fold,
  Expand,
  CaretBottom,
  Monitor,
  DataBoard,
  Connection,
  Cpu,
  Wallet,
  List,
  Setting,
  Document,
  Search,
  FullScreen,
  Moon,
  SwitchButton,
  Close
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()
const isCollapse = ref(false)
const searchKeyword = ref('')
const showThemePanel = ref(false)

// 获取用户信息
const userInfo = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

// 用户头像
const userAvatar = computed(() => {
  return userInfo.value?.avatar || ''
})

// 当前激活的菜单
const activeMenu = computed(() => route.path)

// 当前路由信息
const currentRoute = computed(() => route)

// 主题相关计算属性
const isDark = computed(() => themeStore.isDark)
const tempPrimaryColor = ref(themeStore.primaryColor)

// 切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 切换全屏
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}

// 切换主题
const toggleTheme = () => {
  themeStore.toggleTheme()
}

// 确认主题色更改
const confirmThemeChange = () => {
  themeStore.setPrimaryColor(tempPrimaryColor.value)
  showThemePanel.value = false
}

// 监听面板显示状态，重置临时颜色
watch(showThemePanel, (val) => {
  if (val) {
    tempPrimaryColor.value = themeStore.primaryColor
  }
})

// 初始化主题
onMounted(() => {
  themeStore.initTheme()
})

// 处理下拉菜单命令
const handleCommand = async (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/system')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确认退出登录吗？', '提示', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          type: 'warning'
        })
        localStorage.removeItem('token')
        localStorage.removeItem('refreshToken')
        localStorage.removeItem('user')
        router.push('/login')
      } catch {
        // 取消退出
      }
      break
  }
}
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
  width: 100vw;
  background: var(--el-bg-color);
  overflow: hidden;
}

.aside {
  background: var(--el-menu-bg-color);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  z-index: 10;
  box-shadow: 2px 0 12px rgba(0, 0, 0, 0.12);
  border-right: none;
  border-radius: 0 24px 24px 0;
  overflow: hidden;

  .logo {
    height: 70px;
    padding: 0 24px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
    background: var(--el-menu-bg-color);
    border-bottom: 1px solid rgba(255,255,255,0.08);
    backdrop-filter: blur(8px);

    &.is-collapse {
      padding: 0 16px;
    }

    .logo-content {
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: flex-start;
      gap: 16px;
    }

    .logo-icon {
      font-size: 28px;
      color: var(--el-color-primary);
      filter: drop-shadow(0 0 8px var(--el-color-primary-light-5));
      transition: all 0.4s;

      &:hover {
        transform: rotate(360deg);
      }
    }

    .logo-text {
      font-size: 20px;
      font-weight: 600;
      background: linear-gradient(135deg, var(--el-color-primary), #409EFF 70%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      letter-spacing: 1px;
    }
  }

  .menu {
    border: none;
    padding: 16px;

    :deep(.el-menu-item) {
      height: 56px;
      line-height: 56px;
      margin: 8px 0;
      border-radius: 12px;
      font-size: 15px;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

      .el-menu-tooltip__trigger {
        justify-content: center;
      }

      &.is-active {
        //background: var(--el-color-primary-light-9);
        color: var(--el-color-primary);
        transform: translateX(4px);

        &::before {
          content: '';
          position: absolute;
          left: -16px;
          top: 50%;
          transform: translateY(-50%);
          height: 24px;
          width: 4px;
          background: var(--el-color-primary);
          border-radius: 0 4px 4px 0;
          box-shadow: 0 0 8px var(--el-color-primary);
          animation: glow 1.5s ease-in-out infinite alternate;
        }

        .el-icon {
          animation: iconFloat 3s ease-in-out infinite;
        }
      }

      &:hover {
        background: var(--el-color-primary-light-9);
        transform: translateX(4px);
      }

      .el-icon {
        font-size: 20px;
        transition: all 0.3s;
      }
    }
  }
}

.main-container {
  background: var(--el-bg-color-page);
  position: relative;
  overflow: hidden;
}

.header {
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.05);

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;

    .toggle-sidebar {
      font-size: 20px;
      cursor: pointer;
      color: var(--el-text-color-secondary);
      transition: all 0.3s;

      &:hover {
        color: var(--el-color-primary);
        transform: rotate(180deg);
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 24px;

    .search-box {
      width: 240px;

      :deep(.el-input__wrapper) {
        border-radius: 20px;
        box-shadow: 0 0 0 1px var(--el-border-color-light);

        &:hover, &:focus-within {
          box-shadow: 0 0 0 1px var(--el-color-primary);
        }
      }
    }

    .quick-actions {
      display: flex;
      align-items: center;
      gap: 16px;

      .action-icon {
        font-size: 20px;
        color: var(--el-text-color-secondary);
        cursor: pointer;
        transition: all 0.3s;

        &:hover {
          color: var(--el-color-primary);
          transform: translateY(-2px);
        }
      }

      .action-badge {
        :deep(.el-badge__content) {
          box-shadow: 0 0 0 1px var(--el-bg-color);
        }
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 12px;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 6px;
      transition: all 0.3s;
      outline: none !important;

      &:hover {
        background: var(--el-fill-color-light);
      }

      .user-avatar {
        border: 2px solid var(--el-color-primary-light-5);
        background: var(--el-color-primary-light-9);
        color: var(--el-color-primary);
        font-weight: bold;
      }

      .user-detail {
        display: flex;
        flex-direction: column;
        line-height: 1.2;

        .username {
          font-size: 14px;
          color: var(--el-text-color-primary);
          font-weight: 500;
        }

        .user-role {
          font-size: 12px;
          color: var(--el-text-color-secondary);
        }
      }

      .el-icon--right {
        font-size: 12px;
        color: var(--el-text-color-secondary);
      }
    }
  }
}

.main {
  padding: 20px;
  background: var(--el-bg-color-page);
  height: calc(100vh - 60px);
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--el-scrollbar-bg-color);
    border-radius: 3px;
  }
}

// 路由过渡动画
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

// 添加动画关键帧
@keyframes glow {
  from {
    box-shadow: 0 0 4px var(--el-color-primary);
  }
  to {
    box-shadow: 0 0 12px var(--el-color-primary);
  }
}

@keyframes iconFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

@keyframes pulse {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(var(--el-color-danger-rgb), 0.7);
  }

  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(var(--el-color-danger-rgb), 0);
  }

  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(var(--el-color-danger-rgb), 0);
  }
}

.theme-drawer {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;

  .drawer-header {
    padding: 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--el-border-color-light);

    h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
      color: var(--el-text-color-primary);
    }

    .close-icon {
      font-size: 20px;
      color: var(--el-text-color-secondary);
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        color: var(--el-color-primary);
        transform: rotate(90deg);
      }
    }
  }

  .drawer-content {
    flex: 1;
    padding: 20px;
    overflow-y: auto;

    .setting-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 24px;

      &:last-child {
        margin-bottom: 0;
      }

      .setting-label {
        font-size: 14px;
        color: var(--el-text-color-primary);
      }
    }
  }

  .drawer-footer {
    padding: 20px;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    border-top: 1px solid var(--el-border-color-light);
  }
}

:deep(.el-drawer__body) {
  padding: 0;
}

// 暗色主题样式覆盖
:root.dark {
  --el-color-primary-light-3: var(--el-color-primary);
  --el-color-primary-light-5: var(--el-color-primary);
  --el-color-primary-light-7: var(--el-color-primary);
  --el-color-primary-light-8: var(--el-color-primary);
  --el-color-primary-light-9: rgba(var(--el-color-primary-rgb), 0.2);

  .aside {
    background: var(--el-menu-bg-color);
    box-shadow: 2px 0 12px rgba(0, 0, 0, 0.2);
  }

  .header {
    background: var(--el-bg-color);
    border-color: var(--el-border-color-light);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  }

  .el-card {
    background: var(--el-bg-color);
    border-color: var(--el-border-color-light);

    .el-card__header {
      border-color: var(--el-border-color-light);
    }
  }

  .el-table {
    --el-table-border-color: var(--el-border-color-light);
    --el-table-header-bg-color: var(--el-fill-color-light);
    --el-table-row-hover-bg-color: var(--el-fill-color-light);

    th.el-table__cell {
      background: var(--el-fill-color-light);
    }
  }
}
</style>
