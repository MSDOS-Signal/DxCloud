# Phase 7 报告：Pipeline 引擎

> 状态：已完成并实测通过

## 1. 本阶段目标
自研轻量 Pipeline 引擎：YAML 定义与校验、Redis 队列调度、内嵌 Worker、隔离 Job 容器执行（git/shell 步骤）、状态机（pending/running/success/failed/canceled）、取消、超时、实时日志、前端运行详情页。

## 2. 架构变化
- `internal/pipeline`：引擎（ParseDefinition 白名单校验 → CreateRun 快照步骤+建 run/jobs → Redis LIST 入队 → 2 Worker BLPOP → 顺序执行步骤）；取消 = Redis 标志 + 强杀容器（kill 比 ticker 快，判定取消时双查 ctx 与标志）；崩溃恢复（启动时 running→failed）。
- `internal/runner`：DockerJobRunner —— 一次性 Job 容器（Provider 安全基线强制：非特权/no-new-privileges/CapDrop/PIDs 256/CPU 2/内存 2GB）、workspace 独立卷 `dxw-run-{id}`（运行结束清理）、日志流式落盘 /tmp/dxlogs（实时可读）、超时 kill、**不挂 docker.sock**。
- Provider 扩展 LogsFollow（demux 管道流）与 WaitContainer。
- 步骤类型白名单：git/shell 已实现；docker-build/push/deploy/wait-health 校验通过但执行报 Phase 8（接口已留）。
- 修复：取消优先于失败判定；失败/取消后剩余步骤标记 skipped。

## 3. 新增文件
backend: migrations/000006_pipeline.sql, internal/model/pipeline.go, internal/repository/pipeline_repo.go, internal/runner/job_runner.go, internal/pipeline/engine.go, internal/handler/pipeline.go
frontend: pages/pipelines/{index,[id]}.vue, pages/pipeline-runs/[id].vue

## 4. 修改文件
backend: internal/docker/{provider,docker_provider}.go（LogsFollow/WaitContainer）, internal/api/router.go, cmd/server/main.go, go.mod（yaml.v3）
frontend: types/index.ts

## 5. 数据库变化
000006_pipeline.sql：pipelines / pipeline_steps / pipeline_runs / pipeline_job_runs（trigger 保留字规避为 trigger_type）。

## 6. API 变化
GET/POST/PUT/DELETE /pipelines、GET /pipelines/:id、POST /pipelines/:id/run
GET /pipeline-runs（?pipeline_id）、GET /pipeline-runs/:id、GET /pipeline-runs/:id/jobs、GET /pipeline-runs/:id/logs?job_id=、POST /pipeline-runs/:id/cancel

## 7. 前端页面变化
Pipeline 列表（创建含示例 YAML）、详情（定义查看/编辑/运行/历史）、运行详情（步骤状态表 + 实时日志 2s 轮询 + 取消按钮）。

## 8. Docker 变化
无（Job 容器运行时创建）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（已完成）。

## 11-12. 实测验收
| # | 场景 | 结果 |
|---|---|---|
| 1 | 成功型（echo/ls + allow_failure 失败步骤） | ✅ success，失败步骤 skipped(exit 3)，日志含 hello-pipeline |
| 2 | 失败型（exit 7） | ✅ failed，后续步骤 skipped |
| 3 | **取消（sleep 60 → cancel）** | ✅ canceled，被杀步骤 failed，后续 skipped |
| 4 | YAML 校验（非法类型 rm-rf） | ✅ 400 白名单拒绝 |
| 5 | Job 容器清理 | ✅ 运行后无残留 |

## 13. 浏览器测试
/pipelines /pipelines/1 /pipeline-runs/1 全 200 无编译错误；运行详情页日志实时滚动。

## 14. 常见问题 / 已知限制
- Job 镜像 alpine:3.20 / alpine/git:latest（git 步骤首次需拉取镜像）。
- 日志存容器内 /tmp/dxlogs（重启丢失）；WS 实时推送在 Phase 9。
- docker-* 步骤类型 Phase 8 接入（kaniko 构建 + 平台侧部署）。

## 15. 下一阶段
Phase 8：Runner 增强 + Git Webhook（GitHub/GitLab/Gitee 签名校验）+ kaniko docker-build/push + 自动部署（webhook → pipeline → deploy 闭环）。
