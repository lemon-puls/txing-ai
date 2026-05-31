<template>
  <div class="workflow-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="工作流名称">
          <el-input v-model="searchForm.name" placeholder="请输入工作流名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增工作流</el-button>
          <el-button type="warning" @click="goToTemplateMarket">
            <el-icon><Folder /></el-icon>
            模板市场
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工作流列表 -->
    <el-card class="workflow-list">
      <el-table :data="workflows" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="工作流名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'published' ? 'success' : 'info'" size="small">
              {{ scope.row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)">编辑信息</el-button>
            <el-button type="success" link @click="handleDesign(scope.row)">设计流程</el-button>
            <el-button
              :type="scope.row.status === 'published' ? 'warning' : 'success'"
              link
              @click="handleToggleStatus(scope.row)"
            >
              {{ scope.row.status === 'published' ? '取消发布' : '发布' }}
            </el-button>
            <el-button type="danger" link @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 30, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

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
import { Folder } from '@element-plus/icons-vue'
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
const pageSize = ref(10)
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
    const params = {
      page: currentPage.value,
      limit: pageSize.value,
      name: searchForm.value.name || undefined
    }
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
          data.topology = '' // 默认空拓扑
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
.workflow-container {
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

  .workflow-list {
    border-radius: 16px;
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 24px;
  }
}
</style>
