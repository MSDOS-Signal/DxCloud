# ============================================================
# restore.ps1 —— 从 backups\backup-<时间戳>\ 恢复 DxCloud 数据
# 用法（项目根目录）：
#   .\tools\restore.ps1 -BackupDir .\backups\backup-20260819-120000
# 注意：恢复会覆盖当前数据；建议先停止 backend 服务避免写入竞争
# ============================================================
param(
    [Parameter(Mandatory = $true)]
    [string]$BackupDir
)

$ErrorActionPreference = "Stop"
$BackupDir = (Resolve-Path $BackupDir).Path

if (-not (Test-Path (Join-Path $BackupDir "mysql.sql"))) {
    throw "备份目录缺少 mysql.sql：$BackupDir"
}

Write-Host "[提示] 恢复期间请勿操作平台；如 backend 运行中，建议先 docker compose stop backend"

# ---------- MySQL ----------
Write-Host "[1/3] MySQL 恢复中…"
docker cp (Join-Path $BackupDir "mysql.sql") dx-mysql:/tmp/dxcloud.sql
if ($LASTEXITCODE -ne 0) { throw "拷贝 SQL 失败" }
docker exec dx-mysql sh -c 'exec mysql -udxcloud -p"$MYSQL_PASSWORD" dxcloud < /tmp/dxcloud.sql && rm -f /tmp/dxcloud.sql'
if ($LASTEXITCODE -ne 0) { throw "MySQL 导入失败" }
Write-Host "      mysql 导入完成"

# ---------- Redis ----------
Write-Host "[2/3] Redis 恢复中…"
$redisVol = (docker volume ls --format "{{.Name}}" | Select-String -Pattern "redis-data" | Select-Object -First 1).ToString().Trim()
$srcUnix = (Join-Path $BackupDir "redis-data").Replace("\", "/")
$null = docker run --rm -v "${redisVol}:/data" -v "${srcUnix}:/backup" alpine:3.20 sh -c 'cp -a /backup/. /data/'
if ($LASTEXITCODE -ne 0) { throw "Redis 数据恢复失败" }
Write-Host "      redis 数据已还原（重启 redis 容器后生效：docker compose restart redis）"

# ---------- Registry ----------
Write-Host "[3/3] Registry 恢复中…"
$regVol = (docker volume ls --format "{{.Name}}" | Select-String -Pattern "registry-data" | Select-Object -First 1).ToString().Trim()
$bkUnix = $BackupDir.Replace("\", "/")
if (Test-Path (Join-Path $BackupDir "registry-data.tgz")) {
    $null = docker run --rm -v "${regVol}:/data" -v "${bkUnix}:/backup" alpine:3.20 tar xzf /backup/registry-data.tgz -C /data
    if ($LASTEXITCODE -ne 0) { Write-Warning "registry 卷恢复失败" }
    Write-Host "      registry 卷已还原"
} else {
    Write-Warning "备份中无 registry-data.tgz，跳过"
}

Write-Host ""
Write-Host "恢复完成。请执行：docker compose restart backend redis"
Write-Host "验证：登录控制台检查项目/实例/镜像数据；docker compose ps 全绿。"
