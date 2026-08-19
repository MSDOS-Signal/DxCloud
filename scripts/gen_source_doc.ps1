$ErrorActionPreference = 'Stop'
$root = "d:\deepseek_harness\own_project"

# 按模块顺序定义源文件
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
$fileCount = 0

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
$maxLen = 0
foreach ($l in $allLines) { if ($l.Length -gt $maxLen) { $maxLen = $l.Length } }
Write-Host "源文件: $fileCount 个, 总行数: $total, 总页数: $totalPages (每页50行), 最长行: $maxLen 字符"

$LINES_PER_PAGE = 50
$FRONT_PAGES = 30
$BACK_PAGES = 30
$frontLines = $allLines.GetRange(0, $FRONT_PAGES * $LINES_PER_PAGE)
$backStart = [math]::Max($total - $BACK_PAGES * $LINES_PER_PAGE, 0)
$backLines = $allLines.GetRange($backStart, $total - $backStart)

$NAME = '多晓云DxCloud一体化云平台软件'
$LANGS = 'Go、TypeScript、Vue、SQL、CSS、HTML'

function Get-PageLines([System.Collections.Generic.List[string]]$src, [int]$pageIdx) {
  $pageLines = New-Object System.Collections.Generic.List[string]
  for ($i = 0; $i -lt $LINES_PER_PAGE; $i++) {
    $idx = $pageIdx * $LINES_PER_PAGE + $i
    if ($idx -lt $src.Count) { $pageLines.Add($src[$idx]) } else { $pageLines.Add('') }
  }
  return ,$pageLines
}

function Esc([string]$s) {
  return $s.Replace('&', '&amp;').Replace('<', '&lt;').Replace('>', '&gt;')
}

# ================= Markdown（纯 Markdown，无任何 HTML） =================
$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("# $NAME V1.0")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("## 程序鉴别材料（源程序）")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("| 项目 | 内容 |")
[void]$sb.AppendLine("|------|------|")
[void]$sb.AppendLine("| 软件全称 | $NAME |")
[void]$sb.AppendLine("| 版本号 | V1.0 |")
[void]$sb.AppendLine("| 开发语言 | $LANGS |")
[void]$sb.AppendLine("| 源程序总量 | 约 $total 行（$fileCount 个源文件，合计 $totalPages 页） |")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("**材料构成**：按版权登记要求，提交源程序**前连续 30 页**与**后连续 30 页**，每页 50 行。")
[void]$sb.AppendLine("")
[void]$sb.AppendLine('> 打印/PDF 说明：如需严格按「每页 50 行」分页输出 PDF，请使用同目录配套的 **软著-程序鉴别材料-源代码.html**（浏览器打开后 Ctrl+P 另存为 PDF，自动分页）。本 MD 文件为纯 Markdown 底稿，供阅读与编辑，未内嵌任何 HTML 分页标记。')
[void]$sb.AppendLine("")
[void]$sb.AppendLine("---")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("# 第一部分　前连续 30 页（第 1 页 — 第 30 页）")
[void]$sb.AppendLine("")

for ($p = 0; $p -lt $FRONT_PAGES; $p++) {
  $pageLines = Get-PageLines $frontLines $p
  [void]$sb.AppendLine("**$NAME V1.0　·　源程序　·　第 $($p + 1) 页**")
  [void]$sb.AppendLine("")
  [void]$sb.AppendLine('````')
  foreach ($l in $pageLines) { [void]$sb.AppendLine($l) }
  [void]$sb.AppendLine('````')
  [void]$sb.AppendLine("")
}

[void]$sb.AppendLine("---")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("# 第二部分　后连续 30 页（第 $($totalPages - 29) 页 — 第 $totalPages 页）")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("> 中间部分（第 31 页至第 $($totalPages - 30) 页）按登记要求可省略，此处不再列出。")
[void]$sb.AppendLine("")

for ($p = 0; $p -lt $BACK_PAGES; $p++) {
  $pageLines = Get-PageLines $backLines $p
  $pageNo = $totalPages - $BACK_PAGES + $p + 1
  [void]$sb.AppendLine("**$NAME V1.0　·　源程序　·　第 $pageNo 页**")
  [void]$sb.AppendLine("")
  [void]$sb.AppendLine('````')
  foreach ($l in $pageLines) { [void]$sb.AppendLine($l) }
  [void]$sb.AppendLine('````')
  [void]$sb.AppendLine("")
}

$mdPath = Join-Path $root "软著-程序鉴别材料-源代码.md"
[IO.File]::WriteAllText($mdPath, $sb.ToString(), (New-Object System.Text.UTF8Encoding $true))
Write-Host "已生成: $mdPath"

# ================= HTML（打印版：浏览器 Ctrl+P 另存 PDF，每页 50 行严格分页） =================
$h = New-Object System.Text.StringBuilder
[void]$h.AppendLine('<!DOCTYPE html>')
[void]$h.AppendLine('<html lang="zh-CN">')
[void]$h.AppendLine('<head>')
[void]$h.AppendLine('<meta charset="utf-8">')
[void]$h.AppendLine("<title>$NAME V1.0 程序鉴别材料（源程序）</title>")
[void]$h.AppendLine('<style>')
[void]$h.AppendLine('@page { size: A4; margin: 14mm 13mm; }')
[void]$h.AppendLine('body { margin: 0; font-family: Consolas, "Courier New", monospace; color: #000; }')
[void]$h.AppendLine('.page { page-break-after: always; }')
[void]$h.AppendLine('.page:last-child { page-break-after: auto; }')
[void]$h.AppendLine('.ph { font-family: "Microsoft YaHei", "PingFang SC", sans-serif; font-size: 10.5pt; font-weight: 600; text-align: center; margin: 0 0 10px; }')
[void]$h.AppendLine('pre { font-size: 8.5pt; line-height: 1.3; margin: 0; white-space: pre-wrap; word-break: break-all; font-family: inherit; }')
[void]$h.AppendLine('.zh { font-family: "Microsoft YaHei", "PingFang SC", sans-serif; }')
[void]$h.AppendLine('.cover { text-align: center; padding-top: 55mm; }')
[void]$h.AppendLine('.cover h1 { font-size: 22pt; margin: 0 0 6mm; }')
[void]$h.AppendLine('.cover .ver { font-size: 14pt; color: #333; margin-bottom: 16mm; }')
[void]$h.AppendLine('.cover h2 { font-size: 15pt; margin: 0 0 12mm; }')
[void]$h.AppendLine('.cover table { margin: 0 auto 12mm; border-collapse: collapse; font-size: 11pt; }')
[void]$h.AppendLine('.cover td, .cover th { border: 1px solid #999; padding: 5px 14px; }')
[void]$h.AppendLine('.cover .tip { font-size: 10pt; color: #333; line-height: 1.8; text-align: left; width: 150mm; margin: 0 auto; }')
[void]$h.AppendLine('.section { text-align: center; padding-top: 110mm; }')
[void]$h.AppendLine('.section h1 { font-size: 18pt; margin: 0 0 8mm; }')
[void]$h.AppendLine('.section p { font-size: 12pt; color: #444; }')
[void]$h.AppendLine('</style>')
[void]$h.AppendLine('</head>')
[void]$h.AppendLine('<body>')

[void]$h.AppendLine('<div class="page cover zh">')
[void]$h.AppendLine("<h1>$NAME</h1>")
[void]$h.AppendLine('<div class="ver">V1.0</div>')
[void]$h.AppendLine('<h2>程序鉴别材料（源程序）</h2>')
[void]$h.AppendLine('<table>')
[void]$h.AppendLine("<tr><th>软件全称</th><td>$NAME</td></tr>")
[void]$h.AppendLine('<tr><th>版本号</th><td>V1.0</td></tr>')
[void]$h.AppendLine("<tr><th>开发语言</th><td>$LANGS</td></tr>")
[void]$h.AppendLine("<tr><th>源程序总量</th><td>约 $total 行（$fileCount 个源文件，合计 $totalPages 页）</td></tr>")
[void]$h.AppendLine('</table>')
[void]$h.AppendLine('<div class="tip">')
[void]$h.AppendLine('材料构成：按版权登记要求，提交源程序前连续 30 页与后连续 30 页，每页 50 行。<br>')
[void]$h.AppendLine('打印方法：浏览器（Edge / Chrome）打开本文件 → Ctrl+P → 目标打印机选择“另存为 PDF”→ 保存。')
[void]$h.AppendLine('</div>')
[void]$h.AppendLine('</div>')

[void]$h.AppendLine('<div class="page section zh">')
[void]$h.AppendLine('<h1>第一部分　前连续 30 页</h1>')
[void]$h.AppendLine('<p>第 1 页 — 第 30 页</p>')
[void]$h.AppendLine('</div>')

for ($p = 0; $p -lt $FRONT_PAGES; $p++) {
  $pageLines = Get-PageLines $frontLines $p
  [void]$h.AppendLine('<div class="page">')
  [void]$h.AppendLine("<div class=`"ph`">$NAME V1.0 · 源程序 · 第 $($p + 1) 页</div>")
  [void]$h.AppendLine('<pre>' + (($pageLines | ForEach-Object { Esc $_ }) -join "`n") + '</pre>')
  [void]$h.AppendLine('</div>')
}

[void]$h.AppendLine('<div class="page section zh">')
[void]$h.AppendLine('<h1>第二部分　后连续 30 页</h1>')
[void]$h.AppendLine("<p>第 $($totalPages - 29) 页 — 第 $totalPages 页（中间部分按登记要求省略）</p>")
[void]$h.AppendLine('</div>')

for ($p = 0; $p -lt $BACK_PAGES; $p++) {
  $pageLines = Get-PageLines $backLines $p
  $pageNo = $totalPages - $BACK_PAGES + $p + 1
  [void]$h.AppendLine('<div class="page">')
  [void]$h.AppendLine("<div class=`"ph`">$NAME V1.0 · 源程序 · 第 $pageNo 页</div>")
  [void]$h.AppendLine('<pre>' + (($pageLines | ForEach-Object { Esc $_ }) -join "`n") + '</pre>')
  [void]$h.AppendLine('</div>')
}

[void]$h.AppendLine('</body>')
[void]$h.AppendLine('</html>')

$htmlPath = Join-Path $root "软著-程序鉴别材料-源代码.html"
[IO.File]::WriteAllText($htmlPath, $h.ToString(), (New-Object System.Text.UTF8Encoding $true))
Write-Host "已生成: $htmlPath"
Write-Host ("MD: {0:N0} KB / HTML: {1:N0} KB" -f ((Get-Item $mdPath).Length / 1KB), ((Get-Item $htmlPath).Length / 1KB))
