# Phase 3 报告：ECS 云主机核心

> 状态：已完成并实测通过

## 1. 本阶段目标
ComputeProvider 抽象 + DockerProvider、ECS 创建/启动/停止/重启/强制停止/删除/日志/Stats/事件、双状态模型 + Reconciler 对账、配额初版、端口冲突检测、前端列表/创建/详情三页。

## 2. 架构变化
- `internal/docker`：ComputeProvider 接口 + DockerProvider（Docker SDK v26.1.5，API 协商；容器安全基线代码级强制：非特权/no-new-privileges/CapDrop ALL/PidsLimit 256/swap=0/CPU 内存上限）。
- `internal/service`：ECS 状态机（creating/running/stopped/... + 90s 转换超时）、配额（5 实例/8 核/16GB）、端口双重冲突检测（Docker 运行时 + DB）、事件与审计埋点。
- `internal/scheduler`：Reconciler 每 15s 对账（过渡态推进 + 稳态漂移 + 孤儿容器发现）。
- 命名/标签锚点：容器名 `dx-{instance_no}`，标签 com.dxcloud.{kind,instance-id,owner-id}。
- 工程修复：docker SDK 为 +incompatible 模块 → 显式 `pkg/errors@v0.9.1`；HostConfig 内嵌 Resources 复合字面量写法；工具链升级 go 1.25（golang:1.25-alpine）。

## 3. 新增文件
backend: internal/docker/{provider,docker_provider}.go, internal/model/ecs.go, internal/repository/ecs_repo.go,
internal/service/{ecs_service,ecs_req}.go, internal/scheduler/reconciler.go, internal/handler/ecs.go,
internal/dto/ecs.go, migrations/000003_ecs.sql
frontend: pages/ecs/{index,create,[id]}.vue

## 4. 修改文件
backend: internal/api/router.go, cmd/server/main.go, go.mod/go.sum, Dockerfile(.dev)（go1.25）
frontend: types/index.ts

## 5. 数据库变化
000003_ecs.sql：ecs_instances（双状态列 + container_id 唯一 + owner/state 索引）、ecs_instance_events。

## 6. API 变化
GET/POST /ecs · GET/PUT/DELETE /ecs/:id · POST /ecs/:id/{start,stop,force-stop,restart} · GET /ecs/:id/{logs,stats,events}（全部 RBAC + 属主校验；user 角色 SQL 层强制 owner 过滤）。

## 7. 前端页面变化
ECS 列表（5s 轮询、状态标签、启动/停止/重启/强停/删除/详情）、创建页（镜像预设/规格/端口动态行/环境变量/启动命令/重启策略/只读根盘）、详情页（基本信息/实时监控 5s 刷新/日志/事件）。

## 8. Docker 变化
golang:1.25-alpine 基础镜像（dev+prod）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose build backend && docker compose up -d backend（已完成）。

## 11-12. 实测验收（13 项全过）
| # | 场景 | 结果 |
|---|---|---|
| 1 | 创建 alpine ECS（0.5核/256MB/命令） | ✅ running |
| 2 | 列表/详情 | ✅ |
| 3 | Stats（cpu/mem/net/disk/pids） | ✅ mem_limit=256MB |
| 4 | Logs | ✅ 含 hello-dxcloud |
| 5-6 | stop/start/restart/force-stop 全循环 | ✅ 状态正确 |
| 7 | 事件列表 | ✅ 11 条事件 |
| 8 | 端口冲突（18080×2） | ✅ 409 运行时检出 |
| 9 | alice 停 admin 实例 | ✅ 403（属主校验） |
| 10 | alice 列表仅见自己 | ✅ total=0 |
| 11 | 配额 5 实例 | ✅ 第 6 个 400 |
| 12 | Reconciler：外部 docker rm -f | ✅ 45s 内置 Unknown |
| 13 | 删除实例 | ✅ 404 后不可见 |

## 13. 浏览器测试
http://localhost/ecs → 列表；/ecs/create → 创建（提交后跳详情）；详情页监控/日志/事件卡片。前端无编译错误（Vite 日志检查）。

## 14. 常见问题 / 已知限制
- Docker Desktop 默认 bridge 不分配容器 IP（fixed_ip 为空属预期）→ Phase 5 自定义网络 IPAM 提供静态 IP。
- 磁盘为逻辑配额；日志 REST 为 tail 模式（实时跟随在 Phase 4 WebSocket）。
- docker SDK 是 +incompatible 模块：新增依赖后若出现 errors.As 编译错，检查 pkg/errors ≥ v0.9.1。

## 15. 下一阶段
Phase 4：Web Terminal —— xterm.js + WebSocket + docker exec TTY、resize、一次性 token、空闲超时、会话审计。
