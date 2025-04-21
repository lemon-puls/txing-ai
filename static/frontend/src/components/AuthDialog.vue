<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  User,
  Lock,
  Message,
  Key
} from '@element-plus/icons-vue'

// 弹窗显示状态
defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

// 监听内部状态变化并同步到父组件
const closeDialog = () => {
  emit('update:modelValue', false)
}

// 当前激活标签页
const activeTab = ref('login')
// 验证码图片（模拟数据）
const captchaImage = ref('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIAAAAA...')

// 表单数据
const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  captcha: ''
})

// 记住我选项
const rememberMe = ref(false)

// 提交处理
const handleSubmit = () => {
  if (activeTab.value === 'login') {
    // 登录逻辑
    ElMessage.success('模拟登录成功')
  } else {
    // 注册逻辑
    if (form.value.password !== form.value.confirmPassword) {
      ElMessage.error('两次密码输入不一致')
      return
    }
    ElMessage.success('模拟注册成功')
  }
  closeDialog()
}
</script>

<template>
  <!-- 炫酷的弹窗容器 -->
  <div class="auth-wrapper" v-show="modelValue">
    <!-- 玻璃拟态背景 -->
    <div class="auth-backdrop" @click.self="closeDialog"></div>
    
    <!-- 弹窗主体 -->
    <Transition name="zoom">
      <div class="auth-dialog" v-show="modelValue">
        <!-- Logo 区域 -->
        <div class="logo-container">
          <div class="logo-text">
            <span class="gradient-text">Txing</span>
            <span class="ai-text">AI</span>
          </div>
          <div class="slogan">智能助手，助你远航</div>
        </div>

        <!-- 标签页切换 -->
        <el-tabs v-model="activeTab" class="modern-tabs">
          <el-tab-pane label="登录" name="login">
            <!-- 登录表单 -->
            <el-form 
              @submit.prevent="handleSubmit"
              class="modern-form"
            >
              <el-form-item>
                <el-input 
                  v-model="form.username"
                  placeholder="用户名"
                  :prefix-icon="User"
                />
              </el-form-item>

              <el-form-item>
                <el-input
                  v-model="form.password"
                  type="password"
                  placeholder="密码"
                  show-password
                  :prefix-icon="Lock"
                />
              </el-form-item>

              <div class="captcha-group">
                <el-input 
                  v-model="form.captcha"
                  placeholder="验证码"
                  :prefix-icon="Key"
                />
                <img 
                  :src="captchaImage" 
                  class="captcha-image"
                  alt="验证码"
                />
              </div>

              <div class="form-footer">
                <el-checkbox v-model="rememberMe">记住我</el-checkbox>
                <el-button type="text" class="forget-btn">忘记密码？</el-button>
              </div>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="注册" name="register">
            <!-- 注册表单 -->
            <el-form 
              @submit.prevent="handleSubmit"
              class="modern-form"
            >
              <el-form-item>
                <el-input 
                  v-model="form.username"
                  placeholder="用户名"
                  :prefix-icon="User"
                />
              </el-form-item>

              <el-form-item>
                <el-input
                  v-model="form.email"
                  placeholder="邮箱"
                  :prefix-icon="Message"
                />
              </el-form-item>

              <el-form-item>
                <el-input
                  v-model="form.password"
                  type="password"
                  placeholder="密码"
                  show-password
                  :prefix-icon="Lock"
                />
              </el-form-item>

              <el-form-item>
                <el-input
                  v-model="form.confirmPassword"
                  type="password"
                  placeholder="确认密码"
                  show-password
                  :prefix-icon="Lock"
                />
              </el-form-item>

              <div class="captcha-group">
                <el-input 
                  v-model="form.captcha"
                  placeholder="验证码"
                  :prefix-icon="Key"
                />
                <img 
                  :src="captchaImage" 
                  class="captcha-image"
                  alt="验证码"
                />
              </div>
            </el-form>
          </el-tab-pane>
        </el-tabs>

        <!-- 操作按钮 -->
        <div class="dialog-footer">
          <el-button 
            type="primary" 
            @click="handleSubmit"
            class="submit-button"
          >
            {{ activeTab === 'login' ? '登 录' : '注 册' }}
          </el-button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped lang="scss">
/* 炫酷的弹窗动画 */
.zoom-enter-active,
.zoom-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.zoom-enter-from,
.zoom-leave-to {
  opacity: 0;
  transform: scale(0.9);
}

.auth-wrapper {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.auth-backdrop {
  position: fixed;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(10px);
}

.auth-dialog {
  position: relative;
  width: 400px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 24px;
  padding: 32px;
  box-shadow: 
    0 0 30px rgba(0, 0, 0, 0.1),
    0 0 60px rgba(0, 0, 0, 0.1);
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 200px;
    background: linear-gradient(
      135deg,
      rgba(91, 228, 155, 0.1),
      rgba(0, 167, 225, 0.1)
    );
    z-index: 0;
  }
}

.logo-container {
  text-align: center;
  margin-bottom: 24px;
  position: relative;
}

.logo-text {
  font-size: 36px;
  font-weight: bold;
  margin-bottom: 8px;

  .gradient-text {
    background: linear-gradient(135deg, #5BE49B, #00A7E1);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .ai-text {
    color: #333;
    margin-left: 4px;
  }
}

.slogan {
  font-size: 14px;
  color: #666;
}

.modern-tabs {
  :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }

  :deep(.el-tabs__active-bar) {
    height: 3px;
    border-radius: 3px;
    background: linear-gradient(90deg, #5BE49B, #00A7E1);
  }

  :deep(.el-tabs__item) {
    font-size: 16px;
    color: #666;
    
    &.is-active {
      color: #00A7E1;
      font-weight: 600;
    }
  }
}

.modern-form {
  margin-top: 24px;

  :deep(.el-input__wrapper) {
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    padding: 8px 16px;
  }

  :deep(.el-input__prefix-icon) {
    font-size: 18px;
    color: #999;
  }
}

.captcha-group {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;

  .el-input {
    flex: 1;
  }
}

.captcha-image {
  height: 40px;
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.3s;

  &:hover {
    transform: scale(1.05);
  }
}

.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 16px 0;
}

.forget-btn {
  color: #00A7E1;
  
  &:hover {
    color: #5BE49B;
  }
}

.submit-button {
  width: 100%;
  height: 44px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #5BE49B, #00A7E1);
  border: none;
  
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 167, 225, 0.3);
  }
}
</style> 