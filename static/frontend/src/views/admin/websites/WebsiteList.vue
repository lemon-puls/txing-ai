<template>
  <div class="website-page">
    <!-- 页头 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-icon">
          <el-icon :size="18"><Link /></el-icon>
        </div>
        <div class="header-text">
          <div class="header-title">网站导航管理</div>
          <div class="header-subtitle">管理对外展示的精选站点，支持 AI 智能录入</div>
        </div>
      </div>
      <div class="header-actions">
        <el-button round @click="handleAdd">
          <el-icon class="btn-icon"><Plus /></el-icon>
          新增网站
        </el-button>
        <el-button round class="ai-btn" @click="openOpsDrawer">
          <el-icon class="btn-icon"><MagicStick /></el-icon>
          AI 录入
        </el-button>
      </div>
    </div>

    <!-- 筛选工具栏 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-bar">
        <div class="filter-row">
          <el-input
            v-model="searchForm.name"
            class="search-input"
            placeholder="搜索网站名称，回车确认"
            clearable
            :prefix-icon="Search"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <span class="filter-spacer"></span>
          <el-tooltip content="刷新列表" placement="top">
            <el-button circle class="refresh-btn" @click="loadWebsites">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
        <div class="tag-chips">
          <button class="chip" :class="{ active: !searchForm.tag }" @click="handleTagFilter('')">全部</button>
          <button
            v-for="tag in PRESET_WEBSITE_TAGS"
            :key="`preset-${tag}`"
            class="chip preset"
            :class="{ active: searchForm.tag === tag }"
            @click="handleTagFilter(tag)"
          >
            {{ tag }}
          </button>
          <button
            v-for="tag in customChipTags"
            :key="tag"
            class="chip"
            :class="{ active: searchForm.tag === tag }"
            @click="handleTagFilter(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </el-card>

    <!-- 网站列表 -->
    <el-card class="table-card" shadow="never">
      <div class="table-head">
        <span class="table-title">
          网站列表
          <span class="total-badge">共 {{ total }} 个</span>
        </span>
      </div>

      <el-table v-show="websites.length > 0 || loading" :data="websites" v-loading="loading" style="width: 100%">
        <el-table-column label="网站信息" min-width="330">
          <template #default="{ row }">
            <div class="website-info">
              <div class="website-avatar">
                <span class="avatar-letter">{{ (row.name || '?').slice(0, 1).toUpperCase() }}</span>
                <img
                  v-if="row.avatar"
                  :src="row.avatar"
                  :alt="row.name"
                  @error="handleImageError"
                />
              </div>
              <div class="website-details">
                <a class="website-name" :href="row.url" target="_blank" rel="noopener noreferrer" :title="row.name">
                  {{ row.name }}
                  <el-icon :size="11" class="name-link-icon"><TopRight /></el-icon>
                </a>
                <div class="website-url">{{ row.url }}</div>
                <div class="website-description" :title="row.description">{{ row.description }}</div>
                <div class="tag-list">
                  <el-tag
                    v-for="tag in orderTagsPresetFirst(splitTags(row.tags))"
                    :key="tag"
                    :type="isPresetTag(tag) ? 'primary' : tagType(tag)"
                    :effect="isPresetTag(tag) ? 'light' : 'plain'"
                    size="small"
                    round
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status"
              :active-value="1"
              :inactive-value="0"
              :disabled="statusUpdating.has(row.id)"
              inline-prompt
              active-text="启"
              inactive-text="停"
              @change="(val) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="100">
          <template #default="{ row }">
            <el-tooltip :content="formatISODate(row.createdAt)" placement="top">
              <span class="cell-time">{{ getRelativeTime(row.createdAt) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button link type="primary" :icon="EditPen" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" :icon="Delete" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && websites.length === 0" class="empty-state">
        <el-empty :description="searchForm.name || searchForm.tag ? '没有符合筛选条件的网站' : '还没有收录任何网站'">
          <div class="empty-actions">
            <el-button v-if="isFiltered" round @click="handleReset">清空筛选</el-button>
            <el-button round class="ai-btn" @click="openOpsDrawer">
              <el-icon class="btn-icon"><MagicStick /></el-icon>
              AI 录入
            </el-button>
            <el-button v-if="!isFiltered" round type="primary" @click="handleAdd">
              <el-icon class="btn-icon"><Plus /></el-icon>
              新增网站
            </el-button>
          </div>
        </el-empty>
      </div>

      <!-- 分页 -->
      <div v-if="total > 0" class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          background
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- AI 录入抽屉：运营助手对话，提案确认后刷新列表（destroy-on-close 保证每次打开重新加载会话） -->
    <el-drawer
      v-model="opsDrawerVisible"
      size="480px"
      class="ops-drawer"
      destroy-on-close
    >
      <template #header>
        <span class="ops-drawer-title">
          <span class="title-icon">
            <el-icon :size="14"><MagicStick /></el-icon>
          </span>
          <span class="title-text">
            AI 录入助手
            <small>提案确认后才会入库</small>
          </span>
        </span>
      </template>
      <OpsChatPanel :context="opsContext" style="height: 100%" @inserted="loadWebsites" />
    </el-drawer>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增网站' : '编辑网站'"
      width="620px"
      class="website-dialog"
      align-center
      :close-on-click-modal="false"
    >
      <el-form
        ref="websiteFormRef"
        :model="websiteForm"
        :rules="websiteRules"
        label-width="90px"
        label-position="right"
      >
        <el-form-item label="网站名称" prop="name">
          <el-input v-model="websiteForm.name" placeholder="请输入网站名称" maxlength="50" />
        </el-form-item>

        <el-form-item label="网站地址" prop="url">
          <el-input v-model="websiteForm.url" placeholder="请输入网站地址，如 https://example.com" />
        </el-form-item>

        <el-form-item label="网站描述" prop="description">
          <el-input
            v-model="websiteForm.description"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="一句话介绍这个网站"
          />
        </el-form-item>

        <el-form-item label="网站头像" prop="avatar">
          <div class="avatar-field">
            <div class="avatar-preview">
              <span v-if="!websiteForm.avatar" class="avatar-placeholder">
                <el-icon :size="20"><Link /></el-icon>
              </span>
              <img v-else :src="websiteForm.avatar" alt="头像预览" />
            </div>
            <div class="avatar-ops">
              <el-input
                v-model="websiteForm.avatar"
                placeholder="头像 URL，可自动获取或上传"
                clearable
              />
              <div class="avatar-actions">
                <el-button size="small" round :loading="avatarLoading" @click="autoGetAvatar">
                  <el-icon class="btn-icon"><Download /></el-icon>
                  自动获取
                </el-button>
                <el-upload
                  class="avatar-uploader"
                  :show-file-list="false"
                  :before-upload="beforeAvatarUpload"
                  :http-request="uploadAvatar"
                >
                  <el-button size="small" round>
                    <el-icon class="btn-icon"><Upload /></el-icon>
                    上传头像
                  </el-button>
                </el-upload>
              </div>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="标签" prop="tags">
          <div class="tags-editor">
            <div class="preset-tags">
              <el-tag
                v-for="pt in PRESET_WEBSITE_TAGS"
                :key="pt"
                class="preset-tag"
                :type="websiteForm.tags.includes(pt) ? 'primary' : 'info'"
                :effect="websiteForm.tags.includes(pt) ? 'light' : 'plain'"
                size="small"
                round
                @click="togglePresetTag(pt)"
              >
                {{ pt }}
              </el-tag>
            </div>
            <div class="tags-input">
              <el-tag
                v-for="tag in websiteForm.tags"
                :key="tag"
                :type="isPresetTag(tag) ? 'primary' : tagType(tag)"
                :effect="isPresetTag(tag) ? 'light' : 'plain'"
                closable
                size="small"
                round
                @close="removeTag(tag)"
              >
                {{ tag }}
              </el-tag>
              <el-input
                v-if="inputVisible"
                ref="inputRef"
                v-model="inputValue"
                size="small"
                class="tag-editor-input"
                placeholder="自定义标签，回车确认"
                @keyup.enter="handleInputConfirm"
                @blur="handleInputConfirm"
              />
              <el-button v-else size="small" round link type="primary" @click="showInput">
                <el-icon class="btn-icon"><Plus /></el-icon>
                添加标签
              </el-button>
            </div>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button round @click="dialogVisible = false">取消</el-button>
          <el-button round type="primary" class="submit-btn" :loading="submitLoading" @click="handleSubmit">
            {{ dialogType === 'add' ? '创建' : '保存' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="WebsiteList">
import { ref, onMounted, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Link, MagicStick, Plus, Search, Refresh, TopRight, EditPen, Delete, Download, Upload } from '@element-plus/icons-vue'
import { defaultApi } from '@/api'
import OpsChatPanel from '@/components/ops/OpsChatPanel.vue'
import { formatISODate, getRelativeTime } from '@/utils/timeUtils'
import { PRESET_WEBSITE_TAGS, isPresetTag, splitTags, orderTagsPresetFirst } from '@/constants/websiteTags'

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

// 已知标签并集（跨筛选累积，避免筛选后标签 chips 收缩）
const knownTags = ref(new Set())
const allTags = computed(() => Array.from(knownTags.value))
// 筛选 chips：内定分类标签恒在且置前（来自常量），其余实际标签排后
const customChipTags = computed(() => allTags.value.filter((t) => !isPresetTag(t)))

// 是否处于筛选状态（空状态文案与 CTA 用）
const isFiltered = computed(() => !!(searchForm.value.name || searchForm.value.tag))

// 状态切换中的行（防止重复点击）
const statusUpdating = ref(new Set())

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

// AI 录入抽屉（打开时若新增对话框已填地址则带入草稿上下文）
const opsDrawerVisible = ref(false)
const opsDraftUrl = ref('')
const opsContext = computed(() => {
  const ctx = { page: 'websites' }
  if (opsDraftUrl.value) ctx.draft = { url: opsDraftUrl.value }
  return ctx
})

const openOpsDrawer = () => {
  opsDraftUrl.value = (dialogType.value === 'add' && dialogVisible.value && websiteForm.value.url)
    ? websiteForm.value.url
    : ''
  opsDrawerVisible.value = true
}

// 标签按名称哈希固定配色，同一标签始终同色
const TAG_TYPES = ['primary', 'success', 'warning', 'danger', 'info']
const tagType = (tag) => {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = (hash * 31 + tag.charCodeAt(i)) >>> 0
  }
  return TAG_TYPES[hash % TAG_TYPES.length]
}

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
    const response = await defaultApi.apiWebsitesListGet({
      page: currentPage.value,
      limit: pageSize.value,
      name: searchForm.value.name || undefined,
      tag: searchForm.value.tag || undefined
    })
    if (response.code === 0) {
      websites.value = response.data.records || []
      total.value = response.data.total || 0
      // 累积标签并集，供筛选 chips 使用
      websites.value.forEach((w) => {
        ;(w.tags || '').split(',').forEach((t) => t && knownTags.value.add(t))
      })
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

// 标签筛选
const handleTagFilter = (tag) => {
  searchForm.value.tag = tag
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

// 启用/停用（失败回滚开关状态）
const handleStatusChange = async (row, val) => {
  statusUpdating.value.add(row.id)
  try {
    const response = await defaultApi.apiAdminWebsitesIdPut(row.id, { status: val })
    if (response.code === 0) {
      row.status = val
      ElMessage.success(val === 1 ? '已启用' : '已停用')
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('切换网站状态失败:', error)
    ElMessage.error('操作失败')
  } finally {
    statusUpdating.value.delete(row.id)
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
      // 删完后当前页只剩一条时回退一页，避免空页
      if (websites.value.length === 1 && currentPage.value > 1) {
        currentPage.value -= 1
      }
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
// 点击内定分类标签快捷加入/移除
const togglePresetTag = (tag) => {
  const index = websiteForm.value.tags.indexOf(tag)
  if (index > -1) {
    websiteForm.value.tags.splice(index, 1)
  } else {
    websiteForm.value.tags.push(tag)
  }
}

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
  const value = inputValue.value.trim()
  if (value && !websiteForm.value.tags.includes(value)) {
    websiteForm.value.tags.push(value)
  }
  inputVisible.value = false
  inputValue.value = ''
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
.website-page {
  min-height: 100%;
  padding: 16px 20px 20px;
  box-sizing: border-box;
  // 顶部淡淡的 primary 氛围渐变
  background: linear-gradient(180deg, var(--el-color-primary-light-9), transparent 300px);
}

// ===== 页头 =====
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-icon {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
    box-shadow: 0 4px 12px var(--el-color-primary-light-8);
    flex-shrink: 0;
  }

  .header-title {
    font-size: 17px;
    font-weight: 700;
    line-height: 1.3;
  }

  .header-subtitle {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 2px;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 4px;

    :deep(.el-button) {
      padding: 10px 20px;

      .btn-icon {
        margin-right: 4px;
      }
    }
  }
}

// AI 录入按钮：渐变强化 CTA
.ai-btn {
  border: none !important;
  color: #fff !important;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary)) !important;
  box-shadow: 0 2px 10px var(--el-color-primary-light-7);

  &:hover {
    opacity: 0.88;
  }
}

// ===== 筛选工具栏 =====
.filter-card {
  margin-bottom: 16px;
  border-radius: 14px;

  :deep(.el-card__body) {
    padding: 12px 16px;
  }
}

.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;

  .filter-row {
    display: flex;
    align-items: center;
    gap: 12px;

    .search-input {
      width: 260px;

      :deep(.el-input__wrapper) {
        border-radius: 999px;
      }
    }

    .filter-spacer {
      flex: 1;
    }

    .refresh-btn {
      flex-shrink: 0;
    }
  }

  .tag-chips {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    // 最多两行，其余滚动查看
    max-height: 62px;
    overflow-y: auto;
    scrollbar-width: thin;

    .chip {
      flex-shrink: 0;
      padding: 4px 14px;
      font-size: 12px;
      line-height: 1.5;
      color: var(--el-text-color-regular);
      background: var(--el-fill-color-light);
      border: 1px solid transparent;
      border-radius: 999px;
      cursor: pointer;
      transition: all 0.2s ease;

      // 内定分类标签：常亮 primary 底色以突出
      &.preset {
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
        font-weight: 500;

        &:hover {
          background: var(--el-color-primary-light-8);
        }
      }

      &:hover {
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
      }

      &.active {
        color: #fff;
        background: var(--el-color-primary);
        border-color: var(--el-color-primary);
      }
    }
  }
}

// ===== 列表卡片 =====
.table-card {
  border-radius: 14px;

  :deep(.el-card__body) {
    padding: 16px;
  }

  .table-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;

    .table-title {
      font-size: 15px;
      font-weight: 600;

      .total-badge {
        margin-left: 8px;
        font-size: 12px;
        font-weight: 400;
        color: var(--el-text-color-secondary);
      }
    }
  }

  :deep(.el-table) {
    --el-table-header-bg-color: transparent;

    th.el-table__cell {
      font-weight: 600;
      color: var(--el-text-color-secondary);
    }
  }

  .website-info {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 4px 0;

    .website-avatar {
      position: relative;
      width: 46px;
      height: 46px;
      border-radius: 12px;
      overflow: hidden;
      background: linear-gradient(135deg, var(--el-color-primary-light-8), var(--el-color-primary-light-6));
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
      box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.04);

      .avatar-letter {
        font-size: 18px;
        font-weight: 700;
        color: var(--el-color-primary);
      }

      img {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .website-details {
      flex: 1;
      min-width: 0;

      .website-name {
        display: inline-flex;
        align-items: center;
        gap: 3px;
        font-weight: 600;
        font-size: 14px;
        color: var(--el-text-color-primary);
        text-decoration: none;
        margin-bottom: 2px;
        transition: color 0.2s ease;

        .name-link-icon {
          color: var(--el-text-color-placeholder);
          transition: color 0.2s ease;
        }

        &:hover {
          color: var(--el-color-primary);

          .name-link-icon {
            color: var(--el-color-primary);
          }
        }
      }

      .website-url {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        margin-bottom: 2px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .website-description {
        font-size: 12px;
        color: var(--el-text-color-placeholder);
        line-height: 1.5;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .tag-list {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 4px;
        margin-top: 4px;
      }
    }
  }

  .cell-time {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    font-variant-numeric: tabular-nums;
  }

  .row-actions {
    display: inline-flex;
    align-items: center;

    :deep(.el-button) {
      padding: 6px 5px;
    }

    :deep(.el-button + .el-button) {
      margin-left: 6px;
    }
  }
}

// 分页
.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;

  :deep(.el-pagination) {
    --el-pagination-border-radius: 8px;

    .el-pager li,
    .btn-prev,
    .btn-next {
      border-radius: 8px;
    }
  }
}

// ===== 空状态 =====
.empty-state {
  padding: 30px 0 40px;

  .empty-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;

    .btn-icon {
      margin-right: 4px;
    }
  }
}

// ===== 对话框 =====
.avatar-field {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  width: 100%;

  .avatar-preview {
    width: 64px;
    height: 64px;
    border-radius: 12px;
    overflow: hidden;
    flex-shrink: 0;
    background: var(--el-fill-color-light);
    border: 1px dashed var(--el-border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-placeholder);

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .avatar-ops {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .avatar-actions {
      display: flex;
      gap: 8px;
    }
  }
}

.tags-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;

  .preset-tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
    padding-bottom: 10px;
    border-bottom: 1px dashed var(--el-border-color-lighter);

    .preset-tag {
      cursor: pointer;
      transition: all 0.2s ease;

      &:hover {
        transform: translateY(-1px);
      }
    }
  }
}

.tags-input {
  width: 100%;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;

  .tag-editor-input {
    width: 130px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;

  .submit-btn {
    border: none;
    background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
    box-shadow: 0 2px 8px var(--el-color-primary-light-8);

    &:not(:disabled):hover {
      opacity: 0.88;
    }
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  .website-page {
    padding: 12px;
  }

  .page-header {
    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }

  .filter-bar {
    .filter-row {
      flex-wrap: wrap;

      .search-input {
        width: 100%;
      }
    }
  }
}

// ===== 暗色模式补充：项目暗色主题未覆写部分变量，这里手动补齐 =====
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.website-page {
  html.dark & {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.03), transparent 300px);
  }
}

.filter-card,
.table-card {
  html.dark & {
    background: #1c1c1c;
    border-color: #363637;
  }
}

.website-page .ai-btn {
  html.dark & {
    box-shadow: none;
  }
}

.website-page .filter-bar .tag-chips .chip {
  html.dark & {
    background: #262627;
    color: var(--el-text-color-regular);

    &:hover {
      background: var(--el-color-primary-light-9);
    }

    &.preset {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }

    &.active {
      color: #fff;
      background: var(--el-color-primary);
    }
  }
}

.table-card .website-info .website-avatar {
  html.dark & {
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06);
  }
}
</style>

<style lang="scss">
// 弹层类组件（dialog / drawer）内容不生效 scoped，需全局覆写
.website-dialog {
  border-radius: 18px;

  .el-dialog__header {
    padding: 18px 24px 14px;
    margin-right: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);

    .el-dialog__title {
      font-size: 16px;
      font-weight: 600;
    }
  }

  .el-dialog__body {
    padding: 20px 24px;
  }

  .el-dialog__footer {
    padding: 12px 24px 18px;
    border-top: 1px solid var(--el-border-color-lighter);
  }
}

// AI 录入抽屉
.ops-drawer {
  .el-drawer__header {
    margin-bottom: 0;
    padding: 14px 20px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .el-drawer__body {
    padding: 0;
    overflow: hidden;
  }

  .ops-drawer-title {
    display: inline-flex;
    align-items: center;
    gap: 10px;

    .title-icon {
      width: 30px;
      height: 30px;
      border-radius: 9px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
      box-shadow: 0 2px 8px var(--el-color-primary-light-8);
    }

    .title-text {
      display: flex;
      flex-direction: column;
      font-size: 15px;
      font-weight: 600;
      line-height: 1.3;
      color: var(--el-text-color-primary);

      small {
        font-size: 12px;
        font-weight: 400;
        color: var(--el-text-color-secondary);
      }
    }
  }
}
</style>
