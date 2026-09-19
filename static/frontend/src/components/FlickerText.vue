<template>
  <!-- FlickerText：霓虹灯管式文字闪烁（Originkit Flicker Text 的 Vue 移植，仅文字模式） -->
  <!-- FlickerText: neon-tube text flicker (Vue port of the Originkit Flicker Text, text mode only) -->
  <component
    :is="tag"
    ref="elRef"
    class="flicker-text"
    :style="containerStyle"
    :aria-label="showLetterSpans ? text : undefined"
    @mouseenter="handleMouseEnter"
  >
    <template v-if="showLetterSpans">
      <span
        v-for="(ch, i) in chars"
        :key="i"
        :style="flickerLetters.has(i) ? flickerLetterStyle : undefined"
      >{{ ch }}</span>
    </template>
    <template v-else>{{ text }}</template>
  </component>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

// ===== 缓动引擎（与 Originkit 源码一致） =====
// Easing engine (kept identical to the Originkit source)
function cubicBezier(x1, y1, x2, y2) {
  const cx = 3 * x1
  const bx = 3 * (x2 - x1) - cx
  const ax = 1 - cx - bx
  const cy = 3 * y1
  const by = 3 * (y2 - y1) - cy
  const ay = 1 - cy - by
  const sampleX = (t) => ((ax * t + bx) * t + cx) * t
  const sampleY = (t) => ((ay * t + by) * t + cy) * t
  const sampleDX = (t) => (3 * ax * t + 2 * bx) * t + cx
  return (x) => {
    let t = x
    for (let i = 0; i < 8; i++) {
      const dx = sampleX(t) - x
      const d = sampleDX(t)
      if (Math.abs(dx) < 1e-6) break
      if (d === 0) break
      t -= dx / d
    }
    return sampleY(Math.max(0, Math.min(1, t)))
  }
}

function makeEaseFn(ease) {
  if (Array.isArray(ease) && ease.length === 4) return cubicBezier(ease[0], ease[1], ease[2], ease[3])
  switch (ease) {
    case 'linear': return (t) => t
    case 'easeIn': return (t) => t * t
    case 'easeOut': return (t) => 1 - (1 - t) * (1 - t)
    case 'easeInOut': return (t) => (t < 0.5 ? 2 * t * t : 1 - 2 * (1 - t) * (1 - t))
    case 'circIn': return (t) => 1 - Math.sqrt(1 - t * t)
    case 'circOut': return (t) => Math.sqrt(1 - (t - 1) * (t - 1))
    case 'circInOut':
      return (t) =>
        t < 0.5
          ? (1 - Math.sqrt(1 - 4 * t * t)) / 2
          : (Math.sqrt(1 - (-2 * t + 2) * (-2 * t + 2)) + 1) / 2
    case 'backIn': return (t) => 2.70158 * t * t * t - 1.70158 * t * t
    case 'backOut': return (t) => 1 + 2.70158 * ((t - 1) ** 3) + 1.70158 * ((t - 1) ** 2)
    default: return (t) => t
  }
}

// 闪烁配置归一化（缺省值与 Originkit buildTextCfg 相同）
// Normalize a flicker config (fallbacks match Originkit's buildTextCfg)
function buildTextCfg(m) {
  return {
    duration: m?.ease?.duration ?? 5,
    easeCurve: m?.ease?.ease ?? 'linear',
    flickerCount: m?.flickerCount ?? 4,
    showStroke: m?.showStroke ?? true,
    strokePosition: m?.strokePosition ?? 'start',
    strokeCount: m?.strokeCount ?? 1,
    strokeColor: m?.strokeColor ?? '#ffffff',
    strokeWidth: m?.strokeWidth ?? 1.5,
    restState: m?.restState ?? 'filled',
    delay: m?.delay ?? 0,
    shakeEnabled: m?.shakeEnabled ?? false,
    shakeWidth: m?.shakeWidth ?? 10,
    shakeSpeed: m?.shakeSpeed ?? 10,
    wordFlickerEnabled: m?.wordFlickerEnabled ?? true,
    letterFlickerEnabled: m?.letterFlickerEnabled ?? false,
    letterFlickerMode: m?.letterFlickerMode ?? 'stroke',
    letterFlickerIntensity: m?.letterFlickerIntensity ?? 10,
    letterFlickerOpacity: m?.letterFlickerOpacity ?? 30,
    replay: m?.replay ?? 'no',
    position: m?.position ?? 'above'
  }
}

// Originkit 文字模式预设（入场 / 悬停），props 传入的配置合并其上
// Originkit text-mode presets (enter / hover); prop configs merge over these
const ENTER_PRESET = {
  position: 'above',
  replay: 'yes',
  restState: 'filled',
  delay: 0,
  ease: { duration: 2, ease: 'easeInOut' },
  flickerCount: 10,
  showStroke: false,
  strokePosition: 'start',
  strokeCount: 1,
  strokeColor: '#ffffff',
  strokeWidth: 1.5,
  wordFlickerEnabled: false,
  shakeEnabled: false,
  shakeWidth: 10,
  shakeSpeed: 10,
  letterFlickerEnabled: true,
  letterFlickerMode: 'opacity',
  letterFlickerOpacity: 30,
  letterFlickerIntensity: 10
}
const HOVER_PRESET = {
  ease: { duration: 2, ease: 'easeInOut' },
  flickerCount: 3,
  showStroke: false,
  strokePosition: 'start',
  strokeCount: 1,
  strokeColor: '#ffffff',
  strokeWidth: 1.5,
  wordFlickerEnabled: true,
  shakeEnabled: false,
  shakeWidth: 10,
  shakeSpeed: 10,
  letterFlickerEnabled: true,
  letterFlickerMode: 'opacity',
  letterFlickerOpacity: 30,
  letterFlickerIntensity: 10
}

// 按缓动曲线把 N 个阶段铺满总时长（阶段间隔即时间轴）
// Distribute N phases across the total duration following the ease curve
function generateTimings(count, totalMs, easeCurve) {
  const fn = makeEaseFn(easeCurve)
  const intervals = []
  let prev = 0
  for (let i = 1; i <= count; i++) {
    const cur = fn(i / count) * totalMs
    intervals.push(Math.max(0, cur - prev))
    prev = cur
  }
  return intervals
}

// 闪烁序列里的实填/描边排布（strokeCount 个描边帧插入 flickerCount 帧中）
// Outline/filled layout of the flicker sequence (strokeCount outline frames within flickerCount frames)
function buildVisibleItems(cfg) {
  const sc = Math.min(cfg.strokeCount ?? 1, cfg.flickerCount)
  if (!cfg.showStroke) return Array(cfg.flickerCount).fill('filled')
  const fillCount = Math.max(1, cfg.flickerCount - sc)
  const strokes = Array(sc).fill('outline')
  const pos = cfg.strokePosition ?? 'start'
  if (pos === 'start') return [...strokes, ...Array(fillCount).fill('filled')]
  if (pos === 'end') return [...Array(fillCount - 1).fill('filled'), ...strokes, 'filled']
  const before = Math.floor(fillCount / 2)
  const after = fillCount - before
  return [...Array(before).fill('filled'), ...strokes, ...Array(after).fill('filled')]
}

// ===== Props =====
const props = defineProps({
  text: { type: String, default: 'Flicker Text' },
  tag: { type: String, default: 'span' },        // 渲染标签 h1-h6 / p / span
  colorMode: { type: String, default: 'solid' }, // solid | gradient
  fontColor: { type: String, default: '#FFFFFF' },
  gradientStart: { type: String, default: '#ffffff' },
  gradientEnd: { type: String, default: '#888888' },
  gradientAngle: { type: Number, default: 90 },
  enterEnabled: { type: Boolean, default: true },  // 进入视口时闪烁 / flicker on viewport enter
  hoverEnabled: { type: Boolean, default: true },  // 悬停时闪烁 / flicker on hover
  flicker: { type: Object, default: () => ({}) },       // 入场配置（合并 ENTER_PRESET）
  flickerHover: { type: Object, default: () => ({}) }   // 悬停配置（合并 HOVER_PRESET）
})

const enterCfg = computed(() => buildTextCfg({ ...ENTER_PRESET, ...props.flicker }))
const hoverCfg = computed(() => buildTextCfg({ ...HOVER_PRESET, ...props.flickerHover }))
const chars = computed(() => props.text.split(''))

// ===== 运行态（先于 watch 声明，避免 TDZ） =====
// Runtime state (declared before the watchers to avoid TDZ)
const elRef = ref(null)
const activeCfg = ref(enterCfg.value)
const currentPhase = ref(enterCfg.value.restState)
const moveX = ref(0)
const flickerLetters = ref(new Set())
let timers = []
let hasPlayed = false
let enterDone = !props.enterEnabled
let io = null

const clearTimers = () => {
  timers.forEach(clearTimeout)
  timers = []
}

// ===== 闪烁主流程（与 Originkit runAnimation 一致） =====
// Core flicker flow (same as Originkit's runAnimation)
const runAnimation = (cfg) => {
  clearTimers()
  flickerLetters.value = new Set()
  activeCfg.value = cfg
  if (!cfg.wordFlickerEnabled && !cfg.letterFlickerEnabled) return
  const totalMs = cfg.duration * 1000
  const nonSpaceIndices = props.text.split('').reduce((acc, c, i) => {
    if (c.trim() !== '') acc.push(i)
    return acc
  }, [])

  // 逐字母闪烁：在可见相位窗口内按周期随机点亮 1-2 个字母
  // Letter flicker: light 1-2 random letters per cycle inside visible phase windows
  const scheduleTicks = (windowStart, windowDuration) => {
    if (!cfg.letterFlickerEnabled || nonSpaceIndices.length === 0) return
    const cycleDuration = Math.round(1000 * Math.pow(50 / 1000, (cfg.letterFlickerIntensity - 1) / 19))
    const sub1 = Math.round(cycleDuration / 3)
    const sub2 = Math.round((2 * cycleDuration) / 3)
    const windowEnd = windowStart + windowDuration
    let tickCursor = windowStart
    while (tickCursor < windowEnd) {
      const tFlicker1 = tickCursor
      const tFill = tickCursor + sub1
      const tFlicker2 = tickCursor + sub2
      let sel = new Set()
      timers.push(setTimeout(() => {
        const count = Math.min(nonSpaceIndices.length, Math.floor(Math.random() * 2) + 1)
        const shuffled = [...nonSpaceIndices].sort(() => Math.random() - 0.5)
        sel = new Set(shuffled.slice(0, count))
        flickerLetters.value = sel
      }, tFlicker1))
      if (tFill < windowEnd) {
        timers.push(setTimeout(() => { flickerLetters.value = new Set() }, tFill))
      }
      if (tFlicker2 < windowEnd) {
        timers.push(setTimeout(() => { flickerLetters.value = sel }, tFlicker2))
      }
      tickCursor += cycleDuration
    }
    timers.push(setTimeout(() => { flickerLetters.value = new Set() }, windowEnd))
  }

  if (cfg.wordFlickerEnabled) {
    currentPhase.value = cfg.restState
    moveX.value = 0
    const visibleItems = buildVisibleItems(cfg)
    const sequence = []
    visibleItems.forEach((item) => {
      sequence.push('invisible')
      sequence.push(item)
    })
    const intervals = generateTimings(sequence.length, totalMs, cfg.easeCurve)
    const phaseSlots = []
    let cursor = cfg.delay * 1000
    sequence.forEach((phase, i) => {
      const startMs = cursor
      const durationMs = intervals[i] ?? 0
      phaseSlots.push({ phase, startMs, durationMs })
      timers.push(setTimeout(() => { currentPhase.value = phase }, startMs))
      cursor += durationMs
    })
    timers.push(setTimeout(() => {
      currentPhase.value = cfg.restState
      moveX.value = 0
      flickerLetters.value = new Set()
    }, cursor))
    if (cfg.shakeEnabled) {
      const flipMs = Math.round(500 * Math.pow(30 / 500, (cfg.shakeSpeed - 1) / 19))
      const animEnd = cursor
      let flipCursor = cfg.delay * 1000
      let dir = 1
      while (flipCursor < animEnd) {
        const t = flipCursor
        const d = dir
        timers.push(setTimeout(() => { moveX.value = d * cfg.shakeWidth }, t))
        dir *= -1
        flipCursor += flipMs
      }
    }
    phaseSlots.forEach(({ phase, startMs, durationMs }) => {
      if (phase !== 'filled' && phase !== 'outline') return
      scheduleTicks(startMs, durationMs)
    })
  } else {
    scheduleTicks(cfg.delay * 1000, totalMs)
  }
}

// 进入视口触发；replay=yes 时离开视口复位、再次进入重放
// Trigger on viewport enter; with replay=yes, reset on exit and replay on re-enter
const getThreshold = (amount) => (amount === 'middle' ? 0.5 : amount === 'below' ? 1.0 : 0)

const setupObserver = () => {
  if (io) {
    io.disconnect()
    io = null
  }
  if (!props.enterEnabled || !elRef.value) return
  const threshold = getThreshold(enterCfg.value.position)
  io = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        if (!hasPlayed) {
          hasPlayed = true
          enterDone = false
          runAnimation(enterCfg.value)
          timers.push(setTimeout(() => {
            enterDone = true
          }, (enterCfg.value.delay + enterCfg.value.duration) * 1000))
        }
      } else if (enterCfg.value.replay === 'yes') {
        hasPlayed = false
        enterDone = false
        clearTimers()
        flickerLetters.value = new Set()
        currentPhase.value = enterCfg.value.restState
        moveX.value = 0
      }
    })
  }, { threshold })
  io.observe(elRef.value)
}

const handleMouseEnter = () => {
  if (!props.hoverEnabled) return
  if (props.enterEnabled && !enterDone) return
  runAnimation(hoverCfg.value)
}

// 配置变更时整体复位（对应 React 版的 sig effect）
// Full reset when the config changes (mirrors the React version's sig effect)
const sig = computed(() => JSON.stringify({
  text: props.text,
  colorMode: props.colorMode,
  fontColor: props.fontColor,
  gradientStart: props.gradientStart,
  gradientEnd: props.gradientEnd,
  enterEnabled: props.enterEnabled,
  hoverEnabled: props.hoverEnabled,
  enterCfg: enterCfg.value,
  hoverCfg: hoverCfg.value
}))
watch(sig, () => {
  clearTimers()
  hasPlayed = false
  enterDone = !props.enterEnabled
  flickerLetters.value = new Set()
  const baseCfg = props.enterEnabled ? enterCfg.value : props.hoverEnabled ? hoverCfg.value : enterCfg.value
  activeCfg.value = baseCfg
  currentPhase.value = baseCfg.restState
  moveX.value = 0
  setupObserver()
})

// ===== 样式（相位驱动，全部内联，与 Originkit 相同） =====
// Styles (phase-driven, all inline, same as Originkit)
const filledStyle = computed(() => {
  if (props.colorMode === 'gradient') {
    return {
      background: `linear-gradient(${props.gradientAngle}deg, ${props.gradientStart}, ${props.gradientEnd})`,
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      backgroundClip: 'text',
      WebkitTextStroke: '0px transparent',
      color: 'transparent'
    }
  }
  return {
    color: props.fontColor,
    WebkitTextFillColor: props.fontColor,
    WebkitTextStroke: '0px transparent',
    background: 'none'
  }
})

const textStyle = computed(() => {
  switch (currentPhase.value) {
    case 'outline':
      return {
        color: 'transparent',
        WebkitTextFillColor: 'transparent',
        WebkitTextStroke: `${activeCfg.value.strokeWidth}px ${activeCfg.value.strokeColor}`,
        background: 'none'
      }
    case 'filled':
      return filledStyle.value
    default: // invisible / 其他一律透明 / everything else transparent
      return {
        color: 'transparent',
        WebkitTextFillColor: 'transparent',
        WebkitTextStroke: '0px transparent',
        background: 'none'
      }
  }
})

const containerStyle = computed(() => ({
  transform: `translateX(${moveX.value}px)`,
  transition: 'none',
  cursor: props.hoverEnabled ? 'default' : undefined,
  ...textStyle.value
}))

const showLetterSpans = computed(() => {
  const phase = currentPhase.value
  return activeCfg.value.letterFlickerEnabled &&
    (phase === 'filled' || phase === 'outline') &&
    flickerLetters.value.size > 0
})

const flickerLetterStyle = computed(() => {
  const cfg = activeCfg.value
  if (cfg.letterFlickerMode === 'stroke') {
    if (currentPhase.value === 'outline') {
      return {
        opacity: 0,
        WebkitTextFillColor: 'transparent',
        color: 'transparent',
        WebkitTextStroke: '0px transparent',
        background: 'none'
      }
    }
    return {
      WebkitTextFillColor: 'transparent',
      color: 'transparent',
      WebkitTextStroke: `${cfg.strokeWidth}px ${cfg.strokeColor}`,
      background: 'none'
    }
  }
  return { opacity: cfg.letterFlickerOpacity / 100 }
})

onMounted(() => {
  // 减动效偏好：保持静态实填，不创建观察器、不响应悬停
  // Reduced motion: keep the static filled text; no observer, no hover
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (!reduced) setupObserver()
})

onBeforeUnmount(() => {
  clearTimers()
  io?.disconnect()
  io = null
})
</script>

<style scoped>
.flicker-text {
  display: inline;
}
</style>
