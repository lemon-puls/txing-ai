<template>
  <!-- 极光流光装饰层（WebGL 着色器渲染，纯装饰，不拦截交互；Ribbon Glow 效果的自研移植） -->
  <!-- Aurora-ribbon decorative layer (WebGL shader, decorative only, never intercepts input; an original port of the Ribbon Glow effect) -->
  <div ref="wrapRef" class="aurora-ribbons" aria-hidden="true">
    <canvas ref="canvasRef" class="aurora-ribbons-canvas"></canvas>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watchEffect } from 'vue'

// ===== 观感常量 =====
// Look-and-feel constants
const MAX_DPR = 1.5 // 全屏逐像素 fbm 开销大，DPR 上限取 1.5 平衡清晰度与填充率 / per-pixel fbm is fill-rate heavy

const VERT_SRC = `
attribute vec2 a_pos;
void main(){ gl_Position = vec4(a_pos, 0.0, 1.0); }
`

const FRAG_SRC = `
#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

uniform vec2 uRes;
uniform float uTime;        // 累计漂移相位（CPU 侧取模，防长时运行精度劣化）/ accumulated drift phase
uniform vec2 uPtr;          // 阻尼后的指针（与 p 同空间：x 除以高度）/ damped pointer, height-normalized space
uniform float uPtrStrength; // 指针存在感 0-1（离开渐隐，吸引随之失效）/ pointer presence 0-1
uniform float uWarp;        // 指针吸引强度 / pointer attraction strength
uniform float uIntensity;   // 光带亮度 / ribbon brightness
uniform float uScale;       // 光带疏密 / ribbon density
uniform float uAdd;         // 1=暗色加法发光 0=亮色插值 / 1=dark additive, 0=light lerp
uniform float uExposure;    // 软压顶曝光 / soft filmic exposure
uniform vec3 uC0, uC1, uC2; // 三层光带色 / ribbon hues
uniform vec3 uBgTop, uBgBottom;

float hash12(vec2 p){
  vec3 q = fract(vec3(p.xyx) * 0.1031);
  q += dot(q, q.yzx + 33.33);
  return fract((q.x + q.y) * q.z);
}

float vnoise(vec2 x){
  vec2 i = floor(x), f = fract(x);
  f = f * f * (3.0 - 2.0 * f);
  return mix(mix(hash12(i), hash12(i + vec2(1.0, 0.0)), f.x),
             mix(hash12(i + vec2(0.0, 1.0)), hash12(i + vec2(1.0, 1.0)), f.x), f.y);
}

float fbm(vec2 p){
  float a = 0.5, s = 0.0;
  for (int i = 0; i < 3; i++){
    s += a * vnoise(p);
    p *= 2.03;
    a *= 0.5;
  }
  return s;
}

// 光带轴线：双正弦 + fbm 域扭曲，向指针做高斯衰减吸引（勿用 pow：负底数未定义）
// Ribbon axis: two sines + fbm domain warp, gaussian-falloff attraction toward the pointer (no pow: negative base is UB)
float axisY(vec2 p, float seed){
  float x = p.x * uScale + uTime * (0.12 + 0.07 * seed);
  float y = 0.5
    + 0.20 * sin(x * 1.7 + seed * 11.0 + uTime * 0.35)
    + 0.10 * sin(x * 3.3 - seed * 5.0 - uTime * 0.21)
    + 0.45 * (fbm(vec2(x * 0.8 + seed * 13.0, uTime * 0.10 + seed)) - 0.5);
  float dx = (p.x - uPtr.x) * 2.2;
  float gx = exp(-dx * dx);
  return y + (uPtr.y - y) * gx * uWarp * uPtrStrength;
}

// 轴线上的高斯截面亮度 / gaussian cross-section brightness around the axis
float ribbon(vec2 p, float seed, float thick){
  float d = p.y - axisY(p, seed);
  return exp(-d * d * thick);
}

void main(){
  vec2 p = gl_FragCoord.xy / max(uRes.y, 1.0);
  vec3 base = mix(uBgBottom, uBgTop, smoothstep(0.0, 1.0, p.y));

  // 三层交织光带（缩放 1.13/0.87 + 厚度梯度营造编束感）
  // Three woven layers (scale 1.13/0.87 + thickness spread for the braided look)
  float r0 = ribbon(p, 0.0, 550.0);
  float r1 = ribbon(p * 1.13 + vec2(0.0, 0.08), 1.7, 900.0);
  float r2 = ribbon(p * 0.87 + vec2(0.0, -0.06), 3.9, 1500.0);

  vec3 col;
  if (uAdd > 0.5) {
    // 暗色：加法发光 + 软压顶防削顶 / dark: additive glow with a soft filmic rolloff
    col = base + uC0 * (r0 * uIntensity * 0.9)
              + uC1 * (r1 * uIntensity * 0.6)
              + uC2 * (r2 * uIntensity * 0.4);
    col = 1.0 - exp(-col * uExposure);
  } else {
    // 亮色：向光带色插值（加法会洗成白）/ light: lerp toward ribbon hues (additive would wash to white)
    col = mix(base, uC0, clamp(r0 * uIntensity * 0.9, 0.0, 1.0));
    col = mix(col, uC1, clamp(r1 * uIntensity * 0.6, 0.0, 1.0));
    col = mix(col, uC2, clamp(r2 * uIntensity * 0.4, 0.0, 1.0));
  }

  gl_FragColor = vec4(col, 1.0);
}
`

// ===== 主题配色（光带为品牌渐变三色；亮色底与 CloudSky 地平线同源） =====
// Theme palettes (ribbons use the brand gradient hues; light base matches CloudSky's horizon)
const LIGHT_PALETTE = {
  bgTop: '#f6f9ff',
  bgBottom: '#e6effc',
  c0: '#2B5EFF',
  c1: '#1E88E5',
  c2: '#03A9F4'
}
const DARK_PALETTE = {
  bgTop: '#050814',
  bgBottom: '#0b1226',
  c0: '#2B5EFF',
  c1: '#1E88E5',
  c2: '#03A9F4'
}

// ===== 可调参数 =====
// Tunable props
const props = defineProps({
  speed: { type: Number, default: 55 },     // 时间流速 0-100 / time flow speed
  intensity: { type: Number, default: 65 }, // 光带亮度 0-100 / ribbon brightness
  warp: { type: Number, default: 45 },      // 指针吸引强度 0-100 / pointer attraction
  scale: { type: Number, default: 110 },    // 光带疏密 20-300 / ribbon density
  damping: { type: Number, default: 40 }    // 指针跟随阻尼 1-100 / pointer damping
})

function clampN(v, lo, hi) {
  return v < lo ? lo : v > hi ? hi : v
}

function parseColor(input, fb) {
  if (!input) return fb
  const str = String(input).trim()
  if (str.charAt(0) === '#') {
    let hex = str.slice(1)
    if (hex.length === 3 || hex.length === 4) {
      hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2] + (hex.length === 4 ? hex[3] + hex[3] : '')
    }
    if (hex.length >= 6) {
      const r = parseInt(hex.slice(0, 2), 16)
      const g = parseInt(hex.slice(2, 4), 16)
      const b = parseInt(hex.slice(4, 6), 16)
      const a = hex.length >= 8 ? parseInt(hex.slice(6, 8), 16) / 255 : 1
      if (!isNaN(r) && !isNaN(g) && !isNaN(b)) return [r / 255, g / 255, b / 255, a]
    }
    return fb
  }
  const m = str.match(/[\d.]+/g)
  if (m && m.length >= 3) {
    return [
      Math.min(255, parseFloat(m[0])) / 255,
      Math.min(255, parseFloat(m[1])) / 255,
      Math.min(255, parseFloat(m[2])) / 255,
      m.length >= 4 ? Math.min(1, parseFloat(m[3])) : 1
    ]
  }
  return fb
}

function compile(gl, type, src) {
  const sh = gl.createShader(type)
  if (!sh) return null
  gl.shaderSource(sh, src)
  gl.compileShader(sh)
  if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) {
    console.error('AuroraRibbons shader:', gl.getShaderInfoLog(sh))
    gl.deleteShader(sh)
    return null
  }
  return sh
}

const wrapRef = ref(null)
const canvasRef = ref(null)

// 跟随根节点 dark 类切换主题（与 CloudSky 同一判定方式）
// Follow the root `dark` class (same signal as CloudSky)
const isDark = ref(document.documentElement.classList.contains('dark'))
let themeMo = null

// 每帧 uniform 快照：props + 主题 → 数值（watchEffect 中重算，渲染循环只读）
// Per-frame uniform snapshot: props + theme → numbers (rebuilt in watchEffect, read-only in the loop)
const uni = {
  speed: 1.1, intensity: 0.65, warp: 0.45, scale: 1.1, damping: 40,
  add: 1, exposure: 1.1,
  c0: [0.17, 0.37, 1, 1], c1: [0.12, 0.53, 0.9, 1], c2: [0.01, 0.66, 0.96, 1],
  bgTop: [0.02, 0.03, 0.08, 1], bgBottom: [0.04, 0.07, 0.15, 1]
}

// ===== 运行态 =====
// Runtime state —— 必须先于 watchEffect 声明：其首次同步执行会调用 redrawIfIdle（避免 TDZ）
// Declared before watchEffect: its first synchronous run calls redrawIfIdle (avoids the TDZ error)
let gl = null
let reduced = false
let coarsePtr = false
let dprCap = MAX_DPR
let inView = false
let running = false
let raf = 0
let last = 0
let drift = 0
let strength = 0
const lean = { x: 0, y: 0 }
const ptr = { x: 0, y: 0, inside: false }
let io = null
const locs = {}
const U = (name) => {
  if (!(name in locs)) locs[name] = gl.getUniformLocation(gl.getParameter(gl.CURRENT_PROGRAM), name)
  return locs[name]
}

const draw = () => {
  if (!gl) return
  const canvas = canvasRef.value
  if (!canvas) return
  const dpr = Math.min(window.devicePixelRatio || 1, dprCap)
  const cw = canvas.clientWidth || 1200
  const ch = canvas.clientHeight || 800
  const bw = Math.max(1, Math.round(cw * dpr))
  const bh = Math.max(1, Math.round(ch * dpr))
  if (canvas.width !== bw || canvas.height !== bh) {
    canvas.width = bw
    canvas.height = bh
  }
  gl.viewport(0, 0, bw, bh)

  gl.uniform2f(U('uRes'), bw, bh)
  gl.uniform1f(U('uTime'), drift)
  gl.uniform2f(U('uPtr'), lean.x, lean.y)
  gl.uniform1f(U('uPtrStrength'), strength)
  gl.uniform1f(U('uWarp'), uni.warp)
  gl.uniform1f(U('uIntensity'), uni.intensity)
  gl.uniform1f(U('uScale'), uni.scale)
  gl.uniform1f(U('uAdd'), uni.add)
  gl.uniform1f(U('uExposure'), uni.exposure)
  gl.uniform3f(U('uC0'), uni.c0[0], uni.c0[1], uni.c0[2])
  gl.uniform3f(U('uC1'), uni.c1[0], uni.c1[1], uni.c1[2])
  gl.uniform3f(U('uC2'), uni.c2[0], uni.c2[1], uni.c2[2])
  gl.uniform3f(U('uBgTop'), uni.bgTop[0], uni.bgTop[1], uni.bgTop[2])
  gl.uniform3f(U('uBgBottom'), uni.bgBottom[0], uni.bgBottom[1], uni.bgBottom[2])

  gl.drawArrays(gl.TRIANGLES, 0, 3)
}

const step = (dt) => {
  // 指针位置与存在感双指数趋近（CloudSky 同款公式）；离开后缓归屏幕中心，存在感归零吸引自动失效
  // Pointer position and presence ease exponentially (CloudSky's formula); on leave the target glides home
  const el = wrapRef.value
  const aspect = el && el.clientHeight > 0 ? el.clientWidth / el.clientHeight : 1.6
  const k = 1 - Math.exp(-uni.damping * 0.12 * dt)
  lean.x += ((ptr.inside ? ptr.x : aspect * 0.5) - lean.x) * k
  lean.y += ((ptr.inside ? ptr.y : 0.5) - lean.y) * k
  strength += ((ptr.inside ? 1 : 0) - strength) * (1 - Math.exp(-2.5 * dt))

  // 时间累加而非 now*speed：调整 speed 不跳变；取模防长时运行浮点精度劣化
  // Accumulate time instead of now*speed (no jump when speed changes); wrap to avoid float precision decay
  drift = (drift + dt * uni.speed) % 1000
  draw()
}

const frame = (now) => {
  raf = requestAnimationFrame(frame) // 先挂下一帧，避免空隙 / schedule the next frame first
  const dt = Math.min((now - last) / 1000 || 0.016, 0.05)
  last = now
  step(dt)
}

const ensureLoop = () => {
  if (running || reduced || !gl) return
  running = true
  last = performance.now()
  raf = requestAnimationFrame(frame)
}

const stopLoop = () => {
  running = false
  cancelAnimationFrame(raf)
}

// 静态一帧（减动效模式 / 初始化 / 换肤 / resize 且循环未运行时）
// Single static frame (reduced motion / init / theme change / resize while idle)
const drawOnce = () => {
  if (!gl || running) return
  step(0)
}

const redrawIfIdle = () => {
  if (!running) drawOnce()
}

watchEffect(() => {
  const dark = isDark.value
  const pal = dark ? DARK_PALETTE : LIGHT_PALETTE
  uni.speed = clampN(props.speed, 0, 100) / 50
  // 亮色下压低亮度，避免彩色过艳 / tame saturation on light theme
  uni.intensity = (clampN(props.intensity, 0, 100) / 100) * (dark ? 1.0 : 0.8)
  uni.warp = clampN(props.warp, 0, 100) / 100
  uni.scale = clampN(props.scale, 20, 300) / 100
  uni.damping = clampN(props.damping, 1, 100)
  uni.add = dark ? 1 : 0
  uni.exposure = dark ? 1.1 : 1.0
  uni.c0 = parseColor(pal.c0, [0.17, 0.37, 1, 1])
  uni.c1 = parseColor(pal.c1, [0.12, 0.53, 0.9, 1])
  uni.c2 = parseColor(pal.c2, [0.01, 0.66, 0.96, 1])
  uni.bgTop = parseColor(pal.bgTop, [0.02, 0.03, 0.08, 1])
  uni.bgBottom = parseColor(pal.bgBottom, [0.04, 0.07, 0.15, 1])
  redrawIfIdle()
})

// 指针吸引：window 级监听 —— 组件 pointer-events:none 且被内容遮挡时依旧生效
// Pointer attraction: window-level listeners keep working even though the layer is pointer-transparent and covered
const onPointerMove = (e) => {
  const el = wrapRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  if (r.height <= 0) return
  // 与 shader 的 p 同空间：x 也除以高度（p = fragCoord / uRes.y），y 翻转为自下而上
  // Same space as shader p: x also normalized by height (p = fragCoord / uRes.y), y flipped bottom-up
  ptr.x = (e.clientX - r.left) / r.height
  ptr.y = 1 - (e.clientY - r.top) / r.height
  ptr.inside = true
}
const onPointerGone = () => {
  ptr.inside = false
}
const onResize = () => {
  if (!running) drawOnce()
}
const onIntersect = (entries) => {
  inView = entries[0]?.isIntersecting ?? false
  if (inView) {
    if (reduced) drawOnce()
    else ensureLoop()
  } else {
    stopLoop() // 离开视口即停帧，省电 / off-screen: stop the loop to save power
  }
}
const onThemeChange = () => {
  isDark.value = document.documentElement.classList.contains('dark')
}

onMounted(() => {
  const canvas = canvasRef.value
  if (!canvas) return
  gl = canvas.getContext('webgl', { alpha: false, antialias: false, depth: false })
  if (!gl) return // WebGL 不可用时保留 CSS 渐变兜底 / keep the CSS gradient fallback

  const vs = compile(gl, gl.VERTEX_SHADER, VERT_SRC)
  const fs = compile(gl, gl.FRAGMENT_SHADER, FRAG_SRC)
  if (!vs || !fs) { gl = null; return }
  const prog = gl.createProgram()
  gl.attachShader(prog, vs)
  gl.attachShader(prog, fs)
  gl.linkProgram(prog)
  if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
    console.error('AuroraRibbons link:', gl.getProgramInfoLog(prog))
    gl = null
    return
  }
  gl.useProgram(prog)

  const buf = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, buf)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW)
  const aPos = gl.getAttribLocation(prog, 'a_pos')
  gl.enableVertexAttribArray(aPos)
  gl.vertexAttribPointer(aPos, 2, gl.FLOAT, false, 0, 0)

  reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  coarsePtr = window.matchMedia('(hover: none)').matches
  dprCap = coarsePtr ? 1.25 : MAX_DPR

  window.addEventListener('resize', onResize, { passive: true })
  // 触屏与减动效环境不挂指针监听：光带仅随时间自主流动
  // Touch and reduced-motion skip pointer listeners: ribbons flow autonomously with time
  if (!reduced && !coarsePtr) {
    window.addEventListener('pointermove', onPointerMove, { passive: true })
    window.addEventListener('blur', onPointerGone)
    document.documentElement.addEventListener('mouseleave', onPointerGone)
  }

  // 主题跟随
  // Theme tracking
  themeMo = new MutationObserver(onThemeChange)
  themeMo.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  // 立即呈现一帧，避免兜底渐变闪现；视口外自动停帧
  // Draw one frame up front (no fallback flash); auto-pause off-screen
  drawOnce()
  io = new IntersectionObserver(onIntersect)
  io.observe(wrapRef.value)
})

onBeforeUnmount(() => {
  stopLoop()
  io?.disconnect()
  themeMo?.disconnect()
  window.removeEventListener('resize', onResize)
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('blur', onPointerGone)
  document.documentElement.removeEventListener('mouseleave', onPointerGone)
  // 释放 WebGL 上下文 / release the GL context
  gl?.getExtension('WEBGL_lose_context')?.loseContext()
  gl = null
})
</script>

<style scoped lang="scss">
.aurora-ribbons {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  // WebGL 不可用时的渐变兜底（与两套主题配色一致）
  // CSS gradient fallback when WebGL is unavailable (matches both theme palettes)
  background: linear-gradient(180deg, #f6f9ff 0%, #e6effc 100%);

  html.dark & {
    background: linear-gradient(180deg, #050814 0%, #0b1226 100%);
  }
}

.aurora-ribbons-canvas {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
