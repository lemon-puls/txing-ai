<template>
  <div class="forbidden-container">
    <div class="forbidden-content">
      <div class="error-code">
        <span class="number">4</span>
        <el-icon class="icon"><Lock /></el-icon>
        <span class="number">3</span>
      </div>
      <h1 class="title">访问被拒绝</h1>
      <p class="description">抱歉，您没有权限访问此页面</p>
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
import { Lock, Back, HomeFilled } from '@element-plus/icons-vue'

const router = useRouter()

const goBack = () => {
  router.back()
}

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
.forbidden-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.forbidden-content {
  text-align: center;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 1rem;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  animation: fadeIn 0.5s ease-out;
}

.error-code {
  font-size: 8rem;
  font-weight: bold;
  color: #409EFF;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.error-code .number {
  animation: bounce 1s ease infinite;
}

.error-code .number:last-child {
  animation-delay: 0.2s;
}

.error-code .icon {
  font-size: 6rem;
  margin: 0 1rem;
  color: #F56C6C;
  animation: shake 1s ease infinite;
}

.title {
  font-size: 2rem;
  color: #303133;
  margin-bottom: 1rem;
  font-weight: 600;
}

.description {
  font-size: 1.1rem;
  color: #606266;
  margin-bottom: 2rem;
}

.actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.actions .el-button {
  padding: 0.8rem 1.5rem;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.actions .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes shake {
  0%, 100% {
    transform: rotate(0);
  }
  25% {
    transform: rotate(-10deg);
  }
  75% {
    transform: rotate(10deg);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .error-code {
    font-size: 6rem;
  }
  
  .error-code .icon {
    font-size: 4rem;
  }
  
  .title {
    font-size: 1.5rem;
  }
  
  .description {
    font-size: 1rem;
  }
  
  .actions {
    flex-direction: column;
  }
  
  .actions .el-button {
    width: 100%;
  }
}
</style> 