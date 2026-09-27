<template>
  <div class="gradient-picker">
    <div
      v-for="g in PROJECT_GRADIENTS"
      :key="g.value"
      class="gradient-swatch"
      :class="{ active: current === g.value }"
      :style="{ background: `linear-gradient(135deg, ${g.from} 0%, ${g.to} 100%)` }"
      :title="g.label"
      @click="select(g.value)"
    >
      <el-icon v-if="current === g.value" class="check-icon"><Check /></el-icon>
    </div>
    <span class="form-tip">前台封面回退渐变色（未配置封面图时使用）</span>
  </div>
</template>

<script setup>
// 渐变色板选择器：仅 1-4，色值与前台 .project-gradient-1..4 逐字一致
// Gradient picker limited to 1-4, exact same colors as the front page
import { computed } from 'vue'
import { Check } from '@element-plus/icons-vue'
import { PROJECT_GRADIENTS, normalizeGradient } from '../../constants'

const props = defineProps({
  modelValue: { type: [String, Number], default: 1 }
})
const emit = defineEmits(['update:modelValue'])

// 历史数据可能是 number 或 "5"/"6" 等脏值，归一化展示
// legacy rows may store numbers or out-of-range values; normalize for display
const current = computed(() => normalizeGradient(props.modelValue))

const select = (v) => {
  // 统一以 string 落库，与既有接口约定一致
  // persist as string to match the existing API convention
  emit('update:modelValue', String(v))
}
</script>

<style lang="scss" scoped>
.gradient-picker {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;

  .gradient-swatch {
    position: relative;
    width: 64px;
    height: 40px;
    border-radius: 8px;
    cursor: pointer;
    border: 2px solid transparent;
    box-shadow: var(--el-box-shadow-lighter);
    transition: transform 0.15s ease, border-color 0.15s ease;

    &:hover {
      transform: translateY(-2px);
    }

    &.active {
      border-color: var(--el-color-primary);
      box-shadow: 0 0 0 2px var(--el-color-primary-light-8);
    }

    .check-icon {
      position: absolute;
      inset: 0;
      margin: auto;
      width: fit-content;
      height: fit-content;
      color: #fff;
      filter: drop-shadow(0 1px 2px rgb(0 0 0 / 40%));
    }
  }
}
</style>
