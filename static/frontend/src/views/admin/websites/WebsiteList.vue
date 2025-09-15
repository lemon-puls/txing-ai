<template>
  <div class="website-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="网站名称">
          <el-input v-model="searchForm.name" placeholder="请输入网站名称" clearable />
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="searchForm.tag" placeholder="请选择标签" clearable>
            <el-option v-for="tag in allTags" :key="tag" :value="tag" :label="tag" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增网站</el-button>
        </el-form-item>
      </el-form>
    </el-card>


    <!-- 网站列表 -->
    <el-card class="table-card">
      <el-table
        :data="websites"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />

        <el-table-column label="网站信息" min-width="300">
          <template #default="{ row }">
            <div class="website-info">
              <div class="website-avatar">
                <img
                  v-if="row.avatar"
                  :src="row.avatar"
                  :alt="row.name"
                  @error="handleImageError"
                />
                <el-icon v-else class="default-icon"><Link /></el-icon>
              </div>
              <div class="website-details">
                <div class="website-name">{{ row.name }}</div>
                <div class="website-url">{{ row.url }}</div>
                <div class="website-description">{{ row.description }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="标签" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="tag in row.tags"
              :key="tag"
              size="small"
              type="info"
              effect="plain"
              style="margin-right: 8px; margin-bottom: 4px;"
            >
              {{ tag }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增网站' : '编辑网站'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="websiteFormRef"
        :model="websiteForm"
        :rules="websiteRules"
        label-width="100px"
      >
        <el-form-item label="网站名称" prop="name">
          <el-input v-model="websiteForm.name" placeholder="请输入网站名称" />
        </el-form-item>

        <el-form-item label="网站地址" prop="url">
          <el-input v-model="websiteForm.url" placeholder="请输入网站地址" />
        </el-form-item>

        <el-form-item label="网站描述" prop="description">
          <el-input
            v-model="websiteForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入网站描述"
          />
        </el-form-item>

        <el-form-item label="网站头像" prop="avatar">
          <div class="avatar-upload">
            <el-input
              v-model="websiteForm.avatar"
              placeholder="请输入头像URL或点击自动获取"
            />
            <div class="avatar-actions">
              <el-button @click="autoGetAvatar" :loading="avatarLoading">
                自动获取
              </el-button>
              <el-upload
                class="avatar-uploader"
                :show-file-list="false"
                :before-upload="beforeAvatarUpload"
                :http-request="uploadAvatar"
              >
                <el-button>上传头像</el-button>
              </el-upload>
            </div>
          </div>
          <div v-if="websiteForm.avatar" class="avatar-preview">
            <img :src="websiteForm.avatar" alt="头像预览" />
          </div>
        </el-form-item>

        <el-form-item label="标签" prop="tags">
          <div class="tags-input">
            <el-tag
              v-for="tag in websiteForm.tags"
              :key="tag"
              closable
              @close="removeTag(tag)"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ tag }}
            </el-tag>
            <el-input
              v-if="inputVisible"
              ref="inputRef"
              v-model="inputValue"
              size="small"
              style="width: 100px;"
              @keyup.enter="handleInputConfirm"
              @blur="handleInputConfirm"
            />
            <el-button v-else size="small" @click="showInput">
              + 添加标签
            </el-button>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="WebsiteList">
import { ref, onMounted, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Link } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const avatarLoading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 搜索表单
const searchForm = ref({
  name: '',
  tag: ''
})

// 网站数据
const websites = ref([])

// 对话框数据
const dialogVisible = ref(false)
const dialogType = ref('add')
const websiteFormRef = ref(null)
const websiteForm = ref({
  id: null,
  name: '',
  url: '',
  description: '',
  avatar: '',
  tags: []
})

// 标签输入相关
const inputVisible = ref(false)
const inputValue = ref('')
const inputRef = ref(null)


// 计算所有标签
const allTags = computed(() => {
  const tags = new Set()
  websites.value.forEach(website => {
    // 把 website.tags 按 , 分割，放入 tags 集合
    website.tags.split(',').forEach(tag => tags.add(tag))
  })
  return Array.from(tags)
})

// 表单验证规则
const websiteRules = {
  name: [
    { required: true, message: '请输入网站名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入网站地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL地址', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入网站描述', trigger: 'blur' },
    { max: 200, message: '描述不能超过200个字符', trigger: 'blur' }
  ],
  tags: [
    { required: true, message: '请至少添加一个标签', trigger: 'change' },
    { type: 'array', min: 1, message: '请至少添加一个标签', trigger: 'change' }
  ]
}

// 加载网站列表
const loadWebsites = async () => {
  loading.value = true
  try {
    const response = await defaultApi.apiWebsitesListGet(
      currentPage.value,
      pageSize.value,
      {
        name: searchForm.value.name || undefined,
        tag: searchForm.value.tag || undefined
      }
    )

    if (response.code === 0) {
      websites.value = response.data.records || []
      total.value = response.data.total || 0
    } else {
      ElMessage.error(response.message || '加载失败')
    }
  } catch (error) {
    console.error('加载网站列表失败:', error)
    ElMessage.error('加载网站列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadWebsites()
}

// 重置搜索
const handleReset = () => {
  searchForm.value = {
    name: '',
    tag: ''
  }
  currentPage.value = 1
  loadWebsites()
}

// 新增网站
const handleAdd = () => {
  dialogType.value = 'add'
  websiteForm.value = {
    id: null,
    name: '',
    url: '',
    description: '',
    avatar: '',
    tags: []
  }
  dialogVisible.value = true
}

// 编辑网站
const handleEdit = async (row) => {
  try {
    const response = await defaultApi.apiAdminWebsitesIdGet(row.id)
    if (response.code == 0) {
      dialogType.value = 'edit'
      websiteForm.value = {
        id: response.data.id,
        name: response.data.name,
        url: response.data.url,
        description: response.data.description,
        avatar: response.data.avatar,
        tags: response.data.tags ? response.data.tags.split(',') : []
      }
      dialogVisible.value = true
    } else {
      ElMessage.error(response.message || '获取网站信息失败')
    }
  } catch (error) {
    console.error('获取网站信息失败:', error)
    ElMessage.error('获取网站信息失败')
  }
}

// 删除网站
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除网站 "${row.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await defaultApi.apiAdminWebsitesIdDelete(row.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      await loadWebsites()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!websiteFormRef.value) return

  await websiteFormRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        const formData = {
          ...websiteForm.value,
          tags: websiteForm.value.tags.join(',')
        }

        let response
        if (dialogType.value === 'add') {
          response = await defaultApi.apiAdminWebsitesPost(formData)
        } else {
          response = await defaultApi.apiAdminWebsitesIdPut(formData.id, formData)
        }

        if (response.code === 0) {
          ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
          dialogVisible.value = false
          await loadWebsites()
        } else {
          ElMessage.error(response.message || '操作失败')
        }
      } catch (error) {
        console.error('提交失败:', error)
        ElMessage.error('操作失败')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

// 自动获取头像
const autoGetAvatar = async () => {
  if (!websiteForm.value.url) {
    ElMessage.warning('请先输入网站地址')
    return
  }

  avatarLoading.value = true
  try {
    const response = await defaultApi.apiAdminWebsitesFaviconPost({
      url: websiteForm.value.url
    })

    if (response.code === 0) {
      websiteForm.value.avatar = response.data.favicon
      ElMessage.success('头像获取成功')
    } else {
      ElMessage.error(response.message || '头像获取失败')
    }
  } catch (error) {
    console.error('获取头像失败:', error)
    ElMessage.error('获取头像失败')
  } finally {
    avatarLoading.value = false
  }
}

// 上传头像前的检查
const beforeAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

// 自定义上传
const uploadAvatar = async (options) => {
  try {
    // TODO: 实现文件上传逻辑
    // const formData = new FormData()
    // formData.append('file', options.file)
    // const response = await uploadApi.uploadFile(formData)
    // websiteForm.value.avatar = response.data.url

    // 模拟上传成功
    const reader = new FileReader()
    reader.onload = (e) => {
      websiteForm.value.avatar = e.target.result
    }
    reader.readAsDataURL(options.file)

    ElMessage.success('上传成功')
  } catch (error) {
    console.error('上传失败:', error)
    ElMessage.error('上传失败')
  }
}

// 标签相关方法
const removeTag = (tag) => {
  const index = websiteForm.value.tags.indexOf(tag)
  if (index > -1) {
    websiteForm.value.tags.splice(index, 1)
  }
}

const showInput = () => {
  inputVisible.value = true
  nextTick(() => {
    inputRef.value?.focus()
  })
}

const handleInputConfirm = () => {
  if (inputValue.value && !websiteForm.value.tags.includes(inputValue.value)) {
    websiteForm.value.tags.push(inputValue.value)
  }
  inputVisible.value = false
  inputValue.value = ''
}

// 工具方法
const formatDate = (dateString) => {
  console.log(dateString)
  return dateString || '-'
}

const handleImageError = (event) => {
  event.target.style.display = 'none'
}

// 分页相关
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadWebsites()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  loadWebsites()
}

// 初始化
onMounted(() => {
  loadWebsites()
})
</script>

<style scoped lang="scss">
.website-container {
  padding: 20px;
}

.search-form {
  margin-bottom: 20px;

  .el-form {
    margin-bottom: 0;
  }
}

.table-card {
  .website-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .website-avatar {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      overflow: hidden;
      background: #f5f5f5;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .default-icon {
        font-size: 20px;
        color: #999;
      }
    }

    .website-details {
      flex: 1;
      min-width: 0;

      .website-name {
        font-weight: 600;
        color: #303133;
        margin-bottom: 4px;
      }

      .website-url {
        font-size: 12px;
        color: #409eff;
        margin-bottom: 4px;
        word-break: break-all;
      }

      .website-description {
        font-size: 12px;
        color: #909399;
        line-height: 1.4;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }
    }
  }
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

// 对话框样式
.avatar-upload {
  display: flex;
  gap: 8px;
  align-items: center;

  .el-input {
    flex: 1;
  }

  .avatar-actions {
    display: flex;
    gap: 8px;
  }
}

.avatar-preview {
  margin-top: 10px;

  img {
    width: 60px;
    height: 60px;
    border-radius: 8px;
    object-fit: cover;
    border: 1px solid #dcdfe6;
  }
}

.tags-input {
  .el-tag {
    margin-right: 8px;
    margin-bottom: 8px;
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  .website-container {
    padding: 10px;
  }

  .search-form {
    :deep(.el-form--inline) {
      .el-form-item {
        display: block;
        margin-right: 0;
        margin-bottom: 10px;
      }
    }
  }

  .table-card {
    :deep(.el-table) {
      .el-table__cell {
        padding: 8px 0;
      }
    }
  }
}
</style>
