<script setup lang="ts">
useCursor()
</script>

<template>
  <div class="auth-layout min-h-screen relative flex items-center justify-center px-4 py-10 bg-[#f0f2f5] dark:bg-[#0d1117] overflow-hidden">
    <div class="absolute inset-0 opacity-70 dark:opacity-45">
      <ParticleBg :count="46" :speed="0.32" :glow="0.7" />
    </div>
    <div class="w-full max-w-[400px]">
      <!-- 品牌区 -->
      <div class="mb-6 flex flex-col items-center dx-fade-in">
        <BrandLogo :size="46" subtitle="云平台控制台" />
      </div>

      <div class="dx-fade-in dx-delay-1">
        <slot />
      </div>

      <div class="mt-6 text-center text-xs text-[#86909c] dx-fade-in dx-delay-2">
        Docker 驱动的轻量级 ECS / IaaS / PaaS / SaaS / CI-CD 一体化云平台
      </div>
    </div>
  </div>
</template>
