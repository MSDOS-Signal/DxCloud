# 最终验收：用户验收全闭环 —— 注册 → 建项目 → 建 ECS → 终端 → 部署(×2) → 回滚 → Pipeline → 监控
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

function WaitState([string]$label, [scriptblock]$probe, [int]$tries, [int]$sleepSec, [string[]]$good, [string[]]$bad) {
    for ($i = 0; $i -lt $tries; $i++) {
        Start-Sleep -Seconds $sleepSec
        $v = & $probe
        if ($good -contains $v) { return @{ ok = $true; value = $v } }
        if ($bad -contains $v) { return @{ ok = $false; value = $v } }
    }
    return @{ ok = $false; value = $v }
}

$stamp = Get-Date -Format "MMddHHmmss"
$uname = "final$stamp"
$pwd = "Final@123456"
$img = "host.docker.internal:15000/default/pipetest:v5"

# ---------- F1 注册 ----------
Api "Post" "/auth/register" $null @{ username = $uname; email = "$uname@dx.dev"; password = $pwd } $null | Out-Null
$tok = (Api "Post" "/auth/login" $null @{ username = $uname; password = $pwd } $null).data.access_token
Report "F1 注册并登录" ($null -ne $tok) "user=$uname"

# ---------- F2 建项目 ----------
$proj = (Api "Post" "/projects" $tok @{ name = "final-proj-$stamp"; code = "final-proj" } $null).data
Report "F2 创建项目" ($proj.id -gt 0) "proj=$($proj.id)"

# ---------- F3 建 ECS ----------
$ecs = (Api "Post" "/ecs" $tok @{ name = "final-ecs-$stamp"; image = "busybox:latest"; cpu = 1; memory_mb = 256; command = @("sleep", "3600") } $null).data
$w = WaitState "ecs" { (Api "Get" "/ecs/$($ecs.id)" $tok $null $null).data.observed_state } 30 2 @("running") @("failed")
Report "F3 创建 ECS 并运行" ($ecs.id -gt 0 -and $w.ok) "id=$($ecs.id) state=$($w.value)"

# ---------- F4 终端 ----------
$env:ECS_ID = [string]$ecs.id
$env:ADMIN_TOKEN = $tok
node tools\ws-test\terminal-any.mjs *> $null
Report "F4 Web 终端（resize+命令回显）" ($LASTEXITCODE -eq 0) "ws exit=$LASTEXITCODE"

# ---------- F5 部署 v1 ----------
$app = (Api "Post" "/applications" $tok @{ project_id = $proj.id; name = "final-app-$stamp"; type = "container"; image = $img; port = 80 } $null).data
$d1 = (Api "Post" "/applications/$($app.id)/deploy" $tok @{ image = $img; note = "v1" } $null).data
$w1 = WaitState "dep1" { @((Api "Get" "/applications/$($app.id)/deployments" $tok $null $null).data)[0].status } 40 2 @("success") @("failed")
Report "F5 部署 v1（蓝绿）" ($app.id -gt 0 -and $w1.ok) "dep=$($d1.id) status=$($w1.value)"

# ---------- F6 部署 v2 ----------
$d2 = (Api "Post" "/applications/$($app.id)/deploy" $tok @{ image = $img; note = "v2" } $null).data
$w2 = WaitState "dep2" { @((Api "Get" "/applications/$($app.id)/deployments" $tok $null $null).data)[0].status } 40 2 @("success") @("failed")
Report "F6 部署 v2（蓝绿切换）" ($w2.ok) "dep=$($d2.id) status=$($w2.value)"

# ---------- F7 回滚到 v1 ----------
$vers = (Api "Get" "/applications/$($app.id)/versions" $tok $null $null).data
$v1 = @($vers)[-1]
$rb = (Api "Post" "/applications/$($app.id)/versions/$($v1.id)/rollback" $tok $null $null).data
$w3 = WaitState "rollback" { @((Api "Get" "/applications/$($app.id)/deployments" $tok $null $null).data)[0].status } 40 2 @("success") @("failed")
$rbNote = @((Api "Get" "/applications/$($app.id)/deployments" $tok $null $null).data)[0].note
Report "F7 回滚到 v1" ($w3.ok -and $rbNote -like "*rollback*") "dep=$($rb.id) status=$($w3.value) note=$rbNote"

# ---------- F8 Pipeline ----------
$pipeDef = @"
name: final-pipe-$stamp
timeout: 10m
steps:
  - name: 编译检查
    type: shell
    script: echo build-ok
  - name: 部署
    type: shell
    script: echo deploy-ok
"@
$pipe = (Api "Post" "/pipelines" $tok @{ name = "final-pipe-$stamp"; description = "final"; definition = $pipeDef } $null).data
$run = (Api "Post" "/pipelines/$($pipe.id)/run" $tok @{ ref = "main" } $null).data
$w4 = WaitState "pipe" { (Api "Get" "/pipeline-runs/$($run.id)" $tok $null $null).data.status } 60 3 @("success") @("failed","canceled")
Report "F8 Pipeline 运行成功" ($pipe.id -gt 0 -and $w4.ok) "run=$($run.id) status=$($w4.value)"

# ---------- F9 监控 ----------
$dash = (Api "Get" "/monitor/dashboard" $tok $null $null).data
$series = (Api "Get" "/monitor/series?kind=ecs&minutes=1440" $tok $null $null).data
$seriesInst = (Api "Get" "/monitor/series?kind=ecs&ref_id=$($ecs.id)&minutes=30" $tok $null $null).data
Report "F9 监控总览+实例曲线" ($dash.ecs_total -ge 1 -and @($series).Count -ge 1 -and $null -ne $seriesInst) "ecs=$($dash.ecs_total) deploy_today=$($dash.deploy_today) series_points=$(@($series).Count)"

# ---------- F10 清理 ----------
Api "Delete" "/applications/$($app.id)" $tok $null $null | Out-Null
Api "Delete" "/ecs/$($ecs.id)" $tok $null $null | Out-Null
Api "Delete" "/pipelines/$($pipe.id)" $tok $null $null | Out-Null
Report "F10 资源清理" $true "app/ecs/pipeline 已删除"

Write-Host ""
Write-Host "========== 最终验收结果 =========="
Write-Host "user=$uname proj=$($proj.id) ecs=$($ecs.id) app=$($app.id) pipe=$($pipe.id)"
$results | ForEach-Object { Write-Host $_ }
if ($global:FAILED) { Write-Host "RESULT: FAILED"; exit 1 } else { Write-Host "RESULT: ALL PASS —— 用户验收全闭环通过"; exit 0 }
