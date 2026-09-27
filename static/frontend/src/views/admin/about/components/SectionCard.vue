<template>
  <section :id="id" class="about-section">
    <el-card class="section-card" shadow="never">
      <div class="section-head">
        <div class="section-title">
          <span class="section-icon">
            <el-icon :size="15"><component :is="resolveIcon(icon) || 'Document'" /></el-icon>
          </span>
          <span class="title-text">{{ title }}</span>
          <span v-if="count !== null" class="total-badge">共 {{ count }} 个</span>
        </div>
        <div class="section-actions">
          <slot name="actions" />
        </div>
      </div>
      <slot />
      <slot v-if="count === 0" name="empty">
        <el-empty :description="`还没有${title}内容`" :image-size="90" />
      </slot>
    </el-card>
  </section>
</template>

<script setup>
// 通用分区卡片：锚点挂载 + 渐变图标标题行 + 数量徽标 + 空态
// Shared section card: anchor + gradient icon title + count badge + empty state
import { resolveIcon } from '@/utils/iconResolver.js'

defineProps({
  id: { type: String, required: true },
  title: { type: String, required: true },
  icon: { type: String, default: 'Document' },
  description: { type: String, default: '' },
  // null 表示单例区块（Hero/Contact）不显示数量
  // null hides the badge (singleton sections)
  count: { type: Number, default: null }
})
</script>

<style lang="scss" scoped>
.about-section {
  // 锚点滚动预留顶部间距（sticky 导航高度余量）
  // scroll margin so anchored sections land below the sticky header
  scroll-margin-top: 16px;
  // 分区间距由父级 .sections 的 flex gap 统一控制（抽屉占位节点会使相邻选择器失效）
  // section rhythm is owned by the parent's flex gap (drawer placeholders break sibling selectors)
}

.section-card {
  border-radius: 14px;

  :deep(.el-card__body) {
    padding: 16px;
  }
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 14px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;

  .section-icon {
    width: 30px;
    height: 30px;
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
    box-shadow: 0 3px 8px var(--el-color-primary-light-8);
    flex-shrink: 0;
  }

  .title-text {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .total-badge {
    margin-left: 2px;
    font-size: 12px;
    font-weight: 400;
    color: var(--el-text-color-secondary);
  }
}
</style>
