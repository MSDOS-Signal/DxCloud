# Phase 12 验收：生产部署（HTTPS / 非 root / 资源限额 / 备份恢复 / 优雅停机）
# 前置：生产镜像已构建（docker compose -f docker-compose.prod.yml build）、证书已生成（deploy/certs）
# 本脚本会切换 dev→prod 栈并最终恢复 dev 栈
$ErrorActionPreference = "Continue"   # 原生命令 stderr 不视为致命（结果以 Report 判定）
# 依赖工作目录为项目根（本脚本按脚本文件方式执行时 $PSScriptRoot 亦有效）
$root = if ($PSScriptRoot) { (Resolve-Path (Join-Path $PSScriptRoot "..")).Path } else { (Get-Location).Path }
Set-Location $root
$results = New-Object System.Collections.Generic.List[string]
$global:FAILED = $false

function Report([string]$label, [bool]$ok, [string]$detail) {
    $mark = if ($ok) { "[PASS]" } else { $global:FAILED = $true; "[FAIL]" }
    $line = "$mark $label - $detail"
    $results.Add($line)
    Write-Host $line
}

# ---------- P0 切换生产栈 ----------
Write-Host "== P0 切换 dev → prod（构建 prod 镜像 + 启动） =="
docker compose -f docker-compose.prod.yml build 2>&1 | Select-Object -Last 2
docker compose down 2>&1 | Out-Null
docker compose -f docker-compose.prod.yml up -d 2>&1 | Select-Object -Last 4
$healthy = $false
for ($i = 0; $i -lt 60; $i++) {
    Start-Sleep -Seconds 5
    $s = docker compose -f docker-compose.prod.yml ps --format "{{.Name}} {{.Status}}" 2>$null
    $bad = ($s | Select-String -Pattern "starting|unhealthy|restarting|exited").Count
    if ($s -and $bad -eq 0) { $healthy = $true; break }
}
Report "P0 生产栈启动健康" $healthy ($s -join " | ")
if (-not $healthy) { Write-Host "RESULT: FAILED"; exit 1 }

# ---------- P1 HTTPS 与跳转 ----------
$httpsCode = curl.exe -sk --noproxy "*" -o NUL -w "%{http_code}" "https://localhost/healthz"
Report "P1a HTTPS /healthz 200" ($httpsCode -eq "200") "code=$httpsCode"

$httpCode = curl.exe -s --noproxy "*" -o NUL -w "%{http_code}" "http://localhost/healthz"
Report "P1b HTTP → HTTPS 跳转" ($httpCode -eq "301" -or $httpCode -eq "308") "code=$httpCode"

$secHeader = curl.exe -sk --noproxy "*" -D - -o NUL "https://localhost/" | Select-String -Pattern "X-Content-Type-Options"
Report "P1c 安全响应头" ($secHeader -like "*nosniff*") ($secHeader | ForEach-Object { $_.Line.Trim() })

# ---------- P2 API 全链路（HTTPS） ----------
$bfLogin = Join-Path $env:TEMP "dxc-login.json"
'{"username":"admin","password":"Admin@123456"}' | Out-File -Encoding ascii $bfLogin
$loginResp = curl.exe -sk --noproxy "*" -X POST "https://localhost/api/v1/auth/login" -H "Content-Type: application/json" --data-binary "@$bfLogin"
$login = $loginResp | ConvertFrom-Json
$tok = $login.data.access_token
Report "P2a HTTPS 登录" ($null -ne $tok -and $tok.Length -gt 10) "token len=$($tok.Length)"

$bfEcs = Join-Path $env:TEMP "dxc-ecs.json"
'{"name":"prod-smoke-ecs","image":"busybox:latest","cpu":1,"memory_mb":256,"command":["sleep","3600"]}' | Out-File -Encoding ascii $bfEcs
$ecsResp = curl.exe -sk --noproxy "*" -X POST "https://localhost/api/v1/ecs" -H "Authorization: Bearer $tok" -H "Content-Type: application/json" --data-binary "@$bfEcs"
$ecs = ($ecsResp | ConvertFrom-Json).data
Report "P2b 生产环境创建 ECS" ($ecs.id -gt 0) "id=$($ecs.id)"

Start-Sleep -Seconds 4
$ecsList = (curl.exe -sk --noproxy "*" "https://localhost/api/v1/ecs?page=1&size=5" -H "Authorization: Bearer $tok" | ConvertFrom-Json).data
$found = ($ecsList.total -gt 0) -and (@($ecsList.items | Where-Object { $_.id -eq $ecs.id }).Count -gt 0)
Report "P2c ECS 列表可见" $found "total=$($ecsList.total)"

# ---------- P3 非 root 运行 ----------
$apiUser = (docker exec cloud-api id) -join " "
$webUser = (docker exec cloud-web id) -join " "
Report "P3a backend 非 root（uid 10001）" ($apiUser -like "*10001*" -and $apiUser -notlike "*uid=0*") $apiUser
Report "P3b frontend 非 root（node）" ($webUser -like "*uid=1000*" -and $webUser -notlike "*uid=0*") $webUser

# ---------- P4 备份 / 恢复 ----------
& .\tools\backup.ps1 | Out-Null
$latestBk = Get-ChildItem backups -Directory | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if ($null -eq $latestBk) {
    Report "P4a 全量备份产出" $false "备份目录未生成"
    Write-Host "RESULT: FAILED"; exit 1
}
$bkOk = (Test-Path (Join-Path $latestBk.FullName "mysql.sql")) -and (Test-Path (Join-Path $latestBk.FullName "redis-data")) -and (Test-Path (Join-Path $latestBk.FullName "registry-data.tgz"))
Report "P4a 全量备份产出" $bkOk "dir=$($latestBk.Name) files=$((Get-ChildItem $latestBk.FullName | ForEach-Object { $_.Name }) -join ',')"

& .\tools\restore.ps1 -BackupDir $latestBk.FullName | Out-Null
$restoreOk = ($LASTEXITCODE -eq 0)
Report "P4b 备份恢复执行" $restoreOk "mysql+redis+registry 已还原"

$afterRestore = (curl.exe -sk "https://localhost/api/v1/ecs?page=1&size=5" -H "Authorization: Bearer $tok" | ConvertFrom-Json).data
$stillThere = @($afterRestore.items | Where-Object { $_.id -eq $ecs.id }).Count -gt 0
Report "P4c 恢复后数据完整" $stillThere "total=$($afterRestore.total)"

# ---------- P5 优雅停机 ----------
docker compose -f docker-compose.prod.yml stop backend 2>&1 | Out-Null
Start-Sleep -Seconds 2
$grace = docker logs cloud-api --tail 30 2>&1 | Select-String -Pattern "shutting down|server exited"
$graceFound = [bool]($grace | Select-Object -First 1)
Report "P5 优雅停机日志" $graceFound (($grace | Select-Object -First 1 | ForEach-Object { $_.Line.Trim() }) -join " ")
docker compose -f docker-compose.prod.yml start backend 2>&1 | Out-Null
Start-Sleep -Seconds 8

# ---------- P6 恢复 dev 栈 ----------
Write-Host "== P6 恢复 dev 栈（重建 dev 镜像，避免与 prod 镜像同名冲突） =="
docker compose -f docker-compose.prod.yml down 2>&1 | Out-Null
docker compose build backend 2>&1 | Select-Object -Last 2
docker compose build frontend 2>&1 | Select-Object -Last 2
docker compose up -d 2>&1 | Select-Object -Last 3
$devHealthy = $false
for ($i = 0; $i -lt 60; $i++) {
    Start-Sleep -Seconds 5
    $s = docker compose ps --format "{{.Name}} {{.Status}}" 2>$null
    $bad = ($s | Select-String -Pattern "starting|unhealthy|restarting|exited").Count
    if ($s -and $bad -eq 0) { $devHealthy = $true; break }
}
Report "P6 dev 栈恢复健康" $devHealthy ($s -join " | ")

Write-Host ""
Write-Host "========== Phase 12 验收结果 =========="
$results | ForEach-Object { Write-Host $_ }
if ($global:FAILED) { Write-Host "RESULT: FAILED"; exit 1 } else { Write-Host "RESULT: ALL PASS"; exit 0 }
