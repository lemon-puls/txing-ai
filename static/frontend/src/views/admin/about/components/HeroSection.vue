<template>
  <SectionCard
    id="section-hero"
    title="Hero 顶部"
    icon="UserFilled"
    description="前台首屏：头像、状态徽标、姓名与副标题（支持打字机效果）"
  >
    <div v-loading="saving" class="hero-layout">
      <el-form :model="form" label-width="90px" class="hero-form">
        <el-form-item label="头像文字">
          <el-input v-model="form.avatarText" maxlength="5" show-word-limit placeholder="如 T，展示在圆形头像中心" class="hero-input" />
        </el-form-item>
        <el-form-item label="状态徽标">
          <el-input v-model="form.statusText" placeholder="如 Ready for New Challenges，头像下方胶囊" class="hero-input" />
        </el-form-item>
        <el-form-item label="主标题姓名">
          <el-input v-model="form.name" placeholder="如 Txing，前台以打字机+霓虹效果展示" class="hero-input" />
        </el-form-item>
        <el-form-item label="副标题">
          <el-input v-model="form.subtitle" placeholder="姓名下方打字机逐字输出的副标题" type="textarea" :rows="2" class="hero-input" />
        </el-form-item>
        <el-form-item>
          <el-button round type="primary" :loading="saving" @click="save">保存</el-button>
          <span class="form-tip">保存后前台立即生效</span>
        </el-form-item>
      </el-form>

      <!-- 右侧实时预览卡：纯 CSS 复刻前台 Hero 视觉，随输入即时变化 -->
      <!-- live preview card: pure-CSS replica of the front hero, updates as you type -->
      <div class="hero-preview">
        <div class="hp-avatar">
          <span class="hp-ring hp-ring--1" />
          <span class="hp-ring hp-ring--2" />
          <div class="hp-avatar-circle">{{ form.avatarText || 'T' }}</div>
        </div>
        <transition name="hp-fade" mode="out-in">
          <div v-if="form.statusText" :key="form.statusText" class="hp-status">
            <span class="hp-dot" />
            {{ form.statusText }}
          </div>
        </transition>
        <div class="hp-name">你好，我是 {{ form.name || '...' }}</div>
        <div class="hp-subtitle">{{ form.subtitle || '副标题将在这里逐字打出' }}</div>
      </div>
    </div>
  </SectionCard>
</template>

<script setup>
// Hero 单例区块：表单 + 前台视觉预览
// Hero singleton section: form + front-style visual preview
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { defaultApi } from '@/api'
import SectionCard from './SectionCard.vue'

const props = defineProps({
  data: { type: Object, default: null }
})
const emit = defineEmits(['changed'])

const form = reactive({ avatarText: '', statusText: '', name: '', subtitle: '' })
const saving = ref(false)

// 父组件加载完成后回填（仅已知字段，避免带入多余属性）
// backfill from parent load (known fields only)
watch(
  () => props.data,
  (v) => {
    if (!v) return
    form.avatarText = v.avatarText || ''
    form.statusText = v.statusText || ''
    form.name = v.name || ''
    form.subtitle = v.subtitle || ''
  },
  { immediate: true }
)

const save = async () => {
  saving.value = true
  try {
    const payload = {
      avatarText: form.avatarText,
      statusText: form.statusText,
      name: form.name,
      subtitle: form.subtitle
    }
    const res = await defaultApi.apiAdminAboutHeroPut(payload)
    if (res?.code === 0) {
      ElMessage.success('保存成功')
      emit('changed')
    } else {
      ElMessage.error(res?.msg || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.body?.msg || e.message))
  } finally {
    saving.value = false
  }
}
</script>

<style lang="scss" scoped>
.hero-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(300px, 380px);
  gap: 24px;
  align-items: start;

  @media (max-width: 900px) {
    grid-template-columns: 1fr;
  }
}

.hero-form {
  .hero-input {
    max-width: 420px;
  }

  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
}

// ===== 前台视觉预览卡 =====
.hero-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 36px 24px;
  border-radius: 16px;
  color: #fff;
  text-align: center;
  background: linear-gradient(135deg, var(--el-color-primary), #6366f1);
  box-shadow: 0 10px 30px var(--el-color-primary-light-7);
  overflow: hidden;

  .hp-avatar {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;

    .hp-avatar-circle {
      width: 96px;
      height: 96px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 40px;
      font-weight: 900;
      color: #fff;
      background: rgb(255 255 255 / 18%);
      border: 2px solid rgb(255 255 255 / 35%);
      backdrop-filter: blur(6px);
    }

    // 双层扩散环，复刻前台 pulse-ring
    // double pulse rings, echoing the front hero
    .hp-ring {
      position: absolute;
      inset: 0;
      margin: auto;
      width: 96px;
      height: 96px;
      border-radius: 50%;
      border: 1px solid rgb(255 255 255 / 45%);
      animation: hp-pulse 2.4s ease-out infinite;

      &--2 {
        animation-delay: 1.2s;
      }
    }
  }

  .hp-status {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 5px 14px;
    border-radius: 999px;
    font-size: 12px;
    background: rgb(255 255 255 / 14%);
    border: 1px solid rgb(255 255 255 / 22%);
    backdrop-filter: blur(6px);

    .hp-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: #4ade80;
      box-shadow: 0 0 8px #4ade80;
    }
  }

  .hp-name {
    font-size: 24px;
    font-weight: 800;
    letter-spacing: 0.5px;
  }

  .hp-subtitle {
    font-size: 13px;
    opacity: 0.85;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@keyframes hp-pulse {
  0% {
    transform: scale(1);
    opacity: 0.7;
  }

  100% {
    transform: scale(1.55);
    opacity: 0;
  }
}

.hp-fade-enter-active,
.hp-fade-leave-active {
  transition: opacity 0.2s ease;
}

.hp-fade-enter-from,
.hp-fade-leave-to {
  opacity: 0;
}
</style>
