<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Application, Project } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const route = useRoute()
const rows = ref<Application[]>([])
const projects = ref<Project[]>([])
const loading = ref(false)
const projectId = ref<number | null>(route.query.project_id ? Number(route.query.project_id) : null)

const runtimeCount = computed(() => rows.value.filter((r) => r.type === 'custom').length)
const domainCount = computed(() => rows.value.filter((r) => r.domain).length)
const typeDist = computed(() => {
  const m = new Map<string, number>()
  for (const r of rows.value) m.set(r.type, (m.get(r.type) || 0) + 1)
  return [...m.entries()].map(([type, n]) => ({ type, n }))
})

async function load() {
  loading.value = true
  try {
    const q = projectId.value ? `?project_id=${projectId.value}` : ''
    rows.value = await api.get<Application[]>(`/applications${q}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadProjects() {
  try {
    projects.value = await api.get<Project[]>('/projects')
  } catch {
    // 忽略
  }
}

const runtimePresets = [
  { label: '自定义（镜像直接部署）', value: 'custom' },
  { label: 'Node.js', value: 'node' },
  { label: 'Go', value: 'go' },
  { label: 'Java', value: 'java' },
  { label: 'Python', value: 'python' },
  { label: 'Nginx', value: 'nginx' },
  { label: 'MySQL', value: 'mysql' },
  { label: 'Redis', value: 'redis' },
  { label: 'PostgreSQL', value: 'postgres' },
]

const showCreate = ref(false)
const createModel = reactive({
  project_id: null as number | null,
  name: '',
  type: 'custom',
  image: '',
  port: 80,
  health_check_path: '',
  domain: '',
})

async function handleCreate() {
  try {
    await api.post('/applications', {
      ...createModel,
      project_id: createModel.project_id || projectId.value || 0,
      env: [],
    })
    message.success('应用创建成功')
    showCreate.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleDelete(row: Application) {
  try {
    await api.del(`/applications/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<Application> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 140 },
  {
    title: '类型',
    key: 'type',
    width: 100,
    render: (row) => h(NTag, { size: 'small', bordered: false }, { default: () => row.type }),
  },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true } },
  { title: '端口', key: 'port', width: 70 },
  {
    title: '域名',
    key: 'domain',
    width: 180,
    render: (row) => row.domain || h('span', { class: 'text-slate-400' }, '—（端口访问）'),
  },
  { title: '健康检查', key: 'health_check_path', width: 120, render: (row) => row.health_check_path || 'TCP' },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', type: 'primary', ghost: true, onClick: () => navigateTo(`/apps/${row.id}`) }, { default: () => '详情/部署' }),
          h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
        ],
      }),
  },
]

onMounted(() => {
  load()
  loadProjects()
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="apps" title="应用管理"
      description="PaaS 应用 · 蓝绿部署与域名路由 · 健康检查 · 部署历史与回滚"
      :gradient="'linear-gradient(120deg, #0958d9 0%, #1677ff 45%, #69b1ff 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">应用总数</span></div>
        <div class="hero-pill"><span class="num">{{ domainCount }}</span><span class="lbl">已绑域名</span></div>
        <div class="hero-pill"><span class="num">{{ projects.length }}</span><span class="lbl">项目数</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <n-select
            v-model:value="projectId"
            :options="[{ label: '全部项目', value: null }, ...projects.map(p => ({ label: p.name, value: p.id }))]"
            size="small"
            style="width: 150px"
            @update:value="load"
          />
          <button class="hero-btn" @click="showCreate = true">
            <DxIcon name="plus" :size="14" /> 创建应用
          </button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
      <!-- 列表卡片 -->
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">
          <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">应用列表</span>
          <span class="text-xs text-gray-400">进入详情可部署 / 回滚 / 查看部署历史</span>
        </div>
        <div class="dx-card-body">
          <n-data-table class="dx-table" :columns="columns" :data="rows" :loading="loading" :bordered="false" />
        </div>
      </div>

      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">运行时分布</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="typeDist.length > 0"
              :size="130"
              :center-text="String(rows.length)"
              center-label="应用总数"
              :segments="typeDist.map((t, i) => ({
                value: t.n,
                color: ['#1677ff', '#13c2c2', '#722ed1', '#fa8c16', '#00b42a', '#eb2f96', '#a0d911'][i % 7],
                label: t.type,
              }))"
            />
            <n-empty v-else description="暂无应用" class="py-8" />
          </div>
        </div>
        <StatTile icon="domains" label="域名路由" :value="domainCount" :suffix="` / ${rows.length} 已绑定`" color="#00b42a" hint="Traefik Host 路由" />
      </div>
    </div>

    <!-- 创建应用弹窗 -->
    <n-modal v-model:show="showCreate" preset="card" title="创建应用" style="width: 560px">
      <n-form label-placement="left" label-width="110">
        <n-form-item label="项目">
          <n-select v-model:value="createModel.project_id" :options="projects.map(p => ({ label: p.name, value: p.id }))" placeholder="可选" />
        </n-form-item>
        <n-form-item label="名称" required>
          <n-input v-model:value="createModel.name" placeholder="如 api" />
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="createModel.type" :options="runtimePresets" />
        </n-form-item>
        <n-form-item label="镜像">
          <n-input v-model:value="createModel.image" placeholder="如 registry:2.8 / nginx:latest（部署时也可覆盖）" />
        </n-form-item>
        <n-form-item label="端口">
          <n-input-number v-model:value="createModel.port" :min="1" :max="65535" style="width: 120px" />
        </n-form-item>
        <n-form-item label="健康检查路径">
          <n-input v-model:value="createModel.health_check_path" placeholder="如 /healthz（留空 = TCP 探测）" />
        </n-form-item>
        <n-form-item label="域名">
          <n-input v-model:value="createModel.domain" placeholder="如 api.app.localhost（绑定后部署即走 Traefik 路由）" />
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
