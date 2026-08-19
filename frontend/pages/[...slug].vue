<script setup lang="ts">
import { computed } from 'vue'
import DxIcon from '~/components/DxIcon.vue'

const route = useRoute()
const router = useRouter()

const quickNav = [
  { label: 'Dashboard', desc: '总览面板', icon: 'dashboard', to: '/dashboard' },
  { label: 'ECS 云主机', desc: '计算实例', icon: 'ecs', to: '/ecs' },
  { label: '容器实例', desc: 'Docker 管理', icon: 'containers', to: '/containers' },
  { label: 'CI/CD', desc: '流水线', icon: 'pipelines', to: '/pipelines' },
]

const pathDisplay = computed(() => route.path)
</script>

<template>
  <div class="relative min-h-[calc(100vh-4rem)] flex items-center justify-center overflow-hidden px-4 py-12" style="background: #f0f2f5;">
    <div class="relative w-full max-w-lg text-center">
      <!-- 云图标 -->
      <div class="relative inline-flex items-center justify-center mb-8">
        <div class="relative w-20 h-20 rounded flex items-center justify-center shadow-[0_12px_32px_-8px_rgba(0,110,255,0.6)]" style="background: #006eff;">
          <DxIcon name="cloud" :size="36" :stroke="1.5" class="text-white" />
        </div>
      </div>

      <!-- 404 故障文字 -->
      <div class="dx-fade-up dx-delay-1">
        <h1
          class="glitch text-[120px] leading-none font-black tracking-tight select-none"
          data-text="404"
        >
          404
        </h1>
      </div>

      <!-- 描述 -->
      <div class="dx-fade-up dx-delay-2 mt-4">
        <div class="text-xl font-bold text-gray-800">页面走丢了</div>
        <div class="mt-2 text-sm text-gray-500">
          未找到路径 <code class="px-1.5 py-0.5 rounded-md bg-gray-100 text-[#006eff] font-mono text-xs">{{ pathDisplay }}</code>
        </div>
        <div class="mt-1 text-xs text-gray-400">请检查地址或从下方快捷入口进入功能模块</div>
      </div>

      <!-- 快捷导航 -->
      <div class="dx-fade-up dx-delay-3 mt-8 grid grid-cols-2 sm:grid-cols-4 gap-3">
        <button
          v-for="item in quickNav"
          :key="item.to"
          class="group dx-card rounded p-3 border border-gray-200 hover:border-[#006eff]/50 transition-all hover:-translate-y-1"
          @click="router.push(item.to)"
        >
          <div class="w-9 h-9 mx-auto rounded-lg flex items-center justify-center shadow-md group-hover:scale-110 transition-transform" style="background: #006eff;">
            <DxIcon :name="item.icon" :size="17" class="text-white" />
          </div>
          <div class="mt-2 text-xs font-semibold text-gray-800">{{ item.label }}</div>
          <div class="text-[10px] text-gray-500">{{ item.desc }}</div>
        </button>
      </div>

      <!-- 返回按钮 -->
      <div class="dx-fade-up dx-delay-4 mt-8">
        <button
          class="dx-btn-primary inline-flex items-center gap-2"
          @click="router.push('/dashboard')"
        >
          <DxIcon name="arrow-left" :size="16" />
          返回 Dashboard
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 404 故障文字动画效果 */
.glitch {
  position: relative;
  color: #1f2329;
}
.glitch::before,
.glitch::after {
  content: attr(data-text);
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
.glitch::before {
  color: #006eff;
  animation: glitch-anim-1 2.5s infinite linear alternate-reverse;
  clip-path: polygon(0 0, 100% 0, 100% 45%, 0 45%);
}
.glitch::after {
  color: #ff7d00;
  animation: glitch-anim-2 2s infinite linear alternate-reverse;
  clip-path: polygon(0 55%, 100% 55%, 100% 100%, 0 100%);
}
@keyframes glitch-anim-1 {
  0% { transform: translate(0); }
  20% { transform: translate(-3px, 1px); }
  40% { transform: translate(3px, -1px); }
  60% { transform: translate(-2px, 2px); }
  80% { transform: translate(2px, -2px); }
  100% { transform: translate(0); }
}
@keyframes glitch-anim-2 {
  0% { transform: translate(0); }
  20% { transform: translate(3px, -1px); }
  40% { transform: translate(-3px, 1px); }
  60% { transform: translate(2px, 2px); }
  80% { transform: translate(-2px, -2px); }
  100% { transform: translate(0); }
}
</style>
