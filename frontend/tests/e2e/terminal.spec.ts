import { expect, test } from '@playwright/test'

const apiBase = process.env.PLAYWRIGHT_API_BASE || 'http://localhost/api/v1'

test('网页终端真实连接并回显命令', async ({ browser, request }) => {
  const stamp = Date.now()
  const username = `term-${stamp}`
  const password = 'Term@123456'
  const registered = await request.post(`${apiBase}/auth/register`, {
    data: { username, email: `${username}@dx.dev`, password },
  })
  expect(registered.ok()).toBeTruthy()
  const login = await request.post(`${apiBase}/auth/login`, {
    data: { username, password },
  })
  expect(login.ok()).toBeTruthy()
  const pair = (await login.json()).data
  const token = pair.access_token as string
  const authHeaders = { Authorization: `Bearer ${token}` }

  const created = await request.post(`${apiBase}/ecs`, {
    headers: authHeaders,
    data: {
      name: `ui-terminal-${stamp}`,
      image: 'busybox:latest',
      cpu: 1,
      memory_mb: 256,
      command: ['sleep', '3600'],
    },
  })
  expect(created.ok()).toBeTruthy()
  const instance = (await created.json()).data

  let running = false
  for (let i = 0; i < 30; i++) {
    const check = await request.get(`${apiBase}/ecs/${instance.id}`, { headers: authHeaders })
    const state = (await check.json()).data?.observed_state
    if (state === 'running') {
      running = true
      break
    }
    await new Promise((resolve) => setTimeout(resolve, 1000))
  }
  expect(running).toBeTruthy()

  const context = await browser.newContext()
  await context.addInitScript((value) => {
    localStorage.setItem('cloudx_access', value.access_token)
    localStorage.setItem('cloudx_refresh', value.refresh_token)
  }, pair)
  const page = await context.newPage()

  await page.goto(`/ecs/${instance.id}/terminal`)
  await expect(page.getByText('已连接', { exact: true })).toBeVisible({ timeout: 15_000 })
  await page.locator('.xterm-screen').click()
  await page.keyboard.type('echo dxcloud-terminal-ok\r')
  await expect(page.locator('.xterm-screen')).toContainText('dxcloud-terminal-ok', { timeout: 10_000 })
  await context.close()

  await request.delete(`${apiBase}/ecs/${instance.id}`, { headers: authHeaders })
})
