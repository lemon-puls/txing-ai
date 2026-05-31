<template>
  <div class="node-sidebar">
    <div class="sidebar-header">
      <h3>节点组件</h3>
    </div>

    <!-- 工作流配置 -->
    <div class="workflow-config">
      <div class="config-title">工作流配置</div>
      <div class="config-item">
        <label>默认模型</label>
        <el-select
          v-model="defaultModel"
          size="small"
          placeholder="选择默认模型"
          clearable
          @change="onModelChange"
        >
          <el-option
            v-for="model in modelList"
            :key="model.name"
            :label="model.displayName || model.name"
            :value="model.name"
          />
        </el-select>
      </div>
      <div class="config-item">
        <label>最大步数 <span class="hint">（可选，默认30）</span></label>
        <el-input-number
          v-model="maxRunSteps"
          size="small"
          :min="1"
          :max="1000"
          placeholder="留空使用默认值"
          controls-position="right"
          @change="onMaxRunStepsChange"
        />
      </div>
    </div>

    <div class="node-categories">
      <!-- 基础节点 -->
      <div class="category">
        <div class="category-title">基础</div>
        <div class="category-items">
          <div
            class="node-item start"
            draggable="true"
            @dragstart="onDragStart($event, 'start')"
          >
            <div class="item-icon">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <span>开始节点</span>
          </div>
          <div
            class="node-item end"
            draggable="true"
            @dragstart="onDragStart($event, 'end')"
          >
            <div class="item-icon">
              <el-icon><CircleClose /></el-icon>
            </div>
            <span>结束节点</span>
          </div>
        </div>
      </div>

      <!-- AI节点 -->
      <div class="category">
        <div class="category-title">AI能力</div>
        <div class="category-items">
          <div
            class="node-item llm"
            draggable="true"
            @dragstart="onDragStart($event, 'llm')"
          >
            <div class="item-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <span>大模型节点</span>
          </div>
          <div
            class="node-item tool"
            draggable="true"
            @dragstart="onDragStart($event, 'tool')"
          >
            <div class="item-icon">
              <el-icon><Tools /></el-icon>
            </div>
            <span>工具节点</span>
          </div>
        </div>
      </div>

      <!-- 逻辑节点 -->
      <div class="category">
        <div class="category-title">逻辑控制</div>
        <div class="category-items">
          <div
            class="node-item condition"
            draggable="true"
            @dragstart="onDragStart($event, 'condition')"
          >
            <div class="item-icon">
              <el-icon><Share /></el-icon>
            </div>
            <span>条件分支</span>
          </div>
        </div>
      </div>

      <!-- 集成节点 -->
      <div class="category">
        <div class="category-title">集成</div>
        <div class="category-items">
          <div
            class="node-item code"
            draggable="true"
            @dragstart="onDragStart($event, 'code')"
          >
            <div class="item-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <span>代码节点</span>
          </div>
          <div
            class="node-item http"
            draggable="true"
            @dragstart="onDragStart($event, 'http')"
          >
            <div class="item-icon">
              <el-icon><Link /></el-icon>
            </div>
            <span>HTTP 节点</span>
          </div>
        </div>
      </div>
    </div>

    <div class="sidebar-footer">
      <el-button type="primary" size="small" @click="$emit('save')" :loading="saving">
        <el-icon><Check /></el-icon>
        保存流程
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { VideoPlay, CircleClose, ChatDotRound, Tools, Share, Check, Monitor, Link } from '@element-plus/icons-vue'

const props = defineProps({
  saving: Boolean,
  modelList: {
    type: Array,
    default: () => []
  },
  workflowConfig: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['save', 'config-change'])

const defaultModel = ref(props.workflowConfig?.defaultModel || '')
const maxRunSteps = ref(props.workflowConfig?.maxRunSteps || null)

watch(() => props.workflowConfig, (newConfig) => {
  if (newConfig) {
    defaultModel.value = newConfig.defaultModel || ''
    maxRunSteps.value = newConfig.maxRunSteps || null
  }
}, { deep: true })

const onModelChange = (value) => {
  emit('config-change', {
    ...props.workflowConfig,
    defaultModel: value || ''
  })
}

const onMaxRunStepsChange = (value) => {
  emit('config-change', {
    ...props.workflowConfig,
    maxRunSteps: value || null
  })
}

const onDragStart = (event, nodeType) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/vueflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }
}
</script>

<style lang="scss" scoped>
.node-sidebar {
  width: 220px;
  height: 100%;
  background: #fafafa;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;

  .sidebar-header {
    padding: 16px 20px;
    border-bottom: 1px solid #e0e0e0;

    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
      color: #424242;
    }
  }

  .workflow-config {
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;
    background: #f0f7ff;

    .config-title {
      font-size: 12px;
      font-weight: 500;
      color: #1976d2;
      margin-bottom: 8px;
    }

    .config-item {
      margin-bottom: 8px;

      &:last-child {
        margin-bottom: 0;
      }

      label {
        display: block;
        font-size: 11px;
        color: #666;
        margin-bottom: 4px;

        .hint {
          color: #999;
          font-size: 10px;
        }
      }

      .el-select {
        width: 100%;
      }

      .el-input-number {
        width: 100%;
      }
    }
  }

  .node-categories {
    flex: 1;
    overflow-y: auto;
    padding: 12px;

    .category {
      margin-bottom: 16px;

      .category-title {
        font-size: 12px;
        font-weight: 500;
        color: #757575;
        margin-bottom: 8px;
        padding-left: 8px;
      }

      .category-items {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .node-item {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 12px 14px;
        background: white;
        border: 1px solid #e0e0e0;
        border-radius: 10px;
        cursor: grab;
        transition: all 0.2s ease;

        &:hover {
          border-color: #1976d2;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
          transform: translateY(-1px);
        }

        &:active {
          cursor: grabbing;
        }

        .item-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          border-radius: 8px;
          font-size: 16px;
        }

        span {
          font-size: 13px;
          font-weight: 500;
          color: #424242;
        }

        // 不同类型节点的颜色
        &.start {
          border-left: 3px solid #4caf50;
          .item-icon { background: #e8f5e9; color: #4caf50; }
        }
        &.end {
          border-left: 3px solid #ef5350;
          .item-icon { background: #ffebee; color: #ef5350; }
        }
        &.llm {
          border-left: 3px solid #2196f3;
          .item-icon { background: #e3f2fd; color: #2196f3; }
        }
        &.tool {
          border-left: 3px solid #ff9800;
          .item-icon { background: #fff3e0; color: #ff9800; }
        }
        &.condition {
          border-left: 3px solid #9c27b0;
          .item-icon { background: #f3e5f5; color: #9c27b0; }
        }
        &.code {
          border-left: 3px solid #607d8b;
          .item-icon { background: #eceff1; color: #607d8b; }
        }
        &.http {
          border-left: 3px solid #00bcd4;
          .item-icon { background: #e0f7fa; color: #00bcd4; }
        }
      }
    }
  }

  .sidebar-footer {
    padding: 12px 16px;
    border-top: 1px solid #e0e0e0;
    display: flex;
    justify-content: center;

    .el-button {
      width: 100%;
      border-radius: 10px;
    }
  }
}
</style>