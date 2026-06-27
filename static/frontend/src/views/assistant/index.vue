<template>
  <div class="assistant-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section">
      <div class="search-bg-overlay"></div>
      <div class="search-particles">
        <span v-for="i in 6" :key="i" class="particle" :class="`particle-${i}`"></span>
      </div>
      <div class="search-content">
        <h1 class="title">发现你的 AI 助手</h1>
        <p class="subtitle">探索数百个专业 AI 助手，让智能为你的工作加速</p>
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            placeholder="搜索助手名称或描述..."
            :prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
            :loading="loading"
            size="large"
          />
        </div>
        <div class="action-buttons">
          <el-button class="action-btn primary-btn" @click="startChat">
            <el-icon><Timer /></el-icon>
            开始聊天
          </el-button>
          <el-button v-permission:login class="action-btn ghost-btn" @click="createAssistant">
            <el-icon><Plus /></el-icon>
            创建助手
          </el-button>
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
          <div class="card-body">
            <div class="card-header">
              <el-avatar :size="48" :src="preset.avatar" class="preset-avatar">
                {{ preset.name.charAt(0) }}
              </el-avatar>
              <div class="preset-meta">
                <div class="preset-name-row">
                  <h3 class="preset-name">{{ preset.name }}</h3>
                  <span v-if="preset.type === 'official'" class="official-badge">
                    <el-icon><Star /></el-icon>
                    官方
                  </span>
                </div>
                <div class="preset-categories">
                  <span
                    v-if="preset.type === 'community'"
                    class="community-text"
                  >社区</span>
                  <span
                    v-show="preset.tags"
                    v-for="tag in preset.tags?.split(',')"
                    :key="tag"
                    class="category-tag"
                    :class="getTagType(tag)"
                  >{{ getTagName(tag) }}</span>
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
              @click.stop="useAssistant(preset)"
              v-permission:login
            >
              使用
              <el-icon class="arrow-icon"><ArrowRight /></el-icon>
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
$primary: #4f46e5;
$primary-light: #818cf8;
$primary-gradient: linear-gradient(135deg, $primary, $primary-light);
$gray-50: #f9fafb;
$gray-100: #f3f4f6;
$gray-200: #e5e7eb;
$gray-400: #9ca3af;
$gray-500: #6b7280;
$gray-600: #4b5563;
$gray-700: #374151;
$gray-900: #111827;

.assistant-container {
  min-height: 100vh;
  background: $gray-50;
}

// ========== 搜索区域 ==========
.search-section {
  position: relative;
  padding: 80px 20px 100px;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  overflow: hidden;

  .search-bg-overlay {
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse at 30% 20%, rgba($primary, 0.15) 0%, transparent 50%),
      radial-gradient(ellipse at 70% 80%, rgba($primary-light, 0.1) 0%, transparent 50%);
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
      background: rgba($primary-light, 0.06);
      animation: float 25s infinite ease-in-out;

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
    max-width: 640px;
    margin: 0 auto;

    .title {
      font-size: 2.5em;
      margin: 0 0 12px;
      font-weight: 700;
      color: #fff;
      letter-spacing: -0.02em;
      line-height: 1.2;
    }

    .subtitle {
      font-size: 16px;
      color: rgba(255, 255, 255, 0.6);
      margin: 0 0 32px;
      line-height: 1.5;
    }

    .search-box {
      max-width: 480px;
      margin: 0 auto 24px;

      :deep(.el-input__wrapper) {
        padding: 4px 8px 4px 16px;
        background: rgba(255, 255, 255, 0.1);
        backdrop-filter: blur(12px);
        border: 1px solid rgba(255, 255, 255, 0.15);
        box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
        border-radius: 12px;
        transition: all 0.2s ease;
        height: 48px;

        &:hover, &:focus-within {
          background: rgba(255, 255, 255, 0.15);
          border-color: rgba(255, 255, 255, 0.25);
        }
      }

      :deep(.el-input__inner) {
        font-size: 15px;
        color: #fff;
        &::placeholder { color: rgba(255, 255, 255, 0.5); }
      }

      :deep(.el-input__prefix) { color: rgba(255, 255, 255, 0.5); }
      :deep(.el-input__clear) { color: rgba(255, 255, 255, 0.5); }
    }

    .action-buttons {
      display: flex;
      justify-content: center;
      gap: 14px;

      .action-btn {
        padding: 12px 28px;
        font-size: 15px;
        font-weight: 500;
        border-radius: 10px;
        transition: all 0.2s ease;
        display: flex;
        align-items: center;
        gap: 8px;

        .el-icon {
          font-size: 16px;
        }

        &:hover {
          transform: translateY(-1px);
        }

        &.primary-btn {
          background: rgba(255, 255, 255, 0.15);
          border: 1px solid rgba(255, 255, 255, 0.25);
          color: #fff;
          backdrop-filter: blur(8px);

          &:hover {
            background: rgba(255, 255, 255, 0.25);
            border-color: rgba(255, 255, 255, 0.4);
          }
        }

        &.ghost-btn {
          background: transparent;
          border: 1px solid rgba(255, 255, 255, 0.2);
          color: rgba(255, 255, 255, 0.8);

          &:hover {
            background: rgba(255, 255, 255, 0.1);
            border-color: rgba(255, 255, 255, 0.3);
            color: #fff;
          }
        }
      }
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
  margin-top: -32px;
  padding: 0 24px;
}

.tag-nav {
  display: flex;
  justify-content: center;
  gap: 4px;
  padding: 8px 16px;
  max-width: 900px;
  margin: 0 auto;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);

  .tag-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.15s ease;
    white-space: nowrap;
    font-size: 14px;
    color: $gray-500;
    flex-shrink: 0;

    &:hover {
      color: $gray-700;
      background: $gray-100;
    }

    &.active {
      color: $primary;
      background: rgba($primary, 0.08);
      font-weight: 500;
    }
  }
}

// ========== 助手列表 ==========
.assistants-section {
  padding: 32px 24px 80px;
  max-width: 1200px;
  margin: 0 auto;
}

.assistants-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.preset-card {
  position: relative;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid $gray-200;

  &:hover {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
    border-color: $gray-200;
    transform: translateY(-2px);

    .use-button {
      background: $primary-gradient !important;
    }
  }

  .card-body { padding: 20px 20px 16px; }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 12px;
  }

  .preset-avatar {
    flex-shrink: 0;
    background: $primary-gradient;
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
    color: $gray-900;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .official-badge {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-size: 11px;
    font-weight: 500;
    color: #f59e0b;
    background: rgba(245, 158, 11, 0.1);
    padding: 2px 8px;
    border-radius: 6px;
    flex-shrink: 0;

    .el-icon { font-size: 11px; }
  }

  .preset-categories {
    display: flex;
    gap: 6px;
    align-items: center;
    flex-wrap: wrap;

    .community-text {
      font-size: 12px;
      color: $gray-400;
    }

    .category-tag {
      font-size: 12px;
      padding: 2px 8px;
      border-radius: 4px;
      background: $gray-100;
      color: $gray-600;

      &.warning { background: rgba(245, 158, 11, 0.1); color: #d97706; }
      &.success { background: rgba(16, 185, 129, 0.1); color: #059669; }
      &.danger { background: rgba(239, 68, 68, 0.1); color: #dc2626; }
      &.info { background: rgba(59, 130, 246, 0.1); color: #2563eb; }
      &.primary { background: rgba($primary, 0.1); color: $primary; }
    }
  }

  .preset-description {
    margin: 0;
    font-size: 13px;
    color: $gray-500;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    line-height: 1.6;
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    border-top: 1px solid $gray-100;
  }

  .footer-left {
    display: flex;
    gap: 4px;
  }

  .icon-btn {
    border: 1px solid $gray-200 !important;
    background: #fff !important;
    color: $gray-400;
    transition: all 0.15s ease;

    &:hover {
      color: $primary;
      border-color: rgba($primary, 0.3) !important;
      background: rgba($primary, 0.05) !important;
    }

    &.danger:hover {
      color: #dc2626;
      border-color: rgba(239, 68, 68, 0.3) !important;
      background: rgba(239, 68, 68, 0.05) !important;
    }
  }

  .use-button {
    background: $gray-100 !important;
    border: none !important;
    color: $gray-700 !important;
    font-weight: 500;
    font-size: 13px;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 4px;

    .arrow-icon {
      font-size: 14px;
      transition: transform 0.2s ease;
    }

    &:hover {
      color: #fff !important;

      .arrow-icon { transform: translateX(2px); }
    }
  }
}

// ========== 响应式 ==========
@media (max-width: 768px) {
  .search-section {
    padding: 60px 16px 80px;

    .search-content {
      .title { font-size: 1.8em; }
      .subtitle { font-size: 14px; margin-bottom: 24px; }

      .action-buttons .action-btn {
        padding: 8px 18px;
        font-size: 13px;
      }
    }
  }

  .tag-nav-wrapper { padding: 0 12px; }

  .tag-nav {
    justify-content: flex-start;
    padding: 6px 8px;
    .tag-item { padding: 6px 10px; font-size: 13px; }
  }

  .assistants-section { padding: 24px 12px 60px; }

  .assistants-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

.el-empty {
  grid-column: 1 / -1;
  margin: 60px 0;
}
</style>
