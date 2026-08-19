import { expect, test, type APIRequestContext, type Page } from '@playwright/test'

const apiBase = process.env.PLAYWRIGHT_API_BASE || 'http://localhost/api/v1'

async function loginAdmin(request: APIRequestContext) {
  const response = await request.post(`${apiBase}/auth/login`, {
    data: { username: 'admin', password: 'Admin@123456' },
  })
  expect(response.ok()).toBeTruthy()
  const body = await response.json()
  return body.data.access_token as string
}

async function loginUi(page: Page) {
  await page.goto('/dashboard')
  await expect(page.getByPlaceholder('请输入用户名')).toBeVisible()
  await expect(page).toHaveURL(/\/login$/)
  await page.getByPlaceholder('请输入用户名').fill('admin')
  await page.getByPlaceholder('请输入密码').fill('Admin@123456')
  await page.getByRole('button', { name: '登 录' }).click()
  await expect(page).toHaveURL(/\/dashboard$/)
}

test('登录、图表渲染、中文权限、镜像联想与退出跳转', async ({ page, request }) => {
  await loginUi(page)

  await expect(page.getByText('总览 Dashboard', { exact: true })).toBeVisible()
  await expect(page.locator('canvas')).toHaveCount(4)

  const canvasCheck = await page.evaluate(() => {
    return Array.from(document.querySelectorAll('canvas')).map((canvas) => {
      const ctx = canvas.getContext('2d')
      if (!ctx) return { w: canvas.width, h: canvas.height, nonBlank: false }
      const image = ctx.getImageData(0, 0, canvas.width, canvas.height).data
      let colored = 0
      for (let i = 3; i < image.length; i += 160) {
        if (image[i] > 0) colored += 1
      }
      return { w: canvas.width, h: canvas.height, nonBlank: colored > 8 }
    })
  })
  expect(canvasCheck.length).toBeGreaterThanOrEqual(4)
  expect(canvasCheck.every((item) => item.w > 0 && item.h > 0 && item.nonBlank)).toBeTruthy()

  await page.goto('/iam/permissions')
  await expect(page.getByText('创建云主机', { exact: true })).toBeVisible()
  await expect(page.getByText('ecs:create', { exact: true })).toHaveCount(0)

  await page.goto('/images')
  await page.getByRole('button', { name: '拉取镜像' }).click()
  await page.getByPlaceholder('输入 ngi、mysql、openjdk 等关键词，自动联想').fill('ngi')
  await expect(page.getByText('nginx', { exact: true }).first()).toBeVisible()

  await page.goto('/settings')
  await expect(page.getByText('中国大陆', { exact: true })).toBeVisible()
  await expect(page.getByText('非中国大陆', { exact: true })).toBeVisible()

  await page.getByText('Administrator', { exact: true }).click()
  await page.getByText('退出登录', { exact: true }).click()
  await expect(page).toHaveURL(/\/login$/)

  const token = await loginAdmin(request)
  expect(token.length).toBeGreaterThan(20)
})

test('移动端仪表盘无横向溢出并可展开菜单', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await loginUi(page)
  await expect(page.getByRole('button', { name: '打开菜单' })).toBeVisible()
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth)
  expect(overflow).toBeLessThanOrEqual(2)
  await page.getByRole('button', { name: '打开菜单' }).click()
  await expect(page.getByText('监控', { exact: true }).first()).toBeVisible()
})
