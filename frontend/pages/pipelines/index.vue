<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Pipeline } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const rows = ref<Pipeline[]>([])
const loading = ref(false)

const exampleYaml = `name: build-and-deploy
timeout: 2h
env:
  WHO: dxcloud
steps:
  - name: say-hello
    type: shell
    script: echo "hello from $WHO" && pwd && ls -la
  - name: maybe-fail
    type: shell
    script: echo "this step is allowed to fail"
    allow_failure: true
`

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<Pipeline[]>('/pipelines')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

const showCreate = ref(false)
const createModel = reactive({ name: '', description: '', definition: exampleYaml })

async function handleCreate() {
  try {
    await api.post('/pipelines', createModel)
    message.success('Pipeline 创建成功')
    showCreate.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleRun(row: Pipeline) {
  try {
    const run = await api.post<{ id: number }>(`/pipelines/${row.id}/run`, {})
    message.success(`已触发运行 #${run.id}`)
    navigateTo(`/pipeline-runs/${run.id}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '触发失败')
  }
}

async function handleDelete(row: Pipeline) {
  try {
    await api.del(`/pipelines/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<Pipeline> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 180 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', type: 'primary', ghost: true, onClick: () => navigateTo(`/pipelines/${row.id}`) }, { default: () => '详情/运行' }),
          h(NButton, { size: 'tiny', type: 'success', ghost: true, onClick: () => handleRun(row) }, { default: () => '运行' }),
          h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
        ],
      }),
  },
]

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="pipelines" title="CI/CD Pipeline"
      description="步骤在隔离 Job 容器中执行（CPU/内存/PID 限制、非特权、不挂 docker.sock）· Redis 队列调度 · 支持 git/shell/docker-build/push/deploy/wait-health"
      :gradient="'linear-gradient(120deg, #ad4e00 0%, #fa8c16 45%, #ffc069 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">流水线</span></div>
        <div class="hero-pill"><span class="num">6</span><span class="lbl">步骤类型</span></div>
      </template>
      <template #action>
        <button class="hero-btn" @click="showCreate = true">
          <DxIcon name="plus" :size="14" /> 创建 Pipeline
        </button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="pipelines" label="流水线数量" :value="rows.length" suffix=" 条" color="#fa8c16" hint="当前空间全部定义" />
      <StatTile icon="zap" label="触发方式" value="—" color="#722ed1" hint="手动 / Webhook / 定时（cron）" />
      <StatTile icon="security" label="隔离执行" value="—" color="#13c2c2" hint="Job 容器非特权 + 资源限额" />
    </div>

    <!-- 列表卡片 -->
    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">Pipeline 列表</span>
        <span class="text-xs text-gray-400">运行后自动跳转实时日志页</span>
      </div>
      <div class="dx-card-body">
        <n-data-table class="dx-table" :columns="columns" :data="rows" :loading="loading" :bordered="false" />
      </div>
    </div>

    <!-- 创建 Pipeline 弹窗 -->
    <n-modal v-model:show="showCreate" preset="card" title="创建 Pipeline" style="width: 720px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="名称" required>
          <n-input v-model:value="createModel.name" placeholder="如 build-and-deploy" />
        </n-form-item>
        <n-form-item label="描述">
          <n-input v-model:value="createModel.description" placeholder="可选" />
        </n-form-item>
        <n-form-item label="定义 YAML">
          <n-input v-model:value="createModel.definition" type="textarea" :autosize="{ minRows: 12, maxRows: 24 }" class="font-mono" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate">创建</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
