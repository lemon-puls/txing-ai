<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <div class="error-code">
        <span class="number">4</span>
        <el-icon class="icon"><Search /></el-icon>
        <span class="number">4</span>
      </div>
      <h1 class="title">页面不存在</h1>
      <p class="description">您访问的页面不存在或已被移除</p>
      <div class="actions">
        <el-button type="primary" @click="goBack">
          <el-icon><Back /></el-icon>
          返回上一页
        </el-button>
        <el-button @click="goHome">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { Search, Back, HomeFilled } from '@element-plus/icons-vue'

const router = useRouter()

const goBack = () => {
  router.back()
}

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
.not-found-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.not-found-content {
  text-align: center;
}

.error-code {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 96px;
  font-weight: 700;
  line-height: 1;
  color: var(--el-color-primary);

  .icon {
    font-size: 72px;
    color: var(--el-color-primary-light-3);
  }
}

.title {
  margin: 0 0 8px;
  font-size: 24px;
  color: var(--el-text-color-primary);
}

.description {
  margin: 0 0 28px;
  color: var(--el-text-color-secondary);
}

.actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}
</style>
