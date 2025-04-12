<template>
  <div class="assistant-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section" :style="{ backgroundImage: `url(${bgImage})` }">
      <div class="nav-header">
        <div class="nav-content">
          <div class="nav-left">
            <div class="logo">
              <span class="logo-text">Txing AI</span>
            </div>
          </div>
          <div class="nav-right">
            <el-dropdown trigger="click">
              <div class="user-avatar">
                <el-avatar :size="32" src="https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png" />
                <el-icon class="el-icon--right"><CaretBottom /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item>
                    <el-icon><User /></el-icon>
                    <span>个人中心</span>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-icon><Setting /></el-icon>
                    <span>设置</span>
                  </el-dropdown-item>
                  <el-dropdown-item divided>
                    <el-icon><SwitchButton /></el-icon>
                    <span>退出登录</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>

      <div class="search-content">
        <h1 class="title">做您强大的 AI 助手</h1>
        <div class="action-buttons">
          <el-button type="primary" class="action-btn" @click="startChat">
            <el-icon><Timer /></el-icon>
            开始聊天
          </el-button>
          <el-button type="primary" class="action-btn" @click="createAssistant">
            <el-icon><Plus /></el-icon>
            创建助手
          </el-button>
        </div>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索您的 AI 助手"
            :prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
          />
        </div>
      </div>
    </div>

    <!-- 分类导航 -->
    <div class="category-nav">
      <div
        v-for="category in categories"
        :key="category.id"
        class="category-item"
        :class="{ active: currentCategory === category.id }"
        @click="selectCategory(category.id)"
      >
        <el-icon><component :is="category.icon" /></el-icon>
        <span>{{ category.name }}</span>
      </div>
    </div>

    <!-- AI助手列表 -->
    <div class="assistants-grid">
<!--      <el-card-->
<!--        v-for="assistant in filteredAssistants"-->
<!--        :key="assistant.id"-->
<!--        class="assistant-card"-->
<!--        shadow="hover"-->
<!--      >-->
<!--        <div class="assistant-header">-->
<!--          <el-avatar :src="assistant.avatar" :size="40" />-->
<!--          <div class="assistant-info">-->
<!--            <h3>{{ assistant.name }}</h3>-->
<!--            <p>{{ assistant.description }}</p>-->
<!--          </div>-->
<!--          <el-button-->
<!--            type="primary"-->
<!--            size="small"-->
<!--            class="use-btn"-->
<!--            @click="useAssistant(assistant)"-->
<!--          >-->
<!--            使用-->
<!--          </el-button>-->
<!--        </div>-->
<!--      </el-card>-->


      <div
        v-for="preset in filteredAssistants"
        :key="preset.id"
        class="preset-card"
      >
        <div class="preset-content">
          <div class="preset-avatar">
            <el-avatar :size="36" :src="preset.avatar">
              {{ preset.name.charAt(0) }}
            </el-avatar>
          </div>
          <div class="preset-info">
            <div class="preset-name-row">
              <h3 class="preset-name">{{ preset.name }}</h3>
              <div class="preset-type" :class="preset.type">
                {{ preset.type === 'official' ? '官方' : '社区' }}
              </div>
            </div>
            <p class="preset-description">{{ preset.description }}</p>
          </div>
        </div>
        <div class="preset-actions">
          <el-button
            type="primary"
            class="use-button"
            @click.stop="useAssistant(preset)"
          >
            <span class="button-content">
              <el-icon><ArrowRight /></el-icon>
              使用
            </span>
            <span class="button-background"></span>
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  Search,
  Plus,
  Timer,
  ArrowRight,
  User,
  Setting,
  SwitchButton,
  CaretBottom
} from '@element-plus/icons-vue'
import bgImage from '@/assets/images/header-bg.jpg'
import { useRouter } from 'vue-router'

const router = useRouter()

// 搜索关键词
const searchQuery = ref('')

// 当前选中的分类
const currentCategory = ref('all')

// 分类数据
const categories = [
  { id: 'all', name: '我的助手', icon: 'User' },
  { id: 'popular', name: '热门推荐', icon: 'Star' },
  { id: 'tools', name: '实用工具', icon: 'Tools' },
  { id: 'writing', name: '文案创作', icon: 'Edit' },
  { id: 'coding', name: '编码专家', icon: 'Monitor' },
  { id: 'learning', name: '知识学习', icon: 'Reading' },
  { id: 'life', name: '生活指南', icon: 'House' },
  { id: 'other', name: '其他', icon: 'More' }
]

// 助手数据（模拟数据）
const assistants = [
  {
    id: 1,
    name: '鱼聪明 AI 助手',
    description: '您的专属大脑，双语对话帮手',
    avatar: 'https://placeholder.co/100',
    category: 'all',
    type: 'official'
  },
  {
    id: 2,
    name: 'Midjourney 提示词生成器',
    description: 'midjourney 提示词自动生成',
    avatar: 'https://placeholder.co/100',
    category: 'tools',
    type: 'official'
  },
  {
    id: 3,
    name: '小红书文案写手',
    description: '帮您写出爆款小红书文案',
    avatar: 'https://placeholder.co/100',
    category: 'writing',
    type: 'community'
  },
  // 更多助手数据...
]

// 根据搜索和分类过滤助手
const filteredAssistants = computed(() => {
  return assistants.filter(assistant => {
    const matchesSearch = searchQuery.value === '' ||
      assistant.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      assistant.description.toLowerCase().includes(searchQuery.value.toLowerCase())

    const matchesCategory = currentCategory.value === 'all' ||
      assistant.category === currentCategory.value

    return matchesSearch && matchesCategory
  })
})

// 方法
const handleSearch = () => {
  console.log('搜索:', searchQuery.value)
}

const selectCategory = (categoryId) => {
  currentCategory.value = categoryId
}

const startChat = () => {
  router.push({
    path: '/chat',
    query: {
      newChat: 'true'
    }
  })
}

const createAssistant = () => {
  console.log('创建助手')
}

const useAssistant = (preset) => {
  console.log('使用助手:', preset)
}
</script>

<style lang="scss" scoped>
.assistant-container {
  min-height: 100vh;
  background-color: var(--bg-primary);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background:
      radial-gradient(circle at 0% 0%, rgba(43, 94, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 0%, rgba(30, 136, 229, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 100%, rgba(43, 94, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 0% 100%, rgba(3, 169, 244, 0.1) 0%, transparent 50%);
    filter: blur(60px);
    opacity: 0.5;
    z-index: 0;
    animation: backgroundShift 20s ease-in-out infinite alternate;
  }

  &::after {
    content: '';
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background:
      url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%232B5EFF' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    opacity: 0.5;
    z-index: 0;
    pointer-events: none;
  }
}

.search-section {
  padding: 0 0 60px;
  position: relative;
  color: white;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, #2B5EFF, #1E88E5);
    opacity: 0.95;
    z-index: 1;
  }

  .nav-header {
    position: relative;
    z-index: 10;
    height: 64px;
    backdrop-filter: blur(10px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .nav-content {
      max-width: 1200px;
      height: 100%;
      margin: 0 auto;
      padding: 0 24px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .nav-left {
        .logo {
          .logo-text {
            font-size: 24px;
            font-weight: bold;
            background: linear-gradient(45deg, #fff, rgba(255, 255, 255, 0.8));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
          }
        }
      }

      .nav-right {
        .user-avatar {
          display: flex;
          align-items: center;
          gap: 8px;
          cursor: pointer;
          padding: 4px;
          border-radius: 50%;
          transition: all 0.3s ease;

          &:hover {
            background: rgba(255, 255, 255, 0.1);
          }

          .el-icon--right {
            font-size: 12px;
            color: rgba(255, 255, 255, 0.8);
          }
        }
      }
    }
  }

  .search-content {
    position: relative;
    z-index: 3;
    padding: 40px 20px;
    text-align: center;

    .title {
      font-size: 3.5em;
      margin-bottom: 30px;
      font-weight: 800;
      background: linear-gradient(135deg, #fff 30%, rgba(255, 255, 255, 0.8) 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      text-shadow: 0 0 30px rgba(255, 255, 255, 0.3);
      animation: titleFloat 3s ease-in-out infinite;
    }

    .action-buttons {
      margin-bottom: 30px;
      display: flex;
      justify-content: center;
      gap: 20px;

      .action-btn {
        padding: 12px 32px;
        font-size: 16px;
        font-weight: 500;
        border: none;
        background: rgba(255, 255, 255, 0.15);
        backdrop-filter: blur(10px);
        border-radius: 12px;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        overflow: hidden;

        .el-icon {
          margin-right: 8px;
          transition: transform 0.3s ease;
        }

        &:hover {
          transform: translateY(-2px);
          background: rgba(255, 255, 255, 0.2);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);

          .el-icon {
            transform: scale(1.2);
          }
        }
      }
    }

    .search-box {
      position: relative;
      max-width: 600px;
      margin: 0 auto;
      z-index: 3;

      :deep(.el-input__wrapper) {
        padding: 12px 24px;
        background: rgba(255, 255, 255, 0.1);
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.2);
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        transition: all 0.3s ease;
        border-radius: 30px;

        &:hover, &:focus-within {
          background: rgba(255, 255, 255, 0.15);
          border-color: rgba(255, 255, 255, 0.3);
          transform: translateY(-2px);
          box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
        }
      }

      :deep(.el-input__inner) {
        font-size: 16px;
        color: white;

        &::placeholder {
          color: rgba(255, 255, 255, 0.8);
        }
      }

      :deep(.el-input__prefix) {
        color: rgba(255, 255, 255, 0.8);
      }
    }
  }
}

.category-nav {
  display: flex;
  justify-content: center;
  padding: 20px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 1;
  overflow-x: auto;

  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 2px;
    background: linear-gradient(90deg,
      rgba(43, 94, 255, 0) 0%,
      rgba(43, 94, 255, 0.5) 50%,
      rgba(43, 94, 255, 0) 100%
    );
  }

  .category-item {
    display: flex;
    align-items: center;
    padding: 10px 20px;
    margin: 0 8px;
    cursor: pointer;
    border-radius: 12px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    background: transparent;
    border: 1px solid transparent;

    &:hover {
      background: rgba(43, 94, 255, 0.1);
      border-color: rgba(43, 94, 255, 0.2);
      transform: translateY(-2px);
    }

    &.active {
      background: linear-gradient(135deg, #2B5EFF, #1E88E5);
      color: white;
      border: none;
      box-shadow: 0 4px 15px rgba(43, 94, 255, 0.35);
    }

    .el-icon {
      margin-right: 8px;
      transition: transform 0.3s ease;
    }

    &:hover .el-icon {
      transform: scale(1.2);
    }
  }
}

.assistants-grid {
  padding: 40px 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 30px;
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
  z-index: 1;

  .preset-card {
    background: var(--el-bg-color);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--el-border-color-lighter);
    position: relative;
    overflow: hidden;

    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 4px;
      background: linear-gradient(90deg, var(--el-color-primary), var(--el-color-primary-light-3));
      transform: translateY(-100%);
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
      border-color: var(--el-color-primary-light-5);

      &::before {
        transform: translateY(0);
      }

      .use-button {
        .button-background {
          transform: translateX(0);
        }
      }
    }

    .preset-content {
      display: flex;
      gap: 12px;
    }

    .preset-avatar {
      position: relative;
    }
  }

  .preset-info {
    flex: 1;
    min-width: 0;

    .preset-name-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 4px;
    }

    .preset-name {
      margin: 0;
      font-size: 15px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .preset-type {
      font-size: 10px;
      padding: 1px 6px;
      border-radius: 10px;
      color: white;
      font-weight: 500;
      flex-shrink: 0;

      &.official {
        background: linear-gradient(135deg, #2B5EFF, #1E88E5);
      }

      &.community {
        background: linear-gradient(135deg, #03A9F4, #1E88E5);
      }
    }

    .preset-description {
      margin: 0;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      line-height: 1.5;
    }
  }

  .preset-actions {
    margin-top: auto;

    .use-button {
      width: 100%;
      border: none;
      background: transparent;
      position: relative;
      overflow: hidden;
      transition: all 0.3s ease;
      padding: 8px 0;

      .button-content {
        position: relative;
        z-index: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 4px;
        font-size: 13px;
      }

      .button-background {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(90deg,
        var(--el-color-primary) 0%,
        var(--el-color-primary-light-3) 50%,
        var(--el-color-primary) 100%
        );
        transform: translateX(-100%);
        transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
      }

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.2);
      }
    }
  }

}

@keyframes backgroundShift {
  0% {
    transform: scale(1) rotate(0deg);
  }
  50% {
    transform: scale(1.2) rotate(5deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}

@keyframes glowPulse {
  0% {
    opacity: 0.5;
    transform: scale(0.95);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.05);
  }
  100% {
    opacity: 0.5;
    transform: scale(0.95);
  }
}

@keyframes titleFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes iconBounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

// 响应式布局优化
@media (max-width: 768px) {
  .search-section {
    padding: 40px 20px;

    .title {
      font-size: 2.5em;
    }

    .action-buttons {
      .action-btn {
        padding: 10px 20px;
        font-size: 14px;
      }
    }
  }

  .category-nav {
    padding: 15px 10px;

    .category-item {
      padding: 8px 16px;
      font-size: 14px;
    }
  }

  .assistants-grid {
    padding: 20px 10px;
    gap: 20px;
  }
}
</style>
