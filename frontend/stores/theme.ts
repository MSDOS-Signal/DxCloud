import { defineStore } from 'pinia'

const KEY = 'dx-theme'
export type ThemeMode = 'light' | 'dark'

function readStored(): ThemeMode {
  if (!import.meta.client || typeof localStorage === 'undefined') return 'light'
  return localStorage.getItem(KEY) === 'dark' ? 'dark' : 'light'
}

// 主题模式：light / dark（持久化 localStorage，切换全站 Naive UI 主题与自定义样式）
export const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: readStored() as ThemeMode,
  }),
  getters: {
    isDark: (s) => s.mode === 'dark',
  },
  actions: {
    setMode(mode: ThemeMode) {
      this.mode = mode
      if (import.meta.client && typeof localStorage !== 'undefined') {
        localStorage.setItem(KEY, mode)
      }
      this.apply()
    },
    toggle() {
      this.setMode(this.mode === 'dark' ? 'light' : 'dark')
    },
    // 同步 <html class="dark">（供 Tailwind dark: 变体）与 body 背景
    apply() {
      if (!import.meta.client || typeof document === 'undefined') return
      const root = document.documentElement
      root.classList.toggle('dark', this.mode === 'dark')
      root.style.colorScheme = this.mode
    },
  },
})
