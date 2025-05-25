<template>
  <div class="preset-container">
    <!-- 搜索表单 -->
    <el-card class="search-form">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="助手名称">
          <el-input v-model="searchForm.name" placeholder="请输入助手名称" clearable />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.official" placeholder="请选择类型" clearable>
            <el-option :value="true" label="官方助手" />
            <el-option :value="false" label="社区助手" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="searchForm.tags" placeholder="请选择标签" clearable multiple>
            <el-option v-for="tag in tagOptions" :key="tag.value" :label="tag.label" :value="tag.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增助手</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 助手列表 -->
    <div class="preset-list">
      <el-card v-for="preset in presets" :key="preset.id" class="preset-card">
        <div class="preset-header" @click="toggleExpand(preset)">
          <div class="preset-basic-info">
            <div class="preset-avatar">
              <el-avatar :size="48" :src="preset.avatar" @error="() => true">
                {{ preset.name.charAt(0) }}
              </el-avatar>
            </div>
            <div class="preset-info">
              <div class="preset-title">
                <h3>{{ preset.name }}</h3>
                <el-tag v-if="preset.official" type="success" effect="dark">官方</el-tag>
                <el-tag v-else type="primary" effect="dark">社区</el-tag>
                <el-tag
                  v-for="tag in preset.tags"
                  :key="tag"
                  :type="getTagType(tag)"
                  effect="light"
                  size="small"
                >
                  {{ getTagLabel(tag) }}
                </el-tag>
              </div>
              <div class="preset-description">{{ preset.description || '暂无描述' }}</div>
            </div>
          </div>
          <div class="preset-actions">
            <el-button text type="primary">
              {{ preset.isExpanded ? '收起' : '展开' }}
              <el-icon>
                <ArrowDown v-if="!preset.isExpanded" />
                <ArrowUp v-else />
              </el-icon>
            </el-button>
          </div>
        </div>

        <!-- 展开的详细信息 -->
        <el-collapse-transition>
          <div v-show="preset.isExpanded" class="preset-detail">
            <el-form
              :model="preset"
              label-width="100px"
              class="preset-form"
            >
              <el-form-item label="助手头像" class="avatar-uploader">
                <ImageUploader
                  v-model="preset.avatar"
                  :circle="true"
                  :crop-width="200"
                  :crop-height="200"
                  placeholder="上传头像"
                  @change="(file) => handleAvatarChange(file, preset)"
                  @success="() => handleUpdatePreset(preset)"
                />
              </el-form-item>
              <el-form-item label="助手名称">
                <el-input v-model="preset.name" placeholder="请输入助手名称" />
              </el-form-item>
              <el-form-item label="助手描述">
                <el-input
                  v-model="preset.description"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入助手描述"
                />
              </el-form-item>
              <el-form-item label="上下文设定">
                <el-input
                  v-model="preset.context"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入上下文设定"
                />
              </el-form-item>
              <el-form-item label="官方助手">
                <el-switch v-model="preset.official" />
              </el-form-item>
              <el-form-item label="标签">
                <el-select v-model="preset.tags" placeholder="请选择标签" multiple>
                  <el-option v-for="tag in tagOptions" :key="tag.value" :label="tag.label" :value="tag.value" />
                </el-select>
              </el-form-item>
              <div class="form-actions">
                <el-button type="primary" @click="handleSave(preset)">保存</el-button>
                <el-button type="danger" @click="handleDelete(preset)">删除</el-button>
              </div>
            </el-form>
          </div>
        </el-collapse-transition>
      </el-card>
    </div>

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

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增助手' : '编辑助手'"
      width="600px"
    >
      <el-form
        ref="presetFormRef"
        :model="presetForm"
        :rules="presetRules"
        label-width="100px"
      >
        <el-form-item label="助手头像" prop="avatar" class="avatar-uploader">
          <ImageUploader
            v-model="presetForm.avatar"
            :circle="true"
            :crop-width="200"
            :crop-height="200"
            placeholder="上传头像"
            @success="handleAvatarUploaded"
          />
        </el-form-item>
        <el-form-item label="助手名称" prop="name">
          <el-input v-model="presetForm.name" placeholder="请输入助手名称" />
        </el-form-item>
        <el-form-item label="助手描述" prop="description">
          <el-input
            v-model="presetForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入助手描述"
          />
        </el-form-item>
        <el-form-item label="上下文设定" prop="context">
          <el-input
            v-model="presetForm.context"
            type="textarea"
            :rows="4"
            placeholder="请输入上下文设定"
          />
        </el-form-item>
        <el-form-item label="官方助手" prop="official">
          <el-switch v-model="presetForm.official" />
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-select v-model="presetForm.tags" placeholder="请选择标签" multiple>
            <el-option v-for="tag in tagOptions" :key="tag.value" :label="tag.label" :value="tag.value" />
          </el-select>
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
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import VueCropper from 'vue-cropper/lib/vue-cropper.vue'
import 'vue-cropper/dist/index.css'
import ImageUploader from "@/components/common/ImageUploader.vue"
import { defaultApi } from '@/api/index.js'

// 搜索表单
const searchForm = ref({
  name: '',
  official: '',
  tags: ''
})

// 标签选项
const tagOptions = [
  { value: 'popular', label: '热门推荐' },
  { value: 'tools', label: '实用工具' },
  { value: 'writing', label: '文案创作' },
  { value: 'coding', label: '编码专家' },
  { value: 'learning', label: '知识学习' },
  { value: 'life', label: '生活指南' },
  { value: 'other', label: '其他' }
]

// 分页数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 助手数据
const presets = ref([])

// 对话框数据
const dialogVisible = ref(false)
const dialogType = ref('add')
const presetFormRef = ref(null)
const presetForm = ref({
  name: '',
  description: '',
  context: '',
  avatar: '',
  official: false,
  tags: []
})

// 裁剪相关
const cropperVisible = ref(false)
const cropperImage = ref('')
const cropperRef = ref(null)
const currentPreset = ref(null)

// 表单验证规则
const presetRules = {
  name: [
    { required: true, message: '请输入助手名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 200, message: '描述不能超过200个字符', trigger: 'blur' }
  ],
  context: [
    { required: true, message: '请输入上下文设定', trigger: 'blur' }
  ],
  tags: [
    { required: true, message: '请选择标签', trigger: 'change' },
    {
      validator: (rule, value, callback) => {
        if (value && value.length > 3) {
          callback(new Error('最多选择3个标签'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

// 加载助手列表
const loadPresets = async () => {
  try {
    const response = await defaultApi.apiPresetListGet(
      currentPage.value,
      pageSize.value,
      {
        orderBy: 'id',
        order: 'desc',
        official: searchForm.value.official === '' ? undefined : searchForm.value.official,
        name: searchForm.value.name === '' ? undefined : searchForm.value.name,
        tags: searchForm.value.tags ? searchForm.value.tags.join(',') : undefined
      }
    )
    if (response.code === 0 && response.data) {
      presets.value = response.data.records.map(preset => ({
        ...preset,
        isExpanded: false,
        tags: preset.tags ? preset.tags.split(',') : []
      }))
      total.value = response.data.total || 0
      currentPage.value = response.data.page || 1
      pageSize.value = response.data.limit || 10
    } else {
      ElMessage.error(response.msg || '获取助手列表失败')
    }
  } catch (error) {
    console.error('Load presets error:', error)
    ElMessage.error(error.body?.msg || '获取助手列表失败')
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadPresets()
}

// 重置搜索
const handleReset = () => {
  searchForm.value = {
    name: '',
    official: '',
    tags: ''
  }
  currentPage.value = 1
  loadPresets()
}

// 切换展开/收起
const toggleExpand = (preset) => {
  preset.isExpanded = !preset.isExpanded
}

// 新增助手
const handleAdd = () => {
  dialogType.value = 'add'
  presetForm.value = {
    name: '',
    description: '',
    context: '',
    avatar: '',
    official: false,
    tags: []
  }
  dialogVisible.value = true
}

// 编辑助手
const handleEdit = (preset) => {
  dialogType.value = 'edit'
  presetForm.value = {
    ...preset,
    tags: preset.tags || []
  }
  dialogVisible.value = true
}

// 保存助手
const handleSave = async (preset) => {
  try {
    // 处理头像 URL
    let avatar = preset.avatar
    if (avatar) {
      const urlObj = new URL(avatar)
      avatar = `${urlObj.origin}${urlObj.pathname}`
    }

    const formData = {
      name: preset.name,
      description: preset.description,
      context: preset.context,
      avatar: avatar,
      official: preset.official,
      tags: Array.isArray(preset.tags) ? preset.tags.join(',') : preset.tags
    }

    const response = await defaultApi.apiAdminPresetIdPut(preset.id, formData)
    if (response.code === 0) {
      ElMessage.success('保存成功')
      await loadPresets()
    } else {
      ElMessage.error(response.msg || '保存失败')
    }
  } catch (error) {
    console.error('Save preset error:', error)
    ElMessage.error(error.body?.msg || '保存失败')
  }
}

// 删除助手
const handleDelete = async (preset) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除助手【${preset.name}】吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await defaultApi.apiPresetIdDelete(preset.id)
    if (response.code === 0) {
      ElMessage.success('删除成功')
      await loadPresets()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete preset error:', error)
      ElMessage.error(error.body?.msg || '删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!presetFormRef.value) return
  await presetFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // 处理头像 URL
        let avatar = presetForm.value.avatar
        if (avatar) {
          const urlObj = new URL(avatar)
          avatar = `${urlObj.origin}${urlObj.pathname}`
        }

        // 构建请求数据
        const formData = {
          name: presetForm.value.name,
          description: presetForm.value.description,
          context: presetForm.value.context,
          avatar: avatar,
          official: presetForm.value.official,
          tags: presetForm.value.tags.join(',')
        }

        let response
        if (dialogType.value === 'add') {
          response = await defaultApi.apiPresetPost(formData)
        } else {
          response = await defaultApi.apiPresetIdPut(presetForm.value.id, formData)
        }

        if (response.code === 0) {
          ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
          dialogVisible.value = false
          await loadPresets()
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

// 更新助手头像
const handleUpdatePreset = async (preset) => {
  try {
    // 处理头像 URL
    let avatar = preset.avatar
    if (avatar) {
      const urlObj = new URL(avatar)
      avatar = `${urlObj.origin}${urlObj.pathname}`
    }

    const response = await defaultApi.apiAdminPresetIdPut(preset.id, {
      avatar: avatar
    })

    if (response.code === 0) {
      ElMessage.success('头像更新成功')
      await loadPresets()
    } else {
      ElMessage.error(response.msg || '头像更新失败')
    }
  } catch (error) {
    console.error('Update avatar error:', error)
    ElMessage.error(error.body?.msg || '头像更新失败')
  }
}

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadPresets()
}

// 页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  loadPresets()
}


// 处理头像变更（对话框中）
const handleAvatarUploaded = (url) => {
  presetForm.value.avatar = url
}

// 处理图片裁剪
const handleCropImage = () => {
  if (!cropperRef.value) return

  cropperRef.value.getCropData((data) => {
    if (currentPreset.value) {
      currentPreset.value.avatar = data
    } else {
      presetForm.value.avatar = data
    }
    cropperVisible.value = false
  })
}

// 获取标签类型
const getTagType = (tag) => {
  const typeMap = {
    popular: 'danger',
    tools: 'warning',
    writing: 'success',
    coding: 'info',
    learning: 'primary',
    life: 'warning',
    other: 'info'
  }
  return typeMap[tag] || 'info'
}

// 获取标签显示文本
const getTagLabel = (tag) => {
  const tagOption = tagOptions.find(option => option.value === tag)
  return tagOption ? tagOption.label : tag
}

// 页面加载时获取数据
onMounted(() => {
  loadPresets()
})
</script>

<style lang="scss" scoped>
.preset-container {
  padding: 24px;

  :deep(.el-card) {
    border-radius: 16px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    overflow: hidden;

    &:hover {
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
    }
  }
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

.preset-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
}

.preset-card {
  .preset-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 20px;
    cursor: pointer;
  }

  .preset-basic-info {
    display: flex;
    gap: 20px;
    align-items: flex-start;
  }

  .preset-info {
    flex: 1;
  }

  .preset-title {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 8px;

    h3 {
      margin: 0;
      font-size: 18px;
      line-height: 1.4;
    }

    :deep(.el-tag) {
      margin-right: 4px;
    }
  }

  .preset-description {
    color: var(--el-text-color-secondary);
    font-size: 14px;
    line-height: 1.6;
  }

  .preset-detail {
    padding: 0 20px 20px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .preset-form {
    padding-top: 20px;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
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
