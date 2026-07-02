<template>
  <div class="workflow-container">
    <!-- Hero Header Section -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <div class="title-section">
            <h1 class="page-title">
              <el-icon class="title-icon"><Connection /></el-icon>
              工作流管理
            </h1>
            <p class="page-subtitle">创建和管理您的智能工作流程</p>
          </div>
        </div>
        <div class="header-actions">
          <el-button type="primary" size="large" @click="handleAdd" class="create-btn">
            <el-icon><Plus /></el-icon>
            新建工作流
          </el-button>
        </div>
      </div>
      
      <!-- Stats Cards -->
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-icon blue">
            <el-icon><Grid /></el-icon>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ total }}</span>
            <span class="stat-label">总工作流</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon green">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ publishedCount }}</span>
            <span class="stat-label">已发布</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon purple">
            <el-icon><Edit /></el-icon>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ draftCount }}</span>
            <span class="stat-label">草稿</span>
          </div>
        </div>
        <div class="stat-card" @click="goToTemplateMarket">
          <div class="stat-icon orange">
            <el-icon><Folder /></el-icon>
          </div>
          <div class="stat-info clickable">
            <span class="stat-value">
              模板市场
              <el-icon class="arrow-icon"><ArrowRight /></el-icon>
            </span>
            <span class="stat-label">快速开始</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Search & Filter Bar -->
    <div class="search-bar">
      <div class="search-input-wrapper">
        <el-icon class="search-icon"><Search /></el-icon>
        <el-input 
          v-model="searchForm.name" 
          placeholder="搜索工作流名称..." 
          clearable
          class="search-input"
          @keyup.enter="handleSearch"
        />
      </div>
      <div class="filter-actions">
        <el-button @click="handleReset" class="reset-btn">
          <el-icon><RefreshRight /></el-icon>
          重置
        </el-button>
        <el-button type="primary" @click="handleSearch" class="search-btn">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>

    <!-- Workflow Cards Grid -->
    <div class="workflow-grid" v-loading="loading">
      <!-- Empty State -->
      <div v-if="workflows.length === 0 && !loading" class="empty-state">
        <div class="empty-illustration">
          <div class="empty-circle">
            <el-icon><Connection /></el-icon>
          </div>
          <div class="empty-particles">
            <span></span><span></span><span></span><span></span><span></span>
          </div>
        </div>
        <h3 class="empty-title">暂无工作流</h3>
        <p class="empty-desc">创建您的第一个工作流，开始自动化之旅</p>
        <el-button type="primary" size="large" @click="handleAdd" class="empty-btn">
          <el-icon><Plus /></el-icon>
          创建工作流
        </el-button>
      </div>

      <!-- Workflow Cards -->
      <div
        v-for="(wf, index) in workflows"
        :key="wf.id"
        class="workflow-card"
        :class="{ published: wf.status === 'published' }"
        :style="{ '--delay': `${index * 0.05}s` }"
      >
        <!-- Card Glow Effect -->
        <div class="card-glow"></div>
        
        <!-- Card Header -->
        <div class="card-header-section">
          <div class="card-icon-wrapper" :class="{ published: wf.status === 'published' }">
            <div class="card-icon-inner">
              <el-icon :size="22"><Connection /></el-icon>
            </div>
            <div class="card-icon-ring"></div>
          </div>
          
          <div class="card-title-area">
            <h3 class="workflow-name">{{ wf.name }}</h3>
            <div class="workflow-meta">
              <span class="status-badge" :class="{ published: wf.status === 'published' }">
                <span class="status-dot"></span>
                {{ wf.status === 'published' ? '已发布' : '草稿' }}
              </span>
              <span class="node-count">
                <el-icon><Operation /></el-icon>
                {{ getNodeCount(wf.topology) }} 节点
              </span>
            </div>
          </div>
          
          <!-- More Actions Dropdown -->
          <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, wf)" placement="bottom-end">
            <el-button text class="more-btn">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><EditPen /></el-icon>
                  编辑信息
                </el-dropdown-item>
                <el-dropdown-item command="design">
                  <el-icon><Setting /></el-icon>
                  设计流程
                </el-dropdown-item>
                <el-dropdown-item command="toggle" :divided="true">
                  <el-icon>
                    <Check v-if="wf.status !== 'published'" />
                    <Close v-else />
                  </el-icon>
                  {{ wf.status === 'published' ? '取消发布' : '发布' }}
                </el-dropdown-item>
                <el-dropdown-item command="delete" class="danger-item">
                  <el-icon><Delete /></el-icon>
                  删除
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <!-- Card Body -->
        <div class="card-body">
          <p class="workflow-desc" :class="{ empty: !wf.description }">
            {{ wf.description || '暂无描述' }}
          </p>
        </div>

        <!-- Card Footer -->
        <div class="card-footer-section">
          <div class="footer-info">
            <div class="time-info">
              <el-icon><Clock /></el-icon>
              <span>{{ formatTime(wf.created_at) }}</span>
            </div>
          </div>
          
          <div class="quick-actions">
            <el-tooltip content="设计流程" placement="top">
              <el-button class="action-btn primary" circle @click.stop="handleDesign(wf)">
                <el-icon :size="16"><Setting /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip :content="wf.status === 'published' ? '取消发布' : '发布'" placement="top">
              <el-button 
                class="action-btn" 
                :class="wf.status === 'published' ? 'warning' : 'success'" 
                circle 
                @click.stop="handleToggleStatus(wf)"
              >
                <el-icon :size="16">
                  <component :is="wf.status === 'published' ? 'Close' : 'Check'" />
                </el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </div>
        
        <!-- Card Border Animation -->
        <div class="card-border"></div>
      </div>
    </div>

    <!-- Pagination -->
    <div class="pagination-wrapper" v-if="total > 0">
      <div class="pagination-info">
        共 <span class="highlight">{{ total }}</span> 条记录
      </div>
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[8, 12, 24, 48]"
        :total="total"
        layout="sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        background
      />
    </div>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新建工作流' : '编辑工作流'"
      width="480px"
      class="workflow-dialog"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="workflow-form"
      >
        <el-form-item label="工作流名称" prop="name">
          <el-input 
            v-model="form.name" 
            placeholder="请输入工作流名称" 
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="工作流描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入工作流描述（可选）"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false" size="large">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading" size="large">
            {{ dialogType === 'add' ? '创建' : '保存' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder, Search, RefreshRight, Plus,
  Connection, Clock, EditPen, Setting, Delete,
  Close, Check, ArrowRight, Grid, CircleCheck,
  Edit, MoreFilled, Operation
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

// 统计数据
const publishedCount = computed(() => 
  workflows.value.filter(w => w.status === 'published').length
)
const draftCount = computed(() => 
  workflows.value.filter(w => w.status === 'draft').length
)

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

// 处理下拉菜单命令
const handleCommand = (command, row) => {
  switch (command) {
    case 'edit':
      handleEdit(row)
      break
    case 'design':
      handleDesign(row)
      break
    case 'toggle':
      handleToggleStatus(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

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
$primary: #2B5EFF;
$primary-light: rgba(#2B5EFF, 0.1);
$primary-glow: rgba(#2B5EFF, 0.3);

$success: #10b981;
$success-light: rgba(#10b981, 0.1);

$purple: #8b5cf6;
$purple-light: rgba(#8b5cf6, 0.1);

$orange: #f59e0b;
$orange-light: rgba(#f59e0b, 0.1);

$draft: #94a3b8;
$danger: #ef4444;

.workflow-container {
  padding: 0;
  max-width: 1400px;
  margin: 0 auto;
}

// ==================== Page Header ====================
.page-header {
  margin-bottom: 32px;
  
  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 28px;
  }
  
  .title-section {
    .page-title {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 28px;
      font-weight: 700;
      color: var(--el-text-color-primary);
      margin: 0 0 8px;
      
      .title-icon {
        font-size: 32px;
        color: $primary;
        background: $primary-light;
        padding: 8px;
        border-radius: 12px;
      }
    }
    
    .page-subtitle {
      font-size: 14px;
      color: var(--el-text-color-secondary);
      margin: 0;
    }
  }
  
  .header-actions {
    .create-btn {
      padding: 14px 28px;
      font-size: 15px;
      font-weight: 500;
      border-radius: 12px;
      background: linear-gradient(135deg, $primary, lighten($primary, 10%));
      border: none;
      box-shadow: 0 4px 16px $primary-glow;
      transition: all 0.3s ease;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px $primary-glow;
      }
      
      .el-icon {
        margin-right: 8px;
      }
    }
  }
}

// ==================== Stats Cards ====================
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 28px;
  
  .stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px 24px;
    background: var(--el-bg-color);
    border-radius: 16px;
    border: 1px solid var(--el-border-color-lighter);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
    }
    
    &.clickable {
      cursor: pointer;
      &:hover { border-color: $primary; }
    }
    
    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 22px;
      
      &.blue {
        background: $primary-light;
        color: $primary;
      }
      &.green {
        background: $success-light;
        color: $success;
      }
      &.purple {
        background: $purple-light;
        color: $purple;
      }
      &.orange {
        background: $orange-light;
        color: $orange;
      }
    }
    
    .stat-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
      
      .stat-value {
        font-size: 22px;
        font-weight: 700;
        color: var(--el-text-color-primary);
        display: flex;
        align-items: center;
        gap: 6px;
        
        .arrow-icon {
          font-size: 14px;
          color: $primary;
          transition: transform 0.3s;
        }
      }
      
      &.clickable:hover .arrow-icon {
        transform: translateX(4px);
      }
      
      .stat-label {
        font-size: 13px;
        color: var(--el-text-color-secondary);
      }
    }
  }
}

// ==================== Search Bar ====================
.search-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px 20px;
  background: var(--el-bg-color);
  border-radius: 16px;
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  
  .search-input-wrapper {
    flex: 1;
    position: relative;
    
    .search-icon {
      position: absolute;
      left: 16px;
      top: 50%;
      transform: translateY(-50%);
      font-size: 18px;
      color: var(--el-text-color-placeholder);
      z-index: 1;
    }
    
    .search-input {
      :deep(.el-input__wrapper) {
        padding-left: 44px;
        border-radius: 12px;
        height: 44px;
        box-shadow: none !important;
        border: 1px solid transparent;
        transition: all 0.3s;
        
        &:hover, &:focus-within {
          border-color: $primary;
          box-shadow: 0 0 0 3px $primary-light !important;
        }
        
        .el-input__inner {
          font-size: 14px;
        }
      }
    }
  }
  
  .filter-actions {
    display: flex;
    gap: 12px;
    
    .reset-btn, .search-btn {
      padding: 12px 20px;
      border-radius: 12px;
      font-weight: 500;
      transition: all 0.3s;
    }
    
    .reset-btn {
      &:hover {
        transform: translateY(-1px);
      }
    }
    
    .search-btn {
      background: $primary;
      border: none;
      &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 12px $primary-glow;
      }
      .el-icon { margin-right: 6px; }
    }
  }
}

// ==================== Workflow Grid ====================
.workflow-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
  min-height: 300px;
}

// ==================== Empty State ====================
.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  
  .empty-illustration {
    position: relative;
    margin-bottom: 24px;
    
    .empty-circle {
      width: 120px;
      height: 120px;
      border-radius: 50%;
      background: linear-gradient(135deg, $primary-light, $purple-light);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 48px;
      color: $primary;
      animation: float 3s ease-in-out infinite;
    }
    
    .empty-particles {
      position: absolute;
      inset: 0;
      
      span {
        position: absolute;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: $primary;
        opacity: 0.3;
        animation: particle 2s ease-in-out infinite;
        
        &:nth-child(1) { top: 10%; left: 0; animation-delay: 0s; }
        &:nth-child(2) { top: 20%; right: 0; animation-delay: 0.3s; }
        &:nth-child(3) { bottom: 10%; left: 10%; animation-delay: 0.6s; }
        &:nth-child(4) { bottom: 20%; right: 10%; animation-delay: 0.9s; }
        &:nth-child(5) { top: 50%; left: -10%; animation-delay: 1.2s; }
      }
    }
  }
  
  .empty-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin: 0 0 8px;
  }
  
  .empty-desc {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    margin: 0 0 24px;
  }
  
  .empty-btn {
    padding: 14px 32px;
    font-size: 15px;
    border-radius: 12px;
    background: linear-gradient(135deg, $primary, lighten($primary, 10%));
    border: none;
    box-shadow: 0 4px 16px $primary-glow;
    transition: all 0.3s;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 24px $primary-glow;
    }
    
    .el-icon { margin-right: 8px; }
  }
}

// ==================== Workflow Card ====================
.workflow-card {
  position: relative;
  background: var(--el-bg-color);
  border-radius: 20px;
  border: 1px solid var(--el-border-color-lighter);
  overflow: hidden;
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  animation: cardFadeIn 0.5s ease-out var(--delay, 0s) both;
  
  &:hover {
    transform: translateY(-6px);
    border-color: transparent;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
    
    .card-glow { opacity: 1; }
    .card-icon-wrapper .card-icon-ring { transform: scale(1.1); }
    .card-border { opacity: 1; }
  }
  
  // 卡片发光效果
  .card-glow {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 120px;
    background: linear-gradient(180deg, $primary-light 0%, transparent 100%);
    opacity: 0;
    transition: opacity 0.3s;
    pointer-events: none;
  }
  
  // 卡片边框动画
  .card-border {
    position: absolute;
    inset: 0;
    border-radius: 20px;
    padding: 1px;
    background: linear-gradient(135deg, $primary, $purple, $primary);
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    opacity: 0;
    transition: opacity 0.3s;
    pointer-events: none;
  }
  
  // 卡片头部区域
  .card-header-section {
    position: relative;
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 24px 20px 0;
  }
  
  // 图标容器
  .card-icon-wrapper {
    position: relative;
    flex-shrink: 0;
    
    .card-icon-inner {
      width: 52px;
      height: 52px;
      border-radius: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, $draft, darken($draft, 10%));
      color: #fff;
      transition: all 0.3s;
    }
    
    .card-icon-ring {
      position: absolute;
      inset: -3px;
      border-radius: 19px;
      border: 2px solid $draft;
      opacity: 0.3;
      transition: transform 0.3s;
    }
    
    &.published {
      .card-icon-inner {
        background: linear-gradient(135deg, $primary, darken($primary, 5%));
        box-shadow: 0 4px 16px $primary-glow;
      }
      .card-icon-ring {
        border-color: $primary;
        opacity: 0.5;
      }
    }
  }
  
  // 标题区域
  .card-title-area {
    flex: 1;
    min-width: 0;
    padding-top: 4px;
    
    .workflow-name {
      margin: 0 0 8px;
      font-size: 17px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      line-height: 1.4;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    
    .workflow-meta {
      display: flex;
      align-items: center;
      gap: 12px;
      
      .status-badge {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        padding: 4px 10px;
        border-radius: 20px;
        font-size: 12px;
        font-weight: 500;
        background: rgba($draft, 0.12);
        color: $draft;
        
        .status-dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          background: $draft;
        }
        
        &.published {
          background: $primary-light;
          color: $primary;
          .status-dot {
            background: $primary;
            animation: pulse 2s ease-in-out infinite;
          }
        }
      }
      
      .node-count {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--el-text-color-placeholder);
        
        .el-icon { font-size: 13px; }
      }
    }
  }
  
  // 更多操作按钮
  .more-btn {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    color: var(--el-text-color-secondary);
    transition: all 0.2s;
    
    &:hover {
      background: var(--el-fill-color-light);
      color: var(--el-text-color-primary);
    }
  }
  
  // 卡片内容
  .card-body {
    padding: 16px 20px;
    
    .workflow-desc {
      margin: 0;
      font-size: 13px;
      color: var(--el-text-color-secondary);
      line-height: 1.6;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      
      &.empty {
        color: var(--el-text-color-placeholder);
        font-style: italic;
      }
    }
  }
  
  // 卡片底部
  .card-footer-section {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px 20px;
    border-top: 1px solid var(--el-border-color-extra-light);
    
    .footer-info {
      .time-info {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: var(--el-text-color-placeholder);
        
        .el-icon { font-size: 14px; }
      }
    }
    
    .quick-actions {
      display: flex;
      gap: 8px;
    }
  }
  
  // 操作按钮
  .action-btn {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    border: 1px solid var(--el-border-color-lighter) !important;
    background: transparent !important;
    color: var(--el-text-color-secondary) !important;
    transition: all 0.25s ease;
    
    &:hover {
      transform: scale(1.05);
    }
    
    &.primary:hover {
      color: $primary !important;
      border-color: rgba($primary, 0.3) !important;
      background: $primary-light !important;
    }
    
    &.success:hover {
      color: $success !important;
      border-color: rgba($success, 0.3) !important;
      background: $success-light !important;
    }
    
    &.warning:hover {
      color: $orange !important;
      border-color: rgba($orange, 0.3) !important;
      background: $orange-light !important;
    }
  }
}

// Dropdown danger item
:deep(.el-dropdown-menu__item.danger-item) {
  color: $danger;
  &:hover {
    background: rgba($danger, 0.08);
    color: $danger;
  }
}

// ==================== Pagination ====================
.pagination-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 32px;
  padding: 16px 20px;
  background: var(--el-bg-color);
  border-radius: 16px;
  border: 1px solid var(--el-border-color-lighter);
  
  .pagination-info {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    
    .highlight {
      color: $primary;
      font-weight: 600;
    }
  }
  
  :deep(.el-pagination) {
    .el-pagination__sizes .el-input__wrapper { border-radius: 10px; }
    button, .el-pager li {
      border-radius: 10px;
      transition: all 0.2s;
      &:hover { transform: translateY(-1px); }
    }
    .el-pager li.is-active {
      background: $primary;
      color: #fff;
    }
  }
}

// ==================== Dialog ====================
.workflow-dialog {
  :deep(.el-dialog) {
    border-radius: 20px;
    overflow: hidden;
  }
  
  :deep(.el-dialog__header) {
    padding: 24px 28px 0;
    .el-dialog__title {
      font-size: 20px;
      font-weight: 600;
    }
  }
  
  :deep(.el-dialog__body) {
    padding: 24px 28px;
  }
  
  .workflow-form {
    :deep(.el-form-item__label) {
      font-weight: 500;
      padding-bottom: 8px;
    }
    
    :deep(.el-input__wrapper),
    :deep(.el-textarea__inner) {
      border-radius: 12px;
      padding: 12px 16px;
    }
  }
  
  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding-top: 8px;
    
    .el-button {
      padding: 12px 28px;
      border-radius: 12px;
    }
  }
}

// ==================== Animations ====================
@keyframes cardFadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

@keyframes particle {
  0%, 100% { transform: translateY(0) scale(1); opacity: 0.3; }
  50% { transform: translateY(-10px) scale(1.5); opacity: 0.6; }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

// ==================== Responsive ====================
@media (max-width: 1200px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-row {
    grid-template-columns: 1fr;
  }
  
  .search-bar {
    flex-direction: column;
    align-items: stretch;
    
    .filter-actions {
      justify-content: flex-end;
    }
  }
  
  .workflow-grid {
    grid-template-columns: 1fr;
  }
  
  .pagination-wrapper {
    flex-direction: column;
    gap: 16px;
  }
}
</style>
