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
  background-color: #f5f7fa;
}

.search-section {
  padding: 40px 20px;
  background-size: cover;
  background-position: center;
  text-align: center;
  position: relative;
  color: white;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(to right, #4e54c8, #8f94fb);
    opacity: 0.9;
  }

  .title {
    position: relative;
    font-size: 2.5em;
    margin-bottom: 20px;
    font-weight: bold;
  }

  .action-buttons {
    position: relative;
    margin-bottom: 20px;

    .action-btn {
      margin: 0 10px;
      padding: 12px 24px;
      font-size: 16px;
    }
  }

  .search-box {
    position: relative;
    max-width: 600px;
    margin: 0 auto;

    :deep(.el-input__wrapper) {
      padding: 12px;
      background: rgba(255, 255, 255, 0.9);
    }

    :deep(.el-input__inner) {
      font-size: 16px;
    }
  }
}

.category-nav {
  display: flex;
  justify-content: center;
  padding: 20px;
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow-x: auto;

  .category-item {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    margin: 0 8px;
    cursor: pointer;
    border-radius: 20px;
    transition: all 0.3s;

    &:hover {
      background: #f0f2f5;
    }

    &.active {
      background: #409eff;
      color: white;
    }

    .el-icon {
      margin-right: 8px;
    }
  }
}

.assistants-grid {
  padding: 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  max-width: 1200px;
  margin: 0 auto;

  .assistant-card {
    .assistant-header {
      display: flex;
      align-items: center;
      gap: 12px;

      .assistant-info {
        flex: 1;
        h3 {
          margin: 0;
          font-size: 16px;
        }
        p {
          margin: 4px 0 0;
          font-size: 14px;
          color: #666;
        }
      }

      .use-btn {
        flex-shrink: 0;
      }
    }
  }

}

// 响应式布局
@media (max-width: 768px) {
  .search-section {
    padding: 20px;

    .title {
      font-size: 1.8em;
    }

    .action-buttons {
      .action-btn {
        padding: 8px 16px;
        font-size: 14px;
      }
    }
  }

  .category-nav {
    padding: 10px;

    .category-item {
      padding: 6px 12px;
      font-size: 14px;
    }
  }

  .assistants-grid {
    grid-template-columns: 1fr;
    padding: 10px;
  }
}
</style>
