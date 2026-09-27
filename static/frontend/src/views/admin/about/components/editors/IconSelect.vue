<template>
  <el-select
    :model-value="modelValue"
    filterable
    clearable
    :placeholder="placeholder || '选择图标'"
    class="icon-select"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-option v-for="icon in AVAILABLE_ICONS" :key="icon.key" :label="icon.label" :value="icon.key">
      <div class="icon-option">
        <el-icon class="icon-option-icon">
          <component :is="resolveIcon(icon.key) || 'Picture'" />
        </el-icon>
        <span>{{ icon.label }}</span>
        <span class="icon-option-key">{{ icon.key }}</span>
      </div>
    </el-option>
  </el-select>
</template>

<script setup>
// 图标可视化选择器：封装 AVAILABLE_ICONS，option 内渲染真实图标便于挑选
// Visual icon picker wrapping AVAILABLE_ICONS with real icon previews
import { AVAILABLE_ICONS, resolveIcon } from '@/utils/iconResolver.js'

defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])
</script>

<style lang="scss" scoped>
.icon-select {
  width: 100%;
}

.icon-option {
  display: flex;
  align-items: center;
  gap: 8px;

  .icon-option-icon {
    color: var(--el-color-primary);
  }

  .icon-option-key {
    margin-left: auto;
    font-size: 11px;
    color: var(--el-text-color-placeholder);
  }
}
</style>
