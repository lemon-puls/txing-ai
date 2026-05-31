<template>
  <div class="workflow-market-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section" :style="{ backgroundImage: `url(${bgImage})` }">
      <div class="search-content">
        <h1 class="title">AI 工作流市场</h1>
        <p class="subtitle">发现并使用强大的 AI 工作流，自动化您的任务</p>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索工作流"
            :prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
            :loading="loading"
          />
        </div>
      </div>
    </div>

    <!-- 分类导航 -->
    <div class="tag-nav">
      <div
        v-for="tag in categories"
        :key="tag.id"
        class="tag-item"
        :class="{ active: currentCategory === tag.id }"
        @click="selectCategory(tag.id)"
      >
        <el-icon><component :is="tag.icon" /></el-icon>
        <span>{{ tag.name }}</span>
      </div>
    </div>

    <!-- 工作流列表 -->
    <div class="workflows-grid">
      <el-empty
        v-if="!loading && workflows.length === 0"
        description="暂无可用工作流"
      />
      <div
        v-else
        v-for="workflow in workflows"
        :key="workflow.id"
        class="workflow-card"
      >
        <div class="workflow-content">
          <div class="workflow-icon">
            <el-icon :size="28"><Connection /></el-icon>
          </div>
          <div class="workflow-info">
            <div class="workflow-header">
              <h3 class="workflow-name">{{ workflow.name }}</h3>
              <el-tag
                v-if="workflow.category"
                class="workflow-category"
                effect="plain"
                size="small"
              >
                {{ workflow.category }}
              </el-tag>
            </div>
            <p class="workflow-description">{{ workflow.description || '暂无描述' }}</p>
          </div>
        </div>
        <div class="workflow-actions">
          <el-button
            type="primary"
            class="use-button"
            @click="useWorkflow(workflow)"
          >
            <span class="button-content">
              <el-icon><ArrowRight /></el-icon>
              使用
            </span>
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, ArrowRight, Connection, Grid, Star, Tools, Edit, Monitor, Reading, House, More } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import bgImage from '@/assets/images/header-bg.jpg'

defineOptions({
  name: 'WorkflowMarket'
})

const router = useRouter()

const searchQuery = ref('')
const currentCategory = ref('all')
const workflows = ref([])
const loading = ref(false)

const categories = [
  { id: 'all', name: '全部', icon: 'Grid' },
  { id: 'popular', name: '热门推荐', icon: 'Star' },
  { id: 'tools', name: '实用工具', icon: 'Tools' },
  { id: 'writing', name: '文案创作', icon: 'Edit' },
  { id: 'coding', name: '编码专家', icon: 'Monitor' },
  { id: 'learning', name: '知识学习', icon: 'Reading' },
  { id: 'life', name: '生活指南', icon: 'House' },
  { id: 'other', name: '其他', icon: 'More' }
]

const loadWorkflows = async () => {
  try {
    loading.value = true
    const params = {
      page: 1,
      limit: 999,
      name: searchQuery.value || undefined,
      category: currentCategory.value === 'all' || currentCategory.value === 'popular'
        ? undefined
        : currentCategory.value
    }
    const res = await defaultApi.apiWorkflowPublicGet(
      params.page,
      params.limit,
      {
        name: params.name,
        category: params.category
      }
    )
    if (res.code === 0 && res.data) {
      workflows.value = res.data.records || []
    } else {
      ElMessage.error(res.msg || '获取工作流列表失败')
    }
  } catch (error) {
    console.error('Load workflows error:', error)
    ElMessage.error('获取工作流列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadWorkflows()
}

const selectCategory = (tagId) => {
  currentCategory.value = tagId
  loadWorkflows()
}

const useWorkflow = (workflow) => {
  router.push({
    path: `/workflow/${workflow.id}/execute`
  })
}

onMounted(() => {
  loadWorkflows()
})
</script>

<style lang="scss" scoped>
.workflow-market-container {
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

  .search-content {
    position: relative;
    z-index: 3;
    padding: 40px 20px;
    text-align: center;

    .title {
      font-size: 3.5em;
      margin-bottom: 16px;
      font-weight: 800;
      background: linear-gradient(135deg, #fff 30%, rgba(255, 255, 255, 0.8) 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }

    .subtitle {
      font-size: 1.2em;
      color: rgba(255, 255, 255, 0.85);
      margin-bottom: 30px;
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
        border-radius: 30px;

        &:hover, &:focus-within {
          background: rgba(255, 255, 255, 0.15);
          border-color: rgba(255, 255, 255, 0.3);
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

.tag-nav {
  display: flex;
  justify-content: center;
  padding: 20px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 1;
  overflow-x: auto;

  .tag-item {
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
    }

    &.active {
      background: linear-gradient(135deg, #2B5EFF, #1E88E5);
      color: white;
      border: none;
      box-shadow: 0 4px 15px rgba(43, 94, 255, 0.35);
    }

    .el-icon {
      margin-right: 8px;
    }
  }
}

.workflows-grid {
  padding: 40px 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
  z-index: 1;

  .workflow-card {
    background: var(--el-bg-color);
    border-radius: 12px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
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
    }

    .workflow-content {
      display: flex;
      gap: 16px;
    }

    .workflow-icon {
      width: 52px;
      height: 52px;
      border-radius: 12px;
      background: linear-gradient(135deg, var(--el-color-primary), var(--el-color-primary-light-3));
      display: flex;
      align-items: center;
      justify-content: center;
      color: white;
      flex-shrink: 0;
    }

    .workflow-info {
      flex: 1;
      min-width: 0;
    }

    .workflow-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;
    }

    .workflow-name {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .workflow-category {
      flex-shrink: 0;
    }

    .workflow-description {
      margin: 0;
      font-size: 13px;
      color: var(--el-text-color-secondary);
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      line-height: 1.5;
    }
  }

  .workflow-actions {
    .use-button {
      width: 100%;
      border: none;
      background: linear-gradient(90deg, var(--el-color-primary), var(--el-color-primary-light-3));
      transition: all 0.3s ease;

      .button-content {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 4px;
      }

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.3);
      }
    }
  }
}

@media (max-width: 768px) {
  .search-section .search-content .title {
    font-size: 2.5em;
  }

  .tag-nav {
    padding: 15px 10px;

    .tag-item {
      padding: 8px 16px;
      font-size: 14px;
    }
  }

  .workflows-grid {
    padding: 20px 10px;
    grid-template-columns: 1fr;
  }
}
</style>
