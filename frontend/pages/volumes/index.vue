<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { CloudVolume } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const rows = ref<CloudVolume[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<CloudVolume[]>('/volumes')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

const totalCap = computed(() => rows.value.reduce((s, v) => s + v.capacity_gb, 0))
const totalUsed = computed(() => rows.value.reduce((s, v) => s + (v.used_mb || 0) / 1024, 0))
const usagePct = computed(() => (totalCap.value > 0 ? Math.min(100, (totalUsed.value / totalCap.value) * 100) : 0))

const showCreate = ref(false)
const createModel = reactive({ name: '', capacity_gb: 10 })

async function handleCreate() {
  try {
    await api.post('/volumes', createModel)
    message.success('云磁盘创建成功')
    showCreate.value = false
    createModel.name = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleDelete(row: CloudVolume) {
  try {
    await api.del(`/volumes/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败（可能正被实例挂载）')
  }
}

const columns: DataTableColumns<CloudVolume> = [
  {
    title: '名称',
    key: 'name',
    width: 180,
    render: (row) =>
      h('span', { class: 'font-medium' }, [
        h('span', { class: 'inline-block w-1.5 h-1.5 rounded-full bg-[#006eff] mr-2 align-middle' }),
        row.name,
      ]),
  },
  {
    title: '容量使用',
    key: 'capacity_gb',
    minWidth: 220,
    render: (row) => {
      const used = (row.used_mb || 0) / 1024
      const pct = row.capacity_gb > 0 ? Math.min(100, (used / row.capacity_gb) * 100) : 0
      return h('div', { class: 'flex items-center gap-2' }, [
        h('div', { class: 'flex-1 h-1.5 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden min-w-[90px]' }, [
          h('div', {
            class: 'h-full rounded-full transition-all duration-700',
            style: `width:${pct}%;background:linear-gradient(90deg,#006eff,#00c2ff)`,
          }),
        ]),
        h('span', { class: 'text-xs text-gray-500 tabular-nums whitespace-nowrap' }, `${used.toFixed(1)} / ${row.capacity_gb} GB`),
      ])
    },
  },
  { title: '驱动', key: 'driver', width: 90 },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' })],
      }),
  },
]

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="volumes" title="云磁盘"
      description="Docker named volume 抽象 · 容量软配额 · 变更挂载会重建容器但数据保留 · 可在 ECS 创建/详情页挂载"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">磁盘总数</span></div>
        <div class="hero-pill"><span class="num">{{ totalCap }}</span><span class="lbl">总容量 GB</span></div>
        <div class="hero-pill"><span class="num">{{ usagePct.toFixed(1) }}%</span><span class="lbl">整体使用率</span></div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="volumes" label="磁盘数量" :value="rows.length" suffix=" 块" color="#006eff" hint="当前空间可见的全部卷" class="md:col-span-1" />
      <StatTile icon="database" label="总容量配额" :value="totalCap" suffix=" GB" color="#13c2c2" hint="所有卷 capacity_gb 之和" />
      <StatTile icon="activity" label="已用容量" :value="totalUsed" :decimals="1" suffix=" GB" color="#722ed1" :hint="'使用率 ' + usagePct.toFixed(1) + '%'" />
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">
          <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">磁盘列表</span>
          <button class="dx-btn-primary" @click="showCreate = true">
            <DxIcon name="plus" :size="13" /> 创建云磁盘
          </button>
        </div>
        <div class="dx-card-body">
          <n-data-table class="dx-table" :columns="columns" :data="rows" :loading="loading" :bordered="false" />
        </div>
      </div>

      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">
          <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">容量分布</span>
        </div>
        <div class="dx-card-body flex items-center justify-center py-2">
          <DonutChart
            v-if="rows.length > 0"
            :size="130"
            :center-text="usagePct.toFixed(0) + '%'"
            center-label="整体使用率"
            :segments="rows.slice(0, 6).map((v, i) => ({
              value: v.capacity_gb,
              color: ['#006eff', '#13c2c2', '#722ed1', '#fa8c16', '#00b42a', '#f53f3f'][i % 6],
              label: v.name,
            }))"
          />
          <n-empty v-else description="暂无磁盘" class="py-8" />
        </div>
      </div>
    </div>

    <!-- 创建云磁盘弹窗 -->
    <n-modal v-model:show="showCreate" preset="card" title="创建云磁盘" style="width: 440px">
      <n-form label-placement="left" label-width="100">
        <n-form-item label="名称" required>
          <n-input v-model:value="createModel.name" placeholder="如 mysql-data" />
        </n-form-item>
        <n-form-item label="容量（GB）">
          <n-input-number v-model:value="createModel.capacity_gb" :min="1" :max="500" style="width: 100%" />
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

