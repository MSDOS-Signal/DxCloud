param(
    [ValidateSet('cn', 'global', '')]
    [string]$Region = '',
    [switch]$SkipDockerCheck,
    [switch]$NoBuild,
    [switch]$ConfigureDaemon
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Write-Step([string]$Text) {
    Write-Host "== $Text ==" -ForegroundColor Cyan
}

function Update-EnvValue([string]$Name, [string]$Value) {
    $EnvPath = Join-Path $Root '.env'
    if (-not (Test-Path -LiteralPath $EnvPath)) {
        Copy-Item -LiteralPath (Join-Path $Root '.env.example') -Destination $EnvPath
    }
    $Lines = @(Get-Content -LiteralPath $EnvPath -ErrorAction SilentlyContinue)
    $Found = $false
    for ($i = 0; $i -lt $Lines.Count; $i++) {
        if ($Lines[$i] -match "^\s*${Name}=") {
            $Lines[$i] = "${Name}=${Value}"
            $Found = $true
            break
        }
    }
    if (-not $Found) {
        $Lines += "${Name}=${Value}"
    }
    Set-Content -LiteralPath $EnvPath -Value $Lines -Encoding UTF8
}

if (-not $SkipDockerCheck) {
    $Docker = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $Docker) {
        Write-Host ''
        Write-Host '未检测到 Docker。请先安装 Docker Desktop：' -ForegroundColor Yellow
        Write-Host '  https://www.docker.com/products/docker-desktop/' -ForegroundColor White
        Write-Host ''
        Write-Host '安装完成后打开 Docker Desktop，等待 Engine running，再重新运行本脚本。' -ForegroundColor DarkGray
        exit 1
    }
    docker version *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Host 'Docker 已安装但服务未启动，请打开 Docker Desktop 并等待 Engine running。' -ForegroundColor Yellow
        exit 1
    }
}

if ($Region -eq '') {
    $Region = if ($env:DX_REGION) { $env:DX_REGION.ToLower() } else { 'cn' }
}
if ($Region -ne 'cn' -and $Region -ne 'global') {
    $Region = 'cn'
}

Write-Step '区域配置'
if ($Region -eq 'cn') {
    Write-Host '当前区域：中国大陆。使用镜像加速前缀，启动栈所需的基础镜像会优先从国内可用源拉取。' -ForegroundColor Green
    $Mirror = 'hub.rat.dev'
    Update-EnvValue 'MYSQL_IMAGE' "$Mirror/library/mysql:8.0"
    Update-EnvValue 'REDIS_IMAGE' "$Mirror/library/redis:7-alpine"
    Update-EnvValue 'REGISTRY_IMAGE' "$Mirror/library/registry:2.8"
    Update-EnvValue 'TRAEFIK_IMAGE' "$Mirror/library/traefik:v3"
    Update-EnvValue 'GO_BUILD_IMAGE' "$Mirror/library/golang:1.25-alpine"
    Update-EnvValue 'ALPINE_RUNTIME_IMAGE' "$Mirror/library/alpine:3.20"
    Update-EnvValue 'NODE_IMAGE' "$Mirror/library/node:20-alpine"
} else {
    Write-Host '当前区域：非中国大陆。使用 Docker Hub 官方基础镜像。' -ForegroundColor Green
    Update-EnvValue 'MYSQL_IMAGE' 'mysql:8.0'
    Update-EnvValue 'REDIS_IMAGE' 'redis:7-alpine'
    Update-EnvValue 'REGISTRY_IMAGE' 'registry:2.8'
    Update-EnvValue 'TRAEFIK_IMAGE' 'traefik:v3'
    Update-EnvValue 'GO_BUILD_IMAGE' 'golang:1.25-alpine'
    Update-EnvValue 'ALPINE_RUNTIME_IMAGE' 'alpine:3.20'
    Update-EnvValue 'NODE_IMAGE' 'node:20-alpine'
}

if ($ConfigureDaemon -or ($Region -eq 'cn' -and -not (Test-Path -LiteralPath '.dxcloud\daemon.json'))) {
    $DaemonDir = Join-Path $Root '.dxcloud'
    if (-not (Test-Path -LiteralPath $DaemonDir)) {
        New-Item -ItemType Directory -Path $DaemonDir | Out-Null
    }
    $Daemon = @"
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1ms.run",
    "https://hub.rat.dev"
  ]
}
"@
    Set-Content -LiteralPath (Join-Path $DaemonDir 'daemon.json') -Value $Daemon -Encoding UTF8
    Write-Host 'Docker Desktop 用户可继续在 Settings → Docker Engine 中添加 .dxcloud/daemon.json 的 registry-mirrors，然后 Apply & Restart。' -ForegroundColor DarkGray
}

Write-Step '启动服务'
$Args = @('compose', 'up', '-d')
if (-not $NoBuild) {
    $Args += '--build'
}
& docker @Args
if ($LASTEXITCODE -ne 0) {
    Write-Host '启动失败，请执行 docker compose logs backend 查看日志。' -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host ''
Write-Host '启动完成，请在浏览器打开 http://localhost' -ForegroundColor Green
Write-Host '初始管理员：admin / Admin@123456（首次登录后请立即修改）' -ForegroundColor White
