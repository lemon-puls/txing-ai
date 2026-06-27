<template>
  <div class="assistant-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section">
      <div class="search-bg-overlay"></div>
      <div class="search-particles">
        <span v-for="i in 6" :key="i" class="particle" :class="`particle-${i}`"></span>
      </div>
      <div class="search-content">
        <h1 class="title">做您强大的 AI 助手</h1>
        <div class="action-buttons">
          <el-button type="primary" class="action-btn" @click="startChat">
            <el-icon><Timer /></el-icon>
            开始聊天
          </el-button>
          <el-button v-permission:login type="primary" class="action-btn outline" @click="createAssistant">
            <el-icon><Plus /></el-icon>
            创建助手
          </el-button>
        </div>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索您的 AI 助手..."
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
          v-for="tag in tags"
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

    <!-- AI助手列表 -->
    <div class="assistants-section">
      <el-empty
        v-if="filteredAssistants.length === 0"
        description="暂无数据"
      />
      <div v-else class="assistants-grid">
        <div
          v-for="preset in filteredAssistants"
          :key="preset.id"
          class="preset-card"
          @click="useAssistant(preset)"
        >
          <div class="card-accent"></div>
          <div class="card-body">
            <div class="card-header">
              <el-avatar :size="44" :src="preset.avatar" class="preset-avatar">
                {{ preset.name.charAt(0) }}
              </el-avatar>
              <div class="preset-meta">
                <div class="preset-name-row">
                  <h3 class="preset-name">{{ preset.name }}</h3>
                  <el-tag
                    v-if="preset.type === 'official'"
                    class="preset-type official-tag"
                    effect="dark"
                    size="small"
                    round
                  >
                    <el-icon class="tag-icon"><Star /></el-icon>
                    官方
                  </el-tag>
                  <el-tag
                    v-else
                    class="preset-type community-tag"
                    type="primary"
                    effect="light"
                    size="small"
                    round
                  >
                    <el-icon class="tag-icon"><User /></el-icon>
                    社区
                  </el-tag>
                </div>
                <div class="preset-categories">
                  <div
                    v-show="preset.tags"
                    v-for="tag in preset.tags?.split(',')"
                    :key="tag"
                    class="category-dot"
                    :class="getTagType(tag)"
                    :title="getTagName(tag)"
                  >
                    <el-icon class="category-icon"><component :is="getTagIcon(tag)" /></el-icon>
                  </div>
                </div>
              </div>
            </div>
            <p class="preset-description">{{ preset.description }}</p>
          </div>
          <div class="card-footer">
            <div class="footer-left">
              <template v-if="preset.userId === userStore.userId">
                <el-tooltip content="编辑" placement="top">
                  <el-button class="icon-btn" circle size="small" @click.stop="editAssistant(preset)">
                    <el-icon :size="14"><Edit /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button class="icon-btn danger" circle size="small" @click.stop="deleteAssistant(preset)">
                    <el-icon :size="14"><Delete /></el-icon>
                  </el-button>
                </el-tooltip>
              </template>
            </div>
            <el-button
              type="primary"
              class="use-button"
              round
              @click.stop="useAssistant(preset)"
              v-permission:login
            >
              <el-icon><ArrowRight /></el-icon>
              使用
            </el-button>
          </div>
        </div>
      </div>
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
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import {
  Search, Plus, Timer, ArrowRight, Tools, Edit, Monitor,
  Reading, House, More, Star, User, Delete
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { defaultApi } from '@/api'
import { useRouter } from 'vue-router'
import CreateAssistantDialog from '@/components/assistant/CreateAssistantDialog.vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user.js'

defineOptions({ name: 'AssistantList' })

const router = useRouter()
const themeStore = useThemeStore()
const userStore = useUserStore()

const previousThemeState = ref(null)
const previousTheme = ref("'#409EFF'")

onMounted(() => {
  previousThemeState.value = themeStore.isDark
  previousTheme.value = themeStore.primaryColor
  themeStore.setPrimaryColor('#409EFF')
  if (themeStore.isDark) themeStore.toggleTheme()
  loadAssistants()
})

onBeforeUnmount(() => {
  if (previousThemeState.value) themeStore.setPrimaryColor(previousTheme.value)
  if (previousThemeState.value && !themeStore.isDark) themeStore.toggleTheme()
})

const createDialogVisible = ref(false)
const searchQuery = ref('')

const tags = [
  { id: 'all', name: '全部', icon: 'Grid' },
  { id: 'my', name: '我的助手', icon: 'User' },
  { id: 'popular', name: '热门推荐', icon: 'Star' },
  { id: 'tools', name: '实用工具', icon: 'Tools' },
  { id: 'writing', name: '文案创作', icon: 'Edit' },
  { id: 'coding', name: '编码专家', icon: 'Monitor' },
  { id: 'learning', name: '知识学习', icon: 'Reading' },
  { id: 'life', name: '生活指南', icon: 'House' },
  { id: 'other', name: '其他', icon: 'More' }
]

const assistants = ref([])
const loading = ref(false)
const currentCategory = ref('all')

const loadAssistants = async () => {
  try {
    loading.value = true
    const response = await defaultApi.apiPresetListGet(1, 999, {
      orderBy: 'id', order: 'desc',
      name: searchQuery.value || undefined,
      tags: currentCategory.value === 'all' || currentCategory.value === 'my' ? undefined : currentCategory.value,
      userId: currentCategory.value === 'my' ? userStore.userId : undefined
    })
    if (response.code === 0 && response.data) {
      assistants.value = response.data.records.map(preset => ({ ...preset, type: preset.official ? 'official' : 'community' }))
    } else {
      ElMessage.error(response.msg || '获取助手列表失败')
    }
  } catch (error) {
    ElMessage.error(error.body?.msg || '获取助手列表失败')
  } finally {
    loading.value = false
  }
}

const filteredAssistants = computed(() => assistants.value)
const handleSearch = () => loadAssistants()
const selectCategory = (tagId) => { currentCategory.value = tagId; loadAssistants() }
const startChat = () => router.push({ path: '/chat', query: { newChat: 'true' } })
const createAssistant = () => { createDialogVisible.value = true }
const handleAssistantCreated = () => loadAssistants()
const useAssistant = (preset) => router.push({ path: '/chat', query: { newChat: 'true', presetId: preset.id } })

const getTagIcon = (tagId) => {
  const map = { tools: Tools, writing: Edit, coding: Monitor, learning: Reading, life: House, other: More }
  return map[tagId] || More
}

const getTagName = (tagId) => {
  const tag = tags.find(t => t.id === tagId)
  return tag ? tag.name : tagId
}

const getTagType = (tagId) => {
  const map = { tools: 'warning', writing: 'success', coding: 'primary', learning: 'info', life: 'danger', other: '' }
  return map[tagId] || ''
}

const editDialogVisible = ref(false)
const currentEditAssistant = ref(null)
const editAssistant = (preset) => { currentEditAssistant.value = { ...preset }; editDialogVisible.value = true }
const handleAssistantUpdated = () => loadAssistants()

const deleteAssistant = async (preset) => {
  try {
    await ElMessageBox.confirm('确定要删除这个助手吗？删除后无法恢复。', '删除确认', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    const response = await defaultApi.apiPresetIdDelete(preset.id)
    if (response.code === 0) { ElMessage.success('删除成功'); loadAssistants() }
    else ElMessage.error(response.msg || '删除失败')
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error.body?.msg || '删除失败')
  }
}
</script>

<style lang="scss" scoped>
$blue-500: #2B5EFF;
$blue-400: #4facfe;
$blue-gradient: linear-gradient(135deg, $blue-500, $blue-400);

.assistant-container {
  min-height: 100vh;
  background: var(--el-bg-color-page, #f5f7fa);
}

// ========== 搜索区域 ==========
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
      margin: 0 0 28px;
      font-weight: 800;
      color: #fff;
      letter-spacing: -0.5px;
      text-shadow: 0 2px 10px rgba(0, 0, 0, 0.15);
    }

    .action-buttons {
      margin-bottom: 28px;
      display: flex;
      justify-content: center;
      gap: 14px;

      .action-btn {
        padding: 10px 28px;
        font-size: 15px;
        font-weight: 500;
        border: 1px solid rgba(255, 255, 255, 0.3);
        background: rgba(255, 255, 255, 0.15);
        backdrop-filter: blur(10px);
        border-radius: 12px;
        color: #fff;
        transition: all 0.3s ease;

        .el-icon { margin-right: 6px; }

        &:hover {
          background: rgba(255, 255, 255, 0.25);
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
        }

        &.outline {
          background: transparent;
          border-color: rgba(255, 255, 255, 0.5);

          &:hover { background: rgba(255, 255, 255, 0.12); }
        }
      }
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

// ========== 分类导航 ==========
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

// ========== 助手列表 ==========
.assistants-section {
  padding: 32px 24px 60px;
  max-width: 1200px;
  margin: 0 auto;
}

.assistants-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.preset-card {
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
    .use-button { box-shadow: 0 4px 16px rgba($blue-500, 0.3); }
  }

  .card-accent {
    height: 3px;
    background: $blue-gradient;
    transition: height 0.3s ease;
  }

  .card-body { padding: 18px 18px 12px; }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    margin-bottom: 10px;
  }

  .preset-avatar {
    flex-shrink: 0;
    background: $blue-gradient;
    color: #fff;
    font-weight: 600;
  }

  .preset-meta {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .preset-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }

  .preset-name {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .preset-type {
    flex-shrink: 0;
    font-size: 11px;

    .tag-icon { font-size: 11px; margin-right: 2px; }

    &.official-tag {
      background: linear-gradient(135deg, #ff9500, #ff3852) !important;
      border: none !important;
      color: #fff !important;
    }

    &.community-tag {
      border-color: $blue-500 !important;
      color: $blue-500 !important;
    }
  }

  .preset-categories {
    display: flex;
    gap: 4px;
    align-items: center;

    .category-dot {
      width: 22px;
      height: 22px;
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--el-color-primary);
      color: #fff;

      .category-icon { font-size: 12px; }

      &.warning { background: var(--el-color-warning); }
      &.success { background: var(--el-color-success); }
      &.danger { background: var(--el-color-danger); }
      &.info { background: var(--el-color-info); }
      &.primary { background: $blue-500; }
    }
  }

  .preset-description {
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

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 18px;
    border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  }

  .footer-left {
    display: flex;
    gap: 6px;
  }

  .icon-btn {
    border: 1px solid var(--el-border-color-lighter, #e4e7ed) !important;
    background: var(--el-bg-color, #fff) !important;
    color: var(--el-text-color-secondary);
    transition: all 0.2s ease;

    &:hover {
      color: $blue-500;
      border-color: rgba($blue-500, 0.3) !important;
      background: rgba($blue-500, 0.06) !important;
    }

    &.danger:hover {
      color: var(--el-color-danger);
      border-color: rgba(var(--el-color-danger-rgb), 0.3) !important;
      background: rgba(var(--el-color-danger-rgb), 0.06) !important;
    }
  }

  .use-button {
    background: $blue-gradient !important;
    border: none !important;
    font-weight: 500;
    transition: all 0.3s ease;

    .el-icon { margin-right: 4px; }

    &:hover {
      box-shadow: 0 4px 16px rgba($blue-500, 0.4);
      transform: translateY(-1px);
    }
  }
}

// ========== 响应式 ==========
@media (max-width: 768px) {
  .search-section {
    padding: 40px 16px 50px;
    .search-content .title { font-size: 2em; }
    .action-buttons .action-btn { padding: 8px 20px; font-size: 14px; }
  }

  .tag-nav-wrapper { padding: 0 12px; }

  .tag-nav {
    justify-content: flex-start;
    padding: 8px 12px;
    .tag-item { padding: 6px 12px; font-size: 13px; }
  }

  .assistants-section { padding: 20px 12px 40px; }

  .assistants-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}

.el-empty {
  grid-column: 1 / -1;
  margin: 40px 0;
}
</style>
