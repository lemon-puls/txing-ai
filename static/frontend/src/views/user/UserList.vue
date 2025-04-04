<!-- 用户列表页面 -->
<template>
  <div class="user-list">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="用户名">
          <el-input v-model="searchForm.userName" placeholder="请输入用户名" clearable/>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option :value="0" label="正常" />
            <el-option :value="1" label="禁用" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card>
      <el-table
        :data="users"
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          label="ID"
        />
        <el-table-column
          prop="userId"
          label="用户ID"
        />
        <el-table-column
          prop="userName"
          label="用户名"
        />
        <el-table-column
          prop="phone"
          label="手机号码"
        />
        <el-table-column
          prop="email"
          label="邮箱"
        />
        <el-table-column
          prop="status"
          label="状态"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'danger'">
              {{ row.status === 0 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="gender"
          label="性别"
        >
          <template #default="{ row }">
            <el-tag :type="row.gender === 0 ? 'success' : 'danger'">
              {{ row.gender === 0 ? '男' : '女' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="200"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              :type="row.status === 0 ? 'danger' : 'success'"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 0 ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'

// 搜索表单数据
const searchForm = ref({
  userName: '',
  status: undefined
})

// 数据状态
const users = ref([])
const loading = ref(false)
// 分页数据
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await defaultApi.apiUserListGet(currentPage.value, pageSize.value, {
      ...searchForm.value
    })
    if (response.code === 0 && response.data) {
      users.value = response.data.records || []
      total.value = response.data.total || 0
      currentPage.value = response.data.page || 1
      pageSize.value = response.data.limit || 10
    } else {
      ElMessage.error(response.msg || '获取用户列表失败')
    }
  } catch (error) {
    ElMessage.error(error.body.msg || '获取用户列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadUsers()
}

// 重置搜索
const handleReset = () => {
  searchForm.value = {
    userId: '',
    status: undefined
  }
  currentPage.value = 1
  loadUsers()
}

// 分页变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  loadUsers()
}

// 每页数量变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadUsers()
}


// 页面加载时获取数据
onMounted(async () => {
  await loadUsers()
})
</script>

<style lang="scss" scoped>
.user-list {
  padding: 24px;
}

.search-form {
  margin-bottom: 20px;

  :deep(.el-select) {
    width: 180px;
  }
}

.role-group {
  margin-bottom: 8px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;

  .service-tag {
    margin-right: 8px;
  }

  .role-tag {
    margin-right: 4px;
  }
}

.no-roles {
  color: #909399;
  font-size: 13px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.role-tree-container {
  max-height: 500px;
  overflow-y: auto;
}

.service-role-tree {
  margin-bottom: 16px;

  .service-title {
    margin-bottom: 8px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
