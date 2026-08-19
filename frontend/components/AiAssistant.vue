<script setup lang="ts">
// 全局 AI 助手「多晓」(DuoXiao)：可拖动悬浮球 + 流式聊天窗，后端代理智谱 GLM。
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import DxIcon from '~/components/DxIcon.vue'
import { getAccessToken } from '~/utils/token'

interface Msg {
  role: 'user' | 'assistant'
  content: string
}

const BALL_SIZE = 56
const route = useRoute()

const open = ref(false)
const input = ref('')
const sending = ref(false)
const messages = ref<Msg[]>([])
const listRef = ref<HTMLElement | null>(null)

// ---------- 悬浮球拖动（Pointer Events，位置持久化到 localStorage） ----------
const pos = ref({ x: 0, y: 0 })
let dragState: { startX: number; startY: number; baseX: number; baseY: number; moved: boolean } | null = null

function clamp(v: number, min: number, max: number) {
  return Math.min(Math.max(v, min), max)
}

function defaultPos() {
  return { x: window.innerWidth - BALL_SIZE - 28, y: window.innerHeight - BALL_SIZE - 96 }
}

function savePos() {
  localStorage.setItem('dx-ai-ball', JSON.stringify(pos.value))
}

function loadPos() {
  try {
    const raw = localStorage.getItem('dx-ai-ball')
    if (raw) {
      const p = JSON.parse(raw)
      if (typeof p.x === 'number' && typeof p.y === 'number') {
        pos.value = {
          x: clamp(p.x, 0, window.innerWidth - BALL_SIZE),
          y: clamp(p.y, 0, window.innerHeight - BALL_SIZE),
        }
        return
      }
    }
  } catch { /* 忽略损坏的存储 */ }
  pos.value = defaultPos()
}

function onPointerDown(e: PointerEvent) {
  dragState = {
    startX: e.clientX,
    startY: e.clientY,
    baseX: pos.value.x,
    baseY: pos.value.y,
    moved: false,
  }
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!dragState) return
  const dx = e.clientX - dragState.startX
  const dy = e.clientY - dragState.startY
  if (Math.abs(dx) > 4 || Math.abs(dy) > 4) dragState.moved = true
  pos.value = {
    x: clamp(dragState.baseX + dx, 0, window.innerWidth - BALL_SIZE),
    y: clamp(dragState.baseY + dy, 0, window.innerHeight - BALL_SIZE),
  }
}

function onPointerUp() {
  const st = dragState
  dragState = null
  if (!st) return
  if (st.moved) {
    savePos()
  } else {
    toggleChat()
  }
}

function onResize() {
  pos.value = {
    x: clamp(pos.value.x, 0, window.innerWidth - BALL_SIZE),
    y: clamp(pos.value.y, 0, window.innerHeight - BALL_SIZE),
  }
}

// 球在视口左半边 → 聊天窗出现在球右侧，反之在左侧
const panelStyle = computed(() => {
  const onRight = pos.value.x < window.innerWidth / 2
  return {
    left: onRight ? `${pos.value.x + BALL_SIZE + 14}px` : 'auto',
    right: onRight ? 'auto' : `${window.innerWidth - pos.value.x + 14}px`,
    top: `${clamp(pos.value.y - 20, 12, Math.max(12, window.innerHeight - 560))}px`,
  }
})

// ---------- 对话 ----------
const quickQuestions = [
  '怎么访问我创建的 Nginx 实例？',
  '创建 ECS 实例的具体步骤？',
  'CI/CD 流水线怎么配置？',
  '云磁盘怎么挂载到实例？',
]

function toggleChat() {
  open.value = !open.value
  if (open.value && messages.value.length === 0) {
    messages.value = [{
      role: 'assistant',
      content: '你好，我是「多晓云」DxCloud 平台的 AI 助手「多晓」👋\n\n我与平台同名——多晓（DuoXiao）即 Dx 的谐音，寓意「通晓云上万事」。平台的事，问我准没错！\n\n我熟悉平台的全部功能：ECS 云主机、镜像中心、网络与存储、PaaS 应用、CI/CD 流水线、监控计费等。有任何使用问题，直接问我吧！',
    }]
  }
  nextTick(scrollToBottom)
}

function scrollToBottom() {
  const el = listRef.value
  if (el) el.scrollTop = el.scrollHeight
}

const pageContext = computed(() => {
  const name = (route.meta?.title as string) || route.path
  return `${name}（路由 ${route.path}）`
})

async function send(text?: string) {
  const q = (text ?? input.value).trim()
  if (!q || sending.value) return
  input.value = ''
  messages.value.push({ role: 'user', content: q })
  const reply: Msg = { role: 'assistant', content: '' }
  messages.value.push(reply)
  sending.value = true
  nextTick(scrollToBottom)

  try {
    const res = await fetch('/api/v1/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getAccessToken()}`,
      },
      body: JSON.stringify({
        messages: messages.value.slice(0, -1).map(m => ({ role: m.role, content: m.content })),
        page_context: pageContext.value,
      }),
    })

    if (!res.ok || !res.body) {
      let msg = `请求失败（HTTP ${res.status}）`
      try {
        const j = await res.json()
        if (j?.message) msg = j.message
      } catch { /* 非 JSON 错误体 */ }
      throw new Error(msg)
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buf = ''
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buf += decoder.decode(value, { stream: true })
      const blocks = buf.split('\n\n')
      buf = blocks.pop() ?? ''
      for (const block of blocks) {
        const line = block.trim()
        if (!line.startsWith('data:')) continue
        let data: { delta?: string; error?: string; done?: string }
        try {
          data = JSON.parse(line.slice(5).trim())
        } catch {
          continue
        }
        if (data.delta) {
          reply.content += data.delta
          nextTick(scrollToBottom)
        }
        if (data.error) throw new Error(data.error)
        if (data.done) {
          reader.cancel().catch(() => {})
          break
        }
      }
    }
    if (!reply.content) throw new Error('AI 未返回内容，请重试')
  } catch (e) {
    if (!reply.content) {
      reply.content = `⚠️ ${e instanceof Error ? e.message : '出错了，请稍后重试'}`
    }
  } finally {
    sending.value = false
    nextTick(scrollToBottom)
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function clearChat() {
  messages.value = []
  toggleChat()
}

// ---------- 轻量 Markdown 渲染（先转义再转换，防 XSS） ----------
function esc(s: string) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

function renderMd(text: string) {
  let s = esc(text)
  s = s.replace(/```(\w*)\n?([\s\S]*?)(```|$)/g, (_m, _lang, code) =>
    `<pre class="ai-md-code">${code.replace(/\n$/, '')}</pre>`)
  s = s.replace(/`([^`\n]+)`/g, '<code class="ai-md-inline">$1</code>')
  s = s.replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
  const lines = s.split('\n')
  const out: string[] = []
  let inList = false
  for (const line of lines) {
    const listItem = line.match(/^\s*[-*]\s+(.*)/)
    if (listItem) {
      if (!inList) { out.push('<ul>'); inList = true }
      out.push(`<li>${listItem[1]}</li>`)
      continue
    }
    if (inList) { out.push('</ul>'); inList = false }
    if (line.trim() === '') out.push('<br>')
    else out.push(`<p>${line}</p>`)
  }
  if (inList) out.push('</ul>')
  return out.join('')
}

// ---------- 生命周期 ----------
let resizeHandler = () => onResize()

onMounted(() => {
  loadPos()
  window.addEventListener('resize', resizeHandler)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeHandler)
})

watch(open, v => {
  if (v) document.body.style.overflow = ''
})
</script>

<template>
  <!-- 悬浮球 -->
  <div
    class="ai-ball"
    :class="{ 'ai-ball-active': open }"
    :style="{ left: `${pos.x}px`, top: `${pos.y}px` }"
    @pointerdown.prevent="onPointerDown"
    @pointermove="onPointerMove"
    @pointerup="onPointerUp"
    @pointercancel="onPointerUp"
    title="AI 助手 · 拖动调整位置，点击对话"
  >
    <div class="ai-ball-ring" />
    <img src="/ai.png" alt="AI" draggable="false" class="ai-ball-img">
    <span v-if="!open" class="ai-ball-tip">问我</span>
  </div>

  <!-- 聊天窗 -->
  <Transition name="ai-panel">
    <div v-if="open" class="ai-panel" :style="panelStyle">
      <header class="ai-panel-header">
        <img src="/ai.png" alt="" class="ai-panel-avatar">
        <div class="ai-panel-title">
          <strong>多晓 DuoXiao · DxCloud 助手</strong>
          <span>Powered by GLM · 熟知平台全模块</span>
        </div>
        <button class="ai-panel-btn" title="清空对话" @click="clearChat">
          <DxIcon name="refresh" :size="14" />
        </button>
        <button class="ai-panel-btn" title="收起" @click="open = false">
          <span class="ai-panel-close">✕</span>
        </button>
      </header>

      <div ref="listRef" class="ai-panel-body">
        <template v-if="messages.length === 0">
          <div class="ai-empty">
            <img src="/ai.png" alt="" class="ai-empty-img">
            <p>你好！我是多晓，随时为你解答平台使用问题</p>
          </div>
        </template>
        <div
          v-for="(m, i) in messages"
          :key="i"
          class="ai-msg"
          :class="m.role === 'user' ? 'ai-msg-user' : 'ai-msg-bot'"
        >
          <img v-if="m.role === 'assistant'" src="/ai.png" alt="" class="ai-msg-avatar">
          <div class="ai-msg-bubble" v-html="renderMd(m.content)" />
        </div>
        <div v-if="sending && !messages[messages.length - 1]?.content" class="ai-msg ai-msg-bot">
          <img src="/ai.png" alt="" class="ai-msg-avatar">
          <div class="ai-msg-bubble ai-typing">
            <i /><i /><i />
          </div>
        </div>
      </div>

      <div v-if="messages.length <= 1" class="ai-quick">
        <button v-for="q in quickQuestions" :key="q" @click="send(q)">{{ q }}</button>
      </div>

      <footer class="ai-panel-footer">
        <textarea
          v-model="input"
          rows="1"
          placeholder="输入问题，Enter 发送，Shift+Enter 换行"
          :disabled="sending"
          @keydown="onKeydown"
        />
        <button class="ai-send" :disabled="sending || !input.trim()" @click="send()">
          <DxIcon name="send" :size="16" />
        </button>
      </footer>
    </div>
  </Transition>
</template>

<style scoped>
/* ---------- 悬浮球 ---------- */
.ai-ball {
  position: fixed;
  z-index: 3000;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  cursor: grab;
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  touch-action: none;
  filter: drop-shadow(0 6px 16px rgba(0, 110, 255, 0.35));
  transition: transform 0.2s ease;
}
.ai-ball:active {
  cursor: grabbing;
  transform: scale(0.94);
}
.ai-ball:hover {
  transform: scale(1.08);
}
.ai-ball-active {
  transform: scale(1.02);
}
.ai-ball-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  pointer-events: none;
  border: 2px solid rgba(255, 255, 255, 0.9);
  box-shadow: 0 0 0 2px rgba(0, 110, 255, 0.25);
}
.ai-ball-ring {
  position: absolute;
  inset: -6px;
  border-radius: 50%;
  border: 2px solid rgba(0, 110, 255, 0.35);
  animation: ai-pulse 2.4s ease-out infinite;
  pointer-events: none;
}
@keyframes ai-pulse {
  0% { transform: scale(0.92); opacity: 0.9; }
  70% { transform: scale(1.25); opacity: 0; }
  100% { transform: scale(1.25); opacity: 0; }
}
.ai-ball-tip {
  position: absolute;
  top: -24px;
  left: 50%;
  transform: translateX(-50%);
  background: #1f2329;
  color: #fff;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
  opacity: 0;
  transition: opacity 0.2s;
  pointer-events: none;
}
.ai-ball:hover .ai-ball-tip {
  opacity: 1;
}

/* ---------- 聊天窗 ---------- */
.ai-panel {
  position: fixed;
  z-index: 3001;
  width: 380px;
  max-width: calc(100vw - 24px);
  height: 540px;
  max-height: calc(100vh - 24px);
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}
.ai-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: linear-gradient(135deg, #006eff 0%, #3d8bff 100%);
  color: #fff;
  flex-shrink: 0;
}
.ai-panel-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.8);
  object-fit: cover;
}
.ai-panel-title {
  flex: 1;
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}
.ai-panel-title strong {
  font-size: 14px;
}
.ai-panel-title span {
  font-size: 11px;
  opacity: 0.85;
}
.ai-panel-btn {
  border: none;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}
.ai-panel-btn:hover {
  background: rgba(255, 255, 255, 0.32);
}
.ai-panel-close {
  font-size: 13px;
  line-height: 1;
}

.ai-panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  background: #f7f9fc;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ai-empty {
  text-align: center;
  padding: 32px 0;
  color: #86909c;
  font-size: 13px;
}
.ai-empty-img {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  margin-bottom: 10px;
  opacity: 0.9;
}
.ai-msg {
  display: flex;
  gap: 8px;
  max-width: 92%;
  animation: ai-msg-in 0.25s ease;
}
@keyframes ai-msg-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
.ai-msg-bot {
  align-self: flex-start;
}
.ai-msg-user {
  align-self: flex-end;
  flex-direction: row-reverse;
}
.ai-msg-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 2px;
  object-fit: cover;
}
.ai-msg-bubble {
  padding: 9px 12px;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.65;
  word-break: break-word;
}
.ai-msg-bot .ai-msg-bubble {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-top-left-radius: 2px;
  color: #1f2329;
}
.ai-msg-user .ai-msg-bubble {
  background: #006eff;
  color: #fff;
  border-top-right-radius: 2px;
}
.ai-msg-bubble :deep(p) {
  margin: 0 0 4px;
}
.ai-msg-bubble :deep(p):last-child {
  margin-bottom: 0;
}
.ai-msg-bubble :deep(ul) {
  margin: 4px 0;
  padding-left: 18px;
}
.ai-msg-bubble :deep(li) {
  margin: 2px 0;
}
.ai-md-code,
.ai-msg-bubble :deep(.ai-md-code) {
  background: #0d1117;
  color: #e6edf3;
  border-radius: 6px;
  padding: 10px 12px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
  margin: 6px 0;
  overflow-x: auto;
  white-space: pre;
}
.ai-msg-bubble :deep(.ai-md-inline) {
  background: rgba(0, 110, 255, 0.08);
  color: #006eff;
  padding: 1px 5px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
}
.ai-msg-user .ai-msg-bubble :deep(.ai-md-inline) {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.ai-typing {
  display: flex;
  gap: 4px;
  align-items: center;
  padding: 12px 14px;
}
.ai-typing i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #006eff;
  opacity: 0.4;
  animation: ai-blink 1.2s infinite;
}
.ai-typing i:nth-child(2) { animation-delay: 0.2s; }
.ai-typing i:nth-child(3) { animation-delay: 0.4s; }
@keyframes ai-blink {
  0%, 80%, 100% { opacity: 0.3; transform: scale(0.9); }
  40% { opacity: 1; transform: scale(1.1); }
}

.ai-quick {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0 14px 8px;
  background: #f7f9fc;
  flex-shrink: 0;
}
.ai-quick button {
  border: 1px solid #d4e5ff;
  background: #fff;
  color: #006eff;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.ai-quick button:hover {
  background: #006eff;
  color: #fff;
  border-color: #006eff;
}

.ai-panel-footer {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 12px;
  border-top: 1px solid #e8e8e8;
  background: #fff;
  flex-shrink: 0;
}
.ai-panel-footer textarea {
  flex: 1;
  border: 1px solid #e0e6ed;
  border-radius: 8px;
  padding: 8px 10px;
  font-size: 13px;
  resize: none;
  outline: none;
  max-height: 96px;
  font-family: inherit;
  background: #fafbfc;
  transition: border-color 0.2s;
}
.ai-panel-footer textarea:focus {
  border-color: #006eff;
}
.ai-send {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 8px;
  background: #006eff;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}
.ai-send:hover:not(:disabled) {
  background: #0052d9;
}
.ai-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---------- 弹窗动画 ---------- */
.ai-panel-enter-active {
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.ai-panel-leave-active {
  transition: all 0.18s ease-in;
}
.ai-panel-enter-from,
.ai-panel-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
}

/* ---------- 深色模式 ---------- */
html.dark .ai-ball {
  filter: drop-shadow(0 6px 20px rgba(61, 139, 255, 0.55));
}
html.dark .ai-ball-img {
  border-color: rgba(230, 237, 243, 0.8);
  box-shadow: 0 0 0 2px rgba(61, 139, 255, 0.45);
}
html.dark .ai-ball-ring {
  border-color: rgba(61, 139, 255, 0.6);
}
html.dark .ai-panel {
  background: #161b22;
  border-color: #30363d;
}
html.dark .ai-panel-body,
html.dark .ai-quick {
  background: #0d1117;
}
html.dark .ai-msg-bot .ai-msg-bubble {
  background: #21262d;
  border-color: #30363d;
  color: #e6edf3;
}
html.dark .ai-msg-bubble :deep(.ai-md-inline) {
  background: rgba(61, 139, 255, 0.15);
  color: #3d8bff;
}
html.dark .ai-panel-footer {
  background: #161b22;
  border-color: #30363d;
}
html.dark .ai-panel-footer textarea {
  background: #0d1117;
  border-color: #30363d;
  color: #e6edf3;
}
html.dark .ai-quick button {
  background: #161b22;
  border-color: #2b4a75;
  color: #3d8bff;
}
</style>
