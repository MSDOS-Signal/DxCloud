<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { PermissionItem } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const loading = ref(false)
const groups = ref<{ module: string; items: PermissionItem[] }[]>([])

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

const moduleIcons: Record<string, string> = {
  ecs: 'ecs', image: 'images', network: 'networks', volume: 'volumes', registry: 'images',
  app: 'apps', pipeline: 'pipelines', project: 'projects', domain: 'domains', user: 'users',
  org: 'orgs', quota: 'billing', billing: 'billing', audit: 'logs', settings: 'settings',
  security: 'security', secret: 'permissions',
}

const moduleColors = ['#006eff', '#13c2c2', '#722ed1', '#fa8c16', '#00b42a', '#f53f3f', '#2f54eb', '#eb2f96', '#a0d911', '#faad14']

const totalPerms = computed(() => groups.value.reduce((s, g) => s + g.items.length, 0))

onMounted(async () => {
  loading.value = true
  try {
    const perms = await api.get<PermissionItem[]>('/permissions')
    const map = new Map<string, PermissionItem[]>()
    for (const p of perms) {
      const list = map.get(p.module) || []
      list.push(p)
      map.set(p.module, list)
    }
    groups.value = [...map.entries()]
      .sort((a, b) => b[1].length - a[1].length)
      .map(([module, items]) => ({ module, items }))
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="permissions" title="权限点清单"
      description="后端 RBAC 裁决依据（只读）· 接口层按 permission code 校验 · 角色可组合授权"
      :gradient="'linear-gradient(120deg, #d4380d 0%, #fa541c 45%, #ffa940 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ groups.length }}</span><span class="lbl">功能模块</span></div>
        <div class="hero-pill"><span class="num">{{ totalPerms }}</span><span class="lbl">权限点</span></div>
        <div class="hero-pill"><span class="num">{{ groups.length > 0 ? Math.max(...groups.map(g => g.items.length)) : 0 }}</span><span class="lbl">最大模块点数</span></div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="permissions" label="权限点总数" :value="totalPerms" suffix=" 个" color="#fa541c" hint="后端注册的全部 permission code" />
      <StatTile icon="apps" label="功能模块" :value="groups.length" suffix=" 个" color="#13c2c2" hint="按模块分组展示" />
      <StatTile icon="users" label="授权方式" value="—" color="#722ed1" hint="角色绑定权限点 → 用户绑定角色" />
    </div>

    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <span class="text-[14px] font-semibold text-gray-800 dark:text-gray-200">模块权限点</span>
        <span class="text-xs text-gray-400">按权限点数量排序</span>
      </div>
      <div class="dx-card-body">
        <n-skeleton v-if="loading" :repeat="6" text />
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          <div
            v-for="(g, gi) in groups" :key="g.module"
            class="perm-group"
            :style="{ animationDelay: (gi * 0.04) + 's' }"
          >
            <div class="flex items-center justify-between mb-2.5">
              <div class="flex items-center gap-2 min-w-0">
                <span class="perm-icon" :style="{ background: moduleColors[gi % 10] + '14', color: moduleColors[gi % 10] }">
                  <DxIcon :name="moduleIcons[g.module] || 'default'" :size="14" />
                </span>
                <span class="text-[13px] font-semibold text-gray-700 dark:text-gray-200 truncate">{{ moduleNames[g.module] || g.module }}</span>
                <code class="text-[10px] text-gray-400 shrink-0">{{ g.module }}</code>
              </div>
              <n-tag size="tiny" :bordered="false" :style="{ background: moduleColors[gi % 10] + '12', color: moduleColors[gi % 10] }">{{ g.items.length }}</n-tag>
            </div>
            <div class="flex flex-wrap gap-1.5">
              <n-tooltip v-for="p in g.items" :key="p.code" trigger="hover">
                <template #trigger>
                  <n-tag size="small" bordered class="perm-tag">{{ p.name || p.code }}</n-tag>
                </template>
                {{ p.code }}
              </n-tooltip>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.perm-group {
  padding: 14px;
  border: 1px solid #eef1f5;
  border-radius: 8px;
  background: #fafbfc;
  transition: all 0.2s ease;
  animation: perm-in 0.4s ease both;
}
@keyframes perm-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
.perm-group:hover {
  border-color: #ffd6bb;
  background: #fff;
  box-shadow: 0 4px 14px rgba(250, 84, 28, 0.08);
  transform: translateY(-2px);
}
html.dark .perm-group {
  background: #161b22;
  border-color: #21262d;
}
html.dark .perm-group:hover {
  border-color: #5a3520;
  background: #1a2029;
}
.perm-icon {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.perm-tag {
  cursor: default;
  transition: all 0.15s ease;
}
.perm-tag:hover {
  transform: translateY(-1px);
}
</style>
