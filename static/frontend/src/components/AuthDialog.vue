<script setup>
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  User,
  Lock,
  Message,
  Key
} from '@element-plus/icons-vue'
import { defaultApi } from "@/api/index.js";
import { useUserStore } from '@/stores/user'

// 获取 user store
const userStore = useUserStore()

// 弹窗显示状态
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'login-success'])

// 表单数据
const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  captcha: ''
})

// 登录表单校验规则
const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 4, message: '验证码长度为 4 位', trigger: 'blur' }
  ]
}

// 注册表单校验规则
const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== form.value.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 4, message: '验证码长度为 4 位', trigger: 'blur' }
  ]
}

// 计算当前使用的校验规则
const rules = computed(() => {
  return activeTab.value === 'login' ? loginRules : registerRules
})

// 表单引用
const formRef = ref(null)

// 加载状态
const loading = ref(false)

// 记住我选项
const rememberMe = ref(false)

// 当前激活标签页
const activeTab = ref('login')

// 验证码相关数据
const captchaImage = ref('')
const captchaId = ref('')

// 获取验证码
const getCaptcha = async () => {
  try {
    const res = await defaultApi.apiCaptchaGet()
    if (res.code === 0) {
      captchaImage.value = res.data.image
      captchaId.value = res.data.id
    } else {
      ElMessage.error(res.msg || '获取验证码失败')
    }
  } catch (error) {
    console.error('获取验证码失败:', error)
    ElMessage.error('获取验证码失败')
  }
}

// 监听弹窗显示状态，显示时获取验证码
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    getCaptcha()
    resetForm()
  }
})

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  form.value = {
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    captcha: ''
  }
}

// 关闭弹窗
const closeDialog = () => {
  emit('update:modelValue', false)
  resetForm()
}

// 提交处理
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    // 表单验证
    await formRef.value.validate()

    loading.value = true

    if (activeTab.value === 'login') {
      // 登录请求
      const loginData = {
        username: form.value.username,
        password: form.value.password,
        captchaId: captchaId.value,
        captcha: form.value.captcha,
        rememberMe: rememberMe.value
      }

      const res = await defaultApi.apiUserLoginPost(loginData)
      if (res.code === 0) {
        // 保存用户信息和 token
        userStore.login(res.data)
        ElMessage.success('登录成功')
        emit('login-success', res.data)
        closeDialog()
      } else {
        ElMessage.error(res.msg || '登录失败')
        getCaptcha() // 刷新验证码
      }
    } else {
      // 注册请求
      const registerData = {
        username: form.value.username,
        email: form.value.email,
        password: form.value.password,
        captchaId: captchaId.value,
        captcha: form.value.captcha
      }

      const res = await defaultApi.apiUserRegisterPost(registerData)
      if (res.code === 0) {
        ElMessage.success('注册成功，请登录')
        activeTab.value = 'login'
        resetForm()
        getCaptcha()
      } else {
        ElMessage.error(res.msg || '注册失败')
        getCaptcha() // 刷新验证码
      }
    }
  } catch (error) {
    console.error('表单提交失败:', error)
    if (error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <!-- 使用 Teleport 将弹窗内容传送到 body 元素上 -->
  <Teleport to="body">
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
                ref="formRef"
                :model="form"
                :rules="rules"
                @submit.prevent="handleSubmit"
                class="modern-form"
              >
                <el-form-item prop="username">
                  <el-input
                    v-model="form.username"
                    placeholder="用户名"
                    :prefix-icon="User"
                  />
                </el-form-item>

                <el-form-item prop="password">
                  <el-input
                    v-model="form.password"
                    type="password"
                    placeholder="密码"
                    show-password
                    :prefix-icon="Lock"
                  />
                </el-form-item>

                <div class="captcha-group">
                  <el-form-item prop="captcha" style="margin-bottom: 0; flex: 1;">
                    <el-input
                      v-model="form.captcha"
                      placeholder="验证码"
                      :prefix-icon="Key"
                    />
                  </el-form-item>
                  <img
                    :src="captchaImage"
                    class="captcha-image"
                    alt="验证码"
                    @click="getCaptcha"
                    title="点击刷新验证码"
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
                ref="formRef"
                :model="form"
                :rules="rules"
                @submit.prevent="handleSubmit"
                class="modern-form"
              >
                <el-form-item prop="username">
                  <el-input
                    v-model="form.username"
                    placeholder="用户名"
                    :prefix-icon="User"
                  />
                </el-form-item>

                <el-form-item prop="email">
                  <el-input
                    v-model="form.email"
                    placeholder="邮箱"
                    :prefix-icon="Message"
                  />
                </el-form-item>

                <el-form-item prop="password">
                  <el-input
                    v-model="form.password"
                    type="password"
                    placeholder="密码"
                    show-password
                    :prefix-icon="Lock"
                  />
                </el-form-item>

                <el-form-item prop="confirmPassword">
                  <el-input
                    v-model="form.confirmPassword"
                    type="password"
                    placeholder="确认密码"
                    show-password
                    :prefix-icon="Lock"
                  />
                </el-form-item>

                <div class="captcha-group">
                  <el-form-item prop="captcha" style="margin-bottom: 0; flex: 1;">
                    <el-input
                      v-model="form.captcha"
                      placeholder="验证码"
                      :prefix-icon="Key"
                    />
                  </el-form-item>
                  <img
                    :src="captchaImage"
                    class="captcha-image"
                    alt="验证码"
                    @click="getCaptcha"
                    title="点击刷新验证码"
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
              :loading="loading"
            >
              {{ activeTab === 'login' ? '登 录' : '注 册' }}
            </el-button>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<style lang="scss">
.auth-wrapper {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.auth-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(5px);
  z-index: -1;
}

.auth-dialog {
  width: 400px;
  max-width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  background: var(--el-bg-color);
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  padding: 24px;
  position: relative;
  z-index: 2;
}

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
    opacity: 0.8;
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

/* 全局样式，确保弹窗层级和位置 */
.auth-dialog {
  .el-dialog {
    margin-top: 15vh !important;
  }

  .el-overlay {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 2000 !important;
    background-color: rgba(0, 0, 0, 0.5);
    overflow: auto;
  }
}
</style>
