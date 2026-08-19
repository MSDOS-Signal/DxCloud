<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { EChartsOption } from '~/composables/useEcharts'
import { useEcharts } from '~/composables/useEcharts'
import type { LineSeriesOption } from 'echarts/charts'
import type { TooltipComponentOption } from 'echarts/components'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'

const route = useRoute()
const theme = useThemeStore()
const id = computed(() => Number(route.params.id))
const minutes = ref(30)
const cpuEl = ref<HTMLElement | null>(null)
const memEl = ref<HTMLElement | null>(null)
const netEl = ref<HTMLElement | null>(null)
const cpuChart = useEcharts(cpuEl)
const memChart = useEcharts(memEl)
const netChart = useEcharts(netEl)
const chartError = ref('')
const seriesEmpty = ref(false)
let timer: number | undefined

interface SeriesRow {
  bucket: string
  cpu: number | null
  mem_pct: number | null
  net_rx: number | null
  net_tx: number | null
}

// ---------- 图表视觉规范：腾讯云蓝 · 纯色折线 + 轻渐变面积 · 玻璃拟态 tooltip ----------
const BLUE: [string, string] = ['#006eff', '#00c6ff']
const CYAN: [string, string] = ['#00a4ff', '#7be0ff']
const GREEN: [string, string] = ['#00b42a', '#5ad8a6']

function rgba(hex: string, alpha: number) {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${alpha})`
}

// 深浅色轴主题（与 dashboard 保持一致）
function axisTheme() {
  return theme.isDark
    ? { label: '#64748B', split: '#1f2937', line: '#334155' }
    : { label: '#86909c', split: '#e5e6eb', line: '#e8e8e8' }
}

function fmtBps(v: number) {
  if (v >= 1048576) return `${(v / 1048576).toFixed(1)} MB/s`
  if (v >= 1024) return `${(v / 1024).toFixed(1)} KB/s`
  return `${Math.round(v)} B/s`
}

const glassTooltip: TooltipComponentOption = {
  trigger: 'axis',
  backgroundColor: 'rgba(13,25,42,0.85)',
  borderColor: 'transparent',
  borderWidth: 0,
  padding: [8, 12],
  textStyle: { color: '#fff', fontSize: 12 },
  extraCssText:
    'backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); box-shadow: 0 8px 24px rgba(2,16,43,0.35); border-radius: 6px;',
  axisPointer: { lineStyle: { color: 'rgba(0,110,255,0.4)' } },
}

// 纯色折线 + 两段渐变面积；隐藏 symbol，hover 时聚焦显示
function gradLine(name: string, values: number[], pair: [string, string]): LineSeriesOption {
  const [from, to] = pair
  return {
    name,
    type: 'line',
    smooth: 0.3,
    showSymbol: false,
    symbol: 'circle',
    symbolSize: 7,
    sampling: 'lttb',
    emphasis: { focus: 'series', itemStyle: { borderWidth: 2, borderColor: '#fff', color: from } },
    lineStyle: {
      width: 2.2,
      color: from,
    },
    itemStyle: { color: from },
    areaStyle: {
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: rgba(from, 0.18) },
          { offset: 1, color: rgba(to, 0) },
        ],
      },
    },
    data: values,
  }
}

function buildBase(x: string[], mode: 'pct' | 'bps', legend?: string[]): EChartsOption {
  const t = axisTheme()
  const opt: EChartsOption = {
    animationDuration: 800,
    animationEasing: 'cubicOut',
    grid: { left: 10, right: 14, top: legend ? 30 : 14, bottom: 4, containLabel: true },
    tooltip: {
      ...glassTooltip,
      valueFormatter:
        mode === 'pct'
          ? (v) => `${Number(v ?? 0).toFixed(1)}%`
          : (v) => fmtBps(Number(v ?? 0)),
    },
    xAxis: {
      type: 'category',
      data: x,
      boundaryGap: false,
      axisLine: { lineStyle: { color: t.line } },
      axisTick: { show: false },
      axisLabel: { color: t.label, fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      min: 0,
      ...(mode === 'pct' ? { max: 100 } : {}),
      axisLabel: {
        color: t.label,
        fontSize: 11,
        formatter: mode === 'pct' ? '{value}%' : (v: number) => fmtBps(v),
      },
      splitLine: { lineStyle: { color: t.split, type: 'dashed' } },
    },
  }
  if (legend) {
    opt.legend = {
      data: legend,
      top: 2,
      right: 8,
      icon: 'roundRect',
      itemWidth: 10,
      itemHeight: 3,
      itemGap: 12,
      textStyle: { color: t.label, fontSize: 11 },
    }
  }
  return opt
}

async function load() {
  try {
    const series = await api.get<SeriesRow[]>(`/monitor/series?kind=ecs&ref_id=${id.value}&minutes=${minutes.value}`)
    // 请求成功后清空错误态
    chartError.value = ''
    seriesEmpty.value = series.length === 0

    const x = series.map((r) => r.bucket.slice(11, 16))
    const cpuData = series.map((r) => (r.cpu == null ? 0 : Number(r.cpu)))
    const memData = series.map((r) => (r.mem_pct == null ? 0 : Number(r.mem_pct)))
    const rxData = series.map((r) => (r.net_rx == null ? 0 : Number(r.net_rx)))
    const txData = series.map((r) => (r.net_tx == null ? 0 : Number(r.net_tx)))
    cpuChart.render({
      ...buildBase(x, 'pct'),
      series: [gradLine('CPU 使用率', cpuData, BLUE)],
    })
    memChart.render({
      ...buildBase(x, 'pct'),
      series: [gradLine('内存使用率', memData, CYAN)],
    })
    netChart.render({
      ...buildBase(x, 'bps', ['接收', '发送']),
      series: [
        gradLine('接收', rxData, BLUE),
        gradLine('发送', txData, GREEN),
      ],
    })
  } catch (e) {
    chartError.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 10000)
})

// 深浅色主题切换时按新轴色重绘
watch(() => theme.isDark, () => load())

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
  cpuChart.dispose()
  memChart.dispose()
  netChart.dispose()
})
</script>

<template>
  <div class="space-y-4">
    <div class="dx-card dx-fade-up">
      <div class="dx-card-body">
        <div class="flex items-center justify-between">
          <span class="font-semibold">实例监控 · #{{ id }}</span>
          <n-space>
            <n-radio-group v-model:value="minutes" size="small" @update:value="load">
              <n-radio-button :value="30">30 分钟</n-radio-button>
              <n-radio-button :value="120">2 小时</n-radio-button>
              <n-radio-button :value="360">6 小时</n-radio-button>
            </n-radio-group>
            <n-button size="small" @click="load">刷新</n-button>
          </n-space>
        </div>
        <div class="text-xs text-gray-400 mt-1">指标由后台采样器每分钟采集并落库（metric_samples，保留 7 天）</div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="chartError" class="dx-chart-error dx-fade-in" :title="chartError">
      <span class="dx-chart-error-text">
        <DxIcon name="info" :size="14" />
        监控数据加载失败，点击重试
      </span>
      <button class="dx-chart-retry" @click="load">
        <DxIcon name="refresh" :size="12" /> 重试
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">
          <div class="flex items-center gap-2">
            <DxIcon name="cpu" :size="15" class="text-[#006eff]" />
            <span>CPU 使用率</span>
          </div>
        </div>
        <div class="dx-card-body">
          <div class="relative">
            <div ref="cpuEl" class="dx-chart-box" />
            <div v-if="seriesEmpty && !chartError" class="dx-chart-empty dx-fade-in">
              <svg width="46" height="46" viewBox="0 0 48 48" fill="none" aria-hidden="true">
                <rect x="7" y="9" width="34" height="27" rx="3" stroke="currentColor" stroke-opacity="0.28" stroke-width="1.5" stroke-dasharray="3 3" />
                <path d="M13 29.5l6.5-7 5 4.2L32.5 17" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M13 33.5h22" stroke="currentColor" stroke-opacity="0.18" stroke-width="1.5" stroke-linecap="round" />
                <circle cx="32.5" cy="17" r="1.8" fill="currentColor" fill-opacity="0.5" />
              </svg>
              <span>暂无监控数据</span>
            </div>
          </div>
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">
          <div class="flex items-center gap-2">
            <DxIcon name="memory" :size="15" class="text-[#00a4ff]" />
            <span>内存使用率</span>
          </div>
        </div>
        <div class="dx-card-body">
          <div class="relative">
            <div ref="memEl" class="dx-chart-box" />
            <div v-if="seriesEmpty && !chartError" class="dx-chart-empty dx-fade-in">
              <svg width="46" height="46" viewBox="0 0 48 48" fill="none" aria-hidden="true">
                <rect x="7" y="9" width="34" height="27" rx="3" stroke="currentColor" stroke-opacity="0.28" stroke-width="1.5" stroke-dasharray="3 3" />
                <path d="M13 29.5l6.5-7 5 4.2L32.5 17" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M13 33.5h22" stroke="currentColor" stroke-opacity="0.18" stroke-width="1.5" stroke-linecap="round" />
                <circle cx="32.5" cy="17" r="1.8" fill="currentColor" fill-opacity="0.5" />
              </svg>
              <span>暂无监控数据</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="dx-card dx-fade-up dx-delay-2">
      <div class="dx-card-header">
        <div class="flex items-center gap-2">
          <DxIcon name="networks" :size="15" class="text-[#006eff]" />
          <span>网络吞吐（B/s）</span>
        </div>
      </div>
      <div class="dx-card-body">
        <div class="relative">
          <div ref="netEl" class="dx-chart-box" />
          <div v-if="seriesEmpty && !chartError" class="dx-chart-empty dx-fade-in">
            <svg width="46" height="46" viewBox="0 0 48 48" fill="none" aria-hidden="true">
              <rect x="7" y="9" width="34" height="27" rx="3" stroke="currentColor" stroke-opacity="0.28" stroke-width="1.5" stroke-dasharray="3 3" />
              <path d="M13 29.5l6.5-7 5 4.2L32.5 17" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
              <path d="M13 33.5h22" stroke="currentColor" stroke-opacity="0.18" stroke-width="1.5" stroke-linecap="round" />
              <circle cx="32.5" cy="17" r="1.8" fill="currentColor" fill-opacity="0.5" />
            </svg>
            <span>暂无监控数据</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dx-chart-box {
  width: 100%;
  height: 200px;
}

/* ---------- 空态 ---------- */
.dx-chart-empty {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 12px;
  color: #b3bac6;
  background: #fff;
  pointer-events: none;
}
html.dark .dx-chart-empty {
  background: #161b22;
  color: #57616e;
}

/* ---------- 错误提示 ---------- */
.dx-chart-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border: 1px solid rgba(245, 63, 63, 0.28);
  border-radius: 4px;
  background: rgba(245, 63, 63, 0.05);
  color: #f53f3f;
  font-size: 12px;
}
html.dark .dx-chart-error {
  background: rgba(245, 63, 63, 0.08);
  border-color: rgba(245, 63, 63, 0.35);
}
.dx-chart-error-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.dx-chart-retry {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  height: 24px;
  padding: 0 10px;
  border: 1px solid rgba(245, 63, 63, 0.4);
  border-radius: 3px;
  background: transparent;
  color: #f53f3f;
  font-size: 12px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}
.dx-chart-retry:hover {
  background: rgba(245, 63, 63, 0.1);
  border-color: #f53f3f;
}
</style>
