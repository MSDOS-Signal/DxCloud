<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { EChartsOption } from '~/composables/useEcharts'
import { useEcharts } from '~/composables/useEcharts'
import type { LineSeriesOption } from 'echarts/charts'

export interface CloudLineSeries {
  name: string
  data: number[]
  from?: string
  to?: string
}

const props = withDefaults(defineProps<{
  categories: string[]
  series: CloudLineSeries[]
  mode?: 'percent' | 'bytes'
  height?: number
  loading?: boolean
  empty?: boolean
  error?: string
  showLegend?: boolean
}>(), {
  mode: 'percent',
  height: 220,
  loading: false,
  empty: false,
  error: '',
  showLegend: true,
})

const emit = defineEmits<{ retry: [] }>()
const chartEl = ref<HTMLElement | null>(null)
const chart = useEcharts(chartEl)
const theme = useThemeStore()

function rgba(hex: string, alpha: number) {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${alpha})`
}

function axisTheme() {
  return theme.isDark
    ? { label: '#64748b', split: '#1f2937', line: '#334155' }
    : { label: '#86909c', split: '#e5e6eb', line: '#e8e8e8' }
}

function fmtBytes(v: number) {
  if (v >= 1048576) return `${(v / 1048576).toFixed(1)} MB/s`
  if (v >= 1024) return `${(v / 1024).toFixed(1)} KB/s`
  return `${Math.round(v)} B/s`
}

function seriesOption(item: CloudLineSeries): LineSeriesOption {
  const from = item.from || '#006eff'
  const to = item.to || '#00c6ff'
  return {
    name: item.name,
    type: 'line',
    smooth: 0.3,
    connectNulls: true,
    showSymbol: false,
    symbol: 'circle',
    symbolSize: 7,
    sampling: 'lttb',
    emphasis: {
      focus: 'series',
      lineStyle: { width: 3 },
      itemStyle: { borderWidth: 2, borderColor: '#fff', color: from },
    },
    lineStyle: {
      width: 2.2,
      color: from,
    },
    itemStyle: { color: from },
    areaStyle: {
      color: {
        type: 'linear',
        x: 0,
        y: 0,
        x2: 0,
        y2: 1,
        colorStops: [
          { offset: 0, color: rgba(from, 0.18) },
          { offset: 1, color: rgba(to, 0) },
        ],
      },
    },
    data: item.data,
  }
}

const option = computed<EChartsOption>(() => {
  const t = axisTheme()
  const isPercent = props.mode === 'percent'
  const showLegend = props.showLegend && props.series.length > 1
  return {
    animationDuration: 900,
    animationDurationUpdate: 500,
    animationEasing: 'cubicOut',
    animationDelay: (idx: number) => Math.min(idx * 25, 250),
    animationDelayUpdate: 120,
    grid: {
      left: 10,
      right: 14,
      top: showLegend ? 32 : 14,
      bottom: 4,
      containLabel: true,
    },
    tooltip: {
      trigger: 'axis',
      appendToBody: true,
      backgroundColor: 'rgba(9, 20, 39, 0.92)',
      borderColor: 'rgba(92, 210, 255, 0.18)',
      borderWidth: 1,
      padding: [10, 12],
      textStyle: { color: '#f8fbff', fontSize: 12 },
      extraCssText:
        'backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); box-shadow: 0 16px 42px rgba(0, 12, 34, 0.35); border-radius: 6px;',
      axisPointer: {
        type: 'line',
        lineStyle: { color: 'rgba(0, 110, 255, 0.45)', width: 1 },
      },
      valueFormatter: (value) => (isPercent ? `${Number(value ?? 0).toFixed(1)}%` : fmtBytes(Number(value ?? 0))),
    },
    legend: showLegend
      ? {
          data: props.series.map((s) => s.name),
          top: 0,
          right: 8,
          icon: 'roundRect',
          itemWidth: 11,
          itemHeight: 3,
          itemGap: 14,
          inactiveColor: t.split,
          textStyle: { color: t.label, fontSize: 11 },
        }
      : undefined,
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.categories,
      axisLine: { lineStyle: { color: t.line } },
      axisTick: { show: false },
      axisLabel: { color: t.label, fontSize: 11, hideOverlap: true, margin: 10 },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: isPercent ? 100 : undefined,
      axisLabel: {
        color: t.label,
        fontSize: 11,
        formatter: (value: number) => (isPercent ? `${value}%` : fmtBytes(value)),
      },
      splitLine: { lineStyle: { color: t.split, type: [4, 5] } },
    },
    series: props.series.map(seriesOption),
  }
})

watch([() => props.series, () => props.categories, () => props.mode, () => props.showLegend, () => theme.isDark], () => {
  if (mounted.value) chart.render(option.value)
}, { deep: true })

const mounted = ref(false)
onMounted(() => {
  mounted.value = true
  chart.render(option.value)
})

onBeforeUnmount(() => chart.dispose())
</script>

<template>
  <div class="cloud-line-chart">
    <div v-if="error" class="cloud-line-chart__error" :title="error">
      <span><DxIcon name="info" :size="14" /> 数据加载失败</span>
      <button type="button" @click="emit('retry')"><DxIcon name="refresh" :size="12" /> 重试</button>
    </div>
    <div class="cloud-line-chart__stage" :style="{ height: `${height}px` }">
      <div ref="chartEl" class="cloud-line-chart__canvas" />
      <div v-if="loading" class="cloud-line-chart__overlay">
        <span class="cloud-line-chart__pulse" />
        <span>数据更新中</span>
      </div>
      <div v-else-if="empty" class="cloud-line-chart__overlay">
        <DxIcon name="activity" :size="30" />
        <span>暂无监控数据</span>
        <small>运行实例后约 1 分钟自动生成曲线</small>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cloud-line-chart {
  position: relative;
}

.cloud-line-chart__stage {
  position: relative;
  overflow: hidden;
}

.cloud-line-chart__canvas {
  width: 100%;
  height: 100%;
}

.cloud-line-chart__overlay {
  position: absolute;
  inset: 0;
  z-index: 3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #86909c;
  background: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  pointer-events: none;
}

html.dark .cloud-line-chart__overlay {
  color: #64748b;
  background: rgba(13, 17, 23, 0.8);
}

.cloud-line-chart__overlay small {
  font-size: 11px;
  opacity: 0.8;
}

.cloud-line-chart__pulse {
  width: 28px;
  height: 28px;
  border-radius: 999px;
  border: 2px solid rgba(0, 110, 255, 0.16);
  border-top-color: #006eff;
  animation: cloud-line-spin 0.9s linear infinite;
}

.cloud-line-chart__error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  padding: 8px 12px;
  border: 1px solid rgba(245, 63, 63, 0.28);
  border-radius: 4px;
  background: rgba(245, 63, 63, 0.05);
  color: #f53f3f;
  font-size: 12px;
}

.cloud-line-chart__error span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.cloud-line-chart__error button {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 24px;
  padding: 0 9px;
  border: 1px solid rgba(245, 63, 63, 0.4);
  border-radius: 3px;
  background: transparent;
  color: #f53f3f;
  font-size: 12px;
  cursor: pointer;
}

@keyframes cloud-line-spin {
  to { transform: rotate(360deg); }
}
</style>
