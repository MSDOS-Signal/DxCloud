// 全局路由守卫：未登录跳登录页（真实 JWT 校验在后端，此处仅前端体验控制）
export default defineNuxtRouteMiddleware((to) => {
  const auth = useAuthStore()
  const publicPages = ['/login', '/register']
  if (!auth.isLoggedIn && !publicPages.includes(to.path)) {
    return navigateTo('/login')
  }
  if (auth.isLoggedIn && publicPages.includes(to.path)) {
    return navigateTo('/dashboard')
  }
})
