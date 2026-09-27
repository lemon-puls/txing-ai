<template>
  <div class="welcome-container">
    <!-- 顶部导航栏（沿用原结构：logo / GitHub / 用户头像） -->
    <header class="nav-header">
      <div class="nav-content">
        <div class="nav-left">
          <div class="logo">
            <span class="logo-text">Txing AI</span>
          </div>
        </div>
        <div class="nav-right">
          <a href="https://github.com/lemon-puls/txing-ai" target="_blank" class="github-link">
            <el-tooltip content="在 GitHub 上查看" placement="bottom">
              <el-icon class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg></el-icon>
            </el-tooltip>
          </a>
          <UserAvatar />
        </div>
      </div>
    </header>

    <!-- 流光极光背景（WebGL，随指针弯曲） -->
    <AuroraRibbons class="hero-bg" :speed="55" :intensity="65" :warp="45" />
    <!-- 文字可读性 scrim：柔和压暗左侧文案区，右侧留给地球 -->
    <div class="hero-scrim" aria-hidden="true"></div>

    <!-- 主要内容：左文案 + 右完整地球 -->
    <main class="hero-main">
      <div class="content">
        <p class="brand-line">TXING · AI ASSISTANT</p>
        <h1 class="main-title">Txing AI</h1>
        <p class="subtitle">智能助手，让我们都有光明的未来</p>

        <!-- 直接提问：回车带入新对话（chat 页读取 query.prompt 预填） -->
        <form class="prompt-box" @submit.prevent="goChat">
          <el-icon class="prompt-icon"><Search /></el-icon>
          <input
            v-model="prompt"
            class="prompt-input"
            type="text"
            placeholder="问问 Txing AI…"
            autocomplete="off"
            aria-label="向 Txing AI 提问"
          />
          <button type="submit" class="prompt-send" aria-label="开始对话">
            <el-icon><Promotion /></el-icon>
          </button>
        </form>

        <!-- 四个入口（与旧版目标一致，重做视觉：pill + hover 聚光） -->
        <nav ref="entriesRef" class="entries" @pointermove="trackSpot" @pointerleave="clearSpot">
          <button
            v-for="e in entries"
            :key="e.key"
            type="button"
            class="entry"
            @click="e.go()"
          >
            <el-icon class="entry-icon"><component :is="e.icon" /></el-icon>
            <span>{{ e.label }}</span>
          </button>
        </nav>
      </div>

      <!-- 点阵地球（WebGL，完整呈现：自转、可拖拽、光点标记当前位置） -->
      <!-- Dotted globe (WebGL, shown in full: rotating, draggable, with a pulsing marker at the current location) -->
      <div class="globe-side">
        <DottedGlobe class="hero-globe" :speed="28" :intensity="64" :dot-size="44" />
      </div>
    </main>
  </div>
</template>

<script setup name="WelcomePage">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ChatRound, Shop, Link, Share, Search, Promotion } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'
import AuroraRibbons from '@/components/AuroraRibbons.vue'
import DottedGlobe from '@/components/DottedGlobe.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'

// 与 chat/assistant/websites 页一致：进入页面即应用主题（setup 调用，首帧前挂 dark class 防闪白）
// Same as other pages: apply the stored theme in setup (dark class lands before first paint)
const themeStore = useThemeStore()
themeStore.initTheme()

const router = useRouter()
const prompt = ref('')

// 四个入口与旧版目标完全一致，仅重做视觉 / same four destinations as before, visuals only
const entries = [
  { key: 'chat', label: '开启聊天', icon: ChatRound, go: () => router.push('/chat') },
  { key: 'market', label: 'AI 助手市场', icon: Shop, go: () => router.push('/assistant') },
  { key: 'websites', label: '网站导航', icon: Link, go: () => router.push('/websites') },
  { key: 'workflow', label: 'AI 应用市场', icon: Share, go: () => router.push('/workflow') }
]

// 回车/发送：非空时把输入带入新对话（对象传参由 vue-router 自动编码）
// Enter/send: carry the text into a fresh chat (vue-router encodes object queries)
const goChat = () => {
  const text = prompt.value.trim()
  router.push(text
    ? { path: '/chat', query: { newChat: 'true', prompt: text } }
    : { path: '/chat', query: { newChat: 'true' } })
}

// 入口行聚光：指针坐标写入 CSS 变量，::after 单个径向渐变消费（仅 hover 设备启用）
// Entry spotlight: pointer coords go into CSS vars feeding one radial gradient (::after), hover devices only
const entriesRef = ref(null)
const canHover = window.matchMedia('(hover: hover)').matches
const trackSpot = (e) => {
  if (!canHover || !entriesRef.value) return
  const r = entriesRef.value.getBoundingClientRect()
  entriesRef.value.style.setProperty('--mx', `${e.clientX - r.left}px`)
  entriesRef.value.style.setProperty('--my', `${e.clientY - r.top}px`)
}
const clearSpot = () => {
  entriesRef.value?.style.removeProperty('--mx')
  entriesRef.value?.style.removeProperty('--my')
}
// 本页无任何 setTimeout/setInterval/requestAnimationFrame —— 旧版打字机定时器泄漏从根上消除
// No timers at all on this page — the old typewriter timer leak is gone by design
</script>

<style scoped lang="scss">
.welcome-container {
  position: relative;
  height: 100vh;
  height: 100dvh; // 移动端地址栏收展不跳动 / avoids iOS chrome jump
  overflow: hidden; // 对抗全局 #app overflow:scroll / counters the global scrollable #app
  background: var(--bg-primary);
}

// 背景层
.hero-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

// 可读性 scrim：压暗左侧文案区，右侧留给地球 / dim the copy zone on the left, keep the right for the globe
.hero-scrim {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  background: radial-gradient(
    ellipse 60% 55% at 28% 46%,
    color-mix(in srgb, var(--bg-primary) 45%, transparent),
    transparent 70%
  );
}

// 主区：左文案 + 右完整地球 / hero split: copy left, full globe right
.hero-main {
  position: relative;
  z-index: 3;
  max-width: 1200px;
  min-height: 100%;
  margin: 0 auto;
  padding: 96px 24px 48px;
  display: flex;
  align-items: center;
  gap: clamp(32px, 5vw, 72px);
}

// 点阵地球：完整呈现（可拖拽交互）/ the dotted globe shown in full (interactive, draggable)
.globe-side {
  flex-shrink: 0;
  width: min(38vw, 52vh, 540px);
  aspect-ratio: 1;
}
.hero-globe {
  width: 100%;
  height: 100%;
  animation: globe-in 1.2s cubic-bezier(0.4, 0, 0.2, 1) 0.12s both;
}

// 顶部导航（透明浮在流光上）
.nav-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 4;
  height: 64px;
}
.nav-content {
  max-width: 1200px;
  height: 100%;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.logo-text {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(45deg, #2b5eff, #1e88e5, #03a9f4);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.nav-right {
  display: flex;
  align-items: center;
  gap: 16px;
}
.github-link {
  display: inline-flex;
  color: var(--text-secondary);
  transition: color 0.2s;

  &:hover {
    color: var(--text-primary);
  }
}
.nav-icon {
  display: block;
}

// 文案列（桌面左对齐，窄屏回中）/ copy column (left-aligned on desktop, re-centered when the globe hides)
.content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
}

.brand-line {
  margin: 0 0 14px;
  font-size: 12px;
  letter-spacing: 0.3em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

.main-title {
  margin: 0;
  font-size: clamp(2.5rem, 6vw, 4rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1.1;
  background: linear-gradient(45deg, #2b5eff, #1e88e5, #03a9f4);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  margin: 14px 0 36px;
  font-size: 1.25rem;
  color: var(--text-secondary);
}

// 提问输入框（玻璃胶囊 + 品牌聚焦环）
.prompt-box {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  max-width: 560px;
  height: 52px;
  padding: 0 8px 0 18px;
  border-radius: 999px;
  border: 1px solid var(--border-color);
  background: color-mix(in srgb, var(--bg-primary) 55%, transparent);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  transition: border-color 0.2s, box-shadow 0.2s;

  &:focus-within {
    border-color: var(--el-color-primary);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--el-color-primary) 22%, transparent);
  }
}
.prompt-icon {
  font-size: 18px;
  color: var(--text-secondary);
  flex-shrink: 0;
}
.prompt-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  border: none;
  outline: none;
  background: transparent;
  font-size: 15px;
  color: var(--text-primary);

  &::placeholder {
    color: var(--text-secondary);
  }
}
.prompt-send {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--el-color-primary);
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;

  &:hover {
    box-shadow: 0 2px 10px color-mix(in srgb, var(--el-color-primary) 35%, transparent);
    transform: translateY(-1px);
  }

  &:active {
    transform: none;
  }
}

// 入口行（pill + 聚光）
.entries {
  position: relative;
  margin-top: 28px;
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 10px;
  padding: 6px;
  border-radius: 999px;

  // 聚光层：跟随 --mx/--my 的单个径向渐变，hover 时淡入
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    pointer-events: none;
    opacity: 0;
    transition: opacity 0.25s;
    background: radial-gradient(
      140px circle at var(--mx, 50%) var(--my, 50%),
      color-mix(in srgb, var(--el-color-primary) 10%, transparent),
      transparent 70%
    );
  }

  &:hover::after {
    opacity: 1;
  }
}
.entry {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 999px;
  border: 1px solid transparent;
  background: transparent;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s, background-color 0.2s, transform 0.2s, box-shadow 0.2s;

  .entry-icon {
    font-size: 16px;
  }

  &:hover {
    color: var(--text-primary);
    transform: translateY(-2px);
    border-color: color-mix(in srgb, var(--el-color-primary) 40%, transparent);
    background: color-mix(in srgb, var(--bg-primary) 60%, transparent);
    box-shadow: 0 2px 10px color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  }
}

// 入场动画（CSS only，首屏内容无需 IntersectionObserver）
@keyframes fade-rise {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}
@keyframes globe-in {
  from {
    opacity: 0;
    transform: translateY(28px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: none;
  }
}
.brand-line,
.main-title,
.subtitle,
.prompt-box,
.entries {
  animation: fade-rise 0.7s cubic-bezier(0.4, 0, 0.2, 1) both;
}
.brand-line { animation-delay: 0.05s; }
.main-title { animation-delay: 0.12s; }
.subtitle { animation-delay: 0.22s; }
.prompt-box { animation-delay: 0.32s; }
.entries { animation-delay: 0.42s; }

@media (prefers-reduced-motion: reduce) {
  .brand-line,
  .main-title,
  .subtitle,
  .prompt-box,
  .entries,
  .hero-globe {
    animation: none;
  }
}

// 响应式
// ≤1023：地球退场，文案回中（scrim 对准中央）/ globe bows out, copy re-centers (scrim follows)
@media (max-width: 1023px) {
  .globe-side {
    display: none;
  }
  .content {
    align-items: center;
    text-align: center;
  }
  .entries {
    justify-content: center;
  }
  .hero-scrim {
    background: radial-gradient(
      ellipse 70% 55% at 50% 46%,
      color-mix(in srgb, var(--bg-primary) 42%, transparent),
      transparent 70%
    );
  }
}
@media (max-width: 768px) {
  .hero-main {
    padding: 88px 20px 40px;
  }
  .prompt-box {
    max-width: 100%;
  }
  .entries {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    width: 100%;
    max-width: 420px;
    border-radius: 20px;

    .entry {
      justify-content: flex-start;
    }

    &::after {
      border-radius: 20px;
    }
  }
}
@media (max-width: 480px) {
  .subtitle {
    font-size: 1.05rem;
    margin-bottom: 28px;
  }
  .entries {
    gap: 8px;
  }
  .entry {
    padding: 9px 14px;
    gap: 6px;
    font-size: 13px;
  }
  .hero-scrim {
    background: radial-gradient(
      ellipse 100% 60% at 50% 46%,
      color-mix(in srgb, var(--bg-primary) 46%, transparent),
      transparent 72%
    );
  }
}
</style>
