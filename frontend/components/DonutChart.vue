<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

interface Segment {
  value: number
  color: string
  label?: string
}

const props = withDefaults(defineProps<{
  segments: Segment[]
  size?: number
  thickness?: number
  centerText?: string
  centerLabel?: string
  valueDecimals?: number
}>(), {
  size: 120,
  thickness: 12,
  centerText: '',
  centerLabel: '',
  valueDecimals: -1,
})

const mounted = ref(false)
onMounted(() => {
  requestAnimationFrame(() => { mounted.value = true })
})

const filtered = computed(() => props.segments.filter((s) => s.value > 0))
const total = computed(() => filtered.value.reduce((s, x) => s + x.value, 0))
const radius = computed(() => (props.size - props.thickness) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)

const GAP_DEG = 4

const arcs = computed(() => {
  if (total.value <= 0) return []
  const gapFrac = filtered.value.length > 1 ? GAP_DEG / 360 : 0
  let acc = 0
  return filtered.value.map((s) => {
    const frac = s.value / total.value
    const dashFrac = Math.max(frac - gapFrac, 0.02)
    const arc = {
      color: s.color,
      label: s.label || '',
      value: s.value,
      dash: dashFrac * circumference.value,
      offset: -(acc + gapFrac / 2) * circumference.value,
      pct: Math.round(frac * 100),
      round: filtered.value.length === 1,
    }
    acc += frac
    return arc
  })
})

function fmtVal(v: number): string {
  return props.valueDecimals >= 0 ? v.toFixed(props.valueDecimals) : String(v)
}
</script>

<template>
  <div class="donut">
    <svg :width="size" :height="size" class="donut-svg">
      <circle
        :cx="size / 2" :cy="size / 2" :r="radius"
        fill="none" stroke="#eef1f5" :stroke-width="thickness"
      />
      <circle
        v-for="(a, i) in arcs" :key="i"
        :cx="size / 2" :cy="size / 2" :r="radius"
        fill="none" :stroke="a.color" :stroke-width="thickness"
        :stroke-dasharray="`${mounted ? a.dash : 0} ${circumference}`"
        :stroke-dashoffset="a.offset"
        :stroke-linecap="a.round ? 'round' : 'butt'"
        class="donut-arc"
        :style="{ transition: `stroke-dasharray 0.9s cubic-bezier(0.25, 1, 0.4, 1) ${i * 0.12}s` }"
      >
        <title v-if="a.label">{{ a.label }} · {{ fmtVal(a.value) }}（{{ a.pct }}%）</title>
      </circle>
      <text
        v-if="centerText" :x="size / 2" :y="centerLabel ? size / 2 - 2 : size / 2 + 6"
        text-anchor="middle" class="donut-center"
        :style="{ fontSize: Math.max(13, size / 7) + 'px' }"
      >{{ centerText }}</text>
      <text
        v-if="centerLabel" :x="size / 2" :y="size / 2 + 15"
        text-anchor="middle" class="donut-sub"
        :style="{ fontSize: Math.max(9, size / 12) + 'px' }"
      >{{ centerLabel }}</text>
    </svg>
    <div v-if="arcs.some(a => a.label)" class="donut-legend">
      <div v-for="(a, i) in arcs" :key="i" class="legend-row" :style="{ animationDelay: `${0.3 + i * 0.08}s` }">
        <span class="dot" :style="{ background: a.color }" />
        <span class="name">{{ a.label }}</span>
        <span class="pct">{{ a.pct }}%</span>
        <span class="val">{{ fmtVal(a.value) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.donut {
  width: 100%;
  max-width: 300px;
  margin: 0 auto;
}
.donut-svg {
  display: block;
  margin: 0 auto;
  transform: rotate(-90deg);
}
.donut-arc {
  cursor: default;
  transition-property: stroke-dasharray, filter, stroke-width;
  transition-duration: 0.9s, 0.2s, 0.2s;
}
.donut-arc:hover {
  filter: brightness(1.12) drop-shadow(0 1px 4px rgba(0, 0, 0, 0.18));
}
.donut-center {
  fill: #1f2329;
  font-weight: 700;
  transform: rotate(90deg);
  transform-origin: center;
}
.donut-sub {
  fill: #86909c;
  transform: rotate(90deg);
  transform-origin: center;
}
html.dark .donut-center { fill: #e6edf3; }
html.dark .donut-sub { fill: #8b949e; }
html.dark .donut-svg circle:first-child { stroke: #21262d; }
.donut-legend {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 10px;
  min-width: 0;
}
.legend-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #4e5969;
  padding: 2px 0;
  animation: legend-in 0.4s ease both;
}
html.dark .legend-row { color: #c9d1d9; }
@keyframes legend-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 3px;
  flex-shrink: 0;
}
.name {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pct {
  color: #86909c;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}
html.dark .pct { color: #8b949e; }
.val {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: #1f2329;
}
html.dark .val { color: #e6edf3; }
</style>
