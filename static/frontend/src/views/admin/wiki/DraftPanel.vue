<template>
  <div class="draft-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <span class="panel-label">
          待审核草稿
          <span v-if="total > 0" class="count-chip">{{ total > 99 ? '99+' : total }}</span>
        </span>
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
      </div>
      <div class="toolbar-right">
        <el-button round type="success" :disabled="total === 0" @click="confirmAll">
          <el-icon><Select /></el-icon>
          <span>全部确认</span>
        </el-button>
        <el-button round circle :loading="loading" @click="fetchList">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 草稿列表（展开行直接查看全文并完成审核动作，无需进编辑弹窗） -->
    <el-table :data="list" row-key="id" v-loading="loading" class="rounded-table">
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="draft-expand">
            <div class="expand-meta">
              <el-tag :type="row.targetPageId ? 'warning' : 'success'" round effect="light">
                {{ row.targetPageId ? '更新既有页' : '新建页面' }}
              </el-tag>
              <span v-if="row.aliases && row.aliases.length" class="meta-item">
                别名：{{ row.aliases.join(' / ') }}
              </span>
              <span class="meta-item">v{{ row.version }}</span>
              <span class="meta-item">{{ formatTime(row.createTime) }}</span>
              <div class="expand-actions">
                <el-button round size="small" @click="edit(row)">编辑</el-button>
                <el-button round size="small" type="success" plain @click="confirmOne(row)">确认发布</el-button>
                <el-button round size="small" type="danger" plain @click="remove(row)">丢弃</el-button>
              </div>
            </div>
            <div class="expand-content" v-html="renderRow(row)"></div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="slug" label="slug" width="260" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="slug-code">{{ row.slug }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="140" show-overflow-tooltip />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="TYPE_TAGS[row.pageType] || 'info'" round effect="light">
            {{ TYPE_LABELS[row.pageType] || row.pageType }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="summary" label="摘要" min-width="200" show-overflow-tooltip />
      <el-table-column label="变更" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.targetPageId" type="warning" round effect="light">更新</el-tag>
          <el-tag v-else type="success" round effect="light">新建</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="v" width="60" />
      <el-table-column prop="createTime" label="生成时间" width="170">
        <template #default="{ row }">{{ formatTime(row.createTime) }}</template>
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
      is-draft
      @saved="fetchList"
    />
  </div>
</template>

<script setup name="WikiDraftPanel">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Select, Refresh, Search } from '@element-plus/icons-vue'
import { renderWikiMarkdown } from '@/utils/wikiMarkdown'
import wikiApi from '@/api/wiki'
import PageEditDialog from './PageEditDialog.vue'

// 页面类型的展示文案与标签配色（草稿/已发布两面板共用同一约定）
const TYPE_LABELS = { summary: '概览', entity: '实体', concept: '概念' }
const TYPE_TAGS = { summary: 'warning', entity: 'success', concept: 'primary' }

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 50
const keyword = ref('')
const pageType = ref('')
const loading = ref(false)
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
    const response = await wikiApi.listDrafts(page.value, pageSize, keyword.value.trim(), pageType.value)
    if (response.code === 0) {
      list.value = response.data.list || []
      total.value = response.data.total || 0
    } else {
      ElMessage.error(response.msg || '加载失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载草稿列表失败')
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

const confirmOne = async (row) => {
  try {
    await ElMessageBox.confirm(
      row.targetPageId
        ? `确认将「${row.title}」合并更新到已发布页 ${row.slug}？`
        : `确认发布「${row.title}」？`,
      '确认草稿', { type: 'info' }
    )
  } catch {
    return
  }
  try {
    const response = await wikiApi.confirmDraft(row.id)
    if (response.code === 0) {
      ElMessage.success('已发布')
      fetchList()
    } else {
      ElMessage.error(response.msg || '确认失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('确认失败')
  }
}

const confirmAll = async () => {
  try {
    await ElMessageBox.confirm(`确认一键发布全部 ${total.value} 条草稿？`, '全部确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    const response = await wikiApi.confirmAllDrafts()
    if (response.code === 0) {
      ElMessage.success(`已确认 ${response.data?.confirmed ?? 0} 条`)
      fetchList()
    } else {
      ElMessage.error(response.msg || '批量确认失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('批量确认失败')
  }
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确认丢弃草稿「${row.title}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    const response = await wikiApi.deleteDraft(row.id)
    if (response.code === 0) {
      ElMessage.success('已丢弃')
      fetchList()
    } else {
      ElMessage.error(response.msg || '丢弃失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('丢弃失败')
  }
}

onMounted(fetchList)

// 供父组件在 Tab 切换时触发刷新
defineExpose({ refresh: fetchList })
</script>

<style scoped lang="scss">
.draft-panel {
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
    align-items: center;
    gap: 12px;
  }

  .toolbar-right {
    display: flex;
    gap: 12px;
  }

  .panel-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);

    // 内联计数 chip：替代悬浮角标（el-badge 的 sup 会压到右侧搜索框）
    .count-chip {
      min-width: 20px;
      height: 18px;
      padding: 0 6px;
      border-radius: 999px;
      background: var(--el-color-warning);
      color: #fff;
      font-size: 11.5px;
      font-weight: 600;
      line-height: 18px;
      text-align: center;
      box-sizing: border-box;
    }
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

.draft-expand {
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
    max-height: 55vh;
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
