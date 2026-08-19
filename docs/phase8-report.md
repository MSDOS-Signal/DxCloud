# Phase 8 报告：Webhook + 自动部署闭环

> 状态：已完成并实测通过

## 1. 本阶段目标
Git Webhook（GitHub/GitLab/Gitee 签名校验 + 分支过滤）、docker-build（服务端 BuildKit）/docker-push/docker-deploy 步骤、Webhook → Pipeline → 构建 → 推送 → 自动部署 → 域名验证的完整 DevOps 闭环。

## 2. 架构变化
- **构建方案选择**：放弃 kaniko（gcr.io 国内不可达 + 需拉镜像），改用**服务端 BuildKit**（daemon 内 ImageBuild + workspace 卷 tar 上下文）——CI Job 容器依然零 socket 接触。
- `internal/pipeline`：新增服务端步骤执行器（docker-build：workspace tar → ImageBuild → 日志流解析；docker-push：daemon ImagePush；docker-deploy：经 Deployer 接口调 AppService 蓝绿部署）；Deployer 由路由装配注入；CreateRun 支持 ref+commit_sha。
- `internal/handler/webhook.go`：管理 CRUD（secret 经 AES-GCM 加密存储，创建时一次性返回）+ 公开接收端（GitHub X-Hub-Signature-256 HMAC、GitLab/Gitee Token、分支 glob 过滤、60/min 限流）。
- Provider 扩展：ImageBuild / ImagePush / LogsRaw（无时间戳，二进制安全）。
- 排障实录（4 个真坑）：① Timestamps=true 污染二进制 tar → LogsRaw；② Job 容器缺 WorkingDir → /workspace 统一；③ Dockerfile CMD 用 \015 非法 JSON 转义 → 被 Docker 判为 shell 形式回退（JSONArgsRecommended）→ 改 \r\n；④ busybox nc -l stdin EOF 不发数据 → nc -lk -e resp.sh 规范写法。

## 3. 新增文件
backend: migrations/000007_webhooks.sql, internal/model/webhook.go, internal/repository/webhook_repo.go, pkg/crypto/crypto.go, internal/handler/webhook.go

## 4. 修改文件
backend: internal/pipeline/engine.go（服务端步骤/Deployer/commit sha）、internal/docker/{provider,docker_provider}.go、internal/runner/job_runner.go（WorkingDir）、internal/service/app_service.go（DeployByName）、internal/repository/app_repo.go、internal/handler/pipeline.go、internal/api/router.go
frontend: types/index.ts、pages/pipelines/[id].vue（Webhook 管理区）

## 5. 数据库变化
000007_webhooks.sql：webhooks（secret_enc AES-GCM、hook_code 唯一、branch_filter）。

## 6. API 变化
GET/POST/DELETE /webhooks；POST /api/v1/webhooks/:provider/:code（公开，签名校验，限流）。

## 7. 前端页面变化
Pipeline 详情新增 Webhook 管理（创建/URL+Secret 一次性展示/列表/删除）。

## 8. Docker 变化
无（构建复用 daemon BuildKit）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（已完成）。

## 11-12. 实测验收
| # | 场景 | 结果 |
|---|---|---|
| 1-4 | **Webhook 触发完整闭环**：shell 写 Dockerfile → BuildKit 构建 → push 私有仓 → 蓝绿部署 → 健康检查 | ✅ run success（1.5s） |
| 5 | **域名验证**：Host: pipe.localhost → `<h1>pipe-ok</h1>` | ✅ |
| 6 | 部署记录 trigger=pipeline、health=healthy、蓝绿切换 | ✅ |
| 7 | 错误 HMAC 签名 | ✅ 401 invalid signature |
| 8 | 分支过滤 release/* 推 main | ✅ ignored，不触发 |
| 9 | 推 release/1.0 | ✅ 触发 run |
| 10 | Registry 目录含 default/pipetest[v5] | ✅ |

## 13. 浏览器测试
/pipelines/4（含 Webhook 面板）200 无编译错误；运行详情页四步骤状态+日志实时可见。

## 14. 常见问题 / 已知限制
- 构建上下文全量 tar 入内存（workspace 大时需流式化，留待优化）。
- wait-health 步骤类型校验通过但执行暂缺（Phase 9 与监控一起）。
- git 步骤需拉取 alpine/git 镜像（首次）；国内 git clone 依赖网络环境。

## 15. 下一阶段
Phase 9：监控（指标采样 + Dashboard 图表）/ 日志中心（系统/操作/审计检索）/ 事件与通知。
