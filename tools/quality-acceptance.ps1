# DxCloud 质量验收：Go 单元测试 + Nuxt 生产构建 + Playwright 浏览器端回归。
$ErrorActionPreference = 'Stop'
$Root = if ($PSScriptRoot) { (Resolve-Path (Join-Path $PSScriptRoot '..')).Path } else { (Get-Location).Path }
Set-Location $Root

Write-Host '[1/3] Go unit tests'
Push-Location (Join-Path $Root 'backend')
go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Go unit tests failed' }
Pop-Location

Write-Host '[2/3] Nuxt production build'
Push-Location (Join-Path $Root 'frontend')
npm ci --no-audit --no-fund
npm run build
if ($LASTEXITCODE -ne 0) { throw 'frontend build failed' }

Write-Host '[3/3] Playwright browser regression'
$frontendContainer = docker compose ps -q frontend
if ($frontendContainer) {
    docker compose restart frontend | Out-Null
    $deadline = (Get-Date).AddSeconds(90)
    do {
        $status = docker inspect --format '{{.State.Health.Status}}' cloud-web
        if ($status -eq 'healthy') { break }
        Start-Sleep -Seconds 3
    } while ((Get-Date) -lt $deadline)
}
$browserRoot = Join-Path $env:LOCALAPPDATA 'ms-playwright'
if (-not (Test-Path $browserRoot) -or -not (Get-ChildItem -LiteralPath $browserRoot -Directory -Filter 'chromium-*' -ErrorAction SilentlyContinue | Select-Object -First 1)) {
    npx playwright install chromium
}
npm run test:e2e
if ($LASTEXITCODE -ne 0) { throw 'browser regression failed' }
Pop-Location

Write-Host 'RESULT: ALL QUALITY CHECKS PASSED' -ForegroundColor Green
