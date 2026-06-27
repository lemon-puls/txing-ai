<template>
  <transition name="popup-fade">
    <div v-if="visible" class="app-mention-popup">
      <div class="popup-header" @mousedown.prevent>
        <el-input
          v-model="searchQuery"
          placeholder="搜索应用..."
          size="small"
          clearable
          ref="searchInput"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      <div class="popup-list" ref="listRef" @mousedown.prevent>
        <div v-if="loading" class="popup-loading">
          <el-skeleton :rows="3" animated />
        </div>
        <template v-else>
          <div
            v-for="app in filteredApps"
            :key="app.id"
            class="app-item"
            :class="{ active: selectedIndex === filteredApps.indexOf(app) }"
            @click.stop="selectApp(app)"
            @mouseenter="selectedIndex = filteredApps.indexOf(app)"
          >
            <div class="app-icon">
              <el-icon :size="18"><Share /></el-icon>
            </div>
            <div class="app-info">
              <div class="app-name">{{ app.name }}</div>
              <div class="app-desc">{{ app.description || '暂无描述' }}</div>
            </div>
            <el-tag v-if="app.category" size="small" type="info" effect="plain" round>
              {{ getCategoryLabel(app.category) }}
            </el-tag>
          </div>
          <el-empty v-if="filteredApps.length === 0" description="未找到应用" :image-size="60" />
        </template>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { Search, Share } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const props = defineProps({
  visible: { type: Boolean, default: false }
})

const emit = defineEmits(['select', 'close'])

const searchQuery = ref('')
const apps = ref([])
const loading = ref(false)
const selectedIndex = ref(0)
const searchInput = ref(null)
const listRef = ref(null)

const filteredApps = computed(() => {
  if (!searchQuery.value) return apps.value
  const q = searchQuery.value.toLowerCase()
  return apps.value.filter(app =>
    app.name.toLowerCase().includes(q) ||
    (app.description && app.description.toLowerCase().includes(q))
  )
})

const categoryLabels = {
  general: '通用',
  qa: '问答',
  writing: '写作',
  data: '数据分析',
  tool: '工具',
  tools: '工具',
  other: '其他'
}

const getCategoryLabel = (category) => categoryLabels[category] || category

const loadApps = async () => {
  loading.value = true
  try {
    const res = await defaultApi.apiWorkflowPublicGet(1, 100)
    if (res.code === 0 && res.data) {
      apps.value = res.data.records || []
    }
  } catch (e) {
    console.error('加载应用列表失败:', e)
  } finally {
    loading.value = false
  }
}

const selectApp = (app) => {
  emit('select', { ...app })
  searchQuery.value = ''
  selectedIndex.value = 0
}

const handleKeydown = (e) => {
  if (!props.visible) return

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filteredApps.value.length - 1)
    scrollToSelected()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
    scrollToSelected()
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (filteredApps.value[selectedIndex.value]) {
      selectApp(filteredApps.value[selectedIndex.value])
    }
  } else if (e.key === 'Escape') {
    emit('close')
  }
}

const scrollToSelected = () => {
  nextTick(() => {
    const list = listRef.value
    if (!list) return
    const items = list.querySelectorAll('.app-item')
    if (items[selectedIndex.value]) {
      items[selectedIndex.value].scrollIntoView({ block: 'nearest' })
    }
  })
}

watch(() => props.visible, (val) => {
  if (val) {
    loadApps()
    selectedIndex.value = 0
    nextTick(() => searchInput.value?.focus())
  } else {
    searchQuery.value = ''
  }
})

defineExpose({ handleKeydown })
</script>

<style lang="scss" scoped>
.app-mention-popup {
  position: absolute;
  bottom: 100%;
  left: 0;
  right: 0;
  max-height: 320px;
  background: var(--el-bg-color, #fff);
  border: 1px solid var(--el-border-color-light, #dcdfe6);
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 100;
  margin-bottom: 8px;
}

.popup-header {
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color-extra-light, #f0f0f0);
}

.popup-list {
  overflow-y: auto;
  flex: 1;
  padding: 6px;
}

.popup-loading {
  padding: 12px;
}

.app-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;

  &:hover, &.active {
    background: var(--el-fill-color-light, #f5f7fa);
  }

  .app-icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: rgba(25, 118, 210, 0.08);
    color: #1976d2;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .app-info {
    flex: 1;
    min-width: 0;

    .app-name {
      font-size: 13px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .app-desc {
      font-size: 12px;
      color: var(--el-text-color-secondary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      margin-top: 2px;
    }
  }
}

.popup-fade-enter-active,
.popup-fade-leave-active {
  transition: all 0.2s ease;
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
