import { defineStore } from 'pinia'
import type { Organization } from '~/types'
import { api } from '~/services/http'

const KEY = 'dx-org-id'

function readStored(): number {
  if (!import.meta.client || typeof localStorage === 'undefined') return 0
  const v = Number(localStorage.getItem(KEY) || '0')
  return Number.isFinite(v) && v > 0 ? v : 0
}

// 租户上下文：当前组织 ID（0 = 默认空间/单租户模式）
export const useOrgStore = defineStore('org', {
  state: () => ({
    current: readStored(),
    mine: [] as Organization[],
  }),
  getters: {
    currentOrgId: (s) => s.current,
    currentOrgName: (s) => {
      if (s.current === 0) return '默认空间'
      const o = s.mine.find((x) => x.id === s.current)
      return o ? o.name : `组织 #${s.current}`
    },
    hasOrg: (s) => s.current > 0,
  },
  actions: {
    setOrg(id: number) {
      this.current = id
      if (import.meta.client && typeof localStorage !== 'undefined') {
        if (id === 0) localStorage.removeItem(KEY)
        else localStorage.setItem(KEY, String(id))
      }
    },
    async loadMine() {
      try {
        this.mine = await api.get<Organization[]>('/organizations/mine')
      } catch {
        this.mine = []
      }
    },
  },
})
