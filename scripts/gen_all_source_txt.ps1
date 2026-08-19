$ErrorActionPreference = 'Stop'
$root = "d:\deepseek_harness\own_project"

# 按模块顺序定义源文件（覆盖全部源代码）
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
  @{ dir = "frontend"; ext = "tailwind.config.ts" },
  @{ dir = "frontend"; ext = "playwright.config.ts" },
  @{ dir = "frontend\assets\css"; ext = ".css" },
  @{ dir = "frontend\components"; ext = ".vue" },
  @{ dir = "frontend\composables"; ext = ".ts" },
  @{ dir = "frontend\layouts"; ext = ".vue" },
  @{ dir = "frontend\middleware"; ext = ".ts" },
  @{ dir = "frontend\modules"; ext = ".ts" },
  @{ dir = "frontend\utils"; ext = ".ts" },
  @{ dir = "frontend\pages"; ext = ".vue" },
  @{ dir = "frontend\stores"; ext = ".ts" },
  @{ dir = "frontend\services"; ext = ".ts" },
  @{ dir = "frontend\types"; ext = ".ts" },
  @{ dir = "frontend\plugins"; ext = ".ts" },
  @{ dir = "frontend\tests"; ext = ".ts" }
)

$allLines = New-Object System.Collections.Generic.List[string]
$fileCount = 0

foreach ($o in $order) {
  $full = Join-Path $root $o.dir
  if (-not (Test-Path $full)) { continue }
  $files = Get-ChildItem $full -Recurse -File | Where-Object {
    if ($o.ext -notmatch '\.') { $_.Name -eq $o.ext -and $_.DirectoryName -eq $full }
    elseif ($o.ext.StartsWith('.')) { $_.Extension -eq $o.ext }
    else { $false }
  } | Where-Object { $_.FullName -notmatch 'node_modules|\.nuxt|\.output|test-results' } | Sort-Object FullName
  foreach ($f in $files) {
    $rel = $f.FullName -replace [regex]::Escape("$root\"), ''
    $allLines.Add("/" + "* ============ 源文件: $rel ============ " + "*/")
    $fileCount++
    $lines = [IO.File]::ReadAllLines($f.FullName)
    foreach ($l in $lines) { $allLines.Add($l) }
    $allLines.Add("")
  }
}

$total = $allLines.Count
$totalPages = [math]::Ceiling($total / 50.0)
$blank = 0
$sep = 0
$code = 0
foreach ($l in $allLines) {
  if ($l -eq '') { $blank++ }
  elseif ($l -match '^/\* =+ 源文件: .+ =+ \*/$') { $sep++ }
  else { $code++ }
}
Write-Host "源文件: $fileCount 个"
Write-Host "txt 总行数: $total （代码 $code 行 + 文件分隔注释 $sep 行 + 空行 $blank 行）"
Write-Host "按软著 50 行/页折算: $totalPages 页"

$txtPath = Join-Path $root "多晓云DxCloud-全部源代码.txt"
[IO.File]::WriteAllLines($txtPath, $allLines, (New-Object System.Text.UTF8Encoding $true))
Write-Host "已生成: $txtPath"
Write-Host ("文件大小: {0:N0} KB" -f ((Get-Item $txtPath).Length / 1KB))
