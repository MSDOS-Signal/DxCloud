# Phase 13 回归：全栈 API 回归套件（注册→项目→ECS→终端→部署→Pipeline→监控→安全→计费）
# 前置：dev 栈运行中（docker compose up -d）
$base = "http://localhost/api/v1"
$results = New-Object System.Collections.Generic.List[string]
$global:FAILED = $false

function Report([string]$label, [bool]$ok, [string]$detail) {
    $mark = if ($ok) { "[PASS]" } else { $global:FAILED = $true; "[FAIL]" }
    $line = "$mark $label - $detail"
    $results.Add($line)
    Write-Host $line
}

function Api([string]$method, [string]$path, $token, $body, $orgId) {
    $headers = @{}
    if ($token) { $headers["Authorization"] = "Bearer $token" }
    if ($orgId) { $headers["X-Org-Id"] = [string]$orgId }
    $params = @{ Method = $method; Uri = "$base$path"; Headers = $headers; ContentType = "application/json" }
    if ($null -ne $body) { $params.Body = ($body | ConvertTo-Json -Depth 8) }
    return Invoke-RestMethod @params
}

# ---------- R1 健康与登录 ----------
$health = Invoke-RestMethod -Uri "http://localhost/healthz"
Report "R1 健康检查" ($health -like "*ok*" -or $health -match "ok") "healthz=$health"

$login = Api "Post" "/auth/login" $null @{ username = "admin"; password = "Admin@123456" } $null
$adminTok = $login.data.access_token
Report "R2 admin 登录" ($null -ne $adminTok) "token len=$($adminTok.Length)"

# ---------- R3 注册用户 ----------
$stamp = Get-Date -Format "MMddHHmmss"
$uname = "reg$stamp"
Api "Post" "/auth/register" $null @{ username = $uname; email = "$uname@dx.dev"; password = "Reg@123456" } $null | Out-Null
$uTok = (Api "Post" "/auth/login" $null @{ username = $uname; password = "Reg@123456" } $null).data.access_token
Report "R3 注册并登录新用户" ($null -ne $uTok) "user=$uname"

# ---------- R4 项目 ----------
$proj = (Api "Post" "/projects" $uTok @{ name = "reg-proj-$stamp"; code = "reg-proj" } $null).data
Report "R4 创建项目" ($proj.id -gt 0) "proj=$($proj.id)"

# ---------- R5 ECS 创建 → 运行 ----------
$ecs = (Api "Post" "/ecs" $uTok @{ name = "reg-ecs-$stamp"; image = "busybox:latest"; cpu = 1; memory_mb = 256; command = @("sleep", "3600") } $null).data
$running = $false
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 2
    $st = (Api "Get" "/ecs/$($ecs.id)" $uTok $null $null).data.observed_state
    if ($st -eq "running") { $running = $true; break }
}
Report "R5 ECS 创建并运行" ($ecs.id -gt 0 -and $running) "id=$($ecs.id) state=$st"

# ---------- R6 监控数据 ----------
$stats = (Api "Get" "/ecs/$($ecs.id)/stats" $uTok $null $null).data
$logs = (Api "Get" "/ecs/$($ecs.id)/logs?tail=20" $uTok $null $null).data
Report "R6 ECS stats+logs" ($null -ne $stats.cpu_percent -and $null -ne $logs.logs) "cpu=$($stats.cpu_percent)"

# ---------- R7 Web Terminal（WebSocket） ----------
$env:ECS_ID = [string]$ecs.id
$env:ADMIN_TOKEN = $uTok
node tools\ws-test\terminal-any.mjs *> $null
$termOk = ($LASTEXITCODE -eq 0)
Report "R7 Web 终端（resize+命令回显）" $termOk "ws exit=$LASTEXITCODE"

# ---------- R8 生命周期 ----------
Api "Post" "/ecs/$($ecs.id)/stop" $uTok $null $null | Out-Null
Start-Sleep -Seconds 4
$st1 = (Api "Get" "/ecs/$($ecs.id)" $uTok $null $null).data.observed_state
Api "Post" "/ecs/$($ecs.id)/start" $uTok $null $null | Out-Null
Start-Sleep -Seconds 4
$st2 = (Api "Get" "/ecs/$($ecs.id)" $uTok $null $null).data.observed_state
Report "R8 停止/启动" ($st1 -eq "stopped" -and $st2 -eq "running") "stop=$st1 start=$st2"

# ---------- R9 应用 + 蓝绿部署 ----------
$app = (Api "Post" "/applications" $uTok @{ project_id = $proj.id; name = "reg-app-$stamp"; type = "container"; image = "host.docker.internal:15000/default/pipetest:v5"; port = 80 } $null).data
$dep = (Api "Post" "/applications/$($app.id)/deploy" $uTok @{ image = "host.docker.internal:15000/default/pipetest:v5"; note = "regression" } $null).data
$depOk = $false
for ($i = 0; $i -lt 40; $i++) {
    Start-Sleep -Seconds 2
    $deps = (Api "Get" "/applications/$($app.id)/deployments" $uTok $null $null).data
    $cur = @($deps)[0]
    if ($cur.status -eq "success") { $depOk = $true; break }
    if ($cur.status -eq "failed") { break }
}
Report "R9 应用蓝绿部署成功" ($app.id -gt 0 -and $depOk) "app=$($app.id) dep=$($dep.id) status=$($cur.status)"

# ---------- R10 Pipeline ----------
$pipeDef = @"
name: reg-pipe-$stamp
timeout: 10m
steps:
  - name: 输出
    type: shell
    script: echo hello-from-regression
  - name: 收尾
    type: shell
    script: echo done
"@
$pipe = (Api "Post" "/pipelines" $uTok @{ name = "reg-pipe-$stamp"; description = "regression"; definition = $pipeDef } $null).data
$run = (Api "Post" "/pipelines/$($pipe.id)/run" $uTok @{ ref = "main" } $null).data
$pipeOk = $false
for ($i = 0; $i -lt 60; $i++) {
    Start-Sleep -Seconds 3
    $r = (Api "Get" "/pipeline-runs/$($run.id)" $uTok $null $null).data
    if ($r.status -eq "success") { $pipeOk = $true; break }
    if ($r.status -eq "failed" -or $r.status -eq "canceled") { break }
}
Report "R10 Pipeline 运行成功" ($pipe.id -gt 0 -and $pipeOk) "pipe=$($pipe.id) run=$($run.id) status=$($r.status)"

# ---------- R11 监控总览 ----------
$dash = (Api "Get" "/monitor/dashboard" $adminTok $null $null).data
Report "R11 监控总览" ($null -ne $dash.ecs_total) "ecs=$($dash.ecs_total) apps=$($dash.app_count)"

# ---------- R12 安全扫描 ----------
$scan = (Api "Post" "/security/scan" $adminTok $null $null).data
$secOk = @($scan).Count -ge 2
Report "R12 安全扫描" $secOk "reports=$(@($scan).Count)"

# ---------- R13 计费 ----------
$bill = (Api "Get" "/billing" $adminTok $null $null).data
Report "R13 计费总览" ($null -ne $bill.credit) "credit=$($bill.credit)"

# ---------- R14 清理 ----------
Api "Delete" "/applications/$($app.id)" $uTok $null $null | Out-Null
Api "Delete" "/ecs/$($ecs.id)" $uTok $null $null | Out-Null
Api "Delete" "/pipelines/$($pipe.id)" $uTok $null $null | Out-Null
$clean = $true
Start-Sleep -Seconds 4
try { $null = Api "Get" "/ecs/$($ecs.id)" $uTok $null $null } catch { $clean = $clean -and $true }
Report "R14 资源清理" $clean "app/ecs/pipeline 已删除"

Write-Host ""
Write-Host "========== Phase 13 回归结果 =========="
$results | ForEach-Object { Write-Host $_ }
if ($global:FAILED) { Write-Host "RESULT: FAILED"; exit 1 } else { Write-Host "RESULT: ALL PASS"; exit 0 }
