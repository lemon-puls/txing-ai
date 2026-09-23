<template>
  <div class="ops-assistant-page">
    <el-card class="chat-card" :body-style="{ padding: 0, height: '100%', display: 'flex', flexDirection: 'column' }">
      <!-- 页面头部 -->
      <div class="card-header">
        <div class="header-left">
          <div class="header-icon">
            <el-icon :size="18"><MagicStick /></el-icon>
          </div>
          <div class="header-text">
            <div class="header-title">运营助手</div>
            <div class="header-subtitle">AI 快速录入实用网站，提案确认后才会入库</div>
          </div>
        </div>
        <el-tag effect="plain" round size="small">Beta</el-tag>
      </div>

      <OpsChatPanel :context="chatContext" style="flex: 1; min-height: 0" />
    </el-card>
  </div>
</template>

<script setup name="OpsAssistant">
import { computed } from 'vue'
import OpsChatPanel from '@/components/ops/OpsChatPanel.vue'

// 独立对话页不绑定具体管理页面，仅标记来源
const chatContext = computed(() => ({ page: 'ops-assistant' }))
</script>

<style scoped lang="scss">
.ops-assistant-page {
  height: 100%;
  padding: 16px 20px 20px;
  box-sizing: border-box;
  // 顶部淡淡的primary氛围渐变
  background: linear-gradient(180deg, var(--el-color-primary-light-9), transparent 280px);
}

.chat-card {
  height: 100%;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);

  :deep(.el-card__body) {
    height: 100%;
  }
}

.card-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-icon {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
    box-shadow: 0 4px 12px var(--el-color-primary-light-8);
  }

  .header-title {
    font-size: 16px;
    font-weight: 700;
    line-height: 1.3;
  }

  .header-subtitle {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 2px;
  }
}

// 暗色模式：顶部氛围渐变改为极淡白色（primary-light-9 在暗色下仍是浅色）
// 注意：scoped 下须用 `html.dark &` 写法（与 about/index.vue 一致），`:global(.dark)` 包裹会被编译器丢弃内部选择器
.ops-assistant-page {
  html.dark & {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.03), transparent 280px);
  }
}
</style>
