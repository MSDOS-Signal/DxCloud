<script setup lang="ts">
const props = withDefaults(defineProps<{
  size?: number
  compact?: boolean
  subtitle?: string
}>(), {
  size: 36,
  compact: false,
  subtitle: '云平台控制台',
})
</script>

<template>
  <div class="dx-brand" :class="{ 'dx-brand--compact': compact }">
    <div
      class="dx-brand-mark"
      :style="{ width: `${size}px`, height: `${size}px` }"
      aria-hidden="true"
    >
      <svg viewBox="0 0 64 64" class="dx-brand-svg">
        <defs>
          <linearGradient id="dx-brand-main" x1="8" y1="8" x2="56" y2="56" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stop-color="#64d9ff" />
            <stop offset="45%" stop-color="#1f7bff" />
            <stop offset="100%" stop-color="#0052d9" />
          </linearGradient>
          <linearGradient id="dx-brand-accent" x1="18" y1="14" x2="48" y2="48" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stop-color="rgba(255,255,255,0.95)" />
            <stop offset="100%" stop-color="rgba(255,255,255,0.35)" />
          </linearGradient>
          <filter id="dx-brand-glow" x="-60%" y="-60%" width="220%" height="220%">
            <feGaussianBlur stdDeviation="5" result="blur" />
            <feColorMatrix
              in="blur"
              type="matrix"
              values="1 0 0 0 0
                      0 1 0 0 0.25
                      0 0 1 0 0.7
                      0 0 0 0.32 0"
            />
          </filter>
        </defs>
        <circle cx="32" cy="32" r="23" fill="url(#dx-brand-main)" filter="url(#dx-brand-glow)" opacity="0.75" />
        <rect x="10" y="10" width="44" height="44" rx="14" fill="url(#dx-brand-main)" />
        <path
          d="M18 35.5c0-7.46 6.04-13.5 13.5-13.5 3.77 0 7.18 1.54 9.63 4.02A11.46 11.46 0 0 1 49 37.5c0 6.35-5.15 11.5-11.5 11.5H28A10 10 0 0 1 18 39z"
          fill="url(#dx-brand-accent)"
          opacity="0.9"
        />
        <path
          d="M22 36.5c1.7-5.95 7.2-10.3 13.66-10.3 2.38 0 4.63.59 6.6 1.63"
          stroke="rgba(255,255,255,0.86)"
          stroke-width="3.2"
          stroke-linecap="round"
        />
        <path
          d="M22 42h20.5c3.04 0 5.5-2.46 5.5-5.5 0-2.52-1.7-4.64-4.01-5.29"
          stroke="rgba(255,255,255,0.96)"
          stroke-width="4"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <circle cx="46.5" cy="23.5" r="3.2" fill="#9be7ff" opacity="0.9" />
      </svg>
      <span class="dx-brand-halo dx-brand-halo--one" />
      <span class="dx-brand-halo dx-brand-halo--two" />
    </div>
    <div v-if="!compact" class="dx-brand-copy">
      <div class="dx-brand-title">多晓云</div>
      <div class="dx-brand-subtitle">DxCloud · {{ subtitle }}</div>
    </div>
  </div>
</template>

<style scoped>
.dx-brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.dx-brand--compact {
  gap: 0;
}

.dx-brand-mark {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.dx-brand-svg {
  position: relative;
  z-index: 2;
  width: 100%;
  height: 100%;
  filter: drop-shadow(0 10px 18px rgba(0, 82, 217, 0.2));
}

.dx-brand-copy {
  min-width: 0;
}

.dx-brand-title {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: 0.02em;
  color: #111827;
}

.dx-brand-subtitle {
  margin-top: 3px;
  font-size: 11px;
  line-height: 1.2;
  color: #7b8794;
}

html.dark .dx-brand-title {
  color: #f8fafc;
}

html.dark .dx-brand-subtitle {
  color: #7f8ea3;
}

.dx-brand-halo {
  position: absolute;
  inset: 50% auto auto 50%;
  border-radius: 999px;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 1;
}

.dx-brand-halo--one {
  width: 78%;
  height: 78%;
  background: radial-gradient(circle, rgba(0, 198, 255, 0.28), rgba(0, 110, 255, 0));
  animation: dx-brand-pulse 4.8s ease-in-out infinite;
}

.dx-brand-halo--two {
  width: 112%;
  height: 112%;
  border: 1px solid rgba(100, 217, 255, 0.18);
  animation: dx-brand-ring 5.6s ease-in-out infinite;
}

@keyframes dx-brand-pulse {
  0%, 100% { opacity: 0.45; transform: translate(-50%, -50%) scale(0.92); }
  50% { opacity: 0.9; transform: translate(-50%, -50%) scale(1.06); }
}

@keyframes dx-brand-ring {
  0%, 100% { opacity: 0.2; transform: translate(-50%, -50%) scale(0.98); }
  50% { opacity: 0.55; transform: translate(-50%, -50%) scale(1.05); }
}
</style>
