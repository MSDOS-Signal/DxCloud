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

const data = ref<DashboardData | null>(null)
const series = ref<SeriesRow[]>([])
const loading = ref(false)
const chartError = ref('')
const minutes = ref(60)
let timer: number | undefined

const summaryCards = computed(() => [
  { label: '运行实例', value: data.value?.ecs_running ?? 0, suffix: '', icon: 'ecs', iconClass: 'stat-icon-blue', decimals: 0 },
  { label: 'CPU 均值', value: data.value?.cpu_avg ?? 0, suffix: '%', icon: 'cpu', iconClass: 'stat-icon-cyan', decimals: 1 },
  { label: '内存均值', value: data.value?.mem_avg ?? 0, suffix: '%', icon: 'memory', iconClass: 'stat-icon-cyan', decimals: 1 },
  { label: '24h 部署', value: data.value?.deploy_today ?? 0, suffix: '', icon: 'deployments', iconClass: 'stat-icon-green', decimals: 0 },
])

const categories = computed(() => series.value.map((r) => r.bucket.slice(11, 16)))
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
watch(minutes, load)
onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <div class="dx-page-header dx-fade-up flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2>
          <DxIcon name="monitor" :size="18" class="text-[#006eff]" />
          监控中心
        </h2>
        <p>实时资源水位 · CPU / 内存 / 网络 · 最近 {{ minutes }} 分钟趋势</p>
      </div>
      <n-radio-group v-model:value="minutes" size="small">
        <n-radio-button :value="30">30 分</n-radio-button>
        <n-radio-button :value="60">1 小时</n-radio-button>
        <n-radio-button :value="360">6 小时</n-radio-button>
      </n-radio-group>
    </div>

    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <div
        v-for="(c, i) in summaryCards"
        :key="c.label"
        class="dx-stat-card dx-fade-up"
        :class="`dx-delay-${Math.min(i + 1, 4)}`"
      >
        <div class="flex items-center justify-between">
          <div>
            <div class="stat-label">{{ c.label }}</div>
            <div class="stat-value">
              <CountUp :value="c.value" :decimals="c.decimals" :suffix="c.suffix" />
            </div>
          </div>
          <div class="stat-icon" :class="c.iconClass"><DxIcon :name="c.icon" :size="17" /></div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">
          <div class="flex items-center gap-2"><DxIcon name="cpu" :size="15" class="text-[#006eff]" /><span>CPU 使用率</span></div>
        </div>
        <div class="dx-card-body !p-3">
          <CloudLineChart
            :categories="categories"
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
          <div class="flex items-center gap-2"><DxIcon name="memory" :size="15" class="text-[#00a4ff]" /><span>内存使用率</span></div>
        </div>
        <div class="dx-card-body !p-3">
          <CloudLineChart
            :categories="categories"
            :series="[{ name: '内存', data: memData, from: '#00a4ff', to: '#7be0ff' }]"
            mode="percent"
            :loading="loading"
            :empty="series.length === 0"
            :error="chartError"
            @retry="load"
          />
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
      <div class="dx-card dx-fade-up dx-delay-4">
        <div class="dx-card-header">
          <div class="flex items-center gap-2"><DxIcon name="networks" :size="15" class="text-[#006eff]" /><span>网络流量（接收 / 发送）</span></div>
        </div>
        <div class="dx-card-body !p-3">
          <CloudLineChart
            :categories="categories"
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
      <div class="dx-card dx-fade-up dx-delay-4">
        <div class="dx-card-header">
          <div class="flex items-center gap-2"><DxIcon name="activity" :size="15" class="text-[#00b42a]" /><span>实例状态分布</span></div>
        </div>
        <div class="dx-card-body !p-3">
          <CloudDonutChart :items="stateItems" :height="210" />
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-4 p-4">
        <div class="text-[13px] font-medium mb-3 flex items-center gap-1.5">
          <DxIcon name="info" :size="14" class="text-[#006eff]" /> 说明
        </div>
        <div class="text-xs leading-relaxed text-gray-400 space-y-2">
          <div>采样器每分钟采集一次运行中的 ECS 实例，曲线按分钟聚合。</div>
          <div>网络吞吐使用 B/s 展示；CPU / 内存使用率最大显示 100%。</div>
          <div>切换时间范围会立即从后端重新加载真实数据。</div>
        </div>
      </div>
    </div>
  </div>
</template>
