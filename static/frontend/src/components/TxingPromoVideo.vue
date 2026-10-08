<template>
  <!--
    Txing AI 项目介绍影片（代码驱动的动态图形宣传片）
    Code-driven motion-graphics promo film: 8 scenes / 55s, rAF clock + canvas FX + CSS choreography.
    播放器外壳：自动播放 / 暂停 / 分段进度(可跳场景) / 重播 / 全屏，滚出视口自动暂停。
  -->
  <div
    ref="rootRef"
    class="tpv"
    tabindex="0"
    role="region"
    aria-label="Txing AI 智能对话平台 · 项目介绍影片"
    :class="{ 'is-paused': started && !playing && !ended, 'is-ended': ended, 'ctrl-hide': ctrlHide }"
    @keydown="onKey"
    @mouseenter="ctrlWake"
    @mousemove="ctrlWake"
    @mouseleave="ctrlLease"
  >
    <!-- 1280×720 舞台，整体 transform scale 适配容器宽度 -->
    <!-- 1280x720 stage, scaled as a whole to fit the container width -->
    <div ref="stageRef" class="tpv-stage">
      <!-- 背景光晕（缓慢漂移） -->
      <!-- Slow-drifting background glows -->
      <div class="tpv-bg" aria-hidden="true">
        <i class="tpv-glow g1"></i>
        <i class="tpv-glow g2"></i>
        <i class="tpv-glow g3"></i>
      </div>

      <!-- 特效画布：粒子星野 / 渠道网络流 / Agent 脉冲 -->
      <!-- FX canvas: star particles / gateway network flow / agent pulses -->
      <canvas ref="canvasRef" class="tpv-canvas" aria-hidden="true"></canvas>

      <!-- ══════════ S1 · 开场 ══════════ -->
      <section class="tpv-scene s-open">
        <div class="so-zoom">
          <p class="tpv-kicker so-k">TXING AI · PROJECT FILM</p>
          <h2 class="so-mark"><span class="so-mark-txt">TXING&nbsp;AI</span></h2>
          <div class="so-line"></div>
          <p class="so-sub">智 能 对 话 平 台</p>
        </div>
        <div class="so-foot">
          <span class="so-fl">开源 · 全栈 · 高性能</span>
          <span class="so-fr">PROJECT&nbsp;FILM</span>
        </div>
      </section>

      <!-- ══════════ S2 · 理念与数字 ══════════ -->
      <section class="tpv-scene s-vision">
        <div class="sv-wrap">
          <p class="tpv-kicker sv-k">为什么做它</p>
          <h2 class="sv-h">
            <span class="sv-mask"><span class="sv-l1">一个入口，</span></span>
            <span class="sv-mask"><span class="sv-l2">连接所有大模型</span></span>
          </h2>
          <p class="sv-sub">Go + Vue3 全栈开源 · 从对话到工作流的完整 AI 基础设施</p>
          <div class="sv-stats">
            <div class="sv-stat">
              <span class="sv-num" data-n="24000">0</span>
              <span class="sv-lab">后端代码行</span>
            </div>
            <i class="sv-div"></i>
            <div class="sv-stat">
              <span class="sv-num" data-n="37000">0</span>
              <span class="sv-lab">前端代码行</span>
            </div>
            <i class="sv-div"></i>
            <div class="sv-stat">
              <span class="sv-num sv-flat">100%</span>
              <span class="sv-lab">开源</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ══════════ S3 · 多供应商渠道网关 ══════════ -->
      <section class="tpv-scene s-gate">
        <header class="tpv-head">
          <p class="tpv-kicker">01 / 多供应商渠道网关</p>
          <h2 class="tpv-hl">统一接入 · 智能选路</h2>
        </header>
        <div class="sg-graph">
          <span class="sg-node sg-src">用户请求</span>
          <span class="sg-node sg-core"><i class="sg-core-dot"></i>渠道网关</span>
          <span
            class="sg-node sg-prov"
            v-for="(p, i) in providers"
            :key="p"
            :style="{ left: '1000px', top: sgProvTop[i] + 'px' }"
          >{{ p }}</span>
        </div>
        <div class="tpv-chips sg-chips">
          <span class="tpv-chip">条件模型映射</span>
          <span class="tpv-chip">三层过滤选路</span>
          <span class="tpv-chip">优先级 · 权重</span>
          <span class="tpv-chip">故障启停降级</span>
        </div>
      </section>

      <!-- ══════════ S4 · WebSocket 流式对话 ══════════ -->
      <section class="tpv-scene s-chat">
        <header class="tpv-head">
          <p class="tpv-kicker">02 / WebSocket 流式对话</p>
          <h2 class="tpv-hl">逐 Token 输出 · 实时渲染</h2>
        </header>
        <div class="sc-win">
          <div class="sc-bar">
            <i></i><i></i><i></i>
            <span class="sc-bar-title">Txing AI · 新会话</span>
          </div>
          <div class="sc-msgs">
            <div class="sc-row sc-row-user">
              <div class="sc-bubble sc-user"><span ref="scUserRef"></span><i class="sc-caret"></i></div>
            </div>
            <div class="sc-row sc-row-ai">
              <span class="sc-avatar">T</span>
              <div class="sc-bubble sc-ai">
                <span ref="scAiRef"></span>
                <pre ref="scCodeRef" class="sc-code"></pre>
              </div>
            </div>
          </div>
          <div class="sc-input">
            <span class="sc-wave"><i v-for="i in 28" :key="i"></i></span>
            <span ref="scStopRef" class="sc-stop"><i></i>停止生成</span>
          </div>
        </div>
        <div class="tpv-chips sc-chips">
          <span class="tpv-chip">全双工通道</span>
          <span class="tpv-chip">Markdown 实时渲染</span>
          <span class="tpv-chip">上下文窗口管理</span>
          <span class="tpv-chip">模型参数 · 聊天预设</span>
        </div>
      </section>
      <!-- ══════════ S5 · 可视化工作流 ══════════ -->
      <section class="tpv-scene s-flow">
        <header class="tpv-head">
          <p class="tpv-kicker">03 / 可视化工作流引擎</p>
          <h2 class="tpv-hl">七类节点 · 双重校验 · 全程可视</h2>
        </header>
        <div class="sf-graph">
          <svg viewBox="0 0 880 508" width="880" height="508" class="sf-svg" aria-hidden="true">
            <defs>
              <linearGradient id="tpvFlowGrad" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0" stop-color="#4da3ff" />
                <stop offset="1" stop-color="#a855f7" />
              </linearGradient>
            </defs>
            <path class="sf-edge" d="M 108 240 C 160 240, 178 240, 230 240" />
            <path class="sf-edge" d="M 278 240 C 330 240, 388 130, 440 130" />
            <path class="sf-edge" d="M 278 240 C 330 240, 388 350, 440 350" />
            <path class="sf-edge" d="M 488 130 C 540 130, 588 240, 640 240" />
            <path class="sf-edge" d="M 488 350 C 540 350, 588 240, 640 240" />
            <path class="sf-edge" d="M 688 240 C 740 240, 768 240, 820 240" />
            <circle class="sf-dot" r="5" />
            <circle class="sf-dot" r="5" />
            <circle class="sf-dot" r="5" />
            <circle class="sf-dot" r="5" />
            <circle class="sf-dot" r="5" />
            <circle class="sf-dot" r="5" />
          </svg>
          <div class="sf-node" style="left: 12px;  top: 217px"><i></i><span>开始</span></div>
          <div class="sf-node sf-llm" style="left: 182px; top: 217px"><i></i><span>LLM</span></div>
          <div class="sf-node" style="left: 392px; top: 107px"><i></i><span>HTTP</span></div>
          <div class="sf-node" style="left: 392px; top: 327px"><i></i><span>Code</span></div>
          <div class="sf-node" style="left: 592px; top: 217px"><i></i><span>Join</span></div>
          <div class="sf-node" style="left: 772px; top: 217px"><i></i><span>结束</span></div>
        </div>
        <aside class="sf-log">
          <p class="sf-log-title">执行日志</p>
          <p class="sf-log-time"><span ref="sfElapsedRef">0.00</span><em>s</em></p>
          <ul>
            <li v-for="(l, i) in flowLogs" :key="i" :style="{ animationDelay: l[1] + 's' }">
              <b>{{ l[0] }}</b><span>{{ l[2] }}</span>
            </li>
          </ul>
          <p class="sf-log-done">✓ 执行完成 · 全节点通过</p>
        </aside>
        <div class="tpv-chips sf-chips">
          <span class="tpv-chip">LLM · Code · HTTP</span>
          <span class="tpv-chip">条件 · 并行 · Join · 子工作流</span>
          <span class="tpv-chip">节点级失败重试</span>
          <span class="tpv-chip">执行时间线回放</span>
        </div>
      </section>

      <!-- ══════════ S6 · Agent 与工具体系 ══════════ -->
      <section class="tpv-scene s-agent">
        <header class="tpv-head">
          <p class="tpv-kicker">04 / Agent 与工具体系</p>
          <h2 class="tpv-hl">推理 — 执行 · 完整闭环</h2>
        </header>
        <div class="sa-orbit" ref="saOrbitRef">
          <span class="sa-core"><i></i>AGENT</span>
          <span class="sa-chip" v-for="t in agentTools" :key="t">{{ t }}</span>
        </div>
        <p class="sa-sub">四类 Agent · 工具调用循环 · MCP 多服务器动态接入</p>
      </section>

      <!-- ══════════ S7 · 分层架构与技术栈 ══════════ -->
      <section class="tpv-scene s-stack">
        <header class="tpv-head">
          <p class="tpv-kicker">05 / 分层架构</p>
          <h2 class="tpv-hl">工程化 · 可扩展</h2>
        </header>
        <div class="ss-strata">
          <div class="ss-bar" v-for="(b, i) in strata" :key="b.n" :style="{ animationDelay: 0.25 + i * 0.22 + 's' }">
            <em>0{{ i + 1 }}</em>
            <b>{{ b.n }}</b>
            <span>{{ b.d }}</span>
          </div>
        </div>
        <div class="tpv-chips ss-chips">
          <span class="tpv-chip" v-for="(t, i) in techChips" :key="t" :style="{ animationDelay: 1.6 + i * 0.12 + 's' }">{{ t }}</span>
        </div>
      </section>

      <!-- ══════════ S8 · 尾板 ══════════ -->
      <section class="tpv-scene s-end">
        <div class="se-wrap">
          <p class="tpv-kicker se-k">开源项目</p>
          <h2 class="se-mark"><span class="se-mark-txt">Txing&nbsp;AI</span></h2>
          <p class="se-sub">智能对话平台 · 本站正运行于其上</p>
          <a class="se-link" href="https://github.com/lemon-puls/txing-ai" target="_blank" rel="noopener">
            <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33c.85 0 1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z"/></svg>
            <span>lemon-puls / txing-ai</span>
          </a>
        </div>
        <div class="so-foot se-foot">
          <span class="so-fl">TXING&nbsp;AI</span>
          <span class="so-fr">PROJECT&nbsp;FILM</span>
        </div>
        <i class="se-fade"></i>
      </section>

      <!-- 质感层：胶片颗粒 + 暗角 -->
      <!-- Texture: film grain + vignette -->
      <i class="tpv-grain" aria-hidden="true"></i>
      <i class="tpv-vignette" aria-hidden="true"></i>
    </div>

    <!-- 播放控制条 -->
    <!-- Player controls -->
    <div class="tpv-ctrl" @click.stop>
      <button class="tpv-btn" :aria-label="playing ? '暂停' : '播放'" @click="togglePlay">
        <svg v-if="!playing || ended" viewBox="0 0 24 24" width="19" height="19"><path fill="currentColor" d="M8 5.14v14l11-7l-11-7Z"/></svg>
        <svg v-else viewBox="0 0 24 24" width="19" height="19"><path fill="currentColor" d="M6 5h4v14H6V5Zm8 0h4v14h-4V5Z"/></svg>
      </button>
      <div class="tpv-segs" ref="segsRef">
        <button
          v-for="(s, i) in scenes"
          :key="s.id"
          class="tpv-seg"
          :data-name="s.name"
          :aria-label="'跳转到 ' + s.name"
          @click="seekToScene(i)"
        ><i class="tpv-seg-fill"></i></button>
      </div>
      <span class="tpv-scene-name" ref="sceneNameRef">开场</span>
      <span class="tpv-time"><b ref="timeCurRef">00:00</b>&nbsp;/&nbsp;{{ fmtTime(TOTAL) }}</span>
      <button class="tpv-btn" aria-label="重播" @click="replay">
        <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M12 5V1L7 6l5 5V7a6 6 0 1 1-6 6H4a8 8 0 1 0 8-8Z"/></svg>
      </button>
      <button class="tpv-btn" :aria-label="audioOn && audioLive ? '关闭音效' : '开启音效'" @click="toggleAudio">
        <svg v-if="audioOn && audioLive" viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>
        <svg v-else viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z"/></svg>
      </button>
      <button class="tpv-btn" :aria-label="isFs ? '退出全屏' : '全屏'" @click="toggleFs">
        <svg v-if="!isFs" viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M4 4h6v2H6v4H4V4Zm10 0h6v6h-2V6h-4V4ZM4 14h2v4h4v2H4v-6Zm14 0h2v6h-6v-2h4v-4Z"/></svg>
        <svg v-else viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M9 4h2v6H5V8h4V4Zm4 0h2v4h4v2h-6V4ZM5 14h6v6H9v-4H5v-2Zm10 0h4v2h-4v4h-2v-6Z"/></svg>
      </button>
    </div>

    <!-- 结束遮罩：重播 -->
    <!-- Ended overlay: replay -->
    <transition name="tpv-fade">
      <div v-if="ended" class="tpv-endmask" @click="replay">
        <span class="tpv-end-title">Txing AI · 智能对话平台</span>
        <span class="tpv-end-replay">
          <svg viewBox="0 0 24 24" width="22" height="22"><path fill="currentColor" d="M12 5V1L7 6l5 5V7a6 6 0 1 1-6 6H4a8 8 0 1 0 8-8Z"/></svg>
          重新播放
        </span>
      </div>
    </transition>
  </div>
</template>

<script setup name="TxingPromoVideo">
import { ref, onMounted, onBeforeUnmount } from 'vue'

// ───────────────────────── 影片脚本配置 ─────────────────────────
// Film script: per-scene duration (s). Total ≈ 55s.
const scenes = [
  { id: 'open',   name: '开场',       dur: 6.5 },
  { id: 'vision', name: '理念',       dur: 5.5 },
  { id: 'gate',   name: '渠道网关',   dur: 8.0 },
  { id: 'chat',   name: '流式对话',   dur: 7.5 },
  { id: 'flow',   name: '工作流',     dur: 8.0 },
  { id: 'agent',  name: 'Agent',     dur: 6.5 },
  { id: 'stack',  name: '架构与技术栈', dur: 6.0 },
  { id: 'end',    name: '尾板',       dur: 7.0 }
]
const TOTAL = scenes.reduce((s, x) => s + x.dur, 0)
const STARTS = []
scenes.reduce((acc, s) => { STARTS.push(acc); return acc + s.dur }, 0)
const STAGE_W = 1280
const STAGE_H = 720
const SCENE = Object.fromEntries(scenes.map((s, i) => [s.id, i]))

const providers = ['OpenAI 兼容', '火山引擎', 'Polo', 'Eino']
// 画布中供应商节点中心的 y 坐标；DOM 标签用同一个数组对齐
// Provider node center Y on canvas; DOM labels align via the same array
const sgProvY = [180, 326, 472, 618]
const sgProvTop = sgProvY.map((y) => y - 28)

const agentTools = ['联网搜索', '网页抓取', 'PDF 解析', '代码执行', '图片生成', 'MCP 工具']
const strata = [
  { n: '前端层', d: 'Vue3 · Element Plus · 工作流编排' },
  { n: '接入层', d: 'Gin · WebSocket 网关 · JWT 鉴权' },
  { n: '核心层', d: 'Agent · 工作流引擎 · 工具 · 渠道网关' },
  { n: '适配层', d: 'OpenAI 兼容 · 火山引擎 · Polo · Eino' },
  { n: '基础设施', d: 'MySQL · Redis · COS 对象存储' }
]
const techChips = ['🐹 Go', '💚 Vue3', '🔄 WebSocket', '🐬 MySQL', '⚡ Redis', '🪁 Eino', '🎨 Element Plus']
// [节点, 延迟s, 耗时文本]
const flowLogs = [
  ['▸ 开始', 1.0, '入口节点就绪'],
  ['▸ LLM', 1.55, '推理 · 812ms'],
  ['▸ HTTP', 2.55, '并行分支 A · 341ms'],
  ['▸ Code', 2.6, '并行分支 B · 156ms'],
  ['▸ Join', 3.6, '汇聚分支结果 · 2ms'],
  ['▸ 结束', 4.4, '输出结构化结果']
]
// 工作流边的时间窗 [起点s, 终点s]，与 sf-edge 顺序一一对应
// Edge pulse time windows, aligned with .sf-edge order
const EDGE_T = [[1.0, 1.5], [2.0, 2.5], [2.05, 2.55], [3.0, 3.5], [3.05, 3.55], [3.9, 4.4]]

// 对话场景的打字/流式脚本
// Typing & streaming script for the chat scene
const CHAT_USER = '用 Go 写一个令牌桶限流器'
const CHAT_AI = '好的，这是一个基于令牌桶的限流器核心实现：'
const CHAT_CODE = [
  'func (b *Bucket) Allow() bool {',
  '    b.fill(time.Now())',
  '    if b.tokens >= 1 {',
  '        b.tokens--',
  '        return true',
  '    }',
  '    return false',
  '}'
]

// ───────────────────────── 状态与元素 ─────────────────────────
const rootRef = ref(null)
const stageRef = ref(null)
const canvasRef = ref(null)
const segsRef = ref(null)
const sceneNameRef = ref(null)
const timeCurRef = ref(null)
const scUserRef = ref(null)
const scAiRef = ref(null)
const scCodeRef = ref(null)
const scStopRef = ref(null)
const sfElapsedRef = ref(null)
const saOrbitRef = ref(null)

const started = ref(false)
const playing = ref(false)
const ended = ref(false)
const ctrlHide = ref(false)
const isFs = ref(false)

// 每帧更新的量全部走直接 DOM/Canvas，不进响应式，保证 60fps
// Per-frame values bypass Vue reactivity (direct DOM/canvas writes) for 60fps
let t = 0                 // 全局时间 (s)
let raf = 0
let lastTs = 0
let activeIdx = -1
let ctrlTimer = 0
let autoPaused = false    // 因滚出视口/切后台而自动暂停（回来时自动续播）
let destroyed = false

let sceneEls = []
let segFills = []
let provEls = []
let flowPaths = []
let flowDots = []
let flowNodes = []
let agentChips = []

// ───────────────────────── 时钟引擎 ─────────────────────────
const fmtTime = (s) => {
  const m = Math.floor(s / 60)
  const ss = Math.floor(s % 60)
  return `${String(m).padStart(2, '0')}:${String(ss).padStart(2, '0')}`
}
const idxAt = (time) => {
  for (let i = scenes.length - 1; i >= 0; i--) if (time >= STARTS[i]) return i
  return 0
}
const easeOut = (x) => 1 - Math.pow(1 - Math.min(Math.max(x, 0), 1), 3)
const easeIO = (x) => {
  x = Math.min(Math.max(x, 0), 1)
  return x < 0.5 ? 2 * x * x : 1 - Math.pow(-2 * x + 2, 2) / 2
}

// 进入场景：切换 .is-on（强制回流保证 CSS 动画从头重放，seek 也走这里）
// Enter scene: toggle .is-on (forced reflow restarts CSS animations, also on seek)
const activate = (idx, force = false) => {
  if (idx === activeIdx && !force) return
  if (activeIdx >= 0 && sceneEls[activeIdx]) sceneEls[activeIdx].classList.remove('is-on')
  activeIdx = idx
  const el = sceneEls[idx]
  if (!el) return
  // 工作流场景先复位描边，再点亮，避免进场瞬间闪现完整连线
  // Reset the flow edges before revealing the scene to avoid a full-graph flash
  if (idx === SCENE.flow) resetFlow()
  void el.offsetWidth
  el.classList.add('is-on')
  if (sceneNameRef.value) sceneNameRef.value.textContent = scenes[idx].name
}

const tick = (ts) => {
  if (destroyed) return
  raf = requestAnimationFrame(tick)
  if (!playing.value) { lastTs = ts; return }
  let dt = (ts - lastTs) / 1000
  lastTs = ts
  if (dt > 0.06) dt = 0.06 // 后台节流回来时不跳帧
  t += dt
  if (t >= TOTAL) { t = TOTAL - 0.001; finish() }
  render()
}

const render = () => {
  const idx = idxAt(t)
  activate(idx)
  const s = scenes[idx]
  const local = t - STARTS[idx]
  // 场景内驱动（打字机/计数器/网络流/脉冲…）
  runSceneFx(idx, local)
  drawCanvas(idx, local)
  updAudio() // 音频事件表触发 + 和弦/低音/琶音推进
  // 控制条进度
  for (let i = 0; i < segFills.length; i++) {
    const p = i < idx ? 100 : i > idx ? 0 : (local / s.dur) * 100
    segFills[i].style.width = p + '%'
  }
  if (timeCurRef.value) {
    const txt = fmtTime(t)
    if (timeCurRef.value.textContent !== txt) timeCurRef.value.textContent = txt
  }
}

const finish = () => {
  playing.value = false
  ended.value = true
  ctrlHide.value = false
  audioDuck(0.4) // 片尾音量随结束缓缓收掉
}

const play = () => {
  if (ended.value) { replay(); return }
  started.value = true
  playing.value = true
  audioUnduck()
  ctrlLease()
}
const pause = () => {
  playing.value = false
  ctrlHide.value = false
  audioDuck(0.08)
}
const togglePlay = () => (playing.value ? pause() : play())
const replay = () => {
  ended.value = false
  t = 0
  activate(0, true)
  started.value = true
  playing.value = true
  audioUnduck()
  syncAudio()
  ctrlLease()
}
const seek = (time) => {
  t = Math.min(Math.max(time, 0), TOTAL - 0.001)
  ended.value = false
  activate(idxAt(t), true)
  audioSyncing = true // seek 越过的事件不触发, 只对齐状态
  render()
  audioSyncing = false
  syncAudio()
  if (playing.value) audioUnduck() // 分段跳转等手势交互: 顺带激活音频
}
const seekToScene = (i) => seek(STARTS[i] + 0.01)

const onKey = (e) => {
  if (e.key === ' ' || e.key === 'k') { e.preventDefault(); togglePlay() }
  else if (e.key === 'ArrowRight') seek(t + 5)
  else if (e.key === 'ArrowLeft') seek(t - 5)
  else if (e.key === 'f') toggleFs()
  else if (e.key === 'Home') seek(0)
}

const toggleFs = () => {
  // requestFullscreen 返回 Promise, 被浏览器拒绝(如无用户手势)时兜住, 避免 unhandled rejection
  // requestFullscreen returns a Promise — swallow rejection (e.g. no user gesture)
  if (document.fullscreenElement) document.exitFullscreen().catch(() => {})
  else rootRef.value?.requestFullscreen?.()?.catch(() => {})
}
const onFsChange = () => { isFs.value = document.fullscreenElement === rootRef.value }

// 控制条自动隐藏
const ctrlWake = () => {
  ctrlHide.value = false
  clearTimeout(ctrlTimer)
  if (playing.value) ctrlTimer = setTimeout(() => { ctrlHide.value = true }, 2400)
}
const ctrlLease = () => ctrlWake()

// ───────────────────────── 音频引擎（Web Audio 全合成, 无素材文件）─────────────────────────
// Audio engine: fully synthesized LIGHT soundtrack (music-box / kalimba feel):
// C-major I-V-vi-IV pad + plucked arpeggio + light bass pulse + scene SFX
// routed through: [padVoices / bassOsc / sfxPluck arp] → musBus ┐
//                 [one-shot SFX]                → sfxBus ├→ compressor → master → analyser → out
const audioOn = ref(true)  // 音效开关（用户意图）
const audioLive = ref(false) // 音频是否真正出声（自动播放无手势时 AudioContext 挂起, 首次交互才激活）
let AC = null
let masterBus, musBus, sfxBus, compNode, analyser, lvlBuf
let padFilter, padGain, padVoices = [], bassOsc, bassGain, noiseBuf
let audioEvents = [], evIdx = 0, firedCount = 0
let chordIdx = 0, bassNext = 0, arpNext = 0, arpStep = 0, endFaded = false, audioSyncing = false
let audioLevel = 0 // 输出电平（RMS）, 供开发验证

// MIDI 音号 → 频率
// MIDI note number → frequency
const mtof = (m) => 440 * Math.pow(2, (m - 69) / 12)

// 每幕和声: C 大调 I-V-vi-IV 轻快进行（C→G→Am→F→C→G→F→C 收束回主）+ 滤波亮度 / pad 响度
// Per-scene chord: bright I-V-vi-IV in C major, resolved back to the tonic at the end
const CHORDS = [
  { n: [48, 55, 64, 72], cut: 2200, g: 0.12 }, // S1 开场     C  (C3 G3 E4 C5)
  { n: [43, 59, 62, 67], cut: 2000, g: 0.11 }, // S2 理念     G  (G2 B3 D4 G4)
  { n: [45, 60, 64, 69], cut: 2400, g: 0.12 }, // S3 渠道网关 Am (A2 C4 E4 A4)
  { n: [41, 57, 60, 65], cut: 1700, g: 0.10 }, // S4 流式对话 F  (F2 A3 C4 F4, 避让打字声)
  { n: [48, 55, 64, 72], cut: 2200, g: 0.12 }, // S5 工作流   C
  { n: [43, 59, 67, 74], cut: 2800, g: 0.14 }, // S6 Agent    G  (最亮, D5 顶音)
  { n: [41, 57, 60, 65], cut: 2000, g: 0.12 }, // S7 架构     F
  { n: [48, 55, 64, 72], cut: 2400, g: 0.16 }  // S8 尾板     C  (回归主和弦收束)
]
// 低音: 短促轻点(快速衰减, 不再是沉重心跳), 网关段最带驱动感
// Bass: short light pluck-style pulse instead of a heavy heartbeat
const BASS_INT = [0.85, 0.85, 0.55, 1.1, 1.1, 0.7, 0.85, 1.4]
const BASS_G   = [0.08, 0.08, 0.11, 0.05, 0.07, 0.10, 0.08, 0.09]
// 卡林巴拨弦琶音层(轻快感的来源): 每幕的步进间隔 / 响度
// Kalimba-style plucked arpeggio: the lightness carrier. Step interval & level per scene
const ARP_INT = [0.25, 0.25, 0.25, 0.375, 0.25, 0.25, 0.25, 0.375]
const ARP_G   = [0.07, 0.065, 0.075, 0.045, 0.07, 0.085, 0.065, 0.055]

// 一次性音效事件表（按影片时间排序, 播放中逐个触发; seek 越过的不触发）
// One-shot SFX event list on the film timeline
const buildEvents = () => {
  const ev = []
  const push = (tt, k, f = 0, g = 0, p = 0) => ev.push({ t: tt, k, f, g, p })
  // 幕转场轻 whoosh（每幕起点前 0.18s, 左右交替, 气流感不压画面）
  for (let i = 1; i < STARTS.length; i++) push(STARTS[i] - 0.18, 'whoosh', 0, 0.11, i % 2 ? 0.25 : -0.25)
  // S1 开场: 标题落定(上行琶音) + 高频 sparkle
  push(0.55, 'boom')
  push(1.3, 'blip', mtof(84), 0.04, -0.4)
  push(2.5, 'blip', mtof(88), 0.035, 0.4)
  push(3.9, 'blip', mtof(93), 0.03, -0.3)
  // S2 理念: 计数 tick 流 + 定格提示音
  for (let tt = 0.5; tt < 2.7; tt += 0.11) push(STARTS[1] + tt, 'tick', 0, 0.032)
  push(STARTS[1] + 3.1, 'blip', mtof(84), 0.07)
  // S4 对话: 打字 tick / 消息送达 / 代码行 blip（8 行渐升）/ 停止生成
  const S4 = STARTS[3]
  for (let tt = 0.5; tt < 1.8; tt += 0.085) push(S4 + tt, 'tick', 0, 0.04)
  push(S4 + 1.95, 'blip', mtof(84), 0.08)
  for (let tt = 2.55; tt < 4.1; tt += 0.075) push(S4 + tt, 'tick', 0, 0.03)
  for (let i = 0; i < 8; i++) push(S4 + 4.3 + i * 0.26, 'blip', 520 + i * 40, 0.06)
  push(S4 + 6.45, 'blip', mtof(76), 0.07)
  // S5 工作流: 节点 C 大调五声音阶 blip + 分支轻 whoosh + 完成和弦
  const S5 = STARTS[4]
  const nodeT = [1.0, 1.55, 2.55, 2.6, 3.6, 4.4]
  const nodeF = [72, 74, 76, 79, 81, 84].map(mtof) // C5 D5 E5 G5 A5 C6
  nodeT.forEach((nt, i) => push(S5 + nt, 'blip', nodeF[i], 0.09))
  ;[1.25, 2.25, 3.25].forEach((et, i) => push(S5 + et, 'whoosh', 0, 0.045, i % 2 ? 0.2 : -0.2))
  push(S5 + 4.75, 'chime', 0, 0.1)
  // S6 Agent: 轨道 ping 左右交替（六颗工具芯片, C 大调五声）
  const S6 = STARTS[5]
  ;[84, 86, 88, 91, 93, 96].forEach((m, i) => push(S6 + 0.6 + i, 'blip', mtof(m), 0.04, i % 2 ? 0.45 : -0.45))
  // S7 架构: 分层上行动机(C 大调五声) + 技术栈 shimmer
  const S7 = STARTS[6]
  ;[60, 64, 67, 72, 76].forEach((m, i) => push(S7 + 0.5 + i * 0.45, 'blip', mtof(m), 0.06))
  push(S7 + 3.2, 'blip', mtof(88), 0.03, 0.3)
  // S8 尾板: 上行琶音 swell + 高音 sparkle（淡黑由 updAudio 里随画面收声）
  push(STARTS[7] + 0.35, 'boom')
  push(STARTS[7] + 1.2, 'blip', mtof(93), 0.035, 0.35)
  push(STARTS[7] + 2.4, 'blip', mtof(96), 0.03, -0.35)
  ev.sort((a, b) => a.t - b.t)
  return ev
}

// ── SFX 合成子函数（when = AudioContext 时间）──
// SFX synth units
const sfxWhoosh = (when, g, pan) => {
  const src = AC.createBufferSource(); src.buffer = noiseBuf
  const bp = AC.createBiquadFilter(); bp.type = 'bandpass'; bp.Q.value = 1.4
  bp.frequency.setValueAtTime(500, when)
  bp.frequency.exponentialRampToValueAtTime(1800, when + 0.34)
  const gn = AC.createGain()
  gn.gain.setValueAtTime(0.0001, when)
  gn.gain.exponentialRampToValueAtTime(g, when + 0.1)
  gn.gain.exponentialRampToValueAtTime(0.0001, when + 0.42)
  const pn = AC.createStereoPanner(); pn.pan.value = pan
  src.connect(bp); bp.connect(gn); gn.connect(pn); pn.connect(sfxBus)
  src.start(when); src.stop(when + 0.5)
}
const sfxTick = (when, g) => {
  const src = AC.createBufferSource(); src.buffer = noiseBuf; src.playbackRate.value = 1.6
  const hp = AC.createBiquadFilter(); hp.type = 'highpass'; hp.frequency.value = 3200
  const gn = AC.createGain()
  gn.gain.setValueAtTime(g, when)
  gn.gain.exponentialRampToValueAtTime(0.0001, when + 0.03)
  src.connect(hp); hp.connect(gn); gn.connect(sfxBus)
  src.start(when, Math.random() * 0.4, 0.05); src.stop(when + 0.06)
}
const sfxBlip = (when, f, g, pan = 0, dur = 0.16) => {
  const o = AC.createOscillator(); o.type = 'sine'; o.frequency.value = f
  const o2 = AC.createOscillator(); o2.type = 'triangle'; o2.frequency.value = f * 2
  const g2 = AC.createGain(); g2.gain.value = 0.25
  const gn = AC.createGain()
  gn.gain.setValueAtTime(0.0001, when)
  gn.gain.exponentialRampToValueAtTime(g, when + 0.012)
  gn.gain.exponentialRampToValueAtTime(0.0001, when + dur)
  const pn = AC.createStereoPanner(); pn.pan.value = pan
  o.connect(gn); o2.connect(g2); g2.connect(gn); gn.connect(pn); pn.connect(sfxBus)
  o.start(when); o2.start(when); o.stop(when + dur + 0.02); o2.stop(when + dur + 0.02)
}
// 卡林巴式拨弦: 三角波 + 高八度正弦泛音, 快起急收（琶音层与标题落定共用）
// Kalimba-style pluck: triangle + octave sine partial, fast attack & quick decay
const sfxPluck = (when, f, g) => {
  const o = AC.createOscillator(); o.type = 'triangle'; o.frequency.value = f
  const o2 = AC.createOscillator(); o2.type = 'sine'; o2.frequency.value = f * 2
  const g2 = AC.createGain(); g2.gain.value = 0.3
  const gn = AC.createGain()
  gn.gain.setValueAtTime(0.0001, when)
  gn.gain.exponentialRampToValueAtTime(g, when + 0.008)
  gn.gain.exponentialRampToValueAtTime(0.0001, when + 0.38)
  o.connect(gn); o2.connect(g2); g2.connect(gn); gn.connect(musBus)
  o.start(when); o2.start(when); o.stop(when + 0.42); o2.stop(when + 0.42)
}
// 标题落定/尾板: 轻快上行琶音 + 柔和低音触底（替代沉重 boom）
// Title card / finale: light ascending arpeggio with a soft bass touch
const sfxBoom = (when) => {
  ;[67, 72, 76].forEach((m, i) => sfxPluck(when + i * 0.06, mtof(m), 0.16))
  const o = AC.createOscillator(); o.type = 'sine'; o.frequency.value = mtof(48) // C3 触底
  const gn = AC.createGain()
  gn.gain.setValueAtTime(0.0001, when)
  gn.gain.exponentialRampToValueAtTime(0.18, when + 0.02)
  gn.gain.exponentialRampToValueAtTime(0.0001, when + 0.35)
  o.connect(gn); gn.connect(sfxBus)
  o.start(when); o.stop(when + 0.4)
}
const sfxChime = (when, g) => {
  ;[76, 79, 84, 88].forEach((m, i) => sfxBlip(when + i * 0.07, mtof(m), g, i % 2 ? 0.3 : -0.3, 0.5))
}
const fireEv = (e) => {
  const when = AC.currentTime + 0.02
  if (e.k === 'whoosh') sfxWhoosh(when, e.g, e.p)
  else if (e.k === 'tick') sfxTick(when, e.g)
  else if (e.k === 'blip') sfxBlip(when, e.f, e.g, e.p)
  else if (e.k === 'boom') sfxBoom(when)
  else if (e.k === 'chime') sfxChime(when, e.g)
  firedCount++
}

// 首次播放时构建音频图（在用户手势调用链里, 满足自动播放策略）
// Build the audio graph on first play (inside a user gesture chain)
const ensureAudio = () => {
  if (!AC) {
    const Ctx = window.AudioContext || window.webkitAudioContext
    if (!Ctx) return
    AC = new Ctx()
    compNode = AC.createDynamicsCompressor()
    compNode.threshold.value = -18; compNode.knee.value = 24; compNode.ratio.value = 4
    masterBus = AC.createGain(); masterBus.gain.value = 0
    analyser = AC.createAnalyser(); analyser.fftSize = 256
    lvlBuf = new Uint8Array(analyser.fftSize)
    musBus = AC.createGain(); sfxBus = AC.createGain()
    musBus.connect(compNode); sfxBus.connect(compNode)
    compNode.connect(masterBus); masterBus.connect(analyser); analyser.connect(AC.destination)
    // 白噪声 buffer（whoosh/tick 共用）
    noiseBuf = AC.createBuffer(1, AC.sampleRate, AC.sampleRate)
    const nd = noiseBuf.getChannelData(0)
    for (let i = 0; i < nd.length; i++) nd[i] = Math.random() * 2 - 1
    // pad: 4 声部 ×（三角波 + 高八度正弦泛音）→ 低通 → 轻 LFO 呼吸（柔软无失谐毛刺）
    padFilter = AC.createBiquadFilter(); padFilter.type = 'lowpass'
    padFilter.frequency.value = CHORDS[0].cut; padFilter.Q.value = 0.6
    padGain = AC.createGain(); padGain.gain.value = CHORDS[0].g
    padFilter.connect(padGain); padGain.connect(musBus)
    const lfo = AC.createOscillator(); lfo.frequency.value = 0.3
    const lfoG = AC.createGain(); lfoG.gain.value = 180
    lfo.connect(lfoG); lfoG.connect(padFilter.frequency); lfo.start()
    CHORDS[0].n.forEach((m) => {
      const f = mtof(m)
      const o1 = AC.createOscillator(); o1.type = 'triangle'; o1.frequency.value = f
      const o2 = AC.createOscillator(); o2.type = 'sine'; o2.frequency.value = f * 2
      const vg = AC.createGain(); vg.gain.value = 0.22
      o1.connect(vg); o2.connect(vg); vg.connect(padFilter)
      o1.start(); o2.start()
      padVoices.push({ o1, o2 })
    })
    // 轻点低音: 正弦, 音高随幕根音滑移
    bassOsc = AC.createOscillator(); bassOsc.type = 'sine'; bassOsc.frequency.value = mtof(CHORDS[0].n[0])
    bassGain = AC.createGain(); bassGain.gain.value = 0
    bassOsc.connect(bassGain); bassGain.connect(musBus); bassOsc.start()
    audioEvents = buildEvents()
  }
  if (AC.state === 'suspended') AC.resume().then(() => { audioLive.value = true }).catch(() => {})
}

// 换幕: 和声滑音 + 滤波亮度 + pad 响度
// Scene change: glide chord voices, filter brightness and pad level
const setChord = (idx, instant = false) => {
  if (!AC || chordIdx === idx) return
  chordIdx = idx
  const c = CHORDS[idx], now = AC.currentTime, tc = instant ? 0.02 : 0.5
  c.n.forEach((m, i) => {
    const f = mtof(m)
    padVoices[i]?.o1.frequency.setTargetAtTime(f, now, tc)
    padVoices[i]?.o2.frequency.setTargetAtTime(f, now, tc)
  })
  padFilter.frequency.setTargetAtTime(c.cut, now, tc)
  padGain.gain.setTargetAtTime(c.g, now, instant ? 0.05 : 0.8)
  bassOsc.frequency.setTargetAtTime(mtof(c.n[0]), now, tc)
}

// seek/replay 后对齐音频状态（跳过已越过的事件）
// Align audio state after seek/replay (events crossed by the jump are skipped)
const syncAudio = () => {
  if (!AC) return
  evIdx = 0
  while (evIdx < audioEvents.length && audioEvents[evIdx].t <= t + 0.02) evIdx++
  chordIdx = -1
  setChord(idxAt(t), true)
  endFaded = idxAt(t) === scenes.length - 1 && t - STARTS[scenes.length - 1] > 5.0
  bassNext = t + 0.2
  arpNext = t + 0.2; arpStep = 0
}

// 音量总控: 暂停/结束收声, 播放恢复
// Master duck/unduck for pause, finish and resume
const audioDuck = (tc) => { if (AC) masterBus.gain.setTargetAtTime(0.0001, AC.currentTime, tc) }
const audioUnduck = () => {
  ensureAudio()
  if (AC && audioOn.value) masterBus.gain.setTargetAtTime(0.9, AC.currentTime, 0.25)
}
const toggleAudio = () => {
  // 未出声(自动播放挂起/已静音)时点击 = 开启; 已出声时点击 = 静音
  // Click means "enable" while silent (autoplay-suspended or muted), "mute" while live
  if (!audioOn.value || !audioLive.value) {
    audioOn.value = true
    ensureAudio()
    if (AC && AC.state === 'running') {
      audioLive.value = true
      if (playing.value) masterBus.gain.setTargetAtTime(0.9, AC.currentTime, 0.1)
    }
  } else {
    audioOn.value = false
    audioDuck(0.1)
  }
}

// 每帧驱动: 事件触发 / 和弦推进 / 轻点低音与琶音调度 / 尾板收声 / 电平表
// Per-frame audio driver
const updAudio = () => {
  if (!AC) return
  const now = AC.currentTime
  if (playing.value && !audioSyncing && AC.state === 'running') {
    if (!audioLive.value) audioLive.value = true // 兜底标记（resume.then 之外的路径）
    while (evIdx < audioEvents.length && audioEvents[evIdx].t <= t) fireEv(audioEvents[evIdx++])
    setChord(idxAt(t))
    const bi = idxAt(t)
    if (t >= bassNext) { // 轻点低音以影片时间为准, 暂停自然冻结
      bassNext = Math.max(bassNext + BASS_INT[bi], t + 0.05)
      bassGain.gain.cancelScheduledValues(now)
      bassGain.gain.setValueAtTime(0.0001, now)
      bassGain.gain.exponentialRampToValueAtTime(BASS_G[bi], now + 0.015)
      bassGain.gain.exponentialRampToValueAtTime(0.0001, now + Math.min(BASS_INT[bi] * 0.55, 0.3))
    }
    if (t >= arpNext && !endFaded) { // 琶音步进: 根音高八度 → 三音 → 顶音 → 三音 循环
      arpNext = Math.max(arpNext + ARP_INT[bi], t + 0.05)
      const c = CHORDS[bi]
      sfxPluck(now + 0.01, mtof([c.n[0] + 12, c.n[2], c.n[3], c.n[2]][arpStep++ % 4]), ARP_G[bi])
    }
    if (bi === scenes.length - 1 && t - STARTS[bi] > 5.0 && !endFaded) { // 尾板淡黑同步收声
      endFaded = true
      padGain.gain.setTargetAtTime(0.0001, now, 1.0)
      bassGain.gain.setTargetAtTime(0.0001, now, 0.8)
    }
  }
  if (analyser) { // 输出电平（RMS）, 供开发验证
    analyser.getByteTimeDomainData(lvlBuf)
    let sum = 0
    for (let i = 0; i < lvlBuf.length; i += 4) { const v = (lvlBuf[i] - 128) / 128; sum += v * v }
    audioLevel = Math.sqrt(sum / (lvlBuf.length / 4))
  }
}
// 开发验证辅助: 音频运行态（状态/已触发事件数/输出电平/主增益）
// Dev helper for E2E: audio runtime state
const audioDebug = () => ({ ready: !!AC, state: AC?.state ?? 'none', fired: firedCount, level: +audioLevel.toFixed(4), duck: masterBus ? +masterBus.gain.value.toFixed(3) : null, on: audioOn.value })

// ───────────────────────── 场景内驱动 ─────────────────────────
const clamp01 = (x) => Math.min(Math.max(x, 0), 1)

const runSceneFx = (idx, l) => {
  if (idx === SCENE.vision) updVision(l)
  else if (idx === SCENE.gate) updGate(l)
  else if (idx === SCENE.chat) updChat(l)
  else if (idx === SCENE.flow) updFlow(l)
}

// S2 数字滚动
let visLast = ''
const updVision = (l) => {
  if (!sceneEls[SCENE.vision]) return
  const els = sceneEls[SCENE.vision].querySelectorAll('.sv-num[data-n]')
  const k1 = Math.round(24000 * easeOut(l / 2.2))
  const k2 = Math.round(37000 * easeOut((l - 0.15) / 2.2))
  const sig = k1 + '|' + k2
  if (sig === visLast) return
  visLast = sig
  if (els[0]) els[0].textContent = k1.toLocaleString('en-US') + '+'
  if (els[1]) els[1].textContent = k2.toLocaleString('en-US') + '+'
}

// S3 渠道选路：DOM 标签点亮同步（脉冲本身在 canvas 里画）
let gateLastK = -1
const updGate = (l) => {
  const k = Math.floor(l / 2) % providers.length
  if (k === gateLastK) return
  gateLastK = k
  provEls.forEach((el, i) => el && el.classList.toggle('lit', i === k))
}

// S4 流式对话打字机
let chatSig = ''
const updChat = (l) => {
  const uN = Math.floor(clamp01((l - 0.45) / 1.35) * CHAT_USER.length)
  const aN = Math.floor(clamp01((l - 2.5) / 1.6) * CHAT_AI.length)
  // full = 已完成的行数（0..CHAT_CODE.length），不能用 clamp01（会把行数压成 0/1）
  // full = completed line count (0..CHAT_CODE.length); clamp01 would squash it to 0/1
  const full = Math.min(Math.max(Math.floor((l - 4.3) / 0.26), 0), CHAT_CODE.length)
  const part = clamp01(((l - 4.3) - full * 0.26) / 0.26)
  const sig = uN + '|' + aN + '|' + full + '|' + Math.floor(part * 8)
  if (sig === chatSig) return
  chatSig = sig
  if (scUserRef.value) scUserRef.value.textContent = CHAT_USER.slice(0, uN)
  if (scAiRef.value) scAiRef.value.textContent = CHAT_AI.slice(0, aN)
  if (scCodeRef.value) {
    let html = ''
    for (let i = 0; i < Math.min(full, CHAT_CODE.length); i++) html += codeLine(CHAT_CODE[i], 1)
    if (full < CHAT_CODE.length && full >= 0) html += codeLine(CHAT_CODE[full] || '', part)
    if (full >= CHAT_CODE.length) html = CHAT_CODE.map((c) => codeLine(c, 1)).join('')
    if (scCodeRef.value.__sig !== sig) {
      scCodeRef.value.innerHTML = html
      scCodeRef.value.__sig = sig
    }
  }
  if (scStopRef.value) scStopRef.value.classList.toggle('on', l > 6.4)
}
// 极简语法着色：关键字/布尔着紫，其余默认
const codeLine = (line, ratio) => {
  if (ratio <= 0) return ''
  const cut = line.slice(0, Math.ceil(line.length * Math.min(ratio, 1)))
  const colored = cut
    .replace(/(func|if|return|time)/g, '<em>$1</em>')
    .replace(/(true|false)/g, '<b>$1</b>')
  return `<span class="cl">${colored}</span>`
}

// S5 工作流：进场时复位描边动画（CSS transition 驱动画线，JS 驱动脉冲点）
const resetFlow = () => {
  flowPaths.forEach((p) => {
    if (!p) return
    const len = p.__len || (p.__len = p.getTotalLength())
    p.style.strokeDasharray = len
    p.classList.add('no-trans')
    p.style.strokeDashoffset = len
    p.classList.remove('drawn')
  })
  void (flowPaths[0] && flowPaths[0].getBoundingClientRect())
  flowPaths.forEach((p, i) => {
    if (!p) return
    p.classList.remove('no-trans')
    void p.getBoundingClientRect()
    p.style.transitionDelay = EDGE_T[i][0] + 's'
    requestAnimationFrame(() => { p.style.strokeDashoffset = 0 })
  })
}

const updFlow = (l) => {
  // 脉冲点沿边运动 + 边画完后点亮
  flowPaths.forEach((p, i) => {
    const dot = flowDots[i]
    if (!p || !dot) return
    const [t0, t1] = EDGE_T[i]
    const u = (l - t0) / (t1 - t0)
    if (u < 0 || u > 1.15) { dot.style.opacity = 0; return }
    if (p.__len === undefined) p.__len = p.getTotalLength()
    const pt = p.getPointAtLength(p.__len * Math.min(u, 1))
    dot.style.opacity = u > 1 ? 0 : 1
    dot.setAttribute('cx', pt.x)
    dot.setAttribute('cy', pt.y)
    // 脉冲走完 → 该边切换为渐变高亮
    if (u >= 1 && !p.classList.contains('drawn')) p.classList.add('drawn')
  })
  flowNodes.forEach((n, i) => {
    if (!n) return
    const on = [1.0, 1.55, 2.55, 2.6, 3.6, 4.4][i] <= l
    n.classList.toggle('lit', on)
  })
  if (sfElapsedRef.value) {
    const e = clamp01((l - 1.0) / 3.4) * 1.3
    const txt = e.toFixed(2)
    if (sfElapsedRef.value.textContent !== txt) sfElapsedRef.value.textContent = txt
  }
}

// ───────────────────────── 画布特效 ─────────────────────────
let ctx = null
let stars = []
// 渠道网关几何（舞台坐标 1280×720）
const G_SRC = { x: 190, y: 400 }
const G_CORE = { x: 640, y: 400 }
const G_PROV_X = 1075
const quadPt = (a, c, b, u) => {
  const x = (1 - u) * (1 - u) * a.x + 2 * (1 - u) * u * c.x + u * u * b.x
  const y = (1 - u) * (1 - u) * a.y + 2 * (1 - u) * u * c.y + u * u * b.y
  return { x, y }
}
const initStars = () => {
  stars = []
  for (let i = 0; i < 130; i++) {
    stars.push({
      x: Math.random() * STAGE_W,
      y: Math.random() * STAGE_H,
      r: 0.5 + Math.random() * 1.4,
      spd: 0.3 + Math.random() * 0.9,
      ph: Math.random() * Math.PI * 2,
      tw: 0.4 + Math.random() * 1.2
    })
  }
}

const drawCanvas = (idx, l) => {
  if (!ctx) return
  ctx.clearRect(0, 0, STAGE_W, STAGE_H)
  // 星野：缓慢左移 + 闪烁；重内容场景里调暗
  const dim = idx >= SCENE.gate && idx <= SCENE.agent ? 0.55 : 1
  for (const s of stars) {
    s.x -= s.spd * 6 * (1 / 60)
    if (s.x < -4) s.x = STAGE_W + 4
    const a = (0.16 + 0.3 * (0.5 + 0.5 * Math.sin(l * s.tw + s.ph))) * dim
    ctx.beginPath()
    ctx.fillStyle = `rgba(190,214,255,${a.toFixed(3)})`
    ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2)
    ctx.fill()
  }
  if (idx === SCENE.gate) drawGate(l)
  else if (idx === SCENE.agent) drawAgent(l)
}

const drawGate = (l) => {
  // 底层连线
  ctx.lineWidth = 1.2
  ctx.strokeStyle = 'rgba(125,165,255,0.16)'
  ctx.beginPath()
  ctx.moveTo(G_SRC.x, G_SRC.y)
  ctx.lineTo(G_CORE.x, G_CORE.y)
  ctx.stroke()
  for (let i = 0; i < providers.length; i++) {
    const p = providers[i]
    void p
    const py = sgProvY[i]
    ctx.beginPath()
    ctx.moveTo(G_CORE.x, G_CORE.y)
    ctx.bezierCurveTo(G_CORE.x + 150, G_CORE.y, (G_CORE.x + G_PROV_X) / 2, py, G_PROV_X, py)
    ctx.stroke()
  }
  // 当前选路 k，一个请求脉冲沿「请求 → 网关 → 供应商」行进
  const k = Math.floor(l / 2) % providers.length
  const ph = l % 2
  const u = easeIO(ph / 1.55)
  if (ph < 1.55) {
    for (let g = 2; g >= 0; g--) {
      const uu = u - g * 0.055
      if (uu <= 0) continue
      const pt = u <= 0.42
        ? { x: G_SRC.x + (G_CORE.x - G_SRC.x) * uu / 0.42, y: G_SRC.y }
        : bezierGate(sgProvY[k], (uu - 0.42) / 0.58)
      const a = (g === 0 ? 0.95 : 0.38 / g)
      ctx.beginPath()
      ctx.fillStyle = `rgba(124,232,255,${a})`
      ctx.shadowColor = 'rgba(124,232,255,0.9)'
      ctx.shadowBlur = 14
      ctx.arc(pt.x, pt.y, 4.2, 0, Math.PI * 2)
      ctx.fill()
      ctx.shadowBlur = 0
    }
    // 到达闪环
    if (u > 0.92) {
      const ru = (u - 0.92) / 0.08
      ctx.beginPath()
      ctx.strokeStyle = `rgba(124,232,255,${(0.6 * (1 - ru)).toFixed(3)})`
      ctx.lineWidth = 2
      ctx.arc(G_PROV_X, sgProvY[k], 10 + 26 * ru, 0, Math.PI * 2)
      ctx.stroke()
    }
  }
  // 网关核心脉冲环 + 辉光
  const frac = (l * 0.55) % 1
  ctx.beginPath()
  ctx.strokeStyle = `rgba(168,85,247,${(0.4 * (1 - frac)).toFixed(3)})`
  ctx.lineWidth = 1.6
  ctx.arc(G_CORE.x, G_CORE.y, 30 + 46 * frac, 0, Math.PI * 2)
  ctx.stroke()
  const gr = ctx.createRadialGradient(G_CORE.x, G_CORE.y, 4, G_CORE.x, G_CORE.y, 64)
  gr.addColorStop(0, 'rgba(109,164,255,0.34)')
  gr.addColorStop(1, 'rgba(109,164,255,0)')
  ctx.fillStyle = gr
  ctx.beginPath()
  ctx.arc(G_CORE.x, G_CORE.y, 64, 0, Math.PI * 2)
  ctx.fill()
}

// 网关 → 供应商 k 的三次贝塞尔取点（与底层连线同一形状）
const bezierGate = (py, u) => {
  const c1x = G_CORE.x + 150, c2x = (G_CORE.x + G_PROV_X) / 2
  const iu = 1 - u
  return {
    x: iu * iu * iu * G_CORE.x + 3 * iu * iu * u * c1x + 3 * iu * u * u * c2x + u * u * u * G_PROV_X,
    y: iu * iu * iu * G_CORE.y + 3 * iu * iu * u * G_CORE.y + 3 * iu * u * u * py + u * u * u * py
  }
}

// S6 Agent：核心 ↔ 工具 的双向脉冲
const A_C = { x: 640, y: 402 }
const A_R = 222
const agentPos = agentTools.map((t2, i) => {
  void t2
  const ang = (-90 + i * 60) * Math.PI / 180
  return { x: A_C.x + A_R * Math.cos(ang), y: A_C.y + A_R * Math.sin(ang) }
})
const drawAgent = (l) => {
  ctx.beginPath()
  ctx.strokeStyle = 'rgba(125,165,255,0.13)'
  ctx.lineWidth = 1
  ctx.setLineDash([3, 7])
  ctx.arc(A_C.x, A_C.y, A_R, 0, Math.PI * 2)
  ctx.stroke()
  ctx.setLineDash([])
  // 每个工具一个相位错开的双向往返脉冲
  for (let i = 0; i < agentPos.length; i++) {
    const phase = (l * 0.5 + i / agentPos.length) % 1
    const out = phase < 0.5
    const u = easeIO(out ? phase * 2 : 2 - phase * 2)
    const p = agentPos[i]
    const x = A_C.x + (p.x - A_C.x) * u
    const y = A_C.y + (p.y - A_C.y) * u
    for (let g = 2; g >= 0; g--) {
      const uu = Math.max(u - g * 0.05, 0)
      const gx = A_C.x + (p.x - A_C.x) * uu
      const gy = A_C.y + (p.y - A_C.y) * uu
      const a = g === 0 ? 0.85 : 0.3 / g
      ctx.beginPath()
      ctx.fillStyle = out ? `rgba(124,232,255,${a})` : `rgba(196,140,255,${a})`
      ctx.arc(gx, gy, g === 0 ? 3.8 : 2.6, 0, Math.PI * 2)
      ctx.fill()
    }
  }
  // 核心辉光
  const gr = ctx.createRadialGradient(A_C.x, A_C.y, 4, A_C.x, A_C.y, 90)
  gr.addColorStop(0, 'rgba(109,164,255,0.3)')
  gr.addColorStop(1, 'rgba(109,164,255,0)')
  ctx.fillStyle = gr
  ctx.beginPath()
  ctx.arc(A_C.x, A_C.y, 90, 0, Math.PI * 2)
  ctx.fill()
}

// ───────────────────────── 生命周期 ─────────────────────────
let ro = null
let io = null

const applyScale = () => {
  const root = rootRef.value
  const stage = stageRef.value
  if (!root || !stage) return
  const w = root.clientWidth || STAGE_W
  const s = w / STAGE_W
  stage.style.transform = `scale(${s})`
  const canvas = canvasRef.value
  if (canvas && ctx) {
    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    canvas.width = Math.round(STAGE_W * s * dpr)
    canvas.height = Math.round(STAGE_H * s * dpr)
    ctx.setTransform(s * dpr, 0, 0, s * dpr, 0, 0)
  }
}

const autoPause = () => {
  if (!playing.value) return
  autoPaused = true
  pause()
}
const autoResume = () => {
  if (!autoPaused) return
  autoPaused = false
  play()
}
const onVis = () => (document.hidden ? autoPause() : autoResume())

onMounted(() => {
  // 收集静态 DOM（模板一次性渲染，之后全部手工驱动）
  sceneEls = Array.from(rootRef.value.querySelectorAll('.tpv-scene'))
  segFills = Array.from(segsRef.value.querySelectorAll('.tpv-seg-fill'))
  provEls = Array.from(rootRef.value.querySelectorAll('.sg-prov'))
  flowPaths = Array.from(rootRef.value.querySelectorAll('.sf-edge'))
  flowDots = Array.from(rootRef.value.querySelectorAll('.sf-dot'))
  flowNodes = Array.from(rootRef.value.querySelectorAll('.sf-node'))
  agentChips = Array.from(rootRef.value.querySelectorAll('.sa-chip'))

  // Agent 工具芯片按轨道坐标布点
  agentChips.forEach((el, i) => {
    const p = agentPos[i]
    el.style.left = p.x - 66 + 'px'
    el.style.top = p.y - 24 + 'px'
  })

  // 画布
  ctx = canvasRef.value.getContext('2d')
  initStars()
  applyScale()

  ro = new ResizeObserver(applyScale)
  ro.observe(rootRef.value)
  // 滚出视口自动暂停，回来续播
  io = new IntersectionObserver((es) => es.forEach((e) => (e.isIntersecting ? autoResume() : autoPause())), { threshold: 0.2 })
  io.observe(rootRef.value)
  document.addEventListener('visibilitychange', onVis)
  document.addEventListener('fullscreenchange', onFsChange)

  // 影片自动开播（画面自动放映; 音效受浏览器自动播放策略限制, 首次交互后自动加入）
  // Autoplay on mount; sound joins after the first user gesture (autoplay policy)
  started.value = true
  playing.value = true
  activate(0, true)
  ctrlLease()
  lastTs = performance.now()
  raf = requestAnimationFrame(tick)
})

onBeforeUnmount(() => {
  destroyed = true
  cancelAnimationFrame(raf)
  clearTimeout(ctrlTimer)
  if (AC) { AC.close().catch(() => {}); AC = null } // 释放音频图
  ro?.disconnect()
  io?.disconnect()
  document.removeEventListener('visibilitychange', onVis)
  document.removeEventListener('fullscreenchange', onFsChange)
  if (document.fullscreenElement === rootRef.value) document.exitFullscreen().catch(() => {})
})
</script>

<style scoped lang="scss">
// ───────────────────────── 基础与舞台 ─────────────────────────
.tpv {
  --tb: #04060d;
  --blue: #4da3ff;
  --violet: #a855f7;
  --cyan: #7be8ff;
  --txt: #f2f6ff;
  --dim: rgba(206, 224, 255, 0.62);
  --ease: cubic-bezier(0.22, 1, 0.36, 1);

  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: var(--tb);
  border-radius: 18px;
  overflow: hidden;
  outline: none;
  user-select: none;
  cursor: default;
  box-shadow: 0 24px 70px -22px rgba(8, 18, 48, 0.55), 0 0 0 1px rgba(120, 150, 220, 0.14);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  color: var(--txt);

  &:focus-visible { box-shadow: 0 0 0 2px var(--blue), 0 24px 70px -22px rgba(8, 18, 48, 0.55); }
  &:fullscreen { border-radius: 0; box-shadow: none; }

  // 暂停时冻结场内所有 CSS 动画（时钟同样停在原地，两边一致）
  // Freeze every CSS animation while paused, in step with the clock
  &.is-paused .tpv-stage,
  &.is-paused .tpv-stage * { animation-play-state: paused !important; }
}

.tpv-stage {
  position: absolute;
  top: 0;
  left: 0;
  width: 1280px;
  height: 720px;
  transform-origin: 0 0;
}

.tpv-bg { position: absolute; inset: 0; background: radial-gradient(120% 90% at 50% 0%, #0a1226 0%, #05070f 55%, #04060d 100%); }
.tpv-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  opacity: 0.5;
  animation: tpvDrift 14s ease-in-out infinite alternate;
  &.g1 { width: 560px; height: 560px; left: -140px; top: -180px; background: rgba(45, 96, 200, 0.32); }
  &.g2 { width: 520px; height: 520px; right: -160px; bottom: -200px; background: rgba(126, 58, 220, 0.26); animation-delay: -5s; }
  &.g3 { width: 380px; height: 380px; left: 42%; top: 30%; background: rgba(38, 130, 190, 0.14); animation-delay: -9s; }
}
@keyframes tpvDrift {
  from { transform: translate3d(0, 0, 0) scale(1); }
  to { transform: translate3d(46px, 30px, 0) scale(1.12); }
}

.tpv-canvas { position: absolute; inset: 0; width: 1280px; height: 720px; pointer-events: none; }

// 胶片颗粒 + 暗角
.tpv-grain {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0.05;
  mix-blend-mode: overlay;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='140' height='140'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2'/%3E%3C/filter%3E%3Crect width='140' height='140' filter='url(%23n)' opacity='0.7'/%3E%3C/svg%3E");
}
.tpv-vignette {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(118% 96% at 50% 46%, transparent 58%, rgba(1, 3, 9, 0.5) 100%);
}

// ───────────────────────── 场景切换与通用元素 ─────────────────────────
.tpv-scene {
  position: absolute;
  inset: 0;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.7s ease, visibility 0s linear 0.7s;
  &.is-on {
    opacity: 1;
    visibility: visible;
    transition: opacity 0.7s ease;
  }
}

.tpv-kicker {
  font-size: 15px;
  letter-spacing: 0.42em;
  text-transform: uppercase;
  color: var(--dim);
  font-weight: 500;
}
.tpv-head { position: absolute; left: 80px; top: 62px; z-index: 3;
  .tpv-kicker { margin-bottom: 14px; }
}
.tpv-hl {
  margin: 0;
  font-size: 44px;
  font-weight: 700;
  letter-spacing: 0.02em;
  background: linear-gradient(100deg, #eef4ff 30%, #9dc4ff 75%, #cdb6fd);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  opacity: 0;
  transform: translateY(22px);
}
.is-on .tpv-hl { animation: tpvRise 0.9s 0.12s var(--ease) both; }

.tpv-chips {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  gap: 18px;
  z-index: 3;
}
.tpv-chip {
  padding: 11px 22px;
  font-size: 17px;
  letter-spacing: 0.04em;
  color: var(--txt);
  border: 1px solid rgba(130, 168, 255, 0.26);
  border-radius: 999px;
  background: rgba(22, 36, 72, 0.4);
  backdrop-filter: blur(6px);
  opacity: 0;
  transform: translateY(16px);
}
.is-on .tpv-chip { animation: tpvRise 0.7s var(--ease) both; }

@keyframes tpvRise {
  from { opacity: 0; transform: translateY(22px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes tpvFadeIn { from { opacity: 0; } to { opacity: 1; } }

// ───────────────────────── S1 开场 ─────────────────────────
.s-open { display: flex; align-items: center; justify-content: center; }
.so-zoom {
  text-align: center;
  animation: soZoom 6.8s linear both;
}
@keyframes soZoom { from { transform: scale(1.055); } to { transform: scale(1); } }
.so-k { animation: tpvFadeIn 1.2s 0.2s both; }
.so-mark { margin: 26px 0 0; font-size: 104px; font-weight: 800; letter-spacing: 0.06em; line-height: 1.1; }
.so-mark-txt {
  background: linear-gradient(110deg, #dfe9ff 18%, #9dc4ff 38%, var(--blue) 50%, #c9b2fb 62%, #dfe9ff 84%);
  background-size: 240% 100%;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: soSweep 2.6s 0.4s var(--ease) both, tpvRise 1.1s 0.1s var(--ease) both;
}
@keyframes soSweep { from { background-position: 118% 0; } to { background-position: -32% 0; } }
.so-line {
  width: 0;
  height: 1px;
  margin: 30px auto 26px;
  background: linear-gradient(90deg, transparent, var(--blue) 30%, var(--violet) 70%, transparent);
  animation: soLineGrow 1.4s 1.3s var(--ease) both;
}
@keyframes soLineGrow { to { width: 300px; } }
.so-sub {
  margin: 0;
  font-size: 27px;
  letter-spacing: 0.58em;
  text-indent: 0.58em;
  color: var(--txt);
  animation: tpvFadeIn 1.5s 1.9s both;
}
.so-foot {
  position: absolute;
  left: 56px;
  right: 56px;
  bottom: 44px;
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  letter-spacing: 0.34em;
  color: var(--dim);
  opacity: 0;
}
.is-on .so-foot { animation: tpvFadeIn 1.4s 3.4s both; }

// ───────────────────────── S2 理念与数字 ─────────────────────────
.s-vision { display: flex; align-items: center; justify-content: center; }
.sv-wrap { text-align: center; width: 900px; }
.sv-k { animation: tpvFadeIn 0.8s 0.15s both; }
.sv-h { margin: 30px 0 0; display: flex; flex-direction: column; gap: 6px; }
.sv-mask { display: block; overflow: hidden; line-height: 1.28; }
.sv-l1, .sv-l2 {
  display: inline-block;
  font-size: 62px;
  font-weight: 750;
  letter-spacing: 0.015em;
  transform: translateY(110%);
  opacity: 0;
}
.is-on .sv-l1 { animation: svUp 0.9s 0.3s var(--ease) both; }
.is-on .sv-l2 {
  animation: svUp 0.9s 0.48s var(--ease) both;
  background: linear-gradient(100deg, #b7d4ff 10%, var(--blue) 45%, var(--violet) 90%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
@keyframes svUp { to { transform: translateY(0); opacity: 1; } }
.sv-sub { margin: 30px 0 0; font-size: 20px; letter-spacing: 0.12em; color: var(--dim); animation: tpvFadeIn 1s 1.1s both; }
.sv-stats {
  margin-top: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 44px;
  opacity: 0;
}
.is-on .sv-stats { animation: tpvRise 0.9s 1.5s var(--ease) both; }
.sv-stat { display: flex; flex-direction: column; gap: 8px; }
.sv-num {
  font-size: 54px;
  font-weight: 750;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.01em;
  background: linear-gradient(120deg, #dfe9ff, var(--blue));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.sv-num-s { font-size: 44px; }
.sv-lab { font-size: 15px; letter-spacing: 0.3em; text-indent: 0.3em; color: var(--dim); }
.sv-div { width: 1px; height: 52px; background: linear-gradient(180deg, transparent, rgba(140, 175, 255, 0.4), transparent); }

// ───────────────────────── S3 渠道网关 ─────────────────────────
.sg-graph { position: absolute; inset: 0; z-index: 2; }
.sg-node {
  position: absolute;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  height: 56px;
  padding: 0 24px;
  border-radius: 14px;
  font-size: 18px;
  letter-spacing: 0.05em;
  color: var(--txt);
  background: rgba(20, 34, 68, 0.55);
  border: 1px solid rgba(130, 168, 255, 0.28);
  backdrop-filter: blur(6px);
  box-sizing: border-box;
  opacity: 0;
  transform: translateY(14px);
  transition: border-color 0.35s ease, box-shadow 0.35s ease, color 0.35s ease;
}
.is-on .sg-node { animation: tpvRise 0.8s var(--ease) both; }
.is-on .sg-src { animation-delay: 0.35s; }
.is-on .sg-core { animation-delay: 0.55s; }
.is-on .sg-prov { animation-delay: 0.75s; }
.is-on .sg-prov:nth-of-type(4) { animation-delay: 0.87s; }
.is-on .sg-prov:nth-of-type(5) { animation-delay: 0.99s; }
.is-on .sg-prov:nth-of-type(6) { animation-delay: 1.11s; }
.sg-src { left: 118px; top: 372px; min-width: 144px; }
.sg-core {
  left: 552px;
  top: 372px;
  min-width: 176px;
  border-color: rgba(168, 120, 255, 0.45);
  box-shadow: 0 0 34px rgba(109, 120, 255, 0.28), inset 0 0 18px rgba(109, 140, 255, 0.12);
  animation: tpvRise 0.8s 0.55s var(--ease) both, sgBreath 3.2s 1.4s ease-in-out infinite alternate;
}
@keyframes sgBreath {
  from { box-shadow: 0 0 26px rgba(109, 120, 255, 0.22), inset 0 0 14px rgba(109, 140, 255, 0.1); }
  to { box-shadow: 0 0 52px rgba(140, 110, 255, 0.4), inset 0 0 22px rgba(130, 150, 255, 0.2); }
}
.sg-core-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--cyan);
  box-shadow: 0 0 12px var(--cyan);
  animation: sgBlink 1.6s ease-in-out infinite;
}
@keyframes sgBlink { 0%, 100% { opacity: 1; } 50% { opacity: 0.35; } }
.sg-prov { min-width: 150px; }
.sg-prov.lit {
  border-color: rgba(124, 232, 255, 0.75);
  color: #eaf8ff;
  box-shadow: 0 0 26px rgba(96, 200, 255, 0.35), inset 0 0 14px rgba(96, 200, 255, 0.14);
}
.sg-chips { bottom: 64px; }
.is-on .sg-chips .tpv-chip { animation-delay: 1.1s; }
.is-on .sg-chips .tpv-chip:nth-child(2) { animation-delay: 1.24s; }
.is-on .sg-chips .tpv-chip:nth-child(3) { animation-delay: 1.38s; }
.is-on .sg-chips .tpv-chip:nth-child(4) { animation-delay: 1.52s; }

// ───────────────────────── S4 流式对话 ─────────────────────────
.sc-win {
  position: absolute;
  left: 200px;
  top: 170px;
  width: 880px;
  height: 428px;
  border-radius: 16px;
  background: rgba(13, 21, 44, 0.72);
  border: 1px solid rgba(130, 168, 255, 0.22);
  backdrop-filter: blur(10px);
  box-shadow: 0 30px 70px -30px rgba(0, 0, 0, 0.6), inset 0 1px 0 rgba(160, 190, 255, 0.12);
  overflow: hidden;
  z-index: 2;
  opacity: 0;
  transform: translateY(26px) scale(0.985);
}
.is-on .sc-win { animation: scWinIn 1s 0.3s var(--ease) both; }
@keyframes scWinIn { to { opacity: 1; transform: translateY(0) scale(1); } }
.sc-bar {
  display: flex;
  align-items: center;
  gap: 7px;
  height: 42px;
  padding: 0 16px;
  border-bottom: 1px solid rgba(130, 168, 255, 0.14);
  i { width: 10px; height: 10px; border-radius: 50%; background: rgba(150, 175, 230, 0.3); }
}
.sc-bar-title { margin-left: 12px; font-size: 14px; letter-spacing: 0.08em; color: var(--dim); }
.sc-msgs { padding: 20px 24px; display: flex; flex-direction: column; gap: 16px; }
.sc-row { display: flex; }
.sc-row-user { justify-content: flex-end; }
.sc-row-ai { justify-content: flex-start; align-items: flex-start; gap: 10px; }
.sc-avatar {
  flex: none;
  width: 30px;
  height: 30px;
  border-radius: 9px;
  display: grid;
  place-items: center;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, var(--blue), var(--violet));
  box-shadow: 0 4px 14px rgba(80, 140, 255, 0.4);
  opacity: 0;
}
.is-on .sc-avatar { animation: tpvFadeIn 0.5s 2.45s both; }
.sc-bubble {
  max-width: 640px;
  padding: 12px 16px;
  border-radius: 13px;
  font-size: 16.5px;
  line-height: 1.55;
  letter-spacing: 0.02em;
}
.sc-user {
  background: linear-gradient(120deg, #2f6fdd, #3f7fe8);
  border-bottom-right-radius: 4px;
  box-shadow: 0 8px 22px -8px rgba(47, 111, 221, 0.55);
  opacity: 0;
}
.is-on .sc-user { animation: tpvRise 0.5s 0.5s var(--ease) both; }
.sc-ai {
  background: rgba(30, 44, 80, 0.65);
  border: 1px solid rgba(130, 168, 255, 0.16);
  border-bottom-left-radius: 4px;
  min-width: 380px;
  min-height: 44px;
  opacity: 0;
}
.is-on .sc-ai { animation: tpvFadeIn 0.4s 2.4s both; }
.sc-caret {
  display: inline-block;
  width: 2px;
  height: 15px;
  margin-left: 2px;
  vertical-align: -2px;
  background: var(--cyan);
  animation: scBlinkCaret 0.85s steps(1) infinite;
}
@keyframes scBlinkCaret { 50% { opacity: 0; } }
.sc-code {
  margin: 10px 0 2px;
  padding: 12px 14px;
  border-radius: 10px;
  background: #0a1020;
  border: 1px solid rgba(120, 150, 230, 0.14);
  font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
  font-size: 14px;
  line-height: 1.62;
  color: #c9d8f5;
  min-height: 30px;
  // 代码行由 innerHTML 动态注入, 无 data-v 属性, 必须用 :deep() 穿透
  // Code lines are injected via innerHTML (no data-v attr) — must pierce with :deep()
  :deep(.cl) { display: block; white-space: pre; }
  :deep(em) { font-style: normal; color: #c792ea; }
  :deep(b) { font-weight: 600; color: #82e8c0; }
}
.sc-input {
  position: absolute;
  left: 20px;
  right: 20px;
  bottom: 18px;
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-radius: 12px;
  background: rgba(22, 34, 66, 0.6);
  border: 1px solid rgba(130, 168, 255, 0.18);
}
.sc-wave {
  display: flex;
  align-items: center;
  gap: 3px;
  height: 22px;
  i {
    width: 3px;
    height: 30%;
    border-radius: 2px;
    background: linear-gradient(180deg, var(--cyan), var(--blue));
    animation: scWave 1.1s ease-in-out infinite;
    animation-play-state: inherit;
  }
  i:nth-child(3n) { animation-delay: 0.12s; height: 62%; }
  i:nth-child(3n + 1) { animation-delay: 0.28s; height: 86%; }
  i:nth-child(4n) { animation-delay: 0.42s; height: 44%; }
}
@keyframes scWave { 0%, 100% { transform: scaleY(0.45); } 50% { transform: scaleY(1); } }
.sc-stop {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 999px;
  font-size: 14px;
  color: #ffd9dd;
  border: 1px solid rgba(255, 110, 130, 0.45);
  background: rgba(120, 40, 60, 0.3);
  opacity: 0;
  transform: translateY(8px);
  transition: opacity 0.45s ease, transform 0.45s var(--ease);
  i { width: 10px; height: 10px; border-radius: 2px; background: #ff7d92; }
  &.on { opacity: 1; transform: translateY(0); }
}
.sc-chips { bottom: 52px; }
.is-on .sc-chips .tpv-chip { animation-delay: 1.2s; }
.is-on .sc-chips .tpv-chip:nth-child(2) { animation-delay: 1.33s; }
.is-on .sc-chips .tpv-chip:nth-child(3) { animation-delay: 1.46s; }
.is-on .sc-chips .tpv-chip:nth-child(4) { animation-delay: 1.59s; }

// ───────────────────────── S5 可视化工作流 ─────────────────────────
.sf-graph {
  position: absolute;
  left: 56px;
  top: 178px;
  width: 880px;
  height: 508px;
  z-index: 2;
  opacity: 0;
}
.is-on .sf-graph { animation: tpvFadeIn 0.6s 0.3s both; }
.sf-svg { position: absolute; inset: 0; }
.sf-edge {
  fill: none;
  stroke: rgba(125, 165, 255, 0.2);
  stroke-width: 1.6;
  transition: stroke-dashoffset 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  &.no-trans { transition: none !important; }
  &.drawn { stroke: url(#tpvFlowGrad); stroke-opacity: 0.85; stroke-width: 2; }
}
.sf-dot { fill: var(--cyan); opacity: 0; filter: drop-shadow(0 0 6px rgba(124, 232, 255, 0.9)); }
.sf-node {
  position: absolute;
  width: 96px;
  height: 46px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  font-size: 17px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--txt);
  background: rgba(20, 34, 68, 0.6);
  border: 1px solid rgba(130, 168, 255, 0.28);
  transition: border-color 0.3s ease, box-shadow 0.3s ease, background 0.3s ease;
  &.lit {
    border-color: rgba(140, 160, 255, 0.8);
    background: rgba(40, 56, 110, 0.72);
    box-shadow: 0 0 24px rgba(110, 130, 255, 0.3);
  }
}
.sf-log {
  position: absolute;
  left: 980px;
  top: 178px;
  width: 240px;
  padding: 20px 22px;
  border-radius: 14px;
  background: rgba(13, 21, 44, 0.66);
  border: 1px solid rgba(130, 168, 255, 0.2);
  backdrop-filter: blur(8px);
  z-index: 2;
  opacity: 0;
}
.is-on .sf-log { animation: scWinIn 0.9s 0.5s var(--ease) both; }
.sf-log-title { margin: 0 0 10px; font-size: 14px; letter-spacing: 0.3em; color: var(--dim); }
.sf-log-time {
  margin: 0 0 14px;
  font-size: 34px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  background: linear-gradient(120deg, #dfe9ff, var(--blue));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  em { font-style: normal; font-size: 18px; margin-left: 2px; }
}
.sf-log ul { margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 9px; }
.sf-log li {
  display: flex;
  align-items: baseline;
  gap: 10px;
  font-size: 14.5px;
  opacity: 0;
  transform: translateX(14px);
  b { flex: none; font-weight: 650; color: #b9d2ff; }
  span { color: var(--dim); }
}
.is-on .sf-log li { animation: sfLogIn 0.5s var(--ease) both; }
@keyframes sfLogIn { to { opacity: 1; transform: translateX(0); } }
.sf-log-done {
  margin: 16px 0 0;
  padding-top: 14px;
  border-top: 1px dashed rgba(130, 168, 255, 0.25);
  font-size: 14.5px;
  letter-spacing: 0.06em;
  color: #8ce8c8;
  opacity: 0;
}
.is-on .sf-log-done { animation: tpvFadeIn 0.7s 4.6s both; }
.sf-chips { bottom: 48px; }
.is-on .sf-chips .tpv-chip { animation-delay: 1.4s; }
.is-on .sf-chips .tpv-chip:nth-child(2) { animation-delay: 1.54s; }
.is-on .sf-chips .tpv-chip:nth-child(3) { animation-delay: 1.68s; }
.is-on .sf-chips .tpv-chip:nth-child(4) { animation-delay: 1.82s; }

// ───────────────────────── S6 Agent 与工具 ─────────────────────────
.sa-orbit { position: absolute; inset: 0; z-index: 2; opacity: 0; }
.is-on .sa-orbit { animation: tpvFadeIn 0.7s 0.3s both; }
.sa-core {
  position: absolute;
  left: 552px;
  top: 358px;
  width: 176px;
  height: 88px;
  display: grid;
  place-items: center;
  gap: 4px;
  grid-auto-flow: row;
  border-radius: 22px;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.22em;
  text-indent: 0.22em;
  color: #fff;
  background: linear-gradient(135deg, rgba(77, 163, 255, 0.28), rgba(168, 85, 247, 0.26));
  border: 1px solid rgba(150, 150, 255, 0.5);
  box-shadow: 0 0 44px rgba(110, 120, 255, 0.35), inset 0 0 24px rgba(120, 140, 255, 0.18);
  backdrop-filter: blur(6px);
  animation: sgBreath 3s ease-in-out infinite alternate;
  i {
    position: absolute;
    top: 14px;
    left: 50%;
    width: 8px;
    height: 8px;
    margin-left: -4px;
    border-radius: 50%;
    background: var(--cyan);
    box-shadow: 0 0 12px var(--cyan);
    animation: sgBlink 1.6s ease-in-out infinite;
  }
}
.sa-chip {
  position: absolute;
  width: 132px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  font-size: 16.5px;
  letter-spacing: 0.05em;
  color: var(--txt);
  background: rgba(22, 36, 72, 0.6);
  border: 1px solid rgba(130, 168, 255, 0.3);
  backdrop-filter: blur(6px);
  opacity: 0;
  transform: scale(0.86);
}
.is-on .sa-chip { animation: saChipIn 0.7s var(--ease) both; }
.is-on .sa-chip:nth-of-type(2) { animation-delay: 0.5s; }
.is-on .sa-chip:nth-of-type(3) { animation-delay: 0.62s; }
.is-on .sa-chip:nth-of-type(4) { animation-delay: 0.74s; }
.is-on .sa-chip:nth-of-type(5) { animation-delay: 0.86s; }
.is-on .sa-chip:nth-of-type(6) { animation-delay: 0.98s; }
.is-on .sa-chip:nth-of-type(7) { animation-delay: 1.1s; }
@keyframes saChipIn { to { opacity: 1; transform: scale(1); } }
.sa-sub {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 56px;
  margin: 0;
  text-align: center;
  font-size: 18px;
  letter-spacing: 0.14em;
  color: var(--dim);
  opacity: 0;
}
.is-on .sa-sub { animation: tpvFadeIn 1s 1.5s both; }

// ───────────────────────── S7 架构与技术栈 ─────────────────────────
.ss-strata {
  position: absolute;
  left: 260px;
  top: 168px;
  width: 760px;
  display: flex;
  flex-direction: column;
  gap: 13px;
  z-index: 2;
}
.ss-bar {
  display: flex;
  align-items: center;
  gap: 18px;
  height: 58px;
  padding: 0 24px;
  border-radius: 13px;
  background: linear-gradient(90deg, rgba(32, 52, 104, 0.55), rgba(20, 32, 66, 0.4));
  border: 1px solid rgba(130, 168, 255, 0.22);
  border-left: 3px solid var(--blue);
  backdrop-filter: blur(6px);
  opacity: 0;
  transform: translateX(-36px);
  em { font-style: normal; font-size: 14px; color: rgba(160, 195, 255, 0.7); font-variant-numeric: tabular-nums; }
  b { font-size: 19px; font-weight: 700; letter-spacing: 0.06em; min-width: 96px; }
  span { font-size: 15.5px; color: var(--dim); letter-spacing: 0.03em; }
}
.is-on .ss-bar { animation: ssBarIn 0.8s var(--ease) both; }
@keyframes ssBarIn { to { opacity: 1; transform: translateX(0); } }
.ss-chips { bottom: 54px; }
.is-on .ss-chips .tpv-chip { animation-delay: 1.6s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 2) { animation-delay: 1.72s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 3) { animation-delay: 1.84s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 4) { animation-delay: 1.96s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 5) { animation-delay: 2.08s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 6) { animation-delay: 2.2s; }
.is-on .ss-chips .tpv-chip:nth-child(n + 7) { animation-delay: 2.32s; }

// ───────────────────────── S8 尾板 ─────────────────────────
.s-end { display: flex; align-items: center; justify-content: center; }
.se-wrap { text-align: center; }
.se-k { animation: tpvFadeIn 1s 0.2s both; }
.se-mark { margin: 26px 0 0; font-size: 92px; font-weight: 800; letter-spacing: 0.02em; }
.se-mark-txt {
  background: linear-gradient(110deg, #dfe9ff 20%, #9dc4ff 45%, var(--blue) 62%, #c9b2fb 80%);
  background-size: 200% 100%;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: soSweep 3s 0.4s var(--ease) both, tpvRise 1.1s 0.1s var(--ease) both;
}
.se-sub {
  margin: 26px 0 0;
  font-size: 21px;
  letter-spacing: 0.16em;
  color: var(--txt);
  opacity: 0;
}
.is-on .se-sub { animation: tpvFadeIn 1.2s 1.3s both; }
.se-link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-top: 40px;
  padding: 13px 26px;
  border-radius: 999px;
  font-size: 17px;
  letter-spacing: 0.05em;
  color: var(--txt);
  text-decoration: none;
  border: 1px solid rgba(130, 168, 255, 0.4);
  background: rgba(24, 40, 80, 0.5);
  backdrop-filter: blur(6px);
  transition: border-color 0.3s ease, box-shadow 0.3s ease, transform 0.3s var(--ease);
  opacity: 0;
  &:hover {
    border-color: rgba(124, 232, 255, 0.8);
    box-shadow: 0 0 30px rgba(96, 200, 255, 0.3);
    transform: translateY(-2px);
  }
}
.is-on .se-link { animation: tpvRise 0.9s 1.9s var(--ease) both; }
.se-foot { animation: none; opacity: 0; }
.is-on .se-foot { animation: tpvFadeIn 1.4s 2.6s both; }
.se-fade {
  position: absolute;
  inset: 0;
  background: #02030a;
  opacity: 0;
  pointer-events: none;
  z-index: 5;
}
.is-on .se-fade { animation: seFade 1.6s 5.2s ease both; }
@keyframes seFade { to { opacity: 1; } }

// ───────────────────────── 播放控制条 ─────────────────────────
.tpv-ctrl {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 30px 18px 12px;
  background: linear-gradient(180deg, transparent, rgba(2, 4, 12, 0.82) 62%);
  opacity: 1;
  transition: opacity 0.45s ease;
  cursor: default;
}
.ctrl-hide .tpv-ctrl { opacity: 0; pointer-events: none; }
.tpv-btn {
  flex: none;
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: rgba(230, 240, 255, 0.9);
  cursor: pointer;
  transition: background 0.25s ease, color 0.25s ease;
  &:hover { background: rgba(130, 168, 255, 0.18); color: #fff; }
}
.tpv-segs { flex: 1; display: flex; gap: 5px; align-items: center; }
.tpv-seg {
  position: relative;
  flex: 1;
  height: 14px;
  display: flex;
  align-items: center;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    height: 4px;
    border-radius: 2px;
    background: rgba(160, 185, 240, 0.28);
    transition: height 0.15s ease;
  }
  &:hover::before { height: 7px; }
  &:hover::after {
    content: attr(data-name);
    position: absolute;
    bottom: 22px;
    left: 50%;
    transform: translateX(-50%);
    padding: 5px 10px;
    border-radius: 8px;
    font-size: 12px;
    white-space: nowrap;
    color: #fff;
    background: rgba(8, 14, 30, 0.92);
    border: 1px solid rgba(130, 168, 255, 0.3);
    pointer-events: none;
  }
}
.tpv-seg-fill {
  position: relative;
  display: block;
  width: 0;
  height: 4px;
  border-radius: 2px;
  background: linear-gradient(90deg, var(--blue), var(--violet));
  pointer-events: none;
}
.tpv-scene-name {
  flex: none;
  font-size: 13px;
  letter-spacing: 0.14em;
  color: var(--dim);
  min-width: 72px;
  text-align: center;
}
.tpv-time {
  flex: none;
  font-size: 13px;
  color: var(--dim);
  font-variant-numeric: tabular-nums;
  b { color: rgba(235, 243, 255, 0.95); font-weight: 600; }
}

// 结束遮罩
.tpv-endmask {
  position: absolute;
  inset: 0;
  z-index: 12;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 26px;
  cursor: pointer;
  background: radial-gradient(100% 100% at 50% 40%, rgba(10, 18, 40, 0.6), rgba(2, 3, 10, 0.88));
  backdrop-filter: blur(2px);
}
.tpv-end-title { font-size: 17px; letter-spacing: 0.3em; color: var(--dim); }
.tpv-end-replay {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 30px;
  border-radius: 999px;
  font-size: 16px;
  letter-spacing: 0.08em;
  color: #fff;
  border: 1px solid rgba(130, 168, 255, 0.5);
  background: rgba(30, 48, 96, 0.5);
  transition: box-shadow 0.3s ease, border-color 0.3s ease, transform 0.3s var(--ease);
  &:hover {
    border-color: var(--cyan);
    box-shadow: 0 0 34px rgba(96, 200, 255, 0.35);
    transform: translateY(-2px);
  }
}
.tpv-fade-enter-active, .tpv-fade-leave-active { transition: opacity 0.5s ease; }
.tpv-fade-enter-from, .tpv-fade-leave-to { opacity: 0; }

// 暂停态呼吸角标（左上角小圆点）
.tpv::after {
  content: 'PAUSED';
  position: absolute;
  top: 14px;
  left: 16px;
  z-index: 9;
  padding: 5px 12px;
  border-radius: 999px;
  font-size: 11px;
  letter-spacing: 0.28em;
  text-indent: 0.28em;
  color: rgba(230, 240, 255, 0.85);
  background: rgba(8, 14, 30, 0.6);
  border: 1px solid rgba(130, 168, 255, 0.3);
  opacity: 0;
  transition: opacity 0.4s ease;
  pointer-events: none;
}
.tpv.is-paused::after { opacity: 1; }
</style>
