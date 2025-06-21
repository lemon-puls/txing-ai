<template>
  <div class="page-container">
    <!-- 顶部搜索区域 -->
    <div class="search-section" :style="{ backgroundImage: `url(${bgImage})` }">
      <div class="nav-header">
        <div class="nav-content">
          <div class="nav-left">
            <div class="logo">
              <span class="logo-text">Txing AI</span>
            </div>
          </div>
          <div class="nav-right">
            <a href="https://github.com/yourusername/txing-ai" target="_blank" class="github-link">
              <el-tooltip content="在 GitHub 上查看" placement="bottom">
                <el-icon class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg></el-icon>
              </el-tooltip>
            </a>
            <UserAvatar />
          </div>
        </div>
      </div>

      <div class="profile-header">
        <h1 class="title">个人中心</h1>
        <p class="subtitle">管理您的个人信息和账户安全</p>
      </div>
    </div>

    <!-- 页面内容 -->
    <div class="main-content">
      <el-row :gutter="24">
        <!-- 左侧用户信息卡片 -->
        <el-col :xs="24" :sm="24" :md="8">
          <el-card class="user-card">
            <div class="user-header">
              <ImageUploader
                v-model="userInfo.avatar"
                :circle="true"
                :crop-width="200"
                :crop-height="200"
                placeholder="点击上传头像"
                @success="handleAvatarSuccess"
              />
              <h2 class="username">{{ userInfo.username }}</h2>
              <p class="user-role">{{ userInfo.role === 1 ? '超级管理员' : '普通用户' }}</p>
            </div>
            <el-divider />
            <div class="user-stats">
              <div class="stat-item">
                <div class="stat-icon">
                  <el-icon><User /></el-icon>
                </div>
                <div class="stat-info">
                  <h3>账号状态</h3>
                  <el-tag :type="userInfo.status === 0 ? 'success' : 'danger'" class="status-tag">
                    {{ userInfo.status === 0 ? '正常' : '禁用' }}
                  </el-tag>
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-icon">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="stat-info">
                  <h3>注册时间</h3>
                  <p>{{ formatDate(userInfo.createdAt) }}</p>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 右侧内容 -->
        <el-col :xs="24" :sm="24" :md="16">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>个人信息</span>
                <el-button type="primary" @click="startEdit" v-if="!isEditing">编辑</el-button>
                <div v-else>
                  <el-button type="success" @click="saveUserInfo">保存</el-button>
                  <el-button @click="cancelEdit">取消</el-button>
                </div>
              </div>
            </template>

            <el-form
              ref="formRef"
              :model="editForm"
              :rules="rules"
              label-width="100px"
              :disabled="!isEditing"
            >
              <el-form-item label="用户名" prop="username">
                <el-input v-model="editForm.username" />
              </el-form-item>
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="editForm.email" />
              </el-form-item>
              <el-form-item label="手机号" prop="phone">
                <el-input v-model="editForm.phone" />
              </el-form-item>
              <el-form-item label="性别" prop="gender">
                <el-radio-group v-model="editForm.gender">
                  <el-radio :label="1">男</el-radio>
                  <el-radio :label="2">女</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="年龄" prop="age">
                <el-input-number v-model="editForm.age" :min="1" :max="120" />
              </el-form-item>
            </el-form>
          </el-card>

          <el-card class="security-card">
            <template #header>
              <div class="card-header">
                <span>安全设置</span>
              </div>
            </template>

            <div class="security-items">
              <div class="security-item">
                <div class="security-info">
                  <h4>登录密码</h4>
                  <p>定期更改密码可以保护账号安全</p>
                </div>
                <div class="security-action">
                  <el-button type="primary" link @click="showChangePasswordDialog">
                    修改密码
                  </el-button>
                </div>
              </div>
              <div class="security-item">
                <div class="security-info">
                  <h4>手机绑定</h4>
                  <p>已绑定手机：{{ maskPhone(userInfo.phone) }}</p>
                </div>
                <div class="security-action">
                  <el-button type="primary" link @click="showChangePhoneDialog">
                    更换手机号
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 修改密码弹窗 -->
      <el-dialog
        v-model="changePasswordVisible"
        title="修改密码"
        width="400px"
        destroy-on-close
      >
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="100px"
        >
          <el-form-item label="原密码" prop="oldPassword">
            <el-input
              v-model="passwordForm.oldPassword"
              type="password"
              show-password
            />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input
              v-model="passwordForm.newPassword"
              type="password"
              show-password
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="passwordForm.confirmPassword"
              type="password"
              show-password
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="changePasswordVisible = false">取消</el-button>
            <el-button type="primary" @click="handleChangePassword">
              确认修改
            </el-button>
          </span>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup name="UserProfilePage">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User, Timer } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { defaultApi } from '@/api'
import bgImage from '@/assets/images/header-bg.jpg'
import ImageUploader from '@/components/common/ImageUploader.vue'

const userStore = useUserStore()
const formRef = ref(null)
const passwordFormRef = ref(null)
const isEditing = ref(false)
const changePasswordVisible = ref(false)

// 用户信息
const userInfo = reactive({
  username: '',
  email: '',
  phone: '',
  gender: 1,
  age: 18,
  avatar: '',
  role: 0,
  status: 0,
  createdAt: ''
})

// 编辑表单
const editForm = reactive({
  username: '',
  email: '',
  phone: '',
  gender: 1,
  age: 18
})

// 密码表单
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 表单校验规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ]
}

// 密码校验规则
const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能小于6位', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能小于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 获取用户信息
const getUserInfo = async () => {
  try {
    const res = await defaultApi.apiUserInfoGet()
    if (res.code === 0) {
      Object.assign(userInfo, res.data)
      Object.assign(editForm, res.data)
      // 同步到 store
      userStore.setUserInfo(res.data)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取用户信息失败')
  }
}

// 开始编辑
const startEdit = () => {
  isEditing.value = true
  Object.assign(editForm, userInfo)
}

// 取消编辑
const cancelEdit = () => {
  isEditing.value = false
  Object.assign(editForm, userInfo)
}

// 保存用户信息
const saveUserInfo = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const res = await defaultApi.apiUserProfilePut({
          email: editForm.email,
          phone: editForm.phone,
          gender: editForm.gender,
          age: editForm.age,
          avatar: editForm.avatar
        })
        if (res.code === 0) {
          ElMessage.success('保存成功')
          Object.assign(userInfo, editForm)
          // 同步到 store
          userStore.setUserInfo(editForm)
          isEditing.value = false
        }
      } catch (error) {
        console.error(error)
        ElMessage.error('保存失败')
      }
    }
  })
}

// 修改密码
const handleChangePassword = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const res = await defaultApi.apiUserPasswordPut({
          oldPassword: passwordForm.oldPassword,
          newPassword: passwordForm.newPassword
        })
        if (res.code === 0) {
          ElMessage.success('密码修改成功')
          changePasswordVisible.value = false
          // 清空表单
          passwordForm.oldPassword = ''
          passwordForm.newPassword = ''
          passwordForm.confirmPassword = ''
        }
      } catch (error) {
        console.error(error)
        ElMessage.error('密码修改失败')
      }
    }
  })
}

// 处理头像更新
const handleAvatarSuccess = async (url) => {
  try {
    const res = await defaultApi.apiUserProfilePut({
      avatar: url
    })
    if (res.code === 0) {
      userInfo.avatar = url
      // 同步到 store
      userStore.setUserInfo({
        ...userStore.userInfo,
        avatar: url
      })
      ElMessage.success('头像更新成功')
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('头像更新失败')
  }
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString()
}

// 手机号脱敏
const maskPhone = (phone) => {
  if (!phone) return ''
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

onMounted(() => {
  getUserInfo()
})
</script>

<style scoped lang="scss">
// 添加全局样式确保滚动正常
:global(html),
:global(body) {
  height: auto !important;
  overflow-y: auto !important;
}

.page-container {
  width: 100%;
  min-height: 100vh;
  background-color: var(--el-bg-color-page);
  position: static;
  overflow-y: auto;

  &::before {
    content: '';
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: -1; // 设置为低层级避免影响弹窗
    pointer-events: none;
    background:
      radial-gradient(circle at 0% 0%, rgba(var(--el-color-primary-rgb), 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 0%, rgba(var(--el-color-primary-rgb), 0.1) 0%, transparent 50%),
      radial-gradient(circle at 100% 100%, rgba(var(--el-color-primary-rgb), 0.1) 0%, transparent 50%),
      radial-gradient(circle at 0% 100%, rgba(var(--el-color-primary-rgb), 0.1) 0%, transparent 50%);
    opacity: 0.5;
    filter: blur(60px);
  }
}

.search-section {
  position: relative;
  padding: 0 0 120px;
  color: white;
  background-size: cover;
  background-position: center;
  flex-shrink: 0;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, #2B5EFF, #1E88E5);
    opacity: 0.95;
    z-index: 1;
  }

  .nav-header {
    position: relative;
    z-index: 10;
    height: 64px;
    backdrop-filter: blur(10px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .nav-content {
      max-width: 1200px;
      height: 100%;
      margin: 0 auto;
      padding: 0 24px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .nav-left {
        .logo {
          .logo-text {
            font-size: 24px;
            font-weight: bold;
            background: linear-gradient(45deg, #fff, rgba(255, 255, 255, 0.8));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
          }
        }
      }

      .nav-right {
        display: flex;
        align-items: center;
        gap: 24px;

        .github-link {
          display: flex;
          align-items: center;
          color: rgba(255, 255, 255, 0.7);
          transition: color 0.3s ease;

          &:hover {
            color: white;
          }

          .nav-icon {
            font-size: 24px;
          }
        }
      }
    }
  }

  .profile-header {
    position: relative;
    z-index: 2;
    text-align: center;
    padding: 60px 20px;

    .title {
      font-size: 3em;
      margin-bottom: 16px;
      font-weight: 800;
      background: linear-gradient(135deg, #fff 30%, rgba(255, 255, 255, 0.8) 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      text-shadow: 0 0 30px rgba(255, 255, 255, 0.3);
      animation: titleFloat 3s ease-in-out infinite;
    }

    .subtitle {
      font-size: 1.2em;
      color: rgba(255, 255, 255, 0.8);
      margin: 0;
    }
  }
}

.main-content {
  position: relative;
  z-index: 2;
  flex: 1;
  width: 100%;
  max-width: 1200px;
  margin: -60px auto 40px;
  padding: 0 24px;
}

.user-card {
  border-radius: 16px !important;
  overflow: hidden;
  border: none !important;
  background: var(--el-bg-color) !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1) !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;

  .user-header {
    position: relative;
    padding: 40px 20px !important;
    text-align: center;
    background: linear-gradient(135deg, var(--el-color-primary-light-9), var(--el-bg-color));

    .avatar-uploader {
      position: relative;
      margin-bottom: 20px;

      .el-avatar {
        border: 4px solid var(--el-bg-color);
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
        transition: all 0.3s ease;

        &:hover {
          transform: scale(1.05);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
        }
      }

      .avatar-upload-icon {
        position: absolute;
        bottom: 0;
        right: 0;
        padding: 8px;
        background: var(--el-color-primary);
        color: white;
        border-radius: 50%;
        cursor: pointer;
        transition: all 0.3s ease;

        &:hover {
          transform: scale(1.1);
          background: var(--el-color-primary-dark-2);
        }
      }
    }

    .username {
      font-size: 1.5em;
      margin: 0 0 8px;
      color: var(--el-text-color-primary);
    }

    .user-role {
      color: var(--el-text-color-secondary);
      margin: 0;
    }
  }

  .user-stats {
    padding: 20px;

    .stat-item {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 16px;
      border-radius: 12px;
      background: var(--el-bg-color-page);
      margin-bottom: 16px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateX(4px);
        background: var(--el-color-primary-light-9);
      }

      .stat-icon {
        width: 40px;
        height: 40px;
        border-radius: 10px;
        background: var(--el-color-primary-light-8);
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--el-color-primary);

        .el-icon {
          font-size: 20px;
        }
      }

      .stat-info {
        flex: 1;

        h3 {
          margin: 0 0 4px;
          font-size: 14px;
          color: var(--el-text-color-secondary);
        }

        p {
          margin: 0;
          font-size: 16px;
          color: var(--el-text-color-primary);
        }

        .status-tag {
          font-size: 14px;
          padding: 2px 12px;
          border-radius: 12px;
        }
      }
    }
  }
}

.info-card, .security-card {
  border-radius: 16px !important;
  overflow: hidden;
  border: none !important;
  background: var(--el-bg-color) !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1) !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  margin-bottom: 24px;

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15) !important;
  }

  :deep(.el-card__header) {
    padding: 20px;
    border-bottom: 1px solid var(--el-border-color-light);
    background: linear-gradient(to right, var(--el-bg-color), var(--el-bg-color-page));

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  :deep(.el-form) {
    padding: 20px;

    .el-form-item__label {
      font-weight: 500;
    }

    .el-input__wrapper,
    .el-radio__input {
      box-shadow: none !important;
      background: var(--el-bg-color-page);
      border-radius: 8px;
      transition: all 0.3s ease;

      &:hover, &:focus-within {
        box-shadow: 0 0 0 1px var(--el-color-primary) !important;
        transform: translateY(-2px);
      }
    }
  }
}

.security-card {
  .security-items {
    padding: 20px;

    .security-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 24px;
      border-radius: 12px;
      background: var(--el-bg-color-page);
      margin-bottom: 16px;
      transition: all 0.3s ease;

      &:hover {
        background: var(--el-color-primary-light-9);
      }

      .security-info {
        flex: 1;

        h4 {
          margin: 0 0 8px;
          font-size: 16px;
          font-weight: 500;
          color: var(--el-text-color-primary);
        }

        p {
          margin: 0;
          font-size: 14px;
          color: var(--el-text-color-secondary);
        }
      }

      .security-action {
        margin-left: 24px;

        .el-button {
          font-size: 14px;
          color: var(--el-color-primary);

          &:hover {
            color: var(--el-color-primary-light-3);
          }
        }
      }
    }

    .el-divider {
      margin: 0;
    }
  }
}

@keyframes titleFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes backgroundShift {
  0% {
    transform: scale(1) rotate(0deg);
  }
  50% {
    transform: scale(1.2) rotate(5deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}

// 响应式布局
@media (max-width: 768px) {
  .search-section {
    padding-bottom: 80px;

    .profile-header {
      padding: 40px 20px;

      .title {
        font-size: 2em;
      }
    }
  }

  .main-content {
    margin-top: -40px;
    padding: 0 16px;
  }

  .user-card {
    margin-bottom: 24px;
  }
}
</style>
