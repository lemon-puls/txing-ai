<template>
  <div class="page-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="keyword"
          placeholder="搜索标题 / slug / 摘要"
          clearable
          style="width: 260px"
          @keyup.enter="search"
          @clear="search"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button round @click="search">搜索</el-button>
      </div>
      <div class="toolbar-right">
        <el-button round type="primary" plain @click="exportAll" :loading="exporting">
          <el-icon><Download /></el-icon>
          <span>导出 md</span>
        </el-button>
        <el-button round circle :loading="loading" @click="fetchList">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 页面列表 -->
    <el-table :data="list" v-loading="loading" class="rounded-table">
      <el-table-column prop="slug" label="slug" width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="slug-code">{{ row.slug }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag round effect="light">{{ row.pageType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="80">
        <template #default="{ row }">v{{ row.version }}</template>
      </el-table-column>
      <el-table-column prop="summary" label="摘要" min-width="220" show-overflow-tooltip />
      <el-table-column prop="updateTime" label="更新时间" width="170">
        <template #default="{ row }">{{ formatTime(row.updateTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button round size="small" @click="view(row)">查看</el-button>
          <el-button round size="small" type="primary" plain @click="edit(row)">编辑</el-button>
          <el-button round size="small" type="warning" plain @click="offline(row)">下线</el-button>
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

    <!-- 编辑弹窗 -->
    <PageEditDialog
      v-model:visible="editVisible"
      :page="editing"
      :is-draft="false"
      @saved="fetchList"
    />

    <!-- 查看弹窗（不挂 markdown-body 类，避免 github-markdown-dark 全局污染背景） -->
    <el-dialog v-model="viewVisible" :title="viewing?.title" width="760px" top="6vh">
      <div class="page-view" v-html="renderedContent"></div>
    </el-dialog>
  </div>
</template>

<script setup name="WikiPagePanel">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, Refresh } from '@element-plus/icons-vue'
import { renderWikiMarkdown } from '@/utils/wikiMarkdown'
import wikiApi from '@/api/wiki'
import PageEditDialog from './PageEditDialog.vue'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const keyword = ref('')
const loading = ref(false)
const exporting = ref(false)

const editVisible = ref(false)
const editing = ref(null)
const viewVisible = ref(false)
const viewing = ref(null)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

const renderedContent = computed(() => {
  if (!viewing.value) return ''
  try {
    return renderWikiMarkdown(viewing.value.content)
  } catch {
    return viewing.value.content
  }
})

const fetchList = async () => {
  loading.value = true
  try {
    const response = await wikiApi.listPublished(page.value, pageSize, keyword.value.trim())
    if (response.code === 0) {
      list.value = response.data.list || []
      total.value = response.data.total || 0
    } else {
      ElMessage.error(response.msg || '加载失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载页面列表失败')
  } finally {
    loading.value = false
  }
}

const search = () => {
  page.value = 1
  fetchList()
}

const view = (row) => {
  viewing.value = row
  viewVisible.value = true
}

const edit = (row) => {
  editing.value = row
  editVisible.value = true
}

const offline = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认下线「${row.title}」？下线后 AI 问答不再引用该页。`,
      '下线页面', { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    const response = await wikiApi.offlinePage(row.id)
    if (response.code === 0) {
      ElMessage.success('已下线')
      fetchList()
    } else {
      ElMessage.error(response.msg || '下线失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('下线失败')
  }
}

const exportAll = async () => {
  exporting.value = true
  try {
    const { blob, filename } = await wikiApi.exportAll()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('已导出')
  } catch (e) {
    console.error(e)
    ElMessage.error(e.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(fetchList)

// 供父组件在 Tab 切换时触发刷新
defineExpose({ refresh: fetchList })
</script>

<style scoped lang="scss">
.page-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .toolbar-left,
  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.panel-pagination {
  display: flex;
  justify-content: flex-end;
}

.slug-code {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12.5px;
  background: var(--el-fill-color-light);
  padding: 2px 8px;
  border-radius: 6px;
}

.page-view {
  max-height: 70vh;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.8;

  h1, h2, h3 { margin: 16px 0 8px; }
  h4, h5 { margin: 12px 0 6px; }
  p { margin: 8px 0; }
  blockquote {
    margin: 8px 0;
    padding: 4px 12px;
    border-left: 3px solid var(--el-color-primary-light-5);
    color: var(--el-text-color-secondary);
  }
  a { color: var(--el-color-primary); }
  code {
    background: var(--el-fill-color-light);
    padding: 1px 6px;
    border-radius: 6px;
    font-size: 13px;
  }
  pre {
    background: var(--el-fill-color-light);
    border-radius: 10px;
    padding: 12px;
    overflow-x: auto;
  }

  // [[slug|标题]] 互链标签（utils/wikiMarkdown 转换）
  .wiki-cite {
    display: inline-block;
    padding: 0 8px;
    margin: 0 2px;
    border-radius: 999px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-size: 12.5px;
    line-height: 20px;
    white-space: nowrap;

    &::before { content: '🔗 '; font-size: 11px; }
  }
}
</style>
