<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { RegistryItem, RegistryRepo } from '~/types'
import { api } from '~/services/http'
import PageHero from '~/components/PageHero.vue'
import StatTile from '~/components/StatTile.vue'

const message = useMessage()
const registries = ref<RegistryItem[]>([])
const registryId = ref<number | null>(null)
const repos = ref<RegistryRepo[]>([])
const loading = ref(false)

const totalTags = computed(() => repos.value.reduce((s, r) => s + r.tags.length, 0))
const currentRegistry = computed(() => registries.value.find((r) => r.id === registryId.value))

async function loadRegistries() {
  try {
    registries.value = await api.get<RegistryItem[]>('/registries')
    if (registries.value.length && registryId.value === null) {
      registryId.value = registries.value[0].id
      await loadRepos()
    }
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  }
}

async function loadRepos() {
  if (registryId.value === null) return
  loading.value = true
  try {
    repos.value = await api.get<RegistryRepo[]>(`/registries/${registryId.value}/repositories`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '仓库列表加载失败')
  } finally {
    loading.value = false
  }
}

async function handlePull(repo: RegistryRepo, tag: string) {
  try {
    await api.post(`/registries/${registryId.value}/repositories/pull`, { name: repo.name, tag })
    message.success(`已拉取 ${repo.name}:${tag} 到本机引擎`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '拉取失败')
  }
}

async function handleDeleteTag(repo: RegistryRepo, tag: string) {
  try {
    await api.post(`/registries/${registryId.value}/repositories/delete-tag`, { name: repo.name, tag })
    message.success('Tag 已删除')
    loadRepos()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

const columns: DataTableColumns<RegistryRepo> = [
  { title: '仓库（namespace/name）', key: 'name', ellipsis: { tooltip: true } },
  {
    title: 'Tags',
    key: 'tags',
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () =>
          row.tags.length
            ? row.tags.map((t) =>
                h(NTag, { size: 'small', bordered: false }, { default: () => t }),
              )
            : [h('span', { class: 'text-gray-400' }, '（无）')],
      }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () =>
          row.tags.map((t) => [
            h(NButton, { size: 'tiny', ghost: true, type: 'primary', onClick: () => handlePull(row, t) }, { default: () => `拉取 :${t}` }),
            h(NButton, { size: 'tiny', ghost: true, type: 'error', onClick: () => handleDeleteTag(row, t) }, { default: () => '删除' }),
          ]).flat(),
      }),
  },
]

onMounted(loadRegistries)
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="images" title="私有镜像仓库"
      description="平台内置 Registry（:15000）· Pipeline 构建产物推送至此 · 可一键拉取到本机引擎"
      :gradient="'linear-gradient(120deg, #4c1d95 0%, #6d28d9 45%, #a78bfa 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ repos.length }}</span><span class="lbl">仓库数</span></div>
        <div class="hero-pill"><span class="num">{{ totalTags }}</span><span class="lbl">Tag 总数</span></div>
        <div class="hero-pill"><span class="num">{{ registries.length }}</span><span class="lbl">Registry</span></div>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile icon="database" label="当前 Registry" :value="currentRegistry ? '—' : 0" color="#6d28d9" :hint="currentRegistry?.url || '未选择'" />
      <StatTile icon="images" label="仓库数量" :value="repos.length" suffix=" 个" color="#006eff" hint="namespace/name 形式" />
      <StatTile icon="download" label="Tag 总数" :value="totalTags" suffix=" 个" color="#13c2c2" hint="各仓库版本合计" />
    </div>

    <div class="dx-card dx-fade-up dx-delay-1">
      <div class="dx-card-header">
        <div class="flex items-center gap-3">
          <n-select
            v-model:value="registryId"
            :options="registries.map((r) => ({ label: `${r.name}（${r.url}）`, value: r.id }))"
            style="width: 320px"
            @update:value="loadRepos"
          />
          <n-button size="small" @click="loadRepos">
            <template #icon><DxIcon name="refresh" :size="13" /></template>
            刷新
          </n-button>
        </div>
        <span class="text-xs text-gray-400">命名规范：{org}/{project}/{app} · 推送由 Pipeline 完成</span>
      </div>
      <div class="dx-card-body">
        <n-data-table :columns="columns" :data="repos" :loading="loading" :bordered="false" class="dx-table" />
      </div>
    </div>
  </div>
</template>
