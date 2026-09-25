<template>
  <div class="source-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <el-button round type="primary" @click="mdDialogVisible = true">
          <el-icon><UploadFilled /></el-icon>
          <span>新建 MD 源</span>
        </el-button>
        <el-button round @click="urlDialogVisible = true">
          <el-icon><Link /></el-icon>
          <span>新建 URL 源</span>
        </el-button>
      </div>
      <el-button round circle :loading="loading" @click="fetchList">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <!-- 源列表 -->
    <el-table :data="list" v-loading="loading" class="rounded-table">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="row.sourceType === 'md' ? 'primary' : 'warning'" round effect="light">
            {{ row.sourceType === 'md' ? 'Markdown' : 'URL' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusMeta(row.status).type" round effect="light">
            <el-icon v-if="row.status === 'ingesting'" class="is-loading" style="margin-right:4px">
              <Loading />
            </el-icon>
            {{ statusMeta(row.status).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="errorMsg" label="错误信息" min-width="160" show-overflow-tooltip />
      <el-table-column prop="ingestedAt" label="编译时间" width="170">
        <template #default="{ row }">{{ formatTime(row.ingestedAt) }}</template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="170">
        <template #default="{ row }">{{ formatTime(row.createTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button
            round size="small" type="primary" plain
            :disabled="row.status === 'ingesting'"
            :loading="ingestingId === row.id"
            @click="ingest(row)"
          >编译</el-button>
          <el-button round size="small" type="danger" plain @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="panel-pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        background
        @current-change="fetchList"
      />
    </div>

    <!-- 新建 MD 源 -->
    <el-dialog v-model="mdDialogVisible" title="新建 Markdown 源" width="640px" :close-on-click-modal="false">
      <el-form :model="mdForm" label-width="70px">
        <el-form-item label="标题">
          <el-input v-model="mdForm.title" maxlength="255" placeholder="如：个人简历 2026" />
        </el-form-item>
        <el-form-item label="正文">
          <div class="md-upload-row">
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              accept=".md,.markdown,.txt"
              :on-change="handleFileChange"
            >
              <el-button round size="small">
                <el-icon><Document /></el-icon>
                <span>从文件读取</span>
              </el-button>
            </el-upload>
            <span v-if="mdForm.fileName" class="md-file-name">{{ mdForm.fileName }}</span>
          </div>
          <el-input
            v-model="mdForm.content"
            type="textarea"
            :rows="12"
            class="md-editor"
            placeholder="粘贴 markdown 原文，或点击上方按钮从本地文件读取"
          />
          <div class="md-char-hint" :class="{ 'is-over': mdForm.content.length > 1000000 }">
            {{ mdForm.content.length.toLocaleString() }} 字符
            <template v-if="mdForm.content.length > 12000">
              · 编译时将分片为约 {{ Math.ceil(mdForm.content.length / 12000) }} 段逐片处理
            </template>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button round @click="mdDialogVisible = false">取消</el-button>
        <el-button round type="primary" :loading="creating" @click="createMD">创建</el-button>
      </template>
    </el-dialog>

    <!-- 新建 URL 源 -->
    <el-dialog v-model="urlDialogVisible" title="新建 URL 源" width="520px" :close-on-click-modal="false">
      <el-form :model="urlForm" label-width="70px">
        <el-form-item label="标题">
          <el-input v-model="urlForm.title" maxlength="255" placeholder="如：我的技术博客" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="urlForm.url" placeholder="https://…" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button round @click="urlDialogVisible = false">取消</el-button>
        <el-button round type="primary" :loading="creating" @click="createURL">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="WikiSourcePanel">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, Link, Refresh, Document, Loading } from '@element-plus/icons-vue'
import wikiApi from '@/api/wiki'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const creating = ref(false)
const ingestingId = ref(null)
let pollTimer = null

const mdDialogVisible = ref(false)
const urlDialogVisible = ref(false)
const mdForm = reactive({ title: '', content: '', fileName: '' })
const urlForm = reactive({ title: '', url: '' })

const statusMeta = (status) => ({
  pending: { label: '待编译', type: 'info' },
  ingesting: { label: '编译中', type: 'warning' },
  ingested: { label: '已入库', type: 'success' },
  failed: { label: '失败', type: 'danger' }
}[status] || { label: status, type: 'info' })

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

const fetchList = async () => {
  loading.value = true
  try {
    const response = await wikiApi.listSources(page.value, pageSize)
    if (response.code === 0) {
      list.value = response.data.list || []
      total.value = response.data.total || 0
      schedulePoll()
    } else {
      ElMessage.error(response.msg || '加载失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载源列表失败')
  } finally {
    loading.value = false
  }
}

// 有进行中的编译时轮询刷新（ingest 为后台异步任务）
const schedulePoll = () => {
  const busy = list.value.some(row => row.status === 'pending' || row.status === 'ingesting')
  if (busy && !pollTimer) {
    pollTimer = setTimeout(async () => {
      pollTimer = null
      if (list.value.some(row => row.status === 'pending' || row.status === 'ingesting')) {
        await fetchList()
      }
    }, 3000)
  }
}

const handleFileChange = (file) => {
  const reader = new FileReader()
  reader.onload = () => {
    mdForm.content = String(reader.result || '')
    mdForm.fileName = file.name
    if (!mdForm.title) mdForm.title = file.name.replace(/\.(md|markdown|txt)$/i, '')
  }
  reader.readAsText(file.raw)
}

const createMD = async () => {
  if (!mdForm.title.trim() || !mdForm.content.trim()) {
    ElMessage.warning('标题和正文不能为空')
    return
  }
  // 与后端 max=1000000 对齐的前置校验（超限时后端只回笼统的 invalid params）
  const MAX_CONTENT_CHARS = 1000000
  if (mdForm.content.length > MAX_CONTENT_CHARS) {
    ElMessage.warning(`正文过长（${mdForm.content.length.toLocaleString()} 字符），上限 ${MAX_CONTENT_CHARS.toLocaleString()}，请拆分后分批上传`)
    return
  }
  creating.value = true
  try {
    const response = await wikiApi.createMDSource(mdForm.title.trim(), mdForm.content)
    if (response.code === 0) {
      ElMessage.success('源已创建')
      mdDialogVisible.value = false
      mdForm.title = ''
      mdForm.content = ''
      mdForm.fileName = ''
      fetchList()
    } else {
      ElMessage.error(response.msg || '创建失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

const createURL = async () => {
  if (!urlForm.title.trim() || !urlForm.url.trim()) {
    ElMessage.warning('标题和地址不能为空')
    return
  }
  creating.value = true
  try {
    const response = await wikiApi.createURLSource(urlForm.title.trim(), urlForm.url.trim())
    if (response.code === 0) {
      ElMessage.success('源已创建')
      urlDialogVisible.value = false
      urlForm.title = ''
      urlForm.url = ''
      fetchList()
    } else {
      ElMessage.error(response.msg || '创建失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

const ingest = async (row) => {
  ingestingId.value = row.id
  try {
    const response = await wikiApi.triggerIngest(row.id)
    if (response.code === 0) {
      ElMessage.success('编译已开始，可稍后刷新查看进度')
      fetchList()
    } else {
      ElMessage.error(response.msg || '触发编译失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('触发编译失败')
  } finally {
    ingestingId.value = null
  }
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除源「${row.title}」？已生成的 wiki 页面不受影响。`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    const response = await wikiApi.deleteSource(row.id)
    if (response.code === 0) {
      ElMessage.success('已删除')
      fetchList()
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('删除失败')
  }
}

onMounted(fetchList)
onBeforeUnmount(() => {
  if (pollTimer) clearTimeout(pollTimer)
})

// 供父组件在 Tab 切换时触发刷新
defineExpose({ refresh: fetchList })
</script>

<style scoped lang="scss">
.source-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .toolbar-left {
    display: flex;
    gap: 12px;
  }
}

.panel-pagination {
  display: flex;
  justify-content: flex-end;
}

.md-upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;

  .md-file-name {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}

.md-editor {
  margin-top: 8px;
  :deep(.el-textarea__inner) {
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 13px;
    line-height: 1.6;
  }
}

.md-char-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  text-align: right;

  &.is-over { color: var(--el-color-danger); }
}
</style>
