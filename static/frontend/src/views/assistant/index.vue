<template>
  <div class="assistant-page">
    <div class="page-body">
      <!-- ==================== 左侧分类导航 ==================== -->
      <!-- Left category sidebar (GPT-Store style) -->
      <aside class="side-nav">
        <div class="nav-group">
          <div class="nav-label">探索</div>
          <button
            v-for="tag in exploreTags"
            :key="tag.id"
            class="nav-item"
            :class="{ active: currentCategory === tag.id }"
            @click="selectCategory(tag.id)"
          >
            <el-icon :size="16"><component :is="iconMap[tag.icon]" /></el-icon>
            <span>{{ tag.name }}</span>
          </button>
        </div>

        <div class="nav-group">
          <div class="nav-label">分类</div>
          <button
            v-for="tag in categoryTags"
            :key="tag.id"
            class="nav-item"
            :class="{ active: currentCategory === tag.id }"
            @click="selectCategory(tag.id)"
          >
            <el-icon :size="16"><component :is="iconMap[tag.icon]" /></el-icon>
            <span>{{ tag.name }}</span>
          </button>
        </div>
      </aside>

      <!-- ==================== 右侧主内容 ==================== -->
      <!-- Main content column -->
      <main class="main-content">
        <header class="page-head">
          <h1 class="page-title">AI 助手</h1>
          <p class="page-subtitle">挑选一个助手，直接开始对话</p>
        </header>

        <!-- 工具栏：搜索 + 排序 + 操作 -->
        <div class="toolbar">
          <el-input
            v-model="searchQuery"
            class="search-input"
            placeholder="搜索助手名称或描述"
            :prefix-icon="Search"
            clearable
          />
          <div class="toolbar-right">
            <div class="sort-seg">
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
            <el-button class="toolbar-btn chat-btn" :icon="ChatDotRound" @click="startChat">
              开始聊天
            </el-button>
            <el-button
              v-permission:login
              class="toolbar-btn create-btn"
              type="primary"
              :icon="Plus"
              @click="createAssistant"
            >
              创建助手
            </el-button>
          </div>
        </div>

        <!-- 列表标题 -->
        <div class="section-head">
          <span class="section-title">{{ currentCategoryName }}</span>
          <span class="count-badge">共 {{ assistants.length }} 个</span>
        </div>

        <!-- 助手卡片网格 -->
        <div v-loading="loading" class="assistants-grid">
          <article
            v-for="preset in sortedAssistants"
            :key="preset.id"
            class="assistant-card"
            @click="useAssistant(preset)"
          >
            <div class="card-top">
              <el-avatar :size="44" :src="preset.avatar || undefined" class="card-avatar">
                {{ preset.name.charAt(0) }}
              </el-avatar>
              <div class="card-name-row">
                <span class="card-name" :title="preset.name">{{ preset.name }}</span>
                <el-tag v-if="preset.official" class="official-tag" size="small" type="warning" effect="plain" round>
                  官方
                </el-tag>
              </div>
            </div>

            <p class="card-desc">{{ preset.description || '暂无描述' }}</p>

            <div v-if="getTags(preset.tags).length" class="card-tags">
              <el-tag
                v-for="tag in getTags(preset.tags)"
                :key="tag"
                size="small"
                type="info"
                effect="plain"
                round
              >
                {{ getTagName(tag) }}
              </el-tag>
            </div>

            <footer class="card-footer">
              <span v-if="preset.creatorName || isMyAssistant(preset)" class="card-owner">
                {{ preset.creatorName || '我' }}
              </span>
              <div class="card-actions">
                <template v-if="isMyAssistant(preset)">
                  <el-tooltip content="编辑" placement="top">
                    <button class="icon-btn" @click.stop="editAssistant(preset)">
                      <el-icon :size="14"><Edit /></el-icon>
                    </button>
                  </el-tooltip>
                  <el-tooltip content="删除" placement="top">
                    <button class="icon-btn danger" @click.stop="deleteAssistant(preset)">
                      <el-icon :size="14"><Delete /></el-icon>
                    </button>
                  </el-tooltip>
                </template>
                <button v-permission:login class="use-btn" @click.stop="useAssistant(preset)">
                  使用
                </button>
              </div>
            </footer>
          </article>
        </div>

        <!-- 空状态 -->
        <div v-if="!loading && assistants.length === 0" class="empty-state">
          <el-empty :description="filterActive ? '没有找到符合条件的助手' : '还没有助手'">
            <el-button v-if="filterActive" round @click="resetFilter">清空筛选</el-button>
            <el-button v-else v-permission:login type="primary" round @click="createAssistant">
              创建助手
            </el-button>
          </el-empty>
        </div>
      </main>
    </div>

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
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Plus, ChatDotRound, Edit, Delete, Grid, User, StarFilled,
  ChatLineRound, Tools, Monitor, Reading, House, More
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { defaultApi } from '@/api'
import CreateAssistantDialog from '@/components/assistant/CreateAssistantDialog.vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user.js'

defineOptions({ name: 'AssistantList' })

const router = useRouter()
// 与 chat/websites 页一致：进入本页应用用户主题（挂载 dark class 与主题色变量）
// Same as chat/websites pages: apply the user's theme (dark class + primary color vars)
const themeStore = useThemeStore()
themeStore.initTheme()
const userStore = useUserStore()

const searchQuery = ref('')
// 搜索防抖：输入停顿 300ms 后再请求
// Debounced search: fire the request 300ms after typing stops
let searchTimer = null
watch(searchQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(loadAssistants, 300)
})
onBeforeUnmount(() => clearTimeout(searchTimer))

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const currentEditAssistant = ref(null)
const assistants = ref([])
const loading = ref(false)
const currentCategory = ref('all')
const currentSort = ref('default')

const sortOptions = [
  { label: '默认', value: 'default' },
  { label: '最新', value: 'newest' },
  { label: '名称', value: 'name' }
]

// 分类配置（all/my/popular 为探索项，其余为业务分类）
// Category config (all/my/popular are explore entries, the rest are business categories)
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

// 图标映射（侧栏配置中的 icon 字符串 → 组件）
// Icon map: sidebar config icon strings → components
const iconMap = {
  GridIcon: Grid,
  User,
  StarFilled,
  ChatLineRound,
  Tools,
  Edit,
  Monitor,
  Reading,
  House,
  More
}

// 侧栏分组：探索项 + 业务分类
// Sidebar groups: explore entries + business categories
const exploreTags = tags.filter(t => ['all', 'my', 'popular'].includes(t.id))
const categoryTags = tags.filter(t => !['all', 'my', 'popular'].includes(t.id))

const getTagName = (tagId) => {
  const tag = tags.find(t => t.id === tagId)
  return tag ? tag.name : tagId
}

const getTags = (tagsStr) => {
  return tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(Boolean).slice(0, 3) : []
}

const isMyAssistant = (preset) => preset.userId === userStore.userId

const currentCategoryName = computed(() => {
  const cat = tags.find(c => c.id === currentCategory.value)
  return cat ? cat.name : '全部助手'
})

// 是否处于筛选状态（空态文案与按钮用）
// Whether filters are active (drives empty-state copy and button)
const filterActive = computed(() => !!(searchQuery.value || currentCategory.value !== 'all'))

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

// 清空筛选（先取消挂起的搜索请求，再立即重载）
// Reset filters (cancel pending search first, then reload immediately)
const resetFilter = () => {
  clearTimeout(searchTimer)
  searchQuery.value = ''
  currentCategory.value = 'all'
  loadAssistants()
}

const selectCategory = (tagId) => { currentCategory.value = tagId; loadAssistants() }

// 初始加载
// Initial load
loadAssistants()
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
.assistant-page {
  min-height: 100vh;
  background: #f5f7fa;
}

.page-body {
  display: flex;
  gap: 32px;
  max-width: 1200px;
  margin: 0 auto;
  padding: 36px 24px 64px;
}

// ===== 左侧分类导航 =====
// Left category sidebar
.side-nav {
  position: sticky;
  top: 24px;
  align-self: flex-start;
  flex-shrink: 0;
  width: 200px;
  max-height: calc(100vh - 48px);
  overflow-y: auto;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-thumb {
    background: var(--el-border-color);
    border-radius: 2px;
  }
  &::-webkit-scrollbar-track { background: transparent; }
}

.nav-label {
  margin: 16px 0 6px;
  padding-left: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);

  .nav-group:first-child & { margin-top: 0; }
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 10px;
  margin-bottom: 2px;
  font-size: 14px;
  color: var(--el-text-color-regular);
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  .el-icon { flex-shrink: 0; }

  &:hover {
    color: var(--el-text-color-primary);
    background: var(--el-fill-color-light);
  }

  &.active {
    color: var(--el-color-primary);
    font-weight: 500;
    background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  }
}

// ===== 主内容列 =====
// Main content column
.main-content {
  flex: 1;
  min-width: 0;
}

.page-head {
  .page-title {
    margin: 0;
    font-size: 24px;
    font-weight: 700;
    color: var(--el-text-color-primary);
  }

  .page-subtitle {
    margin: 8px 0 0;
    font-size: 14px;
    color: var(--el-text-color-secondary);
  }
}

// ===== 工具栏 =====
// Toolbar: search + sort + actions
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 24px;

  .search-input {
    flex: 1;
    max-width: 360px;

    :deep(.el-input__wrapper) {
      border-radius: 999px;
    }
  }
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;

  // 工具栏按钮：胶囊形，与搜索框同一套语言
  // Toolbar buttons: pill-shaped, matching the search input language
  .toolbar-btn.el-button {
    border-radius: 999px;
    font-weight: 500;
  }

  // 开始聊天：安静的次要胶囊（白底描边与搜索框同底；fill 色和页面灰底同值会隐形）
  // "Start chat": quiet secondary pill (white + border to read on the gray page bg)
  .chat-btn.el-button {
    color: var(--el-text-color-regular);
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color-lighter);

    &:hover,
    &:focus-visible {
      color: var(--el-color-primary);
      background-color: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
      border-color: var(--el-color-primary-light-5);
    }

    html.dark & {
      // 暗色下 --el-bg-color 即页面底色、--el-border-color-lighter 未映射（近白），需显式给值
      background-color: #1c1c1c;
      border-color: #363637;

      &:hover,
      &:focus-visible {
        background-color: color-mix(in srgb, var(--el-color-primary) 18%, transparent);
        border-color: color-mix(in srgb, var(--el-color-primary) 40%, transparent);
      }
    }
  }

  // 创建助手：主 CTA 胶囊（克制的主题色投影，同 websites 选中 pill 语言）
  // "Create assistant": primary CTA pill with a restrained primary shadow
  .create-btn.el-button {
    box-shadow: 0 2px 8px color-mix(in srgb, var(--el-color-primary) 25%, transparent);

    &:hover,
    &:focus-visible {
      color: #fff;
      background-color: var(--el-color-primary-light-3);
      border-color: var(--el-color-primary-light-3);
      box-shadow: 0 4px 12px color-mix(in srgb, var(--el-color-primary) 35%, transparent);
    }

    &:active {
      color: #fff;
      background-color: color-mix(in srgb, var(--el-color-primary) 88%, black);
      border-color: transparent;
      box-shadow: 0 2px 6px color-mix(in srgb, var(--el-color-primary) 25%, transparent);
    }
  }
}

// 排序分段：白底描边胶囊（fill 色与页面灰底同值会隐形），active 用主题色 tint
// Sort segment: white + border pill (fill color equals page bg and would vanish), primary-tint active
.sort-seg {
  display: flex;
  gap: 2px;
  padding: 3px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 999px;

  .sort-btn {
    padding: 4px 12px;
    font-size: 13px;
    white-space: nowrap;
    color: var(--el-text-color-secondary);
    background: transparent;
    border: none;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover { color: var(--el-text-color-primary); }

    &.active {
      color: var(--el-color-primary);
      background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
    }
  }
}

// ===== 列表标题 =====
// List section head
.section-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 24px 2px 16px;

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .count-badge {
    padding: 2px 10px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border-radius: 999px;
  }
}

// ===== 助手网格 =====
// Assistants grid
.assistants-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 18px;
  min-height: 240px;
}

// ===== 助手卡片 =====
// Assistant card
.assistant-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px 20px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    transform: translateY(-4px);
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 12px 28px color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  }
}

.card-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-avatar {
  flex-shrink: 0;
  border-radius: 12px;
  font-size: 17px;
  font-weight: 600;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-8);
}

.card-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.official-tag {
  flex-shrink: 0;
}

.card-desc {
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
}

.card-footer {
  display: flex;
  align-items: center;
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.card-owner {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: var(--el-color-primary);
    background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  }

  &.danger:hover {
    color: var(--el-color-danger);
    background: color-mix(in srgb, var(--el-color-danger) 12%, transparent);
  }
}

.use-btn {
  padding: 5px 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: color-mix(in srgb, var(--el-color-primary) 14%, transparent);
  }
}

// ===== 空状态 =====
// Empty state
.empty-state {
  padding: 40px 0;
}

// ===== 响应式 =====
// Responsive: sidebar becomes a horizontal pill row
@media screen and (max-width: 900px) {
  .page-body {
    flex-direction: column;
    gap: 20px;
    padding: 28px 16px 48px;
  }

  .side-nav {
    position: static;
    display: flex;
    gap: 16px;
    width: 100%;
    max-height: none;
    overflow-x: auto;
    overflow-y: hidden;
    padding-bottom: 4px;

    &::-webkit-scrollbar { height: 4px; }
    &::-webkit-scrollbar-thumb {
      background: var(--el-border-color);
      border-radius: 2px;
    }
  }

  .nav-label { display: none; }

  .nav-group {
    display: flex;
    gap: 8px;
  }

  .nav-item {
    width: auto;
    margin-bottom: 0;
    padding: 6px 14px;
    white-space: nowrap;
    background: var(--el-fill-color-light);
    border-radius: 999px;
  }
}

@media screen and (max-width: 768px) {
  .toolbar {
    .search-input {
      flex-basis: 100%;
      max-width: none;
    }
  }

  .toolbar-right {
    width: 100%;
    margin-left: 0;
    justify-content: space-between;
  }

  .assistants-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}

// ===== 暗色模式：项目暗色变量未覆盖部分，这里手动补齐 =====
// 注意：scoped 下须用 `html.dark &` 写法（与 websites/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.assistant-page {
  html.dark & {
    background: #141414;
  }
}

.side-nav {
  html.dark & {
    .nav-item:hover {
      background: #262627;
    }

    .nav-item.active {
      background: color-mix(in srgb, var(--el-color-primary) 18%, transparent);
    }
  }
}

.sort-seg {
  html.dark & {
    background: #262627;
    border-color: #363637;

    .sort-btn.active {
      background: color-mix(in srgb, var(--el-color-primary) 18%, transparent);
    }
  }
}

.section-head .count-badge {
  html.dark & {
    background: #262627;
  }
}

.assistant-card {
  html.dark & {
    background: #1c1c1c;
    border-color: #363637;
  }
}

.card-footer {
  html.dark & {
    border-top-color: #363637;
  }
}

.card-avatar {
  html.dark & {
    background: color-mix(in srgb, var(--el-color-primary) 15%, transparent);
  }
}
</style>
