# Phase 6 报告：应用 / 项目 / 环境 / 域名 / 部署（蓝绿）

> 状态：已完成并实测通过

## 1. 本阶段目标
项目（4 内置环境）、应用 CRUD、蓝绿部署（候选容器 → 健康检查 → Traefik 优先级切换零中断 → 旧版降级保留）、版本历史与回滚、域名绑定与 Host 路由。

## 2. 架构变化
- **蓝绿切换机制**：每个部署 = 独立容器 + 独立 Traefik router（`app{id}-v{deployId}`，priority 20）；健康检查通过后把旧版本容器重建为 priority 0 并停止（Traefik 事件驱动，流量即切；旧容器保留可回滚）。回滚 = 用历史版本镜像再部署一次（同一蓝绿路径）。
- 健康检查：HTTP（2xx/3xx）或 TCP 探测容器内网 IP:端口，2s×30 次（60s 窗口）；404 判不健康。
- **应用容器统一加入 dxcloud_edge 网络**（APP_NETWORK 可配）：自定义 bridge 才有 IPAM IP——既解决 Docker Desktop 默认 bridge 无 IP 导致无法探测的问题，又让 Traefik 同网路由。
- 部署配置快照存 config_json（供降级重建/后续滚动更新）。
- 修复：MySQL 保留字 `trigger` → `trigger_type`；域名路由优先级提升至 20。

## 3. 新增文件
backend: migrations/000005_apps.sql, internal/model/{app,project}.go, internal/repository/app_repo.go, internal/service/app_service.go, internal/handler/app.go
frontend: pages/projects/index.vue, pages/apps/{index,[id]}.vue, pages/domains/index.vue

## 4. 修改文件
backend: internal/config/config.go（APP_NETWORK）, internal/api/router.go, internal/service/app_service.go（多次修复）
frontend: types/index.ts

## 5. 数据库变化
000005_apps.sql：applications / application_versions / deployments / domains（项目与环境表 Phase 1 已有，本阶段启用）。

## 6. API 变化
GET/POST/DELETE /projects、GET /projects/:id/environments
GET/POST/GET/PUT/DELETE /applications、POST /applications/:id/deploy、GET /applications/:id/versions、POST /applications/:id/versions/:vid/rollback、GET /applications/:id/deployments
GET/POST/DELETE /domains

## 7. 前端页面变化
项目（创建/删除/环境）、应用列表（项目过滤/运行时类型）、应用详情（部署面板/域名绑定/版本历史回滚/部署记录）、域名管理（绑定/解绑）。

## 8. Docker 变化
无（应用容器运行时加入 dxcloud_edge）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（已完成）。

## 11-12. 实测验收
| # | 场景 | 结果 |
|---|---|---|
| 1 | 创建项目 shop → 4 环境 | ✅ development/testing/staging/production |
| 2 | 创建应用（registry:2.8，域名 app1.localhost，健康 /v2/） | ✅ |
| 3 | 部署 v1 | ✅ success + healthy |
| 4 | **域名路由**：curl -H Host:app1.localhost /v2/ | ✅ 返回 `{}`（registry 本体） |
| 5 | 部署 v2 → 蓝绿切换 | ✅ 新容器 running，旧容器 exited（自动降级） |
| 6 | 回滚到历史版本 | ✅ deploy success，域名服务恢复 |
| 7 | 健康检查失败场景（404 路径） | ✅ deploy failed + unhealthy + 候选清理 |
| 8 | 删除应用 → 容器全清理 | ✅ 无残留 |

## 13. 浏览器测试
/projects /apps /apps/1 /domains 全部 200 无编译错误；应用详情页可完成部署/回滚/域名绑定全流程。

## 14. 常见问题 / 已知限制
- Docker Desktop 默认 bridge 无 IP → 应用容器必须走自定义网络（APP_NETWORK=dxcloud_edge），已内建。
- 同镜像重复部署版本号带部署 ID 后缀保证唯一；无域名应用走宿主端口（切换有秒级中断，架构已知限制）。
- 滚动更新（多副本逐批）留待 Phase 8+（strategy 字段已预留）。

## 15. 下一阶段
Phase 7：Pipeline 引擎（YAML 定义/解析校验/步骤执行/状态机/取消/日志）。
