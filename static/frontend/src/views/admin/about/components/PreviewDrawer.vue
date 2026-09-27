<template>
  <el-drawer
    :model-value="modelValue"
    direction="btt"
    size="100%"
    :with-header="false"
    :destroy-on-close="true"
    class="preview-drawer"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="preview-wrap">
      <div class="preview-toolbar">
        <span class="pt-title">
          <el-icon><View /></el-icon>
          前台预览 · /about
        </span>
        <span class="pt-tip">保存后自动刷新</span>
        <span class="flex-spacer" />
        <el-button round :loading="loading" @click="reload">
          <el-icon class="btn-icon"><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button round @click="openInNewTab">
          <el-icon class="btn-icon"><TopRight /></el-icon>
          新窗口打开
        </el-button>
        <el-button round type="primary" @click="emit('update:modelValue', false)">关闭</el-button>
      </div>
      <div v-loading="loading" class="preview-stage">
        <!-- iframe 整体 key 重建实现刷新（不依赖同源 reload，跨域部署同样成立） -->
        <!-- remount via :key to reload; works regardless of origin -->
        <iframe
          v-if="modelValue"
          :key="frameKey"
          :src="previewUrl"
          class="preview-frame"
          title="前台预览"
          @load="loading = false"
        />
      </div>
    </div>
  </el-drawer>
</template>

<script setup>
// 前台实时预览抽屉：全屏 iframe 渲染 /about，父级 previewToken 变化时自动重建刷新
// Full-screen live preview drawer; parent bumps previewToken to refresh the frame
import { ref, computed, watch } from 'vue'
import { View, Refresh, TopRight } from '@element-plus/icons-vue'
import { previewUrl } from '../constants'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // 父级每次保存后 +1，驱动 iframe 重建
  // parent increments after each save to force a frame rebuild
  reloadToken: { type: Number, default: 0 }
})
const emit = defineEmits(['update:modelValue'])

const loading = ref(false)
const frameKey = ref(0)

// 外部 reloadToken 变化 → 本地 frameKey 变化（抽屉内部刷新按钮也走同一通道）
// external token bumps and the in-drawer refresh button share one channel
const targetKey = computed(() => `p${props.reloadToken}`)
watch(targetKey, () => {
  if (!props.modelValue) return
  loading.value = true
  frameKey.value++
})
watch(
  () => props.modelValue,
  (v) => {
    if (v) loading.value = true
  }
)

const reload = () => {
  loading.value = true
  frameKey.value++
}

const openInNewTab = () => window.open(previewUrl, '_blank')
</script>

<style lang="scss" scoped>
.preview-wrap {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--el-bg-color-page);
}

.preview-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
  flex-shrink: 0;

  .pt-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .pt-tip {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .flex-spacer {
    flex: 1;
  }

  .btn-icon {
    margin-right: 2px;
  }
}

.preview-stage {
  flex: 1;
  min-height: 0;

  .preview-frame {
    display: block;
    width: 100%;
    height: 100%;
    border: 0;
    background: var(--el-bg-color-page);
  }
}
</style>
