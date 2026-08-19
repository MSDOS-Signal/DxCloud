import { ofetch } from 'ofetch'
import type { ApiResponse } from '~/types'
import { clearTokens, getAccessToken, getRefreshToken, setTokens } from '~/utils/token'

// ApiError：把后端统一错误结构 {code,message,request_id} 转成可捕获异常
export class ApiError extends Error {
  code: number
  requestId: string

  constructor(code: number, message: string, requestId: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.requestId = requestId
  }
}

// 底层请求：自动附加 Bearer 与租户上下文 X-Org-Id；HTTP 层错误统一抛 ApiError
const raw = ofetch.create({
  baseURL: '/api/v1',
  onRequest({ options }) {
    const headers: Record<string, string> = { ...((options.headers as Record<string, string> | undefined) || {}) }
    const token = getAccessToken()
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    // 多租户上下文（Phase 10）：当前组织 0 表示默认空间，不发头
    if (import.meta.client && typeof localStorage !== 'undefined') {
      const orgId = localStorage.getItem('dx-org-id')
      if (orgId && orgId !== '0' && !headers['X-Org-Id']) {
        headers['X-Org-Id'] = orgId
      }
    }
    options.headers = headers
  },
  onResponseError({ response }) {
    const body = response._data as Partial<ApiResponse> | null
    const code = body?.code ?? response.status
    const message = body?.message ?? `HTTP ${response.status}`
    throw new ApiError(code, message, body?.request_id ?? '')
  },
})

// Refresh 并发锁：多个 401 同时触发时只刷新一次
let refreshing: Promise<boolean> | null = null

// 会话彻底失效（refresh 也失败）时跳回登录页，避免卡在需要登录的页面
function redirectToLogin() {
  if (!import.meta.client) return
  const path = window.location.pathname
  if (path === '/login' || path === '/register') return
  localStorage.removeItem('dx-org-id')
  window.location.replace('/login')
}

function tryRefresh(): Promise<boolean> {
  if (!refreshing) {
    refreshing = (async () => {
      const rt = getRefreshToken()
      if (!rt) {
        clearTokens()
        return false
      }
      try {
        const res = await raw<ApiResponse<{ access_token: string; refresh_token: string }>>('/auth/refresh', {
          method: 'POST',
          body: { refresh_token: rt },
        })
        setTokens(res.data.access_token, res.data.refresh_token)
        return true
      } catch {
        clearTokens()
        return false
      }
    })().finally(() => {
      refreshing = null
    })
  }
  return refreshing
}

// 业务请求封装：
// 1) 解包统一响应（code!=0 抛 ApiError，code==0 返回 data）
// 2) 401 时自动刷新令牌并重试一次
async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
  const doCall = async (): Promise<T> => {
    const res = await raw<ApiResponse<T>>(url, { method, body } as never)
    if (res && typeof res === 'object' && 'code' in res) {
      if ((res as ApiResponse).code !== 0) {
        throw new ApiError((res as ApiResponse).code, (res as ApiResponse).message, (res as ApiResponse).request_id)
      }
      return (res as ApiResponse<T>).data
    }
    return res as unknown as T
  }
  try {
    return await doCall()
  } catch (e) {
    if (
      e instanceof ApiError &&
      e.code === 40100 &&
      !url.startsWith('/auth/login') &&
      !url.startsWith('/auth/register') &&
      !url.startsWith('/auth/refresh')
    ) {
      const ok = await tryRefresh()
      if (ok) {
        return await doCall()
      }
      // 刷新失败 = 会话彻底失效，跳回登录页
      redirectToLogin()
    }
    throw e
  }
}

export const api = {
  get: <T>(url: string) => request<T>('GET', url),
  post: <T>(url: string, body?: unknown) => request<T>('POST', url, body),
  put: <T>(url: string, body?: unknown) => request<T>('PUT', url, body),
  del: <T>(url: string) => request<T>('DELETE', url),
}
