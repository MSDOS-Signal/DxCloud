# Phase 9 报告：监控 / 日志中心 / 审计 / 通知

> 状态：已完成并实测通过

## 1. 本阶段目标
指标采样（分钟级落库 + 7 天保留）、Dashboard 聚合与曲线（ECharts）、实例监控页、日志中心（操作/审计/登录三合一检索）、Pipeline 完成站内通知。

## 2. 架构变化
- `internal/scheduler` 新增 MetricsCollector：每分钟并行（并发上限 8）对运行中 ECS 采样 StatsOneShot → metric_samples；每 6 小时清理 7 天前数据。
- `internal/service` 新增 MonitorService：Dashboard 聚合（实例/应用/Pipeline/今日部署/24h 成功率/CPU·内存水位）、分钟级曲线聚合（SQL GROUP BY bucket）。
- `internal/handler/monitor.go`：/monitor/{dashboard,series}、/logs（三类型）、/notifications（列表/已读/全部已读）。
- Pipeline finishRun 增加通知写入（触发者收 Pipeline 结果）。
- ECS 关键操作（create/start/stop/restart/force-stop/delete）补操作日志埋点。
- 前端：ECharts 组合式封装 useEcharts、Dashboard 真实指标+双曲线、实例监控页（30m/2h/6h）、日志中心三 Tab、顶栏通知铃铛（未读数+点击已读）。
- 工程修复：nuxt 3.21.11 的 `#app-manifest` 解析错误 → 钉回 3.16.2 并 npm ci 全量重建依赖。

## 3. 新增文件
backend: migrations/000008_ops.sql, internal/model/ops.go, internal/repository/ops_repo.go, internal/service/monitor_service.go, internal/scheduler/metrics.go, internal/handler/monitor.go
frontend: composables/useEcharts.ts, pages/ecs/[id]/monitor.vue, pages/logs/index.vue

## 4. 修改文件
backend: internal/pipeline/engine.go（通知）、internal/service/ecs_service.go（操作日志埋点）、internal/api/router.go、cmd/server/main.go
frontend: package.json（echarts、nuxt 钉版）、pages/dashboard/index.vue（真实数据+图表）、layouts/default.vue（通知铃铛）

## 5. 数据库变化
000008_ops.sql：metric_samples / operation_logs / notifications（含索引）。

## 6. API 变化
GET /monitor/dashboard、GET /monitor/series（kind/ref_id/minutes）
GET /logs?type=operation|audit|login&keyword&page
GET /notifications（unread 过滤）、PUT /notifications/:id/read、POST /notifications/read-all

## 7. 前端页面变化
Dashboard（6 统计卡真实数据 + CPU/内存 ECharts 曲线 + 水位条，30s 刷新）、实例监控页、日志中心（三 Tab + 搜索 + 分页）、顶栏通知（未读角标/点击已读/30s 轮询）。

## 8. Docker 变化
无。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（已完成）。

## 11-12. 实测验收
| # | 场景 | 结果 |
|---|---|---|
| 1 | 运行中实例 → 采样器 75s 落库 | ✅ metric_samples 有行 |
| 2 | Dashboard 聚合 | ✅ 实例/应用/Pipeline/今日部署/成功率/CPU/内存水位 |
| 3 | 30 分钟曲线 | ✅ 分钟级 bucket |
| 4 | 操作日志（stop/start/delete） | ✅ 含结果与耗时 |
| 5 | 审计日志 | ✅ 202 条可检索 |
| 6 | 登录日志 | ✅ 59 条 |
| 7 | **通知闭环**：Pipeline 完成 → 通知 → 未读 1 → 标记已读 → 0 | ✅ |

## 13. 浏览器测试
/dashboard /logs /ecs/:id/monitor 全 200；前端零编译错误（nuxt 钉版后）。

## 14. 常见问题 / 已知限制
- 采样周期 1 分钟（两帧差值 CPU%），曲线最小粒度 1 分钟。
- 应用容器（deployment）采样已预留 kind=app，采集接入随 Phase 11 完善。
- 通知暂仅 Pipeline 结果（部署事件通知后续补）。

## 15. 下一阶段
Phase 10：Multi-Tenant 完整化（组织/成员/角色）+ 配额 UI + 虚拟计费 + 用量报表。
