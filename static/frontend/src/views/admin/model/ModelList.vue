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
          <el-upload
            class="avatar-upload"
            :show-file-list="false"
            :auto-upload="false"
            :on-change="handleAvatarChange"
            accept="image/*"
          >
            <div class="avatar-wrapper">
              <img v-if="modelForm.avatar" :src="modelForm.avatar" class="avatar-image">
              <el-icon v-else class="avatar-icon"><Plus /></el-icon>
            </div>
          </el-upload>
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

    <!-- 图片裁剪对话框 -->
    <el-dialog
      v-model="cropperVisible"
      title="裁剪头像"
      width="600px"
      append-to-body
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="cropper-container">
        <VueCropper
          ref="cropperRef"
          :img="cropperImage"
          :outputSize="1"
          outputType="png"
          :info="true"
          :full="true"
          :canMove="true"
          :canMoveBox="true"
          :original="false"
          :autoCrop="true"
          :autoCropWidth="200"
          :autoCropHeight="200"
          :fixedBox="true"
          :fixed="true"
          :fixedNumber="[1, 1]"
          :centerBox="true"
        />
      </div>
      <template #footer>
        <el-button @click="cropperVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCropImage">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, ArrowUp, Edit, Delete, Plus } from '@element-plus/icons-vue'
import VueCropper from 'vue-cropper/lib/vue-cropper.vue'
import 'vue-cropper/dist/index.css'
import {defaultApi} from "@/api/index.js";

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

// 裁剪相关
const cropperVisible = ref(false)
const cropperImage = ref('')
const cropperRef = ref(null)

// 加载模型列表
const loadModels = async () => {
  // TODO: 调用API获取模型列表
  // 模拟数据
  models.value = [
    {
      id: 1,
      name: 'GPT-4',
      description: '最新的大语言模型，具有强大的理解和生成能力',
      avatar: '',
      default: true,
      high_context: true,
      tag: '通用,对话,创意',
      isExpanded: false
    },
    {
      id: 2,
      name: 'Code Assistant',
      description: '专门用于代码生成和代码审查的模型',
      avatar: '',
      default: false,
      high_context: false,
      tag: '编程',
      isExpanded: false
    }
  ]
  total.value = 100
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
    // TODO: 调用删除API
    ElMessage.success('删除成功')
    loadModels()
  } catch {
    // 取消删除
  }
}

// 设置默认模型
const handleSetDefault = async (model) => {
  // TODO: 调用API设置默认模型
  ElMessage.success(`已${model.default ? '设置' : '取消'}为默认模型`)
  loadModels()
}

// 提交表单
const handleSubmit = async () => {
  if (!modelFormRef.value) return
  await modelFormRef.value.validate(async (valid) => {
    if (valid) {
      // 处理头像 URL
      let avatar = modelForm.value.avatar
      if (avatar) {
        // 从完整 URL 中移除查询参数
        // 例如从 https://static.ai.txing.vip/1745426167657-466.png?q-sign-algorithm=sha1&...
        // 转换为 https://static.ai.txing.vip/1745426167657-466.png
        const urlObj = new URL(avatar)
        avatar = `${urlObj.origin}${urlObj.pathname}`
      }

      // 实际使用 formData
      const formData = {
        ...modelForm.value,
        tag: modelForm.value.tags.join(','),
        avatar: avatar // 使用处理后的 URL
      }

      console.log('Submit data:', formData) // 临时调试
      // TODO: 调用保存API
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      loadModels()
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

// 处理头像变更
const handleAvatarChange = (file) => {
  if (!file) return

  // 验证文件类型和大小
  const isImage = file.raw.type.startsWith('image/')
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('请上传图片文件！')
    return
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB！')
    return
  }

  // 读取文件并显示裁剪对话框
  const reader = new FileReader()
  reader.readAsDataURL(file.raw)
  reader.onload = (e) => {
    cropperImage.value = e.target.result
    cropperVisible.value = true
  }
}

// 处理图片裁剪
const handleCropImage = async () => {
  if (!cropperRef.value) return

  try {
    // 获取裁剪后的图片数据
    const base64Data = await new Promise((resolve) => {
      cropperRef.value.getCropData((data) => resolve(data))
    })


    // 将 base64 转换为 Blob
    const blob = await fetch(base64Data).then(res => res.blob())

    // 生成随机文件名
    const fileExt = 'png'
    const fileName = `${Date.now()}-${Math.floor(Math.random() * 1000)}.${fileExt}`

    try {
      // 获取预签名 URL
      const res = await defaultApi.apiCosPresignedUrlPost({
        type: 'upload',
        key: fileName
      })

      if (res.code !== 0) {
        ElMessage.error('获取预签名URL失败：' + res.msg)
        return
      }

      let presignedData = res.data

      // 上传文件到预签名 URL
      const response = await fetch(presignedData.url, {
        method: 'PUT',
        body: blob,
        headers: {
          'Content-Type': 'image/png'
        }
      })

      if (!response.ok) {
        console.error('上传失败:', response)
        throw new Error('上传失败')
      }

      // 获取待签名下载 URL，用于显示图片
      const res1 = await defaultApi.apiCosPresignedUrlPost({
        key: fileName,
        type: 'download'
      })
      if (res1.code !== 0) {
        ElMessage.error('获取预签名URL失败：' + res1.msg)
        return
      }
      presignedData = res1.data

      // 设置头像 URL
      modelForm.value.avatar = presignedData.url // 或者使用完整的访问 URL

      // 关闭裁剪对话框
      cropperVisible.value = false
      ElMessage.success('头像上传成功')
    } catch (error) {
      console.error('上传失败:', error)
      ElMessage.error('头像上传失败，请重试')
    }
  } catch (error) {
    console.error('裁剪失败:', error)
    ElMessage.error('图片裁剪失败，请重试')
  }
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

.avatar-wrapper {
  width: 120px;
  height: 120px;
  margin: 0 auto;
  border: 2px dashed var(--el-border-color);
  border-radius: 50%;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
  display: flex;
  justify-content: center;
  align-items: center;

  &:hover {
    border-color: var(--el-color-primary);
    transform: translateY(-2px);
  }

  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar-icon {
    font-size: 30px;
    color: var(--el-text-color-secondary);
  }
}

.cropper-container {
  height: 500px;
  background: #262626;

  :deep(.vue-cropper) {
    height: 100%;
    width: 100%;

    .cropper-view-box,
    .cropper-face {
      border-radius: 50%;
    }

    .cropper-dashed {
      display: none;
    }

    .cropper-view-box {
      outline: 1px solid #fff;
      box-shadow: 0 0 0 1px #fff;
    }

    .cropper-face {
      background-color: transparent !important;
    }

    .cropper-point {
      width: 8px !important;
      height: 8px !important;
      opacity: 1 !important;
      background-color: #fff !important;
      border-radius: 50%;

      &.point-se {
        width: 8px !important;
        height: 8px !important;
      }
    }

    .cropper-line {
      background: #fff;
      opacity: 0.3;
    }
  }
}
</style>
