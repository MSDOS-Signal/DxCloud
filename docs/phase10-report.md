# Phase 10 报告：多租户 / 配额 / 虚拟计费

> 状态：已完成并实测通过（验收脚本 19/19 PASS）

## 1. 本阶段目标
组织（org）→ 项目 → 资源的租户隔离；X-Org-Id 租户上下文 + 成员资格校验；组织级资源配额（ECS 实例数/CPU/内存，组织维度汇总）；虚拟计费（每小时用量结算、余额扣减、充值、余额门禁）；前端组织切换器、组织管理页、计费中心页。

## 2. 架构变化
- `internal/middleware/tenant.go`：TenantContext 中间件解析 X-Org-Id（superadmin 免检，其余校验组织成员资格），写入 gin 上下文；ecs/img/net/vol/proj/apps/dom/pipes 全部资源组挂载。
- `internal/service/org_service.go`：OrgService（创建即写入默认配额 + 初始余额 1000 + 创建者为 owner）、QuotaService（CheckEcsQuota 组织维度汇总校验，未配置项回退默认）、BillingService（Collect 按运行中实例折算 cpu_hour/mem_gb_hour/disk_gb_hour 落库并从余额扣费；Recharge 充值；HasCredit 透支门禁 >-1000）。
- `internal/service/ecs_service.go`：Create 增加配额校验 + 余额门禁（ErrNoCredit → 400）；List 按租户上下文强制组织过滤；AccessCtx 增加 OrgID。
- 组织盖章：ECS 创建写 org_id；项目/应用创建写 org_id；应用列表/详情/更新/删除/部署/回滚增加组织归属校验（appAccessible）。
- 配额修复：组织配额改为按 org_id 汇总（EcsOrgQuotaUsage），而非属主维度；未配置配额项回退默认值（避免 0 值误拒）。
- 计费调度：cmd/server/main.go 每小时 goroutine 调用 BillingService.Collect；/billing/tick 支持手动结算（运维/测试）。
- 前端：stores/org.ts（当前组织持久化 localStorage）、services/http.ts 全请求自动附加 X-Org-Id、顶栏组织切换器（切换即刷新重载）、组织页（配额抽屉/成员抽屉）、计费中心页（余额/本月用量/流水/充值/tick）。

## 3. 新增文件
backend: migrations/000009_org_quota_usage.sql, internal/model/org.go, internal/repository/org_repo.go, internal/service/org_service.go, internal/handler/org.go, internal/middleware/tenant.go
frontend: stores/org.ts, pages/orgs/index.vue, pages/billing/index.vue
tools: tools/phase10-acceptance.ps1（19 项自动化验收）

## 4. 修改文件
backend: internal/service/ecs_service.go（配额/余额门禁/组织过滤/ErrNoCredit）、internal/service/app_service.go（组织归属校验/项目名组织内唯一/应用组织过滤）、internal/service/project 相关（List/Create org 参数）、internal/handler/{ecs,app}.go、internal/repository/{ecs,app,org}_repo.go、internal/dto/ecs.go（EcsInfo.org_id）、internal/api/router.go（orgs/quotas/billing 路由 + TenantContext 挂载）、cmd/server/main.go（计费 goroutine）
frontend: services/http.ts（X-Org-Id）、layouts/default.vue（组织切换器）、types/index.ts（组织/配额/计费类型）

## 5. 数据库变化
000009_org_quota_usage.sql：resource_quotas（org_id+resource_type 唯一）、resource_usage（按小时用量）。organizations/organization_members 复用 Phase 1 表；余额存 organizations.credit。

## 6. API 变化
- 组织：GET/POST /organizations、DELETE /organizations/:id、GET /organizations/mine（返回完整组织对象）、GET/POST /organizations/:id/members、DELETE /organizations/:id/members/:uid
- 配额：GET /quotas?org_id=、PUT /quotas?org_id=（resource_type+limit_value）
- 计费：GET /billing?org_id=（余额+本月用量+单价）、GET /billing/records、POST /billing/tick、POST /billing/recharge
- 所有资源接口支持 X-Org-Id 头；非成员组织返回 403「not a member of this organization」

## 7. 前端页面变化
顶栏组织切换器（默认空间 + 我的组织）；/orgs 组织管理（列表/新建/删除/配额抽屉/成员抽屉）；/billing 计费中心（余额卡片、本月用量明细、账单流水、充值弹窗、手动结算）。导航沿用 org:list / billing:view 权限过滤。

## 8. Docker 变化
无（backend 镜像内重编译重启即可）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（前端 HMR 自动生效，零依赖变更）。

## 11-12. 实测验收（tools/phase10-acceptance.ps1，19/19 PASS）
| # | 场景 | 结果 |
|---|---|---|
| T1 | admin 登录 | ✅ |
| T2 | 创建组织 A/B（余额 1000、默认配额落库） | ✅ |
| T3 | 注册并登录 bob | ✅ |
| T4 | bob 加入组织 A（members 2 人） | ✅ |
| T5 | 项目创建带 org 盖章 | ✅ |
| T6 | bob 以 X-Org-Id=B 访问 → 403 非成员 | ✅ |
| T7 | bob 组织 A 项目列表仅见 A（无 B 泄漏） | ✅ |
| T8a | 配额 ecs_count=1 生效 | ✅ |
| T8b | ECS-1 创建成功（org 盖章） | ✅ |
| T8c | ECS-2 超额 → 400「instances 1/1」 | ✅ |
| T9 | ECS 列表按组织隔离 + org_id 字段 | ✅ |
| T10a | 计费 tick 产生用量（cpu 1 / mem 0.5 / disk 10） | ✅ |
| T10b | 余额扣减 1000→999.775 | ✅ |
| T10c | 账单流水可查（3 条） | ✅ |
| T11 | 充值 +500 → 1499.775 | ✅ |
| T12a | 余额 -1001 → 创建被拒 400「余额不足」 | ✅ |
| T12b | 充值 5000 → 创建恢复 | ✅ |
| T13 | 应用列表组织隔离（A 见 B 不见） | ✅ |
| T14 | bob 跨组织读 ECS → 403 | ✅ |

## 13. 浏览器测试
/orgs、/billing SPA 路由 200；GET /organizations/mine、/billing?org_id=3 等新端点实测返回正常；前端 HMR 零编译错误（Nuxt 3.16.2）。

## 14. 常见问题 / 已知限制
- 历史资源（Phase 9 前创建）org_id 为空：在组织上下文中不可见（兼容单租户模式，默认空间可见全部）。
- 组织删除为软删除，成员关系/配额保留（数据可审计，后续 Phase 12 备份/清理策略完善）。
- 计费为虚拟演示：单价 CPU ¥0.1/核时、内存 ¥0.05/GB·时、磁盘 ¥0.01/GB·时；透支门槛 -1000。
- 组织切换采用整页刷新策略（保证全部列表接口按新租户重载）。

## 15. 下一阶段
Phase 11 安全加固：Seccomp/AppArmor 配置、密钥托管、镜像扫描策略、容器安全基线自动化验证、限流/防爆破、敏感日志脱敏与安全审计报告。
