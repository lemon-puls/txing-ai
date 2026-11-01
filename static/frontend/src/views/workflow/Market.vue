<template>
  <div class="workflow-market-container">
    <!-- 顶部标题区：品牌色氛围底（极淡光晕 + 网格渐隐），克制但有焦点 -->
    <!-- Top header: faint brand halo + fading grid — subtle but with a focal point -->
    <section class="hero-section">
      <div class="hero-bg" aria-hidden="true">
        <div class="hero-grid"></div>
      </div>
      <div class="hero-content">
        <div class="hero-heading">
          <div>
            <h1 class="hero-title"><span class="title-bar"></span>AI 应用市场</h1>
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
          v-for="(workflow, index) in workflows"
          :key="workflow.id"
          class="app-card"
          :style="{ animationDelay: `${index * 0.04}s` }"
          @click="useWorkflow(workflow)"
          @mousemove="onCardMove"
        >
          <!-- 跟随鼠标的品牌色光斑（8%，仅在悬停时出现） -->
          <!-- Pointer-following brand glow (8%, hover only) -->
          <span class="card-spot" aria-hidden="true"></span>

          <!-- 右下角水印：同分类的大号淡图标，悬停时微微回正放大 -->
          <!-- Corner watermark: oversized faint category glyph that levels up on hover -->
          <span class="card-watermark" aria-hidden="true">
            <el-icon><component :is="getCategoryIcon(workflow.category)" /></el-icon>
          </span>

          <!-- 头部：图标 + 标题/副行 + 右侧圆形箭头 CTA -->
          <!-- Header: icon + title/subline + a round arrow CTA -->
          <header class="card-top">
            <!-- 优先用后台上传的图标；未上传时回退确定性字母头像 -->
            <!-- Prefer the uploaded icon; fall back to the deterministic letter avatar -->
            <div
              v-if="workflow.icon"
              class="app-icon app-icon--image"
            >
              <img :src="workflow.icon" :alt="workflow.name" loading="lazy" />
            </div>
            <div v-else class="app-icon" :style="{ background: appAvatar(workflow).bg }">
              <span class="app-icon-ch">{{ appAvatar(workflow).ch }}</span>
            </div>
            <div class="card-headline">
              <h3 class="app-title">
                <span class="app-name">{{ workflow.name }}</span>
                <span v-if="workflow.isHot" class="hot-pin">热门</span>
                <span v-if="workflow.isNew" class="new-pin">新</span>
              </h3>
              <span v-if="workflow.category || workflow.model" class="app-sub">
                {{ getCategoryLabel(workflow.category) }}<template v-if="workflow.model"> · {{ workflow.model }}</template>
              </span>
            </div>
          </header>

          <!-- 描述 / Description -->
          <p class="app-desc">{{ workflow.description || '暂无描述' }}</p>

          <!-- 底部：仅有统计数据时渲染，避免空栏 -->
          <!-- Footer renders only when stats exist -->
          <footer
            v-if="workflow.usageCount !== undefined || workflow.rating !== undefined"
            class="card-bottom"
          >
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
          </footer>

          <!-- 圆形箭头 CTA：固定在卡片右下角，整卡可点时作为视觉提示 -->
          <!-- Round arrow CTA pinned to the card corner; the whole card is clickable -->
          <span class="card-arrow" aria-hidden="true">
            <el-icon><ArrowRight /></el-icon>
          </span>
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

// 字母头像：无图标数据时按应用名生成确定性头像（首字符 + 哈希选色），
// 参考主流市场（VS Code/Google Contacts/Raycast）的 letter-avatar 兜底。
// 同一应用永远同色（可记忆），不同应用大概率不同色（避免分类级重复）。
// Letter avatar: deterministic per-app fallback (initial + name-hash colour),
// the mainstream pattern when apps ship no icon of their own
const appAvatar = (workflow) => {
  const name = String(workflow.name || '').trim()
  const chars = Array.from(name)
  const first = chars[0] || 'A'
  const isLatin = /^[a-z]$/i.test(first)
  const ch = isLatin ? first.toUpperCase() : first
  // 种子从首字符出发并叠加次字符：让同前缀（如多个 "AI …"）的应用也大概率散开
  // Seed from the first char and fold in the second, so same-prefix names
  // (e.g. several "AI …") still spread across the hue band
  let h = (first.codePointAt(0) * 131) >>> 0
  if (chars.length > 1) h = (h * 31 + chars[1].codePointAt(0)) >>> 0
  for (let i = 2; i < chars.length; i++) h = (h * 31 + chars[i].codePointAt(0)) >>> 0
  // 色相限定在青→蓝→靛→紫冷色带（185°-305°，16 档）并叠加两档亮度，
  // 共 32 种组合：与品牌蓝同族不花哨，几十个应用也几乎不撞色，白字对比达标
  // Hue spread over a 16-step cool band plus two lightness steps (32 combos):
  // harmonious with the brand colour; dozens of apps rarely share a look
  // 色带围绕品牌蓝（hue≈225）两侧展开：185°-265°，青→蓝→靛，
  // 越过 265° 就开始发紫、与品牌蓝脱钩，故不外扩
  // band centred on the brand blue: 185°-265° (teal→blue→indigo);
  // past ~265° it turns purple and drifts away from the brand hue
  const hue = 185 + (h % 9) * 10
  const light = 38 - ((h >>> 4) % 2) * 3
  return { ch, bg: `hsl(${hue}, 62%, ${light}%)` }
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

// 卡片光斑跟随指针：只写 CSS 变量，样式侧消费，避免每帧触发 Vue 更新
// Track the pointer with CSS variables only — no reactive updates per frame
const onCardMove = (e) => {
  const el = e.currentTarget
  const rect = el.getBoundingClientRect()
  el.style.setProperty('--mx', `${e.clientX - rect.left}px`)
  el.style.setProperty('--my', `${e.clientY - rect.top}px`)
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
  position: relative;
  overflow: hidden;
  background: var(--el-bg-color);
}

// 品牌氛围底：顶部中心的极淡光晕 + 向下渐隐的细网格
// Brand ambience: faint halo at the top centre plus a fading fine grid
.hero-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(
    620px 280px at 50% -80px,
    color-mix(in srgb, var(--el-color-primary) 17%, transparent),
    transparent 72%
  );

  .hero-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(color-mix(in srgb, var(--el-color-primary) 9%, transparent) 1px, transparent 1px),
      linear-gradient(90deg, color-mix(in srgb, var(--el-color-primary) 9%, transparent) 1px, transparent 1px);
    background-size: 48px 48px;
    mask-image: radial-gradient(ellipse 620px 300px at 50% 0%, #000 15%, transparent 78%);
    -webkit-mask-image: radial-gradient(ellipse 620px 300px at 50% 0%, #000 15%, transparent 78%);
  }
}

.hero-content {
  position: relative;
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
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 7px;
  font-size: 30px;
  font-weight: 700;
  letter-spacing: -0.4px;
  line-height: 1.22;
  color: var(--el-text-color-primary);

  // 品牌色竖条：纯色而非渐变，作为标题的视觉起点
  // Solid brand bar (not a gradient) anchoring the title
  .title-bar {
    width: 4px;
    height: 26px;
    border-radius: 2px;
    background: var(--el-color-primary);
    box-shadow: 0 0 12px color-mix(in srgb, var(--el-color-primary) 45%, transparent);
  }
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
  transition: color 0.18s ease, border-color 0.18s ease, background 0.18s ease,
    transform 0.18s ease, box-shadow 0.18s ease;

  .el-icon {
    color: var(--el-text-color-placeholder);
    transition: color 0.18s ease;
  }

  &:hover {
    transform: translateY(-1px);
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-7);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .el-icon {
      color: var(--el-color-primary);
    }
  }

  // 选中：主色实底白字，成为分类行里唯一的强对比焦点
  // Selected: solid brand fill with white text — the single strong contrast in the row
  &.active {
    color: #fff;
    font-weight: 600;
    background: var(--el-color-primary);
    border-color: var(--el-color-primary);
    box-shadow: 0 2px 10px color-mix(in srgb, var(--el-color-primary) 32%, transparent);

    .el-icon {
      color: #fff;
    }

    &:hover {
      color: #fff;
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
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }

  &.view-list {
    grid-template-columns: 1fr;
  }
}

// 卡片：白面 + 发丝边；悬停上浮 2px、顶部扫过一道品牌色亮线，光斑跟随指针
// Cards: white surface + hairline; hover lifts 2px, a brand light sweeps the top edge
// and a pointer-following glow becomes visible
.app-card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 20px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  cursor: pointer;
  overflow: hidden;
  animation: card-in 0.45s cubic-bezier(0.16, 1, 0.3, 1) backwards;
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;

  // 顶部亮线：悬停时从中间向两侧展开
  // Top light: expands from the centre outwards on hover
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 14px;
    right: 14px;
    height: 2px;
    border-radius: 2px;
    background: linear-gradient(90deg, transparent, var(--el-color-primary), transparent);
    transform: scaleX(0);
    transition: transform 0.3s ease;
    pointer-events: none;
  }

  &:hover {
    transform: translateY(-2px);
    border-color: var(--el-color-primary-light-7);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.07);

    &::before {
      transform: scaleX(1);
    }

    .card-spot {
      opacity: 1;
    }
  }

  // 指针光斑：位置由 --mx/--my 驱动，只在悬停时浮现
  // Pointer glow driven by --mx/--my, visible on hover only
  .card-spot {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    opacity: 0;
    pointer-events: none;
    background: radial-gradient(
      220px circle at var(--mx, 50%) var(--my, 0%),
      color-mix(in srgb, var(--el-color-primary) 10%, transparent),
      transparent 66%
    );
    transition: opacity 0.25s ease;
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

    .app-desc {
      flex: 1;
      margin-bottom: 0;
      min-height: 0;
      -webkit-line-clamp: 1;
    }

    // 列表模式信息密集，去掉水印与光斑，只留箭头作为可点提示
    // The list is dense: drop the watermark and glow, keep only the arrow hint
    .card-watermark,
    .card-spot {
      display: none;
    }

    .card-bottom {
      flex-shrink: 0;
      padding-top: 0;
      border-top: none;
    }
  }
}

@keyframes card-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

// 尊重"减少动态效果"偏好：入场与悬停位移全部关闭
// Honour prefers-reduced-motion: no entrance or hover motion
@media (prefers-reduced-motion: reduce) {
  .app-card {
    animation: none;
    transition: none;

    &:hover {
      transform: none;
    }

    &::before,
    .card-spot {
      display: none;
    }
  }
}

.card-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

// 标题 + 分类/模型副行，占据头部中间的弹性宽度
// Title with a category/model subline, taking the flexible middle width
.card-headline {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.app-sub {
  font-size: 12px;
  line-height: 1.4;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

// 字母头像：哈希选色的实底 + 白色首字符，作为每张卡片的视觉锚点
// Letter avatar: hash-picked solid with a white initial — the per-card anchor
.app-icon {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 11px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
  transition: transform 0.18s ease, box-shadow 0.18s ease;

  // 上传的图标：铺满圆角方块，白底兜底透明度
  // Uploaded icon fills the rounded square, on a white backing
  &.app-icon--image {
    background: var(--el-bg-color);
    overflow: hidden;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }
  }

  .app-icon-ch {
    font-size: 19px;
    font-weight: 600;
    line-height: 1;
    color: #fff;
    // 视觉重心微调：汉字/字母都略上移居中
    // nudge the glyph up slightly for optical centring
    transform: translateY(-0.5px);
  }
}

.app-card:hover .app-icon {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px color-mix(in srgb, var(--el-color-primary) 34%, transparent);
}

// 水印：超大号淡色分类图标压在右下角，给卡片一层"图案"而非空白
// Watermark: an oversized, very faint category glyph in the corner
.card-watermark {
  position: absolute;
  right: -20px;
  top: -22px;
  z-index: 0;
  font-size: 84px;
  line-height: 1;
  color: var(--el-color-primary);
  opacity: 0.04;
  transform: rotate(-10deg);
  pointer-events: none;
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.app-card:hover .card-watermark {
  opacity: 0.075;
  transform: rotate(-4deg) scale(1.04);
}

// 标题内的状态徽章：紧跟标题，作为标题的附属信息
// Status pins inline after the title
.hot-pin,
.new-pin {
  display: inline-block;
  margin-left: 6px;
  padding: 0 6px;
  font-size: 10.5px;
  font-weight: 500;
  line-height: 1.7;
  border-radius: 4px;
  vertical-align: 1px;
}

.hot-pin {
  color: #b45309;
  background: rgba(245, 158, 11, 0.12);
}

.new-pin {
  color: #047857;
  background: rgba(16, 185, 129, 0.12);
}

// 描述区：占据剩余高度，底栏自然沉到卡片底部
// Description takes the remaining height so the footer settles at the bottom
.app-desc {
  position: relative;
  z-index: 1;
  flex: 1;
  margin: 0 0 16px;
  min-height: 40px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--el-text-color-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.app-title {
  display: flex;
  align-items: center;
  margin: 0;
  font-size: 15.5px;
  font-weight: 600;
  line-height: 1.35;
  color: var(--el-text-color-primary);
  white-space: nowrap;
  overflow: hidden;

  // 名称单独截断（可收缩），徽章保持原尺寸（不收缩）
  // The name shrinks and ellipsises; pins never shrink
  .app-name {
    flex-shrink: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .hot-pin,
  .new-pin {
    flex-shrink: 0;
  }
}





.card-bottom {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  padding-right: 34px;
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

// 头部圆形箭头：默认浅主色，悬停变实底并右移一格
// Round arrow CTA: brand-tinted by default, solid and nudged right on hover
.card-arrow {
  position: absolute;
  right: 16px;
  bottom: 16px;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  font-size: 14px;
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  border-radius: 50%;
  transition: color 0.18s ease, background 0.18s ease, transform 0.18s ease,
    box-shadow 0.18s ease;
}

.app-card:hover .card-arrow {
  color: #fff;
  background: var(--el-color-primary);
  transform: translateX(2px);
  box-shadow: 0 2px 10px color-mix(in srgb, var(--el-color-primary) 32%, transparent);
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

    .app-desc {
      margin-bottom: 12px;
    }

    .card-bottom {
      padding-top: 12px;
      border-top: 1px solid var(--el-border-color-lighter);
    }
  }
}
</style>