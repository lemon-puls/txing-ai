<template>
  <div class="home-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section" :style="{ backgroundImage: `url(${bgImage})` }">
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
      <el-card
        v-for="assistant in filteredAssistants"
        :key="assistant.id"
        class="assistant-card"
        shadow="hover"
      >
        <div class="assistant-header">
          <el-avatar :src="assistant.avatar" :size="40" />
          <div class="assistant-info">
            <h3>{{ assistant.name }}</h3>
            <p>{{ assistant.description }}</p>
          </div>
          <el-button
            type="primary"
            size="small"
            class="use-btn"
            @click="useAssistant(assistant)"
          >
            使用
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Search, Plus, Timer } from '@element-plus/icons-vue'
import bgImage from '@/assets/images/header-bg.jpg'

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
    category: 'all'
  },
  {
    id: 2,
    name: 'Midjourney 提示词生成器',
    description: 'midjourney 提示词自动生成',
    avatar: 'https://placeholder.co/100',
    category: 'tools'
  },
  {
    id: 3,
    name: '小红书文案写手',
    description: '帮您写出爆款小红书文案',
    avatar: 'https://placeholder.co/100',
    category: 'writing'
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
  console.log('开始聊天')
}

const createAssistant = () => {
  console.log('创建助手')
}

const useAssistant = (assistant) => {
  console.log('使用助手:', assistant)
}
</script>

<style lang="scss" scoped>
.home-container {
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
      radial-gradient(circle at 0% 0%, rgba(64, 158, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 0%, rgba(200, 80, 192, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 100%, rgba(64, 158, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 0% 100%, rgba(200, 80, 192, 0.1) 0%, transparent 50%);
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
      url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%234158D0' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    opacity: 0.5;
    z-index: 0;
    pointer-events: none;
  }
}

.search-section {
  padding: 60px 20px;
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
    background: linear-gradient(135deg, #4158D0, #C850C0);
    opacity: 0.95;
    z-index: 1;
  }

  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: 
      radial-gradient(circle at 20% 50%, rgba(255, 255, 255, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 80% 50%, rgba(255, 255, 255, 0.1) 0%, transparent 50%);
    z-index: 2;
    animation: glowPulse 8s ease-in-out infinite alternate;
  }

  .title {
    position: relative;
    font-size: 3.5em;
    margin-bottom: 30px;
    font-weight: 800;
    z-index: 3;
    background: linear-gradient(135deg, #fff 30%, rgba(255, 255, 255, 0.8) 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    text-shadow: 0 0 30px rgba(255, 255, 255, 0.3);
    animation: titleFloat 3s ease-in-out infinite;
  }

  .action-buttons {
    position: relative;
    margin-bottom: 30px;
    z-index: 3;

    .action-btn {
      margin: 0 10px;
      padding: 12px 32px;
      font-size: 16px;
      font-weight: 500;
      border: none;
      background: rgba(255, 255, 255, 0.15);
      backdrop-filter: blur(10px);
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      position: relative;
      overflow: hidden;

      &::before {
        content: '';
        position: absolute;
        top: -50%;
        left: -50%;
        width: 200%;
        height: 200%;
        background: conic-gradient(
          from 0deg,
          transparent 0%,
          rgba(255, 255, 255, 0.2) 25%,
          transparent 50%
        );
        animation: rotate 4s linear infinite;
      }

      &:hover {
        transform: translateY(-2px) scale(1.05);
        box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
      }

      .el-icon {
        margin-right: 8px;
        animation: iconBounce 2s infinite;
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
      rgba(64, 158, 255, 0) 0%,
      rgba(64, 158, 255, 0.5) 50%,
      rgba(64, 158, 255, 0) 100%
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
      background: rgba(64, 158, 255, 0.1);
      border-color: rgba(64, 158, 255, 0.2);
      transform: translateY(-2px);
    }

    &.active {
      background: linear-gradient(135deg, #4158D0, #C850C0);
      color: white;
      border: none;
      box-shadow: 0 4px 15px rgba(65, 88, 208, 0.35);
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

  .assistant-card {
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(64, 158, 255, 0.1);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;

    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 4px;
      background: linear-gradient(90deg, #4158D0, #C850C0);
      opacity: 0;
      transition: opacity 0.3s ease;
    }

    &:hover {
      transform: translateY(-8px);
      box-shadow: 0 12px 30px rgba(0, 0, 0, 0.1);
      border-color: rgba(64, 158, 255, 0.2);

      &::before {
        opacity: 1;
      }
    }

    .assistant-header {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 20px;

      .el-avatar {
        border: 2px solid transparent;
        background: linear-gradient(135deg, #4158D0, #C850C0);
        transition: all 0.3s ease;

        &:hover {
          transform: rotate(360deg);
        }
      }

      .assistant-info {
        flex: 1;
        h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 600;
          background: linear-gradient(135deg, #4158D0, #C850C0);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }
        p {
          margin: 8px 0 0;
          font-size: 14px;
          color: #666;
          line-height: 1.4;
        }
      }

      .use-btn {
        padding: 8px 24px;
        font-weight: 500;
        background: linear-gradient(135deg, #4158D0, #C850C0);
        border: none;
        position: relative;
        overflow: hidden;

        &::before {
          content: '';
          position: absolute;
          top: -50%;
          left: -50%;
          width: 200%;
          height: 200%;
          background: conic-gradient(
            from 0deg,
            transparent 0%,
            rgba(255, 255, 255, 0.2) 25%,
            transparent 50%
          );
          animation: rotate 4s linear infinite;
        }

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 15px rgba(65, 88, 208, 0.35);
        }
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
