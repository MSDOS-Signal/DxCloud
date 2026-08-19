<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { api } from '~/services/http'

definePageMeta({ layout: 'terminal' })

const route = useRoute()
const id = computed(() => Number(route.params.id))
const termEl = ref<HTMLDivElement | null>(null)
const statusText = ref('连接中…')
const connected = ref(false)

let term: Terminal | null = null
let fit: FitAddon | null = null
let ws: WebSocket | null = null
let destroyed = false

function wsUrl(token: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${window.location.host}/ws/v1/ecs/${id.value}/terminal?token=${encodeURIComponent(token)}`
}

async function connect() {
  if (destroyed) return
  statusText.value = '获取会话令牌…'
  try {
    const res = await api.post<{ token: string }>(`/ecs/${id.value}/exec`)
    const token = res.token
    statusText.value = '连接中…'
    ws = new WebSocket(wsUrl(token))
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      statusText.value = '已连接'
      connected.value = true
      if (fit && term) {
        fit.fit()
        const dims = fit.proposeDimensions()
        if (dims) {
          ws!.send(JSON.stringify({ type: 'resize', cols: dims.cols, rows: dims.rows }))
        }
      }
    }
    ws.onmessage = (ev) => {
      if (ev.data instanceof ArrayBuffer) {
        term?.write(new Uint8Array(ev.data))
      }
    }
    ws.onclose = (ev) => {
      if (destroyed) return
      connected.value = false
      statusText.value = ev.code === 1000 ? '会话已结束' : `连接断开（${ev.code}${ev.reason ? '：' + ev.reason : ''}），可重连`
    }
    ws.onerror = () => {
      connected.value = false
      statusText.value = '连接出错，可重连'
    }
  } catch (e) {
    statusText.value = `获取令牌失败：${e instanceof Error ? e.message : String(e)}`
  }
}

function reconnect() {
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
  if (term) {
    term.reset()
  }
  connect()
}

function clearTerminal() {
  term?.clear()
}

async function copySelection() {
  const text = term?.getSelection()
  if (text && navigator.clipboard) {
    await navigator.clipboard.writeText(text)
  }
}

onMounted(() => {
  if (!termEl.value) return
  term = new Terminal({
    cursorBlink: true,
    fontSize: 13,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#0f172a',
      foreground: '#e2e8f0',
      cursor: '#38bdf8',
    },
    scrollback: 2000,
  })
  fit = new FitAddon()
  term.loadAddon(fit)
  term.open(termEl.value)
  fit.fit()

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      // 终端输入必须走二进制帧（文本帧在服务端被解释为 resize 等 JSON 控制）
      ws.send(new TextEncoder().encode(data))
    }
  })
  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })

  connect()
  window.addEventListener('resize', () => fit?.fit())
})

onBeforeUnmount(() => {
  destroyed = true
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
  term?.dispose()
  term = null
})
</script>

<template>
  <div class="h-full flex flex-col">
    <div class="flex flex-wrap items-center justify-between gap-2 px-3 py-1 bg-slate-800 text-xs text-gray-300">
      <span class="inline-flex items-center gap-2">
        <span class="w-2 h-2 rounded-full" :class="connected ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,.8)]' : 'bg-amber-400'"></span>
        <span>{{ statusText }}</span>
        <span class="text-gray-500">Web 终端 · 实例 #{{ id }}</span>
      </span>
      <div class="flex items-center gap-2">
        <n-button size="tiny" quaternary @click="copySelection">复制所选</n-button>
        <n-button size="tiny" quaternary @click="clearTerminal">清屏</n-button>
        <n-button size="tiny" @click="reconnect">重新连接</n-button>
      </div>
    </div>
    <div ref="termEl" class="flex-1 min-h-0 p-1" />
  </div>
</template>
