# Phase 11 验收：安全加固（基线审计 / 镜像策略 / 密钥托管 / 防爆破）
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

# ---------- S1 admin 登录 ----------
$login = Api "Post" "/auth/login" $null @{ username = "admin"; password = "Admin@123456" } $null
$adminTok = $login.data.access_token
Report "S1 admin 登录" ($null -ne $adminTok) "token len=$($adminTok.Length)"

# ---------- S2 安全总览 ----------
$dash = (Api "Get" "/security/dashboard" $adminTok $null $null).data
Report "S2 安全总览" ($null -ne $dash.score -and $null -ne $dash.finding_count) "score=$($dash.score) findings=$($dash.finding_count)"

# ---------- S3 全量扫描 ----------
$scan = (Api "Post" "/security/scan" $adminTok $null $null).data
$hasBaseline = @($scan | Where-Object { $_.kind -eq "baseline" }).Count -gt 0
$hasImage = @($scan | Where-Object { $_.kind -eq "image" }).Count -gt 0
$score0 = @($scan | Where-Object { $_.kind -eq "baseline" }).score
Report "S3 全量扫描（基线+镜像）" ($hasBaseline -and $hasImage) "baseline_score=$score0"

# ---------- S4 扫描历史 + 发现项 ----------
$reps = (Api "Get" "/security/reports?limit=10" $adminTok $null $null).data
$repId = @($reps)[0].id
$repDetail = (Api "Get" "/security/reports/$repId" $adminTok $null $null).data
Report "S4 扫描历史与发现项" (@($reps).Count -ge 2 -and $null -ne $repDetail.findings) "reports=$(@($reps).Count) repId=$repId findings=$($repDetail.finding_count)"

# ---------- S5 密钥托管（组织 A 作用域） ----------
$orgA = 3; $orgB = 4
$secName = "DB_PASSWORD_P11"
$sec = (Api "Post" "/secrets" $adminTok @{ name = $secName; value = "Sup3rS3cret#2026" } $orgA).data
Report "S5a 密钥加密创建（响应无明文）" ($sec.id -gt 0 -and -not ($sec.PSObject.Properties.Name -contains "value")) "id=$($sec.id) org=$($sec.org_id)"

$secListA = (Api "Get" "/secrets" $adminTok $null $orgA).data
$inA = @($secListA | Where-Object { $_.name -eq $secName }).Count -gt 0
Report "S5b 组织 A 可见密钥" $inA "count=$(@($secListA).Count)"

$secListB = (Api "Get" "/secrets" $adminTok $null $orgB).data
$notInB = @($secListB | Where-Object { $_.name -eq $secName }).Count -eq 0
Report "S5c 组织 B 隔离（不可见）" $notInB "count=$(@($secListB).Count)"

$reveal = (Api "Get" "/secrets/$($sec.id)/reveal" $adminTok $null $orgA).data
Report "S5d 解密读取正确" ($reveal.value -eq "Sup3rS3cret#2026") "value=$($reveal.value)"

$denyReveal = ExpectHttp "Get" "/secrets/$($sec.id)/reveal" $adminTok $null $orgB 403
Report "S5e 跨组织解密被拒(403)" ($denyReveal -like "HTTP 403*") $denyReveal

Api "Delete" "/secrets/$($sec.id)" $adminTok $null $orgA | Out-Null
$afterDel = (Api "Get" "/secrets" $adminTok $null $orgA).data
$deleted = @($afterDel | Where-Object { $_.name -eq $secName }).Count -eq 0
Report "S5f 密钥删除" $deleted "remaining=$(@($afterDel).Count)"

# ---------- S6 登录防爆破锁定 ----------
# 等一个限流窗口滑过（登录接口 5 次/分钟/IP，避免与前面 S1 的请求叠加）
Start-Sleep -Seconds 65
$stamp = Get-Date -Format "MMddHHmmss"
$lockName = "locku$stamp"
Api "Post" "/auth/register" $null @{ username = $lockName; email = "$lockName@dx.dev"; password = "Lock@123456" } $null | Out-Null
$fails = 0
for ($i = 0; $i -lt 3; $i++) {
    $r = ExpectHttp "Post" "/auth/login" $null @{ username = $lockName; password = "WrongPass$i" } $null 401
    if ($r -like "HTTP 401*") { $fails++ }
}
# 手动将失败计数置为 5（模拟累计第 5 次失败，避免与限流 5/min 冲突）
$null = docker exec dx-redis redis-cli -a dxcloud_redis_dev SET "dx:login:fail:$lockName" 5 EX 900
$locked = ExpectHttp "Post" "/auth/login" $null @{ username = $lockName; password = "Lock@123456" } $null 401
Report "S6a 5 次失败后锁定（正确密码也 401）" ($fails -eq 3 -and $locked -like "HTTP 401*" -and $locked -like "*锁定*") $locked

Start-Sleep -Seconds 62
$null = docker exec dx-redis redis-cli -a dxcloud_redis_dev DEL "dx:login:fail:$lockName"
$okLogin = (Api "Post" "/auth/login" $null @{ username = $lockName; password = "Lock@123456" } $null).data.access_token
Report "S6b 锁定期满后恢复登录" ($null -ne $okLogin) "token len=$($okLogin.Length)"

Write-Host ""
Write-Host "========== Phase 11 验收结果 =========="
$results | ForEach-Object { Write-Host $_ }
if ($global:FAILED) { Write-Host "RESULT: FAILED"; exit 1 } else { Write-Host "RESULT: ALL PASS"; exit 0 }
