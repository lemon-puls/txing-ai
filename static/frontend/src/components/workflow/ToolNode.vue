<template>
  <div class="tool-node">
    <Handle type="target" :position="Position.Left" id="input-left" />
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Tools /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="tools-list" v-if="toolConfig?.tools?.length">
        <el-tag v-for="tool in toolConfig.tools" :key="tool" size="small" type="warning" class="tool-tag">
          {{ getToolDisplayName(tool) }}
        </el-tag>
      </div>
      <div class="no-tools" v-else>点击配置工具</div>
    </div>

    <Handle type="source" :position="Position.Right" id="output-right" />
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Tools } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: { type: Object, default: () => ({}) }
})

const label = computed(() => props.data?.label || '工具')
const toolConfig = computed(() => props.data?.toolConfig || {})

const getToolDisplayName = (toolName) => {
  const nameMap = {
    'web_search_tool': '网页搜索',
    'markdown_save_tool': '保存Markdown',
    'markdown_to_pdf_file_tool': '转PDF',
    'image_download_tool': '图片下载',
    'image_search_tool': '图片搜索',
    'web_scraping_tool': '网页抓取',
    'pdf_read_tool': 'PDF读取'
  }
  return nameMap[toolName] || toolName
}
</script>

<style lang="scss" scoped>
$text-primary: #1e293b;
$text-muted: #94a3b8;

.tool-node {
  padding: 0;
  background: white;
  border: 2px solid #f59e0b;
  border-radius: 14px;
  min-width: 180px;
  box-shadow: 0 2px 12px rgba(245, 158, 11, 0.12);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(245, 158, 11, 0.2);
    transform: translateY(-1px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: linear-gradient(135deg, rgba(#f59e0b, 0.1) 0%, rgba(#f59e0b, 0.05) 100%);
    border-bottom: 1px solid rgba(#f59e0b, 0.15);

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: #f59e0b;
      border-radius: 8px;
      color: white;
      font-size: 15px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: $text-primary;
    }
  }

  .node-content {
    padding: 12px 14px;

    .tools-list {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .tool-tag {
      background: rgba(#f59e0b, 0.1);
      border-color: rgba(#f59e0b, 0.3);
      color: darken(#f59e0b, 10%);
      border-radius: 6px;
    }

    .no-tools {
      font-size: 12px;
      color: $text-muted;
      text-align: center;
      padding: 6px 0;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #f59e0b;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
