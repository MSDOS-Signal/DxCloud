<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { EcsEvent, EcsInstance, EcsStats } from '~/types'
import { EcsStateNames, EcsStateType } from '~/types'
import { api } from '~/services/http'

const route = useRoute()
const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()

const id = computed(() => Number(route.params.id))
const inst = ref<EcsInstance | null>(null)
const stats = ref<EcsStats | null>(null)
const logs = ref('')
const events = ref<EcsEvent[]>([])
const loading = ref(false)

let timer: number | undefined

async function load() {
  try {
    inst.value = await api.get<EcsInstance>(`/ecs/${id.value}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

async function loadStats() {
  if (!inst.value || inst.value.observed_state !== 'running') {
    stats.value = null
    return
  }
  try {
    stats.value = await api.get<EcsStats>(`/ecs/${id.value}/stats`)
  } catch {
    stats.value = null
  }
}

async function loadLogs() {
  loading.value = true
  try {
    const r = await api.get<{ logs: string }>(`/ecs/${id.value}/logs?tail=200`)
    logs.value = r.logs || '（暂无日志）'
  } catch (e) {
    message.error(e instanceof Error ? e.message : '日志加载失败')
  } finally {
    loading.value = false
  }
}

async function loadEvents() {
  try {
    events.value = await api.get<EcsEvent[]>(`/ecs/${id.value}/events?limit=100`)
  } catch {
    events.value = []
  }
}

async function act(op: 'start' | 'stop' | 'force-stop' | 'restart' | 'delete') {
  try {
    if (op === 'delete') {
      dialog.warning({
        title: '删除实例',
        content: `确定删除「${inst.value?.name}」吗？`,
        positiveText: '删除',
        negativeText: '取消',
        onPositiveClick: async () => {
          await api.del(`/ecs/${id.value}`)
          message.success('已删除')
          navigateTo('/ecs')
        },
      })
      return
    }
    await api.post(`/ecs/${id.value}/${op}`)
    message.success('操作成功')
    await load()
    await loadStats()
    await loadEvents()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

function fmtBytes(n: number): string {
  if (n >= 1024 * 1024 * 1024) return (n / 1024 / 1024 / 1024).toFixed(2) + ' GB'
  if (n >= 1024 * 1024) return (n / 1024 / 1024).toFixed(2) + ' MB'
  if (n >= 1024) return (n / 1024).toFixed(1) + ' KB'
  return n + ' B'
}

const eventColumns: DataTableColumns<EcsEvent> = [
  { title: '时间', key: 'created_at', width: 170 },
  { title: '类型', key: 'event_type', width: 100 },
  {
    title: '级别',
    key: 'level',
    width: 70,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.level === 'error' ? 'error' : row.level === 'warn' ? 'warning' : 'info', bordered: false },
        { default: () => row.level },
      ),
  },
  { title: '内容', key: 'message' },
]

onMounted(async () => {
  await load()
  await loadEvents()
  if (inst.value?.observed_state === 'running') {
    await loadLogs()
    await loadStats()
    timer = window.setInterval(loadStats, 5000)
  }
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <template v-if="inst">
    <div class="dx-card dx-fade-up">
      <div class="dx-card-body">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <n-tag :type="EcsStateType[inst.observed_state] || 'default'" size="large">
              {{ EcsStateNames[inst.observed_state] || inst.observed_state }}
            </n-tag>
            <span class="text-lg font-semibold">{{ inst.name }}</span>
            <span class="text-xs text-gray-400">{{ inst.instance_no }}</span>
            <span v-if="inst.last_error" class="text-xs text-red-500">错误：{{ inst.last_error }}</span>
          </div>
          <n-space>
            <n-button
              v-if="inst.observed_state === 'stopped' || inst.observed_state === 'failed'"
              type="primary"
              size="small"
              @click="act('start')"
            >启动</n-button>
            <n-button v-if="inst.observed_state === 'running'" size="small" @click="act('stop')">停止</n-button>
            <n-button v-if="inst.observed_state === 'running'" size="small" @click="act('restart')">重启</n-button>
            <n-button v-if="inst.observed_state === 'running'" size="small" type="warning" ghost @click="act('force-stop')">强制停止</n-button>
            <n-button
              v-if="inst.observed_state === 'running' && auth.hasPerm('ecs:console')"
              size="small"
              type="primary"
              ghost
              @click="navigateTo(`/ecs/${inst.id}/terminal`)"
            >控制台</n-button>
            <n-button v-if="auth.hasPerm('ecs:delete')" size="small" type="error" ghost @click="act('delete')">删除</n-button>
            <n-button size="small" @click="load()">刷新</n-button>
          </n-space>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">基本信息</div>
        <div class="dx-card-body">
          <n-descriptions :column="1" size="small" label-placement="left">
            <n-descriptions-item label="镜像">{{ inst.image }}</n-descriptions-item>
            <n-descriptions-item label="规格">{{ inst.cpu }} 核 / {{ inst.memory_mb }} MB / {{ inst.disk_gb }} GB</n-descriptions-item>
            <n-descriptions-item label="内网 IP">{{ inst.fixed_ip || '—' }}</n-descriptions-item>
            <n-descriptions-item label="端口">
              <span v-if="inst.ports.length">
                <n-tag v-for="p in inst.ports" :key="`${p.host_port}-${p.container_port}`" size="small" class="mr-1">
                  {{ p.host_port }}→{{ p.container_port }}/{{ p.protocol || 'tcp' }}
                </n-tag>
              </span>
              <span v-else>—</span>
            </n-descriptions-item>
            <n-descriptions-item label="重启策略">{{ inst.restart_policy }}</n-descriptions-item>
            <n-descriptions-item label="只读根盘">{{ inst.readonly_rootfs ? '是' : '否' }}</n-descriptions-item>
            <n-descriptions-item label="Docker 容器">
              <span class="font-mono text-xs">{{ inst.container_name }}（{{ (inst.container_id || '').slice(0, 12) }}）</span>
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ inst.created_at }}</n-descriptions-item>
          </n-descriptions>
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">实时监控（5s 刷新）</div>
        <div class="dx-card-body">
          <div v-if="stats" class="space-y-3">
            <div class="flex justify-between text-sm"><span class="text-gray-500">CPU</span><span>{{ stats.cpu_percent.toFixed(2) }}%</span></div>
            <n-progress type="line" :percentage="Math.min(stats.cpu_percent, 100)" :height="8" :show-indicator="false" />
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">内存</span>
              <span>{{ fmtBytes(stats.mem_used) }} / {{ fmtBytes(stats.mem_limit) }}（{{ stats.mem_percent.toFixed(1) }}%）</span>
            </div>
            <n-progress type="line" :percentage="Math.min(stats.mem_percent, 100)" :height="8" :show-indicator="false" />
            <div class="grid grid-cols-2 gap-2 text-xs text-gray-500">
              <div>网络收：{{ fmtBytes(stats.net_rx) }}</div>
              <div>网络发：{{ fmtBytes(stats.net_tx) }}</div>
              <div>磁盘读：{{ fmtBytes(stats.disk_read) }}</div>
              <div>磁盘写：{{ fmtBytes(stats.disk_write) }}</div>
              <div>进程数：{{ stats.pids }}</div>
            </div>
          </div>
          <n-empty v-else-if="inst.observed_state === 'running'" description="统计采集中（需 2 秒）" size="small" />
          <n-empty v-else description="实例未运行" size="small" />
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-3">
        <div class="dx-card-header">环境变量 / 启动命令</div>
        <div class="dx-card-body">
          <div class="text-xs font-semibold text-gray-500 mb-1">环境变量</div>
          <div v-if="inst.env.length" class="font-mono text-xs bg-gray-50 p-2 rounded mb-2">
            <div v-for="e in inst.env" :key="e">{{ e }}</div>
          </div>
          <div v-else class="text-xs text-gray-400 mb-2">（无）</div>
          <div class="text-xs font-semibold text-gray-500 mb-1">启动命令</div>
          <div class="font-mono text-xs bg-gray-50 p-2 rounded">{{ inst.command.length ? inst.command.join(' ') : '（镜像默认）' }}</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">
          <div class="flex justify-between items-center">
            <span>容器日志（最近 200 行）</span>
            <n-button size="tiny" :loading="loading" @click="loadLogs">刷新日志</n-button>
          </div>
        </div>
        <div class="dx-card-body">
          <pre class="text-xs font-mono bg-slate-900 text-slate-100 p-3 rounded max-h-80 overflow-auto whitespace-pre-wrap">{{ logs || '点击刷新加载日志' }}</pre>
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-3">
        <div class="dx-card-header">
          <div class="flex justify-between items-center">
            <span>实例事件</span>
            <n-button size="tiny" @click="loadEvents">刷新</n-button>
          </div>
        </div>
        <div class="dx-card-body">
          <n-data-table :columns="eventColumns" :data="events" size="small" :bordered="false" max-height="320" class="dx-table" />
        </div>
      </div>
    </div>
    </template>
    <n-skeleton v-else :repeat="4" text />
  </div>
</template>
