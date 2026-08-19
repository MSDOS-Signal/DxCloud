<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { Organization, OrgMember, ResourceQuota } from '~/types'
import { OrgRoleNames, PlanNames, QuotaTypeNames } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()
const orgStore = useOrgStore()

const loading = ref(false)
const rows = ref<Organization[]>([])

const normalCount = computed(() => rows.value.filter((r) => r.status === 1).length)
const creditSum = computed(() => rows.value.reduce((s, r) => s + (r.credit || 0), 0))
const planSegments = computed(() => {
  const colors: Record<string, string> = { free: '#006eff', pro: '#722ed1', enterprise: '#14c9c9' }
  const m = new Map<string, number>()
  for (const r of rows.value) m.set(r.plan, (m.get(r.plan) || 0) + 1)
  return [...m.entries()].map(([plan, n]) => ({ value: n, color: colors[plan] || '#86909c', label: PlanNames[plan] || plan }))
})

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<Organization[]>('/organizations')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

// ---------- 新建组织 ----------
const showCreate = ref(false)
const createModel = reactive({ name: '', code: '', plan: 'free' })

async function handleCreate() {
  if (!createModel.name.trim()) {
    message.warning('请输入组织名')
    return
  }
  try {
    await api.post('/organizations', createModel)
    message.success('组织创建成功')
    showCreate.value = false
    createModel.name = ''
    createModel.code = ''
    createModel.plan = 'free'
    await load()
    await orgStore.loadMine()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

// ---------- 删除组织 ----------
async function handleDelete(row: Organization) {
  try {
    await api.del(`/organizations/${row.id}`)
    message.success('已删除')
    await load()
    await orgStore.loadMine()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

function confirmDelete(row: Organization) {
  dialog.warning({
    title: '删除组织',
    content: `确定删除「${row.name}」吗？组织下的项目/资源将无法再按组织访问（数据保留但需迁移）。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => handleDelete(row),
  })
}

// ---------- 配额 ----------
const showQuota = ref(false)
const quotaOrg = ref<Organization | null>(null)
const quotaRows = ref<ResourceQuota[]>([])
const quotaLoading = ref(false)

async function openQuota(row: Organization) {
  quotaOrg.value = row
  showQuota.value = true
  quotaLoading.value = true
  try {
    quotaRows.value = await api.get<ResourceQuota[]>(`/quotas?org_id=${row.id}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '配额加载失败')
    quotaRows.value = []
  } finally {
    quotaLoading.value = false
  }
}

async function saveQuota(q: ResourceQuota) {
  if (!quotaOrg.value) return
  try {
    await api.put(`/quotas?org_id=${quotaOrg.value.id}`, {
      resource_type: q.resource_type,
      limit_value: q.limit_value,
    })
    message.success(`配额「${QuotaTypeNames[q.resource_type] || q.resource_type}」已更新为 ${q.limit_value}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '配额更新失败')
    await openQuota(quotaOrg.value)
  }
}

// ---------- 成员 ----------
const showMembers = ref(false)
const memberOrg = ref<Organization | null>(null)
const memberRows = ref<OrgMember[]>([])
const memberLoading = ref(false)
const addUsername = ref('')

async function openMembers(row: Organization) {
  memberOrg.value = row
  showMembers.value = true
  addUsername.value = ''
  await loadMembers()
}

async function loadMembers() {
  if (!memberOrg.value) return
  memberLoading.value = true
  try {
    memberRows.value = await api.get<OrgMember[]>(`/organizations/${memberOrg.value.id}/members`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '成员加载失败')
    memberRows.value = []
  } finally {
    memberLoading.value = false
  }
}

async function handleAddMember() {
  if (!memberOrg.value || !addUsername.value.trim()) {
    message.warning('请输入用户名')
    return
  }
  try {
    await api.post(`/organizations/${memberOrg.value.id}/members`, {
      username: addUsername.value.trim(),
      org_role: 'member',
    })
    message.success('成员已添加')
    addUsername.value = ''
    await loadMembers()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '添加失败')
  }
}

async function handleRemoveMember(m: OrgMember) {
  if (!memberOrg.value) return
  try {
    await api.del(`/organizations/${memberOrg.value.id}/members/${m.user_id}`)
    message.success('成员已移除')
    await loadMembers()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '移除失败')
  }
}

const columns: DataTableColumns<Organization> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '组织名', key: 'name', minWidth: 140 },
  { title: '标识', key: 'code', width: 120 },
  {
    title: '套餐',
    key: 'plan',
    width: 90,
    render: (row) => h(NTag, { size: 'small', type: row.plan === 'free' ? 'default' : 'info' }, { default: () => PlanNames[row.plan] || row.plan }),
  },
  {
    title: '虚拟余额（¥）',
    key: 'credit',
    width: 120,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.credit > 0 ? 'success' : 'error' },
        { default: () => row.credit.toFixed(2) },
      ),
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row) =>
      h(NTag, { size: 'small', type: row.status === 1 ? 'success' : 'warning' }, { default: () => (row.status === 1 ? '正常' : '停用') }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row) =>
      h(NSpace, { size: 4 }, () => [
        h(
          NButton,
          { size: 'small', quaternary: true, type: 'primary', disabled: !auth.hasPerm('quota:view'), onClick: () => openQuota(row) },
          { default: () => '配额' },
        ),
        h(
          NButton,
          { size: 'small', quaternary: true, type: 'primary', disabled: !auth.hasPerm('org:list'), onClick: () => openMembers(row) },
          { default: () => '成员' },
        ),
        h(
          NPopconfirm,
          {
            disabled: !auth.hasPerm('org:delete'),
            onPositiveClick: () => confirmDelete(row),
            positiveText: '删除',
            negativeText: '取消',
          },
          {
            trigger: () =>
              h(NButton, { size: 'small', quaternary: true, type: 'error', disabled: !auth.hasPerm('org:delete') }, { default: () => '删除' }),
            default: () => `确定删除「${row.name}」？`,
          },
        ),
      ]),
  },
]

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="orgs"
      title="组织管理"
      description="多租户隔离：组织 → 项目 → 资源；每个组织拥有独立配额与虚拟计费账户"
      :gradient="'linear-gradient(120deg, #3d1d6e 0%, #6428c8 45%, #9a6bff 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">组织总数</span></div>
        <div class="hero-pill"><span class="num">¥{{ creditSum.toFixed(0) }}</span><span class="lbl">虚拟余额合计</span></div>
      </template>
      <template #action>
        <n-button v-if="auth.hasPerm('org:create')" type="primary" size="small" @click="showCreate = true">
          <template #icon><DxIcon name="plus" :size="14" /></template>
          新建组织
        </n-button>
      </template>
    </PageHero>

    <!-- 统计与套餐分布 -->
    <div class="grid grid-cols-1 lg:grid-cols-[1fr_1fr_1fr_320px] gap-3">
      <StatTile icon="orgs" label="组织总数" :value="rows.length" color="#6428c8" hint="组织 → 项目 → 资源三级隔离" />
      <StatTile icon="check" label="正常状态" :value="normalCount" color="#00b42a" :suffix="` / ${rows.length} 正常`" hint="停用组织无法访问资源" />
      <StatTile icon="billing" label="虚拟余额合计" :value="creditSum" :decimals="2" suffix=" 元" color="#d48806" hint="按组织独立计费账户" />
      <div class="dx-card dx-fade-up dx-delay-2 !p-3">
        <div class="text-[12px] font-semibold text-gray-500 dark:text-gray-400 mb-2 text-center">套餐分布</div>
        <DonutChart
          v-if="planSegments.length > 0"
          :segments="planSegments"
          :size="100"
          :thickness="11"
          :center-text="String(rows.length)"
          center-label="组织"
        />
        <n-empty v-else description="暂无组织" size="small" class="py-6" />
      </div>
    </div>

    <n-card :bordered="false" class="dx-card">
      <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false" class="dx-table" />
    </n-card>

    <!-- 新建组织 -->
    <n-modal v-model:show="showCreate" preset="card" title="新建组织" style="width: 480px">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="组织名" required>
          <n-input v-model:value="createModel.name" placeholder="例如：星辰科技" />
        </n-form-item>
        <n-form-item label="标识">
          <n-input v-model:value="createModel.code" placeholder="唯一标识（可选）" />
        </n-form-item>
        <n-form-item label="套餐">
          <n-select v-model:value="createModel.plan" :options="[
            { label: '免费版（配额 5 实例 / 8 核）', value: 'free' },
            { label: '专业版', value: 'pro' },
            { label: '企业版', value: 'enterprise' },
          ]" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 配额抽屉 -->
    <n-drawer v-model:show="showQuota" :width="420">
      <n-drawer-content :title="`配额 · ${quotaOrg?.name || ''}`">
        <n-spin :show="quotaLoading">
          <div class="space-y-3">
            <p class="text-xs text-gray-500">组织级资源配额（限制组织内所有成员合计占用）</p>
            <div v-for="q in quotaRows" :key="q.resource_type" class="flex items-center gap-3">
              <span class="w-36 text-sm shrink-0">{{ QuotaTypeNames[q.resource_type] || q.resource_type }}</span>
              <n-input-number
                v-model:value="q.limit_value"
                size="small"
                :min="0"
                :max="100000"
                style="flex: 1"
                :disabled="!auth.hasPerm('quota:update')"
              />
              <n-button size="small" type="primary" :disabled="!auth.hasPerm('quota:update')" @click="saveQuota(q)">保存</n-button>
            </div>
          </div>
        </n-spin>
      </n-drawer-content>
    </n-drawer>

    <!-- 成员抽屉 -->
    <n-drawer v-model:show="showMembers" :width="460">
      <n-drawer-content :title="`成员 · ${memberOrg?.name || ''}`">
        <div class="space-y-4">
          <div v-if="auth.hasPerm('org:member:manage')" class="flex gap-2">
            <n-input v-model:value="addUsername" placeholder="输入用户名添加成员" style="flex: 1" @keyup.enter="handleAddMember" />
            <n-button type="primary" @click="handleAddMember">添加</n-button>
          </div>
          <n-spin :show="memberLoading">
            <n-empty v-if="memberRows.length === 0" description="暂无成员" />
            <div v-for="m in memberRows" :key="m.id" class="flex items-center justify-between py-2 border-b border-gray-100">
              <div>
                <div class="text-sm font-medium">用户 #{{ m.user_id }}</div>
                <div class="text-xs text-gray-400 mt-0.5">加入于 {{ m.created_at }}</div>
              </div>
              <div class="flex items-center gap-2">
                <n-tag size="small" :type="m.org_role === 'owner' ? 'success' : 'info'">{{ OrgRoleNames[m.org_role] || m.org_role }}</n-tag>
                <n-popconfirm
                  v-if="auth.hasPerm('org:member:manage') && m.org_role !== 'owner'"
                  :on-positive-click="() => handleRemoveMember(m)"
                  positive-text="移除"
                  negative-text="取消"
                >
                  <template #trigger>
                    <n-button size="tiny" quaternary type="error">移除</n-button>
                  </template>
                  确定将该成员移出组织？
                </n-popconfirm>
              </div>
            </div>
          </n-spin>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>
