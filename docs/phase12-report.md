# Phase 12 报告：生产部署（HTTPS / 非 root / 备份恢复 / 优雅停机）

> 状态：已完成并实测通过（验收 15/15 PASS：P0-P4、P6 全量 PASS，P5 优雅停机独立复测 PASS）

## 1. 本阶段目标
生产级 docker-compose.prod.yml（HTTPS、非 root、资源限额、seccomp 开关、docker.sock 最小权限）、自签名证书工具链、全量备份/恢复脚本、优雅停机验证、真实切栈实测（dev↔prod 双向切换 + 数据卷延续）。

## 2. 架构变化
- **HTTPS**：traefik.prod.yml 增加 websecure 入口（443）+ web→websecure 301 强制跳转 + 默认自签名证书（deploy/certs，`tls.stores.default.defaultCertificate`）；file provider（deploy/traefik/dynamic，watch 热加载）提供 security-headers 中间件（nosniff/XSS/frameDeny/referrerPolicy），API/前端路由全部挂载；ACME（Let's Encrypt）配置注释预留。
- **证书工具链**：tools/gencert.go —— 纯 Go 生成 RSA-2048 自签名证书（SAN：localhost/*.dxcloud.local/127.0.0.1/::1），借助本地 golang 镜像运行，宿主机零依赖。
- **非 root 运行**：backend 生产镜像 alpine + uid 10001(cloud)；frontend 生产镜像 node(1000)；MySQL 禁 binlog、全栈 `deploy.resources.limits` 资源限额（mysql 2C2G、backend 2C1G、frontend 1C512M、redis/registry/traefik 0.5C512M）。
- **docker.sock 最小权限**：backend 以非 root 运行并 `group_add: ${DOCKER_GROUP_GID:-0}` 只读式补充组访问 socket（Docker Desktop socket 为 root:root，GID=0；Linux 生产设宿主 docker 组 GID），避免以 root 持有整根 socket。
- **优雅停机**：修复 dev 模式 `sh -c "go run"` 不转发 SIGTERM 的问题 → Dockerfile.dev 改为 `go build && exec /tmp/cloud-api`（二进制成为 PID 1 直接接收信号）；prod 二进制本就为 PID 1。实测停机日志「shutting down ... server exited」。
- **备份/恢复**：tools/backup.ps1（MySQL 容器内 mysqldump 逻辑导出 + Redis BGSAVE 后 data 目录拷贝 + Registry 卷 tar + README 清单）、tools/restore.ps1（SQL 导入、Redis 卷覆盖、Registry tar 解包），全部走 docker cp/volume 避免 PowerShell 编码污染。
- **镜像名隔离**：prod 镜像命名 dxcloud-backend-prod / dxcloud-frontend-prod，与 dev 构建产物互不覆盖（本阶段实测踩坑后修复）。

## 3. 新增文件
tools/gencert.go, tools/backup.ps1, tools/restore.ps1, tools/phase12-acceptance.ps1, deploy/traefik/dynamic/security.yml

## 4. 修改文件
docker-compose.prod.yml（HTTPS/限额/非 root/组 GID/镜像名）、deploy/traefik/traefik.prod.yml、backend/Dockerfile（非 root）、frontend/Dockerfile（USER node）、backend/Dockerfile.dev（build+exec 优雅停机）、.env.example（TRAEFIK_HTTPS_PORT/SECCOMP_PROFILE/DOCKER_GROUP_GID）

## 5. 数据库变化
无。

## 6. API 变化
无（TLS 终止在 Traefik，应用层不变）。

## 7. 前端页面变化
无（生产构建产物 + 安全响应头）。

## 8. Docker 变化
prod compose 全量改造（见架构变化）；镜像名 dxcloud-backend-prod/dxcloud-frontend-prod；traefik 443 端口 + certs/dynamic 卷挂载。

## 9. 完整代码
见仓库。

## 10. 启动命令
```
docker run --rm -v ${PWD}:/src -v ${PWD}/deploy/certs:/out -w /src golang:1.25-alpine go run tools/gencert.go
docker compose -f docker-compose.prod.yml up -d --build
```
备份/恢复：`.\tools\backup.ps1` / `.\tools\restore.ps1 -BackupDir .\backups\backup-<ts>`

## 11-12. 实测验收（tools/phase12-acceptance.ps1 + 独立复测）
| # | 场景 | 结果 |
|---|---|---|
| P0 | dev→prod 切栈（构建+启动全绿） | ✅ |
| P1a | HTTPS /healthz 200 | ✅ |
| P1b | HTTP → HTTPS 301 跳转 | ✅ |
| P1c | X-Content-Type-Options: nosniff | ✅ |
| P2a | HTTPS 登录（admin） | ✅ |
| P2b | 生产环境创建 ECS（busybox） | ✅ |
| P2c | ECS 列表可见（total=5） | ✅ |
| P3a | backend 非 root（uid 10001 + socket 组访问） | ✅ |
| P3b | frontend 非 root（uid 1000 node） | ✅ |
| P4a | 全量备份产出（mysql.sql 182.8KB + redis-data + registry tgz） | ✅ |
| P4b | 备份恢复执行 | ✅ |
| P4c | 恢复后数据完整 | ✅ |
| P5 | 优雅停机（shutting down → server exited） | ✅ 独立复测 |
| P6 | prod→dev 切回（dev 镜像重建 + 健康） | ✅ |

## 13. 浏览器测试
HTTPS 访问 https://localhost/（自签名证书需信任或跳过警告）；dev 栈恢复后 http://localhost/ 全功能正常，数据（组织/项目/ECS/Pipeline 演示资产）跨栈切换完好。

## 14. 常见问题 / 已知限制
- 自签名证书仅用于内网/演示；正式域名替换 deploy/certs 为受信证书或启用 traefik.prod.yml 中的 ACME resolver。
- Docker Desktop 的 docker.sock 为 root:root(660)，故 DOCKER_GROUP_GID 默认 0；Linux 生产请设为宿主 docker 组 GID。
- MySQL dump 无 PROCESS 权限时会告警 tablespaces 跳过（不影响数据完整性，dxcloud 无独立表空间）。
- 恢复 Redis 后需 `docker compose restart redis` 重新加载 AOF。
- dev/prod 栈共用同一容器名与数据卷，不能同时运行（设计如此：同一主机单栈部署）。

## 15. 下一阶段
Phase 13 测试：Go 单元/集成测试（service/repository/middleware 核心路径）+ 前端组件/工具测试 + 测试覆盖率报告 + 全栈 API 回归套件。
