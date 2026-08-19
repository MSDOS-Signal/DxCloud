import { defineNuxtModule } from '@nuxt/kit'

// Nuxt 3.16 dev 模式 HMR 竞态会把内置 validate 全局中间件注册两次，
// 生成的 middleware.mjs 出现重复 import 导致浏览器 SyntaxError。
// 上游未修复：https://github.com/nuxt/nuxt/issues/33249
// 此模块在模板渲染前按 name 去重，保证生成的 middleware.mjs 永远干净。
export default defineNuxtModule({
  meta: { name: 'fix-middleware-dup' },
  setup(_options, nuxt) {
    nuxt.hook('app:templates', (app) => {
      const seen = new Set<string>()
      app.middleware = app.middleware.filter((mw) => {
        if (seen.has(mw.name)) return false
        seen.add(mw.name)
        return true
      })
    })
  },
})
