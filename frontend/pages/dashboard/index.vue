<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import CountUp from '~/components/CountUp.vue'

interface DashboardData {
  ecs_total: number
  ecs_running: number
  ecs_stopped: number
  app_count: number
  pipeline_count: number
  deploy_today: number
  pipeline_success_rate: number
  pipeline_runs_24h: number
  cpu_avg: number
  mem_avg: number
}

interface SeriesRow {
  bucket: string
  cpu: number | null
  mem_pct: number | null
  net_rx: number | null
  net_tx: number | null
}

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()
const data = ref<DashboardData | null>(null)
const series = ref<SeriesRow[]>([])
const loading = ref(false)
const chartError = ref('')
const minutes = ref(60)
let timer: number | undefined

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const statCards = computed(() => [
  { label: 'ECS 实例', value: data.value?.ecs_total ?? 0, sub: `运行 ${data.value?.ecs_running ?? 0} · 停止 ${data.value?.ecs_stopped ?? 0}`, icon: 'ecs', iconClass: 'stat-icon-blue', decimals: 0, suffix: '' },
  { label: '应用', value: data.value?.app_count ?? 0, sub: 'PaaS 蓝绿部署', icon: 'apps', iconClass: 'stat-icon-cyan', decimals: 0, suffix: '' },
  { label: 'Pipeline', value: data.value?.pipeline_count ?? 0, sub: `24h 运行 ${data.value?.pipeline_runs_24h ?? 0} 次`, icon: 'pipelines', iconClass: 'stat-icon-blue', decimals: 0, suffix: '' },
  { label: '今日部署', value: data.value?.deploy_today ?? 0, sub: '蓝绿切换零中断', icon: 'deployments', iconClass: 'stat-icon-green', decimals: 0, suffix: '' },
  { label: '成功率 24h', value: data.value?.pipeline_success_rate ?? 0, sub: 'CI/CD 交付质量', icon: 'check-circle', iconClass: 'stat-icon-green', decimals: 1, suffix: '%' },
  { label: 'CPU 水位', value: data.value?.cpu_avg ?? 0, sub: `内存 ${data.value ? data.value.mem_avg.toFixed(1) : '—'}%`, icon: 'activity', iconClass: 'stat-icon-orange', decimals: 1, suffix: '%' },
])

const quickLinks = [
  { label: '创建 ECS', desc: '秒级启动云主机', icon: 'ecs', iconClass: 'stat-icon-blue', to: '/ecs/create' },
  { label: '部署应用', desc: 'PaaS 一键部署', icon: 'apps', iconClass: 'stat-icon-cyan', to: '/apps' },
  { label: '运行 Pipeline', desc: 'CI/CD 自动化', icon: 'pipelines', iconClass: 'stat-icon-blue', to: '/pipelines' },
  { label: '查看监控', desc: '实时资源水位', icon: 'monitor', iconClass: 'stat-icon-green', to: '/monitor' },
]

const chartCategories = computed(() => series.value.map((r) => r.bucket.slice(11, 16)))
const cpuData = computed(() => series.value.map((r) => (r.cpu == null ? 0 : Number(r.cpu))))
const memData = computed(() => series.value.map((r) => (r.mem_pct == null ? 0 : Number(r.mem_pct))))
const rxData = computed(() => series.value.map((r) => (r.net_rx == null ? 0 : Number(r.net_rx))))
const txData = computed(() => series.value.map((r) => (r.net_tx == null ? 0 : Number(r.net_tx))))

const stateItems = computed(() => [
  { name: '运行中', value: data.value?.ecs_running ?? 0, color: '#00b42a' },
  { name: '已停止', value: data.value?.ecs_stopped ?? 0, color: '#86909c' },
  { name: '其他状态', value: Math.max((data.value?.ecs_total ?? 0) - (data.value?.ecs_running ?? 0) - (data.value?.ecs_stopped ?? 0), 0), color: '#ff9500' },
])

async function load() {
  loading.value = true
  try {
    const [dash, rows] = await Promise.all([
      api.get<DashboardData>('/monitor/dashboard'),
      api.get<SeriesRow[]>(`/monitor/series?kind=ecs&minutes=${minutes.value}`),
    ])
    data.value = dash
    series.value = rows
    chartError.value = ''
  } catch (e) {
    chartError.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 30000)
})

watch(() => theme.isDark, () => load())
watch(minutes, () => load())

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="dashboard-page relative">
    <div class="dashboard-aurora" />
    <div class="dashboard-content relative space-y-4">
      <div class="dx-fade-up flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div class="text-[16px] font-semibold text-gray-800 dark:text-gray-100">
            {{ greeting }}，{{ auth.nickname || '朋友' }}
          </div>
          <div class="text-xs text-gray-400 mt-1">
            今日已部署 <span class="text-gray-600 dark:text-gray-300 font-medium">{{ data?.deploy_today ?? 0 }}</span> 次 ·
            24h Pipeline 成功率 <span class="text-gray-600 dark:text-gray-300 font-medium">{{ data ? data.pipeline_success_rate.toFixed(1) : '—' }}%</span> ·
            运行中实例 <span class="text-gray-600 dark:text-gray-300 font-medium">{{ data?.ecs_running ?? 0 }}</span> 台
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <n-radio-group v-model:value="minutes" size="small">
            <n-radio-button :value="30">30 分</n-radio-button>
            <n-radio-button :value="60">1 小时</n-radio-button>
            <n-radio-button :value="180">3 小时</n-radio-button>
          </n-radio-group>
          <button class="dx-btn-primary" @click="router.push('/ecs/create')">
            <DxIcon name="plus" :size="14" /> 创建实例
          </button>
          <button class="dx-btn-secondary" @click="router.push('/pipelines')">
            <DxIcon name="play" :size="13" /> 运行 Pipeline
          </button>
        </div>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        <div
          v-for="(c, i) in statCards"
          :key="c.label"
          class="dx-stat-card dx-fade-up stat-card-glow"
          :class="`dx-delay-${Math.min(i + 1, 4)}`"
        >
          <div class="flex items-start justify-between">
            <div class="min-w-0">
              <div class="stat-label">{{ c.label }}</div>
              <div class="stat-value">
                <CountUp :value="c.value" :decimals="c.decimals" :suffix="c.suffix" :duration="800" />
              </div>
              <div class="stat-sub truncate">{{ c.sub }}</div>
            </div>
            <div class="stat-icon" :class="c.iconClass">
              <DxIcon :name="c.icon" :size="18" />
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <button
          v-for="(q, i) in quickLinks"
          :key="q.label"
          class="dx-card dx-fade-up text-left p-3 hover:border-[#006eff]"
          :class="`dx-delay-${Math.min(i + 1, 4)}`"
          @click="router.push(q.to)"
        >
          <div class="flex items-center gap-3">
            <div class="stat-icon" :class="q.iconClass" style="width: 32px; height: 32px;">
              <DxIcon :name="q.icon" :size="16" />
            </div>
            <div class="min-w-0">
              <div class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ q.label }}</div>
              <div class="text-[11px] text-gray-400 truncate">{{ q.desc }}</div>
            </div>
          </div>
        </button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <div class="flex items-center gap-2">
              <DxIcon name="cpu" :size="15" class="text-[#006eff]" />
              <span>CPU 使用率</span>
            </div>
            <span class="text-[11px] text-gray-400 font-normal">最近 {{ minutes }} 分钟</span>
          </div>
          <div class="dx-card-body !p-3">
            <CloudLineChart
              :categories="chartCategories"
              :series="[{ name: 'CPU', data: cpuData, from: '#006eff', to: '#00c6ff' }]"
              mode="percent"
              :loading="loading"
              :empty="series.length === 0"
              :error="chartError"
              @retry="load"
            />
          </div>
        </div>
        <div class="dx-card dx-fade-up dx-delay-3">
          <div class="dx-card-header">
            <div class="flex items-center gap-2">
              <DxIcon name="memory" :size="15" class="text-[#00a4ff]" />
              <span>内存使用率</span>
            </div>
            <span class="text-[11px] text-gray-400 font-normal">最近 {{ minutes }} 分钟</span>
          </div>
          <div class="dx-card-body !p-3">
            <CloudLineChart
              :categories="chartCategories"
              :series="[{ name: '内存', data: memData, from: '#00a4ff', to: '#7be0ff' }]"
              mode="percent"
              :loading="loading"
              :empty="series.length === 0"
              :error="chartError"
              @retry="load"
            />
          </div>
        </div>
        <div class="dx-card dx-fade-up dx-delay-4">
          <div class="dx-card-header">
            <div class="flex items-center gap-2">
              <DxIcon name="networks" :size="15" class="text-[#00b42a]" />
              <span>网络流量</span>
            </div>
            <span class="text-[11px] text-gray-400 font-normal">接收 / 发送</span>
          </div>
          <div class="dx-card-body !p-3">
            <CloudLineChart
              :categories="chartCategories"
              :series="[
                { name: '接收', data: rxData, from: '#006eff', to: '#00c6ff' },
                { name: '发送', data: txData, from: '#00b42a', to: '#5ad8a6' },
              ]"
              mode="bytes"
              :loading="loading"
              :empty="series.length === 0"
              :error="chartError"
              @retry="load"
            />
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div class="dx-card dx-fade-up dx-delay-4 p-4">
          <div class="flex items-center justify-between text-[13px] mb-3">
            <span class="font-medium text-gray-700 dark:text-gray-200 flex items-center gap-1.5">
              <DxIcon name="cpu" :size="14" class="text-[#006eff]" />
              当前 CPU 水位
            </span>
            <span class="text-[#006eff] font-semibold tabular-nums">
              <CountUp :value="data?.cpu_avg ?? 0" :decimals="1" suffix="%" />
            </span>
          </div>
          <n-progress type="line" :percentage="Math.min(data?.cpu_avg || 0, 100)" :height="8" :show-indicator="false" color="#006eff" :rail-color="theme.isDark ? '#1E293B' : '#e8f3ff'" />
        </div>
        <div class="dx-card dx-fade-up dx-delay-4 p-4">
          <div class="flex items-center justify-between text-[13px] mb-3">
            <span class="font-medium text-gray-700 dark:text-gray-200 flex items-center gap-1.5">
              <DxIcon name="memory" :size="14" class="text-[#00a4ff]" />
              当前内存水位
            </span>
            <span class="text-[#00a4ff] font-semibold tabular-nums">
              <CountUp :value="data?.mem_avg ?? 0" :decimals="1" suffix="%" />
            </span>
          </div>
          <n-progress type="line" :percentage="Math.min(data?.mem_avg || 0, 100)" :height="8" :show-indicator="false" color="#00a4ff" :rail-color="theme.isDark ? '#1E293B' : '#e8f7ff'" />
        </div>
        <div class="dx-card dx-fade-up dx-delay-4 p-4">
          <div class="flex items-center justify-between text-[13px] mb-3">
            <span class="font-medium text-gray-700 dark:text-gray-200 flex items-center gap-1.5">
              <DxIcon name="ecs" :size="14" class="text-[#00b42a]" />
              实例状态分布
            </span>
            <span class="text-xs text-gray-400">总计 {{ data?.ecs_total ?? 0 }} 台</span>
          </div>
          <CloudDonutChart :items="stateItems" :height="180" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-page {
  overflow: hidden;
}

.dashboard-aurora {
  position: absolute;
  inset: -120px -100px auto auto;
  width: 420px;
  height: 420px;
  pointer-events: none;
  background:
    radial-gradient(circle at 50% 50%, rgba(0, 198, 255, 0.12), transparent 62%),
    radial-gradient(circle at 30% 40%, rgba(0, 180, 42, 0.06), transparent 55%);
  filter: blur(10px);
}

.dashboard-content {
  z-index: 1;
}

.stat-card-glow {
  position: relative;
  overflow: hidden;
}

.stat-card-glow::after {
  content: '';
  position: absolute;
  width: 80px;
  height: 80px;
  right: -40px;
  top: -40px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(0, 198, 255, 0.12), transparent 68%);
  pointer-events: none;
}
</style>
