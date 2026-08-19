<script setup lang="ts">
import DxIcon from '~/components/DxIcon.vue'
import ParticleBg from '~/components/ParticleBg.vue'

withDefaults(defineProps<{
  icon: string
  title: string
  description?: string
  gradient?: string
}>(), {
  description: '',
  gradient: 'linear-gradient(120deg, #0052d9 0%, #006eff 45%, #00a1e4 100%)',
})
</script>

<template>
  <div class="page-hero dx-fade-up" :style="{ background: gradient }">
    <ParticleBg :count="42" :speed="0.32" :link-distance="110" :glow="0.9" class="hero-particles" />
    <div class="hero-shine" />
    <div class="hero-grid" />
    <div class="hero-content">
      <div class="hero-icon">
        <DxIcon :name="icon" :size="24" />
      </div>
      <div class="min-w-0">
        <h2 class="hero-title">{{ title }}</h2>
        <p v-if="description" class="hero-desc">{{ description }}</p>
      </div>
    </div>
    <div v-if="$slots.stats" class="hero-stats">
      <slot name="stats" />
    </div>
    <div v-if="$slots.action" class="hero-action">
      <slot name="action" />
    </div>
  </div>
</template>

<style scoped>
.page-hero {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  padding: 18px 22px;
  display: flex;
  align-items: center;
  gap: 18px;
  flex-wrap: wrap;
  color: #fff;
  box-shadow: 0 4px 20px rgba(0, 82, 217, 0.25);
  isolation: isolate;
}
.hero-particles {
  z-index: 1;
}
.hero-shine {
  position: absolute;
  top: -60%;
  right: -10%;
  width: 46%;
  height: 220%;
  background: radial-gradient(ellipse at center, rgba(255, 255, 255, 0.22) 0%, transparent 60%);
  transform: rotate(18deg);
  pointer-events: none;
  animation: hero-shimmer 7s ease-in-out infinite alternate;
  z-index: 0;
}
.hero-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 26px 26px;
  mask-image: radial-gradient(ellipse at right, black 0%, transparent 68%);
  -webkit-mask-image: radial-gradient(ellipse at right, black 0%, transparent 68%);
  pointer-events: none;
  z-index: 0;
}
@keyframes hero-shimmer {
  from { transform: rotate(18deg) translateX(-24px); opacity: 0.7; }
  to { transform: rotate(18deg) translateX(24px); opacity: 1; }
}
.hero-content {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  z-index: 2;
}
.hero-icon {
  width: 46px;
  height: 46px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.28);
  flex-shrink: 0;
  animation: hero-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
@keyframes hero-pop {
  from { transform: scale(0.6); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
.hero-title {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 0.01em;
  line-height: 1.3;
  text-shadow: 0 1px 3px rgba(0, 40, 100, 0.3);
}
.hero-desc {
  margin-top: 3px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.82);
  max-width: 560px;
  line-height: 1.5;
}
.hero-stats {
  position: relative;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-left: auto;
  z-index: 2;
}
.hero-action {
  position: relative;
  z-index: 2;
  margin-left: auto;
}
.hero-stats + .hero-action {
  margin-left: 0;
}
@media (max-width: 900px) {
  .hero-stats,
  .hero-action {
    margin-left: 0;
  }
}
</style>
