<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { EChartsOption } from '~/composables/useEcharts'
import { useEcharts } from '~/composables/useEcharts'

interface DonutItem {
  name: string
  value: number
  color: string
}

const props = withDefaults(defineProps<{
  items: DonutItem[]
  height?: number
  centerText?: string
  centerLabel?: string
}>(), {
  height: 220,
  centerText: '',
  centerLabel: '实例总数',
})

const theme = useThemeStore()
const chartEl = ref<HTMLElement | null>(null)
const chart = useEcharts(chartEl)

const visibleItems = computed(() => props.items.filter((i) => i.value > 0))
const total = computed(() => visibleItems.value.reduce((sum, i) => sum + i.value, 0))

const option = computed<EChartsOption>(() => {
  const dark = theme.isDark
  const hasData = visibleItems.value.length > 0
  return {
    animationDuration: 1000,
    animationEasing: 'cubicOut',
    animationDelay: (idx: number) => idx * 110,
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(9,20,39,.92)',
      borderColor: 'rgba(92,210,255,.18)',
      textStyle: { color: '#f8fbff', fontSize: 12 },
      extraCssText: 'backdrop-filter: blur(12px); border-radius: 6px; box-shadow: 0 12px 30px rgba(0,12,34,.3);',
      formatter: hasData ? '{b}<br/><b>{c}</b> 台（{d}%）' : () => '暂无实例',
    },
    legend: {
      bottom: 0,
      left: 'center',
      icon: 'circle',
      itemWidth: 8,
      itemHeight: 8,
      itemGap: 18,
      textStyle: { color: dark ? '#8b949e' : '#4e5969', fontSize: 11 },
    },
    series: [
      {
        type: 'pie',
        radius: ['64%', '82%'],
        center: ['50%', '42%'],
        avoidLabelOverlap: true,
        padAngle: 2,
        cursor: hasData ? 'pointer' : 'default',
        itemStyle: {
          borderColor: dark ? '#161b22' : '#ffffff',
          borderWidth: 3,
          borderRadius: 5,
          shadowBlur: 10,
          shadowColor: 'rgba(0,110,255,.12)',
        },
        label: { show: false },
        emphasis: {
          scaleSize: 7,
          itemStyle: { shadowBlur: 18, shadowColor: 'rgba(0,110,255,.3)' },
        },
        data: hasData
          ? visibleItems.value.map((i) => ({ name: i.name, value: i.value, itemStyle: { color: i.color } }))
          : [
              {
                name: '暂无实例',
                value: 1,
                itemStyle: { color: dark ? '#21262d' : '#ebedf0' },
                tooltip: { show: false },
                emphasis: { disabled: true },
              },
            ],
      },
    ],
  }
})

const mounted = ref(false)
watch([() => props.items, () => props.centerText, () => props.centerLabel, () => theme.isDark], () => {
  if (mounted.value) chart.render(option.value)
}, { deep: true })

onMounted(() => {
  mounted.value = true
  chart.render(option.value)
})
onBeforeUnmount(() => chart.dispose())
</script>

<template>
  <div class="cloud-donut-chart">
    <div class="cloud-donut-chart__stage" :style="{ height: `${height}px` }">
      <div ref="chartEl" class="cloud-donut-chart__canvas" />
      <div class="cloud-donut-chart__center">
        <span class="cloud-donut-chart__num">{{ centerText || total }}</span>
        <span class="cloud-donut-chart__lbl">{{ centerLabel }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cloud-donut-chart__stage {
  position: relative;
}

.cloud-donut-chart__canvas {
  width: 100%;
  height: 100%;
}

.cloud-donut-chart__center {
  position: absolute;
  left: 0;
  right: 0;
  top: 42%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  pointer-events: none;
}

.cloud-donut-chart__num {
  font-size: 22px;
  font-weight: 600;
  line-height: 1.1;
  color: #1f2329;
  font-variant-numeric: tabular-nums;
}

.cloud-donut-chart__lbl {
  font-size: 11px;
  color: #86909c;
}

html.dark .cloud-donut-chart__num {
  color: #e6edf3;
}

html.dark .cloud-donut-chart__lbl {
  color: #64748b;
}
</style>
