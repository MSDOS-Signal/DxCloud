<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { PermissionItem, RoleItem } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const auth = useAuthStore()

const loading = ref(false)
const rows = ref<RoleItem[]>([])
const allPerms = ref<PermissionItem[]>([])

const sysRoleCount = computed(() => rows.value.filter((r) => r.is_system).length)
const avgPerms = computed(() => (rows.value.length ? Math.round(rows.value.reduce((s, r) => s + r.permissions.length, 0) / rows.value.length) : 0))

const moduleNames: Record<string, string> = {
  ecs: 'ECS 云主机',
  image: '镜像',
  network: '网络',
  volume: '存储',
  registry: 'Registry',
  app: '应用',
  pipeline: 'Pipeline',
  project: '项目',
  domain: '域名',
  user: '用户管理',
  org: '组织',
  quota: '配额',
  billing: '计费',
  audit: '审计',
  settings: '设置',
  security: '安全中心',
  secret: '密钥管理',
}

async function load() {
  loading.value = true
  try {
    rows.value = await api.get<RoleItem[]>('/roles')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadPerms() {
  try {
    allPerms.value = await api.get<PermissionItem[]>('/permissions')
  } catch {
    // 忽略
  }
}

// ---------- 新建角色 ----------
const showCreate = ref(false)
const createModel = reactive({ code: '', name: '', description: '' })

async function handleCreate() {
  try {
    await api.post('/roles', createModel)
    message.success('角色创建成功')
    showCreate.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  }
}

async function handleDelete(row: RoleItem) {
  try {
    await api.del(`/roles/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

// ---------- 权限配置 ----------
const showPerm = ref(false)
const permTarget = ref<RoleItem | null>(null)
const permCodes = ref<string[]>([])

function openPerm(row: RoleItem) {
  permTarget.value = row
  permCodes.value = [...row.permissions]
  showPerm.value = true
}

async function handleGrant() {
  if (!permTarget.value) return
  try {
    await api.put(`/roles/${permTarget.value.id}/permissions`, { perm_codes: permCodes.value })
    message.success('权限已保存（受影响用户缓存已失效）')
    showPerm.value = false
    permTarget.value = null
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '保存失败')
  }
}

// 按模块分组权限
const moduleGroups = computed(() => {
  const map = new Map<string, PermissionItem[]>()
  for (const p of allPerms.value) {
    const list = map.get(p.module) || []
    list.push(p)
    map.set(p.module, list)
  }
  return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]))
})

const columns: DataTableColumns<RoleItem> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '代码', key: 'code', width: 130 },
  { title: '名称', key: 'name', width: 130 },
  { title: '说明', key: 'description', ellipsis: { tooltip: true } },
  {
    title: '类型',
    key: 'is_system',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: row.is_system ? 'info' : 'default', bordered: false }, { default: () => (row.is_system ? '系统' : '自定义') }),
  },
  {
    title: '权限数',
    key: 'permissions',
    width: 90,
    render: (row) => row.permissions.length,
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) =>
      h(NSpace, { size: 6 }, {
        default: () => [
          h(NButton, { size: 'small', type: 'primary', ghost: true, onClick: () => openPerm(row) }, { default: () => '配置权限' }),
          h(
            NButton,
            { size: 'small', type: 'error', ghost: true, disabled: row.is_system, onClick: () => handleDelete(row) },
            { default: () => '删除' },
          ),
        ],
      }),
  },
]

onMounted(() => {
  load()
  if (auth.hasPerm('user:grant')) {
    loadPerms()
  }
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="roles" title="角色管理"
      description="角色与权限分配 · 系统角色 / 自定义角色 / 权限点配置 · 用户可叠加多角色"
      :gradient="'linear-gradient(120deg, #132e5e 0%, #1e4d8f 45%, #3a7bd5 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ rows.length }}</span><span class="lbl">角色总数</span></div>
        <div class="hero-pill"><span class="num">{{ sysRoleCount }}</span><span class="lbl">系统角色</span></div>
        <div class="hero-pill"><span class="num">{{ allPerms.length }}</span><span class="lbl">权限点</span></div>
      </template>
      <template #action>
        <n-button v-if="auth.hasPerm('user:create')" size="small" type="primary" @click="showCreate = true">
          <template #icon><DxIcon name="plus" :size="13" /></template>
          新建角色
        </n-button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="roles" label="角色数量" :value="rows.length" suffix=" 个" color="#1e4d8f" hint="系统 + 自定义" />
      <StatTile icon="permissions" label="平均权限点" :value="avgPerms" suffix=" 个/角色" color="#13c2c2" hint="衡量角色粒度" />
      <StatTile icon="users" label="授权模型" value="—" color="#722ed1" hint="用户 ↔ 角色 ↔ 权限点" />
    </div>

    <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false" class="dx-card dx-table dx-fade-up dx-delay-1" />

    <!-- 新建角色 -->
    <n-modal v-model:show="showCreate" preset="card" title="新建角色" style="width: 520px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="代码">
          <n-input v-model:value="createModel.code" placeholder="小写字母/数字/短横线，如 qa-admin" />
        </n-form-item>
        <n-form-item label="名称">
          <n-input v-model:value="createModel.name" placeholder="角色显示名" />
        </n-form-item>
        <n-form-item label="说明">
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

    <!-- 权限配置 -->
    <n-modal
      v-model:show="showPerm"
      preset="card"
      :title="`配置权限 · ${permTarget?.name ?? ''}（${permTarget?.code ?? ''}）`"
      style="width: 720px"
    >
      <div class="max-h-96 overflow-auto">
        <div v-for="[module, perms] in moduleGroups" :key="module" class="mb-3">
          <div class="text-sm font-semibold text-gray-600 mb-1">{{ moduleNames[module] || module }}</div>
          <n-checkbox-group v-if="permTarget" v-model:value="permCodes">
            <n-space size="small" wrap>
              <n-tooltip v-for="p in perms" :key="p.code" trigger="hover">
                <template #trigger>
                  <n-checkbox :value="p.code" :label="p.name || p.code" size="small" />
                </template>
                {{ p.code }}
              </n-tooltip>
            </n-space>
          </n-checkbox-group>
        </div>
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showPerm = false">取消</n-button>
          <n-button type="primary" @click="handleGrant">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
