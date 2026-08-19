<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { CloudNetwork, EcsInstance, PageResult } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const rows = ref<CloudNetwork[]>([])
const loading = ref(false)

const internalCount = computed(() => rows.value.filter((r) => r.internal).length)
const stdCount = computed(() => rows.value.filter((r) => !r.internal).length)

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<CloudNetwork[]>('/networks')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

// 创建
const showCreate = ref(false)
const createModel = reactive({ name: '', subnet: '10.10.0.0/16', gateway: '', ip_range: '', internal: false })

async function handleCreate() {
  try {
    await api.post('/networks', createModel)
    message.success('网络创建成功')
    showCreate.value = false
    createModel.name = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

// 详情
const showDetail = ref(false)
const detail = ref<{ name: string; subnet: string; gateway: string; containers: Record<string, string> } | null>(null)

async function openDetail(row: CloudNetwork) {
  try {
    detail.value = await api.get(`/networks/${row.id}`)
    showDetail.value = true
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

// 连接容器
const connectNet = ref<CloudNetwork | null>(null)
const showConnect = ref(false)
const instances = ref<EcsInstance[]>([])
const connectModel = reactive({ instance_id: null as number | null, fixed_ip: '' })

async function openConnect(row: CloudNetwork) {
  connectNet.value = row
  connectModel.instance_id = null
  connectModel.fixed_ip = ''
  try {
    const res = await api.get<PageResult<EcsInstance>>('/ecs?size=100')
    instances.value = res.items.filter((i) => i.observed_state === 'running')
  } catch {
    instances.value = []
  }
  showConnect.value = true
}

async function handleConnect() {
  if (!connectNet.value || !connectModel.instance_id) return
  try {
    await api.post(`/networks/${connectNet.value.id}/connect`, {
      instance_id: connectModel.instance_id,
      fixed_ip: connectModel.fixed_ip || '',
    })
    message.success('已加入网络')
    showConnect.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '连接失败')
  }
}

async function handleDisconnect(row: CloudNetwork, containerName: string) {
  // 找到容器对应实例 id：简化——通过详情容器名匹配前端列表
  try {
    const res = await api.get<PageResult<EcsInstance>>('/ecs?size=100')
    const inst = res.items.find((i) => i.container_name === containerName)
    if (!inst) {
      message.error('未找到对应实例')
      return
    }
    await api.post(`/networks/${row.id}/disconnect`, { instance_id: inst.id })
    message.success('已退出网络')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

async function handleDelete(row: CloudNetwork) {
  try {
    await api.del(`/networks/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<CloudNetwork> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 140 },
  {
    title: '类型',
    key: 'internal',
    width: 90,
    render: (row) => h(NTag, { size: 'small', bordered: false }, { default: () => (row.internal ? '内部' : '标准') }),
  },
  { title: '子网', key: 'subnet', width: 140, render: (row) => row.subnet || '—' },
  { title: '网关', key: 'gateway', width: 130, render: (row) => row.gateway || '—' },
  { title: 'IP 段', key: 'ip_range', width: 130, render: (row) => row.ip_range || '—' },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', ghost: true, type: 'primary', onClick: () => openDetail(row) }, { default: () => '详情' }),
          h(NButton, { size: 'tiny', onClick: () => openConnect(row) }, { default: () => '连接容器' }),
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
      icon="networks" title="网络管理"
      description="每个网络 = 一个自定义子网的隔离 bridge（Docker IPAM）· 实例可分配静态 IP · 支持容器热接入"
      :gradient="'linear-gradient(120deg, #1d39c4 0%, #2f54eb 45%, #597ef7 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">网络总数</span></div>
        <div class="hero-pill"><span class="num">{{ stdCount }}</span><span class="lbl">标准网络</span></div>
        <div class="hero-pill"><span class="num">{{ internalCount }}</span><span class="lbl">内部网络</span></div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">
          <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">网络列表</span>
          <button class="dx-btn-primary" @click="showCreate = true">
            <DxIcon name="plus" :size="13" /> 创建网络
          </button>
        </div>
        <div class="dx-card-body">
          <n-data-table class="dx-table" :columns="columns" :data="rows" :loading="loading" :bordered="false" />
        </div>
      </div>

      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">类型分布</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="rows.length > 0"
              :size="130"
              :center-text="String(rows.length)"
              center-label="网络总数"
              :segments="[
                { value: stdCount, color: '#2f54eb', label: '标准 bridge' },
                { value: internalCount, color: '#13c2c2', label: '内部网络' },
              ].filter(s => s.value > 0)"
            />
            <n-empty v-else description="暂无网络" class="py-8" />
          </div>
        </div>
        <StatTile icon="globe" label="子网容量估算" :value="rows.length * 65534" suffix=" IP" color="#597ef7" hint="按 /16 子网粗略估算" />
      </div>
    </div>

    <!-- 创建网络弹窗 -->
    <n-modal v-model:show="showCreate" preset="card" title="创建网络" style="width: 500px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="名称" required>
          <n-input v-model:value="createModel.name" placeholder="如 shop-net" />
        </n-form-item>
        <n-form-item label="子网 CIDR">
          <n-input v-model:value="createModel.subnet" placeholder="如 10.10.0.0/16" />
        </n-form-item>
        <n-form-item label="网关">
          <n-input v-model:value="createModel.gateway" placeholder="如 10.10.0.1（可留空自动）" />
        </n-form-item>
        <n-form-item label="IP 段">
          <n-input v-model:value="createModel.ip_range" placeholder="如 10.10.0.0/24（可留空）" />
        </n-form-item>
        <n-form-item label="内部网络">
          <n-switch v-model:value="createModel.internal" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 网络详情弹窗 -->
    <n-modal v-model:show="showDetail" preset="card" :title="`网络详情 · ${detail?.name ?? ''}`" style="width: 560px">
      <n-descriptions :column="1" size="small" label-placement="left">
        <n-descriptions-item label="子网">{{ detail?.subnet || '—' }}</n-descriptions-item>
        <n-descriptions-item label="网关">{{ detail?.gateway || '—' }}</n-descriptions-item>
        <n-descriptions-item label="已连接容器">
          <div v-if="detail && Object.keys(detail.containers).length">
            <div v-for="(ip, name) in detail.containers" :key="name" class="flex justify-between text-xs py-0.5">
              <span class="font-mono">{{ name }}</span>
              <span class="font-mono text-gray-500 dark:text-gray-400">{{ ip }}</span>
            </div>
          </div>
          <span v-else class="text-gray-400">（无）</span>
        </n-descriptions-item>
      </n-descriptions>
    </n-modal>

    <!-- 容器加入网络弹窗 -->
    <n-modal v-model:show="showConnect" preset="card" :title="`容器加入网络 · ${connectNet?.name ?? ''}`" style="width: 500px">
      <n-form label-placement="left" label-width="110">
        <n-form-item label="实例">
          <n-select
            v-model:value="connectModel.instance_id"
            :options="instances.map((i) => ({ label: `${i.name}（${i.instance_no}）`, value: i.id }))"
            placeholder="选择运行中的实例"
            filterable
          />
        </n-form-item>
        <n-form-item label="静态 IP（可选）">
          <n-input v-model:value="connectModel.fixed_ip" :placeholder="`如 ${(connectNet?.subnet || '10.10.0.0/16').split('/')[0].split('.').slice(0, 3).join('.')}.10`" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showConnect = false">取消</n-button>
          <n-button type="primary" @click="handleConnect">加入网络</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
