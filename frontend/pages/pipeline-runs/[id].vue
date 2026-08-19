<script setup lang="ts">
import { h, onBeforeUnmount, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { PipelineJob, PipelineRun } from '~/types'
import { JobStatusNames, JobStatusType, PipeStatusNames, PipeStatusType } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'

const route = useRoute()
const message = useMessage()
const id = computed(() => Number(route.params.id))

const run = ref<PipelineRun | null>(null)
const jobs = ref<PipelineJob[]>([])
const activeJobId = ref<number | null>(null)
const logs = ref('（选择左侧步骤查看日志）')

const successJobs = computed(() => jobs.value.filter((j) => j.status === 'success').length)
const failedJobs = computed(() => jobs.value.filter((j) => j.status === 'failed' || j.status === 'canceled').length)
const runningJobs = computed(() => jobs.value.filter((j) => j.status === 'running' || j.status === 'pending').length)
const jobSegments = computed(() => [
  { value: successJobs.value, color: '#00b42a', label: '成功' },
  { value: runningJobs.value, color: '#006eff', label: '进行中' },
  { value: failedJobs.value, color: '#f53f3f', label: '失败/取消' },
])
const heroGradient = computed(() => {
  const s = run.value?.status
  if (s === 'success') return 'linear-gradient(120deg, #0b5d2a 0%, #12a35a 45%, #3ddc84 100%)'
  if (s === 'failed' || s === 'canceled') return 'linear-gradient(120deg, #5c1a1a 0%, #a02020 45%, #e05656 100%)'
  return 'linear-gradient(120deg, #06307a 0%, #0a5ad6 45%, #3a8dff 100%)'
})

let timer: number | undefined
let logTimer: number | undefined

async function load() {
  try {
    run.value = await api.get<PipelineRun>(`/pipeline-runs/${id.value}`)
    jobs.value = await api.get<PipelineJob[]>(`/pipeline-runs/${id.value}/jobs`)
    if (activeJobId.value === null && jobs.value.length) {
      activeJobId.value = jobs.value[0].id
      await loadLogs()
    }
    if (run.value.status === 'pending' || run.value.status === 'running') {
      if (!timer) {
        timer = window.setInterval(load, 2000)
      }
    } else if (timer) {
      window.clearInterval(timer)
      timer = undefined
      await loadLogs()
    }
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

async function loadLogs() {
  if (activeJobId.value === null) return
  try {
    const r = await api.get<{ logs: string }>(`/pipeline-runs/${id.value}/logs?job_id=${activeJobId.value}`)
    logs.value = r.logs || '（暂无日志）'
  } catch {
    // 忽略
  }
}

async function handleCancel() {
  try {
    await api.post(`/pipeline-runs/${id.value}/cancel`)
    message.success('已请求取消')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '取消失败')
  }
}

const columns: DataTableColumns<PipelineJob> = [
  { title: '步骤', key: 'name', width: 160 },
  { title: '类型', key: 'type', width: 90 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: JobStatusType[row.status] || 'default' }, { default: () => JobStatusNames[row.status] || row.status }),
  },
  { title: '退出码', key: 'exit_code', width: 80, render: (row) => (row.status === 'success' ? '0' : row.exit_code || '—') },
  { title: '开始', key: 'started_at', width: 160, render: (row) => row.started_at || '—' },
  {
    title: '',
    key: 'watch',
    width: 90,
    render: (row) =>
      h(
        NButton,
        { size: 'tiny', type: activeJobId.value === row.id ? 'primary' : 'default', onClick: () => { activeJobId.value = row.id; loadLogs() } },
        { default: () => '日志' },
      ),
  },
]

onMounted(() => {
  load()
  logTimer = window.setInterval(() => {
    if (run.value && (run.value.status === 'running' || run.value.status === 'pending')) loadLogs()
  }, 2000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
  if (logTimer) window.clearInterval(logTimer)
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="pipelines"
      :title="`Pipeline Run #${id}`"
      :description="run ? `run_no #${run.run_no} · 触发方式 ${run.trigger} · 耗时 ${run.duration_ms ? (run.duration_ms / 1000).toFixed(1) + 's' : '—'}` : '加载中…'"
      :gradient="heroGradient"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ PipeStatusNames[run?.status || ''] || run?.status || '…' }}</span><span class="lbl">运行状态</span></div>
        <div class="hero-pill"><span class="num">{{ jobs.length }}</span><span class="lbl">执行步骤</span></div>
        <div class="hero-pill"><span class="num">{{ successJobs }}</span><span class="lbl">成功</span></div>
        <div class="hero-pill"><span class="num">{{ failedJobs }}</span><span class="lbl">失败/取消</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <n-button v-if="run && (run.status === 'running' || run.status === 'pending')" size="small" type="warning" @click="handleCancel">取消运行</n-button>
          <n-button size="small" quaternary class="!text-white" @click="navigateTo('/pipelines')">返回流水线</n-button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header flex items-center gap-2">
          <span>步骤</span>
          <span class="ml-auto flex items-center gap-3 text-[11px] text-gray-400">
            <span class="flex items-center gap-1"><span class="inline-block w-2 h-2 rounded-full bg-[#00b42a]" />成功 {{ successJobs }}</span>
            <span class="flex items-center gap-1"><span class="inline-block w-2 h-2 rounded-full bg-[#006eff]" />进行中 {{ runningJobs }}</span>
            <span class="flex items-center gap-1"><span class="inline-block w-2 h-2 rounded-full bg-[#f53f3f]" />失败 {{ failedJobs }}</span>
          </span>
        </div>
        <div class="dx-card-body">
          <n-data-table :columns="columns" :data="jobs" size="small" :bordered="false" max-height="360" class="dx-table" />
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">实时日志（2s 刷新）</div>
        <div class="dx-card-body">
          <pre class="text-xs font-mono bg-slate-900 text-slate-100 p-3 rounded max-h-96 overflow-auto whitespace-pre-wrap">{{ logs }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
