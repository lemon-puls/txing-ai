<template>
  <div class="template-market-container">
    <!-- 顶部搜索 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="模板名称">
          <el-input v-model="searchForm.name" placeholder="搜索模板" clearable />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchForm.category" placeholder="全部分类" clearable>
            <el-option label="通用" value="general" />
            <el-option label="问答" value="qa" />
            <el-option label="写作" value="writing" />
            <el-option label="数据分析" value="data" />
            <el-option label="工具调用" value="tool" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="default" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回工作流列表
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 模板列表 -->
    <div class="template-grid" v-loading="loading">
      <div v-if="templates.length === 0 && !loading" class="empty-state">
        <el-icon :size="48"><Folder /></el-icon>
        <p>暂无模板</p>
      </div>

      <div
        v-for="tpl in templates"
        :key="tpl.id"
        class="template-card"
      >
        <div class="card-header">
          <div class="card-title">
            <span class="name">{{ tpl.name }}</span>
            <el-tag v-if="tpl.category" size="small" type="info">{{ getCategoryLabel(tpl.category) }}</el-tag>
          </div>
        </div>

        <div class="card-body">
          <p class="description">{{ tpl.description || '暂无描述' }}</p>
          <div class="meta">
            <span class="time">
              <el-icon><Clock /></el-icon>
              {{ formatTime(tpl.createTime) }}
            </span>
          </div>
        </div>

        <div class="card-footer">
          <el-button type="primary" size="small" @click="handleUseTemplate(tpl)">
            <el-icon><CopyDocument /></el-icon>
            使用此模板
          </el-button>
          <el-button type="danger" size="small" text @click="handleDeleteTemplate(tpl)">
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-container" v-if="total > 0">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[12, 24, 48]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Folder, Clock, CopyDocument, Delete } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const router = useRouter()

// 搜索
const searchForm = ref({ name: '', category: '' })

// 列表
const loading = ref(false)
const templates = ref([])
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)

// 分类标签
const categoryLabels = {
  general: '通用',
  qa: '问答',
  writing: '写作',
  data: '数据分析',
  tool: '工具调用',
  other: '其他'
}

const getCategoryLabel = (category) => categoryLabels[category] || category

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', { hour12: false })
}

// 加载模板列表
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }
    if (searchForm.value.name) params.name = searchForm.value.name
    if (searchForm.value.category) params.category = searchForm.value.category

    const res = await defaultApi.apiWorkflowTemplatesGet(params.page, params.limit, {
      name: params.name,
      category: params.category
    })
    const result = res.data || res
    templates.value = result.records || []
    total.value = result.total || 0
  } catch (error) {
    console.error('加载模板列表失败:', error)
    ElMessage.error('加载模板列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadData()
}

const handleReset = () => {
  searchForm.value = { name: '', category: '' }
  handleSearch()
}

// 使用模板（克隆为新工作流）
const handleUseTemplate = async (tpl) => {
  try {
    const { value: name } = await ElMessageBox.prompt(
      '请输入新工作流的名称',
      '使用模板',
      {
        confirmButtonText: '创建',
        cancelButtonText: '取消',
        inputValue: tpl.name + ' (副本)',
        inputPlaceholder: '工作流名称'
      }
    )

    if (!name) return

    const res = await defaultApi.apiWorkflowTemplatesClonePost({
      templateId: tpl.id,
      name: name
    })
    if (res.code === 0) {
      ElMessage.success('工作流创建成功')
      // 跳转到新工作流编辑器
      const newId = res.data?.id
      if (newId) {
        router.push(`/admin/workflow/editor/${newId}`)
      }
    } else {
      ElMessage.error(res.msg || '创建失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('使用模板失败:', error)
      ElMessage.error('使用模板失败')
    }
  }
}

// 删除模板
const handleDeleteTemplate = async (tpl) => {
  try {
    await ElMessageBox.confirm(
      `确定删除模板"${tpl.name}"吗？`,
      '删除确认',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )

    // 没有专门的删除模板接口，使用通用的删除工作流接口
    // 后端 DeleteTemplate service 是软删除（设置 is_template=false）
    // 但 controller 没有暴露该接口，这里调用通用删除
    const res = await defaultApi.apiWorkflowIdDelete(tpl.id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      loadData()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模板失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleSizeChange = (val) => {
  pageSize.value = val
  loadData()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadData()
}

const goBack = () => {
  router.push('/admin/workflow')
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.template-market-container {
  padding: 24px;

  .search-form {
    margin-bottom: 24px;
    border-radius: 16px;
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);

    :deep(.el-form-item) {
      margin-bottom: 0;
    }
    :deep(.el-input__wrapper) {
      border-radius: 12px;
    }
    :deep(.el-button) {
      border-radius: 12px;
      padding: 12px 24px;
    }
  }

  .template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
    min-height: 200px;

    .empty-state {
      grid-column: 1 / -1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 60px 0;
      color: #909399;

      p { margin-top: 12px; font-size: 14px; }
    }

    .template-card {
      background: #fff;
      border-radius: 14px;
      border: 1px solid #e4e7ed;
      overflow: hidden;
      transition: all 0.2s ease;
      display: flex;
      flex-direction: column;

      &:hover {
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
        border-color: #c0c4cc;
      }

      .card-header {
        padding: 16px 20px 0;

        .card-title {
          display: flex;
          align-items: center;
          gap: 8px;

          .name {
            font-size: 16px;
            font-weight: 600;
            color: #303133;
          }
        }
      }

      .card-body {
        padding: 12px 20px;
        flex: 1;

        .description {
          font-size: 13px;
          color: #606266;
          line-height: 1.5;
          margin: 0 0 8px;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }

        .meta {
          font-size: 12px;
          color: #909399;

          .time {
            display: flex;
            align-items: center;
            gap: 4px;
          }
        }
      }

      .card-footer {
        padding: 12px 20px;
        border-top: 1px solid #f0f0f0;
        display: flex;
        align-items: center;
        justify-content: space-between;
      }
    }
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 24px;
  }
}
</style>
