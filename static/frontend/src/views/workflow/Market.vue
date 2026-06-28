<template>
  <div class="workflow-market-container">
    <!-- 顶部 Hero 区域 / Top Hero Section -->
    <section class="hero-section">
      <div class="hero-bg">
        <div class="hero-aurora aurora-1"></div>
        <div class="hero-aurora aurora-2"></div>
        <div class="hero-aurora aurora-3"></div>
        <div class="hero-grid"></div>
        <div class="hero-noise"></div>
      </div>

      <div class="hero-content">
        <div class="hero-badge">
          <span class="badge-dot"></span>
          <span>AI App Store · {{ workflows.length }}+ 应用</span>
        </div>
        <h1 class="hero-title">
          <span class="title-line">探索强大的</span>
          <span class="title-line gradient-text">AI 应用市场</span>
        </h1>
        <p class="hero-subtitle">
          即插即用的智能工作流，让 AI 真正为你所用 —— 从内容创作到代码生成，从知识学习到生活助手。
        </p>

        <div class="search-wrapper">
          <div class="search-box" :class="{ focused: searchFocused }">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
              v-model="searchQuery"
              class="search-input"
              placeholder="搜索你需要的 AI 应用..."
              clearable
              @keyup.enter="handleSearch"
              @focus="searchFocused = true"
              @blur="searchFocused = false"
            />
            <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
              <el-icon><Close /></el-icon>
            </button>
            <button class="search-btn" @click="handleSearch">
              <el-icon><Search /></el-icon>
              <span>搜索</span>
            </button>
          </div>
          <div class="hot-tags">
            <span class="hot-label">热门:</span>
            <span
              v-for="tag in hotSearches"
              :key="tag"
              class="hot-tag"
              @click="quickSearch(tag)"
            >{{ tag }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 分类导航 / Category Navigation -->
    <section class="category-section">
      <div class="category-wrapper">
        <div class="category-header">
          <h2 class="section-title">
            <span class="title-decor"></span>
            应用分类
          </h2>
          <span class="category-count">{{ workflows.length }} 个应用</span>
        </div>
        <div class="category-scroll">
          <div class="category-nav">
            <div
              v-for="tag in categories"
              :key="tag.id"
              class="category-item"
              :class="[`cat-${tag.id}`, { active: currentCategory === tag.id }]"
              @click="selectCategory(tag.id)"
            >
              <div class="cat-icon">
                <el-icon :size="18"><component :is="tag.icon" /></el-icon>
              </div>
              <div class="cat-info">
                <span class="cat-name">{{ tag.name }}</span>
                <span class="cat-desc">{{ tag.desc }}</span>
              </div>
              <div class="cat-glow"></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 应用列表 / App Grid -->
    <section class="apps-section">
      <div class="apps-header">
        <h2 class="section-title">
          <span class="title-decor"></span>
          {{ currentCategoryName }}
        </h2>
        <div class="view-toggle">
          <button
            class="view-btn"
            :class="{ active: viewMode === 'grid' }"
            @click="viewMode = 'grid'"
            title="网格视图"
          >
            <el-icon><Grid /></el-icon>
          </button>
          <button
            class="view-btn"
            :class="{ active: viewMode === 'list' }"
            @click="viewMode = 'list'"
            title="列表视图"
          >
            <el-icon><List /></el-icon>
          </button>
        </div>
      </div>

      <!-- 加载状态 / Loading -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner">
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
        </div>
        <p>正在加载应用...</p>
      </div>

      <!-- 空状态 / Empty -->
      <el-empty
        v-else-if="workflows.length === 0"
        class="empty-state"
        description="暂无可用应用"
      >
        <template #image>
          <div class="empty-illustration">
            <el-icon :size="80"><Box /></el-icon>
          </div>
        </template>
      </el-empty>

      <!-- 应用卡片网格 / Apps Grid -->
      <div v-else class="apps-grid" :class="`view-${viewMode}`">
        <article
          v-for="(workflow, index) in workflows"
          :key="workflow.id"
          class="app-card"
          :class="[`cat-${workflow.category || 'default'}`]"
          :style="{ animationDelay: `${index * 0.05}s` }"
          @click="useWorkflow(workflow)"
        >
          <!-- 卡片光效 / Card glow -->
          <div class="card-glow"></div>

          <!-- 头部 / Header -->
          <header class="card-top">
            <div class="app-icon">
              <div class="icon-bg"></div>
              <el-icon :size="26"><component :is="getCategoryIcon(workflow.category)" /></el-icon>
              <div class="icon-shine"></div>
            </div>
            <div class="app-meta-tags">
              <span v-if="workflow.isHot" class="hot-pin">
                <el-icon><StarFilled /></el-icon>
                HOT
              </span>
              <span v-if="workflow.isNew" class="new-pin">NEW</span>
            </div>
          </header>

          <!-- 内容 / Content -->
          <div class="card-content">
            <h3 class="app-title">{{ workflow.name }}</h3>
            <p class="app-desc">{{ workflow.description || '暂无描述' }}</p>
          </div>

          <!-- 标签 / Tags -->
          <div class="card-tags">
            <span class="tag-pill">
              <el-icon :size="12"><Folder /></el-icon>
              {{ getCategoryLabel(workflow.category) }}
            </span>
            <span v-if="workflow.model" class="tag-pill">
              <el-icon :size="12"><Cpu /></el-icon>
              {{ workflow.model }}
            </span>
          </div>

          <!-- 底部 / Footer -->
          <footer class="card-bottom">
            <div class="app-stats">
              <div class="stat-item" v-if="workflow.usageCount !== undefined">
                <el-icon :size="14"><View /></el-icon>
                <span>{{ formatCount(workflow.usageCount) }}</span>
              </div>
              <div class="stat-item" v-if="workflow.rating !== undefined">
                <el-icon :size="14" class="star-icon"><Star /></el-icon>
                <span>{{ workflow.rating.toFixed ? workflow.rating.toFixed(1) : workflow.rating }}</span>
              </div>
            </div>
            <button class="run-btn" @click.stop="useWorkflow(workflow)">
              <span>立即使用</span>
              <el-icon><ArrowRight /></el-icon>
            </button>
          </footer>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, ArrowRight, Grid, Star, Tools, Edit, Monitor, Reading, House, More,
  Close, StarFilled, Box, List, View, Cpu, Folder, ChatLineRound, Document,
  Picture, Promotion, DataAnalysis
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'

defineOptions({
  name: 'WorkflowMarket'
})

const router = useRouter()

const searchQuery = ref('')
const searchFocused = ref(false)
const currentCategory = ref('all')
const workflows = ref([])
const loading = ref(false)
const viewMode = ref('grid')

// 分类配置 / Categories
const categories = [
  { id: 'all',      name: '全部',     icon: 'Grid',         desc: '查看所有' },
  { id: 'popular',  name: '热门推荐', icon: 'StarFilled',   desc: '精选热门' },
  { id: 'chat',     name: '智能对话', icon: 'ChatLineRound', desc: '聊天助手' },
  { id: 'writing',  name: '文案创作', icon: 'Edit',          desc: '写作辅助' },
  { id: 'tools',    name: '实用工具', icon: 'Tools',         desc: '提效利器' },
  { id: 'coding',   name: '编码专家', icon: 'Monitor',       desc: '代码生成' },
  { id: 'learning', name: '知识学习', icon: 'Reading',       desc: '学习伴侣' },
  { id: 'image',    name: '图像处理', icon: 'Picture',       desc: '图片生成' },
  { id: 'analysis', name: '数据分析', icon: 'DataAnalysis',  desc: '数据洞察' },
  { id: 'life',     name: '生活指南', icon: 'House',         desc: '生活助手' },
  { id: 'other',    name: '其他',     icon: 'More',          desc: '更多分类' }
]

const hotSearches = ['代码生成', '智能写作', '翻译助手', 'PPT 生成', '数据分析']

// 图标映射 / Icon mapping
const iconMap = {
  Grid: Grid,
  Star: Star,
  StarFilled: StarFilled,
  Tools: Tools,
  Edit: Edit,
  Monitor: Monitor,
  Reading: Reading,
  House: House,
  More: More,
  ChatLineRound: ChatLineRound,
  Document: Document,
  Picture: Picture,
  Promotion: Promotion,
  DataAnalysis: DataAnalysis
}

const getCategoryIcon = (category) => {
  const map = {
    popular: StarFilled,
    chat: ChatLineRound,
    writing: Edit,
    tools: Tools,
    coding: Monitor,
    learning: Reading,
    image: Picture,
    analysis: DataAnalysis,
    life: House
  }
  return iconMap[map[category] || 'Grid']
}

const getCategoryLabel = (category) => {
  const map = {
    life: '生活',
    tools: '工具',
    writing: '创作',
    coding: '编码',
    learning: '学习',
    chat: '对话',
    image: '图像',
    analysis: '分析',
    popular: '热门',
    other: '其他'
  }
  return map[category] || category
}

const currentCategoryName = computed(() => {
  const cat = categories.find(c => c.id === currentCategory.value)
  return cat ? cat.name : '全部应用'
})

const formatCount = (count) => {
  if (count >= 10000) return `${(count / 10000).toFixed(1)}w`
  if (count >= 1000) return `${(count / 1000).toFixed(1)}k`
  return count
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
      ElMessage.error(res.msg || '获取应用列表失败')
    }
  } catch (error) {
    console.error('Load workflows error:', error)
    ElMessage.error('获取应用列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadWorkflows()
}

const quickSearch = (keyword) => {
  searchQuery.value = keyword
  handleSearch()
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
// ====================
// 主题变量 / Theme Variables
// ====================
$primary: var(--el-color-primary, #409EFF);
$primary-rgb: var(--el-color-primary-rgb, 64, 158, 255);

// 渐变色板 / Gradient Palette
$grad-blue: linear-gradient(135deg, #2B5EFF 0%, #4facfe 100%);
$grad-purple: linear-gradient(135deg, #8B5CF6 0%, #EC4899 100%);
$grad-cyan: linear-gradient(135deg, #06B6D4 0%, #3B82F6 100%);
$grad-orange: linear-gradient(135deg, #F59E0B 0%, #EF4444 100%);
$grad-green: linear-gradient(135deg, #10B981 0%, #34D399 100%);
$grad-pink: linear-gradient(135deg, #EC4899 0%, #F472B6 100%);
$grad-amber: linear-gradient(135deg, #F59E0B 0%, #FBBF24 100%);
$grad-indigo: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%);
$grad-teal: linear-gradient(135deg, #14B8A6 0%, #06B6D4 100%);
$grad-rose: linear-gradient(135deg, #F43F5E 0%, #FB7185 100%);
$grad-violet: linear-gradient(135deg, #A78BFA 0%, #C084FC 100%);

// 分类颜色映射 / Category Color Mapping
$cat-colors: (
  popular: $grad-orange,
  chat: $grad-blue,
  writing: $grad-purple,
  tools: $grad-amber,
  coding: $grad-cyan,
  learning: $grad-green,
  image: $grad-pink,
  analysis: $grad-indigo,
  life: $grad-teal,
  other: $grad-violet,
  default: $grad-blue
);

// ====================
// 容器 / Container
// ====================
.workflow-market-container {
  min-height: 100vh;
  background: var(--el-bg-color-page, #f5f7fa);
  position: relative;
  overflow-x: hidden;
}

// ====================
// Hero 区域
// ====================
.hero-section {
  position: relative;
  padding: 80px 24px 120px;
  overflow: hidden;
  isolation: isolate;
}

.hero-bg {
  position: absolute;
  inset: 0;
  z-index: -1;
  overflow: hidden;

  // 极光效果 / Aurora effect
  .hero-aurora {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.5;
    animation: aurora-float 20s ease-in-out infinite;

    &.aurora-1 {
      width: 600px;
      height: 600px;
      background: radial-gradient(circle, rgba(43, 94, 255, 0.4), transparent 70%);
      top: -200px;
      left: -100px;
      animation-delay: 0s;
    }

    &.aurora-2 {
      width: 500px;
      height: 500px;
      background: radial-gradient(circle, rgba(139, 92, 246, 0.35), transparent 70%);
      top: -100px;
      right: -100px;
      animation-delay: -7s;
    }

    &.aurora-3 {
      width: 400px;
      height: 400px;
      background: radial-gradient(circle, rgba(236, 72, 153, 0.25), transparent 70%);
      bottom: -100px;
      left: 50%;
      animation-delay: -14s;
    }
  }

  // 网格背景 / Grid background
  .hero-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(var(--divider-rgb, 0, 0, 0), 0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(var(--divider-rgb, 0, 0, 0), 0.04) 1px, transparent 1px);
    background-size: 60px 60px;
    mask-image: radial-gradient(ellipse at center, #000 30%, transparent 80%);
    -webkit-mask-image: radial-gradient(ellipse at center, #000 30%, transparent 80%);
  }

  // 噪点纹理 / Noise texture
  .hero-noise {
    position: absolute;
    inset: 0;
    opacity: 0.03;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
  }
}

@keyframes aurora-float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33%      { transform: translate(30px, -30px) scale(1.1); }
  66%      { transform: translate(-20px, 20px) scale(0.95); }
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 760px;
  margin: 0 auto;
  text-align: center;
}

// Hero Badge
.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  margin-bottom: 24px;
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-regular);
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 999px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  backdrop-filter: blur(8px);

  .badge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: linear-gradient(135deg, #10B981, #34D399);
    box-shadow: 0 0 8px rgba(16, 185, 129, 0.6);
    animation: pulse-dot 2s ease-in-out infinite;
  }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%      { opacity: 0.6; transform: scale(1.2); }
}

// Hero Title
.hero-title {
  margin: 0 0 20px;
  font-size: 3.4em;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--el-text-color-primary);

  .title-line {
    display: block;
  }

  .gradient-text {
    background: linear-gradient(135deg, #2B5EFF 0%, #8B5CF6 50%, #EC4899 100%);
    background-size: 200% auto;
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    color: transparent;
    animation: gradient-shift 6s ease-in-out infinite;
  }
}

@keyframes gradient-shift {
  0%, 100% { background-position: 0% center; }
  50%      { background-position: 100% center; }
}

.hero-subtitle {
  margin: 0 auto 40px;
  max-width: 580px;
  font-size: 1.05em;
  line-height: 1.7;
  color: var(--el-text-color-secondary);
}

// Search Box
.search-wrapper {
  max-width: 600px;
  margin: 0 auto;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 56px;
  padding: 6px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  border-radius: 18px;
  box-shadow:
    0 4px 24px rgba(0, 0, 0, 0.04),
    0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &.focused {
    border-color: $primary;
    box-shadow:
      0 0 0 4px rgba($primary-rgb, 0.08),
      0 8px 32px rgba(0, 0, 0, 0.08);
    transform: translateY(-2px);
  }

  .search-icon {
    margin-left: 12px;
    font-size: 18px;
    color: var(--el-text-color-secondary);
  }

  .search-input {
    flex: 1;
    height: 100%;
    padding: 0 8px;
    font-size: 15px;
    color: var(--el-text-color-primary);
    background: transparent;
    border: none;
    outline: none;

    &::placeholder {
      color: var(--el-text-color-placeholder);
    }
  }

  .clear-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border: none;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      color: var(--el-text-color-primary);
      background: var(--el-fill-color);
    }
  }

  .search-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 100%;
    padding: 0 20px;
    font-size: 14px;
    font-weight: 500;
    color: #fff;
    background: linear-gradient(135deg, #2B5EFF 0%, #4facfe 100%);
    border: none;
    border-radius: 14px;
    cursor: pointer;
    transition: all 0.25s ease;
    box-shadow: 0 4px 12px rgba(43, 94, 255, 0.25);

    &:hover {
      transform: scale(1.02);
      box-shadow: 0 6px 18px rgba(43, 94, 255, 0.35);
    }

    &:active {
      transform: scale(0.98);
    }
  }
}

// Hot Tags
.hot-tags {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 18px;
  font-size: 13px;

  .hot-label {
    color: var(--el-text-color-secondary);
  }

  .hot-tag {
    padding: 4px 12px;
    color: var(--el-text-color-regular);
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      color: $primary;
      border-color: $primary;
      background: rgba($primary-rgb, 0.04);
      transform: translateY(-1px);
    }
  }
}

// ====================
// 分类导航
// ====================
.category-section {
  margin-top: -60px;
  padding: 0 24px;
  position: relative;
  z-index: 5;
}

.category-wrapper {
  max-width: 1200px;
  margin: 0 auto;
}

.category-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 0 4px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
  font-size: 1.4em;
  font-weight: 700;
  color: var(--el-text-color-primary);

  .title-decor {
    width: 4px;
    height: 20px;
    background: linear-gradient(135deg, #2B5EFF 0%, #8B5CF6 100%);
    border-radius: 2px;
  }
}

.category-count {
  font-size: 13px;
  color: var(--el-text-color-secondary);

  &::before {
    content: '·';
    margin-right: 6px;
    color: $primary;
  }
}

.category-scroll {
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 4px;

  &::-webkit-scrollbar {
    height: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color);
    border-radius: 2px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

.category-nav {
  display: flex;
  gap: 12px;
  padding: 4px;
  min-width: max-content;
}

.category-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  min-width: 140px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;

  .cat-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    color: #fff;
    background: $grad-blue;
    border-radius: 12px;
    transition: transform 0.3s ease;
    flex-shrink: 0;
  }

  .cat-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .cat-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    line-height: 1.2;
  }

  .cat-desc {
    font-size: 11px;
    color: var(--el-text-color-secondary);
    line-height: 1.2;
  }

  .cat-glow {
    position: absolute;
    inset: 0;
    border-radius: 14px;
    opacity: 0;
    transition: opacity 0.3s ease;
    pointer-events: none;
    background: radial-gradient(circle at center, rgba($primary-rgb, 0.08), transparent 70%);
  }

  &:hover {
    transform: translateY(-2px);
    border-color: rgba($primary-rgb, 0.3);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);

    .cat-icon {
      transform: scale(1.08) rotate(-5deg);
    }
  }

  &.active {
    border-color: transparent;
    box-shadow: 0 8px 24px rgba(43, 94, 255, 0.18);

    .cat-glow {
      opacity: 1;
    }

    .cat-icon {
      background: linear-gradient(135deg, #2B5EFF 0%, #8B5CF6 100%);
    }
  }
}

// 分类图标颜色 / Category icon colors
@each $name, $gradient in $cat-colors {
  .category-item.cat-#{$name} .cat-icon {
    background: $gradient;
  }
}

// ====================
// 应用列表
// ====================
.apps-section {
  max-width: 1200px;
  padding: 48px 24px 80px;
  margin: 0 auto;
}

.apps-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.view-toggle {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--el-fill-color-light);
  border-radius: 10px;

  .view-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 28px;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      color: var(--el-text-color-primary);
    }

    &.active {
      color: $primary;
      background: var(--el-bg-color);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    }
  }
}

// 加载状态 / Loading state
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: 20px;
  color: var(--el-text-color-secondary);

  .loading-spinner {
    position: relative;
    width: 60px;
    height: 60px;
  }

  .spinner-ring {
    position: absolute;
    inset: 0;
    border: 3px solid transparent;
    border-top-color: $primary;
    border-radius: 50%;
    animation: spin 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;

    &:nth-child(2) {
      inset: 8px;
      border-top-color: #8B5CF6;
      animation-duration: 1.6s;
      animation-direction: reverse;
    }

    &:nth-child(3) {
      inset: 16px;
      border-top-color: #EC4899;
      animation-duration: 2s;
    }
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// 空状态 / Empty state
.empty-state {
  padding: 60px 0;

  .empty-illustration {
    color: var(--el-text-color-placeholder);
    opacity: 0.5;
  }
}

// 应用网格 / Apps grid
.apps-grid {
  display: grid;
  gap: 20px;

  &.view-grid {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  }

  &.view-list {
    grid-template-columns: 1fr;
  }
}

// 应用卡片 / App card
.app-card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 22px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 18px;
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  animation: card-enter 0.6s cubic-bezier(0.16, 1, 0.3, 1) backwards;

  // 顶部光效 / Top glow
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 80px;
    opacity: 0;
    transition: opacity 0.4s ease;
    pointer-events: none;
    z-index: 0;
  }

  // 悬停光斑 / Hover glow
  .card-glow {
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    opacity: 0;
    transition: opacity 0.4s ease;
    pointer-events: none;
    background: radial-gradient(
      circle at var(--mouse-x, 50%) var(--mouse-y, 0%),
      rgba($primary-rgb, 0.08),
      transparent 40%
    );
  }

  &:hover {
    transform: translateY(-6px);
    border-color: rgba($primary-rgb, 0.2);
    box-shadow:
      0 16px 48px rgba(0, 0, 0, 0.08),
      0 4px 12px rgba(0, 0, 0, 0.04);

    .card-glow {
      opacity: 1;
    }

    .run-btn {
      background: linear-gradient(135deg, #2B5EFF 0%, #4facfe 100%);
      color: #fff;
      box-shadow: 0 6px 18px rgba(43, 94, 255, 0.3);

      .el-icon {
        transform: translateX(4px);
      }
    }
  }

  // 列表视图样式 / List view style
  .view-list & {
    flex-direction: row;
    align-items: center;
    gap: 20px;
    padding: 18px 24px;

    .card-top {
      flex-shrink: 0;
      margin-bottom: 0;
    }

    .card-content {
      flex: 1;
      margin-bottom: 0;
    }

    .card-tags {
      flex-shrink: 0;
    }

    .card-bottom {
      flex-shrink: 0;
      padding-top: 0;
      border-top: none;
    }
  }
}

@keyframes card-enter {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

// 卡片顶部 / Card top
.card-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.app-icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  color: #fff;
  background: $grad-blue;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 6px 16px rgba(43, 94, 255, 0.25);
  flex-shrink: 0;

  .icon-bg {
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.15), transparent 60%);
  }

  .icon-shine {
    position: absolute;
    top: -50%;
    left: -50%;
    width: 50%;
    height: 200%;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
    transform: skewX(-20deg);
    animation: shine 4s ease-in-out infinite;
  }
}

@keyframes shine {
  0%, 100% { transform: translateX(0) skewX(-20deg); }
  50%      { transform: translateX(300%) skewX(-20deg); }
}

// 分类图标颜色
@each $name, $gradient in $cat-colors {
  .app-card.cat-#{$name} .app-icon {
    background: $gradient;
  }
}

.app-meta-tags {
  display: flex;
  gap: 6px;

  .hot-pin,
  .new-pin {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 8px;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.5px;
    border-radius: 999px;
  }

  .hot-pin {
    color: #fff;
    background: linear-gradient(135deg, #F59E0B 0%, #EF4444 100%);
    box-shadow: 0 2px 8px rgba(245, 158, 11, 0.35);
  }

  .new-pin {
    color: #fff;
    background: linear-gradient(135deg, #10B981 0%, #34D399 100%);
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.35);
  }
}

// 卡片内容 / Card content
.card-content {
  position: relative;
  z-index: 1;
  flex: 1;
  margin-bottom: 14px;
}

.app-title {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 700;
  line-height: 1.3;
  color: var(--el-text-color-primary);
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.app-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 42px;
}

// 卡片标签 / Card tags
.card-tags {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 14px;
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  font-size: 11px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-radius: 999px;
  transition: all 0.2s ease;
}

.app-card:hover .tag-pill {
  color: var(--el-text-color-regular);
}

// 卡片底部 / Card bottom
.card-bottom {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 14px;
  border-top: 1px dashed var(--el-border-color-lighter);
}

.app-stats {
  display: flex;
  gap: 12px;

  .stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);

    .star-icon {
      color: #F59E0B;
    }
  }
}

.run-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 600;
  color: $primary;
  background: rgba($primary-rgb, 0.08);
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;

  .el-icon {
    transition: transform 0.3s ease;
  }

  &:hover {
    background: rgba($primary-rgb, 0.12);
  }
}

// ====================
// 暗色模式适配
// ====================
:global(.dark) {
  .hero-section {
    background: linear-gradient(180deg, #0a0a0a 0%, #141414 100%);
  }

  .hero-bg {
    .hero-grid {
      background-image:
        linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
    }

    .hero-aurora {
      opacity: 0.3;
    }
  }

  .search-box,
  .category-item,
  .app-card {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(12px);

    &:hover {
      background: rgba(255, 255, 255, 0.05);
      border-color: rgba(255, 255, 255, 0.12);
    }
  }

  .hero-badge {
    background: rgba(255, 255, 255, 0.04);
    border-color: rgba(255, 255, 255, 0.08);
  }

  .hot-tag {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.08);

    &:hover {
      background: rgba(64, 158, 255, 0.1);
    }
  }

  .tag-pill {
    background: rgba(255, 255, 255, 0.06);
  }
}

// ====================
// 响应式 / Responsive
// ====================
@media (max-width: 768px) {
  .hero-section {
    padding: 50px 16px 90px;
  }

  .hero-title {
    font-size: 2.2em;

    .title-line {
      display: inline;
    }
  }

  .hero-subtitle {
    font-size: 0.95em;
    margin-bottom: 28px;
  }

  .search-box {
    height: 50px;
    border-radius: 14px;

    .search-btn {
      padding: 0 14px;

      span {
        display: none;
      }
    }
  }

  .category-section {
    margin-top: -40px;
    padding: 0 12px;
  }

  .category-item {
    padding: 12px 14px;
    min-width: 120px;

    .cat-icon {
      width: 36px;
      height: 36px;
    }

    .cat-name {
      font-size: 13px;
    }

    .cat-desc {
      font-size: 10px;
    }
  }

  .apps-section {
    padding: 32px 12px 60px;
  }

  .section-title {
    font-size: 1.2em;
  }

  .apps-grid {
    gap: 14px;

    &.view-grid {
      grid-template-columns: 1fr;
    }
  }

  .app-card {
    padding: 18px;
  }
}

@media (max-width: 480px) {
  .hero-badge {
    font-size: 12px;
    padding: 5px 12px;
  }

  .app-icon {
    width: 46px;
    height: 46px;
  }
}
</style>