<template>
  <div class="draft-panel">
    <!-- 工具栏 -->
    <div class="panel-toolbar">
      <div class="toolbar-left">
        <el-badge :value="total" :hidden="total === 0" type="warning">
          <span class="panel-label">待审核草稿</span>
        </el-badge>
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

    <!-- 草稿列表 -->
    <el-table :data="list" v-loading="loading" class="rounded-table">
      <el-table-column prop="slug" label="slug" width="180" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="slug-code">{{ row.slug }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag round effect="light">{{ row.pageType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="summary" label="摘要" min-width="200" show-overflow-tooltip />
      <el-table-column label="变更" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.targetPageId" type="warning" round effect="light">更新</el-tag>
          <el-tag v-else type="success" round effect="light">新建</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="v" width="60" />
      <el-table-column prop="createTime" label="生成时间" width="170">
        <template #default="{ row }">{{ formatTime(row.createTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button round size="small" @click="edit(row)">编辑</el-button>
          <el-button round size="small" type="success" plain @click="confirmOne(row)">确认</el-button>
          <el-button round size="small" type="danger" plain @click="remove(row)">丢弃</el-button>
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
      is-draft
      @saved="fetchList"
    />
  </div>
</template>

<script setup name="WikiDraftPanel">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Select, Refresh } from '@element-plus/icons-vue'
import wikiApi from '@/api/wiki'
import PageEditDialog from './PageEditDialog.vue'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 50
const loading = ref(false)
const editVisible = ref(false)
const editing = ref(null)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '—')

const fetchList = async () => {
  loading.value = true
  try {
    const response = await wikiApi.listDrafts(page.value, pageSize)
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

  .toolbar-right {
    display: flex;
    gap: 12px;
  }

  .panel-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
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
</style>
