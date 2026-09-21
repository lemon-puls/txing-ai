<template>
  <!-- 云海天空装饰层（WebGL 着色器渲染，纯装饰，不拦截交互） -->
  <!-- Cloud-sky decorative layer (WebGL shader, decorative only, never intercepts input) -->
  <div ref="wrapRef" class="cloud-sky" aria-hidden="true">
    <canvas ref="canvasRef" class="cloud-sky-canvas"></canvas>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watchEffect } from 'vue'

// ===== 观感常量（与 Originkit CloudSky 保持一致） =====
// Look-and-feel constants (kept identical to the Originkit CloudSky source)
const MAX_DPR = 1.5 // 全屏逐像素 fbm 开销大，DPR 上限取 1.5 平衡清晰度与填充率 / per-pixel fbm is fill-rate heavy
const PUFF_UP = 0.34
const PUFF_DOWN = 0.19
const ERODE = 0.7
const SHADOW_STEP = 0.085
const NEAR_CELL = 1.05
const FAR_CELL = 2.15
const FAR_MIX = 0.55
const NEAR_DRIFT = 0.055
const FAR_DRIFT = 0.026
const CIRRUS_DRIFT = 0.014
const PUFF_WMAX = 2.15
const SHADE_BLEND = 12.0

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
uniform float uNearX, uFarX, uCirrusX;
uniform float uCoverage, uSize, uSoftness, uShadow, uCirrus;
uniform vec3 uZenith, uHorizon, uCloud;
uniform vec4 uGlow;
uniform vec2 uSun;
uniform vec2 uParallax;

vec2 hash22(vec2 p){
  vec3 q = fract(vec3(p.xyx) * vec3(0.1031, 0.1030, 0.0973));
  q += dot(q, q.yzx + 33.33);
  return fract((q.xx + q.yz) * q.zy);
}

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
  for (int i = 0; i < 4; i++){
    s += a * vnoise(p);
    p *= 2.03;
    a *= 0.5;
  }
  return s;
}

vec2 blobs(vec2 uv, float seed){
  vec2 id = floor(uv), f = fract(uv);

  float best = -1e4;
  float wsum = 0.0, ysum = 0.0;
  float wMax = min(${PUFF_WMAX.toFixed(3)}, 0.72 * uSize);
  float reach = min(2.0, ceil(wMax + 0.85) - 1.0);
  for (int j = -2; j <= 2; j++){
    for (int i = -2; i <= 2; i++){
      vec2 o = vec2(float(i), float(j));
      if (max(abs(o.x), abs(o.y)) > reach) continue;
      vec2 h = hash22(id + o + seed);

      if (fract(h.x * 37.1) > uCoverage) continue;
      vec2 c = o + 0.15 + h * 0.7;
      float w = min(${PUFF_WMAX.toFixed(3)}, (0.30 + 0.42 * fract(h.y * 19.7)) * uSize);
      vec2 d = f - c;

      float ry = (d.y > 0.0 ? ${PUFF_UP.toFixed(3)} : ${PUFF_DOWN.toFixed(3)}) * uSize * (0.8 + 0.5 * fract(h.y * 7.3));
      float e = length(vec2(d.x / max(w, 1e-3), d.y / max(ry, 1e-3)));
      float val = 1.0 - e;
      float yN = d.y / max(ry, 1e-3);
      if (val > best){
        float k = exp(${SHADE_BLEND.toFixed(1)} * (best - val));
        wsum = wsum * k + 1.0;
        ysum = ysum * k + yN;
        best = val;
      } else {
        float g = exp(${SHADE_BLEND.toFixed(1)} * (val - best));
        wsum += g;
        ysum += g * yN;
      }
    }
  }
  return vec2(best, ysum / max(wsum, 1e-4));
}

vec2 cloudField(vec2 uv, float seed, float detailScale){
  vec2 b = blobs(uv, seed);

  float n = fbm(uv * detailScale + seed * 3.1) * 0.72
          + fbm(uv * detailScale * 3.3 + seed * 7.7) * 0.28;
  return vec2(b.x - (1.0 - n) * ${ERODE.toFixed(3)}, b.y);
}

vec3 shadeCloud(float dyNorm, vec3 sky){
  float t = smoothstep(-0.95, 0.25, dyNorm);
  vec3 base = mix(uCloud * 0.52, sky, 0.34);
  return mix(mix(uCloud, base, uShadow), uCloud, t);
}

void main(){
  vec2 frag = gl_FragCoord.xy / max(uRes.y, 1.0);
  float aspect = uRes.x / max(uRes.y, 1.0);
  vec2 p = vec2(frag.x, frag.y);

  vec3 sky = mix(uHorizon, uZenith, smoothstep(-0.15, 1.05, p.y));
  vec2 sunP = vec2(uSun.x * aspect, uSun.y);
  float sd = length(p - sunP);

  sky += uGlow.rgb * uGlow.a * exp(-sd * 3.4) * 0.30;

  vec3 col = sky;

  if (uCirrus > 0.0) {
    vec2 cuv = vec2(p.x * 1.4 + uCirrusX, p.y * 5.5);
    float veil = fbm(cuv) * fbm(cuv * 2.3 + 9.0);
    veil = smoothstep(0.24, 0.55, veil) * smoothstep(0.15, 0.7, p.y);
    col = mix(col, uCloud, veil * uCirrus * 0.5);
  }

  vec2 fuv = vec2(p.x + uFarX, p.y) * ${FAR_CELL.toFixed(3)} + uParallax * 0.4;
  vec2 fd = cloudField(fuv, 17.0, 11.0);
  float fa = clamp(fd.x * uSoftness, 0.0, 1.0);
  if (fa > 0.0) {
    vec3 lit = shadeCloud(fd.y, sky);

    col = mix(col, mix(lit, sky, ${FAR_MIX.toFixed(3)}), fa);
  }

  vec2 nuv = vec2(p.x + uNearX, p.y) * ${NEAR_CELL.toFixed(3)} + uParallax;
  vec2 nd = cloudField(nuv, 3.0, 8.5);
  float na = clamp(nd.x * uSoftness, 0.0, 1.0);
  if (na > 0.0) {
    vec3 lit = shadeCloud(nd.y, sky);

    float above = clamp(cloudField(nuv + vec2(0.0, ${SHADOW_STEP.toFixed(3)}), 3.0, 8.5).x * uSoftness, 0.0, 1.0);
    lit *= 1.0 - 0.18 * uShadow * above;

    lit += uGlow.rgb * uGlow.a * 0.22 * exp(-length(p - sunP) * 1.6);
    col = mix(col, lit, na);
  }

  gl_FragColor = vec4(col, 1.0);
}
`

// ===== 主题配色（与 about 页星空底色衔接） =====
// Theme palettes (matched to the about page starfield base gradient)
// 亮色：晴日蓝天，地平线渐白，融入星空底色 #eaf1ff
// Light: clear-day sky, whitening toward the horizon into the starfield's #eaf1ff
const LIGHT_PALETTE = {
  zenith: '#4b93e8',
  horizon: '#e6effc',
  cloud: '#ffffff',
  glow: 'rgba(248, 251, 255, 0.85)',
  sunX: 80,
  sunY: 92
}
// 暗色：深夜蓝调 + 月晕，融入星空底色 #04070f
// Dark: night-blue dome with a moon glow, into the starfield's #04070f
const DARK_PALETTE = {
  zenith: '#060d20',
  horizon: '#1a2a55',
  cloud: '#96abd6',
  glow: 'rgba(186, 208, 255, 0.5)',
  sunX: 74,
  sunY: 88
}

// ===== 可调参数（默认值已针对本页调校） =====
// Tunable props (defaults tuned for the about page)
const props = defineProps({
  density: { type: Number, default: 58 },   // 云量 0-100 / coverage
  speed: { type: Number, default: 55 },     // 飘动速度 0-100 / drift speed
  size: { type: Number, default: 125 },     // 团块大小 20-300 / puff size
  softness: { type: Number, default: 210 }, // 边缘柔化 20-300（越大越柔）/ edge softness
  shadow: { type: Number, default: 80 },    // 明暗立体感 0-200 / shading strength
  cirrus: { type: Number, default: 65 },    // 卷云纱 0-100 / cirrus veil
  parallax: { type: Number, default: 130 }, // 指针视差 0-300 / pointer parallax
  wind: { type: Number, default: 90 },      // 指针扰动阵风 0-300 / pointer gust
  damping: { type: Number, default: 40 }    // 视差阻尼 1-100 / parallax damping
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
    console.error('CloudSky shader:', gl.getShaderInfoLog(sh))
    gl.deleteShader(sh)
    return null
  }
  return sh
}

const wrapRef = ref(null)
const canvasRef = ref(null)

// 跟随根节点 dark 类切换主题（与页面墨水层同一判定方式）
// Follow the root `dark` class (same signal the page ink layer uses)
const isDark = ref(document.documentElement.classList.contains('dark'))
let themeMo = null

// 每帧 uniform 快照：props + 主题 → 数值（watchEffect 中重算，渲染循环只读）
// Per-frame uniform snapshot: props + theme → numbers (rebuilt in watchEffect, read-only in the loop)
const uni = {
  coverage: 0.58, speed: 1.1, size: 1.25, softness: 2.14, shadow: 0.8, cirrus: 0.65,
  sunX: 0.8, sunY: 0.92, parallax: 1.3, wind: 0.9, damping: 40,
  zenith: [0.29, 0.58, 0.91, 1], horizon: [0.9, 0.94, 0.99, 1],
  cloud: [1, 1, 1, 1], glow: [0.97, 0.98, 1, 0.85]
}

// ===== 运行态 =====
// Runtime state —— 必须先于 watchEffect 声明：其首次同步执行会调用 redrawIfIdle（避免 TDZ）
// Declared before watchEffect: its first synchronous run calls redrawIfIdle (avoids the TDZ error)
let gl = null
let reduced = false
let inView = false
let running = false
let raf = 0
let last = 0
let nearX = 0
let farX = 0
let cirrusX = 0
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
  const dpr = Math.min(window.devicePixelRatio || 1, MAX_DPR)
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
  gl.uniform1f(U('uNearX'), nearX)
  gl.uniform1f(U('uFarX'), farX)
  gl.uniform1f(U('uCirrusX'), cirrusX)
  gl.uniform1f(U('uCoverage'), uni.coverage)
  gl.uniform1f(U('uSize'), uni.size)
  gl.uniform1f(U('uSoftness'), uni.softness)
  gl.uniform1f(U('uShadow'), uni.shadow)
  gl.uniform1f(U('uCirrus'), uni.cirrus)
  gl.uniform2f(U('uSun'), uni.sunX, uni.sunY)
  gl.uniform2f(U('uParallax'), -lean.x * uni.parallax * 0.07, -lean.y * uni.parallax * 0.05)
  gl.uniform3f(U('uZenith'), uni.zenith[0], uni.zenith[1], uni.zenith[2])
  gl.uniform3f(U('uHorizon'), uni.horizon[0], uni.horizon[1], uni.horizon[2])
  gl.uniform3f(U('uCloud'), uni.cloud[0], uni.cloud[1], uni.cloud[2])
  gl.uniform4f(U('uGlow'), uni.glow[0], uni.glow[1], uni.glow[2], uni.glow[3])

  gl.drawArrays(gl.TRIANGLES, 0, 3)
}

const step = (dt) => {
  // 视差与阵风共用同一平滑指针（指数趋近）
  // Parallax and gust share one smoothed pointer (exponential ease)
  const k = 1 - Math.exp(-uni.damping * 0.12 * dt)
  lean.x += ((ptr.inside ? ptr.x : 0) - lean.x) * k
  lean.y += ((ptr.inside ? ptr.y : 0) - lean.y) * k

  const gust = 1 + lean.x * uni.wind
  const rate = uni.speed * gust
  nearX = (nearX - NEAR_DRIFT * rate * dt) % 1000
  farX = (farX - FAR_DRIFT * rate * dt) % 1000
  cirrusX = (cirrusX - CIRRUS_DRIFT * rate * dt) % 1000

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
  const darkK = dark ? 0.72 : 1 // 暗色降低云量，给星空留出透气感 / fewer clouds at night to keep the stars breathable
  uni.coverage = (clampN(props.density, 0, 100) / 100) * darkK
  uni.speed = clampN(props.speed, 0, 100) / 50
  uni.size = clampN(props.size, 20, 300) / 100
  uni.softness = 4.5 / Math.max(0.15, clampN(props.softness, 20, 300) / 100)
  uni.shadow = clampN(props.shadow, 0, 200) / 100
  uni.cirrus = (clampN(props.cirrus, 0, 100) / 100) * (dark ? 0.9 : 1)
  uni.sunX = clampN(pal.sunX, 0, 100) / 100
  uni.sunY = clampN(pal.sunY, 0, 100) / 100
  uni.parallax = clampN(props.parallax, 0, 300) / 100
  uni.wind = clampN(props.wind, 0, 300) / 100
  uni.damping = clampN(props.damping, 1, 100)
  uni.zenith = parseColor(pal.zenith, [0.29, 0.58, 0.91, 1])
  uni.horizon = parseColor(pal.horizon, [0.9, 0.94, 0.99, 1])
  uni.cloud = parseColor(pal.cloud, [1, 1, 1, 1])
  uni.glow = parseColor(pal.glow, [0.97, 0.98, 1, 0.85])
  redrawIfIdle()
})

// 指针视差：window 级监听 —— 组件 pointer-events:none 且被内容遮挡时依旧生效
// Pointer parallax: window-level listeners keep working even though the layer is pointer-transparent and covered
const onPointerMove = (e) => {
  const el = wrapRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  if (r.width <= 0 || r.height <= 0) return
  ptr.x = ((e.clientX - r.left) / r.width) * 2 - 1
  ptr.y = 1 - ((e.clientY - r.top) / r.height) * 2
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
    console.error('CloudSky link:', gl.getProgramInfoLog(prog))
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

  window.addEventListener('resize', onResize, { passive: true })
  if (!reduced) {
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
.cloud-sky {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  // WebGL 不可用时的渐变兜底（与两套主题配色一致）
  // CSS gradient fallback when WebGL is unavailable (matches both theme palettes)
  background: linear-gradient(180deg, #4b93e8 0%, #b9d5f2 70%, #e6effc 100%);

  html.dark & {
    background: linear-gradient(180deg, #060d20 0%, #10204a 70%, #1a2a55 100%);
  }
}

.cloud-sky-canvas {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
