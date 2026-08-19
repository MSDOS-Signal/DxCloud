<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import DxIcon from '~/components/DxIcon.vue'
import PageHero from '~/components/PageHero.vue'
import { api } from '~/services/http'
import { useAuthStore } from '~/stores/auth'

definePageMeta({ title: '个人信息' })

const message = useMessage()
const auth = useAuthStore()

interface SessionRow {
  jti: string
  expires_in: number
  created_at: string
}

const savingProfile = ref(false)
const profileForm = reactive({ nickname: '', avatar_url: '' })

const pwdForm = reactive({ old_password: '', new_password: '', confirm: '' })
const savingPwd = ref(false)
const sessions = ref<SessionRow[]>([])
const loadingSessions = ref(false)

// ---------- 头像 ----------
const presetAvatars = [
  'linear-gradient(135deg,#006eff,#00c2ff)',
  'linear-gradient(135deg,#722ed1,#b37feb)',
  'linear-gradient(135deg,#13c2c2,#5cdbd3)',
  'linear-gradient(135deg,#52c41a,#95de64)',
  'linear-gradient(135deg,#fa8c16,#ffc069)',
  'linear-gradient(135deg,#f5222d,#ff7875)',
  'linear-gradient(135deg,#eb2f96,#ff85c0)',
  'linear-gradient(135deg,#2f54eb,#85a5ff)',
]

function presetDataURI(i: number) {
  const first = (auth.nickname || auth.username || 'D').charAt(0).toUpperCase()
  const [c1, c2] = presetAvatars[i].match(/#[0-9a-f]{6}/gi) || ['#006eff', '#00c2ff']
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${c1}"/><stop offset="1" stop-color="${c2}"/></linearGradient></defs><rect width="128" height="128" fill="url(#g)"/><text x="64" y="82" font-family="-apple-system,'Segoe UI','PingFang SC','Microsoft YaHei',sans-serif" font-size="56" font-weight="600" fill="#fff" text-anchor="middle">${first}</text></svg>`
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
}

const avatarSrc = computed(() => profileForm.avatar_url || presetDataURI(0))

function pickPreset(i: number) {
  profileForm.avatar_url = presetDataURI(i)
}

const fileInput = ref<HTMLInputElement | null>(null)

function triggerUpload() {
  fileInput.value?.click()
}

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    message.error('请选择图片文件')
    return
  }
  if (file.size > 4 * 1024 * 1024) {
    message.error('图片不能超过 4MB')
    return
  }
  const reader = new FileReader()
  reader.onload = () => {
    const img = new Image()
    img.onload = () => {
      // 压缩到 128x128 圆形裁剪，JPEG 0.8 保证 base64 足够小
      const size = 128
      const canvas = document.createElement('canvas')
      canvas.width = size
      canvas.height = size
      const ctx = canvas.getContext('2d')!
      const scale = Math.max(size / img.width, size / img.height)
      const w = img.width * scale
      const h = img.height * scale
      ctx.drawImage(img, (size - w) / 2, (size - h) / 2, w, h)
      const dataURI = canvas.toDataURL('image/jpeg', 0.8)
      if (dataURI.length > 32768) {
        message.error('压缩后图片仍过大，请换一张')
        return
      }
      profileForm.avatar_url = dataURI
      message.success('头像已预览，点击保存生效')
    }
    img.src = reader.result as string
  }
  reader.readAsDataURL(file)
  ;(e.target as HTMLInputElement).value = ''
}

// ---------- 保存资料 ----------
async function saveProfile() {
  if (!profileForm.nickname.trim()) {
    message.error('昵称不能为空')
    return
  }
  savingProfile.value = true
  try {
    await api.put('/auth/profile', {
      nickname: profileForm.nickname.trim(),
      avatar_url: profileForm.avatar_url.trim(),
    })
    await auth.fetchMe()
    profileForm.nickname = auth.user?.nickname || ''
    profileForm.avatar_url = auth.user?.avatar_url || ''
    message.success('个人信息已更新')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    savingProfile.value = false
  }
}

// ---------- 修改密码 ----------
const pwdStrength = computed(() => {
  const p = pwdForm.new_password
  if (!p) return 0
  let score = 0
  if (p.length >= 8) score++
  if (/[a-z]/.test(p) && /[A-Z]/.test(p)) score++
  if (/\d/.test(p)) score++
  if (/[^\w]/.test(p)) score++
  return Math.min(score, 4)
})

const strengthText = computed(() => ['', '弱', '一般', '较强', '强'][pwdStrength.value])
const strengthColor = computed(() => ['', '#f53f3f', '#ff7d00', '#006eff', '#00b42a'][pwdStrength.value])

async function changePassword() {
  if (!pwdForm.old_password || !pwdForm.new_password) {
    message.error('请填写完整')
    return
  }
  if (pwdForm.new_password.length < 8) {
    message.error('新密码至少 8 位')
    return
  }
  if (pwdForm.new_password !== pwdForm.confirm) {
    message.error('两次输入的新密码不一致')
    return
  }
  savingPwd.value = true
  try {
    await api.put('/auth/password', {
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
    })
    message.success('密码已修改，请重新登录', { duration: 3000 })
    setTimeout(async () => {
      await auth.logout()
      window.location.href = '/login'
    }, 1200)
  } catch (e) {
    message.error(e instanceof Error ? e.message : '修改失败')
  } finally {
    savingPwd.value = false
  }
}

// ---------- 会话管理 ----------
async function loadSessions() {
  loadingSessions.value = true
  try {
    const data = await api.get<SessionRow[] | { items: SessionRow[] }>('/auth/sessions')
    const list = Array.isArray(data) ? data : (data.items || [])
    sessions.value = [...list].sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
  } catch {
    sessions.value = []
  } finally {
    loadingSessions.value = false
  }
}

async function killSession(row: SessionRow) {
  try {
    await api.del(`/auth/sessions/${row.jti}`)
    message.success('会话已下线')
    loadSessions()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

function fmtTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN', { hour12: false })
}

function fmtRemain(sec: number) {
  if (!sec || sec <= 0) return '已过期'
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  if (d > 0) return `${d} 天 ${h} 小时`
  return `${h} 小时`
}

onMounted(async () => {
  if (!auth.user) await auth.fetchMe()
  profileForm.nickname = auth.user?.nickname || ''
  profileForm.avatar_url = auth.user?.avatar_url || ''
  loadSessions()
})
</script>

<template>
  <div class="space-y-4">
    <PageHero
      icon="users"
      title="个人信息"
      description="管理你的账号资料、登录密码与会话安全"
    >
      <template #action>
        <div class="hero-user">
          <img :src="avatarSrc" alt="" class="hero-avatar">
          <div>
            <div class="hero-name">{{ auth.nickname }}</div>
            <div class="hero-roles">{{ (auth.user?.role_names || []).join(' · ') || '普通用户' }}</div>
          </div>
        </div>
      </template>
    </PageHero>

    <div class="profile-grid">
      <!-- 头像卡 -->
      <div class="dx-card avatar-card dx-fade-up">
        <div class="dx-card-header">
          <span class="card-title"><DxIcon name="users" :size="15" /> 头像</span>
        </div>
        <div class="dx-card-body">
          <div class="avatar-preview-wrap">
            <div class="avatar-ring">
              <img :src="avatarSrc" alt="头像" class="avatar-preview">
            </div>
            <div class="avatar-actions">
              <button class="avatar-btn primary" @click="triggerUpload">
                <DxIcon name="images" :size="14" /> 本地上传
              </button>
              <input ref="fileInput" type="file" accept="image/*" hidden @change="onFileChange">
            </div>
          </div>

          <div class="preset-label">预设头像</div>
          <div class="preset-grid">
            <button
              v-for="(g, i) in presetAvatars"
              :key="i"
              class="preset-item"
              :class="{ active: profileForm.avatar_url === presetDataURI(i) }"
              :style="{ background: g }"
              @click="pickPreset(i)"
            >
              {{ (auth.nickname || auth.username || 'D').charAt(0).toUpperCase() }}
            </button>
          </div>

          <div class="preset-label">或输入图片链接</div>
          <input
            v-model="profileForm.avatar_url"
            type="text"
            class="avatar-url-input"
            placeholder="https://example.com/avatar.png"
          >
        </div>
      </div>

      <div class="right-col">
        <!-- 基本信息卡 -->
        <div class="dx-card dx-fade-up dx-delay-1">
          <div class="dx-card-header">
            <span class="card-title"><DxIcon name="info" :size="15" /> 基本信息</span>
          </div>
          <div class="dx-card-body">
            <div class="form-row">
              <label class="form-label">用户名</label>
              <div class="readonly-field">
                {{ auth.user?.username }}
                <span class="dx-tag dx-tag-blue">不可修改</span>
              </div>
            </div>
            <div class="form-row">
              <label class="form-label">邮箱</label>
              <div class="readonly-field">{{ auth.user?.email }}</div>
            </div>
            <div class="form-row">
              <label class="form-label">角色</label>
              <div class="readonly-field">
                <span v-for="r in (auth.user?.role_names || [])" :key="r" class="dx-tag dx-tag-blue">{{ r }}</span>
                <span v-if="!(auth.user?.role_names || []).length" class="dx-tag dx-tag-gray">普通用户</span>
              </div>
            </div>
            <div class="form-row">
              <label class="form-label">昵称 <span class="required">*</span></label>
              <input v-model="profileForm.nickname" type="text" class="form-input" placeholder="输入新昵称" maxlength="32">
            </div>
            <button class="save-btn" :disabled="savingProfile" @click="saveProfile">
              <DxIcon name="refresh" :size="14" /> {{ savingProfile ? '保存中…' : '保存资料' }}
            </button>
          </div>
        </div>

        <!-- 修改密码卡 -->
        <div class="dx-card dx-fade-up dx-delay-2">
          <div class="dx-card-header">
            <span class="card-title"><DxIcon name="security" :size="15" /> 修改密码</span>
          </div>
          <div class="dx-card-body">
            <div class="form-row">
              <label class="form-label">原密码</label>
              <input v-model="pwdForm.old_password" type="password" class="form-input" placeholder="输入当前密码">
            </div>
            <div class="form-row">
              <label class="form-label">新密码</label>
              <input v-model="pwdForm.new_password" type="password" class="form-input" placeholder="8-72 位，建议字母+数字+符号">
              <div v-if="pwdForm.new_password" class="strength-bar">
                <div class="strength-track">
                  <div class="strength-fill" :style="{ width: `${pwdStrength * 25}%`, background: strengthColor }" />
                </div>
                <span class="strength-text" :style="{ color: strengthColor }">{{ strengthText }}</span>
              </div>
            </div>
            <div class="form-row">
              <label class="form-label">确认新密码</label>
              <input v-model="pwdForm.confirm" type="password" class="form-input" placeholder="再次输入新密码">
            </div>
            <button class="save-btn" :disabled="savingPwd" @click="changePassword">
              <DxIcon name="security" :size="14" /> {{ savingPwd ? '提交中…' : '修改密码' }}
            </button>
            <p class="pwd-tip">修改成功后所有会话将失效，需重新登录</p>
          </div>
        </div>

        <!-- 会话管理卡 -->
        <div class="dx-card dx-fade-up dx-delay-3">
          <div class="dx-card-header">
            <span class="card-title"><DxIcon name="server" :size="15" /> 登录会话</span>
            <button class="refresh-mini" title="刷新" @click="loadSessions">
              <DxIcon name="refresh" :size="13" />
            </button>
          </div>
          <div class="dx-card-body">
            <n-skeleton v-if="loadingSessions" :repeat="2" text />
            <div v-else-if="sessions.length === 0" class="empty-sessions">暂无活跃会话</div>
            <div v-else class="session-list">
              <div class="session-count">共 {{ sessions.length }} 个活跃会话{{ sessions.length > 10 ? '，展示最近 10 个' : '' }}</div>
              <div v-for="s in sessions.slice(0, 10)" :key="s.jti" class="session-item">
                <div class="session-info">
                  <span class="session-ua">会话 {{ s.jti.slice(0, 8) }}…</span>
                  <span class="session-time">登录于 {{ fmtTime(s.created_at) }} · 剩余 {{ fmtRemain(s.expires_in) }}</span>
                </div>
                <button class="kick-btn" @click="killSession(s)">下线</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hero-user {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
}
.hero-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.85);
  object-fit: cover;
}
.hero-name {
  font-size: 15px;
  font-weight: 600;
}
.hero-roles {
  font-size: 11px;
  opacity: 0.85;
}

.profile-grid {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 16px;
  align-items: start;
}
.right-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
}

/* 头像卡 */
.avatar-preview-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0 4px;
}
.avatar-ring {
  width: 112px;
  height: 112px;
  border-radius: 50%;
  padding: 3px;
  background: conic-gradient(#006eff, #00c2ff, #722ed1, #006eff);
  animation: ring-rotate 6s linear infinite;
}
@keyframes ring-rotate {
  to { transform: rotate(360deg); }
}
.avatar-preview {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #fff;
  background: #f0f2f5;
}
html.dark .avatar-preview {
  border-color: #161b22;
}
.avatar-actions {
  margin-top: 14px;
  display: flex;
  gap: 8px;
}
.avatar-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: 1px solid #d4e5ff;
  background: #fff;
  color: #006eff;
  transition: all 0.2s;
}
.avatar-btn:hover {
  border-color: #006eff;
  background: #f0f7ff;
}
.avatar-btn.primary {
  background: #006eff;
  border-color: #006eff;
  color: #fff;
}
.avatar-btn.primary:hover {
  background: #0052d9;
}
html.dark .avatar-btn {
  background: #161b22;
  border-color: #2b4a75;
  color: #3d8bff;
}

.preset-label {
  font-size: 12px;
  color: #86909c;
  margin: 14px 0 8px;
}
.preset-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}
.preset-item {
  aspect-ratio: 1;
  border-radius: 50%;
  border: 2px solid transparent;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.preset-item:hover {
  transform: scale(1.1);
}
.preset-item.active {
  border-color: #006eff;
  box-shadow: 0 0 0 3px rgba(0, 110, 255, 0.2);
}
.avatar-url-input {
  width: 100%;
  border: 1px solid #e0e6ed;
  border-radius: 6px;
  padding: 8px 10px;
  font-size: 12px;
  outline: none;
  transition: border-color 0.2s;
}
.avatar-url-input:focus {
  border-color: #006eff;
}
html.dark .avatar-url-input {
  background: #0d1117;
  border-color: #30363d;
  color: #e6edf3;
}

/* 表单 */
.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
}
.form-label {
  font-size: 13px;
  color: #4e5969;
  font-weight: 500;
}
.required {
  color: #f53f3f;
}
html.dark .form-label {
  color: #8b949e;
}
.form-input {
  border: 1px solid #e0e6ed;
  border-radius: 6px;
  padding: 9px 12px;
  font-size: 13px;
  outline: none;
  transition: all 0.2s;
  background: #fff;
}
.form-input:focus {
  border-color: #006eff;
  box-shadow: 0 0 0 2px rgba(0, 110, 255, 0.1);
}
html.dark .form-input {
  background: #0d1117;
  border-color: #30363d;
  color: #e6edf3;
}
.readonly-field {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  background: #f7f9fc;
  border-radius: 6px;
  font-size: 13px;
  color: #4e5969;
}
html.dark .readonly-field {
  background: #0d1117;
  color: #8b949e;
}

.strength-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}
.strength-track {
  flex: 1;
  height: 4px;
  background: #e8e8e8;
  border-radius: 2px;
  overflow: hidden;
}
html.dark .strength-track {
  background: #30363d;
}
.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition: all 0.3s;
}
.strength-text {
  font-size: 11px;
  min-width: 24px;
}

.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 20px;
  background: #006eff;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.save-btn:hover:not(:disabled) {
  background: #0052d9;
}
.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.pwd-tip {
  margin-top: 10px;
  font-size: 12px;
  color: #86909c;
}

/* 会话 */
.refresh-mini {
  border: none;
  background: transparent;
  color: #86909c;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}
.refresh-mini:hover {
  color: #006eff;
  background: rgba(0, 110, 255, 0.08);
}
.empty-sessions {
  text-align: center;
  color: #86909c;
  font-size: 13px;
  padding: 16px 0;
}
.session-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.session-count {
  font-size: 12px;
  color: #86909c;
}
.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  background: #f7f9fc;
  border-radius: 6px;
}
html.dark .session-item {
  background: #0d1117;
}
.session-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.session-ua {
  font-size: 13px;
  color: #1f2329;
  font-weight: 500;
}
html.dark .session-ua {
  color: #e6edf3;
}
.session-time {
  font-size: 11px;
  color: #86909c;
}
.kick-btn {
  border: 1px solid #ffd6d6;
  background: #fff;
  color: #f53f3f;
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}
.kick-btn:hover {
  background: #f53f3f;
  color: #fff;
  border-color: #f53f3f;
}
html.dark .kick-btn {
  background: #161b22;
  border-color: #5c2e2e;
}

@media (max-width: 900px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
