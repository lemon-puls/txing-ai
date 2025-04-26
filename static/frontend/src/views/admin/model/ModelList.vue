<template>
  <div class="model-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="模型名称">
          <el-input v-model="searchForm.name" placeholder="请输入模型名称" clearable />
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="searchForm.tag" placeholder="请选择标签" clearable>
            <el-option v-for="tag in modelTags" :key="tag" :value="tag" :label="tag" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增模型</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 模型列表 -->
    <div class="model-grid">
      <el-card v-for="model in models" :key="model.id" class="model-card">
        <template #header>
          <div class="card-header">
            <div class="model-info">
              <el-avatar :size="40" :src="model.avatar">{{ model.name.charAt(0) }}</el-avatar>
              <div class="model-title">
                <h3>{{ model.name }}</h3>
                <div class="model-tags">
                  <el-tag v-if="model.default" type="success" effect="dark" size="small">默认</el-tag>
                  <el-tag v-if="model.high_context" type="warning" effect="dark" size="small">高上下文</el-tag>
                  <el-tag
                    v-for="tag in model.tag?.split(',')"
                    :key="tag"
                    type="info"
                    effect="plain"
                    size="small"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="card-actions">
              <el-button text type="primary" @click="toggleExpand(model)">
                {{ model.isExpanded ? '收起' : '展开' }}
                <el-icon>
                  <ArrowDown v-if="!model.isExpanded" />
                  <ArrowUp v-else />
                </el-icon>
              </el-button>
            </div>
          </div>
        </template>

        <div class="model-content" :class="{ expanded: model.isExpanded }">
          <div class="model-description">{{ model.description || '暂无描述' }}</div>
          <div v-show="model.isExpanded" class="model-actions">
            <el-button-group>
              <el-button type="primary" @click="handleEdit(model)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                type="danger"
                @click="handleDelete(model)"
                :disabled="model.default"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </el-button-group>
            <el-switch
              v-model="model.default"
              active-text="设为默认"
              @change="handleSetDefault(model)"
            />
          </div>
        </div>
      </el-card>
    </div>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[8, 16, 24, 32]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 编辑/新增对话框 -->
    <el-dialog
      :title="dialogType === 'add' ? '新增模型' : '编辑模型'"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form
        ref="modelFormRef"
        :model="modelForm"
        :rules="modelRules"
        label-width="100px"
      >
        <el-form-item label="模型头像" prop="avatar" class="avatar-uploader">
          <ImageUploader
            v-model="modelForm.avatar"
            :circle="true"
            :fixed="true"
            :enable-compress="true"
            :compress-quality="0.5"
            :max-width="1080"
            :max-height="1080"
            :crop-width="200"
            :crop-height="200"
            :enable-crop="true"
            placeholder="点击上传头像"
            @success="(url) => {
              console.log('Upload success:', url)
            }"
            @error="(error) => ElMessage.error(error.message || '上传失败')"
          />
        </el-form-item>
        <el-form-item label="模型名称" prop="name">
          <el-input v-model="modelForm.name" placeholder="请输入模型名称" />
        </el-form-item>
        <el-form-item label="模型描述" prop="description">
          <el-input
            v-model="modelForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入模型描述"
          />
        </el-form-item>
        <el-form-item label="模型标签" prop="tag">
          <el-select
            v-model="modelForm.tags"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="请选择或输入标签"
          >
            <el-option
              v-for="tag in modelTags"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="高上下文" prop="high_context">
          <el-switch v-model="modelForm.high_context" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, ArrowUp, Edit, Delete } from '@element-plus/icons-vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import { defaultApi } from '@/api/index.js'

// 搜索表单
const searchForm = ref({
  name: '',
  tag: ''
})

// 分页数据
const currentPage = ref(1)
const pageSize = ref(8)
const total = ref(0)

// 模型数据
const models = ref([])
const modelTags = ref(['通用', '对话', '编程', '创意', '分析'])

// 对话框数据
const dialogVisible = ref(false)
const dialogType = ref('add')
const modelForm = ref({
  name: '',
  description: '',
  avatar: '',
  tags: [],
  high_context: false
})
const modelFormRef = ref(null)

// 表单验证规则
const modelRules = {
  name: [
    { required: true, message: '请输入模型名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 500, message: '描述不能超过500个字符', trigger: 'blur' }
  ]
}

// 加载模型列表
const loadModels = async () => {
  try {
    const response = await defaultApi.apiModelListGet(
      currentPage.value,
      pageSize.value,
      {
        orderBy: 'id',
        order: 'desc',
        tag: searchForm.value.tag || undefined,
        name: searchForm.value.name || undefined
      }
    )
    if (response.code === 0 && response.data) {
      models.value = response.data.records.map(model => ({
        ...model,
        isExpanded: false
      }))
      total.value = response.data.total || 0
      currentPage.value = response.data.page || 1
      pageSize.value = response.data.limit || 8
    } else {
      ElMessage.error(response.msg || '获取模型列表失败')
    }
  } catch (error) {
    console.error('Load models error:', error)
    ElMessage.error(error.body?.msg || '获取模型列表失败')
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadModels()
}

// 重置搜索
const handleReset = () => {
  searchForm.value = {
    name: '',
    tag: ''
  }
  currentPage.value = 1
  loadModels()
}

// 切换展开/收起
const toggleExpand = (model) => {
  model.isExpanded = !model.isExpanded
}

// 新增模型
const handleAdd = () => {
  dialogType.value = 'add'
  modelForm.value = {
    name: '',
    description: '',
    avatar: '',
    tags: [],
    high_context: false
  }
  dialogVisible.value = true
}

// 编辑模型
const handleEdit = (model) => {
  dialogType.value = 'edit'
  modelForm.value = {
    ...model,
    tags: model.tag ? model.tag.split(',') : []
  }
  dialogVisible.value = true
}

// 删除模型
const handleDelete = async (model) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模型【${model.name}】吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await defaultApi.apiAdminModelIdDelete(model.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      await loadModels()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.body?.msg || '删除失败')
    }
  }
}

// 设置默认模型
const handleSetDefault = async (model) => {
  try {
    const response = await defaultApi.apiAdminModelIdPut(model.id, {
      default: model.default
    })
    if (response.code === 0) {
      ElMessage.success(`已${model.default ? '设置' : '取消'}为默认模型`)
      await loadModels()
    } else {
      model.default = !model.default // 恢复状态
      ElMessage.error(response.msg || '操作失败')
    }
  } catch (error) {
    model.default = !model.default // 恢复状态
    ElMessage.error(error.body?.msg || '操作失败')
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!modelFormRef.value) return
  await modelFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // 处理头像 URL
        let avatar = modelForm.value.avatar
        if (avatar) {
          const urlObj = new URL(avatar)
          avatar = `${urlObj.origin}${urlObj.pathname}`
        }

        // 构建请求数据
        const formData = {
          name: modelForm.value.name,
          description: modelForm.value.description,
          avatar: avatar,
          tag: modelForm.value.tags.join(','),
          high_context: modelForm.value.high_context
        }

        let response
        if (dialogType.value === 'add') {
          response = await defaultApi.apiAdminModelPost(formData)
        } else {
          response = await defaultApi.apiAdminModelIdPut(modelForm.value.id, formData)
        }

        if (response.code === 0) {
          ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
          dialogVisible.value = false
          await loadModels()
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('Submit error:', error)
        ElMessage.error(error.body?.msg || '操作失败')
      }
    }
  })
}

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadModels()
}

// 页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  loadModels()
}

// 页面加载时获取数据
onMounted(() => {
  loadModels()
})
</script>

<style lang="scss" scoped>
.model-container {
  padding: 24px;

  :deep(.el-card) {
    border-radius: 16px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    overflow: hidden;

    &:hover {
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
      transform: translateY(-2px);
    }
  }
}

.model-card .el-button-group .el-button {
   margin-right: 10px;
}

.search-form {
  margin-bottom: 24px;

  :deep(.el-form-item) {
    margin-bottom: 0;
  }

  :deep(.el-input__wrapper) {
    border-radius: 12px;
  }

  :deep(.el-select) {
    width: 180px;
  }

  :deep(.el-button) {
    border-radius: 12px;
    padding: 12px 24px;
    transition: all 0.3s;

    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    &.el-button--primary {
      &:hover {
        box-shadow: 0 4px 12px rgba(var(--el-color-primary-rgb), 0.3);
      }
    }
  }
}

.model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.model-card {
  min-height: 200px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 20px;
  }

  .model-info {
    display: flex;
    gap: 20px;
    align-items: flex-start;
  }

  .model-title {
    h3 {
      margin: 0 0 12px;
      font-size: 18px;
      line-height: 1.4;
    }
  }

  .model-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .model-content {
    padding: 0 20px 20px;
  }

  .model-description {
    color: var(--el-text-color-secondary);
    font-size: 14px;
    line-height: 1.6;
    margin-bottom: 20px;
    min-height: 44px;
  }

  .model-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  :deep(.el-tag) {
    border-radius: 6px;
  }

  :deep(.el-button-group) {
    .el-button {
      border-radius: 8px;
    }
  }
}

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

.avatar-uploader {
  :deep(.el-upload) {
    width: 100%;
    text-align: center;
  }
}
</style>
