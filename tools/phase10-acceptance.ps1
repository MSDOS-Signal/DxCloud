# Phase 10 验收：多租户 / 配额 / 虚拟计费
# 前置：docker compose 全栈已启动（cloud-api 健康）
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

function ExpectHttp([string]$method, [string]$path, $token, $body, $orgId, [int]$expectedHttp) {
    try {
        Api $method $path $token $body $orgId | Out-Null
        return "unexpected 200"
    } catch {
        $code = [int]$_.Exception.Response.StatusCode
        $msg = ""
        if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
            try { $msg = ($_.ErrorDetails.Message | ConvertFrom-Json).message } catch { $msg = $_.ErrorDetails.Message }
        }
        if ($code -eq $expectedHttp) { return "HTTP $code : $msg" }
        return "HTTP $code (expect $expectedHttp) : $msg"
    }
}

# ---------- T1 admin 登录 ----------
$login = Api "Post" "/auth/login" $null @{ username = "admin"; password = "Admin@123456" } $null
$adminTok = $login.data.access_token
Report "T1 admin 登录" ($null -ne $adminTok) "token len=$($adminTok.Length)"

# ---------- T2 创建组织 A / B ----------
$stamp = Get-Date -Format "MMddHHmmss"
$orgA = (Api "Post" "/organizations" $adminTok @{ name = "租户A-$stamp"; code = "tenant-a-$stamp"; plan = "pro" } $null).data
$orgB = (Api "Post" "/organizations" $adminTok @{ name = "租户B-$stamp"; code = "tenant-b-$stamp"; plan = "free" } $null).data
$idA = [uint64]$orgA.id; $idB = [uint64]$orgB.id
Report "T2 创建组织 A/B" ($idA -gt 0 -and $idB -gt 0) "orgA=$idA credit=$($orgA.credit), orgB=$idB credit=$($orgB.credit)"

# ---------- T3 注册并登录 bob ----------
$bobName = "bob$stamp"
$bobReg = Api "Post" "/auth/register" $null @{ username = $bobName; email = "$bobName@dx.dev"; password = "Bob@123456" } $null
$bobTok = (Api "Post" "/auth/login" $null @{ username = $bobName; password = "Bob@123456" } $null).data.access_token
$bobUid = ((Api "Get" ('/users?keyword=' + $bobName) $adminTok $null $null).data.items | Where-Object { $_.username -eq $bobName } | Select-Object -First 1).id
Report "T3 bob 注册登录" ($null -ne $bobTok -and $bobUid -gt 0) "user=$bobName uid=$bobUid"

# ---------- T4 bob 加入组织 A ----------
Api "Post" "/organizations/$idA/members" $adminTok @{ username = $bobName; org_role = "member" } $null | Out-Null
$members = (Api "Get" "/organizations/$idA/members" $adminTok $null $null).data
$bobInA = @($members | Where-Object { $_.user_id -eq $bobUid }).Count -gt 0
Report "T4 bob 加入组织 A" $bobInA "members=$(@($members).Count) bob_uid=$bobUid"

# ---------- T5/T6 项目 + 跨组织隔离 ----------
$projA = (Api "Post" "/projects" $adminTok @{ name = "pa-web-$stamp"; code = "pa-web" } $idA).data
$projB = (Api "Post" "/projects" $adminTok @{ name = "pb-web-$stamp"; code = "pb-web" } $idB).data
Report "T5 项目创建" ($projA.id -gt 0 -and $projB.id -gt 0) "projA=$($projA.id) org=$($projA.org_id), projB=$($projB.id) org=$($projB.org_id)"

$deny = ExpectHttp "Get" "/projects" $bobTok $null $idB 403
Report "T6 bob 访问组织 B 项目 → 403" ($deny -like "HTTP 403*") $deny

$bobProjs = (Api "Get" "/projects" $bobTok $null $idA).data
$names = @($bobProjs | ForEach-Object { $_.name })
$onlyA = ($names -contains $projA.name) -and (-not ($names -contains $projB.name))
Report "T7 bob 组织 A 列表仅见 A 项目" $onlyA "names=$($names -join ',')"

# ---------- T8 配额：ecs_count=1 → 第 2 台被拒 ----------
Api "Put" "/quotas?org_id=$idA" $adminTok @{ resource_type = "ecs_count"; limit_value = 1 } $null | Out-Null
$qList = (Api "Get" "/quotas?org_id=$idA" $adminTok $null $null).data
$qCount = @($qList | Where-Object { $_.resource_type -eq "ecs_count" }).limit_value
Report "T8a 组织配额设置 ecs_count=1" ($qCount -eq 1) "quota=$qCount"

$ecsBody1 = @{ name = "tenant-a-ecs1"; image = "busybox:latest"; cpu = 1; memory_mb = 512; disk_gb = 10; command = @("sleep", "3600") }
$ecs1 = (Api "Post" "/ecs" $adminTok $ecsBody1 $idA).data
Report "T8b ECS-1 创建成功" ($ecs1.id -gt 0) "id=$($ecs1.id) org=$($ecs1.org_id)"

Start-Sleep -Seconds 3
$deny2 = ExpectHttp "Post" "/ecs" $adminTok $ecsBody1 $idA 400
Report "T8c ECS-2 超额被拒(400)" ($deny2 -like "HTTP 400*") $deny2

# ---------- T9 组织盖章 + 列表隔离 ----------
$ecsPage = (Api "Get" '/ecs?page=1&size=50' $adminTok $null $idA).data
$stamped = @($ecsPage.items | Where-Object { $_.id -eq $ecs1.id -and [string]$_.org_id -eq [string]$idA }).Count -gt 0
Report "T9 ECS 组织盖章/列表隔离" $stamped "total=$($ecsPage.total)"

# ---------- T10 计费：tick + summary ----------
Api "Post" "/billing/tick" $adminTok $null $null | Out-Null
$sum1 = (Api "Get" "/billing?org_id=$idA" $adminTok $null $null).data
$cpuH = $sum1.usage_month.cpu_hour
$creditAfter = [double]$sum1.credit
Report "T10a 计费 tick 产生用量" ($cpuH -gt 0) "cpu_hour=$cpuH mem_gb_hour=$($sum1.usage_month.mem_gb_hour) disk_gb_hour=$($sum1.usage_month.disk_gb_hour)"
Report "T10b 余额被扣减" ($creditAfter -lt 1000) "credit=$creditAfter (initial 1000)"

$recs = (Api "Get" ('/billing/records?org_id=' + $idA + '&limit=10') $adminTok $null $null).data
Report "T10c 账单记录可查" (@($recs).Count -gt 0) "records=$(@($recs).Count)"

# ---------- T11 充值 ----------
$sum2 = (Api "Post" "/billing/recharge" $adminTok @{ org_id = $idA; amount = 500 } $null).code
$sum3 = (Api "Get" "/billing?org_id=$idA" $adminTok $null $null).data
Report "T11 虚拟充值 +500" ([double]$sum3.credit -gt $creditAfter) "credit=$($sum3.credit) (before $creditAfter)"

# ---------- T12 余额门禁 ----------
Api "Put" "/quotas?org_id=$idA" $adminTok @{ resource_type = "ecs_count"; limit_value = 10 } $null | Out-Null
$null = docker exec dx-mysql mysql -udxcloud -pdxcloud_dev_pass -e "UPDATE dxcloud.organizations SET credit=-1001 WHERE id=$idA;"
$gateBody = @{ name = "tenant-a-ecs-gate"; image = "busybox:latest"; cpu = 1; memory_mb = 512; command = @("sleep", "3600") }
$deny3 = ExpectHttp "Post" "/ecs" $adminTok $gateBody $idA 400
Report "T12a 余额不足拒绝创建" ($deny3 -like "HTTP 400*") $deny3
Api "Post" "/billing/recharge" $adminTok @{ org_id = $idA; amount = 5000 } $null | Out-Null
$ecs3 = (Api "Post" "/ecs" $adminTok @{ name = "tenant-a-ecs3"; image = "busybox:latest"; cpu = 1; memory_mb = 512; command = @("sleep", "3600") } $idA).data
Report "T12b 充值后创建恢复" ($ecs3.id -gt 0) "ecs3 id=$($ecs3.id)"

# ---------- T13 应用列表隔离 ----------
$appA = (Api "Post" "/applications" $adminTok @{ project_id = $projA.id; name = "appa-$stamp"; type = "container"; image = "busybox:latest"; port = 80 } $idA).data
$bobApps = (Api "Get" "/applications" $bobTok $null $idA).data
$appInA = @($bobApps | Where-Object { $_.id -eq $appA.id }).Count -gt 0
$bobAppsB = (Api "Get" "/applications" $bobTok $null $idB).data
$appNotInB = @($bobAppsB | Where-Object { $_.id -eq $appA.id }).Count -eq 0
Report "T13 应用列表组织隔离" ($appInA -and $appNotInB) "appA=$($appA.id) org=$($appA.org_id)"

# ---------- T14 跨组织资源读取被拒 ----------
$deny4 = ExpectHttp "Get" "/ecs/$($ecs1.id)" $bobTok $null $idB 403
Report "T14 bob 跨组织读 ECS → 403" ($deny4 -like "HTTP 403*") $deny4

# ---------- 清理临时资源 ----------
$null = docker exec dx-mysql mysql -udxcloud -pdxcloud_dev_pass -e "UPDATE dxcloud.organizations SET credit=1000 WHERE id=$idA;"
Write-Host ""
Write-Host "========== Phase 10 验收结果 =========="
Write-Host "orgA=$idA orgB=$idB bob=$bobName projA=$($projA.id) projB=$($projB.id) ecs1=$($ecs1.id) appA=$($appA.id)"
$results | ForEach-Object { Write-Host $_ }
if ($global:FAILED) { Write-Host "RESULT: FAILED"; exit 1 } else { Write-Host "RESULT: ALL PASS"; exit 0 }
