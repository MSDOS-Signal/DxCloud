<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  color?: string
  palette?: string[]
  linkDistance?: number
  count?: number
  speed?: number
  glow?: number
  interactive?: boolean
}>(), {
  color: '',
  palette: () => ['0, 110, 255', '0, 164, 255', '0, 198, 255', '0, 180, 42'],
  linkDistance: 130,
  count: 0,
  speed: 0.4,
  glow: 0.8,
  interactive: true,
})

const canvas = ref<HTMLCanvasElement | null>(null)
let ctx: CanvasRenderingContext2D | null = null
let raf = 0
let particles: { x: number; y: number; vx: number; vy: number; r: number; color: string; phase: number }[] = []
let mouse = { x: -9999, y: -9999 }
let w = 0
let h = 0
let dpr = 1

function resolveColors(): string[] {
  if (props.color) return [props.color]
  return props.palette
}

function resize() {
  const el = canvas.value
  if (!el || !ctx) return
  const parent = el.parentElement
  if (!parent) return
  dpr = Math.min(window.devicePixelRatio || 1, 2)
  w = parent.clientWidth
  h = parent.clientHeight
  el.width = w * dpr
  el.height = h * dpr
  el.style.width = w + 'px'
  el.style.height = h + 'px'
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  initParticles()
}

function initParticles() {
  const target = props.count || Math.min(104, Math.max(34, Math.floor((w * h) / 12000)))
  const colors = resolveColors()
  particles = []
  for (let i = 0; i < target; i++) {
    particles.push({
      x: Math.random() * w,
      y: Math.random() * h,
      vx: (Math.random() - 0.5) * props.speed,
      vy: (Math.random() - 0.5) * props.speed,
      r: Math.random() * 1.6 + 0.8,
      color: colors[Math.floor(Math.random() * colors.length)] || colors[0] || '0, 110, 255',
      phase: Math.random() * Math.PI * 2,
    })
  }
}

function draw() {
  if (!ctx) return
  ctx.clearRect(0, 0, w, h)

  for (const p of particles) {
    p.x += p.vx
    p.y += p.vy
    if (p.x < 0 || p.x > w) p.vx *= -1
    if (p.y < 0 || p.y > h) p.vy *= -1

    if (props.interactive) {
      const dx = p.x - mouse.x
      const dy = p.y - mouse.y
      const dist = Math.hypot(dx, dy)
      if (dist < 130 && dist > 0.1) {
        const f = (130 - dist) / 130
        p.x += (dx / dist) * f * 1.4
        p.y += (dy / dist) * f * 1.4
      }
    }

    ctx.beginPath()
    ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
    ctx.fillStyle = `rgba(${p.color}, ${0.48 + Math.sin(p.phase) * 0.16})`
    if (props.glow > 0) {
      ctx.shadowBlur = props.glow * 6
      ctx.shadowColor = `rgba(${p.color}, ${props.glow})`
    }
    ctx.fill()
    ctx.shadowBlur = 0
  }

  for (let i = 0; i < particles.length; i++) {
    for (let j = i + 1; j < particles.length; j++) {
      const a = particles[i]
      const b = particles[j]
      const d = Math.hypot(a.x - b.x, a.y - b.y)
      if (d < props.linkDistance) {
        const op = (1 - d / props.linkDistance) * 0.38
        ctx.beginPath()
        ctx.moveTo(a.x, a.y)
        ctx.lineTo(b.x, b.y)
        const gradient = ctx.createLinearGradient(a.x, a.y, b.x, b.y)
        gradient.addColorStop(0, `rgba(${a.color}, ${op})`)
        gradient.addColorStop(1, `rgba(${b.color}, ${op * 0.55})`)
        ctx.strokeStyle = gradient
        ctx.lineWidth = 0.8
        ctx.stroke()
      }
    }
  }
  raf = requestAnimationFrame(draw)
}

function onMouseMove(e: MouseEvent) {
  const el = canvas.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  mouse.x = e.clientX - rect.left
  mouse.y = e.clientY - rect.top
}

function onMouseLeave() {
  mouse.x = -9999
  mouse.y = -9999
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  const el = canvas.value
  if (!el) return
  ctx = el.getContext('2d')
  if (!ctx) return
  resize()
  resizeObserver = new ResizeObserver(resize)
  if (el.parentElement) resizeObserver.observe(el.parentElement)
  if (props.interactive) {
    window.addEventListener('mousemove', onMouseMove, { passive: true })
    el.addEventListener('mouseleave', onMouseLeave)
  }
  raf = requestAnimationFrame(draw)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(raf)
  resizeObserver?.disconnect()
  window.removeEventListener('mousemove', onMouseMove)
  canvas.value?.removeEventListener('mouseleave', onMouseLeave)
})
</script>

<template>
  <canvas ref="canvas" class="absolute inset-0 w-full h-full pointer-events-none" />
</template>
