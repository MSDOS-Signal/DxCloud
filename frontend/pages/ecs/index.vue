<script setup lang="ts">
import { h, onBeforeUnmount, onMounted, ref } from 'vue'
import { NButton, NDropdown, NSpace, NTag, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { EcsInstance, PageResult } from '~/types'
import { EcsStateNames, EcsStateType } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()

const loading = ref(false)
const rows = ref<EcsInstance[]>([])
const total = ref(0)
const page = ref(1)
const size = 20
const keyword = ref('')
const status = ref<string | null>(null)

let timer: number | undefined

const runningCount = computed(() => rows.value.filter((r) => r.observed_state === 'running').length)
const stoppedCount = computed(() => rows.value.filter((r) => r.observed_state === 'stopped').length)
const otherCount = computed(() => rows.value.length - runningCount.value - stoppedCount.value)
const totalCpu = computed(() => rows.value.reduce((s, r) => s + (r.cpu || 0), 0))
const totalMem = computed(() => rows.value.reduce((s, r) => s + (r.memory_mb || 0), 0) / 1024)
const totalDisk = computed(() => rows.value.reduce((s, r) => s + (r.disk_gb || 0), 0))

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams()
    q.set('page', String(page.value))
    q.set('size', String(size))
    if (keyword.value) q.set('keyword', keyword.value)
    if (status.value) q.set('status', status.value)
    const res = await api.get<PageResult<EcsInstance>>(`/ecs?${q.toString()}`)
    rows.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function act(row: EcsInstance, op: 'start' | 'stop' | 'force-stop' | 'restart' | 'delete') {
  try {
    if (op === 'delete') {
      dialog.warning({
        title: '删除实例',
        content: `确定删除实例「${row.name}」（${row.instance_no}）吗？镜像与磁盘将保留，操作不可撤销。`,
        positiveText: '删除',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await api.del(`/ecs/${row.id}`)
            message.success('已删除')
            load()
          } catch (e) {
            message.error(e instanceof Error ? e.message : '删除失败')
            return false
          }
        },
      })
      return
    }
    await api.post(`/ecs/${row.id}/${op}`)
    message.success('操作成功')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

function stateTag(row: EcsInstance) {
  const type = EcsStateType[row.observed_state] || 'default'
  return h(NTag, { type, size: 'small' }, { default: () => EcsStateNames[row.observed_state] || row.observed_state })
}

const columns: DataTableColumns<EcsInstance> = [
  { title: '实例ID', key: 'instance_no', width: 170 },
  { title: '名称', key: 'name', width: 150 },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true } },
  {
    title: '规格',
    key: 'spec',
    width: 140,
    render: (row) => `${row.cpu} 核 / ${row.memory_mb} MB / ${row.disk_gb} GB`,
  },
  { title: '状态', key: 'observed_state', width: 90, render: (row) => stateTag(row) },
  { title: 'IP', key: 'fixed_ip', width: 130, render: (row) => row.fixed_ip || '—' },
  {
    title: '端口',
    key: 'ports',
    width: 140,
    render: (row) => (row.ports.length ? row.ports.map((p) => `${p.host_port}→${p.container_port}/${p.protocol || 'tcp'}`).join('，') : '—'),
  },
  { title: '创建时间', key: 'created_at', width: 150 },
  {
    title: '操作',
    key: 'actions',
    width: 190,
    fixed: 'right',
    render: (row) => {
      const isRunning = row.observed_state === 'running'
      const isDown = row.observed_state === 'stopped' || row.observed_state === 'failed'
      const moreOptions: { label: string; key: string; disabled?: boolean }[] = []
      if (isRunning) {
        moreOptions.push({ label: '重启', key: 'restart' })
        moreOptions.push({ label: '强制停止', key: 'force-stop' })
      }
      if (auth.hasPerm('ecs:delete')) moreOptions.push({ label: '删除', key: 'delete' })
      return h(NSpace, { size: 4, wrap: false }, {
        default: () => [
          h(NButton, { size: 'tiny', type: 'primary', ghost: true, onClick: () => navigateTo(`/ecs/${row.id}`) }, { default: () => '详情' }),
          isDown
            ? h(NButton, { size: 'tiny', onClick: () => act(row, 'start') }, { default: () => '启动' })
            : null,
          isRunning
            ? h(NButton, { size: 'tiny', onClick: () => act(row, 'stop') }, { default: () => '停止' })
            : null,
          moreOptions.length
            ? h(NDropdown, {
                size: 'small',
                trigger: 'click',
                options: moreOptions,
                onSelect: (key: string) => act(row, key as 'restart' | 'force-stop' | 'delete'),
              }, {
                default: () => h(NButton, { size: 'tiny', quaternary: true }, { default: () => '更多 ▾' }),
              })
            : null,
        ],
      })
    },
  },
]

onMounted(() => {
  load()
  timer = window.setInterval(load, 5000) // 状态轮询（Reconciler 15s 对账，前端 5s 拉取）
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="ecs" title="云主机 ECS"
      description="容器实例生命周期管理 · Reconciler 每 15s 与 Docker 引擎对账 · 页面每 5s 自动刷新"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ runningCount }}</span><span class="lbl">运行中</span></div>
        <div class="hero-pill"><span class="num">{{ total }}</span><span class="lbl">实例总数</span></div>
        <div class="hero-pill"><span class="num">{{ totalCpu.toFixed(2) }}</span><span class="lbl">vCPU 合计</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <n-input v-model:value="keyword" placeholder="搜索名称 / 实例ID" clearable size="small" style="width: 200px" @keyup.enter="page = 1; load()">
            <template #prefix><DxIcon name="search" :size="14" /></template>
          </n-input>
          <n-select
            v-model:value="status"
            :options="[
              { label: '全部状态', value: null },
              ...Object.entries(EcsStateNames).map(([value, label]) => ({ label, value })),
            ]"
            size="small"
            style="width: 120px"
            @update:value="page = 1; load()"
          />
          <button class="hero-btn" @click="page = 1; load()">搜索</button>
          <button v-if="auth.hasPerm('ecs:create')" class="hero-btn" style="background:#fff;color:#0052d9;font-weight:600" @click="navigateTo('/ecs/create')">
            <DxIcon name="plus" :size="14" /> 创建实例
          </button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
      <StatTile icon="play" label="运行中实例" :value="runningCount" suffix=" 台" color="#00b42a" hint="状态 running" />
      <StatTile icon="cpu" label="vCPU 合计" :value="totalCpu" :decimals="2" suffix=" 核" color="#006eff" hint="当前页实例规格总和" />
      <StatTile icon="memory" label="内存合计" :value="totalMem" :decimals="1" suffix=" GB" color="#722ed1" hint="当前页实例规格总和" />
      <StatTile icon="database" label="磁盘合计" :value="totalDisk" suffix=" GB" color="#13c2c2" hint="逻辑配额" />
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_280px] gap-4 items-start">
      <!-- 列表卡片 -->
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">
          <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">实例列表</span>
          <span class="text-xs text-gray-400">点击「详情」查看监控 / 终端 / 日志</span>
        </div>
        <div class="dx-card-body">
          <n-data-table
            class="dx-table"
            :columns="columns"
            :data="rows"
            :loading="loading"
            :pagination="{ page: page, pageSize: size, itemCount: total, onChange: (p: number) => { page = p; load() } }"
            :bordered="false"
          />
        </div>
      </div>

      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">状态分布</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="rows.length > 0"
              :size="130"
              :center-text="String(total)"
              center-label="实例总数"
              :segments="[
                { value: runningCount, color: '#00b42a', label: '运行中' },
                { value: stoppedCount, color: '#fa8c16', label: '已停止' },
                { value: otherCount, color: '#86909c', label: '其他' },
              ].filter(s => s.value > 0)"
            />
            <n-empty v-else description="暂无实例" class="py-8" />
          </div>
        </div>
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">资源用量</span>
          </div>
          <div class="dx-card-body space-y-3">
            <div v-for="r in [
              { label: 'vCPU', pct: Math.min(100, totalCpu / 16 * 100), txt: totalCpu.toFixed(2) + ' / 16 核', color: '#006eff' },
              { label: '内存', pct: Math.min(100, totalMem / 32 * 100), txt: totalMem.toFixed(1) + ' / 32 GB', color: '#722ed1' },
              { label: '磁盘', pct: Math.min(100, totalDisk / 200 * 100), txt: totalDisk + ' / 200 GB', color: '#13c2c2' },
            ]" :key="r.label" class="res-row">
              <div class="flex justify-between text-xs mb-1">
                <span class="text-gray-500 dark:text-gray-400">{{ r.label }}</span>
                <span class="text-gray-600 dark:text-gray-300 tabular-nums">{{ r.txt }}</span>
              </div>
              <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
                <div class="h-full rounded-full transition-all duration-1000" :style="`width:${r.pct}%;background:linear-gradient(90deg,${r.color},${r.color}99)`" />
              </div>
            </div>
            <div class="text-[11px] text-gray-400 pt-1">* 上限为演示参考值，实际以组织配额为准</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.res-row {
  animation: res-in 0.5s ease both;
}
@keyframes res-in {
  from { opacity: 0; transform: translateX(10px); }
  to { opacity: 1; transform: translateX(0); }
}
</style>
