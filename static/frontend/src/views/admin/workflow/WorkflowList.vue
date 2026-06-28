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
        <el-empty description="暂无工作流">
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增工作流
          </el-button>
        </el-empty>
      </div>

      <!-- 工作流卡片 -->
      <div
        v-for="(wf, index) in workflows"
        :key="wf.id"
        class="workflow-card"
        :class="{ published: wf.status === 'published' }"
        :style="{ '--delay': `${index * 0.06}s` }"
      >

        <!-- 卡片内容 -->
        <div class="card-body">
          <div class="card-header">
            <div class="card-icon" :class="{ published: wf.status === 'published' }">
              <el-icon :size="18"><Connection /></el-icon>
            </div>
            <div class="card-title-area">
              <h3 class="workflow-name">{{ wf.name }}</h3>
              <span
                class="status-tag"
                :class="{ published: wf.status === 'published' }"
              >
                {{ wf.status === 'published' ? '已发布' : '草稿' }}
              </span>
            </div>
          </div>
          <p class="workflow-desc">{{ wf.description || '暂无描述' }}</p>
        </div>

        <!-- 卡片底部 -->
        <div class="card-footer">
          <div class="footer-meta">
            <span class="meta-item">
              <el-icon><Clock /></el-icon>
              {{ formatTime(wf.created_at) }}
            </span>
            <span class="meta-sep">·</span>
            <span class="meta-item">
              {{ getNodeCount(wf.topology) }} 个节点
            </span>
          </div>
          <div class="footer-actions">
            <el-tooltip content="编辑信息" placement="top">
              <el-button class="action-btn" circle size="small" @click.stop="handleEdit(wf)">
                <el-icon :size="14"><EditPen /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="设计流程" placement="top">
              <el-button class="action-btn primary" circle size="small" @click.stop="handleDesign(wf)">
                <el-icon :size="14"><Setting /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip :content="wf.status === 'published' ? '取消发布' : '发布'" placement="top">
              <el-button
                class="action-btn"
                :class="{ warning: wf.status === 'published', success: wf.status !== 'published' }"
                circle
                size="small"
                @click.stop="handleToggleStatus(wf)"
              >
                <el-icon :size="14">
                  <component :is="wf.status === 'published' ? 'Close' : 'Check'" />
                </el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button class="action-btn danger" circle size="small" @click.stop="handleDelete(wf)">
                <el-icon :size="14"><Delete /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
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
  Folder, Search, RefreshRight, Plus,
  Connection, Clock, EditPen, Setting, Delete,
  Close, Check
} from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

const router = useRouter()

// 搜索表单
const searchForm = ref({ name: '' })

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
const form = ref({ name: '', description: '' })

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

const handleSearch = () => { currentPage.value = 1; loadData() }
const handleReset = () => { searchForm.value.name = ''; handleSearch() }

const handleAdd = () => {
  dialogType.value = 'add'
  form.value = { name: '', description: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogType.value = 'edit'
  form.value = { id: row.id, name: row.name, description: row.description, topology: row.topology }
  dialogVisible.value = true
}

const handleDesign = (row) => { router.push(`/admin/workflow/editor/${row.id}`) }
const goToTemplateMarket = () => { router.push('/admin/workflow/templates') }

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该工作流吗？删除后不可恢复！', '警告', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(async () => {
    try {
      const res = await defaultApi.apiWorkflowIdDelete(row.id)
      if (res.code === 0) { ElMessage.success('删除成功'); loadData() }
      else ElMessage.error(res.msg || '删除失败')
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
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'info'
    })
    const res = await defaultApi.apiWorkflowIdStatusPut(row.id, { status: newStatus })
    if (res.code === 0) { ElMessage.success(`${actionText}成功`); loadData() }
    else ElMessage.error(res.msg || `${actionText}失败`)
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
        const data = { name: form.value.name, description: form.value.description }
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

const handleSizeChange = (val) => { pageSize.value = val; loadData() }
const handleCurrentChange = (val) => { currentPage.value = val; loadData() }

onMounted(() => { loadData() })
</script>

<style lang="scss" scoped>
// 主题色
$blue-500: #2B5EFF;
$blue-400: #4facfe;

// 状态色
$published: #10b981;
$draft: #94a3b8;

.workflow-container {
  padding: 24px;
}

// 搜索表单
.search-form {
  margin-bottom: 24px;
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);

  :deep(.el-form-item) { margin-bottom: 0; }
  :deep(.el-input__wrapper) { border-radius: 8px; }
  :deep(.el-button) {
    border-radius: 8px;
    padding: 8px 20px;
    transition: all 0.2s ease;
    &:hover { transform: translateY(-1px); box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08); }
  }
}

// 卡片网格
.workflow-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  min-height: 200px;
}

// 空状态
.empty-state {
  grid-column: 1 / -1;
  padding: 60px 0;
}

// 工作流卡片
.workflow-card {
  position: relative;
  background: var(--el-bg-color, #fff);
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--el-border-color-lighter, #e4e7ed);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  animation: cardFadeIn 0.4s ease-out var(--delay, 0s) both;

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
    border-color: var(--el-border-color, #c0c4cc);

    .action-btn { opacity: 1; }
  }


  // 卡片内容
  .card-body {
    padding: 20px 20px 16px;
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 12px;
  }

  // 图标
  .card-icon {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: $draft;
    color: #fff;
    flex-shrink: 0;
    transition: all 0.25s ease;

    &.published {
      background: $published;
    }

    .workflow-card:hover & {
      transform: scale(1.05);
    }
  }

  .card-title-area {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .workflow-name {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  // 状态标签
  .status-tag {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
    line-height: 1.5;
    width: fit-content;
    background: rgba($draft, 0.1);
    color: $draft;

    &.published {
      background: rgba($published, 0.1);
      color: $published;
    }
  }

  .workflow-desc {
    margin: 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  // 卡片底部
  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    border-top: 1px solid var(--el-border-color-extra-light, #f0f0f0);
  }

  .footer-meta {
    display: flex;
    align-items: center;
    gap: 6px;

    .meta-item {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      color: var(--el-text-color-placeholder);

      .el-icon { font-size: 13px; }
    }

    .meta-sep {
      color: var(--el-border-color);
      font-size: 10px;
    }
  }

  .footer-actions {
    display: flex;
    gap: 4px;
  }

  // 操作按钮（圆形图标按钮）
  .action-btn {
    border: 1px solid var(--el-border-color-lighter) !important;
    background: #fff !important;
    color: var(--el-text-color-placeholder) !important;
    opacity: 0.5;
    transition: all 0.2s ease;

    &:hover {
      opacity: 1;
    }

    &.primary:hover {
      color: $blue-500 !important;
      border-color: rgba($blue-500, 0.3) !important;
      background: rgba($blue-500, 0.05) !important;
    }

    &.success:hover {
      color: $published !important;
      border-color: rgba($published, 0.3) !important;
      background: rgba($published, 0.05) !important;
    }

    &.warning:hover {
      color: #f59e0b !important;
      border-color: rgba(245, 158, 11, 0.3) !important;
      background: rgba(245, 158, 11, 0.05) !important;
    }

    &.danger:hover {
      color: #ef4444 !important;
      border-color: rgba(239, 68, 68, 0.3) !important;
      background: rgba(239, 68, 68, 0.05) !important;
    }
  }
}

// 分页
.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;

  :deep(.el-pagination) {
    padding: 8px 16px;
    border-radius: 10px;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);

    .el-pagination__sizes .el-input__wrapper { border-radius: 6px; }
    button { border-radius: 6px; transition: all 0.2s; &:hover { transform: translateY(-1px); } }
    .el-pager li {
      border-radius: 6px;
      transition: all 0.2s;
      &:hover { transform: translateY(-1px); }
      &.is-active { font-weight: 600; }
    }
  }
}

// 入场动画
@keyframes cardFadeIn {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
