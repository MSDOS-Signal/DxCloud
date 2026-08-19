<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { PageResult, RoleItem, UserRow } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const auth = useAuthStore()

const loading = ref(false)
const rows = ref<UserRow[]>([])
const total = ref(0)
const page = ref(1)
const size = 20
const keyword = ref('')

// 角色下拉（分配角色用）
const roles = ref<RoleItem[]>([])

const activeCount = computed(() => rows.value.filter((r) => r.status === 1).length)
const disabledCount = computed(() => total.value - activeCount.value)
const roleDist = computed(() => {
  const m = new Map<string, number>()
  for (const r of rows.value) for (const c of r.roles) m.set(c, (m.get(c) || 0) + 1)
  return [...m.entries()].map(([code, n]) => ({ code, name: roles.value.find((x) => x.code === code)?.name || code, n }))
})

async function load() {
  loading.value = true
  try {
    const res = await api.get<PageResult<UserRow>>(
      `/users?page=${page.value}&size=${size}${keyword.value ? `&keyword=${encodeURIComponent(keyword.value)}` : ''}`,
    )
    rows.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadRoles() {
  try {
    roles.value = await api.get<RoleItem[]>('/roles')
  } catch {
    // 无权限时忽略（后端会拒绝分配动作）
  }
}

function roleName(code: string) {
  return roles.value.find((r) => r.code === code)?.name || code
}

// ---------- 新建用户 ----------
const showCreate = ref(false)
const createModel = reactive({ username: '', email: '', password: '', nickname: '', role_codes: [] as string[] })

async function handleCreate() {
  try {
    await api.post('/users', createModel)
    message.success('用户创建成功')
    showCreate.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

// ---------- 分配角色 ----------
const showGrant = ref(false)
const grantTarget = ref<UserRow | null>(null)
const grantCodes = ref<string[]>([])

function openGrant(row: UserRow) {
  grantTarget.value = row
  grantCodes.value = [...row.roles]
  showGrant.value = true
}

async function handleGrant() {
  if (!grantTarget.value) return
  try {
    await api.put(`/users/${grantTarget.value.id}/roles`, { role_codes: grantCodes.value })
    message.success('角色已更新')
    showGrant.value = false
    grantTarget.value = null
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '更新失败')
  }
}

// ---------- 禁用/启用、删除 ----------
async function toggleStatus(row: UserRow) {
  const next = row.status === 1 ? 2 : 1
  try {
    await api.put(`/users/${row.id}`, { status: next })
    message.success(next === 1 ? '已启用' : '已禁用')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

async function handleDelete(row: UserRow) {
  try {
    await api.del(`/users/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<UserRow> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '用户名', key: 'username', width: 140 },
  { title: '昵称', key: 'nickname', width: 120 },
  { title: '邮箱', key: 'email', width: 200 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { type: row.status === 1 ? 'success' : 'warning', size: 'small' },
        { default: () => (row.status === 1 ? '正常' : '禁用') },
      ),
  },
  {
    title: '角色',
    key: 'roles',
    render: (row) =>
      h(NSpace, { size: 4 }, { default: () => row.roles.map((r) => h(NTag, { size: 'small', bordered: false }, { default: () => roleName(r) })) }),
  },
  { title: '创建时间', key: 'created_at', width: 170 },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) =>
      h(NSpace, { size: 6 }, {
        default: () => [
          h(NButton, { size: 'small', type: 'primary', ghost: true, onClick: () => openGrant(row) }, { default: () => '分配角色' }),
          h(NButton, { size: 'small', onClick: () => toggleStatus(row) }, { default: () => (row.status === 1 ? '禁用' : '启用') }),
          h(
            NButton,
            { size: 'small', type: 'error', ghost: true, disabled: row.id === auth.user?.id, onClick: () => handleDelete(row) },
            { default: () => '删除' },
          ),
        ],
      }),
  },
]

onMounted(() => {
  load()
  loadRoles()
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="users" title="用户管理"
      description="用户账号与角色分配 · 创建 / 禁用 / 删除 / 授权 · RBAC 权限随角色生效"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ total }}</span><span class="lbl">用户总数</span></div>
        <div class="hero-pill"><span class="num">{{ activeCount }}</span><span class="lbl">正常</span></div>
        <div class="hero-pill"><span class="num">{{ roles.length }}</span><span class="lbl">角色数</span></div>
      </template>
      <template #action>
        <div class="flex items-center gap-2">
          <n-input v-model:value="keyword" placeholder="搜索用户名/邮箱/昵称" clearable size="small" style="width: 220px" @keyup.enter="page = 1; load()">
            <template #prefix><DxIcon name="search" :size="14" /></template>
          </n-input>
          <n-button size="small" @click="page = 1; load()">搜索</n-button>
          <n-button v-if="auth.hasPerm('user:create')" size="small" type="primary" @click="showCreate = true">
            <template #icon><DxIcon name="plus" :size="13" /></template>
            新建用户
          </n-button>
        </div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
      <n-data-table
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="{ page: page, pageSize: size, itemCount: total, onChange: (p: number) => { page = p; load() } }"
        :bordered="false"
        class="dx-card dx-table dx-fade-up dx-delay-1"
      />
      <div class="space-y-4">
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">角色分布</span>
          </div>
          <div class="dx-card-body flex items-center justify-center py-2">
            <DonutChart
              v-if="roleDist.length > 0"
              :size="130"
              :center-text="String(total)"
              center-label="用户总数"
              :segments="roleDist.map((r, i) => ({
                value: r.n,
                color: ['#006eff', '#13c2c2', '#722ed1', '#fa8c16', '#00b42a', '#eb2f96'][i % 6],
                label: r.name,
              }))"
            />
            <n-empty v-else description="暂无数据" class="py-8" />
          </div>
        </div>
        <StatTile icon="users" label="账号状态" :value="activeCount" :suffix="` / ${total} 正常`" color="#00b42a" hint="禁用账号无法登录" />
      </div>
    </div>

    <!-- 新建用户 -->
    <n-modal v-model:show="showCreate" preset="card" title="新建用户" style="width: 520px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="用户名">
          <n-input v-model:value="createModel.username" placeholder="3-32 位" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="createModel.email" placeholder="you@example.com" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="createModel.password" type="password" show-password-on="click" placeholder="至少 8 位" />
        </n-form-item>
        <n-form-item label="昵称">
          <n-input v-model:value="createModel.nickname" placeholder="可选" />
        </n-form-item>
        <n-form-item label="角色">
          <n-select v-model:value="createModel.role_codes" multiple :options="roles.map(r => ({ label: `${r.name}(${r.code})`, value: r.code }))" placeholder="默认 user" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 分配角色 -->
    <n-modal v-model:show="showGrant" preset="card" :title="`分配角色 · ${grantTarget?.username ?? ''}`" style="width: 520px">
      <n-select
        v-if="grantTarget"
        v-model:value="grantCodes"
        multiple
        :options="roles.map(r => ({ label: `${r.name}(${r.code})`, value: r.code }))"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showGrant = false">取消</n-button>
          <n-button type="primary" @click="handleGrant">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
