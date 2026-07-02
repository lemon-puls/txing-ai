<template>
  <div class="assistant-container">
    <!-- ==================== Hero 区域 ==================== -->
    <section class="hero-section">
      <div class="hero-bg">
        <div class="hero-aurora aurora-1"></div>
        <div class="hero-aurora aurora-2"></div>
        <div class="hero-aurora aurora-3"></div>
        <div class="hero-grid"></div>
      </div>

      <div class="hero-content">
        <!-- 状态徽章 -->
        <div class="hero-badge">
          <span class="badge-dot"></span>
          <span>{{ assistants.length }}+ 精选助手</span>
        </div>

        <!-- 标题 -->
        <h1 class="hero-title">
          <span class="title-line">发现你的</span>
          <span class="title-line gradient-text">专属 AI 助手</span>
        </h1>
        <p class="hero-subtitle">
          从专业编码到创意写作，从生活助手到知识导师 —— 找到最懂你的那个 AI 伙伴。
        </p>

        <!-- 搜索框 -->
        <div class="search-wrapper">
          <div class="search-box" :class="{ focused: searchFocused }">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
              v-model="searchQuery"
              class="search-input"
              placeholder="搜索助手名称或描述..."
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
          <!-- 热门标签 -->
          <div class="quick-tags">
            <span
              v-for="tag in quickTags"
              :key="tag"
              class="quick-tag"
              @click="quickSearch(tag)"
            >{{ tag }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-buttons">
          <button class="action-btn primary" @click="startChat">
            <el-icon><ChatDotRound /></el-icon>
            <span>开始聊天</span>
          </button>
          <button v-permission:login class="action-btn secondary" @click="createAssistant">
            <el-icon><Plus /></el-icon>
            <span>创建助手</span>
          </button>
        </div>
      </div>
    </section>

    <!-- ==================== 分类导航 ==================== -->
    <section class="category-section">
      <div class="category-wrapper">
        <div class="category-header">
          <h2 class="section-title">
            <span class="title-decor"></span>
            助手分类
          </h2>
          <span class="assistant-count">{{ assistants.length }} 个助手</span>
        </div>

        <div class="category-scroll">
          <div class="category-nav">
            <div
              v-for="tag in tags"
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

    <!-- ==================== 助手列表 ==================== -->
    <section class="assistants-section">
      <div class="assistants-header">
        <h2 class="section-title">
          <span class="title-decor"></span>
          {{ currentCategoryName }}
        </h2>
        <div class="header-actions">
          <div class="sort-options">
            <button
              v-for="opt in sortOptions"
              :key="opt.value"
              class="sort-btn"
              :class="{ active: currentSort === opt.value }"
              @click="currentSort = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
          <div class="view-toggle">
            <button
              class="view-btn"
              :class="{ active: viewMode === 'grid' }"
              @click="viewMode = 'grid'"
            >
              <el-icon><Grid /></el-icon>
            </button>
            <button
              class="view-btn"
              :class="{ active: viewMode === 'list' }"
              @click="viewMode = 'list'"
            >
              <el-icon><List /></el-icon>
            </button>
          </div>
        </div>
      </div>

      <!-- 加载骨架屏 -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner">
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
        </div>
        <p>正在加载助手...</p>
      </div>

      <!-- 空状态 -->
      <el-empty
        v-else-if="assistants.length === 0"
        class="empty-state"
        description="暂无匹配的助手"
      >
        <template #image>
          <div class="empty-illustration">
            <div class="empty-icon-wrapper">
              <el-icon :size="64"><UserFilled /></el-icon>
            </div>
          </div>
        </template>
        <template #default>
          <el-button type="primary" @click="createAssistant">
            <el-icon><Plus /></el-icon>
            创建第一个助手
          </el-button>
        </template>
      </el-empty>

      <!-- 助手卡片网格 -->
      <div v-else class="assistants-grid" :class="`view-${viewMode}`">
        <article
          v-for="(preset, index) in sortedAssistants"
          :key="preset.id"
          class="assistant-card"
          :class="[`cat-${getCategory(preset)}`, { 'is-mine': isMyAssistant(preset) }]"
          :style="{ animationDelay: `${index * 0.05}s` }"
          @click="useAssistant(preset)"
        >
          <!-- 卡片光效 -->
          <div class="card-glow"></div>

          <!-- 头部 -->
          <header class="card-top">
            <div class="avatar-wrapper">
              <el-avatar :size="52" :src="preset.avatar" class="assistant-avatar">
                {{ preset.name.charAt(0) }}
              </el-avatar>
            </div>
            <div class="meta-badges">
              <span v-if="preset.official" class="badge-official">
                <el-icon :size="10"><Star /></el-icon>
                官方
              </span>
              <span v-if="isMyAssistant(preset)" class="badge-mine">
                <el-icon :size="10"><User /></el-icon>
                我的
              </span>
            </div>
          </header>

          <!-- 内容 -->
          <div class="card-body">
            <h3 class="assistant-name">{{ preset.name }}</h3>
            <p class="assistant-desc">{{ preset.description || '暂无描述' }}</p>
          </div>

          <!-- 标签 -->
          <div class="assistant-tags" v-if="preset.tags">
            <span
              v-for="tag in getTags(preset.tags)"
              :key="tag"
              class="tag-item"
              :class="`tag-${tag}`"
            >
              <el-icon :size="10"><component :is="getTagIcon(tag)" /></el-icon>
              {{ getTagName(tag) }}
            </span>
          </div>

          <!-- 底部 -->
          <footer class="card-footer">
            <div class="creator-info" v-if="preset.creatorName || preset.userId === userStore.userId">
              <el-icon :size="12"><Avatar /></el-icon>
              <span class="creator-name">{{ preset.creatorName || '我' }}</span>
            </div>
            <div v-else class="creator-info placeholder"></div>

            <div class="card-actions">
              <template v-if="isMyAssistant(preset)">
                <el-tooltip content="编辑" placement="top">
                  <button class="action-icon-btn" @click.stop="editAssistant(preset)">
                    <el-icon :size="14"><Edit /></el-icon>
                  </button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <button class="action-icon-btn danger" @click.stop="deleteAssistant(preset)">
                    <el-icon :size="14"><Delete /></el-icon>
                  </button>
                </el-tooltip>
              </template>
              <button class="use-btn" @click.stop="useAssistant(preset)" v-permission:login>
                <span>使用</span>
                <el-icon><ArrowRight /></el-icon>
              </button>
            </div>
          </footer>
        </article>
      </div>
    </section>

    <!-- 创建助手弹窗 -->
    <create-assistant-dialog
      v-model:visible="createDialogVisible"
      @created="handleAssistantCreated"
    />

    <!-- 编辑助手弹窗 -->
    <create-assistant-dialog
      v-model:visible="editDialogVisible"
      :edit-data="currentEditAssistant"
      @updated="handleAssistantUpdated"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Plus, ChatDotRound, ArrowRight, Tools, Edit, Monitor,
  Reading, House, More, Star, User, Delete, UserFilled, Close,
  MagicStick, Grid, Bell, DocumentCopy, Collection, TrendCharts, Setting,
  List, Avatar, Grid as GridIcon, StarFilled, ChatLineRound
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { defaultApi } from '@/api'
import CreateAssistantDialog from '@/components/assistant/CreateAssistantDialog.vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user.js'

defineOptions({ name: 'AssistantList' })

const router = useRouter()
const themeStore = useThemeStore()
const userStore = useUserStore()

const previousThemeState = ref(null)
const previousTheme = ref("'#3B82F6'")

onMounted(() => {
  previousThemeState.value = themeStore.isDark
  previousTheme.value = themeStore.primaryColor
  themeStore.setPrimaryColor('#3B82F6')
  if (themeStore.isDark) themeStore.toggleTheme()
  loadAssistants()
})

onBeforeUnmount(() => {
  if (previousThemeState.value) themeStore.setPrimaryColor(previousTheme.value)
  if (previousThemeState.value && !themeStore.isDark) themeStore.toggleTheme()
})

const searchQuery = ref('')
const searchFocused = ref(false)
const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const currentEditAssistant = ref(null)
const assistants = ref([])
const loading = ref(false)
const currentCategory = ref('all')
const currentSort = ref('default')
const viewMode = ref('grid')

const quickTags = ['智能对话', '代码专家', '文案创作', '学习助手', '生活助手']

const sortOptions = [
  { label: '默认', value: 'default' },
  { label: '最新', value: 'newest' },
  { label: '名称', value: 'name' }
]

// 分类配置
const tags = [
  { id: 'all',      name: '全部',     icon: 'GridIcon',       desc: '所有助手' },
  { id: 'my',       name: '我的助手', icon: 'User',           desc: '我的创作' },
  { id: 'popular',  name: '热门推荐', icon: 'StarFilled',     desc: '精选热门' },
  { id: 'chat',     name: '智能对话', icon: 'ChatLineRound',  desc: '聊天陪伴' },
  { id: 'tools',    name: '实用工具', icon: 'Tools',          desc: '效率提升' },
  { id: 'writing',  name: '文案创作', icon: 'Edit',           desc: '写作灵感' },
  { id: 'coding',   name: '编码专家', icon: 'Monitor',        desc: '代码专家' },
  { id: 'learning', name: '知识学习', icon: 'Reading',        desc: '学习助手' },
  { id: 'life',     name: '生活指南', icon: 'House',          desc: '生活帮手' },
  { id: 'other',    name: '其他',     icon: 'More',           desc: '更多分类' }
]

// 图标映射
const iconMap = {
  Grid: Grid, Star: Star, StarFilled: StarFilled, Tools: Tools, Edit: Edit, Monitor: Monitor,
  Reading: Reading, House: House, More: More, User: User, Avatar: Avatar,
  ChatDotRound: ChatDotRound, ChatLineRound: ChatLineRound, TrendCharts: TrendCharts,
  Bell: Bell, DocumentCopy: DocumentCopy, Collection: Collection, Setting: Setting,
  MagicStick: MagicStick, GridIcon: GridIcon
}

// 分类图标映射
const tagIconMap = {
  popular: 'StarFilled',
  chat: 'ChatLineRound',
  tools: 'Tools',
  writing: 'Edit',
  coding: 'Monitor',
  learning: 'Reading',
  life: 'House',
  other: 'More',
  all: 'GridIcon',
  my: 'User'
}

const getTagIcon = (tagId) => {
  const iconName = tagIconMap[tagId] || 'GridIcon'
  return iconMap[iconName] || GridIcon
}

const getTagName = (tagId) => {
  const tag = tags.find(t => t.id === tagId)
  return tag ? tag.name : tagId
}

const getCategory = (preset) => {
  if (preset.tags) {
    const firstTag = preset.tags.split(',')[0].trim()
    if (['tools', 'writing', 'coding', 'learning', 'life', 'chat', 'other', 'popular'].includes(firstTag)) {
      return firstTag
    }
  }
  return 'default'
}

const getTags = (tagsStr) => {
  return tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(Boolean).slice(0, 3) : []
}

const isMyAssistant = (preset) => preset.userId === userStore.userId

const currentCategoryName = computed(() => {
  const cat = tags.find(c => c.id === currentCategory.value)
  return cat ? cat.name : '全部助手'
})

const sortedAssistants = computed(() => {
  const list = [...assistants.value]
  switch (currentSort.value) {
    case 'newest':
      return list.sort((a, b) => b.id - a.id)
    case 'name':
      return list.sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
    default:
      return list.sort((a, b) => {
        if (a.official && !b.official) return -1
        if (!a.official && b.official) return 1
        if (isMyAssistant(a) && !isMyAssistant(b)) return -1
        if (!isMyAssistant(a) && isMyAssistant(b)) return 1
        return b.id - a.id
      })
  }
})

const loadAssistants = async () => {
  try {
    loading.value = true
    const response = await defaultApi.apiPresetListGet(1, 999, {
      orderBy: 'id',
      order: 'desc',
      name: searchQuery.value || undefined,
      tags: currentCategory.value === 'all' || currentCategory.value === 'my' ? undefined : currentCategory.value,
      userId: currentCategory.value === 'my' ? userStore.userId : undefined
    })
    if (response.code === 0 && response.data) {
      assistants.value = response.data.records
    } else {
      ElMessage.error(response.msg || '获取助手列表失败')
    }
  } catch (error) {
    ElMessage.error(error.body?.msg || '获取助手列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => loadAssistants()
const quickSearch = (keyword) => {
  searchQuery.value = keyword
  handleSearch()
}
const selectCategory = (tagId) => { currentCategory.value = tagId; loadAssistants() }
const startChat = () => router.push({ path: '/chat', query: { newChat: 'true' } })
const createAssistant = () => { createDialogVisible.value = true }

const useAssistant = (preset) => {
  router.push({ path: '/chat', query: { newChat: 'true', presetId: preset.id } })
}

const editAssistant = (preset) => {
  currentEditAssistant.value = { ...preset }
  editDialogVisible.value = true
}

const handleAssistantCreated = () => loadAssistants()
const handleAssistantUpdated = () => loadAssistants()

const deleteAssistant = async (preset) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这个助手吗？删除后无法恢复。',
      '删除确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const response = await defaultApi.apiPresetIdDelete(preset.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      loadAssistants()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error.body?.msg || '删除失败')
  }
}
</script>

<style lang="scss" scoped>
// ====================
// 主题变量
// ====================
$primary: #3B82F6;

// 渐变色板
$grad-blue: linear-gradient(135deg, #2B5EFF 0%, #4facfe 100%);
$grad-cyan: linear-gradient(135deg, #06B6D4 0%, #3B82F6 100%);
$grad-orange: linear-gradient(135deg, #F59E0B 0%, #EF4444 100%);
$grad-green: linear-gradient(135deg, #10B981 0%, #34D399 100%);
$grad-pink: linear-gradient(135deg, #EC4899 0%, #F472B6 100%);
$grad-amber: linear-gradient(135deg, #F59E0B 0%, #FBBF24 100%);
$grad-teal: linear-gradient(135deg, #14B8A6 0%, #06B6D4 100%);
$grad-rose: linear-gradient(135deg, #F43F5E 0%, #FB7185 100%);

$cat-grads: (
  popular:  $grad-orange,
  chat:     $grad-blue,
  tools:    $grad-amber,
  writing:  $grad-blue,
  coding:   $grad-cyan,
  learning: $grad-green,
  life:     $grad-teal,
  other:    $grad-cyan,
  default:  $grad-blue
);

$cat-colors: (
  popular:  #F59E0B,
  chat:     #3B82F6,
  tools:    #F59E0B,
  writing:  #3B82F6,
  coding:   #06B6D4,
  learning: #10B981,
  life:     #14B8A6,
  other:    #06B6D4,
  default:  #3B82F6
);

// ====================
// 容器
// ====================
.assistant-container {
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

  .hero-aurora {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.45;
    animation: aurora-float 20s ease-in-out infinite;

    &.aurora-1 {
      width: 600px;
      height: 600px;
      background: radial-gradient(circle, rgba(99, 102, 241, 0.5), transparent 70%);
      top: -200px;
      left: -100px;
      animation-delay: 0s;
    }

    &.aurora-2 {
      width: 500px;
      height: 500px;
      background: radial-gradient(circle, rgba(59, 130, 246, 0.4), transparent 70%);
      top: -100px;
      right: -100px;
      animation-delay: -7s;
    }

    &.aurora-3 {
      width: 400px;
      height: 400px;
      background: radial-gradient(circle, rgba(59, 130, 246, 0.3), transparent 70%);
      bottom: -100px;
      left: 50%;
      animation-delay: -14s;
    }
  }

  .hero-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(var(--divider-rgb, 0, 0, 0), 0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(var(--divider-rgb, 0, 0, 0), 0.04) 1px, transparent 1px);
    background-size: 60px 60px;
    mask-image: radial-gradient(ellipse at 50% 0%, #000 30%, transparent 75%);
    -webkit-mask-image: radial-gradient(ellipse at 50% 0%, #000 30%, transparent 75%);
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
  max-width: 720px;
  margin: 0 auto;
  text-align: center;
}

// Badge
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

// Title
.hero-title {
  margin: 0 0 20px;
  font-size: 3.4em;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -1px;
  color: var(--el-text-color-primary);

  .title-line { display: block; }

  .gradient-text {
    background: linear-gradient(135deg, #2B5EFF 0%, #4facfe 50%, #06B6D4 100%);
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
  max-width: 560px;
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
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &.focused {
    border-color: $primary;
    box-shadow: 0 0 0 4px rgba($primary, 0.08), 0 8px 32px rgba(0, 0, 0, 0.08);
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

    &::placeholder { color: var(--el-text-color-placeholder); }
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
    background: $grad-blue;
    border: none;
    border-radius: 14px;
    cursor: pointer;
    transition: all 0.25s ease;
    box-shadow: 0 4px 12px rgba($primary, 0.25);

    &:hover {
      transform: scale(1.02);
      box-shadow: 0 6px 18px rgba($primary, 0.35);
    }

    &:active { transform: scale(0.98); }
  }
}

// Quick Tags
.quick-tags {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 18px;

  .quick-tag {
    padding: 4px 12px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      color: $primary;
      border-color: $primary;
      background: rgba($primary, 0.04);
      transform: translateY(-1px);
    }
  }
}

// Action Buttons
.action-buttons {
  display: flex;
  justify-content: center;
  gap: 14px;
  margin-top: 28px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 26px;
  font-size: 14px;
  font-weight: 600;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.25s ease;

  &:hover { transform: translateY(-2px); }

  &.primary {
    color: #fff;
    background: $grad-blue;
    border: none;
    box-shadow: 0 4px 16px rgba($primary, 0.35);

    &:hover { box-shadow: 0 8px 24px rgba($primary, 0.45); }
  }

  &.secondary {
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

    &:hover {
      background: var(--el-fill-color-light);
      border-color: var(--el-border-color);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
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
    background: linear-gradient(135deg, #3B82F6 0%, #4facfe 100%);
    border-radius: 2px;
  }
}

.assistant-count {
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

  &::-webkit-scrollbar { height: 4px; }
  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color);
    border-radius: 2px;
  }
  &::-webkit-scrollbar-track { background: transparent; }
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
  border-radius: 16px;
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
    flex-shrink: 0;
    transition: transform 0.3s ease;
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
    opacity: 0;
    border-radius: 16px;
    transition: opacity 0.3s ease;
    pointer-events: none;
    background: radial-gradient(circle at center, rgba($primary, 0.08), transparent 70%);
  }

  &:hover {
    transform: translateY(-3px);
    border-color: rgba($primary, 0.25);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);

    .cat-icon { transform: scale(1.08) rotate(-3deg); }
  }

  &.active {
    border-color: transparent;
    box-shadow: 0 8px 24px rgba($primary, 0.15);

    .cat-glow { opacity: 1; }
  }
}

// 分类图标颜色
@each $name, $gradient in $cat-grads {
  .category-item.cat-#{$name} .cat-icon {
    background: $gradient;
  }
}

// ====================
// 助手列表
// ====================
.assistants-section {
  max-width: 1200px;
  padding: 40px 24px 80px;
  margin: 0 auto;
}

.assistants-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sort-options {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--el-fill-color-light);
  border-radius: 10px;

  .sort-btn {
    padding: 6px 14px;
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover { color: var(--el-text-color-primary); }

    &.active {
      color: $primary;
      background: var(--el-bg-color);
      box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
    }
  }
}

.view-toggle {
  display: flex;
  gap: 2px;
  padding: 4px;
  background: var(--el-fill-color-light);
  border-radius: 10px;

  .view-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover { color: var(--el-text-color-primary); }

    &.active {
      color: $primary;
      background: var(--el-bg-color);
      box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
    }
  }
}

// 空状态
.empty-state {
  padding: 60px 0;

  .empty-illustration {
    color: var(--el-text-color-placeholder);
    opacity: 0.5;
  }

  .empty-icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 120px;
    height: 120px;
    background: var(--el-fill-color-light);
    border-radius: 24px;
  }
}

// 加载状态
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  color: var(--el-text-color-secondary);

  p {
    margin-top: 20px;
    font-size: 14px;
  }
}

.loading-spinner {
  position: relative;
  width: 60px;
  height: 60px;

  .spinner-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 3px solid transparent;
    border-top-color: $primary;
    animation: spin 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;

    &:nth-child(1) { animation-delay: -0.45s; }
    &:nth-child(2) { animation-delay: -0.3s; }
    &:nth-child(3) { animation-delay: -0.15s; }
  }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

// 助手网格
.assistants-grid {
  display: grid;
  gap: 20px;

  &.view-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  }

  &.view-list {
    grid-template-columns: 1fr;

    .assistant-card {
      flex-direction: row;
      align-items: center;
      padding: 20px 24px;

      .card-top {
        padding: 0;
        flex-direction: column;
        align-items: center;
      }

      .card-body {
        flex: 1;
        padding: 0 20px;
      }

      .assistant-tags {
        display: none;
      }

      .card-footer {
        padding: 0;
        border: none;
      }

      &:hover {
        transform: translateX(4px);
      }
    }
  }
}

// ====================
// 助手卡片
// ====================
.assistant-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 20px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  animation: card-rise 0.5s cubic-bezier(0.16, 1, 0.3, 1) backwards;

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
      rgba($primary, 0.06),
      transparent 40%
    );
  }

  &:hover {
    transform: translateY(-6px);
    border-color: rgba($primary, 0.2);
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.08), 0 4px 12px rgba(0, 0, 0, 0.04);

    .card-glow { opacity: 1; }
    .use-btn {
      background: $grad-blue;
      color: #fff;
      box-shadow: 0 6px 16px rgba($primary, 0.3);

      .el-icon { transform: translateX(4px); }
    }
  }

  &.is-mine {
    border-color: rgba($primary, 0.2);
  }
}

@keyframes card-rise {
  from { opacity: 0; transform: translateY(20px); }
  to   { opacity: 1; transform: translateY(0); }
}

// 头部
.card-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 24px 24px 0;
}

.avatar-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.assistant-avatar {
  background: $grad-blue;
  color: #fff;
  font-weight: 700;
  font-size: 20px;
  box-shadow: 0 4px 12px rgba($primary, 0.25);
}

// 元数据徽章
.meta-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-end;
}

.badge-official,
.badge-mine {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
  letter-spacing: 0.3px;
}

.badge-official {
  color: #fff;
  background: $grad-orange;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.badge-mine {
  color: #fff;
  background: $grad-blue;
  box-shadow: 0 2px 8px rgba($primary, 0.25);
}

// 卡片内容
.card-body {
  position: relative;
  z-index: 1;
  flex: 1;
  padding: 16px 24px;
}

.assistant-name {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.assistant-desc {
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

// 标签
.assistant-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0 24px 16px;
}

.tag-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 999px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  transition: all 0.2s ease;
}

@each $name, $color in $cat-colors {
  .tag-#{$name} {
    background: rgba($color, 0.1);
    color: $color;
  }
}

// 卡片底部
.card-footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 24px 20px;
  border-top: 1px dashed var(--el-border-color-lighter);

  .placeholder {
    visibility: hidden;
  }
}

.creator-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-text-color-placeholder);

  .creator-name {
    font-size: 12px;
  }
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: $primary;
    background: rgba($primary, 0.08);
  }

  &.danger:hover {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.08);
  }
}

.use-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 7px 14px;
  font-size: 13px;
  font-weight: 600;
  color: $primary;
  background: rgba($primary, 0.08);
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;

  .el-icon {
    font-size: 14px;
    transition: transform 0.3s ease;
  }

  &:hover {
    background: rgba($primary, 0.12);
  }
}

// ====================
// 暗色模式
// ====================
:global(.dark) {
  .hero-aurora {
    opacity: 0.3;
  }

  .assistant-card,
  .category-item,
  .search-box {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(12px);

    &:hover {
      background: rgba(255, 255, 255, 0.05);
      border-color: rgba(255, 255, 255, 0.12);
    }
  }
}

// ====================
// 响应式
// ====================
@media (max-width: 768px) {
  .hero-section {
    padding: 60px 16px 100px;
  }

  .hero-title {
    font-size: 2.4em;
  }

  .hero-subtitle {
    font-size: 0.95em;
    margin-bottom: 32px;
  }

  .search-box {
    height: 50px;

    .search-btn span { display: none; }
    .search-btn { padding: 0 14px; }
  }

  .action-buttons {
    gap: 12px;

    .action-btn {
      padding: 10px 20px;
      font-size: 13px;
    }
  }

  .category-section {
    margin-top: -50px;
    padding: 0 12px;
  }

  .category-item {
    padding: 12px 14px;
    min-width: 120px;

    .cat-icon {
      width: 36px;
      height: 36px;
    }

    .cat-name { font-size: 13px; }
    .cat-desc { display: none; }
  }

  .assistants-section {
    padding: 32px 12px 60px;
  }

  .assistants-grid {
    &.view-grid {
      grid-template-columns: 1fr;
    }
  }

  .assistants-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
