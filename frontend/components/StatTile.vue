<script setup lang="ts">
import { computed } from 'vue'
import DxIcon from '~/components/DxIcon.vue'
import CountUp from '~/components/CountUp.vue'

const props = withDefaults(defineProps<{
  icon: string
  label: string
  value: number | string
  decimals?: number
  suffix?: string
  hint?: string
  color?: string
  trend?: number | null
}>(), {
  decimals: 0,
  suffix: '',
  hint: '',
  color: '#006eff',
  trend: null,
})

const isDash = computed(() => props.value === '—' || props.value === null || props.value === undefined)
</script>

<template>
  <div class="stat-tile dx-fade-up">
    <div class="stat-icon" :style="{ background: color + '14', color }">
      <DxIcon :name="icon" :size="17" />
    </div>
    <div class="min-w-0 flex-1">
      <div class="stat-label">{{ label }}</div>
      <div class="stat-value" :style="{ color }">
        <span v-if="isDash">—</span>
        <CountUp v-else :value="value" :decimals="decimals" :suffix="suffix" />
        <span v-if="trend !== null && !isDash" class="stat-trend" :class="trend >= 0 ? 'up' : 'down'">
          {{ trend >= 0 ? '↑' : '↓' }} {{ Math.abs(trend).toFixed(1) }}%
        </span>
      </div>
      <div v-if="hint" class="stat-hint">{{ hint }}</div>
    </div>
  </div>
</template>

<style scoped>
.stat-tile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  position: relative;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.stat-tile::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: currentColor;
  opacity: 0;
  transition: opacity 0.2s ease;
}
.stat-tile:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(0, 20, 60, 0.08);
  border-color: #d0e3ff;
}
html.dark .stat-tile {
  background: #161b22;
  border-color: #30363d;
}
html.dark .stat-tile:hover {
  border-color: #2b4a75;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.4);
}
.stat-icon {
  width: 38px;
  height: 38px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-label {
  font-size: 12px;
  color: #86909c;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
html.dark .stat-label {
  color: #8b949e;
}
.stat-value {
  font-size: 22px;
  font-weight: 700;
  line-height: 1.25;
  letter-spacing: -0.01em;
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-variant-numeric: tabular-nums;
}
.stat-trend {
  font-size: 11px;
  font-weight: 500;
}
.stat-trend.up { color: #00b42a; }
.stat-trend.down { color: #f53f3f; }
.stat-hint {
  font-size: 11px;
  color: #c9cdd4;
  margin-top: 1px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
html.dark .stat-hint {
  color: #6e7681;
}
</style>
