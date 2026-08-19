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
const model = reactive({ username: '', password: '' })

const rules: FormRules = {
  username: { required: true, message: '请输入用户名', trigger: ['input', 'blur'] },
  password: { required: true, message: '请输入密码', trigger: ['input', 'blur'] },
}

async function handleLogin() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    await auth.login(model.username, model.password)
    message.success(`欢迎回来，${auth.nickname}`)
    router.push('/dashboard')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-form dx-card p-7 dx-fade-up" style="box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);">
    <div class="mb-5">
      <div class="text-base font-semibold text-[#1f2329] dark:text-[#e6edf3]">欢迎登录多晓云</div>
      <div class="mt-1 text-xs text-[#86909c]">登录后进入多晓云 DxCloud 控制台</div>
    </div>

    <n-form ref="formRef" :model="model" :rules="rules" label-placement="top" size="medium" @keyup.enter="handleLogin">
      <n-form-item path="username" label="用户名">
        <n-input v-model:value="model.username" placeholder="请输入用户名" clearable>
          <template #prefix>
            <DxIcon name="users" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="password" label="密码">
        <n-input
          v-model:value="model.password"
          type="password"
          show-password-on="click"
          placeholder="请输入密码"
        >
          <template #prefix>
            <DxIcon name="permissions" :size="16" class="text-[#86909c]" />
          </template>
        </n-input>
      </n-form-item>

      <n-button
        type="primary"
        block
        :loading="loading"
        :disabled="loading"
        class="mt-2"
        @click="handleLogin"
      >
        {{ loading ? '登录中...' : '登 录' }}
      </n-button>
    </n-form>

    <div class="mt-4 flex items-center justify-between text-xs">
      <span class="text-[#86909c]">体验账号 <code class="text-[#006eff] font-medium">admin / Admin@123456</code></span>
      <NuxtLink to="/register" class="text-[#006eff] font-medium hover:text-[#0052d9] transition-colors">
        注册新账号
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
