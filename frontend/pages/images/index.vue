<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { DockerImage, PageResult } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'
import DonutChart from '~/components/DonutChart.vue'

const message = useMessage()
const loading = ref(false)
const rows = ref<DockerImage[]>([])
const total = ref(0)
const page = ref(1)
const size = 20
const keyword = ref('')

const readyCount = computed(() => rows.value.filter((r) => r.status === 'ready').length)
const pullingCount = computed(() => rows.value.filter((r) => r.status === 'pulling').length)
const failedCount = computed(() => rows.value.filter((r) => r.status === 'failed').length)
const totalSize = computed(() => rows.value.reduce((s, r) => s + (r.size_bytes || 0), 0))
const totalSizeGb = computed(() => totalSize.value / 1024 / 1024 / 1024)
const statusSegments = computed(() => [
  { value: readyCount.value, color: '#00b42a', label: '就绪' },
  { value: pullingCount.value, color: '#ff7d00', label: '拉取中' },
  { value: failedCount.value, color: '#f53f3f', label: '失败' },
])

let timer: number | undefined
let pullTimer: number | undefined
let searchTimer: number | undefined

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({ page: String(page.value), size: String(size) })
    if (keyword.value) q.set('keyword', keyword.value)
    const res = await api.get<PageResult<DockerImage>>(`/images?${q.toString()}`)
    rows.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function fmtSize(n: number): string {
  if (n >= 1024 * 1024 * 1024) return (n / 1024 / 1024 / 1024).toFixed(2) + ' GB'
  if (n >= 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + ' MB'
  if (n >= 1024) return (n / 1024).toFixed(1) + ' KB'
  return n + ' B'
}

// 拉取
const showPull = ref(false)
const pullImage = ref('')
const searchOptions = ref<any[]>([])
const searchLoading = ref(false)

interface ImageSearchItem {
  name: string
  description: string
  official: boolean
  source: string
}

watch(pullImage, (value) => {
  if (searchTimer) window.clearTimeout(searchTimer)
  const query = value.trim()
  if (query.length < 2) {
    searchOptions.value = []
    return
  }
  searchTimer = window.setTimeout(async () => {
    searchLoading.value = true
    try {
      const items = await api.get<ImageSearchItem[]>(`/images/search?q=${encodeURIComponent(query)}&limit=10`)
      searchOptions.value = items.map((item) => ({
        ...item,
        label: item.name,
        value: item.name,
      }))
    } catch {
      searchOptions.value = []
    } finally {
      searchLoading.value = false
    }
  }, 140)
})

interface PullTask {
  id: number
  image: string
  status: string
  pull_error: string
  logs: string
  progress: number | null
}

const showPullProgress = ref(false)
const pullTask = ref<PullTask | null>(null)
const pullLogBox = ref<HTMLElement | null>(null)
const visiblePullLogs = computed(() => {
  const logs = pullTask.value?.logs || ''
  if (logs.length <= 16_000) return logs
  return `... 日志较长，仅显示最后 16000 字符 ...\n${logs.slice(-16_000)}`
})

async function handlePull() {
  try {
    const image = pullImage.value.trim()
    const r = await api.post<{ id: number; status: string; repo: string; tag: string }>('/images/pull', { image })
    message.success(`已提交拉取任务（${r.status}），完成后自动更新`)
    showPull.value = false
    pullTask.value = {
      id: r.id,
      image,
      status: r.status,
      pull_error: '',
      logs: `正在准备拉取 ${image}\n`,
      progress: null,
    }
    showPullProgress.value = true
    if (pullTimer) window.clearInterval(pullTimer)
    pullTimer = window.setInterval(loadPullProgress, 1500)
    loadPullProgress()
    pullImage.value = ''
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '拉取失败')
  }
}

async function loadPullProgress() {
  if (!pullTask.value) return
  try {
    const r = await api.get<{ status: string; pull_error: string; logs: string; progress: number | null }>(`/images/${pullTask.value.id}/logs`)
    if (pullTask.value) {
      pullTask.value.status = r.status
      pullTask.value.pull_error = r.pull_error
      pullTask.value.logs = r.logs || pullTask.value.logs
      pullTask.value.progress = typeof r.progress === 'number' ? r.progress : null
    }
  } catch {
    // 任务轮询失败不覆盖当前内容，下一轮继续
  }
}

// 打标签
const tagTarget = ref<DockerImage | null>(null)
const showTag = ref(false)
const newRepo = ref('')
const newTag = ref('latest')

function openTag(row: DockerImage) {
  tagTarget.value = row
  newRepo.value = ''
  newTag.value = 'latest'
  showTag.value = true
}

async function handleTag() {
  if (!tagTarget.value || !newRepo.value.trim()) return
  try {
    await api.post(`/images/${tagTarget.value.id}/tag`, { repo: newRepo.value.trim(), tag: newTag.value.trim() })
    message.success('已打标签')
    showTag.value = false
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '打标签失败')
  }
}

async function handleDelete(row: DockerImage) {
  try {
    await api.del(`/images/${row.id}`)
    message.success('已删除')
    load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<DockerImage> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '仓库', key: 'repo', ellipsis: { tooltip: true } },
  { title: 'Tag', key: 'tag', width: 110 },
  { title: '大小', key: 'size_bytes', width: 100, render: (row) => fmtSize(row.size_bytes) },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      h(
        NTag,
        { size: 'small', type: row.status === 'ready' ? 'success' : row.status === 'pulling' ? 'warning' : 'error' },
        { default: () => ({ ready: '就绪', pulling: '拉取中', failed: '失败' }[row.status] || row.status) },
      ),
  },
  { title: '来源', key: 'source', width: 80 },
  {
    title: '说明',
    key: 'pull_error',
    ellipsis: { tooltip: true },
    render: (row) => row.pull_error || '—',
  },
  {
    title: '操作',
    key: 'actions',
    width: 170,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', ghost: true, type: 'primary', onClick: () => openTag(row) }, { default: () => '打标签' }),
          h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
        ],
      }),
  },
]

onMounted(() => {
  load()
  timer = window.setInterval(() => {
    if (rows.value.some((r) => r.status === 'pulling')) load()
  }, 3000)
})

watch(
  () => pullTask.value?.logs,
  () => {
    if (pullLogBox.value) pullLogBox.value.scrollTop = pullLogBox.value.scrollHeight
  },
)

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
  if (pullTimer) window.clearInterval(pullTimer)
  if (searchTimer) window.clearTimeout(searchTimer)
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="images"
      title="镜像中心"
      description="管理 Docker 镜像，支持自动联想搜索、异步拉取（国内自动走加速源）与打标签"
      :gradient="'linear-gradient(120deg, #085b52 0%, #0aa1a0 45%, #14c9c9 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ total }}</span><span class="lbl">镜像总数</span></div>
        <div class="hero-pill"><span class="num">{{ totalSizeGb.toFixed(2) }} GB</span><span class="lbl">本页容量</span></div>
      </template>
      <template #action>
        <n-button type="primary" size="small" @click="showPull = true">
          <template #icon><DxIcon name="download" :size="14" /></template>
          拉取镜像
        </n-button>
      </template>
    </PageHero>

    <!-- 统计与分布 -->
    <div class="grid grid-cols-1 lg:grid-cols-[1fr_1fr_1fr_1fr_320px] gap-3">
      <StatTile icon="images" label="镜像总数" :value="total" color="#0aa1a0" hint="分页统计全部镜像" />
      <StatTile icon="check" label="就绪可用" :value="readyCount" color="#00b42a" hint="可直接创建实例" />
      <StatTile icon="download" label="拉取中" :value="pullingCount" color="#ff7d00" hint="异步任务自动更新" />
      <StatTile icon="x" label="拉取失败" :value="failedCount" color="#f53f3f" hint="查看说明列排查" />
      <div class="dx-card dx-fade-up dx-delay-2 !p-3">
        <div class="text-[12px] font-semibold text-gray-500 dark:text-gray-400 mb-2 text-center">状态分布</div>
        <DonutChart
          v-if="statusSegments.some((s) => s.value > 0)"
          :segments="statusSegments"
          :size="100"
          :thickness="11"
          :center-text="String(rows.length)"
          center-label="本页镜像"
        />
        <n-empty v-else description="暂无镜像" size="small" class="py-6" />
      </div>
    </div>

    <!-- 列表卡片 -->
    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <div class="flex items-center gap-3">
          <n-input v-model:value="keyword" placeholder="搜索仓库 / Tag" clearable size="small" style="width: 240px" @keyup.enter="page = 1; load()" />
          <button class="dx-btn-secondary" @click="page = 1; load()">搜索</button>
        </div>
        <button class="dx-btn-primary" @click="showPull = true">拉取镜像</button>
      </div>
      <div class="dx-card-body">
        <n-data-table
          class="dx-table"
          :columns="columns"
          :data="rows"
          :loading="loading"
          :pagination="{ page: page, pageSize: size, itemCount: total, onChange: (p: number) => { page = p; load() } }"
          :bordered="false"
        />
      </div>
    </div>

    <!-- 拉取镜像弹窗 -->
    <n-modal v-model:show="showPull" preset="card" title="拉取镜像" style="width: 480px">
      <n-auto-complete
        v-model:value="pullImage"
        :options="searchOptions"
        :loading="searchLoading"
        placeholder="输入 ngi、mysql、openjdk 等关键词，自动联想"
        size="medium"
        @keyup.enter="handlePull"
        @select="(value: string) => { pullImage = value; }"
      >
        <template #option="option">
          <div class="py-1.5 px-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-[13px] font-medium text-gray-800 dark:text-gray-100">{{ option.label }}</span>
              <span v-if="option.official" class="dx-tag dx-tag-blue !h-4 !px-1.5 !text-[10px]">官方</span>
              <span class="text-[10px] text-gray-400">{{ option.source === 'hub' ? 'Docker Hub' : '常用目录' }}</span>
            </div>
            <div class="text-xs text-gray-400 mt-0.5 truncate">{{ option.description || '暂无描述' }}</div>
          </div>
        </template>
      </n-auto-complete>
      <div class="text-xs text-gray-400 mt-2">输入至少 2 个字符自动搜索；中国大陆网络不可达时会自动使用内置常用镜像目录</div>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showPull = false">取消</n-button>
          <n-button type="primary" @click="handlePull">拉取</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 拉取进度 -->
    <n-modal v-model:show="showPullProgress" preset="card" title="拉取任务日志" style="width: min(680px, calc(100vw - 24px))">
      <div v-if="pullTask" class="space-y-3">
        <div class="flex items-center gap-2">
          <span class="dx-status-dot" :class="pullTask.status === 'ready' ? 'dx-status-dot-running' : pullTask.status === 'failed' ? 'dx-status-dot-error' : 'dx-status-dot-pending'" />
          <span class="text-[13px] font-medium">{{ pullTask.image }}</span>
          <span class="dx-tag dx-tag-gray">{{ { ready: '拉取完成', pulling: '拉取中', failed: '失败' }[pullTask.status] || pullTask.status }}</span>
        </div>
        <div v-if="pullTask.status === 'pulling'" class="pull-progress">
          <div
            class="pull-progress-bar"
            :class="{ 'pull-progress-bar--indeterminate': pullTask.progress === null }"
            :style="{ width: pullTask.progress === null ? '28%' : `${Math.max(4, Math.min(pullTask.progress, 100))}%` }"
          />
        </div>
        <div v-if="pullTask.status === 'pulling'" class="text-[11px] text-gray-400">
          {{ pullTask.progress === null ? '当前镜像源未返回字节进度，正在等待传输' : `已下载 ${pullTask.progress.toFixed(1)}%` }}
        </div>
        <pre ref="pullLogBox" class="h-64 overflow-auto rounded bg-slate-950 p-3 text-[11px] leading-relaxed text-slate-100 whitespace-pre-wrap font-mono">{{ visiblePullLogs || '等待日志...' }}</pre>
        <div v-if="pullTask.pull_error" class="text-xs text-red-500 bg-red-50 dark:bg-red-950/30 rounded px-3 py-2">
          {{ pullTask.pull_error }}
        </div>
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button v-if="pullTask && pullTask.status === 'pulling'" @click="loadPullProgress">刷新日志</n-button>
          <n-button type="primary" @click="showPullProgress = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 打标签弹窗 -->
    <n-modal v-model:show="showTag" preset="card" :title="`打标签 · ${tagTarget?.repo ?? ''}:${tagTarget?.tag ?? ''}`" style="width: 480px">
      <n-space vertical>
        <n-input v-model:value="newRepo" placeholder="目标仓库，如 registry:5000/default/myapp" />
        <n-input v-model:value="newTag" placeholder="目标 tag" />
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showTag = false">取消</n-button>
          <n-button type="primary" @click="handleTag">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.pull-progress {
  height: 6px;
  overflow: hidden;
  border-radius: 999px;
  background: #e5e6eb;
}

html.dark .pull-progress {
  background: #30363d;
}

.pull-progress-bar {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #006eff, #00c6ff, #00b42a);
  transition: width 0.35s ease;
}

.pull-progress-bar--indeterminate {
  animation: pull-progress-sweep 1.4s ease-in-out infinite;
}

@keyframes pull-progress-sweep {
  0% { margin-left: -28%; }
  50% { margin-left: 52%; }
  100% { margin-left: 104%; }
}
</style>
