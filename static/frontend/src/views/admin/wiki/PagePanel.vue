<template>
  <div class="page-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="keyword"
          placeholder="搜索标题 / slug / 摘要"
          clearable
          style="width: 220px"
          @keyup.enter="search"
          @clear="search"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="pageType" clearable placeholder="全部类型" style="width: 130px" @change="search">
          <el-option v-for="(label, val) in TYPE_LABELS" :key="val" :label="label" :value="val" />
        </el-select>
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

    <!-- 页面列表（展开行直接查看全文，无需弹窗） -->
    <el-table :data="list" row-key="id" v-loading="loading" class="rounded-table">
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="page-expand">
            <div class="expand-meta">
              <el-tag :type="TYPE_TAGS[row.pageType] || 'info'" round effect="light">
                {{ TYPE_LABELS[row.pageType] || row.pageType }}
              </el-tag>
              <span v-if="row.aliases && row.aliases.length" class="meta-item">
                别名：{{ row.aliases.join(' / ') }}
              </span>
              <span class="meta-item">v{{ row.version }}</span>
              <span class="meta-item">{{ formatTime(row.updateTime) }}</span>
              <div class="expand-actions">
                <el-button round size="small" type="primary" plain @click="edit(row)">编辑</el-button>
                <el-button round size="small" type="warning" plain @click="offline(row)">下线</el-button>
              </div>
            </div>
            <div class="expand-content" v-html="renderRow(row)"></div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="slug" label="slug" width="280" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="slug-code">{{ row.slug }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="TYPE_TAGS[row.pageType] || 'info'" round effect="light">
            {{ TYPE_LABELS[row.pageType] || row.pageType }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="80">
        <template #default="{ row }">v{{ row.version }}</template>
      </el-table-column>
      <el-table-column prop="summary" label="摘要" min-width="220" show-overflow-tooltip />
      <el-table-column prop="updateTime" label="更新时间" width="170">
        <template #default="{ row }">{{ formatTime(row.updateTime) }}</template>
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
  </div>
</template>

<script setup name="WikiPagePanel">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, Refresh } from '@element-plus/icons-vue'
import { renderWikiMarkdown } from '@/utils/wikiMarkdown'
import wikiApi from '@/api/wiki'
import PageEditDialog from './PageEditDialog.vue'

// 页面类型的展示文案与标签配色（与草稿面板保持一致）
const TYPE_LABELS = { summary: '概览', entity: '实体', concept: '概念' }
const TYPE_TAGS = { summary: 'warning', entity: 'success', concept: 'primary' }

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const keyword = ref('')
const pageType = ref('')
const loading = ref(false)
const exporting = ref(false)

const editVisible = ref(false)
const editing = ref(null)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

const renderRow = (row) => {
  try {
    return renderWikiMarkdown(row.content)
  } catch {
    return row.content
  }
}

const fetchList = async () => {
  loading.value = true
  try {
    const response = await wikiApi.listPublished(page.value, pageSize, keyword.value.trim(), pageType.value)
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

.page-expand {
  padding: 4px 16px 16px 56px;

  .expand-meta {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 12px;

    .meta-item {
      font-size: 12.5px;
      color: var(--el-text-color-secondary);
    }

    .expand-actions {
      margin-left: auto;
      display: flex;
      gap: 8px;
    }
  }

  .expand-content {
    max-height: 60vh;
    overflow-y: auto;
    padding: 16px 20px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-blank);
    font-size: 14px;
    line-height: 1.8;

    h1, h2, h3 { margin: 16px 0 8px; }
    h4, h5 { margin: 12px 0 6px; }
    p { margin: 8px 0; }
    ul, ol { margin: 4px 0 8px; padding-left: 20px; }
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
}
</style>
