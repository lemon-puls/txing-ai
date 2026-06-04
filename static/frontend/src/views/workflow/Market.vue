<template>
  <div class="workflow-market-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section">
      <div class="search-bg-overlay"></div>
      <div class="search-particles">
        <span v-for="i in 6" :key="i" class="particle" :class="`particle-${i}`"></span>
      </div>
      <div class="search-content">
        <h1 class="title">AI 工作流市场</h1>
        <p class="subtitle">发现并使用强大的 AI 工作流，自动化您的任务</p>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索工作流..."
            :prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
            :loading="loading"
            size="large"
          />
        </div>
      </div>
    </div>

    <!-- 分类导航 -->
    <div class="tag-nav-wrapper">
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
    </div>

    <!-- 工作流列表 -->
    <div class="workflows-section">
      <el-empty
        v-if="!loading && workflows.length === 0"
        description="暂无可用工作流"
      />
      <div v-else class="workflows-grid">
        <div
          v-for="workflow in workflows"
          :key="workflow.id"
          class="workflow-card"
          @click="useWorkflow(workflow)"
        >
          <div class="card-accent"></div>
          <div class="card-body">
            <div class="card-header">
              <div class="workflow-icon" :class="getCategoryClass(workflow.category)">
                <el-icon :size="24"><Connection /></el-icon>
              </div>
              <div class="workflow-meta">
                <h3 class="workflow-name">{{ workflow.name }}</h3>
                <el-tag
                  v-if="workflow.category"
                  class="workflow-category"
                  :type="getCategoryTagType(workflow.category)"
                  effect="light"
                  size="small"
                  round
                >
                  {{ getCategoryLabel(workflow.category) }}
                </el-tag>
              </div>
            </div>
            <p class="workflow-description">{{ workflow.description || '暂无描述' }}</p>
          </div>
          <div class="card-footer">
            <div class="footer-left">
              <span class="run-hint">点击运行</span>
            </div>
            <el-button
              type="primary"
              class="use-button"
              circle
              @click.stop="useWorkflow(workflow)"
            >
              <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
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

const getCategoryClass = (category) => {
  const map = {
    life: 'cat-life',
    tools: 'cat-tools',
    writing: 'cat-writing',
    coding: 'cat-coding',
    learning: 'cat-learning'
  }
  return map[category] || 'cat-default'
}

const getCategoryTagType = (category) => {
  const map = {
    life: 'success',
    tools: 'warning',
    writing: 'primary',
    coding: '',
    learning: 'danger'
  }
  return map[category] || 'info'
}

const getCategoryLabel = (category) => {
  const map = {
    life: '生活',
    tools: '工具',
    writing: '创作',
    coding: '编码',
    learning: '学习',
    popular: '热门',
    other: '其他'
  }
  return map[category] || category
}

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
// 蓝色主题变量
$blue-500: #2B5EFF;
$blue-400: #4facfe;
$blue-600: #1E88E5;
$blue-gradient: linear-gradient(135deg, $blue-500, $blue-400);

.workflow-market-container {
  min-height: 100vh;
  background: var(--el-bg-color-page, #f5f7fa);
}

.search-section {
  position: relative;
  padding: 60px 20px 70px;
  background: $blue-gradient;
  overflow: hidden;

  .search-bg-overlay {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse at 20% 50%, rgba(255, 255, 255, 0.12) 0%, transparent 60%),
      radial-gradient(ellipse at 80% 20%, rgba(255, 255, 255, 0.08) 0%, transparent 50%);
    z-index: 1;
  }

  .search-particles {
    position: absolute;
    inset: 0;
    z-index: 1;
    overflow: hidden;

    .particle {
      position: absolute;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.08);
      animation: float 20s infinite ease-in-out;

      &.particle-1 { width: 200px; height: 200px; top: -50px; left: 10%; animation-delay: 0s; }
      &.particle-2 { width: 120px; height: 120px; top: 30%; right: 15%; animation-delay: -5s; }
      &.particle-3 { width: 80px; height: 80px; bottom: 10%; left: 30%; animation-delay: -10s; }
      &.particle-4 { width: 150px; height: 150px; top: 10%; right: 30%; animation-delay: -3s; }
      &.particle-5 { width: 60px; height: 60px; bottom: 20%; right: 10%; animation-delay: -8s; }
      &.particle-6 { width: 100px; height: 100px; top: 50%; left: 5%; animation-delay: -12s; }
    }
  }

  .search-content {
    position: relative;
    z-index: 2;
    text-align: center;
    max-width: 700px;
    margin: 0 auto;

    .title {
      font-size: 2.8em;
      margin: 0 0 12px;
      font-weight: 800;
      color: #fff;
      letter-spacing: -0.5px;
      text-shadow: 0 2px 10px rgba(0, 0, 0, 0.15);
    }

    .subtitle {
      font-size: 1.1em;
      color: rgba(255, 255, 255, 0.85);
      margin: 0 0 32px;
    }

    .search-box {
      max-width: 520px;
      margin: 0 auto;

      :deep(.el-input__wrapper) {
        padding: 6px 8px 6px 20px;
        background: rgba(255, 255, 255, 0.15);
        backdrop-filter: blur(12px);
        border: 1px solid rgba(255, 255, 255, 0.25);
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
        border-radius: 14px;
        transition: all 0.3s ease;

        &:hover, &:focus-within {
          background: rgba(255, 255, 255, 0.22);
          border-color: rgba(255, 255, 255, 0.4);
        }
      }

      :deep(.el-input__inner) {
        font-size: 15px;
        color: #fff;
        &::placeholder { color: rgba(255, 255, 255, 0.7); }
      }

      :deep(.el-input__prefix) { color: rgba(255, 255, 255, 0.7); }
      :deep(.el-input__clear) { color: rgba(255, 255, 255, 0.7); }
    }
  }
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  33% { transform: translateY(-20px) scale(1.05); }
  66% { transform: translateY(10px) scale(0.95); }
}

.tag-nav-wrapper {
  position: relative;
  z-index: 10;
  margin-top: -28px;
  padding: 0 20px;
}

.tag-nav {
  display: flex;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  max-width: 900px;
  margin: 0 auto;
  background: var(--el-bg-color, #fff);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  overflow-x: auto;

  &::-webkit-scrollbar { display: none; }

  .tag-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    cursor: pointer;
    border-radius: 10px;
    transition: all 0.25s ease;
    white-space: nowrap;
    font-size: 14px;
    color: var(--el-text-color-regular);
    flex-shrink: 0;

    &:hover {
      background: rgba($blue-500, 0.08);
      color: $blue-500;
    }

    &.active {
      background: $blue-gradient;
      color: #fff;
      box-shadow: 0 2px 12px rgba($blue-500, 0.35);
    }
  }
}

.workflows-section {
  padding: 32px 24px 60px;
  max-width: 1200px;
  margin: 0 auto;
}

.workflows-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.workflow-card {
  position: relative;
  background: var(--el-bg-color, #fff);
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--el-border-color-lighter, #e4e7ed);

  &:hover {
    transform: translateY(-6px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.1);
    border-color: transparent;

    .card-accent { height: 4px; }
    .use-button { transform: scale(1.1); }
    .run-hint { opacity: 1; transform: translateX(0); }
  }

  .card-accent {
    height: 3px;
    background: $blue-gradient;
    transition: height 0.3s ease;
  }

  .card-body { padding: 20px 20px 14px; }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    margin-bottom: 12px;
  }

  .workflow-icon {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    flex-shrink: 0;
    background: $blue-gradient;

    &.cat-life { background: linear-gradient(135deg, #43e97b, #38f9d7); }
    &.cat-tools { background: linear-gradient(135deg, #f6d365, #fda085); }
    &.cat-writing { background: linear-gradient(135deg, #a18cd1, #fbc2eb); }
    &.cat-coding { background: $blue-gradient; }
    &.cat-learning { background: linear-gradient(135deg, #ff9a9e, #fecfef); }
    &.cat-default { background: linear-gradient(135deg, #89f7fe, #66a6ff); }
  }

  .workflow-meta {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .workflow-name {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .workflow-description {
    margin: 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px;
    border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  }

  .run-hint {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
    opacity: 0;
    transform: translateX(-8px);
    transition: all 0.3s ease;
  }

  .use-button {
    width: 36px;
    height: 36px;
    background: $blue-gradient;
    border: none;
    box-shadow: 0 2px 8px rgba($blue-500, 0.3);
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 4px 16px rgba($blue-500, 0.4);
    }
  }
}

@media (max-width: 768px) {
  .search-section {
    padding: 40px 16px 50px;
    .search-content .title { font-size: 2em; }
  }

  .tag-nav-wrapper { padding: 0 12px; }

  .tag-nav {
    justify-content: flex-start;
    padding: 8px 12px;
    .tag-item { padding: 6px 12px; font-size: 13px; }
  }

  .workflows-section { padding: 20px 12px 40px; }

  .workflows-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}
</style>
