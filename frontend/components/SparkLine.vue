<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  values: number[]
  color?: string
  width?: number
  height?: number
}>(), {
  color: '#006eff',
  width: 120,
  height: 36,
})

const uid = Math.random().toString(36).slice(2, 8)
const gradId = `sg-${uid}`

const points = computed(() => {
  const vals = props.values.length ? props.values : [0, 0]
  const max = Math.max(...vals)
  const min = Math.min(...vals)
  const range = max - min || 1
  const stepX = props.width / Math.max(vals.length - 1, 1)
  return vals.map((v, i) => ({
    x: i * stepX,
    y: props.height - 3 - ((v - min) / range) * (props.height - 8),
  }))
})

const last = computed(() => points.value[points.value.length - 1])

const linePath = computed(() => {
  const pts = points.value
  if (pts.length < 3) {
    return pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' ')
  }
  let d = `M${pts[0].x.toFixed(1)},${pts[0].y.toFixed(1)}`
  for (let i = 1; i < pts.length - 1; i++) {
    const xc = ((pts[i].x + pts[i + 1].x) / 2).toFixed(1)
    const yc = ((pts[i].y + pts[i + 1].y) / 2).toFixed(1)
    d += ` Q${pts[i].x.toFixed(1)},${pts[i].y.toFixed(1)} ${xc},${yc}`
  }
  const tail = pts[pts.length - 1]
  d += ` L${tail.x.toFixed(1)},${tail.y.toFixed(1)}`
  return d
})

const areaPath = computed(() =>
  `${linePath.value} L${props.width},${props.height} L0,${props.height} Z`,
)
</script>

<template>
  <svg :width="width" :height="height" class="spark" preserveAspectRatio="none">
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.25" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path :d="areaPath" :fill="`url(#${gradId})`" class="spark-area" />
    <path :d="linePath" fill="none" :stroke="color" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="spark-line" />
    <circle v-if="last" :cx="last.x" :cy="last.y" r="3" :fill="color" class="spark-dot" />
    <circle v-if="last" :cx="last.x" :cy="last.y" r="3" fill="none" :stroke="color" stroke-width="1.5" class="spark-dot-pulse" />
  </svg>
</template>

<style scoped>
.spark {
  display: block;
  overflow: visible;
}
.spark-area {
  animation: spark-in 0.8s ease both;
}
.spark-line {
  stroke-dasharray: 600;
  stroke-dashoffset: 600;
  animation: spark-draw 1.1s ease-out 0.15s forwards;
}
.spark-dot {
  animation: spark-in 0.4s ease 0.9s both;
}
.spark-dot-pulse {
  animation: spark-pulse 2s ease-out 1.2s infinite;
  transform-origin: center;
  transform-box: fill-box;
}
@keyframes spark-draw {
  to { stroke-dashoffset: 0; }
}
@keyframes spark-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes spark-pulse {
  0% { opacity: 0.8; transform: scale(1); }
  70% { opacity: 0; transform: scale(2.8); }
  100% { opacity: 0; transform: scale(2.8); }
}
</style>
