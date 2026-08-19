# ============================================================
# backup.ps1 —— DxCloud 全量备份（MySQL 逻辑导出 + Redis AOF/RDB + Registry 卷）
# 用法（项目根目录）：
#   .\tools\backup.ps1                     # 备份到 backups\backup-<时间戳>\
#   .\tools\backup.ps1 -OutDir D:\bk\dx    # 指定目录
# ============================================================
param(
    [string]$OutDir = ""
)

$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

if (-not $OutDir) {
    $ts = Get-Date -Format "yyyyMMdd-HHmmss"
    $OutDir = Join-Path $root ("backups\backup-" + $ts)
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
Write-Host "[1/5] 备份目录：$OutDir"

# ---------- MySQL 逻辑备份（容器内执行，避免 PowerShell 编码污染） ----------
Write-Host "[2/5] MySQL 导出中…"
docker exec dx-mysql sh -c 'rm -f /tmp/dxcloud.sql && exec mysqldump -udxcloud -p"$MYSQL_PASSWORD" --single-transaction --routines --triggers dxcloud > /tmp/dxcloud.sql'
if ($LASTEXITCODE -ne 0) { throw "mysqldump 失败" }
docker cp dx-mysql:/tmp/dxcloud.sql (Join-Path $OutDir "mysql.sql") | Out-Null
$mysqlSize = (Get-Item (Join-Path $OutDir "mysql.sql")).Length
Write-Host "      mysql.sql  $([math]::Round($mysqlSize/1KB,1)) KB"

# ---------- Redis（BGSAVE + 拷贝整个 data 目录，含 AOF） ----------
Write-Host "[3/5] Redis BGSAVE…"
docker exec dx-redis redis-cli -a dxcloud_redis_dev BGSAVE | Out-Null
docker cp dx-redis:/data (Join-Path $OutDir "redis-data") | Out-Null
Write-Host "      redis-data 已拷贝"

# ---------- Registry 数据卷（tar 打包） ----------
Write-Host "[4/5] Registry 卷打包…"
$volName = (docker volume ls --format "{{.Name}}" | Select-String -Pattern "registry-data" | Select-Object -First 1).ToString().Trim()
$outUnix = $OutDir.Replace("\", "/")
$null = docker run --rm -v "${volName}:/data" -v "${outUnix}:/backup" alpine:3.20 tar czf /backup/registry-data.tgz -C /data .
if ($LASTEXITCODE -ne 0) { Write-Warning "registry 卷打包失败（可能卷为空或 alpine 镜像未拉取）" }

# ---------- 清单 ----------
Write-Host "[5/5] 生成清单…"
@(
    "DxCloud 备份清单"
    "时间：$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    "mysql.sql      MySQL 逻辑导出（single-transaction + routines + triggers）"
    "redis-data/    Redis AOF + RDB（BGSAVE 后拷贝）"
    "registry-data.tgz  Registry 存储卷 tar 包"
    ""
    "恢复：.\tools\restore.ps1 -BackupDir `"$OutDir`""
) | Out-File -Encoding utf8 (Join-Path $OutDir "README.txt")

Write-Host "备份完成：$OutDir"
Write-Host "  ls：" + ((Get-ChildItem $OutDir | ForEach-Object { $_.Name }) -join ", ")
