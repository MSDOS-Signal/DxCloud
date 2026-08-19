import naive from 'naive-ui'

// 全局注册 Naive UI 组件（Phase 1 简单可靠；后续可换按需引入优化体积）
export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(naive)
})
