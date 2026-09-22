<template>
  <div class="ops-assistant-page">
    <el-card class="chat-card" :body-style="{ padding: 0, height: '100%', display: 'flex' }">
      <OpsChatPanel :context="chatContext" style="flex: 1" />
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
  padding: 20px;
  box-sizing: border-box;
}

.chat-card {
  height: 100%;
  border-radius: 16px;
  overflow: hidden;

  :deep(.el-card__body) {
    height: 100%;
  }
}
</style>
