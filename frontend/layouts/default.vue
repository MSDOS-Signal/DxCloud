<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import DxIcon from '~/components/DxIcon.vue'

useCursor()

const router = useRouter()
const auth = useAuthStore()
const org = useOrgStore()
const theme = useThemeStore()
const message = useMessage()
const handleResize = () => {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) showMobileNav.value = false
}

// 通知
interface NotifItem {
  id: number
  type: string
  title: string
  content: string
  link: string
  created_at: string
  read_at: string | null
}
const notifTotal = ref(0)
const notifItems = ref<NotifItem[]>([])
const showNotif = ref(false)

async function loadNotifs() {
  try {
    const r = await api.get<{ total: number; items: NotifItem[] }>('/notifications?limit=10')
    notifTotal.value = r.items.filter((n) => !n.read_at).length
    notifItems.value = r.items
  } catch {
    // 忽略
  }
}

const notifMeta: Record<string, { icon: string; label: string; color: string }> = {
  ecs: { icon: 'ecs', label: '云主机', color: '#006eff' },
  image: { icon: 'images', label: '镜像', color: '#13c2c2' },
  deploy: { icon: 'deployments', label: '部署', color: '#722ed1' },
  pipeline: { icon: 'pipelines', label: '流水线', color: '#fa8c16' },
  security: { icon: 'security', label: '安全', color: '#f5222d' },
}

function notifOf(t: string) {
  return notifMeta[t] || { icon: 'bell', label: '通知', color: '#8c8c8c' }
}

function timeAgo(s: string): string {
  const diff = Date.now() - new Date(s).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return '刚刚'
  if (m < 60) return `${m} 分钟前`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h} 小时前`
  const d = Math.floor(h / 24)
  if (d < 7) return `${d} 天前`
  return new Date(s).toLocaleDateString('zh-CN')
}

async function openNotif(n: NotifItem) {
  try {
    if (!n.read_at) await api.put(`/notifications/${n.id}/read`)
  } catch {
    // 忽略
  }
  if (n.link) {
    showNotif.value = false
    router.push(n.link)
  }
  loadNotifs()
}

async function markAllRead() {
  try {
    await api.post('/notifications/read-all')
    await loadNotifs()
    message.success('已全部标记为已读')
  } catch {
    message.error('操作失败')
  }
}

const orgOptions = computed(() => [
  { label: '默认空间（单租户）', value: 0 },
  ...org.mine.map((o) => ({ label: `${o.name}（${o.plan}）`, value: o.id })),
])

function onOrgChange(v: number) {
  if (v === org.current) return
  org.setOrg(v)
  message.info(`已切换到「${v === 0 ? '默认空间' : org.currentOrgName}」`)
  window.location.reload()
}

interface NavItem {
  label: string
  key: string
  perm?: string
  icon: string
}

const navGroups: { group: string; items: NavItem[] }[] = [
  { group: '概览', items: [{ label: '总览 Dashboard', key: '/dashboard', icon: 'dashboard' }] },
  {
    group: '计算',
    items: [
      { label: 'ECS 云主机', key: '/ecs', icon: 'ecs' },
      { label: '容器实例', key: '/containers', icon: 'containers' },
    ],
  },
  {
    group: '资源',
    items: [
      { label: '镜像中心', key: '/images', icon: 'images' },
      { label: '网络', key: '/networks', icon: 'networks' },
      { label: '存储', key: '/volumes', icon: 'volumes' },
    ],
  },
  {
    group: 'PaaS',
    items: [
      { label: '应用', key: '/apps', icon: 'apps' },
      { label: '项目', key: '/projects', icon: 'projects' },
      { label: '域名', key: '/domains', icon: 'domains' },
    ],
  },
  {
    group: 'DevOps',
    items: [
      { label: 'CI/CD Pipeline', key: '/pipelines', icon: 'pipelines' },
      { label: '部署', key: '/deployments', icon: 'deployments' },
    ],
  },
  {
    group: '运营',
    items: [
      { label: '监控', key: '/monitor', icon: 'monitor' },
      { label: '日志', key: '/logs', icon: 'logs' },
      { label: '组织', key: '/orgs', perm: 'org:list', icon: 'orgs' },
      { label: '计费', key: '/billing', perm: 'billing:view', icon: 'billing' },
      { label: '安全中心', key: '/security', perm: 'security:view', icon: 'security' },
    ],
  },
  {
    group: '系统',
    items: [
      { label: 'IAM · 用户', key: '/iam/users', perm: 'user:list', icon: 'users' },
      { label: 'IAM · 角色', key: '/iam/roles', perm: 'user:list', icon: 'roles' },
      { label: 'IAM · 权限', key: '/iam/permissions', perm: 'user:list', icon: 'permissions' },
      { label: '设置', key: '/settings', perm: 'settings:view', icon: 'settings' },
    ],
  },
]

const menuOptions = computed<MenuOption[]>(() => {
  const opts: MenuOption[] = []
  for (const g of navGroups) {
    const items = g.items.filter((it) => !it.perm || auth.hasPerm(it.perm))
    if (items.length === 0) continue
    opts.push({ type: 'group', label: g.group, key: 'group-' + g.group })
    for (const it of items) {
      opts.push({ label: it.label, key: it.key, icon: () => h(DxIcon, { name: it.icon, size: 16 }) })
    }
  }
  return opts
})

const activeKey = ref(router.currentRoute.value.path)
const isMobile = ref(false)
const showMobileNav = ref(false)
watch(
  () => router.currentRoute.value.path,
  (p) => { activeKey.value = p },
)

function onMenuSelect(key: string) {
  if (key.startsWith('group-')) return
  showMobileNav.value = false
  router.push(key)
}

const userOptions = [
  { label: '个人信息', key: 'profile', icon: () => h(DxIcon, { name: 'users', size: 15 }) },
  { type: 'divider', key: 'd1' },
  { label: '退出登录', key: 'logout', icon: () => h(DxIcon, { name: 'logout', size: 15 }) },
]

async function onUserSelect(key: string) {
  if (key === 'logout') {
    await auth.logout()
    router.push('/login')
    return
  }
  if (key === 'profile') {
    router.push('/profile')
  }
}

onMounted(async () => {
  isMobile.value = window.innerWidth < 768
  window.addEventListener('resize', handleResize)
  if (auth.isLoggedIn && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      auth.logout()
      router.push('/login')
    }
  }
  await org.loadMine()
  loadNotifs()
  setInterval(loadNotifs, 30000)
})

onBeforeUnmount(() => window.removeEventListener('resize', handleResize))
</script>

<template>
  <n-drawer
    v-model:show="showMobileNav"
    placement="left"
    :width="230"
    :show-mask="true"
    style="padding: 0"
  >
    <div class="h-full flex flex-col bg-white dark:bg-[#0d1117]">
      <div class="h-12 flex items-center px-4 border-b border-gray-100 dark:border-gray-800">
        <BrandLogo :size="30" compact />
      </div>
      <div class="flex-1 overflow-y-auto py-1">
        <n-menu
          class="sider-light dark:sider-dark"
          :options="menuOptions"
          :value="activeKey"
          :root-indent="18"
          :indent="14"
          @update:value="onMenuSelect"
        />
      </div>
      <div class="px-4 py-2.5 border-t border-gray-100 dark:border-gray-800 text-[11px] text-gray-400">
        引擎运行中 · v1.0
      </div>
    </div>
  </n-drawer>
  <n-layout has-sider class="h-screen">
    <!-- 腾讯云风格浅色侧栏 -->
    <n-layout-sider
      v-if="!isMobile"
      bordered
      :width="200"
      collapse-mode="width"
      :collapsed-width="56"
      show-trigger="arrow-circle"
    >
      <div class="h-full flex flex-col bg-white sider-light dark:bg-[#0d1117]">
        <!-- 品牌区 -->
        <div class="h-12 flex items-center gap-2 px-4 border-b border-gray-100 dark:border-gray-800 shrink-0">
          <BrandLogo :size="28" compact />
          <div class="min-w-0 overflow-hidden">
            <div class="text-[14px] font-semibold text-gray-800 dark:text-gray-100 leading-none truncate">多晓云</div>
            <div class="text-[10px] text-gray-400 mt-0.5 truncate">DxCloud · 控制台</div>
          </div>
        </div>

        <!-- 菜单 -->
        <div class="flex-1 overflow-y-auto py-1">
          <n-menu
            class="sider-light dark:sider-dark"
            :options="menuOptions"
            :value="activeKey"
            :root-indent="18"
            :indent="14"
            @update:value="onMenuSelect"
          />
        </div>

        <!-- 底部版本 -->
        <div class="px-4 py-2.5 border-t border-gray-100 dark:border-gray-800">
          <div class="flex items-center gap-1.5 text-[11px] text-gray-400">
            <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
            引擎运行中 · v1.0
          </div>
        </div>
      </div>
    </n-layout-sider>

    <n-layout>
      <!-- 头部 -->
      <n-layout-header bordered class="h-12 px-4 flex items-center justify-between bg-white dark:bg-[#0d1117] sticky top-0 z-10">
        <div class="text-[13px] text-gray-500 flex items-center gap-1.5 min-w-0">
          <n-button v-if="isMobile" quaternary circle size="small" aria-label="打开菜单" @click="showMobileNav = true">
            <DxIcon name="dashboard" :size="16" />
          </n-button>
          <span class="text-gray-400">控制台</span>
          <span class="text-gray-300">/</span>
          <span class="font-medium text-gray-700 dark:text-gray-200 truncate">{{ router.currentRoute.value.meta?.title || '概览' }}</span>
        </div>
        <div class="flex items-center gap-2">
          <n-popover trigger="hover" placement="bottom" :width="300">
            <template #trigger>
              <n-select
                :value="org.current"
                :options="orgOptions"
                size="small"
                :style="isMobile ? { width: '110px' } : { width: '180px' }"
                placeholder="选择空间"
                @update:value="onOrgChange"
              />
            </template>
            <div class="text-xs leading-relaxed">
              <p class="mb-1"><b>默认空间（单租户）</b>：平台级共享空间，所有资源全局可见，适合个人使用与快速体验。</p>
              <p><b>组织空间</b>：按组织隔离资源（实例、镜像、应用各自独立），拥有独立配额与虚拟余额，适合团队多租户协作。切换后页面会自动刷新。</p>
            </div>
          </n-popover>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button quaternary circle size="small" @click="theme.toggle()">
                <DxIcon :name="theme.isDark ? 'sun' : 'moon'" :size="16" />
              </n-button>
            </template>
            切换{{ theme.isDark ? '浅色' : '深色' }}模式
          </n-tooltip>
          <n-popover v-model:show="showNotif" trigger="click" placement="bottom-end" :width="360" raw>
            <template #trigger>
              <n-badge :value="notifTotal" :max="99" :show="notifTotal > 0">
                <n-button quaternary circle size="small">
                  <DxIcon name="bell" :size="16" />
                </n-button>
              </n-badge>
            </template>
            <div class="notif-panel">
              <div class="flex items-center justify-between px-3.5 h-11 border-b border-gray-100 dark:border-gray-800">
                <div class="text-[13px] font-medium text-gray-700 dark:text-gray-200">消息通知</div>
                <n-button v-if="notifTotal > 0" size="tiny" quaternary type="primary" @click="markAllRead">全部已读</n-button>
              </div>
              <div class="max-h-96 overflow-auto">
                <div v-if="notifItems.length === 0" class="py-10 text-center">
                  <DxIcon name="bell" :size="28" class="text-gray-300 dark:text-gray-600" />
                  <div class="text-xs text-gray-400 mt-2">暂无通知</div>
                  <div class="text-[11px] text-gray-300 dark:text-gray-600 mt-0.5">实例、镜像、部署等操作结果会在这里提醒</div>
                </div>
                <div
                  v-for="n in notifItems" :key="n.id"
                  class="notif-item"
                  :class="{ unread: !n.read_at }"
                  @click="openNotif(n)"
                >
                  <div class="notif-icon" :style="{ background: notifOf(n.type).color + '1a', color: notifOf(n.type).color }">
                    <DxIcon :name="notifOf(n.type).icon" :size="14" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between gap-2">
                      <span class="text-[13px] font-medium truncate" :class="n.read_at ? 'text-gray-500 dark:text-gray-400' : 'text-gray-800 dark:text-gray-100'">{{ n.title }}</span>
                      <span class="text-[10px] text-gray-400 shrink-0">{{ timeAgo(n.created_at) }}</span>
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 leading-relaxed line-clamp-2">{{ n.content }}</div>
                    <div v-if="n.link" class="text-[11px] text-[#006eff] mt-1 flex items-center gap-0.5">
                      查看详情 <DxIcon name="arrow-right" :size="10" />
                    </div>
                  </div>
                  <span v-if="!n.read_at" class="w-1.5 h-1.5 rounded-full bg-red-500 shrink-0 mt-1.5" />
                </div>
              </div>
            </div>
          </n-popover>
          <div class="w-px h-4 bg-gray-200 dark:bg-gray-700 mx-0.5" />
          <n-dropdown :options="userOptions" @select="onUserSelect">
            <div class="flex items-center gap-2 cursor-pointer">
              <img
                v-if="auth.user?.avatar_url"
                :src="auth.user.avatar_url"
                alt=""
                class="w-7 h-7 rounded-full object-cover border border-gray-200 dark:border-gray-700"
              >
              <div v-else class="w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-medium" style="background: #006eff;">
                {{ auth.nickname ? auth.nickname.slice(0, 1).toUpperCase() : 'U' }}
              </div>
              <div class="leading-tight hidden sm:block">
                <div class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ auth.nickname || '未登录' }}</div>
                <div class="text-[10px] text-gray-400">{{ auth.user?.role_names?.join(' / ') || '普通用户' }}</div>
              </div>
            </div>
          </n-dropdown>
        </div>
      </n-layout-header>

      <n-layout-content class="p-4 bg-[#f0f2f5] dark:bg-[#0d1117]" :native-scrollbar="false">
        <slot />
      </n-layout-content>
    </n-layout>
  </n-layout>

  <!-- 全局 AI 助手悬浮球（可拖动，点击对话） -->
  <AiAssistant />
</template>
