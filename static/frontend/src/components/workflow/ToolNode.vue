<template>
  <div class="tool-node">
    <!-- 输入连接点 - 左侧 -->
    <Handle type="target" :position="Position.Left" id="input-left" />
    <!-- 输入连接点 - 右侧 -->
    <Handle type="target" :position="Position.Right" id="input-right" />

    <div class="node-header">
      <div class="node-icon">
        <el-icon><Tools /></el-icon>
      </div>
      <div class="node-title">{{ label }}</div>
    </div>

    <div class="node-content">
      <div class="tools-list" v-if="toolConfig?.tools?.length">
        <el-tag
          v-for="tool in toolConfig.tools"
          :key="tool"
          size="small"
          type="warning"
          class="tool-tag"
        >
          {{ getToolDisplayName(tool) }}
        </el-tag>
      </div>
      <div class="no-tools" v-else>
        点击配置工具
      </div>
    </div>

    <!-- 输出连接点 - 右侧 -->
    <Handle type="source" :position="Position.Right" id="output-right" />
    <!-- 输出连接点 - 左侧 -->
    <Handle type="source" :position="Position.Left" id="output-left" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Tools } from '@element-plus/icons-vue'

const props = defineProps({
  id: String,
  data: {
    type: Object,
    default: () => ({})
  }
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
.tool-node {
  padding: 0;
  background: white;
  border: 2px solid #ff9800;
  border-radius: 12px;
  min-width: 180px;
  box-shadow: 0 2px 8px rgba(255, 152, 0, 0.15);
  transition: all 0.2s ease;
  overflow: hidden;

  &:hover {
    box-shadow: 0 4px 16px rgba(255, 152, 0, 0.25);
    transform: translateY(-2px);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: linear-gradient(135deg, #fff3e0 0%, #ffe0b2 100%);
    border-bottom: 1px solid #ffcc80;

    .node-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #ff9800;
      border-radius: 6px;
      color: white;
      font-size: 14px;
    }

    .node-title {
      font-size: 14px;
      font-weight: 600;
      color: #e65100;
    }
  }

  .node-content {
    padding: 10px 12px;

    .tools-list {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    .tool-tag {
      font-size: 11px;
    }

    .no-tools {
      font-size: 12px;
      color: #bdbdbd;
      text-align: center;
      padding: 4px 0;
    }
  }
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  background: #ff9800;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  &:hover {
    transform: scale(1.2);
  }
}
</style>
