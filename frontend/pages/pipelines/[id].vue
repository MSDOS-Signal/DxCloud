<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Pipeline, PipelineRun, WebhookCreated, WebhookItem } from '~/types'
import { PipeStatusNames, PipeStatusType } from '~/types'
import { api } from '~/services/http'

const route = useRoute()
const message = useMessage()
const id = computed(() => Number(route.params.id))

const pipe = ref<Pipeline | null>(null)
const runs = ref<PipelineRun[]>([])
const webhooks = ref<WebhookItem[]>([])
const editing = ref(false)
const defDraft = ref('')
const nameDraft = ref('')

// ---------- webhook ----------
const showWh = ref(false)
const whModel = reactive({ provider: 'github', branch_filter: '', secret: '' })
const whCreated = ref<WebhookCreated | null>(null)

async function loadWebhooks() {
  try {
    webhooks.value = await api.get<WebhookItem[]>(`/webhooks?pipeline_id=${id.value}`)
  } catch {
    webhooks.value = []
  }
}

async function handleCreateWebhook() {
  try {
    whCreated.value = await api.post<WebhookCreated>('/webhooks', {
      pipeline_id: id.value,
      provider: whModel.provider,
      branch_filter: whModel.branch_filter,
      secret: whModel.secret,
    })
    message.success('Webhook 创建成功（Secret 仅此一次展示，请保存）')
    loadWebhooks()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleDeleteWebhook(w: WebhookItem) {
  try {
    await api.del(`/webhooks/${w.id}`)
    message.success('已删除')
    loadWebhooks()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

async function load() {
  try {
    pipe.value = await api.get<Pipeline>(`/pipelines/${id.value}`)
    defDraft.value = pipe.value.definition
    nameDraft.value = pipe.value.name
    runs.value = await api.get<PipelineRun[]>(`/pipeline-runs?pipeline_id=${id.value}&limit=30`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

async function handleRun() {
  try {
    const run = await api.post<{ id: number }>(`/pipelines/${id.value}/run`, {})
    message.success(`已触发运行 #${run.id}`)
    navigateTo(`/pipeline-runs/${run.id}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '触发失败')
  }
}

async function handleSave() {
  try {
    await api.put(`/pipelines/${id.value}`, { name: nameDraft.value, description: pipe.value?.description || '', definition: defDraft.value })
    message.success('已保存（下次运行生效）')
    editing.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '保存失败')
  }
}

const columns: DataTableColumns<PipelineRun> = [
  { title: '运行', key: 'run_no', width: 70, render: (row) => `#${row.run_no}` },
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: PipeStatusType[row.status] || 'default' }, { default: () => PipeStatusNames[row.status] || row.status }),
  },
  { title: '触发', key: 'trigger', width: 80 },
  { title: '分支', key: 'ref', width: 100, render: (row) => row.ref || '—' },
  { title: '耗时', key: 'duration_ms', width: 100, render: (row) => (row.duration_ms ? `${(row.duration_ms / 1000).toFixed(1)}s` : '—') },
  { title: '开始时间', key: 'started_at', width: 165, render: (row) => row.started_at || '—' },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) =>
      h(NButton, { size: 'tiny', ghost: true, type: 'primary', onClick: () => navigateTo(`/pipeline-runs/${row.id}`) }, { default: () => '详情' }),
  },
]

const whColumns: DataTableColumns<WebhookItem> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: 'Provider', key: 'provider', width: 90 },
  { title: '分支过滤', key: 'branch_filter', width: 120, render: (row) => row.branch_filter || '全部' },
  { title: 'URL', key: 'hook_code', render: (row) => `/api/v1/webhooks/${row.provider}/${row.hook_code}` },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    render: (row) => h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDeleteWebhook(row) }, { default: () => '删除' }),
  },
]

onMounted(() => {
  load()
  loadWebhooks()
})
</script>

<template>
  <div v-if="pipe" class="space-y-4">
    <div class="dx-card dx-fade-up">
      <div class="dx-card-body">
        <div class="flex items-center justify-between mb-3">
          <div>
            <span class="text-lg font-semibold">{{ pipe.name }}</span>
            <span class="text-xs text-gray-400 ml-3">{{ pipe.description }}</span>
          </div>
          <n-space>
            <n-button size="small" @click="editing = !editing">{{ editing ? '取消编辑' : '编辑定义' }}</n-button>
            <n-button size="small" @click="showWh = !showWh">Webhook</n-button>
            <n-button size="small" type="success" @click="handleRun">▶ 运行</n-button>
          </n-space>
        </div>
        <pre v-if="!editing" class="text-xs font-mono bg-slate-900 text-slate-100 p-3 rounded overflow-auto max-h-80">{{ pipe.definition }}</pre>
        <div v-else class="space-y-2">
          <n-input v-model:value="nameDraft" placeholder="名称" />
          <n-input v-model:value="defDraft" type="textarea" :autosize="{ minRows: 12, maxRows: 24 }" class="font-mono" />
          <n-button size="small" type="primary" @click="handleSave">保存</n-button>
        </div>
      </div>
    </div>

    <div v-if="showWh" class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">Git Webhook（push 自动触发）</div>
      <div class="dx-card-body">
        <n-space class="mb-3">
          <n-select v-model:value="whModel.provider" :options="[{label:'GitHub',value:'github'},{label:'GitLab',value:'gitlab'},{label:'Gitee',value:'gitee'}]" style="width: 110px" />
          <n-input v-model:value="whModel.branch_filter" placeholder="分支过滤（如 main 或 release/*，空=全部）" style="width: 240px" />
          <n-input v-model:value="whModel.secret" placeholder="Secret（留空自动生成）" style="width: 220px" />
          <n-button type="primary" size="small" @click="handleCreateWebhook">创建</n-button>
        </n-space>
        <n-alert v-if="whCreated" type="success" class="mb-3" :show-icon="false">
          <div class="text-xs">
            URL：<span class="font-mono">http://localhost/api/v1/webhooks/{{ whCreated.provider }}/{{ whCreated.hook_code }}</span><br />
            Secret：<span class="font-mono">{{ whCreated.secret }}</span>（仅展示一次，配置到 Git 平台 Webhook 中）
          </div>
        </n-alert>
        <n-data-table :columns="whColumns" :data="webhooks" size="small" :bordered="false" class="dx-table" />
      </div>
    </div>

    <div class="dx-card dx-fade-up dx-delay-2">
      <div class="dx-card-header">运行历史</div>
      <div class="dx-card-body">
        <n-data-table :columns="columns" :data="runs" size="small" :bordered="false" class="dx-table" />
      </div>
    </div>
  </div>
  <n-skeleton v-else :repeat="3" text />
</template>
