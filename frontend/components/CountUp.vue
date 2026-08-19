<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  value: number | string
  duration?: number
  decimals?: number
  suffix?: string
}>(), {
  duration: 900,
  decimals: 0,
  suffix: '',
})

const display = ref(0)
let raf = 0

function easeOutExpo(t: number): number {
  return t === 1 ? 1 : 1 - Math.pow(2, -10 * t)
}

function animate(from: number, to: number) {
  if (raf) cancelAnimationFrame(raf)
  const start = performance.now()
  const diff = to - from

  function tick(now: number) {
    const elapsed = now - start
    const progress = Math.min(elapsed / props.duration, 1)
    display.value = from + diff * easeOutExpo(progress)
    if (progress < 1) {
      raf = requestAnimationFrame(tick)
    } else {
      display.value = to
    }
  }
  raf = requestAnimationFrame(tick)
}

const numericValue = computed(() => {
  if (typeof props.value === 'string') {
    const parsed = parseFloat(props.value)
    return isNaN(parsed) ? 0 : parsed
  }
  return props.value
})

watch(
  () => props.value,
  (newVal) => {
    const target = typeof newVal === 'string' ? parseFloat(newVal) : newVal
    if (!isNaN(target)) {
      animate(display.value, target)
    }
  },
)

onMounted(() => {
  if (numericValue.value > 0) {
    animate(0, numericValue.value)
  }
})

onBeforeUnmount(() => {
  if (raf) cancelAnimationFrame(raf)
})

const formatted = computed(() => {
  if (typeof props.value === 'string' && props.value === '—') return '—'
  return display.value.toFixed(props.decimals) + props.suffix
})
</script>

<template>
  <span class="tabular-nums">{{ formatted }}</span>
</template>
