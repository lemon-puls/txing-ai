<template>
  <div class="wiki-admin">
    <el-card class="wiki-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon class="title-icon"><Collection /></el-icon>
            <span>知识库管理</span>
          </div>
          <span class="header-desc">LLM Wiki：Raw 源 → AI 编译草稿 → 人工审核发布 → 面试官问答</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="wiki-tabs" @tab-change="handleTabChange">
        <el-tab-pane name="sources">
          <template #label>
            <span class="tab-label"><el-icon><Files /></el-icon> 源管理</span>
          </template>
          <SourcePanel ref="sourceRef" />
        </el-tab-pane>
        <el-tab-pane name="drafts">
          <template #label>
            <span class="tab-label"><el-icon><EditPen /></el-icon> 草稿审核</span>
          </template>
          <DraftPanel ref="draftRef" />
        </el-tab-pane>
        <el-tab-pane name="pages">
          <template #label>
            <span class="tab-label"><el-icon><Reading /></el-icon> 已发布页</span>
          </template>
          <PagePanel ref="pageRef" />
        </el-tab-pane>
        <el-tab-pane name="graph">
          <template #label>
            <span class="tab-label"><el-icon><Share /></el-icon> 知识图谱</span>
          </template>
          <GraphPanel ref="graphRef" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup name="WikiAdmin">
import { ref, nextTick } from 'vue'
import { Collection, Files, EditPen, Reading, Share } from '@element-plus/icons-vue'
import SourcePanel from './SourcePanel.vue'
import DraftPanel from './DraftPanel.vue'
import PagePanel from './PagePanel.vue'
import GraphPanel from './GraphPanel.vue'

const activeTab = ref('sources')
const sourceRef = ref(null)
const draftRef = ref(null)
const pageRef = ref(null)
const graphRef = ref(null)
// 已挂载过的 Tab：首次进入由面板 onMounted 拉取，再进入时手动刷新，
// 保证 ingest 产出草稿、确认发布后跨 Tab 立即可见
const visited = new Set(['sources'])

const refreshers = {
  sources: () => sourceRef.value?.refresh(),
  drafts: () => draftRef.value?.refresh(),
  pages: () => pageRef.value?.refresh(),
  graph: () => graphRef.value?.refresh()
}

const handleTabChange = (name) => {
  if (visited.has(name)) {
    nextTick(() => refreshers[name]?.())
  } else {
    visited.add(name)
  }
}
</script>

<style scoped lang="scss">
.wiki-admin {
  .wiki-card {
    border-radius: 18px;
    border: 1px solid var(--el-border-color-lighter);
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;

    .header-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 17px;
      font-weight: 700;
      color: var(--el-text-color-primary);

      .title-icon {
        color: var(--el-color-primary);
        font-size: 20px;
      }
    }

    .header-desc {
      font-size: 13px;
      color: var(--el-text-color-secondary);
    }
  }

  .wiki-tabs {
    :deep(.el-tabs__item) {
      font-size: 15px;
    }

    .tab-label {
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }
  }
}
</style>
