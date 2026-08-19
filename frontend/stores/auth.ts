import { defineStore } from 'pinia'
import type { TokenPair, UserInfo } from '~/types'
import { api } from '~/services/http'
import { clearTokens, getAccessToken, getRefreshToken, setTokens } from '~/utils/token'

// 会话状态：Access/Refresh Token + 用户资料 + 权限（后端为最终裁决，前端仅做体验控制）
export const useAuthStore = defineStore('auth', {
  state: () => ({
    access: import.meta.client ? getAccessToken() : '',
    user: null as UserInfo | null,
  }),
  getters: {
    isLoggedIn: (state) => state.access !== '',
    username: (state) => state.user?.username || '',
    nickname: (state) => state.user?.nickname || state.user?.username || '',
  },
  actions: {
    hasPerm(code: string): boolean {
      return !!this.user && this.user.permissions.includes(code)
    },
    async login(username: string, password: string) {
      const pair = await api.post<TokenPair>('/auth/login', { username, password })
      setTokens(pair.access_token, pair.refresh_token)
      this.access = pair.access_token
      await this.fetchMe()
    },
    async register(username: string, email: string, password: string) {
      const pair = await api.post<TokenPair>('/auth/register', { username, email, password })
      setTokens(pair.access_token, pair.refresh_token)
      this.access = pair.access_token
      await this.fetchMe()
    },
    async fetchMe() {
      this.user = await api.get<UserInfo>('/auth/me')
    },
    async logout() {
      try {
        // 后端登出（拉黑 Refresh Token）尽力而为，失败不影响本地退出
        const refresh = getRefreshToken()
        if (this.access) {
          await api.post('/auth/logout', { refresh_token: refresh })
        }
      } catch {
        // 忽略登出接口错误，本地会话已清理
      } finally {
        clearTokens()
        this.access = ''
        this.user = null
      }
    },
  },
})
