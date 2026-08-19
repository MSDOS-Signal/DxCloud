<script setup lang="ts">
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'

definePageMeta({ layout: 'auth' })

const message = useMessage()
const auth = useAuthStore()
const router = useRouter()

const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const model = reactive({ username: '', email: '', password: '', confirm: '' })

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名（3-32 位）', trigger: ['input', 'blur'] },
    { min: 3, max: 32, message: '用户名长度 3-32 位', trigger: ['input', 'blur'] },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: ['input', 'blur'] },
    { type: 'email', message: '邮箱格式不正确', trigger: ['input', 'blur'] },
  ],
  password: [
    { required: true, message: '请输入密码（至少 8 位）', trigger: ['input', 'blur'] },
    { min: 8, message: '密码至少 8 位', trigger: ['input', 'blur'] },
  ],
  confirm: [
    {
      required: true,
      validator: (_rule: unknown, value: string) => value === model.password || new Error('两次密码不一致'),
      trigger: ['input', 'blur'],
    },
  ],
}

async function handleRegister() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    await auth.register(model.username, model.email, model.password)
    message.success('注册成功，已自动登录')
    router.push('/dashboard')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-form dx-card p-7 dx-fade-up" style="box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);">
    <div class="mb-5">
      <div class="text-base font-semibold text-[#1f2329] dark:text-[#e6edf3]">创建账号</div>
      <div class="mt-1 text-xs text-[#86909c]">注册即开通默认空间（5 实例 / 8 核免费配额）</div>
    </div>

    <n-form ref="formRef" :model="model" :rules="rules" label-placement="top" size="medium" @keyup.enter="handleRegister">
      <n-form-item path="username" label="用户名">
        <n-input v-model:value="model.username" placeholder="3-32 位，用于登录" clearable>
          <template #prefix>
            <DxIcon name="users" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="email" label="邮箱">
        <n-input v-model:value="model.email" placeholder="you@example.com" clearable>
          <template #prefix>
            <DxIcon name="domains" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="password" label="密码">
        <n-input v-model:value="model.password" type="password" show-password-on="click" placeholder="至少 8 位">
          <template #prefix>
            <DxIcon name="permissions" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="confirm" label="确认密码">
        <n-input v-model:value="model.confirm" type="password" show-password-on="click" placeholder="再次输入密码">
          <template #prefix>
            <DxIcon name="check-circle" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>

      <n-button
        type="primary"
        block
        :loading="loading"
        :disabled="loading"
        class="mt-2"
        @click="handleRegister"
      >
        {{ loading ? '注册中...' : '注册并登录' }}
      </n-button>
    </n-form>

    <div class="mt-4 text-center text-xs text-[#86909c]">
      已有账号？<NuxtLink to="/login" class="text-[#006eff] font-medium hover:text-[#0052d9] transition-colors">
        去登录
      </NuxtLink>
    </div>
  </div>
</template>

<style scoped>
/* n-input: 腾讯云风格边框 */
.auth-form :deep(.n-input) {
  --n-border: 1px solid #c9cdd4 !important;
  --n-border-hover: 1px solid #006eff !important;
  --n-border-focus: 1px solid #006eff !important;
  --n-border-radius: 3px !important;
  --n-box-shadow-focus: unset !important;
}
.auth-form :deep(.n-input .n-input__border, .n-input .n-input__border-hover) {
  border-color: #c9cdd4;
}
.auth-form :deep(.n-input--focus .n-input__border) {
  border-color: #006eff;
}
html.dark .auth-form :deep(.n-input) {
  --n-border: 1px solid #30363d !important;
  --n-border-hover: 1px solid #006eff !important;
  --n-border-focus: 1px solid #006eff !important;
}
html.dark .auth-form :deep(.n-input .n-input__border, .n-input .n-input__border-hover) {
  border-color: #30363d;
}

/* n-button: 腾讯云蓝色主按钮 */
.auth-form :deep(.n-button--primary-type) {
  --n-color: #006eff !important;
  --n-color-hover: #0052d9 !important;
  --n-color-pressed: #003aab !important;
  --n-color-focus: #0052d9 !important;
  --n-border: 1px solid #006eff !important;
  --n-border-hover: 1px solid #0052d9 !important;
  --n-border-pressed: 1px solid #003aab !important;
  --n-border-focus: 1px solid #0052d9 !important;
  --n-border-radius: 3px !important;
  --n-font-size: 13px !important;
  --n-height: 36px !important;
  font-weight: 500;
}
</style>
