<script setup lang="ts">
import { h, onMounted, ref, watch } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { BillingSummary, ResourceUsage } from '~/types'
import { UsageTypeNames } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const auth = useAuthStore()
const orgStore = useOrgStore()

const loading = ref(false)
const summary = ref<BillingSummary | null>(null)
const records = ref<ResourceUsage[]>([])
const recLoading = ref(false)

function orgQuery(): string {
  return orgStore.currentOrgId > 0 ? `?org_id=${orgStore.currentOrgId}` : ''
}

async function load() {
  loading.value = true
  try {
    summary.value = await api.get<BillingSummary>(`/billing${orgQuery()}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '账单加载失败')
    summary.value = null
  } finally {
    loading.value = false
  }
  await loadRecords()
}

async function loadRecords() {
  recLoading.value = true
  try {
    records.value = await api.get<ResourceUsage[]>(`/billing/records${orgQuery()}`)
  } catch {
    records.value = []
  } finally {
    recLoading.value = false
  }
}

// ---------- 手动结算（运维/测试） ----------
const ticking = ref(false)
async function handleTick() {
  ticking.value = true
  try {
    await api.post('/billing/tick')
    message.success('已按当前运行中实例结算本小时用量')
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '结算失败')
  } finally {
    ticking.value = false
  }
}

// ---------- 充值 ----------
const showRecharge = ref(false)
const rechargeModel = reactive({ org_id: 0, amount: 100 })

function openRecharge() {
  rechargeModel.org_id = orgStore.currentOrgId
  showRecharge.value = true
}

async function handleRecharge() {
  if (rechargeModel.org_id <= 0) {
    message.warning('请先选择组织（充值按组织入账）')
    return
  }
  if (rechargeModel.amount <= 0) {
    message.warning('充值金额需大于 0')
    return
  }
  try {
    await api.post('/billing/recharge', rechargeModel)
    message.success('充值成功')
    showRecharge.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '充值失败')
  }
}

const usageItems = computed(() => {
  const m = summary.value?.usage_month || {}
  return (Object.keys(UsageTypeNames) as (keyof typeof UsageTypeNames)[]).map((k) => ({
    key: k,
    name: UsageTypeNames[k],
    value: Number(m[k] || 0),
    price: Number(summary.value?.price?.[k] || 0),
  }))
})

const totalCost = computed(() =>
  usageItems.value.reduce((s, u) => s + u.value * u.price, 0),
)

const columns: DataTableColumns<ResourceUsage> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: '类型',
    key: 'resource_type',
    width: 140,
    render: (row) => h(NTag, { size: 'small' }, { default: () => UsageTypeNames[row.resource_type] || row.resource_type }),
  },
  { title: '用量', key: 'used_value', width: 100, render: (row) => row.used_value.toFixed(2) },
  { title: '所属小时', key: 'period', minWidth: 160 },
  { title: '组织 ID', key: 'org_id', width: 80 },
]

watch(() => orgStore.currentOrgId, () => load())
onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="billing" title="计费中心"
      description="虚拟计费演示 · 单价：CPU ¥0.1/核时 · 内存 ¥0.05/GB·时 · 磁盘 ¥0.01/GB·时"
      :gradient="'linear-gradient(120deg, #5b3a1e 0%, #d48806 45%, #ffc53d 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">¥{{ (summary?.credit || 0).toFixed(2) }}</span><span class="lbl">虚拟余额</span></div>
        <div class="hero-pill"><span class="num">¥{{ totalCost.toFixed(2) }}</span><span class="lbl">本月费用</span></div>
        <div class="hero-pill"><span class="num">{{ records.length }}</span><span class="lbl">流水条数</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <button v-if="auth.hasPerm('quota:update')" class="hero-btn" @click="handleTick">
            <DxIcon name="refresh" :size="14" /> {{ ticking ? '结算中...' : '手动结算' }}
          </button>
          <button v-if="auth.hasPerm('quota:update')" class="hero-btn" style="font-weight:600" @click="openRecharge">
            <DxIcon name="plus" :size="14" /> 充值
          </button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="billing" label="组织虚拟余额" :value="summary?.credit ?? 0" :decimals="2" prefix="¥ " color="#d48806" :hint="(summary?.credit || 0) > 0 ? '可用' : '已透支'" />
      <StatTile icon="activity" label="本月累计费用" :value="totalCost" :decimals="2" prefix="¥ " color="#fa541c" hint="按小时结算累计" />
      <StatTile icon="database" label="计费资源种类" :value="usageItems.filter((u) => u.value > 0).length" suffix=" 种" color="#13c2c2" hint="CPU / 内存 / 磁盘" />
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_320px] gap-4 items-start">
      <n-card :bordered="false" title="本月用量明细" class="dx-card dx-fade-up dx-delay-1">
        <n-empty v-if="usageItems.every((u) => u.value <= 0)" description="本月暂无用量记录" class="py-6" />
        <div v-else class="space-y-2">
          <div v-for="u in usageItems.filter((x) => x.value > 0)" :key="u.key" class="flex items-center justify-between py-2 border-b border-gray-100">
            <span class="text-sm">{{ u.name }}</span>
            <span class="text-sm text-gray-500">{{ u.value.toFixed(2) }} × ¥{{ u.price }}/单位 = <b class="text-gray-700">¥{{ (u.value * u.price).toFixed(2) }}</b></span>
          </div>
        </div>
      </n-card>

      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">费用构成</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="totalCost > 0"
              :size="130"
              :center-text="'¥' + totalCost.toFixed(2)"
              center-label="本月合计"
              :value-decimals="2"
              :segments="usageItems.filter((u) => u.value > 0).map((u, i) => ({
                value: Number((u.value * u.price).toFixed(2)),
                color: ['#d48806', '#13c2c2', '#722ed1'][i % 3],
                label: u.name,
              }))"
            />
            <n-empty v-else description="暂无费用" class="py-8" />
          </div>
        </div>
        <n-card :bordered="false" class="dx-card dx-fade-up dx-delay-2" title="账单流水（用量记录）">
          <n-data-table :columns="columns" :data="records" :loading="recLoading" :bordered="false" size="small" :max-height="360" class="dx-table" />
        </n-card>
      </div>
    </div>

    <!-- 充值 -->
    <n-modal v-model:show="showRecharge" preset="card" title="组织充值（虚拟）" style="width: 420px">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="组织 ID" required>
          <n-input-number v-model:value="rechargeModel.org_id" :min="1" style="width: 100%" placeholder="充值入账的组织" />
        </n-form-item>
        <n-form-item label="金额（¥）" required>
          <n-input-number v-model:value="rechargeModel.amount" :min="0.01" :max="1000000" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRecharge = false">取消</n-button>
          <n-button type="primary" @click="handleRecharge">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
