<template>
  <el-drawer
    v-model="visible"
    title="主题设置"
    size="300px"
    :with-header="false"
  >
    <div class="theme-drawer">
      <div class="drawer-header">
        <h2>主题设置</h2>
        <el-icon class="close-icon" @click="handleClose"><Close /></el-icon>
      </div>
      <div class="drawer-content">
        <div class="setting-item">
          <span class="setting-label">主题色</span>
          <el-color-picker
            v-model="tempPrimaryColor"
            :predefine="[
              '#409EFF',
              '#67C23A',
              '#E6A23C',
              '#F56C6C',
              '#909399'
            ]"
            show-alpha
          />
        </div>
        <div class="setting-item">
          <span class="setting-label">深色模式</span>
          <el-switch
            v-model="isDark"
            @change="toggleTheme"
          />
        </div>
      </div>
      <div class="drawer-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleConfirm">确认</el-button>
      </div>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Close } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const themeStore = useThemeStore()
const visible = ref(props.modelValue)
const isDark = computed(() => themeStore.isDark)
const tempPrimaryColor = ref(themeStore.primaryColor)

// 监听 modelValue 变化
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    tempPrimaryColor.value = themeStore.primaryColor
  }
})

// 监听 visible 变化
watch(visible, (val) => {
  emit('update:modelValue', val)
})

// 切换主题
const toggleTheme = () => {
  themeStore.toggleTheme()
}

// 确认主题色更改
const handleConfirm = () => {
  themeStore.setPrimaryColor(tempPrimaryColor.value)
  handleClose()
}

// 关闭抽屉
const handleClose = () => {
  visible.value = false
}
</script>

<style lang="scss" scoped>
.theme-drawer {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;

  .drawer-header {
    padding: 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--el-border-color-light);

    h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
      color: var(--el-text-color-primary);
    }

    .close-icon {
      font-size: 20px;
      color: var(--el-text-color-secondary);
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        color: var(--el-color-primary);
        transform: rotate(90deg);
      }
    }
  }

  .drawer-content {
    flex: 1;
    padding: 20px;
    overflow-y: auto;

    .setting-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 24px;

      &:last-child {
        margin-bottom: 0;
      }

      .setting-label {
        font-size: 14px;
        color: var(--el-text-color-primary);
      }
    }
  }

  .drawer-footer {
    padding: 20px;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    border-top: 1px solid var(--el-border-color-light);
  }
}

:deep(.el-drawer__body) {
  padding: 0;
}
</style> 