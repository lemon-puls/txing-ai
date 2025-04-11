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
        :prefix-icon="Search"
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
import { ref, computed } from 'vue'
import { Search, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

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

// 模拟数据
const presets = ref([
  {
    id: 1,
    name: '全能代码助手',
    description: '精通前后端开发，帮你解决各类编程难题，优化代码质量，提供最佳实践建议。',
    avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    type: 'official'
  },
  {
    id: 2,
    name: '英语教练',
    description: '专业英语教师，帮你提高口语、写作和语法，准备雅思托福考试。',
    avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
    type: 'official'
  },
  {
    id: 3,
    name: '数学导师',
    description: '擅长微积分、线性代数、概率统计等，深入浅出讲解数学概念。',
    avatar: 'https://cube.elemecdn.com/9/c2/f0ee8a3c7c9638a54940382568c9dpng.png',
    type: 'official'
  },
  {
    id: 4,
    name: '写作助手',
    description: '帮你润色文章，改进写作风格，提供创意灵感和写作建议。',
    avatar: 'https://cube.elemecdn.com/d/e6/c4d93a3805b3ce3f323f7974e6f78png.png',
    type: 'official'
  },
  {
    id: 5,
    name: '产品经理',
    description: '协助产品规划、需求分析、用户研究，提供产品设计建议。',
    avatar: 'https://cube.elemecdn.com/6/94/4d3ea53c084bad6931a56d5158a48jpeg.jpeg',
    type: 'official'
  },
  {
    id: 6,
    name: '心理咨询师',
    description: '提供情感咨询、压力管理、心理疏导等服务，帮你保持心理健康。',
    avatar: 'https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png',
    type: 'official'
  },
  {
    id: 7,
    name: '健身教练',
    description: '制定个性化健身计划，指导运动姿势，提供营养建议。',
    avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    type: 'official'
  },
  {
    id: 8,
    name: '职业规划师',
    description: '帮你分析职业发展方向，提供求职建议，优化简历面试。',
    avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
    type: 'community'
  },
  {
    id: 9,
    name: '设计导师',
    description: 'UI/UX 设计指导，把控设计趋势，提供设计改进建议。',
    avatar: 'https://cube.elemecdn.com/9/c2/f0ee8a3c7c9638a54940382568c9dpng.png',
    type: 'official'
  },
  {
    id: 10,
    name: '创业顾问',
    description: '创业全程指导，商业模式分析，风险控制建议。',
    avatar: 'https://cube.elemecdn.com/d/e6/c4d93a3805b3ce3f323f7974e6f78png.png',
    type: 'community'
  },
  {
    id: 11,
    name: '理财专家',
    description: '个人理财规划，投资组合建议，风险管理指导。',
    avatar: 'https://cube.elemecdn.com/6/94/4d3ea53c084bad6931a56d5158a48jpeg.jpeg',
    type: 'community'
  },
  {
    id: 12,
    name: '旅行规划师',
    description: '定制旅行路线，景点推荐，旅行攻略分享。',
    avatar: 'https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png',
    type: 'community'
  },
  {
    id: 13,
    name: '营养师',
    description: '膳食营养搭配，健康饮食建议，体重管理指导。',
    avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    type: 'community'
  },
  {
    id: 14,
    name: '法律顾问',
    description: '提供法律咨询，合同审查建议，法律风险提醒。',
    avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
    type: 'official'
  },
  {
    id: 15,
    name: '市场专家',
    description: '市场分析，营销策略，品牌建设指导。',
    avatar: 'https://cube.elemecdn.com/9/c2/f0ee8a3c7c9638a54940382568c9dpng.png',
    type: 'community'
  },
  {
    id: 16,
    name: '学习规划师',
    description: '学习方法指导，时间管理，考试备考建议。',
    avatar: 'https://cube.elemecdn.com/d/e6/c4d93a3805b3ce3f323f7974e6f78png.png',
    type: 'community'
  },
  {
    id: 17,
    name: '生活管家',
    description: '日常生活建议，家居整理，时间规划指导。',
    avatar: 'https://cube.elemecdn.com/6/94/4d3ea53c084bad6931a56d5158a48jpeg.jpeg',
    type: 'community'
  },
  {
    id: 18,
    name: '艺术导师',
    description: '艺术创作指导，艺术鉴赏，创意灵感激发。',
    avatar: 'https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png',
    type: 'community'
  },
  {
    id: 19,
    name: '科技资讯师',
    description: '最新科技动态解读，技术趋势分析，创新见解分享。',
    avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    type: 'official'
  },
  {
    id: 20,
    name: '情感咨询师',
    description: '情感关系指导，沟通技巧建议，人际关系处理。',
    avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
    type: 'community'
  }
])

// 更新总数
total.value = presets.value.length

// 过滤预设
const filteredPresets = computed(() => {
  return presets.value.filter(preset => {
    const matchesSearch = preset.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                         preset.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = filterType.value === 'all' || preset.type === filterType.value
    return matchesSearch && matchesType
  })
})

// 选择预设
const selectPreset = (preset) => {
  emit('select', preset)
  dialogVisible.value = false
  ElMessage.success(`已选择 ${preset.name}`)
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  // TODO: 重新加载数据
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  // TODO: 重新加载数据
}
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
