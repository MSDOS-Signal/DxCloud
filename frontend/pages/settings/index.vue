<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'

const message = useMessage()
const theme = useThemeStore()
const auth = useAuthStore()

interface SystemInfo {
  version: string
  docker_version: string
  docker_api_version: string
  os: string
  arch: string
  kernel: string
  mem_total: number
  cpu_count: number
}

const sysInfo = ref<SystemInfo | null>(null)
const loading = ref(false)

async function loadInfo() {
  loading.value = true
  try {
    sysInfo.value = await api.get<SystemInfo>('/health')
  } catch {
    // 忽略：系统信息获取失败不阻塞页面
  } finally {
    loading.value = false
  }
}

function fmtMem(mb: number): string {
  if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
  return mb + ' MB'
}

// ---------- 区域与镜像加速源 ----------
interface SettingsOverview {
  region: string
  registry_mirror: string
  mirror_candidates: string[]
  default_mirror: string
}

const region = ref<'cn' | 'global'>('cn')
const mirror = ref('')
const mirrorCandidates = ref<string[]>([])
const settingsLoading = ref(false)
const savingSettings = ref(false)
const testingMirror = ref(false)
const mirrorTestResult = ref<{ reachable: boolean; message: string; status_code?: number } | null>(null)
// 只有具备 settings:update 权限（管理员）才能保存，其余用户只读
const canEditSettings = auth.hasPerm('settings:update')

async function loadSettings() {
  settingsLoading.value = true
  try {
    const s = await api.get<SettingsOverview>('/settings')
    region.value = s.region === 'global' ? 'global' : 'cn'
    mirror.value = s.registry_mirror || ''
    mirrorCandidates.value = s.mirror_candidates || []
  } catch {
    // 忽略：读取失败使用默认值
  } finally {
    settingsLoading.value = false
  }
}

async function saveSettings() {
  if (!canEditSettings) {
    message.warning('当前账号没有修改系统设置的权限')
    return
  }
  savingSettings.value = true
  try {
    await api.put('/settings', { region: region.value, registry_mirror: mirror.value })
    message.success('区域与镜像源已保存')
    await loadSettings()
  } catch (e: any) {
    message.error(e?.message || '保存失败')
  } finally {
    savingSettings.value = false
  }
}

async function testMirror() {
  testingMirror.value = true
  mirrorTestResult.value = null
  try {
    const r = await api.post<{ reachable: boolean; message: string; status_code?: number }>('/settings/test-mirror', { mirror: mirror.value })
    mirrorTestResult.value = r
    message.success(r.reachable ? '镜像源连接正常' : '镜像源暂不可达')
  } catch (e: any) {
    message.error(e?.message || '测试失败')
  } finally {
    testingMirror.value = false
  }
}

const regionOptions = [
  {
    label: '中国大陆',
    value: 'cn',
    desc: '拉取官方 Docker Hub 镜像自动经加速源，速度更快',
  },
  {
    label: '非中国大陆',
    value: 'global',
    desc: '直连 Docker Hub 官方源',
  },
]

onMounted(() => {
  loadInfo()
  loadSettings()
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="settings"
      title="系统设置"
      description="平台系统信息 · 外观偏好 · 区域与镜像加速源 · 引擎状态"
      :gradient="'linear-gradient(120deg, #1f2937 0%, #37475d 45%, #5a7a9b 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ sysInfo?.cpu_count || '—' }}</span><span class="lbl">CPU 核心</span></div>
        <div class="hero-pill"><span class="num">{{ sysInfo?.mem_total ? fmtMem(sysInfo.mem_total) : '—' }}</span><span class="lbl">总内存</span></div>
        <div class="hero-pill"><span class="num">{{ region === 'cn' ? '中国大陆' : '全球' }}</span><span class="lbl">部署区域</span></div>
      </template>
    </PageHero>

    <!-- 外观设置 -->
    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <div class="flex items-center gap-2">
          <DxIcon name="theme" :size="15" class="text-[#006eff]" />
          <span>外观偏好</span>
        </div>
      </div>
      <div class="dx-card-body">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-[13px] font-medium text-gray-700 dark:text-gray-200">深色模式</div>
            <div class="text-xs text-gray-400 mt-0.5">切换浅色 / 深色主题</div>
          </div>
          <n-switch :value="theme.isDark" @update:value="theme.toggle()">
            <template #checked>深色</template>
            <template #unchecked>浅色</template>
          </n-switch>
        </div>
      </div>
    </div>

    <!-- 区域与镜像加速源 -->
    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <div class="flex items-center gap-2">
          <DxIcon name="globe" :size="15" class="text-[#006eff]" />
          <span>区域与镜像加速源</span>
        </div>
        <span class="text-[11px] text-gray-400">影响镜像拉取速度与下载源</span>
      </div>
      <div class="dx-card-body space-y-4">
        <!-- 区域选择 -->
        <div>
          <div class="text-[12px] text-gray-400 mb-2">部署区域</div>
          <div class="grid grid-cols-2 gap-3">
            <div
              v-for="opt in regionOptions"
              :key="opt.value"
              class="region-opt cursor-pointer rounded-lg border px-3.5 py-3 transition-all"
              :class="region === opt.value ? 'region-opt-active' : 'border-gray-200 dark:border-gray-700 hover:border-[#006eff]/50'"
              @click="region = opt.value as 'cn' | 'global'"
            >
              <div class="flex items-center gap-2">
                <span class="w-3.5 h-3.5 rounded-full border flex items-center justify-center" :class="region === opt.value ? 'border-[#006eff]' : 'border-gray-300 dark:border-gray-600'">
                  <span v-if="region === opt.value" class="w-2 h-2 rounded-full bg-[#006eff]" />
                </span>
                <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ opt.label }}</span>
              </div>
              <div class="text-[11px] text-gray-400 mt-1.5 leading-relaxed">{{ opt.desc }}</div>
            </div>
          </div>
        </div>

        <!-- 加速源（仅中国大陆区域生效） -->
        <div v-if="region === 'cn'">
          <div class="text-[12px] text-gray-400 mb-2">镜像加速源</div>
          <n-select
            v-model:value="mirror"
            :options="mirrorCandidates.map((m) => ({ label: m, value: m }))"
            filterable
            tag
            clearable
            placeholder="选择或输入加速源域名（如 hub.rat.dev）"
            :disabled="!canEditSettings"
          />
          <div class="flex items-center gap-2 mt-2">
            <n-button size="small" :loading="testingMirror" @click="testMirror">测试当前源</n-button>
            <span v-if="mirrorTestResult" class="text-[11px]" :class="mirrorTestResult.reachable ? 'text-green-600' : 'text-red-500'">
              {{ mirrorTestResult.message }}{{ mirrorTestResult.status_code ? `（HTTP ${mirrorTestResult.status_code}）` : '' }}
            </span>
          </div>
          <div class="text-[11px] text-gray-400 mt-1.5 leading-relaxed">
            中国大陆区域拉取官方镜像（如 nginx、mysql）时自动经此加速源，拉取完成后仍按原镜像名使用。若当前源拉取失败，平台会自动依次尝试其他国内候选源；仍失败时再在此更换源。
          </div>
        </div>
        <div v-else class="text-[12px] text-gray-400 bg-gray-50 dark:bg-gray-800/60 rounded-lg px-3 py-2.5">
          非中国大陆区域将直连 Docker Hub 官方源，无需配置加速源。
        </div>

        <div class="flex justify-end">
          <n-button type="primary" :loading="savingSettings" :disabled="!canEditSettings" @click="saveSettings">
            {{ canEditSettings ? '保存设置' : '无权限修改' }}
          </n-button>
        </div>
      </div>
    </div>

    <!-- 系统信息 -->
    <div class="dx-card dx-fade-up dx-delay-2">
      <div class="dx-card-header">
        <div class="flex items-center gap-2">
          <DxIcon name="info" :size="15" class="text-[#006eff]" />
          <span>系统信息</span>
        </div>
      </div>
      <div class="dx-card-body">
        <div class="grid grid-cols-2 gap-x-6">
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">平台版本</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">DxCloud v1.0</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">Docker 版本</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.docker_version || '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">Docker API</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.docker_api_version || '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">操作系统</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.os || '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">架构</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.arch || '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">CPU 核数</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.cpu_count || '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">总内存</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.mem_total ? fmtMem(sysInfo.mem_total) : '—' }}</span>
          </div>
          <div class="flex items-center justify-between py-2.5 border-b border-gray-100 dark:border-gray-800">
            <span class="text-[12px] text-gray-400">内核版本</span>
            <span class="text-[13px] font-medium text-gray-700 dark:text-gray-200">{{ sysInfo?.kernel || '—' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 引擎状态 -->
    <div class="dx-card dx-fade-up dx-delay-3">
      <div class="dx-card-header">
        <div class="flex items-center gap-2">
          <DxIcon name="activity" :size="15" class="text-[#006eff]" />
          <span>引擎状态</span>
        </div>
      </div>
      <div class="dx-card-body">
        <div class="flex items-center py-1">
          <span class="dx-status-dot dx-status-dot-running" />
          <span class="text-[13px] text-gray-700 dark:text-gray-200">Docker 引擎运行中</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.region-opt-active {
  border-color: #006eff;
  background: linear-gradient(135deg, rgba(0, 110, 255, 0.06), rgba(0, 110, 255, 0.02));
  box-shadow: 0 0 0 1px rgba(0, 110, 255, 0.35), 0 4px 14px rgba(0, 110, 255, 0.12);
}
</style>
