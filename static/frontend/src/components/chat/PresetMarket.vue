<template>
  <el-dialog
    v-model="dialogVisible"
    title="AI 助手市场"
    width="800px"
    class="preset-market-dialog"
    :close-on-click-modal="false"
  >
    <!-- 搜索区域 -->
    <div class="search-area">
      <el-input
        v-model="searchQuery"
        placeholder="搜索 AI 助手..."
        class="search-input"
        clearable
      >
        <template #prefix>
          <el-icon class="search-icon"><Search /></el-icon>
        </template>
      </el-input>
      <el-radio-group v-model="filterType" class="filter-group">
        <el-radio-button label="all">全部</el-radio-button>
        <el-radio-button label="official">官方</el-radio-button>
        <el-radio-button label="community">社区</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 助手列表 -->
    <div class="presets-grid">
      <div
        v-for="preset in filteredPresets"
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
            <div class="preset-name-row">
              <h3 class="preset-name">{{ preset.name }}</h3>
              <div class="preset-type" :class="preset.type">
                {{ preset.type === 'official' ? '官方' : '社区' }}
              </div>
            </div>
            <p class="preset-description">{{ preset.description }}</p>
          </div>
        </div>
        <div class="preset-actions">
          <el-button
            type="primary"
            class="use-button"
            @click.stop="selectPreset(preset)"
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

    <!-- 分页 -->
    <div class="pagination-area">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[12, 24, 36, 48]"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Search, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import {useUserStore} from "@/stores/user.js";

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:visible', 'select'])

// 双向绑定对话框可见性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

// 搜索和筛选
const searchQuery = ref('')
const filterType = ref('all')
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)

// 助手数据
const presets = ref([])
const loading = ref(false)

const userStore = useUserStore()

// 加载助手列表
const loadPresets = async () => {
  try {
    loading.value = true
    const params = {
      orderBy: 'id',
      order: 'desc',
      name: searchQuery.value || undefined,
      official: filterType.value === 'official' ? true : filterType.value === 'community' ? false : undefined,
      page: currentPage.value,
      limit: pageSize.value
    }
    const response = await defaultApi.apiPresetListGet(
      currentPage.value,
      pageSize.value,
      params
    )
    if (response.code === 0 && response.data) {
      presets.value = (response.data.records || []).map(preset => ({
        ...preset,
        type: preset.official ? 'official' : 'community'
      }))
      total.value = response.data.total || 0
    } else {
      ElMessage.error(response.msg || '获取助手列表失败')
    }
  } catch (error) {
    console.error('Load presets error:', error)
    ElMessage.error(error.body?.msg || '获取助手列表失败')
  } finally {
    loading.value = false
  }
}

// 过滤预设（仅本地过滤搜索和类型，已由后端处理）
const filteredPresets = computed(() => presets.value)

// 选择预设
const selectPreset = (preset) => {
  emit('select', preset)
  dialogVisible.value = false
  ElMessage.success(`已选择 ${preset.name}`)
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
  loadPresets()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadPresets()
}

// 搜索和筛选变化时重新加载
watch([searchQuery, filterType], () => {
  currentPage.value = 1
  loadPresets()
})

onMounted(() => {

  // 登陆后才允许加载助手列表
  if (userStore.isLoggedIn) {
    loadPresets()
  }
})
</script>

<style scoped lang="scss">
.preset-market-dialog {

  // TODO 解决这里样式不生效的问题
  :deep(.el-dialog__body) {
    padding: 0;
    overflow-y: initial;
  }
}
.search-area {
  padding: 16px 20px;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  gap: 16px;
  align-items: center;
  background: var(--el-bg-color);

  .search-input {
    max-width: 300px;

    :deep(.el-input__wrapper) {
      box-shadow: 0 0 0 1px var(--el-border-color-light) inset;
      transition: all 0.3s ease;

      &:hover {
        box-shadow: 0 0 0 1px var(--el-color-primary) inset;
      }

      &.is-focus {
        box-shadow: 0 0 0 1px var(--el-color-primary) inset;
      }
    }

    .search-icon {
      color: var(--el-text-color-secondary);
      transition: color 0.3s ease;
    }

    &:hover .search-icon,
    &:focus-within .search-icon {
      color: var(--el-color-primary);
    }
  }

  .filter-group {
    margin-left: auto;
  }
}

.presets-grid {
  padding: 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
  //min-height: 400px;
  max-height: 600px;
  overflow-y: auto;

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

    .preset-name-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 4px;
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

    .preset-type {
      font-size: 10px;
      padding: 1px 6px;
      border-radius: 10px;
      color: white;
      font-weight: 500;
      flex-shrink: 0;

      &.official {
        background: linear-gradient(135deg, #4158D0, #C850C0);
      }

      &.community {
        background: linear-gradient(135deg, #11998e, #38ef7d);
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

    .use-button {
      width: 100%;
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

.pagination-area {
  padding: 16px 20px;
  display: flex;
  justify-content: center;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-light);
}

// 添加进入和离开动画
.preset-card-enter-active,
.preset-card-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.preset-card-enter-from,
.preset-card-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>
