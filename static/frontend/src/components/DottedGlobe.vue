<template>
  <!-- 点阵地球（WebGL 着色器渲染；originkit Globe 的自研移植：自转、拖拽惯性、悬停暂停、位置标记） -->
  <!-- Dotted globe (WebGL; an original port of originkit's Globe: rotation, drag momentum, hover pause, location marker) -->
  <div ref="wrapRef" class="dotted-globe" role="img" aria-label="点阵地球：自转中，可拖拽旋转，光点标记当前位置">
    <canvas ref="canvasRef" class="dotted-globe-canvas"></canvas>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watchEffect } from 'vue'
import landMaskUrl from '@/assets/images/land-mask.png'
import { OUTLINE_SEG_COUNT, OUTLINE_SEGS_B64 } from './data/globe-outlines.js'

// 陆地标定：Natural Earth 110m land（公有领域）烘焙的等距圆柱 PNG，运行时采样出陆地点
// Land mask: equirectangular PNG baked from Natural Earth 110m land (public domain), sampled for land points

// ===== 观感常量 =====
// Look-and-feel constants
const MAX_DPR = 1.5 // 点/线渲染轻量，可承受 1.5 / points+lines are light enough for 1.5
const CAM_DIST = 2.7 // 相机距离（球半径 = 1）/ camera distance in sphere radii
const YAW0 = 0.4 // 初始经度对准 -23°（originkit initialLongitude 默认）/ face lon -23° (originkit initialLongitude)
const TILT0 = 0.4014 // 初始俯仰 23°（originkit initialLatitude 默认）/ initial pitch 23° (originkit initialLatitude)
const HALF_PI = Math.PI / 2
const SPHERE_FIT = 0.88 // 球半径占画布半宽；留 12% 边距让光晕自然衰减，防画布裁切出矩形边 / sphere radius in clip units; 12% margin so the halo decays before the canvas edge
const POINT_CANDIDATES = 42000 // Fibonacci 候选数（约 1/3 落在陆地）/ Fibonacci candidates (~1/3 on land)
const GA = Math.PI * (3 - Math.sqrt(5)) // 黄金角分布 / golden-angle spacing

const VERT_SRC = `
attribute vec3 a_pos;
attribute float a_seed;

uniform float uYaw;      // 自转相位（含拖拽/惯性）/ rotation phase (drag & momentum included)
uniform float uPitch;    // 俯仰相位（含拖拽/惯性）/ pitch phase (drag & momentum included)
uniform float uDist;     // 相机距离 / camera distance
uniform float uFit;      // 球半径 → 裁剪空间缩放 / sphere radius → clip scale
uniform float uDotPx;    // 点基准像素尺寸 / base point size in px
uniform float uTime;     // 色源光波相位 / color-source wave phase
uniform float uMode;     // 0=点阵 1=经纬线 2=位置标记 / 0=points 1=graticule 2=location marker

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
  // 自转 + 俯仰（均含拖拽/惯性，originkit 的双向 drag-to-spin；无深度测试，背面仅变暗）
  // Yaw + pitch (both with drag & momentum, originkit's two-axis drag-to-spin; no depth test; back side just dims)
  vec3 n = a_pos;
  vec3 p = rotX(rotY(n, uYaw), uPitch);
  float w = uDist - p.z;

  vFacing = p.z;
  vSeed = a_seed;
  vGlow = vec3(0.0);

  if (uMode < 0.5) {
    float g0 = srcGlow(n, 0.0, uTime);
    float g1 = srcGlow(n, 1.0, uTime);
    float g2 = srcGlow(n, 2.0, uTime);
    vGlow = vec3(g0, g1, g2);
    // 光波处轻微抬升，剪影上可见波纹 / lift at glow waves so they read on the silhouette
    p *= 1.0 + (g0 + g1 + g2) * 0.012;
    w = uDist - p.z;
    gl_PointSize = uDotPx * (uDist / w) * (0.82 + 0.36 * a_seed);
  } else if (uMode > 1.5) {
    // 位置标记：稍抬离球面，放大成精灵 / location marker: lifted off the surface, sized as a sprite
    p *= 1.004;
    w = uDist - p.z;
    gl_PointSize = uDotPx * 7.0 * (uDist / w);
  } else {
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
uniform float uTime;       // 标记脉冲相位 / marker pulse phase
uniform float uMode;
uniform vec3 uLand, uGrid, uOutline, uMarker, uC0, uC1, uC2;

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
  } else if (uMode > 2.5) {
    // 国界描线：originkit 的白色轮廓（独立配色、不随点阵亮度缩放）
    // Country outlines: originkit's white lines (own color, not scaled by dot brightness)
    col = uOutline;
    a = facing * 0.7;
  } else if (uMode > 1.5) {
    // 位置标记：实心核 + 固定细环 + 扩散脉冲，三层叠加（勿对负底数用 pow）
    // Marker: solid core + fixed thin ring + expanding pulse (no pow on negative base)
    vec2 q = gl_PointCoord - 0.5;
    float r = length(q) * 2.0;
    float core = smoothstep(0.30, 0.20, r);
    float e = (r - 0.52) * 14.0;
    float fix = exp(-e * e) * 0.6;
    float ph = fract(uTime * 2.7);
    float d = (r - ph * 0.95) * 9.0;
    float pulse = exp(-d * d) * (1.0 - ph);
    col = uMarker;
    a = (core + fix + pulse * 0.85) * facing; // 不随 intensity 缩放：信息标记始终清晰 / informational marker stays crisp
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

uniform float uFit; // 球半径（裁剪空间）/ sphere radius in clip units
uniform float uIntensity;
uniform vec3 uRim;
varying vec2 vP;

void main(){
  float r = length(vP);
  // 球缘光晕：指数急衰减，画布边缘处幅值已不可见（高斯宽尾会被画布裁出矩形边）
  // Limb halo: steep exponential decay, invisible at the canvas edge (a wide gaussian tail gets cropped into rectangle edges)
  float halo = exp(-abs(r - uFit) * 26.0);
  float a = halo * 0.55 * uIntensity;
  gl_FragColor = vec4(uRim * a, a); // 预乘 over / premultiplied over
}
`

// ===== 主题配色（点阵与光波取品牌三色） =====
// Theme palettes (dots and waves use the brand gradient hues)
const LIGHT_PALETTE = {
  land: '#2b52c8',
  grid: '#8fb0e8',
  outline: '#33549e',
  rim: '#2B5EFF',
  marker: '#0288d1',
  c0: '#2B5EFF',
  c1: '#1E88E5',
  c2: '#03A9F4'
}
const DARK_PALETTE = {
  land: '#d5e4ff',
  grid: '#3d5f9e',
  outline: '#dbe8ff',
  rim: '#2B5EFF',
  marker: '#8bd4ff',
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
  speed: 0.32, intensity: 0.58, dotPx: 2.8,
  land: [0.84, 0.89, 1, 1], grid: [0.24, 0.37, 0.62, 1], rim: [0.17, 0.37, 1, 1],
  outline: [0.86, 0.91, 1, 1], marker: [0.55, 0.83, 1, 1],
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
let bufOutline = null
let bufTri = null
let pointCount = 0
let gridCount = 0
let outlineCount = 0
let pointsReady = false
let reduced = false
let coarsePtr = false
let dprCap = MAX_DPR
let inView = false
let running = false
let raf = 0
let last = 0
let yaw = YAW0
let pitch = TILT0 // 俯仰角，拖拽上下调整 / pitch, adjusted by vertical drag
let drift = 0
// 拖拽旋转（originkit 的 drag-to-spin + 惯性）/ drag-to-spin with momentum, like originkit
let dragging = false
let yawVel = 0 // 释放后的惯性角速度 / post-release angular momentum
let pitchVel = 0 // 俯仰惯性角速度 / pitch angular momentum
let hovering = false // stopOnHover：悬停暂停自转 / pause auto-rotation while hovered
let rotBlend = 1 // 自转权重（悬停/拖拽时缓落 0）/ auto-rotation weight, eases to 0
let lastDragX = 0
let lastDragY = 0
let lastDragT = 0
// 当前位置标记 / current-location marker
let markerPos = null // 单位球坐标 [x,y,z] / unit-sphere coords
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

// 国界描线：解码烘焙数据（量化 Uint16 → 单位球坐标；lonQ 最大 36000 超出 Int16，勿改回）
// Country outlines: decode baked data (quantized Uint16 → unit-sphere coords; lonQ peaks at 36000, past Int16 range)
const buildOutlines = () => {
  const bin = atob(OUTLINE_SEGS_B64)
  const bytes = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
  const q = new Uint16Array(bytes.buffer)
  const verts = new Float32Array(OUTLINE_SEG_COUNT * 6)
  const d2r = Math.PI / 180
  for (let i = 0; i < OUTLINE_SEG_COUNT; i++) {
    const v1 = ll2v((q[i * 4] / 100 - 180) * d2r, (q[i * 4 + 1] / 100 - 90) * d2r)
    const v2 = ll2v((q[i * 4 + 2] / 100 - 180) * d2r, (q[i * 4 + 3] / 100 - 90) * d2r)
    verts.set(v1, i * 6)
    verts.set(v2, i * 6 + 3)
  }
  outlineCount = OUTLINE_SEG_COUNT * 2
  gl.bindBuffer(gl.ARRAY_BUFFER, bufOutline)
  gl.bufferData(gl.ARRAY_BUFFER, verts, gl.STATIC_DRAW)
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
  gl.uniform1f(locRim.uFit, SPHERE_FIT)
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
  gl.uniform1f(locMain.uPitch, pitch)
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

  // 国界描线（originkit 的白色轮廓）/ country outlines (originkit's white lines)
  gl.uniform1f(locMain.uMode, 3)
  gl.uniform3f(locMain.uOutline, uni.outline[0], uni.outline[1], uni.outline[2])
  gl.bindBuffer(gl.ARRAY_BUFFER, bufOutline)
  gl.vertexAttribPointer(aPosMain, 3, gl.FLOAT, false, 0, 0)
  gl.drawArrays(gl.LINES, 0, outlineCount)

  // 陆地点阵（单次 draw call）/ land dots (single draw call)
  if (pointsReady) {
    gl.uniform1f(locMain.uMode, 0)
    gl.bindBuffer(gl.ARRAY_BUFFER, bufPoints)
    gl.enableVertexAttribArray(aSeedMain)
    gl.vertexAttribPointer(aPosMain, 3, gl.FLOAT, false, 16, 0)
    gl.vertexAttribPointer(aSeedMain, 1, gl.FLOAT, false, 16, 12)
    gl.drawArrays(gl.POINTS, 0, pointCount)
  }

  // 当前位置标记（单顶点常量属性，无需缓冲）/ location marker (single vertex via constant attributes, no buffer)
  if (markerPos) {
    gl.uniform1f(locMain.uMode, 2)
    gl.uniform3f(locMain.uMarker, uni.marker[0], uni.marker[1], uni.marker[2])
    gl.disableVertexAttribArray(aPosMain)
    gl.vertexAttrib3f(aPosMain, markerPos[0], markerPos[1], markerPos[2])
    gl.vertexAttrib1f(aSeedMain, 0.9)
    gl.drawArrays(gl.POINTS, 0, 1)
  }
}

const step = (dt) => {
  // 自转：悬停/拖拽时缓停（stopOnHover）；释放后惯性指数衰减
  // Rotation: eases to a stop on hover/drag (stopOnHover); momentum decays exponentially after release
  rotBlend += (((dragging || hovering) ? 0 : 1) - rotBlend) * (1 - Math.exp(-3 * dt))
  if (!dragging) {
    // direction: left（originkit 默认）→ 表面向左移动 / surface drifts left (originkit's default direction)
    yaw = (yaw + (yawVel - uni.speed * 0.13 * rotBlend) * dt) % (Math.PI * 2)
    yawVel *= Math.exp(-2.2 * dt)
    if (pitchVel !== 0) {
      pitch = clampN(pitch + pitchVel * dt, -HALF_PI, HALF_PI)
      if (pitch === -HALF_PI || pitch === HALF_PI) {
        pitchVel = 0 // 抵达极点即停 / stop at the poles
      } else {
        pitchVel *= Math.exp(-2.2 * dt)
      }
    }
  }

  // 时间累加而非 now*speed：调整 speed 不跳变；取模防精度劣化
  // Accumulate time instead of now*speed (no jump when speed changes); wrap to avoid precision decay
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
  // 亮色下略压亮度 / slightly tame brightness on light theme
  uni.intensity = (clampN(props.intensity, 0, 100) / 100) * (dark ? 1.0 : 0.95)
  uni.dotPx = (clampN(props.dotSize, 0, 100) / 50) * 3 // 50 → 700px 画布下 3px / 50 → 3px at the 700px baseline
  uni.land = parseColor(pal.land, [0.84, 0.89, 1, 1])
  uni.grid = parseColor(pal.grid, [0.24, 0.37, 0.62, 1])
  uni.rim = parseColor(pal.rim, [0.17, 0.37, 1, 1])
  uni.outline = parseColor(pal.outline, [0.86, 0.91, 1, 1])
  uni.marker = parseColor(pal.marker, [0.55, 0.83, 1, 1])
  uni.c0 = parseColor(pal.c0, [0.17, 0.37, 1, 1])
  uni.c1 = parseColor(pal.c1, [0.12, 0.53, 0.9, 1])
  uni.c2 = parseColor(pal.c2, [0.01, 0.66, 0.96, 1])
  redrawIfIdle()
})

// 拖拽旋转：按住拖动 → 球面跟随；释放 → 惯性滚动（originkit 的 drag-to-spin）
// Drag-to-spin: the surface follows the pointer; momentum on release (originkit's drag-to-spin)
const onPointerDown = (e) => {
  if (e.pointerType === 'mouse' && e.button !== 0) return
  dragging = true
  yawVel = 0
  pitchVel = 0
  lastDragX = e.clientX
  lastDragY = e.clientY
  lastDragT = performance.now()
  wrapRef.value?.classList.add('is-dragging')
  window.addEventListener('pointermove', onDragMove)
  window.addEventListener('pointerup', onDragEnd)
  window.addEventListener('pointercancel', onDragEnd)
}
const onDragMove = (e) => {
  if (!dragging || !gl) return
  const el = wrapRef.value
  if (!el) return
  // 按球面像素半径换算角位移：表面跟随光标（横→自转，纵→俯仰）/ angular steps from the sphere's pixel radius: x→yaw, y→pitch
  const radiusPx = Math.max(1, (el.clientWidth * SPHERE_FIT) / 2)
  const dyaw = (e.clientX - lastDragX) / radiusPx
  const dpitch = (e.clientY - lastDragY) / radiusPx
  yaw += dyaw
  const prevPitch = pitch
  pitch = clampN(pitch + dpitch, -HALF_PI, HALF_PI)
  const now = performance.now()
  const dt = Math.max(0.001, (now - lastDragT) / 1000)
  // 指数平滑采样角速度作为释放惯性 / exponentially-smoothed angular velocities → release momentum
  yawVel = yawVel * 0.7 + (dyaw / dt) * 0.3
  pitchVel = pitchVel * 0.7 + ((pitch - prevPitch) / dt) * 0.3
  lastDragX = e.clientX
  lastDragY = e.clientY
  lastDragT = now
  if (!running) drawOnce() // 减动效下拖拽也即时呈现 / drag still renders instantly under reduced motion
}
const onDragEnd = () => {
  dragging = false
  yawVel = clampN(yawVel, -3, 3)
  pitchVel = clampN(pitchVel, -3, 3)
  wrapRef.value?.classList.remove('is-dragging')
  window.removeEventListener('pointermove', onDragMove)
  window.removeEventListener('pointerup', onDragEnd)
  window.removeEventListener('pointercancel', onDragEnd)
}
// stopOnHover：悬停暂停自转，离开恢复（仅精确指针）/ stopOnHover: pause auto-rotation while hovered (fine pointers only)
const onHoverEnter = (e) => {
  if (e.pointerType === 'mouse') hovering = true
}
const onHoverLeave = (e) => {
  if (e.pointerType === 'mouse') hovering = false
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

// ===== 当前位置标记 =====
// Current-location marker
// 时区 → 城市坐标（Geolocation 被拒/不可用时的兜底）/ timezone → city coords (fallback when geolocation is denied/unavailable)
const TZ_COORDS = {
  'Asia/Shanghai': [31.23, 121.47], 'Asia/Urumqi': [43.83, 87.62],
  'Asia/Hong_Kong': [22.32, 114.17], 'Asia/Macau': [22.2, 113.55],
  'Asia/Taipei': [25.03, 121.53], 'Asia/Tokyo': [35.68, 139.69],
  'Asia/Seoul': [37.57, 126.98], 'Asia/Singapore': [1.35, 103.82],
  'Asia/Bangkok': [13.76, 100.5], 'Asia/Jakarta': [-6.21, 106.85],
  'Asia/Kuala_Lumpur': [3.14, 101.69], 'Asia/Manila': [14.6, 120.98],
  'Asia/Ho_Chi_Minh': [10.82, 106.63], 'Asia/Kolkata': [28.61, 77.21],
  'Asia/Karachi': [24.86, 67.01], 'Asia/Dubai': [25.2, 55.27],
  'Asia/Riyadh': [24.71, 46.68], 'Asia/Tehran': [35.69, 51.39],
  'Europe/Istanbul': [41.01, 28.98], 'Europe/London': [51.51, -0.13],
  'Europe/Paris': [48.86, 2.35], 'Europe/Berlin': [52.52, 13.4],
  'Europe/Moscow': [55.76, 37.62], 'Europe/Madrid': [40.42, -3.7],
  'Europe/Rome': [41.9, 12.5], 'Europe/Amsterdam': [52.37, 4.9],
  'America/New_York': [40.71, -74.01], 'America/Chicago': [41.88, -87.63],
  'America/Denver': [39.74, -104.99], 'America/Los_Angeles': [34.05, -118.24],
  'America/Vancouver': [49.28, -123.12], 'America/Toronto': [43.65, -79.38],
  'America/Mexico_City': [19.43, -99.13], 'America/Sao_Paulo': [-23.55, -46.63],
  'America/Argentina/Buenos_Aires': [-34.6, -58.38],
  'Australia/Sydney': [-33.87, 151.21], 'Australia/Perth': [-31.95, 115.86],
  'Pacific/Auckland': [-36.85, 174.76], 'Africa/Cairo': [30.04, 31.24],
  'Africa/Johannesburg': [-26.2, 28.05], 'Africa/Lagos': [6.52, 3.38],
  'Africa/Nairobi': [-1.29, 36.82]
}
const setMarkerLatLon = (lat, lon) => {
  markerPos = ll2v((lon * Math.PI) / 180, (lat * Math.PI) / 180)
  redrawIfIdle()
}
const locateByTimezone = () => {
  try {
    const c = TZ_COORDS[Intl.DateTimeFormat().resolvedOptions().timeZone]
    if (c) setMarkerLatLon(c[0], c[1])
  } catch (err) {
    // 无时区信息则不显示标记 / no timezone info: skip the marker
  }
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
  locMain = cacheLocs(progMain, ['uYaw', 'uPitch', 'uDist', 'uFit', 'uDotPx', 'uTime', 'uMode', 'uIntensity', 'uLand', 'uGrid', 'uOutline', 'uMarker', 'uC0', 'uC1', 'uC2'])
  aPosMain = gl.getAttribLocation(progMain, 'a_pos')
  aSeedMain = gl.getAttribLocation(progMain, 'a_seed')
  gl.useProgram(progRim)
  locRim = cacheLocs(progRim, ['uFit', 'uIntensity', 'uRim'])
  aPosRim = gl.getAttribLocation(progRim, 'a_pos')

  bufPoints = gl.createBuffer()
  bufGrid = gl.createBuffer()
  bufOutline = gl.createBuffer()
  bufTri = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, bufTri)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW)
  buildGrid()
  buildOutlines()

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

  // 当前位置：Geolocation 授权优先，被拒/不可用回退时区推断
  // Current location: geolocation first; timezone inference when denied/unavailable
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition(
      (pos) => setMarkerLatLon(pos.coords.latitude, pos.coords.longitude),
      locateByTimezone,
      { timeout: 8000, maximumAge: 600000 }
    )
  } else {
    locateByTimezone()
  }

  reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  coarsePtr = window.matchMedia('(hover: none)').matches
  dprCap = coarsePtr ? 1.25 : MAX_DPR

  window.addEventListener('resize', onResize, { passive: true })
  // 拖拽旋转随时可用（含触屏与减动效）；悬停暂停仅对精确指针生效
  // Drag is always available (touch & reduced motion included); hover-pause applies to fine pointers only
  const el = wrapRef.value
  el.addEventListener('pointerdown', onPointerDown)
  el.addEventListener('pointerenter', onHoverEnter)
  el.addEventListener('pointerleave', onHoverLeave)

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
  const el = wrapRef.value
  el?.removeEventListener('pointerdown', onPointerDown)
  el?.removeEventListener('pointerenter', onHoverEnter)
  el?.removeEventListener('pointerleave', onHoverLeave)
  window.removeEventListener('resize', onResize)
  onDragEnd() // 卸载时若在拖拽，解绑 window 监听 / unmount mid-drag: unbind the window listeners
  // 释放 WebGL 上下文 / release the GL context
  gl?.getExtension('WEBGL_lose_context')?.loseContext()
  gl = null
  progMain = null
  progRim = null
  bufPoints = null
  bufGrid = null
  bufOutline = null
  bufTri = null
})
</script>

<style scoped lang="scss">
.dotted-globe {
  width: 100%;
  height: 100%;
  overflow: hidden;
  cursor: grab; // 可拖拽旋转 / draggable
  touch-action: none; // 触屏拖拽归地球而不是页面 / touch drags spin the globe, not the page

  &.is-dragging {
    cursor: grabbing;
  }
}

.dotted-globe-canvas {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
