<template>
  <div class="workflow-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="工作流名称">
          <el-input v-model="searchForm.name" placeholder="请输入工作流名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleReset">
            <el-icon><RefreshRight /></el-icon>
            重置
          </el-button>
          <el-button type="success" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增工作流
          </el-button>
          <el-button type="warning" @click="goToTemplateMarket">
            <el-icon><Folder /></el-icon>
            模板市场
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工作流卡片网格 -->
    <div class="workflow-grid" v-loading="loading">
      <!-- 空状态 -->
      <div v-if="workflows.length === 0 && !loading" class="empty-state">
        <div class="empty-icon-wrapper">
          <el-icon :size="64"><FolderOpened /></el-icon>
        </div>
        <p class="empty-title">暂无工作流</p>
        <p class="empty-desc">点击「新增工作流」开始创建您的第一个工作流</p>
      </div>

      <!-- 工作流卡片 -->
      <div
        v-for="(wf, index) in workflows"
        :key="wf.id"
        class="workflow-card"
        :style="{ animationDelay: `${index * 80}ms` }"
      >
        <!-- 顶部渐变色条 -->
        <div
          class="card-top-line"
          :class="{ published: wf.status === 'published' }"
        />

        <div class="card-main">
          <!-- 图标 -->
          <div class="card-icon-wrapper">
            <div class="card-icon" :class="{ published: wf.status === 'published' }">
              <el-icon :size="24"><Connection /></el-icon>
            </div>
          </div>

          <!-- 信息区 -->
          <div class="card-info">
            <div class="card-title-row">
              <span class="card-name">{{ wf.name }}</span>
              <span
                class="status-badge"
                :class="{ published: wf.status === 'published' }"
              >
                <span class="status-dot" />
                {{ wf.status === 'published' ? '已发布' : '草稿' }}
              </span>
            </div>

            <p class="card-desc">{{ wf.description || '暂无描述' }}</p>

            <div class="card-meta">
              <span class="meta-item">
                <el-icon><Clock /></el-icon>
                {{ formatTime(wf.created_at) }}
              </span>
              <span class="meta-item">
                <el-icon><Share /></el-icon>
                {{ getNodeCount(wf.topology) }} 个节点
              </span>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="card-actions">
          <el-button type="primary" size="small" plain @click="handleEdit(wf)">
            <el-icon><EditPen /></el-icon>
            编辑信息
          </el-button>
          <el-button type="success" size="small" plain @click="handleDesign(wf)">
            <el-icon><Setting /></el-icon>
            设计流程
          </el-button>
          <el-button
            :type="wf.status === 'published' ? 'warning' : 'success'"
            size="small"
            plain
            @click="handleToggleStatus(wf)"
          >
            <el-icon><component :is="wf.status === 'published' ? 'Close' : 'Check'" /></el-icon>
            {{ wf.status === 'published' ? '取消发布' : '发布' }}
          </el-button>
          <el-button type="danger" size="small" plain @click="handleDelete(wf)">
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
        :page-sizes="[8, 12, 24, 48]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 新增/编辑基本信息对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增工作流' : '编辑工作流'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入工作流名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入工作流描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder, FolderOpened, Search, RefreshRight, Plus,
  Connection, Clock, Share, EditPen, Setting, Delete,
  Close, Check
} from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const router = useRouter()

// 搜索表单
const searchForm = ref({
  name: ''
})

// 列表数据
const loading = ref(false)
const workflows = ref([])
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)

// 对话框数据
const dialogVisible = ref(false)
const dialogType = ref('add')
const submitLoading = ref(false)
const formRef = ref(null)
const form = ref({
  name: '',
  description: ''
})

const rules = {
  name: [
    { required: true, message: '请输入工作流名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ]
}

// 解析节点数量
const getNodeCount = (topology) => {
  if (!topology) return 0
  try {
    const parsed = typeof topology === 'string' ? JSON.parse(topology) : topology
    return parsed.nodes?.length || 0
  } catch {
    return 0
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', { hour12: false })
}

// 加载列表
const loadData = async () => {
  loading.value = true
  try {
    const res = await defaultApi.apiWorkflowGet(currentPage.value, pageSize.value, {
      name: searchForm.value.name || undefined
    })
    if (res.code === 0 && res.data) {
      workflows.value = res.data.records || []
      total.value = res.data.total || 0
    } else {
      ElMessage.error(res.msg || '获取列表失败')
    }
  } catch (error) {
    console.error('Load error:', error)
    ElMessage.error('获取列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadData()
}

const handleReset = () => {
  searchForm.value.name = ''
  handleSearch()
}

const handleAdd = () => {
  dialogType.value = 'add'
  form.value = {
    name: '',
    description: ''
  }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogType.value = 'edit'
  form.value = {
    id: row.id,
    name: row.name,
    description: row.description,
    topology: row.topology
  }
  dialogVisible.value = true
}

const handleDesign = (row) => {
  router.push(`/admin/workflow/editor/${row.id}`)
}

const goToTemplateMarket = () => {
  router.push('/admin/workflow/templates')
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该工作流吗？删除后不可恢复！', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await defaultApi.apiWorkflowIdDelete(row.id)
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      console.error('Delete error:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const handleToggleStatus = async (row) => {
  const newStatus = row.status === 'published' ? 'draft' : 'published'
  const actionText = newStatus === 'published' ? '发布' : '取消发布'
  try {
    await ElMessageBox.confirm(`确认${actionText}该工作流吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    })
    const res = await defaultApi.apiWorkflowIdStatusPut(row.id, { status: newStatus })
    if (res.code === 0) {
      ElMessage.success(`${actionText}成功`)
      loadData()
    } else {
      ElMessage.error(res.msg || `${actionText}失败`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Toggle status error:', error)
      ElMessage.error(`${actionText}失败`)
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        const data = {
          name: form.value.name,
          description: form.value.description
        }

        let res
        if (dialogType.value === 'add') {
          data.topology = ''
          res = await defaultApi.apiWorkflowPost(data)
        } else {
          data.topology = form.value.topology || ''
          res = await defaultApi.apiWorkflowIdPut(form.value.id, data)
        }

        if (res.code === 0) {
          ElMessage.success(dialogType.value === 'add' ? '创建成功' : '修改成功')
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.msg || '操作失败')
        }
      } catch (error) {
        console.error('Submit error:', error)
        ElMessage.error('操作失败')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleSizeChange = (val) => {
  pageSize.value = val
  loadData()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
// 动画关键帧
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes cardFadeIn {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.6); }
}

@keyframes breathe {
  0%, 100% { transform: scale(1); opacity: 0.6; }
  50% { transform: scale(1.08); opacity: 1; }
}

.workflow-container {
  padding: 24px;
  animation: fadeInUp 0.6s ease-out;

  // 搜索表单
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
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
      }

      &.el-button--primary:hover {
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.35);
      }
      &.el-button--success:hover {
        box-shadow: 0 4px 12px rgba(var(--el-color-success-rgb), 0.35);
      }
      &.el-button--warning:hover {
        box-shadow: 0 4px 12px rgba(var(--el-color-warning-rgb), 0.35);
      }
    }
  }

  // 卡片网格
  .workflow-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
    gap: 20px;
    min-height: 200px;
  }

  // 空状态
  .empty-state {
    grid-column: 1 / -1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 0;
    color: var(--el-text-color-placeholder);

    .empty-icon-wrapper {
      animation: breathe 3s ease-in-out infinite;
      margin-bottom: 16px;
    }

    .empty-title {
      font-size: 18px;
      font-weight: 600;
      color: var(--el-text-color-secondary);
      margin: 0 0 8px;
    }

    .empty-desc {
      font-size: 14px;
      margin: 0;
    }
  }

  // 工作流卡片
  .workflow-card {
    background: var(--el-bg-color);
    border-radius: 16px;
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
    animation: cardFadeIn 0.5s ease-out both;
    position: relative;

    &:hover {
      transform: translateY(-6px) scale(1.02);
      box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

      .card-top-line {
        transform: scaleX(1);
      }

      .card-icon {
        transform: scale(1.15) rotate(5deg);
      }

      .card-actions {
        opacity: 1;
        transform: translateY(0);
      }
    }

    // 顶部渐变色条
    .card-top-line {
      height: 4px;
      background: linear-gradient(90deg, #909399, #c0c4cc);
      transform: scaleX(0.6);
      transform-origin: left;
      transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);

      &.published {
        background: linear-gradient(90deg, #409eff, #67c23a, #e6a23c);
        transform: scaleX(1);
      }
    }

    // 卡片主体
    .card-main {
      display: flex;
      gap: 16px;
      padding: 20px 20px 0;
      flex: 1;
    }

    // 图标
    .card-icon-wrapper {
      flex-shrink: 0;
    }

    .card-icon {
      width: 52px;
      height: 52px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #e8eaed, #d0d3d8);
      color: #606266;
      transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);

      &.published {
        background: linear-gradient(135deg, #409eff, #7c4dff);
        color: #fff;
        box-shadow: 0 4px 16px rgba(64, 158, 255, 0.3);
      }
    }

    // 信息区
    .card-info {
      flex: 1;
      min-width: 0;
    }

    .card-title-row {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 8px;
    }

    .card-name {
      font-size: 16px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    // 状态徽章
    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      padding: 2px 10px;
      border-radius: 20px;
      background: var(--el-fill-color-light);
      color: var(--el-text-color-secondary);
      flex-shrink: 0;

      &.published {
        background: rgba(103, 194, 58, 0.1);
        color: #67c23a;

        .status-dot {
          background: #67c23a;
          box-shadow: 0 0 0 3px rgba(103, 194, 58, 0.2);
          animation: pulse 2s ease-in-out infinite;
        }
      }

      .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: #909399;
        transition: all 0.3s;
      }
    }

    .card-desc {
      font-size: 13px;
      color: var(--el-text-color-secondary);
      line-height: 1.6;
      margin: 0 0 12px;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .card-meta {
      display: flex;
      gap: 16px;
      flex-wrap: wrap;

      .meta-item {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--el-text-color-placeholder);

        .el-icon {
          font-size: 14px;
        }
      }
    }

    // 操作按钮
    .card-actions {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 14px 20px;
      margin-top: 12px;
      border-top: 1px solid var(--el-border-color-lighter);
      opacity: 0.7;
      transform: translateY(4px);
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      flex-wrap: wrap;

      :deep(.el-button) {
        border-radius: 8px;
        transition: all 0.3s;

        &:hover {
          transform: translateY(-1px);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
      }
    }
  }

  // 分页
  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 24px;

    :deep(.el-pagination) {
      padding: 12px 24px;
      border-radius: 12px;
      background: var(--el-bg-color);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);

      .el-pagination__sizes {
        .el-input__wrapper {
          border-radius: 8px;
        }
      }

      button {
        border-radius: 8px;
        transition: all 0.3s;

        &:hover {
          transform: translateY(-1px);
        }
      }

      .el-pager li {
        border-radius: 8px;
        transition: all 0.3s;

        &:hover {
          transform: translateY(-1px);
        }

        &.is-active {
          font-weight: 600;
        }
      }
    }
  }
}
</style>
