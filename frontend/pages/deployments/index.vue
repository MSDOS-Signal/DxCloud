<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Application, Deployment } from '~/types'
import { DeployStatusNames, DeployStatusType } from '~/types'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'
import SparkLine from '~/components/SparkLine.vue'

const loading = ref(false)
const rows = ref<(Deployment & { app_name?: string })[]>([])
const apps = ref<Application[]>([])

async function load() {
  loading.value = true
  try {
    // 获取所有应用
    apps.value = await api.get<Application[]>('/applications')
    // 并发获取每个应用的部署记录
    const allDeployments: (Deployment & { app_name?: string })[] = []
    await Promise.all(
      apps.value.map(async (app) => {
        try {
          const deps = await api.get<Deployment[]>(`/applications/${app.id}/deployments`)
          deps.forEach((d) => {
            allDeployments.push({ ...d, app_name: app.name })
          })
        } catch {
          // 忽略单个应用的错误
        }
      }),
    )
    // 按创建时间倒序
    allDeployments.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    rows.value = allDeployments
  } catch {
    // 忽略
  } finally {
    loading.value = false
  }
}

const successCount = computed(() => rows.value.filter((r) => r.status === 'deployed' || r.status === 'success').length)
const failedCount = computed(() => rows.value.filter((r) => r.status === 'failed').length)
const runningCount = computed(() => rows.value.filter((r) => r.status === 'deploying' || r.status === 'pending').length)
const successRate = computed(() => {
  const done = successCount.value + failedCount.value
  return done > 0 ? (successCount.value / done) * 100 : 100
})

// 最近 14 次部署的成功趋势（成功=1，失败=0，滚动累计）
const trendValues = computed(() => {
  const recent = rows.value.slice(0, 14).reverse()
  let acc = 0
  return recent.map((r, i) => {
    acc += r.status === 'failed' ? 0 : 1
    return acc / (i + 1) * 100
  })
})

function fmtTime(v: string | null): string {
  if (!v) return '—'
  return v.replace('T', ' ').slice(0, 19)
}

const columns: DataTableColumns<Deployment & { app_name?: string }> = [
  { title: '应用', key: 'app_name', width: 140, render: (r) => r.app_name || '—' },
  { title: '版本', key: 'version', width: 120 },
  { title: '镜像', key: 'image_ref', minWidth: 200, ellipsis: { tooltip: true } },
  { title: '策略', key: 'strategy', width: 100, render: (r) => r.strategy === 'blue-green' ? '蓝绿' : r.strategy === 'rolling' ? '滚动' : r.strategy || '—' },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (r) => h(NTag, { type: DeployStatusType[r.status] || 'default', size: 'small', round: true }, { default: () => DeployStatusNames[r.status] || r.status }),
  },
  {
    title: '健康',
    key: 'health_status',
    width: 90,
    render: (r) => {
      if (!r.health_status) return '—'
      const healthy = r.health_status === 'healthy'
      return h(NTag, { type: healthy ? 'success' : 'error', size: 'small', round: true }, { default: () => healthy ? '健康' : '异常' })
    },
  },
  { title: '触发', key: 'trigger', width: 90 },
  { title: '开始时间', key: 'started_at', width: 160, render: (r) => fmtTime(r.started_at) },
  { title: '完成时间', key: 'finished_at', width: 160, render: (r) => fmtTime(r.finished_at) },
]

onMounted(() => load())
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="deployments" title="部署记录"
      description="所有应用的部署历史 · 蓝绿 / 滚动更新 · 健康检查结果"
      :gradient="'linear-gradient(120deg, #531dab 0%, #722ed1 45%, #9254de 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">部署总数</span></div>
        <div class="hero-pill"><span class="num">{{ successRate.toFixed(0) }}%</span><span class="lbl">成功率</span></div>
        <div class="hero-pill"><span class="num">{{ apps.length }}</span><span class="lbl">应用数</span></div>
      </template>
      <template #action>
        <button class="hero-btn" @click="load">
          <DxIcon name="refresh" :size="14" /> 刷新
        </button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
      <StatTile icon="check-circle" label="部署成功" :value="successCount" suffix=" 次" color="#00b42a" hint="含蓝绿切换完成" />
      <StatTile icon="alert-triangle" label="部署失败" :value="failedCount" suffix=" 次" color="#f53f3f" hint="查看应用详情定位原因" />
      <StatTile icon="activity" label="进行中" :value="runningCount" suffix=" 次" color="#fa8c16" hint="部署 / 待健康检查" />
      <div class="stat-tile-static dx-fade-up">
        <div class="flex items-center justify-between w-full">
          <div>
            <div class="stat-label-2">近期成功率趋势</div>
            <div class="text-[22px] font-bold mt-0.5" style="color:#722ed1; font-variant-numeric: tabular-nums;">{{ successRate.toFixed(1) }}%</div>
          </div>
          <SparkLine v-if="trendValues.length > 1" :values="trendValues" color="#722ed1" :width="150" :height="48" />
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
      <n-data-table
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="{ pageSize: 20 }"
        :bordered="false"
        class="dx-card dx-table dx-fade-up dx-delay-1 overflow-hidden"
      />
      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">状态分布</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="rows.length > 0"
              :size="130"
              :center-text="successRate.toFixed(0) + '%'"
              center-label="成功率"
              :segments="[
                { value: successCount, color: '#00b42a', label: '成功' },
                { value: failedCount, color: '#f53f3f', label: '失败' },
                { value: runningCount, color: '#fa8c16', label: '进行中' },
              ].filter(s => s.value > 0)"
            />
            <n-empty v-else description="暂无部署记录" class="py-8" />
          </div>
        </div>
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">说明</span>
          </div>
          <div class="dx-card-body text-xs leading-relaxed text-gray-500 dark:text-gray-400 space-y-1.5">
            <p>· 蓝绿策略：新版本容器健康检查通过后自动切流，旧容器保留可回滚</p>
            <p>· 记录来自全部应用，按开始时间倒序</p>
            <p>· 失败记录可在「应用详情 → 部署历史」查看具体原因</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stat-tile-static {
  padding: 14px 16px;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.stat-tile-static:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(0, 20, 60, 0.08);
}
html.dark .stat-tile-static {
  background: #161b22;
  border-color: #30363d;
}
.stat-label-2 {
  font-size: 12px;
  color: #86909c;
}
</style>
