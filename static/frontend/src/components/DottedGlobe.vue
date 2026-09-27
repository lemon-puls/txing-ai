<template>
  <!-- 点阵地球装饰层（WebGL 着色器渲染，纯装饰，不拦截交互；originkit Globe 效果的自研移植） -->
  <!-- Dotted-globe decorative layer (WebGL shader, decorative only, never intercepts input; an original port of originkit's Globe) -->
  <div ref="wrapRef" class="dotted-globe" aria-hidden="true">
    <canvas ref="canvasRef" class="dotted-globe-canvas"></canvas>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watchEffect } from 'vue'
import landMaskUrl from '@/assets/images/land-mask.png'

// 陆地标定：Natural Earth 110m land（公有领域）烘焙的等距圆柱 PNG，运行时采样出陆地点
// Land mask: equirectangular PNG baked from Natural Earth 110m land (public domain), sampled for land points

// ===== 观感常量 =====
// Look-and-feel constants
const MAX_DPR = 1.5 // 点/线渲染轻量，可承受 1.5 / points+lines are light enough for 1.5
const SPHERE_FIT = 0.93 // 球半径占画布半宽（裁剪空间）/ sphere radius in clip units
const CAM_DIST = 2.7 // 相机距离（球半径 = 1）/ camera distance in sphere radii
const YAW0 = 0.4 // 初始经度对准 -23°（originkit 默认视角）/ face lon -23° like originkit
const POINT_CANDIDATES = 42000 // Fibonacci 候选数（约 1/3 落在陆地）/ Fibonacci candidates (~1/3 on land)
const GA = Math.PI * (3 - Math.sqrt(5)) // 黄金角分布 / golden-angle spacing

const VERT_SRC = `
attribute vec3 a_pos;
attribute float a_seed;

uniform float uYaw;       // 自转相位 / auto-rotation phase
uniform float uLeanYaw;   // 指针带来的轻微偏航 / pointer-driven yaw
uniform float uLeanTilt;  // 指针带来的轻微俯仰 / pointer-driven tilt
uniform float uDist;      // 相机距离 / camera distance
uniform float uFit;       // 球半径 → 裁剪空间缩放 / sphere radius → clip scale
uniform float uDotPx;     // 点基准像素尺寸 / base point size in px
uniform float uTime;      // 色源光波相位 / color-source wave phase
uniform float uMode;      // 0=点阵 1=经纬线 / 0=points 1=graticule

varying vec3 vGlow;    // 三色源各自的光强 / per-source glow
varying float vFacing; // 朝向相机的分量 / facing toward camera
varying float vSeed;

vec3 rotY(vec3 p, float a){
  float c = cos(a), s = sin(a);
  return vec3(c * p.x + s * p.z, p.y, -s * p.x + c * p.z);
}
vec3 rotX(vec3 p, float a){
  float c = cos(a), s = sin(a);
  return vec3(p.x, c * p.y - s * p.z, s * p.y + c * p.z);
}

// 色源方向沿慢轨道漂移 / source directions drift on slow orbits
vec3 srcDir(float i, float t){
  float a = t * (0.055 + 0.012 * i) + i * 2.1;
  float b = sin(t * 0.04 + i * 1.7) * 0.75;
  return normalize(vec3(cos(a) * cos(b), sin(b), sin(a) * cos(b)));
}

// 扩散环 + 色斑（originkit 的色源签名）/ expanding ring + color patch (originkit's signature)
float srcGlow(vec3 n, float i, float t){
  vec3 d = srcDir(i, t);
  float ang = acos(clamp(dot(n, d), -1.0, 1.0));
  float r = fract(t * (0.035 + 0.011 * i) + i * 0.37) * 3.5;
  float x = (ang - r) * 3.4;
  float ring = exp(-x * x);
  float y = ang * 1.8;
  float patch = exp(-y * y) * 0.75;
  return ring + patch;
}

void main(){
  // 自转 + 地轴倾角 23.5° + 指针轻微姿态偏转（无深度测试，背面仅变暗）
  // Auto-rotation + 23.5° axial tilt + subtle pointer attitude (no depth test; back side just dims)
  vec3 n = a_pos;
  vec3 p = rotX(rotY(n, uYaw + uLeanYaw), 0.41 + uLeanTilt);
  float w = uDist - p.z;

  vFacing = p.z;
  vSeed = a_seed;

  if (uMode < 0.5) {
    float g0 = srcGlow(n, 0.0, uTime);
    float g1 = srcGlow(n, 1.0, uTime);
    float g2 = srcGlow(n, 2.0, uTime);
    vGlow = vec3(g0, g1, g2);
    // 光波处轻微抬升，剪影上可见波纹 / lift at glow waves so they read on the silhouette
    p *= 1.0 + (g0 + g1 + g2) * 0.012;
    w = uDist - p.z;
    gl_PointSize = uDotPx * (uDist / w) * (0.82 + 0.36 * a_seed);
  } else {
    vGlow = vec3(0.0);
    gl_PointSize = 1.0;
  }

  gl_Position = vec4(p.xy * (uFit / w), 0.0, 1.0);
}
`

const FRAG_SRC = `
#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

uniform float uIntensity;  // 整体亮度 / overall brightness
uniform float uMode;
uniform vec3 uLand, uGrid, uC0, uC1, uC2;

varying vec3 vGlow;
varying float vFacing;
varying float vSeed;

void main(){
  // 背面点阵仍可见但变暗（无深度测试 → 不会因深度交换闪烁）
  // Back-face dots stay visible but dimmed (no depth test → no depth-swap flicker)
  float facing = mix(0.14, 1.0, smoothstep(-0.55, 0.5, vFacing));

  vec3 col;
  float a;
  if (uMode < 0.5) {
    vec2 q = gl_PointCoord - 0.5;
    float d2 = dot(q, q);
    float soft = smoothstep(0.25, 0.02, d2); // 软圆点 / soft round sprite
    col = uLand + uC0 * vGlow.x + uC1 * vGlow.y + uC2 * vGlow.z;
    a = soft * facing * uIntensity * (0.55 + 0.45 * vSeed);
  } else {
    col = uGrid;
    a = facing * uIntensity * 0.3;
  }

  // 预乘 over 混合（勿写 a=0 的加法：合成器按非预乘解读会除零出整块雾）
  // Premultiplied over — never additive-at-alpha-0: the compositor un-premultiplies and divides by zero (full-canvas haze)
  gl_FragColor = vec4(col * a, a);
}
`

// 球缘大气光晕（全屏三角形）/ limb atmosphere halo (fullscreen triangle)
const RIM_VERT_SRC = `
attribute vec2 a_pos;
varying vec2 vP;
void main(){
  vP = a_pos;
  gl_Position = vec4(a_pos, 0.0, 1.0);
}
`

const RIM_FRAG_SRC = `
#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

uniform float uIntensity;
uniform vec3 uRim;
varying vec2 vP;

void main(){
  float r = length(vP);
  float x = (r - 0.93) * 5.5;
  float halo = exp(-x * x); // 球缘光晕（勿加球内平铺光晕：那是难看的色块）/ limb halo (no inner flat bloom: it reads as an ugly disc)
  float a = halo * 0.5 * uIntensity;
  gl_FragColor = vec4(uRim * a, a); // 预乘 over / premultiplied over
}
`

// ===== 主题配色（点阵与光波取品牌三色） =====
// Theme palettes (dots and waves use the brand gradient hues)
const LIGHT_PALETTE = {
  land: '#2b52c8',
  grid: '#8fb0e8',
  rim: '#2B5EFF',
  c0: '#2B5EFF',
  c1: '#1E88E5',
  c2: '#03A9F4'
}
const DARK_PALETTE = {
  land: '#d5e4ff',
  grid: '#3d5f9e',
  rim: '#2B5EFF',
  c0: '#2B5EFF',
  c1: '#1E88E5',
  c2: '#03A9F4'
}

// ===== 可调参数 =====
// Tunable props
const props = defineProps({
  speed: { type: Number, default: 32 },     // 自转速度 0-100 / rotation speed
  intensity: { type: Number, default: 58 }, // 点阵亮度 0-100 / dot brightness
  dotSize: { type: Number, default: 46 },   // 点尺寸 0-100 / dot size
  lean: { type: Number, default: 30 },      // 指针姿态偏转 0-100 / pointer attitude sway
  grid: { type: Boolean, default: true }    // 15° 经纬网 / 15° graticule
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
    console.error('DottedGlobe shader:', gl.getShaderInfoLog(sh))
    gl.deleteShader(sh)
    return null
  }
  return sh
}

function link(gl, vsSrc, fsSrc) {
  const vs = compile(gl, gl.VERTEX_SHADER, vsSrc)
  const fs = compile(gl, gl.FRAGMENT_SHADER, fsSrc)
  if (!vs || !fs) return null
  const prog = gl.createProgram()
  gl.attachShader(prog, vs)
  gl.attachShader(prog, fs)
  gl.linkProgram(prog)
  if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
    console.error('DottedGlobe link:', gl.getProgramInfoLog(prog))
    return null
  }
  return prog
}

const wrapRef = ref(null)
const canvasRef = ref(null)

// 跟随根节点 dark 类切换主题（与 CloudSky/AuroraRibbons 同一判定方式）
// Follow the root `dark` class (same signal as CloudSky/AuroraRibbons)
const isDark = ref(document.documentElement.classList.contains('dark'))
let themeMo = null

// 每帧 uniform 快照：props + 主题 → 数值（watchEffect 中重算，渲染循环只读）
// Per-frame uniform snapshot: props + theme → numbers (rebuilt in watchEffect, read-only in the loop)
const uni = {
  speed: 0.32, intensity: 0.58, dotPx: 2.8, lean: 0.3,
  land: [0.84, 0.89, 1, 1], grid: [0.24, 0.37, 0.62, 1], rim: [0.17, 0.37, 1, 1],
  c0: [0.17, 0.37, 1, 1], c1: [0.12, 0.53, 0.9, 1], c2: [0.01, 0.66, 0.96, 1]
}

// ===== 运行态 =====
// Runtime state —— 必须先于 watchEffect 声明：其首次同步执行会调用 redrawIfIdle（避免 TDZ）
// Declared before watchEffect: its first synchronous run calls redrawIfIdle (avoids the TDZ error)
let gl = null
let progMain = null
let progRim = null
let locMain = null
let locRim = null
let aPosMain = -1
let aSeedMain = -1
let aPosRim = -1
let bufPoints = null
let bufGrid = null
let bufTri = null
let pointCount = 0
let gridCount = 0
let pointsReady = false
let reduced = false
let coarsePtr = false
let dprCap = MAX_DPR
let inView = false
let running = false
let raf = 0
let last = 0
let yaw = YAW0
let drift = 0
let strength = 0
const lean = { x: 0, y: 0 }
const ptr = { x: 0, y: 0, inside: false }
let io = null

// 两套 program 各自缓存 uniform 位置（attribute 必须走 getAttribLocation）
// Cache uniform locations per program (attributes must use getAttribLocation)
const cacheLocs = (prog, names) => {
  const out = {}
  for (const n of names) out[n] = gl.getUniformLocation(prog, n)
  return out
}

// 经纬度 → 单位球坐标（经度 0 指向 +z）/ lat-lon → unit sphere (lon 0 at +z)
const ll2v = (lon, lat) => [
  Math.cos(lat) * Math.sin(lon),
  Math.sin(lat),
  Math.cos(lat) * Math.cos(lon)
]

// Fibonacci 球面均匀采样 + 陆地标定过滤 → 单次 draw call 的点阵
// Fibonacci sphere sampling + land-mask filter → one draw call of dots
const buildPoints = (pixels, mw, mh) => {
  const pos = []
  const seeds = []
  for (let i = 0; i < POINT_CANDIDATES; i++) {
    const y = 1 - (i + 0.5) * (2 / POINT_CANDIDATES)
    const r = Math.sqrt(Math.max(0, 1 - y * y))
    const th = i * GA
    const x = Math.cos(th) * r
    const z = Math.sin(th) * r
    // 等距圆柱投影采样陆地 / sample land via equirectangular projection
    const lon = Math.atan2(x, z)
    const lat = Math.asin(clampN(y, -1, 1))
    const u = (lon + Math.PI) / (2 * Math.PI)
    const v = (Math.PI / 2 - lat) / Math.PI
    const mx = Math.min(mw - 1, Math.max(0, Math.floor(u * mw)))
    const my = Math.min(mh - 1, Math.max(0, Math.floor(v * mh)))
    if (pixels[(my * mw + mx) * 4] > 127) {
      pos.push(x, y, z)
      // 确定性伪随机种子（重载视觉一致）/ deterministic pseudo-random seed (stable across reloads)
      seeds.push(Math.abs(Math.sin(i * 127.1) * 43758.5453) % 1)
    }
  }
  pointCount = seeds.length
  const data = new Float32Array(pointCount * 4)
  for (let i = 0; i < pointCount; i++) {
    data[i * 4] = pos[i * 3]
    data[i * 4 + 1] = pos[i * 3 + 1]
    data[i * 4 + 2] = pos[i * 3 + 2]
    data[i * 4 + 3] = seeds[i]
  }
  gl.bindBuffer(gl.ARRAY_BUFFER, bufPoints)
  gl.bufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW)
  pointsReady = true
}

// 15° 经纬网（经线 -180..165、纬线 -75..75，步进 3°）/ 15° graticule (meridians -180..165, parallels -75..75, 3° steps)
const buildGrid = () => {
  const seg = []
  const step = (3 * Math.PI) / 180
  for (let lonD = -180; lonD < 180; lonD += 15) {
    const lon = (lonD * Math.PI) / 180
    for (let lat = -Math.PI / 2; lat < Math.PI / 2 - 1e-6; lat += step) {
      const lat2 = Math.min(Math.PI / 2, lat + step)
      seg.push(...ll2v(lon, lat), ...ll2v(lon, lat2))
    }
  }
  for (let latD = -75; latD <= 75; latD += 15) {
    const lat = (latD * Math.PI) / 180
    for (let lon = -Math.PI; lon < Math.PI - 1e-6; lon += step) {
      const lon2 = Math.min(Math.PI, lon + step)
      seg.push(...ll2v(lon, lat), ...ll2v(lon2, lat))
    }
  }
  gridCount = seg.length / 3
  gl.bindBuffer(gl.ARRAY_BUFFER, bufGrid)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(seg), gl.STATIC_DRAW)
}

const draw = () => {
  if (!gl) return
  const canvas = canvasRef.value
  if (!canvas) return
  const cw = canvas.clientWidth || 600
  if (cw <= 0) return // 隐藏（移动端 display:none）时跳过 / hidden (mobile display:none): skip
  const dpr = Math.min(window.devicePixelRatio || 1, dprCap)
  const bh = Math.max(1, Math.round(cw * dpr)) // 画布为正方形 / canvas is square
  if (canvas.width !== bh) {
    canvas.width = bh
    canvas.height = bh
  }
  gl.viewport(0, 0, bh, bh)

  // 预乘 over 混合：透明画布叠在流光之上 / premultiplied over: transparent canvas composited on the aurora
  gl.blendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

  gl.clearColor(0, 0, 0, 0)
  gl.clear(gl.COLOR_BUFFER_BIT)

  // 先大气光晕 / atmosphere halo first
  gl.useProgram(progRim)
  gl.uniform1f(locRim.uIntensity, uni.intensity)
  gl.uniform3f(locRim.uRim, uni.rim[0], uni.rim[1], uni.rim[2])
  gl.bindBuffer(gl.ARRAY_BUFFER, bufTri)
  gl.enableVertexAttribArray(aPosRim)
  gl.vertexAttribPointer(aPosRim, 2, gl.FLOAT, false, 0, 0)
  gl.drawArrays(gl.TRIANGLES, 0, 3)
  gl.disableVertexAttribArray(aPosRim)

  // 点阵与经纬网共用主 program / points and graticule share the main program
  gl.useProgram(progMain)
  gl.uniform1f(locMain.uYaw, yaw)
  gl.uniform1f(locMain.uLeanYaw, lean.x * 0.3 * strength * uni.lean)
  gl.uniform1f(locMain.uLeanTilt, -lean.y * 0.26 * strength * uni.lean)
  gl.uniform1f(locMain.uDist, CAM_DIST)
  gl.uniform1f(locMain.uFit, SPHERE_FIT * CAM_DIST)
  gl.uniform1f(locMain.uDotPx, uni.dotPx * (bh / 700)) // 700px 画布基准尺寸 / 700px baseline
  gl.uniform1f(locMain.uTime, drift)
  gl.uniform1f(locMain.uIntensity, uni.intensity)
  gl.uniform3f(locMain.uLand, uni.land[0], uni.land[1], uni.land[2])
  gl.uniform3f(locMain.uGrid, uni.grid[0], uni.grid[1], uni.grid[2])
  gl.uniform3f(locMain.uC0, uni.c0[0], uni.c0[1], uni.c0[2])
  gl.uniform3f(locMain.uC1, uni.c1[0], uni.c1[1], uni.c1[2])
  gl.uniform3f(locMain.uC2, uni.c2[0], uni.c2[1], uni.c2[2])
  gl.enableVertexAttribArray(aPosMain)

  // 经纬网（1px 线，a_seed 用常量）/ graticule (1px lines, a_seed as a constant)
  gl.uniform1f(locMain.uMode, 1)
  gl.disableVertexAttribArray(aSeedMain)
  gl.vertexAttrib1f(aSeedMain, 0.5)
  gl.bindBuffer(gl.ARRAY_BUFFER, bufGrid)
  gl.vertexAttribPointer(aPosMain, 3, gl.FLOAT, false, 0, 0)
  gl.drawArrays(gl.LINES, 0, gridCount)

  // 陆地点阵（单次 draw call）/ land dots (single draw call)
  if (pointsReady) {
    gl.uniform1f(locMain.uMode, 0)
    gl.bindBuffer(gl.ARRAY_BUFFER, bufPoints)
    gl.enableVertexAttribArray(aSeedMain)
    gl.vertexAttribPointer(aPosMain, 3, gl.FLOAT, false, 16, 0)
    gl.vertexAttribPointer(aSeedMain, 1, gl.FLOAT, false, 16, 12)
    gl.drawArrays(gl.POINTS, 0, pointCount)
  }
}

const step = (dt) => {
  // 指针姿态偏转双指数趋近（CloudSky 同款公式）；离开缓归中性
  // Pointer attitude eases exponentially (CloudSky's formula); drifts home on leave
  const k = 1 - Math.exp(-40 * 0.12 * dt)
  lean.x += ((ptr.inside ? ptr.x : 0) - lean.x) * k
  lean.y += ((ptr.inside ? ptr.y : 0) - lean.y) * k
  strength += ((ptr.inside ? 1 : 0) - strength) * (1 - Math.exp(-2.5 * dt))

  // 时间累加而非 now*speed：调整 speed 不跳变；取模防精度劣化
  // Accumulate time instead of now*speed (no jump when speed changes); wrap to avoid precision decay
  yaw = (yaw + dt * uni.speed * 0.13) % (Math.PI * 2)
  drift = (drift + dt * uni.speed * 0.55) % 1000
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
  uni.speed = clampN(props.speed, 0, 100) / 100
  // 亮色下压低亮度 / tame brightness on light theme
  uni.intensity = (clampN(props.intensity, 0, 100) / 100) * (dark ? 1.0 : 0.85)
  uni.dotPx = (clampN(props.dotSize, 0, 100) / 50) * 3 // 50 → 700px 画布下 3px / 50 → 3px at the 700px baseline
  uni.lean = clampN(props.lean, 0, 100) / 100
  uni.land = parseColor(pal.land, [0.84, 0.89, 1, 1])
  uni.grid = parseColor(pal.grid, [0.24, 0.37, 0.62, 1])
  uni.rim = parseColor(pal.rim, [0.17, 0.37, 1, 1])
  uni.c0 = parseColor(pal.c0, [0.17, 0.37, 1, 1])
  uni.c1 = parseColor(pal.c1, [0.12, 0.53, 0.9, 1])
  uni.c2 = parseColor(pal.c2, [0.01, 0.66, 0.96, 1])
  redrawIfIdle()
})

// 指针姿态：window 级监听，相对视口中心取偏转目标
// Pointer attitude: window-level listeners, tilt target relative to viewport center
const onPointerMove = (e) => {
  ptr.x = e.clientX / Math.max(1, window.innerWidth) - 0.5
  ptr.y = e.clientY / Math.max(1, window.innerHeight) - 0.5
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
  // alpha:true 画布用预乘混合叠在极光流光之上（区别于 AuroraRibbons 的自绘底色）
  // alpha:true canvas composited over the aurora via premultiplied blending (unlike AuroraRibbons' own base fill)
  gl = canvas.getContext('webgl', {
    alpha: true,
    antialias: true, // 1px 经纬线需要 AA / 1px graticule lines want AA
    depth: false,
    powerPreference: 'low-power'
  })
  if (!gl) return // WebGL 不可用时静默退场（纯装饰）/ no WebGL: quietly no-op (pure decoration)

  progMain = link(gl, VERT_SRC, FRAG_SRC)
  progRim = link(gl, RIM_VERT_SRC, RIM_FRAG_SRC)
  if (!progMain || !progRim) {
    gl = null
    return
  }
  gl.useProgram(progMain)
  locMain = cacheLocs(progMain, ['uYaw', 'uLeanYaw', 'uLeanTilt', 'uDist', 'uFit', 'uDotPx', 'uTime', 'uMode', 'uIntensity', 'uLand', 'uGrid', 'uC0', 'uC1', 'uC2'])
  aPosMain = gl.getAttribLocation(progMain, 'a_pos')
  aSeedMain = gl.getAttribLocation(progMain, 'a_seed')
  gl.useProgram(progRim)
  locRim = cacheLocs(progRim, ['uIntensity', 'uRim'])
  aPosRim = gl.getAttribLocation(progRim, 'a_pos')

  bufPoints = gl.createBuffer()
  bufGrid = gl.createBuffer()
  bufTri = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, bufTri)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW)
  buildGrid()

  // 陆地标定异步解码 → 采样出点阵；失败则仅保留经纬网与光晕
  // Land mask decodes async → sample into dots; on failure keep grid + halo only
  const img = new Image()
  img.onload = () => {
    if (!gl) return
    const c = document.createElement('canvas')
    c.width = img.width
    c.height = img.height
    const ctx = c.getContext('2d', { willReadFrequently: true })
    ctx.drawImage(img, 0, 0)
    buildPoints(ctx.getImageData(0, 0, c.width, c.height).data, c.width, c.height)
    redrawIfIdle()
  }
  img.src = landMaskUrl

  reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  coarsePtr = window.matchMedia('(hover: none)').matches
  dprCap = coarsePtr ? 1.25 : MAX_DPR

  window.addEventListener('resize', onResize, { passive: true })
  // 触屏与减动效环境不挂指针监听：星球仅自转
  // Touch and reduced-motion skip pointer listeners: the globe only auto-rotates
  if (!reduced && !coarsePtr) {
    window.addEventListener('pointermove', onPointerMove, { passive: true })
    window.addEventListener('blur', onPointerGone)
    document.documentElement.addEventListener('mouseleave', onPointerGone)
  }

  // 主题跟随
  // Theme tracking
  themeMo = new MutationObserver(onThemeChange)
  themeMo.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  // 立即呈现一帧；视口外自动停帧
  // Draw one frame up front; auto-pause off-screen
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
  progMain = null
  progRim = null
  bufPoints = null
  bufGrid = null
  bufTri = null
})
</script>

<style scoped lang="scss">
.dotted-globe {
  width: 100%;
  height: 100%;
  overflow: hidden;
  pointer-events: none;
}

.dotted-globe-canvas {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
