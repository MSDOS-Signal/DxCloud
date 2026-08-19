<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Application, DomainItem } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const rows = ref<DomainItem[]>([])
const apps = ref<Application[]>([])
const loading = ref(false)

const tlsCount = computed(() => rows.value.filter((r) => r.tls).length)
const appName = (id: number) => apps.value.find((a) => a.id === id)?.name || `应用 ${id}`

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<DomainItem[]>('/domains')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

const showBind = ref(false)
const bindModel = reactive({ domain: '', application_id: null as number | null, target_port: 80 })

async function handleBind() {
  try {
    await api.post('/domains', {
      domain: bindModel.domain.trim(),
      application_id: bindModel.application_id || 0,
      target_port: bindModel.target_port || 80,
    })
    message.success('绑定成功')
    showBind.value = false
    bindModel.domain = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '绑定失败')
  }
}

async function handleUnbind(row: DomainItem) {
  try {
    await api.del(`/domains/${row.id}`)
    message.success('已解绑')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '解绑失败')
  }
}

const columns: DataTableColumns<DomainItem> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '域名', key: 'domain', width: 240 },
  { title: '应用', key: 'application_id', width: 140, render: (row) => (row.application_id ? appName(row.application_id) : '—') },
  { title: '端口', key: 'target_port', width: 80 },
  {
    title: 'TLS',
    key: 'tls',
    width: 70,
    render: (row) => h(NTag, { size: 'small', bordered: false, type: row.tls ? 'success' : 'default' }, { default: () => (row.tls ? 'HTTPS' : 'HTTP') }),
  },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleUnbind(row) }, { default: () => '解绑' }),
  },
]

onMounted(async () => {
  load()
  try {
    apps.value = await api.get<Application[]>('/applications')
  } catch {
    // 忽略
  }
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="domains" title="域名管理"
      description="Traefik Host 路由到应用容器 · 测试可用 *.localhost（自动解析 127.0.0.1）· 绑定后部署自动走域名"
      :gradient="'linear-gradient(120deg, #135200 0%, #237804 45%, #52c41a 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">绑定域名</span></div>
        <div class="hero-pill"><span class="num">{{ tlsCount }}</span><span class="lbl">HTTPS</span></div>
        <div class="hero-pill"><span class="num">{{ apps.length }}</span><span class="lbl">可绑应用</span></div>
      </template>
      <template #action>
        <button class="hero-btn" @click="showBind = true">
          <DxIcon name="plus" :size="14" /> 绑定域名
        </button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="domains" label="域名总数" :value="rows.length" suffix=" 个" color="#237804" hint="Traefik Host 规则" />
      <StatTile icon="security" label="TLS / HTTPS" :value="tlsCount" suffix=" 个" color="#13c2c2" hint="Traefik 自动证书" />
      <StatTile icon="apps" label="可绑定应用" :value="apps.length" suffix=" 个" color="#722ed1" hint="应用中心内全部应用" />
    </div>

    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">域名列表</span>
        <span class="text-xs text-gray-400">浏览器直接访问 *.localhost 即可验证路由</span>
      </div>
      <div class="dx-card-body">
        <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false" class="dx-table" />
      </div>
    </div>

    <n-modal v-model:show="showBind" preset="card" title="绑定域名" style="width: 480px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="域名" required>
          <n-input v-model:value="bindModel.domain" placeholder="api.app.localhost" />
        </n-form-item>
        <n-form-item label="应用">
          <n-select v-model:value="bindModel.application_id" :options="apps.map(a => ({ label: a.name, value: a.id }))" placeholder="可选（绑定后部署走域名路由）" />
        </n-form-item>
        <n-form-item label="端口">
          <n-input-number v-model:value="bindModel.target_port" :min="1" :max="65535" style="width: 120px" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showBind = false">取消</n-button>
          <n-button type="primary" @click="handleBind">绑定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
