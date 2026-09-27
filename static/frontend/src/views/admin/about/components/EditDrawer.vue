<template>
  <el-drawer
    :model-value="modelValue"
    :size="size"
    :destroy-on-close="true"
    :close-on-click-modal="false"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <template #header>
      <div class="edit-drawer-header">
        <span class="header-title">{{ title }}</span>
        <span v-if="subtitle" class="header-subtitle">{{ subtitle }}</span>
      </div>
    </template>
    <div class="edit-drawer-body">
      <slot />
    </div>
    <template #footer>
      <div class="edit-drawer-footer">
        <el-button round @click="emit('update:modelValue', false)">取消</el-button>
        <el-button round type="primary" :loading="saving" @click="emit('save')">
          {{ saveText }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
// 通用编辑抽屉壳：零业务，表单体由调用方通过默认插槽提供
// Shared edit drawer shell; the form body is provided by the caller's default slot
defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  size: { type: String, default: '520px' },
  saving: { type: Boolean, default: false },
  saveText: { type: String, default: '保存' }
})
const emit = defineEmits(['update:modelValue', 'save'])
</script>

<style lang="scss" scoped>
.edit-drawer-header {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .header-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .header-subtitle {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

.edit-drawer-body {
  height: 100%;
}

.edit-drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
