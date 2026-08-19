export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',

  // 控制台类应用采用 SPA：无 SSR 水合问题，WebSocket / 状态管理更简单
  ssr: false,
  devtools: { enabled: false },

  modules: ['@pinia/nuxt', '@nuxtjs/tailwindcss', '~/modules/fix-middleware-dup'],

  css: ['~/assets/css/main.css'],

  tailwindcss: {
    config: {
      darkMode: 'class', // 深色模式由 <html class="dark"> 控制（stores/theme.ts）
    },
  },

  typescript: {
    strict: true,
    typeCheck: false,
  },

  // 大依赖（naive-ui / echarts / vueuc）预构建，避免 dev 模式超大 chunk 加载中断（ERR_ABORTED）
  vite: {
    optimizeDeps: {
      include: [
        'naive-ui',
        'vueuc',
        'date-fns-tz',
        'echarts/core',
        'echarts/charts',
        'echarts/components',
        'echarts/renderers',
      ],
    },
  },

  app: {
    head: {
      title: '多晓云 DxCloud Console',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '多晓云 DxCloud · 做懂你心的云枢 — 基于 Docker 的一体化云平台控制台' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
      ],
    },
  },

  runtimeConfig: {
    public: {
      apiBase: '/api/v1',
      apiProxyTarget: process.env.NUXT_API_PROXY_TARGET || 'http://127.0.0.1:8080',
    },
  },

  // 开发期把 /api/v1、/ws 反向代理到 Go 后端（仅 dev 生效；生产由 Traefik 路由）
  nitro: {
    devProxy: {
      '/api/v1': {
        target: process.env.NUXT_API_PROXY_TARGET || 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: process.env.NUXT_API_PROXY_TARGET || 'http://127.0.0.1:8080',
        changeOrigin: true,
        ws: true,
      },
    },
  },

  devServer: {
    host: '0.0.0.0',
    port: 3000,
  },
})
