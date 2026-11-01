<template>
  <div class="workflow-market-container">
    <!-- 顶部标题区：扁平、紧凑，不用极光/网格/渐变字 -->
    <!-- Top header: flat and compact — no aurora, grid or gradient text -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="hero-heading">
          <div>
            <h1 class="hero-title">AI 应用市场</h1>
            <p class="hero-subtitle">即插即用的智能工作流，从内容创作到代码生成</p>
          </div>
          <span class="hero-count">{{ workflows.length }} 个应用</span>
        </div>

        <div class="search-wrapper">
          <div class="search-box" :class="{ focused: searchFocused }">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
              v-model="searchQuery"
              class="search-input"
              placeholder="搜索你需要的 AI 应用"
              clearable
              @keyup.enter="handleSearch"
              @focus="searchFocused = true"
              @blur="searchFocused = false"
            />
            <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
              <el-icon><Close /></el-icon>
            </button>
            <button class="search-btn" @click="handleSearch">搜索</button>
          </div>
          <div class="hot-tags">
            <span class="hot-label">热门</span>
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

    <!-- 分类导航：紧凑 pill，选中态只有主色，不再每类一个彩色图标 -->
    <!-- Category nav: compact pills; the selected one only uses the brand color -->
    <section class="category-section">
      <div class="category-wrapper">
        <div class="category-scroll">
          <div class="category-nav">
            <button
              v-for="tag in categories"
              :key="tag.id"
              class="category-item"
              :class="{ active: currentCategory === tag.id }"
              @click="selectCategory(tag.id)"
            >
              <el-icon :size="14"><component :is="tag.icon" /></el-icon>
              <span>{{ tag.name }}</span>
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- 应用列表 / App Grid -->
    <section class="apps-section">
      <div class="apps-header">
        <h2 class="section-title">{{ currentCategoryName }}</h2>
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
        <el-icon class="is-loading" :size="22"><Loading /></el-icon>
        <span>正在加载应用…</span>
      </div>

      <!-- 空状态 / Empty -->
      <el-empty
        v-else-if="workflows.length === 0"
        class="empty-state"
        description="暂无可用应用"
      >
        <template #image>
          <div class="empty-illustration">
            <el-icon :size="56"><Box /></el-icon>
          </div>
        </template>
      </el-empty>

      <!-- 应用卡片网格 / Apps Grid -->
      <div v-else class="apps-grid" :class="`view-${viewMode}`">
        <article
          v-for="workflow in workflows"
          :key="workflow.id"
          class="app-card"
          @click="useWorkflow(workflow)"
        >
          <!-- 头部 / Header -->
          <header class="card-top">
            <div class="app-icon">
              <el-icon :size="20"><component :is="getCategoryIcon(workflow.category)" /></el-icon>
            </div>
            <div class="app-meta-tags">
              <span v-if="workflow.isHot" class="hot-pin">热门</span>
              <span v-if="workflow.isNew" class="new-pin">新</span>
            </div>
          </header>

          <!-- 内容 / Content -->
          <div class="card-content">
            <h3 class="app-title">{{ workflow.name }}</h3>
            <p class="app-desc">{{ workflow.description || '暂无描述' }}</p>
          </div>

          <!-- 标签 / Tags（无分类时不渲染空占位，避免卡片高度参差） -->
          <div class="card-tags">
            <span v-if="workflow.category" class="tag-pill">{{ getCategoryLabel(workflow.category) }}</span>
            <span v-if="workflow.model" class="tag-pill">{{ workflow.model }}</span>
          </div>

          <!-- 底部 / Footer -->
          <footer class="card-bottom">
            <div class="app-stats">
              <div class="stat-item" v-if="workflow.usageCount !== undefined">
                <el-icon :size="13"><View /></el-icon>
                <span>{{ formatCount(workflow.usageCount) }}</span>
              </div>
              <div class="stat-item" v-if="workflow.rating !== undefined">
                <el-icon :size="13" class="star-icon"><Star /></el-icon>
                <span>{{ workflow.rating.toFixed ? workflow.rating.toFixed(1) : workflow.rating }}</span>
              </div>
            </div>
            <span class="run-link">
              使用
              <el-icon><ArrowRight /></el-icon>
            </span>
          </footer>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, ArrowRight, Grid, Star, Tools, Edit, Monitor, Reading, House, More,
  Close, StarFilled, Box, List, View, Loading, ChatLineRound, Document,
  Picture, Promotion, DataAnalysis
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import { useThemeStore } from '@/stores/theme'
import { authResolvedTick, isAuthError } from '@/utils/authPrompt'

defineOptions({
  name: 'WorkflowMarket'
})

const router = useRouter()
const themeStore = useThemeStore()

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
  // 映射表存「组件名字符串」再查 iconMap；直接存组件对象会让 iconMap[对象] 恒为 undefined（图标空白）
  // Keep strings here: indexing iconMap with a component object always yields undefined (blank icons)
  const map = {
    popular: 'StarFilled',
    chat: 'ChatLineRound',
    writing: 'Edit',
    tools: 'Tools',
    coding: 'Monitor',
    learning: 'Reading',
    image: 'Picture',
    analysis: 'DataAnalysis',
    life: 'House'
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
    // 登录类错误不重复报错：全局登录弹窗本身就是提示，登录成功后自动重载
    // auth errors stay quiet: the global login dialog is the message, and we refetch on success
    if (!isAuthError(error)) {
      ElMessage.error('获取应用列表失败')
    }
  } finally {
    loading.value = false
  }
}

// 全局登录弹窗登录成功后，自动重载本页数据（无需用户手动刷新）
// after a successful re-login via the global dialog, reload automatically
watch(authResolvedTick, () => {
  loadWorkflows()
})

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
  // 应用用户主题偏好：本页此前不调用，深色偏好直开/刷新时失效（与其他公开页一致）
  // Apply the saved theme; without this a direct load ignores the dark preference
  themeStore.initTheme()
  loadWorkflows()
})
</script>

<style lang="scss" scoped>
// ====================
// 视觉基调 / Visual language
// 去装饰化：无渐变文字、无极光/网格/噪点、无光晕与流光、无每类一色的图标；
// 只用品牌色 + 中性灰阶，靠排版、间距与边框建立层次
// Decoration-free: no gradient text, aurora, grid, noise, glows or shine, and no
// per-category colors — hierarchy comes from type, spacing and hairlines instead
// ====================

.workflow-market-container {
  min-height: 100vh;
  // 页面底色与顶栏/卡片的白面拉开一层，但保持中性
  // Neutral page ground, one step below the white header and cards
  background: var(--el-bg-color-page, #f7f8fa);
  color: var(--el-text-color-primary);
}

// ====================
// 标题区 / Header
// ====================
.hero-section {
  background: var(--el-bg-color);
}

.hero-content {
  max-width: 1120px;
  margin: 0 auto;
  padding: 46px 24px 30px;
}

.hero-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.hero-title {
  margin: 0 0 7px;
  font-size: 30px;
  font-weight: 700;
  letter-spacing: -0.4px;
  line-height: 1.22;
  color: var(--el-text-color-primary);
}

.hero-subtitle {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.hero-count {
  padding: 3px 11px;
  font-size: 12.5px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-radius: 999px;
  white-space: nowrap;
}

// 搜索：扁平输入框 + 实心按钮，聚焦只变边框与一圈极淡的主色环
// Search: flat field + solid button; focus only shifts the border and adds a faint ring
.search-wrapper {
  max-width: 560px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 48px;
  padding: 0 4px 0 14px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;

  &.focused {
    border-color: var(--el-color-primary);
    box-shadow:
      0 0 0 3px color-mix(in srgb, var(--el-color-primary) 12%, transparent),
      0 1px 3px rgba(0, 0, 0, 0.04);
  }

  .search-icon {
    font-size: 17px;
    color: var(--el-text-color-placeholder);
  }

  .search-input {
    flex: 1;
    height: 100%;
    padding: 0 8px;
    font-size: 14px;
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
    width: 22px;
    height: 22px;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 50%;
    cursor: pointer;
    transition: color 0.18s ease, background 0.18s ease;

    &:hover {
      color: var(--el-text-color-primary);
      background: var(--el-fill-color-light);
    }
  }

  .search-btn {
    height: 38px;
    padding: 0 20px;
    font-size: 13.5px;
    font-weight: 500;
    color: #fff;
    background: var(--el-color-primary);
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: opacity 0.18s ease;

    &:hover {
      opacity: 0.88;
    }
  }
}

.hot-tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 14px;
  font-size: 13px;

  .hot-label {
    margin-right: 2px;
    color: var(--el-text-color-placeholder);
  }

  .hot-tag {
    padding: 2px 9px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border-radius: 6px;
    cursor: pointer;
    transition: color 0.18s ease, background 0.18s ease;

    &:hover {
      color: var(--el-color-primary);
      background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
    }
  }
}

// ====================
// 分类导航：灰底上的白 pill，选中用主色浅底
// Category nav: white pills on the page ground; the selected one is brand-tinted
// ====================
.category-section {
  padding-top: 24px;
}

.category-wrapper {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
}

.category-scroll {
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    height: 0;
  }
}

.category-nav {
  display: flex;
  gap: 8px;
  min-width: max-content;
}

.category-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 999px;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.18s ease, border-color 0.18s ease, background 0.18s ease;

  .el-icon {
    color: var(--el-text-color-placeholder);
    transition: color 0.18s ease;
  }

  &:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-7);

    .el-icon {
      color: var(--el-color-primary);
    }
  }

  &.active {
    color: var(--el-color-primary);
    font-weight: 600;
    background: color-mix(in srgb, var(--el-color-primary) 9%, transparent);
    border-color: color-mix(in srgb, var(--el-color-primary) 28%, transparent);

    .el-icon {
      color: var(--el-color-primary);
    }
  }
}

// ====================
// 应用列表 / Apps
// ====================
.apps-section {
  max-width: 1120px;
  margin: 0 auto;
  padding: 28px 24px 64px;
}

.apps-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  padding: 0 2px;
}

.section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.view-toggle {
  display: flex;
  gap: 2px;

  .view-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 26px;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: color 0.18s ease, background 0.18s ease;

    &:hover {
      color: var(--el-text-color-primary);
      background: var(--el-fill-color-light);
    }

    &.active {
      color: var(--el-color-primary);
      background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
    }
  }
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 64px 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.empty-state {
  padding: 56px 0;

  .empty-illustration {
    color: var(--el-text-color-placeholder);
    opacity: 0.6;
  }
}

.apps-grid {
  display: grid;
  gap: 16px;

  &.view-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  }

  &.view-list {
    grid-template-columns: 1fr;
  }
}

// 卡片：白面 + 发丝边；悬停只上浮 2px 并加一层柔和阴影，不发光
// Cards: white surface + hairline; hover lifts 2px with a soft shadow, no glow
.app-card {
  display: flex;
  flex-direction: column;
  padding: 20px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;

  &:hover {
    transform: translateY(-2px);
    border-color: var(--el-color-primary-light-7);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.07);
  }

  // 列表视图：单行横排
  // List view: a single horizontal row
  .view-list & {
    flex-direction: row;
    align-items: center;
    gap: 18px;
    padding: 14px 18px;

    .card-top {
      flex-shrink: 0;
      margin-bottom: 0;
    }

    .card-content {
      flex: 1;
      margin-bottom: 0;
    }

    .app-desc {
      min-height: 0;
      -webkit-line-clamp: 1;
    }

    .card-tags {
      flex-shrink: 0;
      margin-bottom: 0;
    }

    .card-bottom {
      flex-shrink: 0;
      padding-top: 0;
      border-top: none;
    }
  }
}

.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 14px;
}

// 图标：统一的品牌色实底 + 白色线性图标，作为每张卡片的视觉锚点
// Icons: solid brand fill with a white glyph — one consistent anchor per card
.app-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  color: #fff;
  background: var(--el-color-primary);
  border-radius: 11px;
  box-shadow: 0 2px 6px color-mix(in srgb, var(--el-color-primary) 28%, transparent);
  flex-shrink: 0;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.app-card:hover .app-icon {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px color-mix(in srgb, var(--el-color-primary) 34%, transparent);
}

.app-meta-tags {
  display: flex;
  gap: 6px;

  .hot-pin,
  .new-pin {
    padding: 1px 7px;
    font-size: 11px;
    line-height: 1.6;
    border-radius: 4px;
  }

  .hot-pin {
    color: #b45309;
    background: rgba(245, 158, 11, 0.12);
  }

  .new-pin {
    color: #047857;
    background: rgba(16, 185, 129, 0.12);
  }
}

.card-content {
  flex: 1;
  margin-bottom: 12px;
}

.app-title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--el-text-color-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
  // 固定占位：没有标签的卡片也能与有标签的保持同一节奏
  // Reserved row so tag-less cards keep the same rhythm
  min-height: 22px;
}

.tag-pill {
  padding: 2px 8px;
  font-size: 11.5px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-radius: 5px;
}

.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.app-stats {
  display: flex;
  gap: 12px;

  .stat-item {
    display: flex;
    align-items: center;
    gap: 3px;
    font-size: 12px;
    color: var(--el-text-color-placeholder);

    .star-icon {
      color: #f59e0b;
    }
  }
}

// 「使用」：浅主色底的小按钮，悬停加深一档
// "Use": a small brand-tinted button that deepens on hover
.run-link {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 5px 11px;
  font-size: 12.5px;
  font-weight: 500;
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  border-radius: 7px;
  transition: background 0.18s ease;

  .el-icon {
    transition: transform 0.18s ease;
  }
}

.app-card:hover .run-link {
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent);

  .el-icon {
    transform: translateX(2px);
  }
}

// ====================
// 暗色模式：只调对比度，不引入第二套装饰
// Dark mode: contrast adjustments only, no second decoration pass
// 注意 scoped 下必须用 html.dark &，:global(.dark) 嵌套会坍缩成裸 .dark 糊到 html 上
// ====================
.workflow-market-container {
  html.dark & {
    .hero-section {
      background: #17181a;
    }

    .search-box,
    .app-card {
      background: #17181a;
      border-color: #2b2c30;
    }

    .app-card:hover {
      border-color: #3d3e43;
      box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
    }

    .category-item {
      background: #17181a;
      border-color: #2b2c30;
    }

    .tag-pill,
    .hot-tag {
      background: rgba(255, 255, 255, 0.06);
    }

    .hot-pin {
      color: #fbbf24;
      background: rgba(245, 158, 11, 0.16);
    }

    .new-pin {
      color: #34d399;
      background: rgba(16, 185, 129, 0.16);
    }
  }
}

// ====================
// 响应式 / Responsive
// ====================
@media (max-width: 768px) {
  .hero-content {
    padding: 28px 16px 22px;
  }

  .hero-title {
    font-size: 22px;
  }

  .hero-heading {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .hero-count {
    padding-bottom: 0;
  }

  .category-wrapper {
    padding: 0 16px;
  }

  .apps-section {
    padding: 22px 16px 48px;
  }

  .apps-grid {
    gap: 10px;

    &.view-grid {
      grid-template-columns: 1fr;
    }
  }

  // 窄屏下列表视图回到纵向，避免挤压
  .view-list .app-card {
    flex-direction: column;
    align-items: stretch;
    gap: 0;

    .card-top {
      margin-bottom: 12px;
    }

    .card-content {
      margin-bottom: 12px;
    }

    .card-tags {
      margin-bottom: 12px;
    }

    .card-bottom {
      padding-top: 12px;
      border-top: 1px solid var(--el-border-color-lighter);
    }
  }
}
</style>