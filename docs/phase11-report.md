# Phase 11 报告：安全加固

> 状态：已完成并实测通过（验收脚本 12/12 PASS）

## 1. 本阶段目标
容器安全基线自动化审计、镜像策略扫描、密钥托管（加密存储 + 租户隔离 + 解密审计）、登录防爆破锁定、敏感日志脱敏，以及安全中心 Web 页面。

## 2. 架构变化
- `internal/service/security_service.go`：SecurityService（ScanBaseline 容器基线审计 / ScanImages 镜像策略 / ScanAll / Dashboard / Reports / ReportFindings，发现项分级 high/medium/low/info 计分：-10/-5/-2/0）；SecretService（AES-256-GCM 加密存储，组织维度隔离，Reveal 后端校验归属）。
- `internal/docker`：新增 `SecurityAudit` 能力与 `ContainerSecurity` 结构（特权/no-new-privileges/CapAdd/CapDrop/只读根/PidsLimit/CPU/内存上限/User/SecurityOpt），DockerProvider 全量 Inspect 实现；Provider 接口同步扩展。
- 基线规则（仅平台托管容器 `com.dxcloud.kind` 计分，基础设施容器仅 info 提示）：privileged → 高危；缺 no-new-privileges → 高危；未 CapDrop ALL → 高危；危险能力（ALL/SYS_ADMIN/NET_ADMIN/SYS_PTRACE/SYS_MODULE）→ 高危；root 运行 → 中危；缺 CPU/内存上限、PidsLimit 越界 → 低危。
- 镜像策略：`:latest` 标签 → 中危（生产应使用不可变版本）；悬空镜像 → info；体积 >2GB → 低危。
- `internal/iam`：登录防爆破 —— Redis 计数 `dx:login:fail:{username}`，连续失败 5 次锁定 15 分钟（正确密码同样拒绝，返回「账号已锁定」），成功登录清零；Audit 统一敏感字段脱敏（pkg/redact：password/token/secret/authorization 等键 → `***`，递归）。
- `internal/handler/security.go` + 路由：/security/{dashboard,scan,reports,reports/:id}（security:view / security:scan）、/secrets CRUD+reveal（secret:* 权限 + TenantContext 组织隔离）。
- 种子：新增 security:* / secret:* 6 个权限点；admin 全量、developer 增 secret:list/create/reveal、superadmin 自动全量。
- 前端：/security 安全中心页（综合得分、规则清单、最新发现、扫描历史、密钥托管面板）+ 顶栏导航入口。

## 3. 新增文件
backend: migrations/000010_security.sql, internal/model/security.go, internal/repository/security_repo.go, internal/service/security_service.go, internal/handler/security.go, pkg/redact/redact.go
frontend: pages/security/index.vue
tools: tools/phase11-acceptance.ps1（12 项自动化验收）

## 4. 修改文件
backend: internal/docker/provider.go（ContainerSecurity + SecurityAudit 接口）、internal/docker/docker_provider.go（实现）、internal/iam/service.go（防爆破 + Audit 脱敏 + ErrAccountLocked）、internal/handler/auth.go（锁定提示）、internal/api/router.go（security/secrets 路由）、internal/database/seed.go（权限点与角色映射）
frontend: types/index.ts（SecurityFinding/SecurityDashboard/SecretItem 等）、layouts/default.vue（安全中心导航）

## 5. 数据库变化
000010_security.sql：secrets（org_id+name+deleted_at 唯一，value_cipher 密文列）、security_reports（kind/score/finding_count/summary JSON）。

## 6. API 变化
- GET /security/dashboard（综合得分 + 最新各维度发现 + 规则清单）
- POST /security/scan（基线 + 镜像全量扫描，落报告）
- GET /security/reports、GET /security/reports/:id（发现项明细）
- GET/POST /secrets、GET /secrets/:id/reveal、DELETE /secrets/:id（X-Org-Id 组织隔离，跨组织 403）
- POST /auth/login：连续失败 5 次 → 401「账号已锁定：连续失败 5 次，请 15 分钟后重试」

## 7. 前端页面变化
新增 /security 安全中心：综合得分大数字（颜色分级）、容器基线/镜像策略规则清单、最新扫描发现（严重级别标签）、扫描历史表格、密钥托管面板（新建/解密查看/删除 + 当前组织提示）；导航「运营 → 安全中心」。

## 8. Docker 变化
无。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（迁移 000010 + 权限种子自动应用；前端 HMR 生效）。

## 11-12. 实测验收（tools/phase11-acceptance.ps1，12/12 PASS）
| # | 场景 | 结果 |
|---|---|---|
| S1 | admin 登录 | ✅ |
| S2 | 安全总览（score/finding_count） | ✅ |
| S3 | 全量扫描产出基线+镜像报告（baseline 65 分） | ✅ |
| S4 | 扫描历史 + 报告发现项明细 | ✅ |
| S5a | 密钥加密创建（响应不含明文/密文） | ✅ |
| S5b | 组织 A 可见本组织密钥 | ✅ |
| S5c | 组织 B 隔离（列表 0 条） | ✅ |
| S5d | 解密读取值一致（Sup3rS3cret#2026） | ✅ |
| S5e | 跨组织解密 → 403 | ✅ |
| S5f | 密钥删除 | ✅ |
| S6a | 5 次失败后锁定：正确密码也 401「账号已锁定」 | ✅ |
| S6b | 锁定期满（计数清除）恢复登录 | ✅ |

## 13. 浏览器测试
/security SPA 路由 200；HMR 路由注册零错误（Nuxt 3.16.2）；扫描报告实时可见（baseline 65 分 / 26 项发现）。

## 14. 常见问题 / 已知限制
- 基线审计覆盖平台上创建的全部容器；MySQL/Redis/Traefik 等 Compose 基础设施容器仅 info 提示不计分（未贴 com.dxcloud.kind 标签）。
- 基线得分 65 主要来自既有演示实例以 root 运行、部分容器未设资源上限 —— 平台新创建的资源均自动满足基线，存量资源可在控制台重建后消除。
- Seccomp：容器默认启用 Docker 守护进程 seccomp 配置（未使用 unconfined）；自定义 seccomp profile 需宿主机放置 profile 文件，文档已说明开关预留（SECCOMP_PROFILE 环境变量，Phase 12 生产 compose 启用）。
- 镜像漏洞扫描（trivy 等）受网络环境限制未接入，以策略扫描（latest/悬空/体积）+ 私有 Registry 管控替代，接口已预留扩展点。
- 登录限流 5 次/分钟/IP 与账号锁定 5 次/15 分钟叠加生效（双层防爆破）。

## 15. 下一阶段
Phase 12 生产部署：docker-compose.prod.yml 收紧（HTTPS/TLS、非 root 运行、资源限制、seccomp 开关）、数据备份/恢复脚本（MySQL/Redis/卷）、健康探针与优雅停机、生产环境启动实测。
