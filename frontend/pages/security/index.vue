<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { SecretItem, SecurityDashboard, SecurityReportItem } from '~/types'
import { SeverityNames, SeverityType } from '~/types'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const auth = useAuthStore()
const orgStore = useOrgStore()

const loading = ref(false)
const scanning = ref(false)
const dash = ref<SecurityDashboard | null>(null)
const reports = ref<SecurityReportItem[]>([])

async function load() {
  loading.value = true
  try {
    dash.value = await api.get<SecurityDashboard>('/security/dashboard')
    reports.value = await api.get<SecurityReportItem[]>('/security/reports?limit=30')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleScan() {
  scanning.value = true
  try {
    await api.post('/security/scan')
    message.success('扫描完成，报告已生成')
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '扫描失败')
  } finally {
    scanning.value = false
  }
}

const scoreColor = computed(() => {
  const s = dash.value?.score ?? 100
  if (s >= 90) return '#18a058'
  if (s >= 70) return '#f0a020'
  return '#d03050'
})

const scoreSegments = computed(() => {
  const s = dash.value?.score ?? 0
  return [
    { value: s, color: scoreColor.value },
    { value: Math.max(0, 100 - s), color: '#eef1f5' },
  ]
})

const latestReports = computed(() => reports.value.slice(0, 5))
const scoreTrend = computed(() => {
  if (latestReports.value.length < 2) return null
  const cur = latestReports.value[0]?.score ?? 0
  const prev = latestReports.value[1]?.score ?? 0
  return prev === 0 ? null : ((cur - prev) / prev) * 100
})

// ---------- 密钥托管 ----------
const secrets = ref<SecretItem[]>([])
const showCreate = ref(false)
const createModel = reactive({ name: '', value: '' })
const revealMap = ref<Record<number, string>>({})

async function loadSecrets() {
  try {
    secrets.value = await api.get<SecretItem[]>('/secrets')
  } catch {
    secrets.value = []
  }
}

async function handleCreateSecret() {
  if (!createModel.name.trim() || !createModel.value) {
    message.warning('请填写密钥名与密钥值')
    return
  }
  try {
    await api.post('/secrets', createModel)
    message.success('密钥已加密保存')
    showCreate.value = false
    createModel.name = ''
    createModel.value = ''
    await loadSecrets()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleReveal(s: SecretItem) {
  try {
    const r = await api.get<{ id: number; value: string }>(`/secrets/${s.id}/reveal`)
    revealMap.value[s.id] = r.value
    message.success('已解密显示（仅本次）')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '读取失败')
  }
}

async function handleDeleteSecret(s: SecretItem) {
  try {
    await api.del(`/secrets/${s.id}`)
    message.success('已删除')
    await loadSecrets()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const secretColumns: DataTableColumns<SecretItem> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '密钥名', key: 'name', minWidth: 140 },
  { title: '组织', key: 'org_id', width: 90, render: (row) => (row.org_id === 0 ? '默认空间' : `#${row.org_id}`) },
  { title: '创建时间', key: 'created_at', minWidth: 160 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) =>
      h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, type: 'primary', disabled: !auth.hasPerm('secret:reveal'), onClick: () => handleReveal(row) }, { default: () => '解密查看' }),
        h(
          NPopconfirm,
          { disabled: !auth.hasPerm('secret:delete'), onPositiveClick: () => handleDeleteSecret(row), positiveText: '删除', negativeText: '取消' },
          {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error', disabled: !auth.hasPerm('secret:delete') }, { default: () => '删除' }),
            default: () => `确定删除密钥「${row.name}」？`,
          },
        ),
      ]),
  },
]

const reportColumns: DataTableColumns<SecurityReportItem> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '类型', key: 'kind', width: 100, render: (row) => h(NTag, { size: 'small' }, { default: () => (row.kind === 'baseline' ? '容器基线' : '镜像策略') }) },
  {
    title: '得分',
    key: 'score',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: row.score >= 90 ? 'success' : row.score >= 70 ? 'warning' : 'error' }, { default: () => String(row.score) }),
  },
  { title: '发现项', key: 'finding_count', width: 90 },
  { title: '扫描时间', key: 'created_at', minWidth: 170 },
]

onMounted(() => {
  load()
  loadSecrets()
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="security"
      title="安全中心"
      description="容器安全基线审计 · 镜像策略扫描 · 密钥托管（AES-256-GCM 加密）"
      :gradient="'linear-gradient(120deg, #5c1a1a 0%, #a02020 45%, #e05656 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num" :style="{ color: '#fff' }">{{ dash?.score ?? '--' }}</span><span class="lbl">综合得分</span></div>
        <div class="hero-pill"><span class="num">{{ dash?.finding_count ?? 0 }}</span><span class="lbl">发现项</span></div>
        <div class="hero-pill"><span class="num">{{ secrets.length }}</span><span class="lbl">托管密钥</span></div>
      </template>
      <template #action>
        <n-button v-if="auth.hasPerm('security:scan')" type="primary" size="small" :loading="scanning" @click="handleScan">
          <template #icon><DxIcon name="zap" :size="14" /></template>
          立即扫描
        </n-button>
      </template>
    </PageHero>

    <!-- 安全概览 -->
    <div class="grid grid-cols-1 lg:grid-cols-[320px_1fr_1fr_1fr] gap-3">
      <div class="dx-card dx-fade-up dx-delay-1 !p-3">
        <div class="text-[12px] font-semibold text-gray-500 dark:text-gray-400 mb-2 text-center">安全评分</div>
        <DonutChart
          :segments="scoreSegments"
          :size="120"
          :thickness="12"
          :center-text="String(dash?.score ?? '--')"
          center-label="综合得分"
        />
        <div class="mt-2 text-center text-[11px] text-gray-400 leading-relaxed">
          <span class="font-medium" :style="{ color: scoreColor }">{{ (dash?.score ?? 0) >= 90 ? '安全状况良好' : (dash?.score ?? 0) >= 70 ? '存在改进空间' : '需要立即处理' }}</span><br>
          基线规则 {{ dash?.baseline_rules?.length || 0 }} 条 · 镜像策略 {{ dash?.image_rules?.length || 0 }} 条
        </div>
      </div>
      <StatTile icon="zap" label="综合得分" :value="dash?.score ?? 0" :color="scoreColor" :trend="scoreTrend" hint="基线 + 镜像策略加权" />
      <StatTile icon="activity" label="发现项" :value="dash?.finding_count ?? 0" color="#ff7d00" hint="待处置的安全风险" />
      <StatTile icon="check" label="托管密钥" :value="secrets.length" color="#006eff" hint="AES-256-GCM 加密落库" />
    </div>

    <n-spin :show="loading">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <n-card :bordered="false" title="容器基线规则" class="dx-card">
          <ul class="text-xs text-gray-600 space-y-1">
            <li v-for="r in dash?.baseline_rules || []" :key="r" class="flex items-start gap-1.5"><DxIcon name="check" :size="12" class="text-emerald-500 mt-0.5 shrink-0" /> {{ r }}</li>
          </ul>
        </n-card>
        <n-card :bordered="false" title="镜像策略规则" class="dx-card">
          <ul class="text-xs text-gray-600 space-y-1">
            <li v-for="r in dash?.image_rules || []" :key="r" class="flex items-start gap-1.5"><DxIcon name="check" :size="12" class="text-emerald-500 mt-0.5 shrink-0" /> {{ r }}</li>
          </ul>
        </n-card>
      </div>

      <n-card :bordered="false" title="最新扫描发现" class="mt-4 dx-card">
        <n-empty v-if="!(dash?.reports || []).length" description="暂无扫描报告，点击「立即扫描」生成" class="py-6" />
        <div v-for="rep in dash?.reports || []" :key="rep.kind" class="mb-4 last:mb-0">
          <div class="flex items-center gap-2 mb-2">
            <n-tag size="small" :type="rep.score >= 90 ? 'success' : rep.score >= 70 ? 'warning' : 'error'">
              {{ rep.kind === 'baseline' ? '容器基线' : '镜像策略' }} · {{ rep.score }} 分
            </n-tag>
            <span class="text-xs text-gray-400">{{ rep.scanned_at }}</span>
          </div>
          <div v-if="rep.findings.length === 0" class="text-xs text-gray-400 flex items-center gap-1.5"><DxIcon name="check-circle" :size="14" class="text-emerald-500" /> 无发现项</div>
          <div v-for="(f, i) in rep.findings.slice(0, 12)" :key="i" class="flex items-start gap-2 py-1 text-xs">
            <n-tag size="tiny" :type="SeverityType[f.severity]">{{ SeverityNames[f.severity] }}</n-tag>
            <span class="text-gray-600"><b>{{ f.target }}</b>：{{ f.message }}</span>
          </div>
          <div v-if="rep.findings.length > 12" class="text-xs text-gray-400">… 其余 {{ rep.findings.length - 12 }} 项见扫描历史</div>
        </div>
      </n-card>

      <n-card :bordered="false" title="扫描历史" class="mt-4 dx-card">
        <n-data-table :columns="reportColumns" :data="reports" :bordered="false" size="small" :max-height="300" class="dx-table" />
      </n-card>

      <n-card :bordered="false" title="密钥托管" class="mt-4 dx-card">
        <template #header-extra>
          <n-button v-if="auth.hasPerm('secret:create')" size="small" type="primary" @click="showCreate = true">新建密钥</n-button>
        </template>
        <div class="text-xs text-gray-500 mb-3">
          当前空间：{{ orgStore.currentOrgName }}（org_id={{ orgStore.currentOrgId }}）· 密钥值以 AES-256-GCM 加密落库，接口永不返回明文，解密需 secret:reveal 权限并留审计
        </div>
        <n-data-table :columns="secretColumns" :data="secrets" :bordered="false" size="small" class="dx-table" />
        <div v-if="Object.keys(revealMap).length" class="mt-3 p-2 bg-amber-50 rounded text-xs">
          <div v-for="(v, id) in revealMap" :key="id" class="flex gap-2 items-center">
            <span class="text-gray-500 shrink-0">#{{ id }} 明文：</span>
            <code class="text-amber-700 break-all">{{ v }}</code>
          </div>
        </div>
      </n-card>
    </n-spin>

    <n-modal v-model:show="showCreate" preset="card" title="新建密钥" style="width: 460px">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="密钥名" required>
          <n-input v-model:value="createModel.name" placeholder="例如：DB_PASSWORD（组织内唯一）" />
        </n-form-item>
        <n-form-item label="密钥值" required>
          <n-input v-model:value="createModel.value" type="textarea" :rows="2" placeholder="明文输入，后端加密存储" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreateSecret">加密保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
