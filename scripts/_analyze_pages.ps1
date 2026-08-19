$ErrorActionPreference = 'Stop'
$root = "d:\deepseek_harness\own_project"
$order = @(
  @{ dir = "backend\cmd\server"; ext = ".go" },
  @{ dir = "backend\internal\config"; ext = ".go" },
  @{ dir = "backend\internal\api"; ext = ".go" },
  @{ dir = "backend\internal\database"; ext = ".go" },
  @{ dir = "backend\internal\model"; ext = ".go" },
  @{ dir = "backend\internal\dto"; ext = ".go" },
  @{ dir = "backend\internal\repository"; ext = ".go" },
  @{ dir = "backend\internal\service"; ext = ".go" },
  @{ dir = "backend\internal\handler"; ext = ".go" },
  @{ dir = "backend\internal\iam"; ext = ".go" },
  @{ dir = "backend\internal\middleware"; ext = ".go" },
  @{ dir = "backend\internal\docker"; ext = ".go" },
  @{ dir = "backend\internal\pipeline"; ext = ".go" },
  @{ dir = "backend\internal\runner"; ext = ".go" },
  @{ dir = "backend\internal\scheduler"; ext = ".go" },
  @{ dir = "backend\internal\websocket"; ext = ".go" },
  @{ dir = "backend\pkg"; ext = ".go" },
  @{ dir = "backend\migrations"; ext = ".sql" },
  @{ dir = "frontend"; ext = "app.vue" },
  @{ dir = "frontend"; ext = "nuxt.config.ts" },
  @{ dir = "frontend\assets\css"; ext = ".css" },
  @{ dir = "frontend\components"; ext = ".vue" },
  @{ dir = "frontend\composables"; ext = ".ts" },
  @{ dir = "frontend\layouts"; ext = ".vue" },
  @{ dir = "frontend\middleware"; ext = ".ts" },
  @{ dir = "frontend\pages"; ext = ".vue" },
  @{ dir = "frontend\stores"; ext = ".ts" },
  @{ dir = "frontend\services"; ext = ".ts" },
  @{ dir = "frontend\types"; ext = ".ts" },
  @{ dir = "frontend\plugins"; ext = ".ts" }
)
$allLines = New-Object System.Collections.Generic.List[string]
foreach ($o in $order) {
  $full = Join-Path $root $o.dir
  if (-not (Test-Path $full)) { continue }
  $files = Get-ChildItem $full -Recurse -File | Where-Object {
    if ($o.ext -eq 'app.vue') { $_.Name -eq 'app.vue' -and $_.DirectoryName -eq $full }
    elseif ($o.ext -eq 'nuxt.config.ts') { $_.Name -eq 'nuxt.config.ts' -and $_.DirectoryName -eq $full }
    elseif ($o.ext.StartsWith('.')) { $_.Extension -eq $o.ext }
    else { $false }
  } | Where-Object { $_.FullName -notmatch 'node_modules|\.nuxt|\.output' } | Sort-Object FullName
  foreach ($f in $files) {
    $allLines.Add("SEP")
    foreach ($l in [IO.File]::ReadAllLines($f.FullName)) { $allLines.Add($l) }
    $allLines.Add("")
  }
}
$CH = 105
$pageCount = [math]::Ceiling($allLines.Count / 50.0)
$risk = New-Object System.Collections.Generic.List[string]
for ($pg = 0; $pg -lt $pageCount; $pg++) {
  $sum = 0
  for ($i = 0; $i -lt 50; $i++) {
    $idx = $pg * 50 + $i
    if ($idx -ge $allLines.Count) { break }
    $sum += [math]::Max(1, [math]::Ceiling($allLines[$idx].Length / $CH))
  }
  if ($sum -gt 62) { $risk.Add(("第 {0} 页: 估算渲染 {1} 行" -f ($pg + 1), $sum)) }
}
if ($risk.Count) { $risk | Select-Object -First 10; "共 $($risk.Count) 个风险页（>62渲染行，可能溢出A4）" } else { "全部页安全：无页超过 62 渲染行" }
"总行 $($allLines.Count), 总页 $pageCount"
$over300 = @($allLines | Where-Object { $_.Length -gt 300 })
$over120 = @($allLines | Where-Object { $_.Length -gt 120 })
"超300字符行: $($over300.Count) 个; 超120字符行: $($over120.Count) 个"
foreach ($l in ($over300 | Select-Object -First 3)) { "  [预览] " + $l.Substring(0, [math]::Min(90, $l.Length)) }
