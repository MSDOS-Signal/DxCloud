<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'

const message = useMessage()
const tab = ref<'operation' | 'audit' | 'login'>('operation')
const rows = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = 20
const keyword = ref('')
const loading = ref(false)

const tabNames = { operation: '操作日志', audit: '审计日志', login: '登录日志' } as const

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({ type: tab.value, page: String(page.value), size: String(size) })
    if (keyword.value) q.set('keyword', keyword.value)
    const r = await api.get<{ total: number; items: any[] }>(`/logs?${q.toString()}`)
    rows.value = r.items
    total.value = r.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function fmtTime(v: string | null): string {
  if (!v) return '—'
  return v.replace('T', ' ').slice(0, 19)
}

const opColumns: DataTableColumns<any> = [
  { title: '时间', key: 'created_at', width: 170, render: (r) => fmtTime(r.created_at) },
  { title: '模块', key: 'module', width: 90 },
  { title: '动作', key: 'action', width: 130 },
  { title: '对象', key: 'target_name', width: 140, render: (r) => r.target_name || r.target_id },
  { title: '用户', key: 'user_id', width: 80 },
  { title: '结果', key: 'result', width: 80, render: (r) => h(NTag, { size: 'small', type: r.result === 1 ? 'success' : 'error' }, { default: () => (r.result === 1 ? '成功' : '失败') }) },
  { title: '耗时', key: 'duration_ms', width: 90, render: (r) => `${r.duration_ms}ms` },
  { title: 'IP', key: 'ip', width: 120 },
]

const auditColumns: DataTableColumns<any> = [
  { title: '时间', key: 'created_at', width: 170, render: (r) => fmtTime(r.created_at) },
  { title: '动作', key: 'action', width: 170 },
  { title: '资源', key: 'resource_type', width: 90 },
  { title: '资源ID', key: 'resource_id', width: 200, ellipsis: { tooltip: true } },
  { title: '用户', key: 'user_id', width: 80 },
  { title: '结果', key: 'status', width: 80, render: (r) => h(NTag, { size: 'small', type: r.status === 1 ? 'success' : 'error' }, { default: () => (r.status === 1 ? '成功' : '拒绝') }) },
  { title: 'IP', key: 'ip', width: 120 },
  { title: 'RequestID', key: 'request_id', width: 110 },
]

const loginColumns: DataTableColumns<any> = [
  { title: '时间', key: 'created_at', width: 170, render: (r) => fmtTime(r.created_at) },
  { title: '用户', key: 'user_id', width: 80 },
  { title: '结果', key: 'status', width: 80, render: (r) => h(NTag, { size: 'small', type: r.status === 1 ? 'success' : 'error' }, { default: () => (r.status === 1 ? '成功' : '失败') }) },
  { title: '说明', key: 'message', ellipsis: { tooltip: true } },
  { title: 'IP', key: 'ip', width: 130 },
  { title: 'UA', key: 'user_agent', ellipsis: { tooltip: true } },
]

const columns = computed(() => {
  if (tab.value === 'operation') return opColumns
  if (tab.value === 'audit') return auditColumns
  return loginColumns
})

function switchTab(t: 'operation' | 'audit' | 'login') {
  tab.value = t
  page.value = 1
  load()
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="logs" title="日志中心"
      description="操作日志（who did what）· 审计日志（RBAC 裁决）· 登录日志（认证成功/失败）"
      :gradient="'linear-gradient(120deg, #262626 0%, #434343 45%, #595959 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ total }}</span><span class="lbl">{{ tabNames[tab] }}条数</span></div>
        <div class="hero-pill"><span class="num">3</span><span class="lbl">日志类型</span></div>
      </template>
    </PageHero>

    <n-card class="dx-card dx-fade-up dx-delay-1">
      <n-tabs type="line" :value="tab" @update:value="(v: any) => switchTab(v)">
        <n-tab-pane name="operation" tab="操作日志" />
        <n-tab-pane name="audit" tab="审计日志" />
        <n-tab-pane name="login" tab="登录日志" />
      </n-tabs>
      <div class="flex items-center gap-2 my-3">
        <n-input v-model:value="keyword" placeholder="搜索动作/对象/IP" clearable style="width: 260px" @keyup.enter="page = 1; load()">
          <template #prefix><DxIcon name="search" :size="14" class="text-gray-400" /></template>
        </n-input>
        <n-button @click="page = 1; load()">搜索</n-button>
        <span class="text-xs text-gray-400">共 {{ total }} 条</span>
      </div>
      <n-data-table
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="{ page: page, pageSize: size, itemCount: total, onChange: (p: number) => { page = p; load() } }"
        :bordered="false"
        size="small"
        class="dx-table"
      />
    </n-card>
  </div>
</template>
