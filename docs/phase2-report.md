# Phase 2 报告：认证 + RBAC

> 状态：已完成并实测通过

## 1. 本阶段目标
注册/登录/JWT + Refresh 轮换/退出/改密/会话管理、6 默认角色 + 69 权限点、后端 RBAC 强制链、登录日志与审计日志、限流，前端真实登录/注册与 IAM 管理页。

## 2. 架构变化
- 新增 `internal/iam`（认证/令牌/RBAC/审计决策）、`internal/middleware`（auth/rbac/ratelimit）、`internal/repository`（含 scope.go 租户隔离铁律）、`pkg/jwt`、`pkg/ratelimit`。
- 认证链路：JWT Access（15min，jti 黑名单） + 不透明 Refresh（Redis HASH，7 天，一次性轮换）。
- RBAC 强制链：中间件 AuthRequired → RequirePerm（Redis 权限缓存 5min，角色变更主动失效）；拒绝写审计（authz.deny）。
- 统一响应增强：40000→HTTP 400、40009→409、42900→429；前端信封 code!=0 一律抛错（修复了 40000 曾以 HTTP 200 返回的缺陷）。

## 3. 新增文件
backend: internal/model/iam.go, internal/repository/{repo,scope}.go, internal/iam/{service,admin,json}.go,
internal/middleware/{auth,rbac,ratelimit}.go, internal/handler/{auth,user,role,permission}.go,
internal/database/seed.go, migrations/000002_audit_logs.sql, pkg/{jwt,ratelimit}, pkg/redisx/helpers.go
frontend: utils/token.ts, pages/register.vue, pages/iam/users|roles|permissions/index.vue

## 4. 修改文件
backend: internal/api/router.go, cmd/server/main.go, pkg/resp/resp.go, pkg/errcode/errcode.go, internal/iam/service.go
frontend: services/http.ts, stores/auth.ts, layouts/default.vue, pages/login.vue, pages/dashboard/index.vue, middleware/auth.global.ts, types/index.ts

## 5. 数据库变化
- 000002_audit_logs.sql：新增 audit_logs（含 org_id/user_id/action/request_id/status 索引）。
- 种子（Go 幂等）：6 角色、69 权限点、角色-权限映射、初始管理员 admin/Admin@123456（环境变量 ADMIN_INIT_PASSWORD 可覆盖，默认密码仅提示修改）。

## 6. API 变化
POST /auth/register|login|refresh（限流 20/5/60 每分钟）· POST /auth/logout · GET /auth/me · PUT /auth/password · GET /auth/sessions · DELETE /auth/sessions/:id
GET/POST /users · GET/PUT/DELETE /users/:id · PUT /users/:id/roles · GET/POST /roles · PUT/DELETE /roles/:id · PUT /roles/:id/permissions · GET /permissions（全部经 user:* 权限强制）

## 7. 前端页面变化
登录页接真实 API（提示初始管理员）、新增注册页、IAM 用户/角色/权限三页（列表+新建+分配角色+权限配置，按权限隐藏按钮）、顶栏用户菜单显示昵称并支持退出、菜单按权限过滤。

## 8. Docker 变化
无（沿用 Phase 1；后端热重启即生效）。

## 9. 完整代码
见仓库（backend/、frontend/，不重复粘贴）。

## 10. 启动命令
docker compose restart backend（已完成）；前端 HMR 自动生效。

## 11. 测试方法
go vet/build 通过；curl/Invoke-RestMethod 12 项验收；MySQL 查询登录日志。

## 12. curl 测试（实测结果）
| # | 场景 | 结果 |
|---|---|---|
| 1 | alice 注册 → 发 Token | ✅ code=0 |
| 2 | admin 登录（superadmin，69 权限） | ✅ |
| 3 | 错误密码登录 | ✅ 40100 |
| 4 | alice GET /users | ✅ 403 40001（RBAC 生效） |
| 5 | alice /auth/me | ✅ roles=[user] |
| 6 | 改密-错误旧密码 | ✅ 40000 "old password incorrect" |
| 7 | Refresh 轮换 | ✅ 新对签发 |
| 8 | 旧 refresh 复用 | ✅ 401 |
| 9 | Logout 后 access 复用 | ✅ 401（黑名单） |
| 10 | Logout 后 refresh 复用 | ✅ 401 |
| 11 | 会话列表 | ✅ 3 会话，7 天过期 |
| 12 | 登录日志落库 | ✅ 含 IP/结果 |
| 13 | 登录限流（5/min） | ✅ 触发 42900 |

## 13. 浏览器测试
http://localhost → 登录页（admin/Admin@123456）→ Dashboard 连通性卡片 code=0 → 左侧「系统 → IAM·用户」可管理用户/角色/权限；注册页 http://localhost/register 可注册并自动登录。

## 14. 常见问题
- 429 too many requests：登录限流 5 次/分钟/IP，等待 1 分钟。
- 改密后自动退出：设计行为（改密撤销全部会话）。
- 前端 401 自动刷新重试：http.ts 内建并发锁，无需手动处理。

## 15. 下一阶段
Phase 3：ECS 核心 —— CloudProvider 抽象 + DockerProvider、ecs_instances 迁移、创建/启动/停止/重启/删除/日志/Stats/事件、Reconciler 第一版、配额初版、ECS 列表/创建/详情前端页。
