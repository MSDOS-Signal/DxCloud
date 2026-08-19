<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { CloudNetwork, CloudVolume, DockerImage, EcsInstance, PortMapping } from '~/types'
import { api } from '~/services/http'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'

const message = useMessage()
const router = useRouter()

const loading = ref(false)
const model = reactive({
  name: '',
  description: '',
  image: '',
  cpu: 1,
  memory_mb: 512,
  disk_gb: 10,
  restart_policy: 'no' as string,
  readonly_rootfs: false,
  command_text: '',
})

const ports = ref<PortMapping[]>([{ container_port: 80, host_port: 0, protocol: 'tcp' }])
const envText = ref('')
const networks = ref<CloudNetwork[]>([])
const volumes = ref<CloudVolume[]>([])
const mounts = ref<{ volume_id: number | null; target: string; read_only: boolean }[]>([])
const networkId = ref<string>('')
const fixedIp = ref('')
const images = ref<DockerImage[]>([])

// ---------- 镜像中心联动 ----------
const readyImages = computed(() => images.value.filter((i) => i.status === 'ready'))
const imageInCenter = computed(() => readyImages.value.some((i) => `${i.repo}:${i.tag}` === model.image.trim()))
const selectedImage = computed(() => readyImages.value.find((i) => `${i.repo}:${i.tag}` === model.image.trim()))

// 常用镜像（不在镜像中心时提示去拉取）
const suggestImages = ['alpine:3.20', 'nginx:latest', 'redis:7-alpine', 'mysql:8.0', 'registry:2.8']

function pickImage(ref: string) {
  model.image = ref
}

async function loadInfra() {
  try {
    const [ns, vs, imgs] = await Promise.all([
      api.get<CloudNetwork[]>('/networks'),
      api.get<CloudVolume[]>('/volumes'),
      api.get<PageResult<DockerImage>>('/images?page=1&size=100'),
    ])
    networks.value = ns
    volumes.value = vs
    images.value = imgs.items
    // 默认选中第一个可用镜像
    if (!model.image && readyImages.value.length > 0) {
      const first = readyImages.value[0]
      model.image = `${first.repo}:${first.tag}`
    } else if (!model.image) {
      model.image = 'alpine:3.20'
    }
  } catch {
    if (!model.image) model.image = 'alpine:3.20'
  }
}

// ---------- 规格预设 ----------
interface SpecPreset {
  key: string
  name: string
  desc: string
  cpu: number
  memory_mb: number
}
const specPresets: SpecPreset[] = [
  { key: 'xs', name: '轻量型', desc: '适合测试 / 轻量工具', cpu: 0.5, memory_mb: 256 },
  { key: 's', name: '标准型', desc: '适合个人网站 / 小型应用', cpu: 1, memory_mb: 512 },
  { key: 'm', name: '均衡型', desc: '适合数据库 / 中型应用', cpu: 2, memory_mb: 2048 },
  { key: 'l', name: '进阶型', desc: '适合高并发服务', cpu: 4, memory_mb: 4096 },
]
const activePreset = computed(() => specPresets.find((p) => p.cpu === model.cpu && p.memory_mb === model.memory_mb)?.key ?? 'custom')

function applyPreset(p: SpecPreset) {
  model.cpu = p.cpu
  model.memory_mb = p.memory_mb
}

// ---------- 汇总 ----------
const validPorts = computed(() => ports.value.filter((p) => p.container_port > 0 && p.host_port > 0))
const validMounts = computed(() => mounts.value.filter((m) => m.volume_id !== null && m.target.trim() !== ''))
// 虚拟计费单价（与后端计费一致）：CPU ¥0.1/核时 内存 ¥0.05/GB时 磁盘 ¥0.01/GB时
const hourlyCost = computed(() => {
  const c = model.cpu * 0.1 + (model.memory_mb / 1024) * 0.05 + model.disk_gb * 0.01
  return `¥${c.toFixed(3)}/小时`
})

function fmtSize(bytes: number): string {
  if (!bytes) return ''
  const mb = bytes / 1024 / 1024
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`
  return `${mb.toFixed(0)} MB`
}

function addMount() {
  mounts.value.push({ volume_id: null, target: '', read_only: false })
}

function removeMount(idx: number) {
  mounts.value.splice(idx, 1)
}

onMounted(loadInfra)

function addPort() {
  ports.value.push({ container_port: 0, host_port: 0, protocol: 'tcp' })
}

function removePort(idx: number) {
  ports.value.splice(idx, 1)
}

async function submit() {
  if (!model.name.trim()) {
    message.error('请输入实例名称')
    return
  }
  if (!model.image.trim()) {
    message.error('请选择或输入镜像')
    return
  }
  if (!imageInCenter.value) {
    message.warning(`镜像「${model.image}」不在镜像中心，创建可能失败。建议先到镜像中心拉取。`)
  }
  const env = envText.value
    .split('\n')
    .map((s) => s.trim())
    .filter((s) => s !== '')
  const command = model.command_text
    .split(/\s+/)
    .map((s) => s.trim())
    .filter((s) => s !== '')

  loading.value = true
  try {
    const inst = await api.post<EcsInstance>('/ecs', {
      name: model.name.trim(),
      description: model.description,
      image: model.image.trim(),
      cpu: model.cpu,
      memory_mb: model.memory_mb,
      disk_gb: model.disk_gb,
      ports: validPorts.value,
      env,
      command,
      restart_policy: model.restart_policy,
      readonly_rootfs: model.readonly_rootfs,
      network_id: networkId.value || '',
      fixed_ip: fixedIp.value.trim(),
      mounts: validMounts.value.map((m) => ({ volume_id: m.volume_id, target: m.target.trim(), read_only: m.read_only })),
    })
    message.success(`实例 ${inst.instance_no} 创建成功，正在启动`)
    router.push(`/ecs/${inst.id}`)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="create-page">
    <PageHero
      icon="ecs"
      title="创建 ECS 实例"
      description="配置容器云主机实例的镜像、规格、网络与存储，实时预估每小时费用"
      :gradient="'linear-gradient(120deg, #06307a 0%, #0a5ad6 45%, #3a8dff 100%)'"
    >
      <template #stats>
        <div class="hero-pill"><span class="num">{{ model.cpu }}</span><span class="lbl">vCPU</span></div>
        <div class="hero-pill"><span class="num">{{ model.memory_mb >= 1024 ? (model.memory_mb / 1024).toFixed(1) + ' GB' : model.memory_mb + ' MB' }}</span><span class="lbl">内存</span></div>
        <div class="hero-pill"><span class="num">{{ model.disk_gb }} GB</span><span class="lbl">磁盘</span></div>
        <div class="hero-pill"><span class="num">{{ hourlyCost }}</span><span class="lbl">预估费用</span></div>
      </template>
      <template #action>
        <n-button quaternary size="small" class="!text-white" @click="router.push('/ecs')">
          <template #icon><DxIcon name="arrow-left" :size="14" /></template>
          返回实例列表
        </n-button>
      </template>
    </PageHero>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_360px] gap-4 items-start">
      <!-- 左侧：分步配置 -->
      <div class="space-y-4">
        <!-- ① 基础配置 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">1</span>基础配置
          </div>
          <div class="dx-card-body grid grid-cols-1 md:grid-cols-2 gap-x-6">
            <n-form-item label="实例名称" required :show-feedback="false" class="mb-4">
              <n-input v-model:value="model.name" placeholder="2-64 位，字母/数字/._-，如 web-server" />
            </n-form-item>
            <n-form-item label="描述" :show-feedback="false" class="mb-4">
              <n-input v-model:value="model.description" placeholder="可选，便于识别用途" />
            </n-form-item>
          </div>
        </section>

        <!-- ② 镜像 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">2</span>镜像
            <span class="ml-auto text-xs text-gray-400 font-normal">来自镜像中心，可直接选用</span>
          </div>
          <div class="dx-card-body space-y-3">
            <div v-if="readyImages.length > 0" class="grid grid-cols-2 sm:grid-cols-3 gap-2">
              <button
                v-for="img in readyImages" :key="img.id"
                type="button"
                class="image-tile"
                :class="{ active: model.image === `${img.repo}:${img.tag}` }"
                @click="pickImage(`${img.repo}:${img.tag}`)"
              >
                <div class="flex items-center gap-1.5 min-w-0">
                  <DxIcon name="images" :size="14" class="text-[#006eff] shrink-0" />
                  <span class="font-medium truncate text-[13px]">{{ img.repo }}</span>
                </div>
                <div class="flex items-center justify-between mt-1 text-[11px] text-gray-400">
                  <span class="dx-tag-mini">{{ img.tag }}</span>
                  <span v-if="img.size_bytes">{{ fmtSize(img.size_bytes) }}</span>
                </div>
              </button>
            </div>
            <n-alert v-else type="warning" :show-icon="false" class="text-xs">
              镜像中心暂无可用镜像，可先到
              <a class="text-[#006eff] font-medium" @click.prevent="router.push('/images')">镜像中心</a>
              拉取，或直接在下方手动输入镜像名（创建时将尝试使用）。
            </n-alert>

            <div class="flex flex-wrap gap-1.5" v-if="readyImages.length > 0">
              <span class="text-[11px] text-gray-400 self-center mr-1">快捷输入：</span>
              <n-tag
                v-for="s in suggestImages.filter(s => !readyImages.some(i => `${i.repo}:${i.tag}` === s))"
                :key="s" size="small" :bordered="true" class="cursor-pointer" @click="pickImage(s)"
              >{{ s }}</n-tag>
            </div>

            <n-form-item label="镜像地址" required :show-feedback="false">
              <n-input v-model:value="model.image" placeholder="如 nginx:latest，支持任意有效镜像引用">
                <template #prefix><DxIcon name="images" :size="14" /></template>
              </n-input>
            </n-form-item>
            <n-alert v-if="model.image && !imageInCenter" type="warning" :show-icon="true" class="text-xs">
              「{{ model.image }}」不在镜像中心。若本机引擎无此镜像，创建会失败，请先到
              <a class="text-[#006eff] font-medium" @click.prevent="router.push('/images')">镜像中心</a>
              拉取（中国大陆区域会自动走加速源）。
            </n-alert>
            <n-alert v-else-if="selectedImage" type="success" :show-icon="true" class="text-xs">
              已选用镜像中心镜像「{{ model.image }}」<template v-if="selectedImage.size_bytes">（{{ fmtSize(selectedImage.size_bytes) }}）</template>，创建后立即可启动。
            </n-alert>
          </div>
        </section>

        <!-- ③ 实例规格 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">3</span>实例规格
          </div>
          <div class="dx-card-body space-y-4">
            <div class="grid grid-cols-2 lg:grid-cols-4 gap-2">
              <button
                v-for="p in specPresets" :key="p.key" type="button"
                class="spec-tile" :class="{ active: activePreset === p.key }"
                @click="applyPreset(p)"
              >
                <div class="flex items-center justify-between">
                  <span class="font-medium text-[13px]">{{ p.name }}</span>
                  <span v-if="activePreset === p.key" class="w-4 h-4 rounded-full bg-[#006eff] text-white flex items-center justify-center">
                    <DxIcon name="check" :size="10" :stroke="3" />
                  </span>
                </div>
                <div class="text-[13px] mt-1.5 text-gray-700 dark:text-gray-200">{{ p.cpu }} 核 · {{ p.memory_mb >= 1024 ? `${p.memory_mb / 1024} GB` : `${p.memory_mb} MB` }}</div>
                <div class="text-[11px] text-gray-400 mt-0.5 truncate">{{ p.desc }}</div>
              </button>
            </div>
            <div class="rounded-lg bg-gray-50 dark:bg-gray-800/50 p-3.5 border border-gray-100 dark:border-gray-800">
              <div class="text-xs text-gray-500 mb-3 flex items-center gap-1">
                <DxIcon name="cpu" :size="13" />
                {{ activePreset === 'custom' ? '自定义规格' : '在预设基础上微调' }}
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <div class="text-[11px] text-gray-400 mb-1.5">CPU</div>
                  <n-input-number v-model:value="model.cpu" :min="0.25" :max="16" :step="0.25" class="w-full">
                    <template #suffix>核</template>
                  </n-input-number>
                </div>
                <div>
                  <div class="text-[11px] text-gray-400 mb-1.5">内存</div>
                  <n-input-number v-model:value="model.memory_mb" :min="128" :max="32768" :step="128" class="w-full">
                    <template #suffix>MB</template>
                  </n-input-number>
                </div>
                <div>
                  <div class="text-[11px] text-gray-400 mb-1.5">磁盘</div>
                  <n-input-number v-model:value="model.disk_gb" :min="1" :max="500" class="w-full">
                    <template #suffix>GB</template>
                  </n-input-number>
                </div>
              </div>
              <div class="text-[11px] text-gray-400 mt-2">磁盘为逻辑配额（Phase 11 落硬配额）；规格受组织配额约束</div>
            </div>
          </div>
        </section>

        <!-- ④ 网络与端口 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">4</span>网络与端口
          </div>
          <div class="dx-card-body space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <n-form-item label="所属网络" :show-feedback="false">
                <n-select
                  v-model:value="networkId"
                  :options="[{ label: '默认网络（bridge）', value: '' }, ...networks.map(n => ({ label: `${n.name}（${n.subnet || '无子网'}）`, value: String(n.id) }))]"
                />
              </n-form-item>
              <n-form-item label="静态 IP" :show-feedback="false">
                <n-input v-model:value="fixedIp" placeholder="可选，指定容器 IP" />
              </n-form-item>
            </div>
            <div>
              <div class="flex items-center justify-between mb-2">
                <div class="text-xs text-gray-500">端口映射（宿主机 → 容器）</div>
                <n-button size="tiny" dashed @click="addPort">+ 添加端口</n-button>
              </div>
              <div class="space-y-2">
                <div v-for="(p, i) in ports" :key="i" class="flex items-center gap-2 flex-wrap">
                  <n-input-number v-model:value="p.host_port" :min="0" :max="65535" placeholder="宿主端口" style="width: 140px" />
                  <span class="text-gray-400">→</span>
                  <n-input-number v-model:value="p.container_port" :min="0" :max="65535" placeholder="容器端口" style="width: 140px" />
                  <n-select v-model:value="p.protocol" :options="[{ label: 'tcp', value: 'tcp' }, { label: 'udp', value: 'udp' }]" style="width: 90px" />
                  <n-button size="tiny" quaternary type="error" @click="removePort(i)">移除</n-button>
                </div>
                <div class="text-[11px] text-gray-400">host_port 为 0 表示不映射；端口冲突会在创建时报错</div>
              </div>
            </div>
          </div>
        </section>

        <!-- ⑤ 存储 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">5</span>存储挂载
          </div>
          <div class="dx-card-body space-y-2">
            <div v-for="(m, i) in mounts" :key="i" class="flex items-center gap-2 flex-wrap">
              <n-select
                v-model:value="m.volume_id"
                :options="volumes.map(v => ({ label: `${v.name}（${v.capacity_gb}GB）`, value: v.id }))"
                placeholder="选择云磁盘" style="width: 220px"
              />
              <n-input v-model:value="m.target" placeholder="挂载路径，如 /data" style="width: 200px" />
              <n-checkbox v-model:checked="m.read_only">只读</n-checkbox>
              <n-button size="tiny" quaternary type="error" @click="removeMount(i)">移除</n-button>
            </div>
            <n-button size="tiny" dashed @click="addMount">+ 添加磁盘（变更挂载将重建容器，数据保留）</n-button>
            <div v-if="volumes.length === 0" class="text-[11px] text-gray-400">
              暂无云磁盘，可到「存储」页创建后挂载
            </div>
          </div>
        </section>

        <!-- ⑥ 高级设置 -->
        <section class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header flex items-center gap-2">
            <span class="step-no">6</span>高级设置
          </div>
          <div class="dx-card-body space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <n-form-item label="重启策略" :show-feedback="false">
                <n-select v-model:value="model.restart_policy" :options="[
                  { label: '不自动重启（默认）', value: 'no' },
                  { label: '除非手动停止（unless-stopped）', value: 'unless-stopped' },
                  { label: '总是重启（always）', value: 'always' },
                  { label: '失败时重试（on-failure）', value: 'on-failure' },
                ]" />
              </n-form-item>
              <n-form-item label="启动命令" :show-feedback="false">
                <n-input v-model:value="model.command_text" placeholder="空格分隔，如：sleep 3600（留空用镜像默认）" />
              </n-form-item>
            </div>
            <n-form-item label="环境变量" :show-feedback="false">
              <n-input v-model:value="envText" type="textarea" placeholder="每行一个：KEY=VALUE" :autosize="{ minRows: 2, maxRows: 5 }" />
            </n-form-item>
            <div class="flex items-center gap-2">
              <n-switch v-model:value="model.readonly_rootfs" size="small" />
              <span class="text-xs text-gray-600 dark:text-gray-300">只读根文件系统（更安全，需配合挂载磁盘写入数据）</span>
            </div>
          </div>
        </section>
      </div>

      <!-- 右侧：配置清单 -->
      <div class="xl:sticky xl:top-16 space-y-4 dx-fade-up dx-delay-2">
        <section class="dx-card">
          <div class="dx-card-header">配置清单</div>
          <div class="dx-card-body">
            <dl class="summary-list">
              <div class="row">
                <dt>实例名称</dt>
                <dd>{{ model.name || '未填写' }}</dd>
              </div>
              <div class="row">
                <dt>镜像</dt>
                <dd class="flex items-center gap-1">
                  <span class="truncate max-w-[170px]">{{ model.image || '未选择' }}</span>
                  <n-tag v-if="imageInCenter" size="tiny" type="success" :bordered="false">已就绪</n-tag>
                  <n-tag v-else-if="model.image" size="tiny" type="warning" :bordered="false">未拉取</n-tag>
                </dd>
              </div>
              <div class="row">
                <dt>规格</dt>
                <dd>{{ model.cpu }} 核 / {{ model.memory_mb >= 1024 ? `${(model.memory_mb / 1024).toFixed(model.memory_mb % 1024 === 0 ? 0 : 1)} GB` : `${model.memory_mb} MB` }} / {{ model.disk_gb }} GB</dd>
              </div>
              <div class="row">
                <dt>端口映射</dt>
                <dd>{{ validPorts.length > 0 ? validPorts.map(p => `${p.host_port}:${p.container_port}`).join('，') : '不映射' }}</dd>
              </div>
              <div class="row">
                <dt>挂载磁盘</dt>
                <dd>{{ validMounts.length > 0 ? `${validMounts.length} 块` : '无' }}</dd>
              </div>
              <div class="row">
                <dt>网络</dt>
                <dd>{{ networkId ? (networks.find(n => String(n.id) === networkId)?.name || '自定义') : '默认网络' }}</dd>
              </div>
              <div class="row">
                <dt>重启策略</dt>
                <dd>{{ { no: '不自动重启', 'unless-stopped': '除非手动停止', always: '总是重启', 'on-failure': '失败时重试' }[model.restart_policy] || model.restart_policy }}</dd>
              </div>
            </dl>
            <div class="mt-4 rounded-lg bg-gradient-to-r from-[#006eff]/10 to-[#006eff]/5 dark:from-[#006eff]/20 dark:to-[#006eff]/10 p-3 border border-[#006eff]/20">
              <div class="text-[11px] text-gray-500">预估费用（虚拟计费）</div>
              <div class="text-lg font-semibold text-[#006eff] mt-0.5">{{ hourlyCost }}</div>
              <div class="text-[11px] text-gray-400 mt-0.5">CPU ¥0.1/核时 + 内存 ¥0.05/GB时 + 磁盘 ¥0.01/GB时</div>
            </div>
            <div class="flex gap-2 mt-4">
              <n-button class="flex-1" @click="router.push('/ecs')">取消</n-button>
              <n-button class="flex-1" type="primary" :loading="loading" @click="submit">创建实例</n-button>
            </div>
          </div>
        </section>

        <n-alert type="info" :show-icon="true">
          <div class="text-xs leading-relaxed">
            安全基线（后端强制，不可绕过）：非特权容器、no-new-privileges、CapDrop ALL、PID 限制 256、CPU/内存配额内上限、禁止宿主网络与 docker.sock。
          </div>
        </n-alert>
      </div>
    </div>
  </div>
</template>

<style scoped>
.step-no {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #006eff, #00c2ff);
}

.image-tile {
  text-align: left;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s ease;
}
.dark .image-tile {
  background: transparent;
  border-color: #30363d;
}
.image-tile:hover {
  border-color: #006eff;
  box-shadow: 0 2px 8px rgba(0, 110, 255, 0.12);
}
.image-tile.active {
  border-color: #006eff;
  background: rgba(0, 110, 255, 0.04);
}
.dark .image-tile.active {
  background: rgba(0, 110, 255, 0.1);
}

.spec-tile {
  text-align: left;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s ease;
}
.dark .spec-tile {
  background: transparent;
  border-color: #30363d;
}
.spec-tile:hover {
  border-color: #006eff;
  box-shadow: 0 2px 8px rgba(0, 110, 255, 0.12);
}
.spec-tile.active {
  border-color: #006eff;
  background: rgba(0, 110, 255, 0.04);
  box-shadow: 0 0 0 1px #006eff inset;
}
.dark .spec-tile.active {
  background: rgba(0, 110, 255, 0.1);
}

.summary-list .row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 12px;
  padding: 7px 0;
  border-bottom: 1px dashed #f0f0f0;
}
.dark .summary-list .row {
  border-bottom-color: #21262d;
}
.summary-list dt {
  font-size: 12px;
  color: #9ca3af;
  flex-shrink: 0;
}
.summary-list dd {
  font-size: 12px;
  color: #374151;
  text-align: right;
  min-width: 0;
  word-break: break-all;
}
.dark .summary-list dd {
  color: #d1d5db;
}
.dx-tag-mini {
  display: inline-block;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 10px;
  line-height: 18px;
  background: rgba(0, 110, 255, 0.08);
  color: #006eff;
}
</style>
