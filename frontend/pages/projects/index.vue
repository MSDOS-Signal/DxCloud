<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Project } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const rows = ref<Project[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<Project[]>('/projects')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

const showCreate = ref(false)
const createModel = reactive({ name: '', code: '', description: '' })

async function handleCreate() {
  try {
    await api.post('/projects', createModel)
    message.success('项目创建成功（已内置 4 个环境：development/testing/staging/production）')
    showCreate.value = false
    createModel.name = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleDelete(row: Project) {
  try {
    await api.del(`/projects/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<Project> = [
  {
    title: '名称',
    key: 'name',
    width: 180,
    render: (row) =>
      h('span', { class: 'font-medium' }, [
        h('span', { class: 'inline-block w-1.5 h-1.5 rounded-full bg-[#00b42a] mr-2 align-middle' }),
        row.name,
      ]),
  },
  { title: '代码', key: 'code', width: 120 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row) => h(NTag, { size: 'small', type: row.status === 1 ? 'success' : 'default' }, { default: () => (row.status === 1 ? '正常' : '停用') }),
  },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', ghost: true, type: 'primary', onClick: () => navigateTo(`/apps?project_id=${row.id}`) }, { default: () => '应用' }),
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
      icon="projects" title="项目管理"
      description="项目 = 应用分组 · 每个项目内置 development / testing / staging / production 四环境 · 应用与流水线按项目组织"
      :gradient="'linear-gradient(120deg, #062e7d 0%, #0a4bb0 45%, #4096ff 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">项目总数</span></div>
        <div class="hero-pill"><span class="num">{{ rows.length * 4 }}</span><span class="lbl">环境合计</span></div>
      </template>
      <template #action>
        <button class="hero-btn" @click="showCreate = true">
          <DxIcon name="plus" :size="14" /> 创建项目
        </button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="projects" label="项目数量" :value="rows.length" suffix=" 个" color="#0a4bb0" hint="当前空间可见" />
      <StatTile icon="activity" label="部署环境" :value="rows.length * 4" suffix=" 个" color="#13c2c2" hint="每项目 4 环境" />
      <StatTile icon="apps" label="下一步" value="—" color="#722ed1" hint="在应用页创建应用并部署" />
    </div>

    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">项目列表</span>
        <span class="text-xs text-gray-400">项目删除前需先清空应用</span>
      </div>
      <div class="dx-card-body">
        <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false" class="dx-table" />
      </div>
    </div>

    <n-modal v-model:show="showCreate" preset="card" title="创建项目" style="width: 480px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="名称" required>
          <n-input v-model:value="createModel.name" placeholder="如 shop" />
        </n-form-item>
        <n-form-item label="代码">
          <n-input v-model:value="createModel.code" placeholder="可选，如 shop" />
        </n-form-item>
        <n-form-item label="描述">
          <n-input v-model:value="createModel.description" placeholder="可选" />
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
