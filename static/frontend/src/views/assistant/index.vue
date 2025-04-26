<template>
  <div class="assistant-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section" :style="{ backgroundImage: `url(${bgImage})` }">
      <div class="nav-header">
        <div class="nav-content">
          <div class="nav-left">
            <div class="logo">
              <span class="logo-text">Txing AI</span>
            </div>
          </div>
          <div class="nav-right">
            <a href="https://github.com/yourusername/txing-ai" target="_blank" class="github-link">
              <el-tooltip content="在 GitHub 上查看" placement="bottom">
                <el-icon class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg></el-icon>
              </el-tooltip>
            </a>
            <UserAvatar />
          </div>
        </div>
      </div>

      <div class="search-content">
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
            :loading="loading"
          />
        </div>
      </div>
    </div>

    <!-- 分类导航 -->
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

    <!-- AI助手列表 -->
    <div class="assistants-grid">
      <el-empty
        v-if="filteredAssistants.length === 0"
        description="暂无数据"
      />
      <div
        v-else
        v-for="preset in filteredAssistants"
        :key="preset.id"
        class="preset-card"
      >
        <div class="preset-content">
          <div class="preset-avatar">
            <el-avatar :size="36" :src="preset.avatar">
              {{ preset.name.charAt(0) }}
            </el-avatar>
          </div>
          <div class="preset-info">
            <div class="preset-header">
              <div class="preset-name-row">
                <h3 class="preset-name">{{ preset.name }}</h3>
                <el-tag
                  v-if="preset.type === 'official'"
                  class="preset-type official-tag"
                  effect="dark"
                  size="small"
                >
                  <el-icon class="tag-icon"><Star /></el-icon>
                  官方
                </el-tag>
                <el-tag
                  v-else
                  class="preset-type community-tag"
                  type="primary"
                  effect="plain"
                  size="small"
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
            <p class="preset-description">{{ preset.description }}</p>
          </div>
        </div>
        <div class="preset-actions">
          <div class="action-wrapper">
            <template v-if="preset.userId === userStore.userId">
              <div class="action-icons">
                <el-tooltip content="编辑" placement="top">
                  <el-button
                    class="icon-button edit-button"
                    circle
                    @click.stop="editAssistant(preset)"
                  >
                    <el-icon><Edit /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button
                    class="icon-button delete-button"
                    circle
                    @click.stop="deleteAssistant(preset)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </template>
            <el-button
              type="primary"
              class="use-button"
              @click.stop="useAssistant(preset)"
            >
              <span class="button-content">
                <el-icon><ArrowRight /></el-icon>
                使用
              </span>
              <span class="button-background"></span>
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
  Search,
  Plus,
  Timer,
  ArrowRight,
  Tools,
  Edit,
  Monitor,
  Reading,
  House,
  More,
  Star,
  User,
  Delete
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { defaultApi } from '@/api'
import bgImage from '@/assets/images/header-bg.jpg'
import { useRouter } from 'vue-router'
import CreateAssistantDialog from '@/components/assistant/CreateAssistantDialog.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useThemeStore } from '@/stores/theme'
import {useUserStore} from "@/stores/user.js";

defineOptions({
  name: 'AssistantList'
})

const router = useRouter()
const themeStore = useThemeStore()
const userStore = useUserStore();

// 存储进入页面时的主题状态
const previousThemeState = ref(null)
const previousTheme = ref("'#409EFF'")

// 在组件挂载时切换为明亮主题
onMounted(() => {
  // 保存当前主题状态
  previousThemeState.value = themeStore.isDark
  previousTheme.value = themeStore.primaryColor
  themeStore.setPrimaryColor('#409EFF')
  // 如果当前是暗色主题，切换为明亮主题
  if (themeStore.isDark) {
    themeStore.toggleTheme()
  }
  loadAssistants()
})

// 在组件卸载前恢复原来的主题
onBeforeUnmount(() => {
  // 恢复原来的主题状态
  if (previousThemeState.value) {
    themeStore.setPrimaryColor(previousTheme.value)
  }
  // 如果之前是暗色主题，现在是明亮主题，则切换回暗色主题
  if (previousThemeState.value && !themeStore.isDark) {
    themeStore.toggleTheme()
  }
})

// 创建助手弹窗可见性
const createDialogVisible = ref(false)

// 搜索关键词
const searchQuery = ref('')

// 标签数据
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

// 助手数据
const assistants = ref([])
const loading = ref(false)

// 当前选中的标签
const currentCategory = ref('all')

// 加载助手列表
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
      assistants.value = response.data.records.map(preset => ({
        ...preset,
        type: preset.official ? 'official' : 'community'
      }))
    } else {
      ElMessage.error(response.msg || '获取助手列表失败')
    }
  } catch (error) {
    console.error('Load assistants error:', error)
    ElMessage.error(error.body?.msg || '获取助手列表失败')
  } finally {
    loading.value = false
  }
}

// 根据搜索和分类过滤助手
const filteredAssistants = computed(() => {
  return assistants.value
})

// 处理搜索
const handleSearch = () => {
  loadAssistants()
}

// 选择标签
const selectCategory = (tagId) => {
  currentCategory.value = tagId
  loadAssistants()
}

const startChat = () => {
  router.push({
    path: '/chat',
    query: {
      newChat: 'true'
    }
  })
}

const createAssistant = () => {
  createDialogVisible.value = true
}

const handleAssistantCreated = () => {
  loadAssistants()
}

const useAssistant = (preset) => {
  router.push({
    path: '/chat',
    query: {
      newChat: 'true',
      assistantId: preset.id,
      assistantName: preset.name,
      assistantAvatar: preset.avatar,
      assistantDescription: preset.description,
      assistantType: preset.type
    }
  })
}

// 获取标签图标
const getTagIcon = (tagId) => {
  const iconMap = {
    tools: Tools,
    writing: Edit,
    coding: Monitor,
    learning: Reading,
    life: House,
    other: More
  }
  return iconMap[tagId] || More
}

// 获取标签名称
const getTagName = (tagId) => {
  const tag = tags.find(t => t.id === tagId)
  return tag ? tag.name : tagId
}

// 获取标签类型
const getTagType = (tagId) => {
  const typeMap = {
    tools: 'warning',
    writing: 'success',
    coding: 'primary',
    learning: 'info',
    life: 'danger',
    other: ''
  }
  return typeMap[tagId] || ''
}

// 编辑助手相关
const editDialogVisible = ref(false)
const currentEditAssistant = ref(null)

const editAssistant = (preset) => {
  currentEditAssistant.value = { ...preset }
  editDialogVisible.value = true
}

const handleAssistantUpdated = () => {
  loadAssistants()
}

// 删除助手
const deleteAssistant = async (preset) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这个助手吗？删除后无法恢复。',
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await defaultApi.apiPresetIdDelete(preset.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      loadAssistants()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete assistant error:', error)
      ElMessage.error(error.body?.msg || '删除失败')
    }
  }
}
</script>

<style lang="scss" scoped>
.assistant-container {
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
    animation: backgroundShift 20s ease-in-out infinite alternate;
  }

  &::after {
    content: '';
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background:
      url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%232B5EFF' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
    opacity: 0.5;
    z-index: 0;
    pointer-events: none;
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

  .nav-header {
    position: relative;
    z-index: 10;
    height: 64px;
    backdrop-filter: blur(10px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .nav-content {
      max-width: 1200px;
      height: 100%;
      margin: 0 auto;
      padding: 0 24px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .nav-left {
        .logo {
          .logo-text {
            font-size: 24px;
            font-weight: bold;
            background: linear-gradient(45deg, #fff, rgba(255, 255, 255, 0.8));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
          }
        }
      }

      .nav-right {
        display: flex;
        align-items: center;
        gap: 24px;

        .github-link {
          display: flex;
          align-items: center;
          color: rgba(255, 255, 255, 0.7);
          transition: color 0.3s ease;

          &:hover {
            color: var(--el-color-primary);
          }

          .nav-icon {
            font-size: 24px;
          }
        }
      }
    }
  }

  .search-content {
    position: relative;
    z-index: 3;
    padding: 40px 20px;
    text-align: center;

    .title {
      font-size: 3.5em;
      margin-bottom: 30px;
      font-weight: 800;
      background: linear-gradient(135deg, #fff 30%, rgba(255, 255, 255, 0.8) 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      text-shadow: 0 0 30px rgba(255, 255, 255, 0.3);
      animation: titleFloat 3s ease-in-out infinite;
    }

    .action-buttons {
      margin-bottom: 30px;
      display: flex;
      justify-content: center;
      gap: 20px;

      .action-btn {
        padding: 12px 32px;
        font-size: 16px;
        font-weight: 500;
        border: none;
        background: rgba(255, 255, 255, 0.15);
        backdrop-filter: blur(10px);
        border-radius: 12px;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        overflow: hidden;

        .el-icon {
          margin-right: 8px;
          transition: transform 0.3s ease;
        }

        &:hover {
          transform: translateY(-2px);
          background: rgba(255, 255, 255, 0.2);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);

          .el-icon {
            transform: scale(1.2);
          }
        }
      }
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
        transition: all 0.3s ease;
        border-radius: 30px;

        &:hover, &:focus-within {
          background: rgba(255, 255, 255, 0.15);
          border-color: rgba(255, 255, 255, 0.3);
          transform: translateY(-2px);
          box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
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

  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 2px;
    background: linear-gradient(90deg,
      rgba(43, 94, 255, 0) 0%,
      rgba(43, 94, 255, 0.5) 50%,
      rgba(43, 94, 255, 0) 100%
    );
  }

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
      transform: translateY(-2px);
    }

    &.active {
      background: linear-gradient(135deg, #2B5EFF, #1E88E5);
      color: white;
      border: none;
      box-shadow: 0 4px 15px rgba(43, 94, 255, 0.35);
    }

    .el-icon {
      margin-right: 8px;
      transition: transform 0.3s ease;
    }

    &:hover .el-icon {
      transform: scale(1.2);
    }
  }
}

.assistants-grid {
  padding: 40px 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 30px;
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
  z-index: 1;

  .preset-card {
    background: var(--el-bg-color);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    cursor: pointer;
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

      .use-button {
        .button-background {
          transform: translateX(0);
        }
      }
    }

    .preset-content {
      display: flex;
      gap: 12px;
    }

    .preset-avatar {
      position: relative;
    }
  }

  .preset-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .preset-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 12px;
    }

    .preset-name-row {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 1;
      min-width: 0;

      .preset-type {
        flex-shrink: 0;
        font-size: 12px;
        padding: 0 8px;
        height: 22px;
        line-height: 20px;
        border-radius: 11px;
        display: inline-flex;
        align-items: center;
        gap: 4px;
        font-weight: 500;

        .tag-icon {
          font-size: 12px;
          margin-right: 2px;
        }

        &.official-tag {
          background: linear-gradient(135deg, #ff9500, #ff3852);
          border: none;
          color: white;
          padding: 0 10px;
          box-shadow: 0 2px 8px rgba(255, 56, 82, 0.25);

          .tag-icon {
            animation: starRotate 4s linear infinite;
          }
        }

        &.community-tag {
          background: rgba(64, 158, 255, 0.1);
          border-color: var(--el-color-primary);
          color: var(--el-color-primary);
          font-weight: 500;
          padding: 0 10px;

          &:hover {
            background: var(--el-color-primary);
            color: white;
          }

          .tag-icon {
            animation: iconBounce 2s ease-in-out infinite;
          }
        }
      }
    }

    .preset-name {
      margin: 0;
      font-size: 15px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .preset-categories {
      display: flex;
      gap: 4px;
      align-items: center;
      margin-top: 2px;

      .category-dot {
        width: 24px;
        height: 24px;
        border-radius: 6px;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.3s;
        background: var(--el-color-primary);
        color: white;

        &:hover {
          transform: translateY(-2px);
          filter: brightness(1.1);
        }

        .category-icon {
          font-size: 14px;
        }

        &.warning {
          background: var(--el-color-warning);
        }

        &.success {
          background: var(--el-color-success);
        }

        &.danger {
          background: var(--el-color-danger);
        }

        &.info {
          background: var(--el-color-info);
        }

        &.primary {
          background: var(--el-color-primary);
        }
      }
    }

    .preset-description {
      margin: 0;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      line-height: 1.5;
    }
  }

  .preset-actions {
    margin-top: auto;

    .action-wrapper {
      display: flex;
      gap: 12px;
      align-items: center;
      
      .action-icons {
        display: flex;
        gap: 8px;

        .icon-button {
          width: 32px;
          height: 32px;
          padding: 0;
          border: 1px solid var(--el-border-color-lighter);
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
          background: var(--el-bg-color);
          opacity: 0.8;

          .el-icon {
            font-size: 14px;
            transition: transform 0.3s ease;
          }

          &:hover {
            transform: translateY(-2px);
            opacity: 1;
            border-color: transparent;

            .el-icon {
              transform: scale(1.1);
            }
          }

          &.edit-button {
            color: var(--el-color-primary);

            &:hover {
              background: var(--el-color-primary-light-9);
              color: var(--el-color-primary);
              box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.2);
            }
          }

          &.delete-button {
            color: var(--el-color-danger);

            &:hover {
              background: var(--el-color-danger-light-9);
              color: var(--el-color-danger);
              box-shadow: 0 4px 12px rgba(var(--el-color-danger-rgb), 0.2);
            }
          }
        }
      }
    }

    .use-button {
      flex: 1;
      border: none;
      background: transparent;
      position: relative;
      overflow: hidden;
      transition: all 0.3s ease;
      padding: 8px 0;

      .button-content {
        position: relative;
        z-index: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 4px;
        font-size: 13px;
      }

      .button-background {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(90deg,
        var(--el-color-primary) 0%,
        var(--el-color-primary-light-3) 50%,
        var(--el-color-primary) 100%
        );
        transform: translateX(-100%);
        transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
      }

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.2);
      }
    }
  }
}

@keyframes backgroundShift {
  0% {
    transform: scale(1) rotate(0deg);
  }
  50% {
    transform: scale(1.2) rotate(5deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}

@keyframes glowPulse {
  0% {
    opacity: 0.5;
    transform: scale(0.95);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.05);
  }
  100% {
    opacity: 0.5;
    transform: scale(0.95);
  }
}

@keyframes titleFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes iconBounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}

@keyframes starRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 响应式布局优化
@media (max-width: 768px) {
  .search-section {
    padding: 40px 20px;

    .title {
      font-size: 2.5em;
    }

    .action-buttons {
      .action-btn {
        padding: 10px 20px;
        font-size: 14px;
      }
    }
  }

  .tag-nav {
    padding: 15px 10px;

    .tag-item {
      padding: 8px 16px;
      font-size: 14px;
    }
  }

  .assistants-grid {
    padding: 20px 10px;
    gap: 20px;
  }
}

.el-empty {
  grid-column: 1 / -1;
  margin: 40px 0;
}
</style>
