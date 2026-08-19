<script setup lang="ts">
import { h, onBeforeUnmount, onMounted, ref } from 'vue'
import { NButton, NTag, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { EcsInstance, PageResult } from '~/types'
import { EcsStateNames, EcsStateType } from '~/types'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const dialog = useDialog()
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
const failedCount = computed(() => rows.value.filter((r) => r.observed_state === 'failed' || r.observed_state === 'exited').length)
const totalCpu = computed(() => rows.value.reduce((s, r) => s + (r.cpu || 0), 0))
const totalMem = computed(() => rows.value.reduce((s, r) => s + (r.memory_mb || 0), 0) / 1024)

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({ page: String(page.value), size: String(size) })
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
  if (op === 'delete') {
    dialog.warning({
      title: '删除容器',
      content: `确定删除容器「${row.name}」（${row.instance_no}）吗？操作不可撤销。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await api.del(`/ecs/${row.id}`)
          message.success('删除成功')
          load()
        } catch (e) {
          message.error(e instanceof Error ? e.message : '删除失败')
          return false
        }
      },
    })
    return
  }
  try {
    await api.post(`/ecs/${row.id}/${op}`)
    message.success('操作已提交')
    setTimeout(load, 500)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

function fmtPorts(ports: { host_port: number; container_port: number; protocol: string }[]): string {
  if (!ports || ports.length === 0) return '—'
  return ports.map((p) => `${p.host_port}:${p.container_port}/${p.protocol}`).join(', ')
}

function shortId(id: string): string {
  if (!id) return '—'
  return id.length > 12 ? id.slice(0, 12) : id
}

const columns: DataTableColumns<EcsInstance> = [
  { title: '容器名', key: 'name', minWidth: 140, render: (r) => r.name || r.container_name || '—' },
  { title: '容器 ID', key: 'container_id', width: 130, render: (r) => shortId(r.container_id) },
  { title: '镜像', key: 'image', minWidth: 160, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'observed_state',
    width: 100,
    render: (r) => h(NTag, { type: EcsStateType[r.observed_state] || 'default', size: 'small', round: true }, { default: () => EcsStateNames[r.observed_state] || r.observed_state }),
  },
  { title: '端口映射', key: 'ports', minWidth: 160, ellipsis: { tooltip: true }, render: (r) => fmtPorts(r.ports) },
  { title: 'IP', key: 'fixed_ip', width: 130, render: (r) => r.fixed_ip || '—' },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    fixed: 'right',
    render: (r) =>
      h('div', { class: 'flex gap-1' }, [
        h(NButton, { size: 'tiny', type: 'primary', tertiary: true, disabled: r.observed_state === 'running', onClick: () => act(r, 'start') }, { default: () => '启动' }),
        h(NButton, { size: 'tiny', type: 'warning', tertiary: true, disabled: r.observed_state !== 'running', onClick: () => act(r, 'stop') }, { default: () => '停止' }),
        h(NButton, { size: 'tiny', type: 'info', tertiary: true, disabled: r.observed_state !== 'running', onClick: () => act(r, 'restart') }, { default: () => '重启' }),
        h(NButton, { size: 'tiny', type: 'error', tertiary: true, onClick: () => act(r, 'delete') }, { default: () => '删除' }),
      ]),
  },
]

onMounted(() => {
  load()
  timer = window.setInterval(load, 15000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="containers" title="容器实例"
      description="Docker 容器实例管理 · 启停 / 重启 / 删除 / 端口映射 · 15s 自动刷新"
      :gradient="'linear-gradient(120deg, #0f766e 0%, #14b8a6 50%, #22d3ee 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ runningCount }}</span><span class="lbl">运行中</span></div>
        <div class="hero-pill"><span class="num">{{ stoppedCount }}</span><span class="lbl">已停止</span></div>
        <div class="hero-pill"><span class="num">{{ total }}</span><span class="lbl">总实例</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <n-input v-model:value="keyword" placeholder="搜索容器名 / 镜像" clearable size="small" style="width: 200px; --n-border-radius: 4px" @keyup.enter="load" @clear="load">
            <template #prefix><DxIcon name="search" :size="14" /></template>
          </n-input>
          <n-select v-model:value="status" :options="[
            { label: '全部状态', value: null },
            { label: '运行中', value: 'running' },
            { label: '已停止', value: 'stopped' },
            { label: '失败', value: 'failed' },
          ]" size="small" style="width: 120px" @update:value="load" />
          <button class="hero-btn" @click="load">
            <DxIcon name="refresh" :size="14" /> 刷新
          </button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
      <StatTile icon="play" label="运行中容器" :value="runningCount" suffix=" 个" color="#00b42a" hint="observed_state = running" />
      <StatTile icon="server" label="已停止 / 失败" :value="stoppedCount + failedCount" suffix=" 个" color="#fa8c16" hint="可重新启动或删除" />
      <StatTile icon="cpu" label="CPU 合计" :value="totalCpu" :decimals="2" suffix=" 核" color="#006eff" hint="当前页实例规格总和" />
      <StatTile icon="memory" label="内存合计" :value="totalMem" :decimals="1" suffix=" GB" color="#722ed1" hint="当前页实例规格总和" />
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_280px] gap-4 items-start">
      <n-data-table
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="{ page, pageSize: size, itemCount: total, showSizePicker: false, prefix: ({ itemCount }) => `共 ${itemCount} 条` }"
        :bordered="false"
        remote
        class="dx-card dx-table dx-fade-up dx-delay-1 overflow-hidden"
        @update:page="(p: number) => { page = p; load() }"
      />
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
              { value: failedCount, color: '#f53f3f', label: '失败' },
            ].filter(s => s.value > 0)"
          />
          <n-empty v-else description="暂无容器" class="py-8" />
        </div>
      </div>
    </div>
  </div>
</template>

