<template>
  <div class="h-screen flex flex-col bg-slate-900">
    <div class="h-12 px-4 flex items-center justify-between bg-slate-800 text-slate-200 border-b border-slate-700">
      <div class="flex items-center gap-3">
        <n-button size="tiny" quaternary @click="navigateTo('/ecs')">← 返回 ECS</n-button>
        <span class="text-sm font-semibold">Web Terminal</span>
        <span v-if="route.params.id" class="text-xs text-slate-400">实例 #{{ route.params.id }}</span>
      </div>
      <span class="text-xs text-slate-400">空闲 15 分钟自动断开</span>
    </div>
    <div class="flex-1 min-h-0">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
</script>
