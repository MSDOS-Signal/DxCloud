<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { AppVersion, Application, Deployment } from '~/types'
import { DeployStatusNames, DeployStatusType } from '~/types'
import { api } from '~/services/http'

const route = useRoute()
const message = useMessage()
const id = computed(() => Number(route.params.id))

const app = ref<Application | null>(null)
const versions = ref<AppVersion[]>([])
const deployments = ref<Deployment[]>([])
const loading = ref(false)

async function load() {
  try {
    app.value = await api.get<Application>(`/applications/${id.value}`)
    versions.value = await api.get<AppVersion[]>(`/applications/${id.value}/versions`)
    deployments.value = await api.get<Deployment[]>(`/applications/${id.value}/deployments?limit=50`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

// ---------- 部署 ----------
const deployModel = reactive({
  image: '',
  host_port: 0,
  note: '',
  env_text: '',
})

async function handleDeploy() {
  loading.value = true
  try {
    const env: Record<string, string> = {}
    for (const line of deployModel.env_text.split('\n')) {
      const t = line.trim()
      if (!t) continue
      const i = t.indexOf('=')
      if (i > 0) env[t.slice(0, i).trim()] = t.slice(i + 1).trim()
    }
    await api.post(`/applications/${id.value}/deploy`, {
      image: deployModel.image || app.value?.image || undefined,
      env,
      host_port: deployModel.host_port || 0,
      note: deployModel.note,
    })
    message.success('部署完成（蓝绿切换 + 健康检查）')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '部署失败')
  } finally {
    loading.value = false
  }
}

async function handleRollback(v: AppVersion) {
  try {
    await api.post(`/applications/${id.value}/versions/${v.id}/rollback`)
    message.success(`已回滚到版本 ${v.version}`)
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '回滚失败')
  }
}

// ---------- 域名绑定 ----------
const domainText = ref('')

async function handleBindDomain() {
  try {
    await api.post('/domains', { domain: domainText.value.trim(), application_id: id.value, target_port: app.value?.port || 80 })
    message.success('域名已绑定，下次部署生效')
    domainText.value = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '绑定失败')
  }
}

function fmtEnv(envJSON: string): string {
  try {
    return (JSON.parse(envJSON) as string[]).join('，')
  } catch {
    return envJSON || '—'
  }
}

const versionColumns: DataTableColumns<AppVersion> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '版本', key: 'version', width: 110 },
  { title: '镜像', key: 'image_ref', ellipsis: { tooltip: true } },
  { title: '状态', key: 'status', width: 90 },
  { title: '创建时间', key: 'created_at', width: 165 },
  {
    title: '操作',
    key: 'actions',
    width: 110,
    render: (row) =>
      h(NButton, { size: 'tiny', type: 'warning', ghost: true, onClick: () => handleRollback(row) }, { default: () => '回滚' }),
  },
]

const deployColumns: DataTableColumns<Deployment> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '版本', key: 'version', width: 100 },
  { title: '镜像', key: 'image_ref', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: DeployStatusType[row.status] || 'default' }, { default: () => DeployStatusNames[row.status] || row.status }),
  },
  { title: '健康', key: 'health_status', width: 90, render: (row) => row.health_status || '—' },
  { title: '触发', key: 'trigger', width: 80 },
  { title: '说明', key: 'note', ellipsis: { tooltip: true } },
  { title: '时间', key: 'created_at', width: 165 },
]

onMounted(() => {
  load()
  if (app.value) deployModel.image = app.value.image
})

// 应用加载后回填部署镜像
watch(app, (a) => {
  if (a && !deployModel.image) deployModel.image = a.image
})
</script>

<template>
  <div class="space-y-4">
    <template v-if="app">
    <div class="dx-card dx-fade-up">
      <div class="dx-card-body">
        <div class="flex items-center justify-between">
          <div>
            <span class="text-lg font-semibold mr-3">{{ app.name }}</span>
            <n-tag size="small" bordered>{{ app.type }}</n-tag>
            <n-tag size="small" class="ml-2" type="info" bordered>{{ app.port }} 端口</n-tag>
            <n-tag v-if="app.domain" size="small" class="ml-2" type="success" bordered>{{ app.domain }}</n-tag>
          </div>
          <div class="text-xs text-gray-400">健康检查：{{ app.health_check_path || 'TCP' }} · 环境变量：{{ fmtEnv(app.env) }}</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-1">
        <div class="dx-card-header">部署新版本（蓝绿）</div>
        <div class="dx-card-body">
          <n-form label-placement="left" label-width="90" size="small">
            <n-form-item label="镜像">
              <n-input v-model:value="deployModel.image" placeholder="镜像（默认用应用配置）" />
            </n-form-item>
            <n-form-item label="宿主端口">
              <n-input-number v-model:value="deployModel.host_port" :min="0" :max="65535" style="width: 140px" placeholder="0=不发布" />
            </n-form-item>
            <n-form-item label="环境变量">
              <n-input v-model:value="deployModel.env_text" type="textarea" placeholder="每行 KEY=VALUE（覆盖应用默认）" :autosize="{ minRows: 2 }" />
            </n-form-item>
            <n-form-item label="备注">
              <n-input v-model:value="deployModel.note" placeholder="可选" />
            </n-form-item>
          </n-form>
          <div class="text-xs text-gray-400 mt-2">
            流程：创建候选容器 → 健康检查（60s）→ Traefik 优先级切换（域名零中断）→ 旧版本降级保留（可回滚）
          </div>
        </div>
        <div class="px-4 py-3 border-t border-gray-100 dark:border-gray-800 flex justify-end">
          <n-button type="primary" :loading="loading" @click="handleDeploy">部署</n-button>
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">域名绑定</div>
        <div class="dx-card-body">
          <n-space>
            <n-input v-model:value="domainText" placeholder="如 api.app.localhost（.localhost 免改 hosts）" style="width: 280px" />
            <n-button type="primary" ghost @click="handleBindDomain">绑定</n-button>
          </n-space>
          <div class="text-xs text-gray-400 mt-2">域名绑定后，部署容器自动携带 Traefik 路由标签，实现 Host 路由与蓝绿切换</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div class="dx-card dx-fade-up dx-delay-2">
        <div class="dx-card-header">版本历史（可回滚）</div>
        <div class="dx-card-body">
          <n-data-table :columns="versionColumns" :data="versions" size="small" :bordered="false" max-height="300" class="dx-table" />
        </div>
      </div>
      <div class="dx-card dx-fade-up dx-delay-3">
        <div class="dx-card-header">部署记录</div>
        <div class="dx-card-body">
          <n-data-table :columns="deployColumns" :data="deployments" size="small" :bordered="false" max-height="300" class="dx-table" />
        </div>
      </div>
    </div>
    </template>
    <n-skeleton v-else :repeat="4" text />
  </div>
</template>
