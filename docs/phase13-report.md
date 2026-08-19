# Phase 13 报告：测试（单元测试 + 全栈回归）

> 状态：已完成并实测通过（单元测试全绿；回归套件 14/14 PASS）

## 1. 本阶段目标
Go 单元测试（核心纯逻辑 + sqlite 内存库服务层）、测试覆盖率统计、全栈 API 回归套件（注册→项目→ECS→终端→部署→Pipeline→监控→安全→计费→清理）。

## 2. 架构变化
- `internal/service/security_service.go` 抽出纯函数 `computeScore`（发现项计分）供单测，无行为变化。
- 测试依赖引入 `github.com/glebarez/sqlite`（纯 Go，无 CGO，宿主机/容器均可跑）。
- tools/ws-test/terminal-any.mjs：参数化（ECS_ID/ADMIN_TOKEN 环境变量）的 Web 终端 E2E。
- tools/phase13-regression.ps1：14 项全栈回归套件（BOM UTF-8，PS5.1 兼容）。

## 3. 新增文件
backend: pkg/crypto/crypto_test.go, pkg/redact/redact_test.go, pkg/jwt/jwt_test.go, internal/pipeline/engine_test.go, internal/service/security_service_test.go, internal/service/org_service_test.go
tools: tools/phase13-regression.ps1, tools/ws-test/terminal-any.mjs

## 4. 修改文件
backend: internal/service/security_service.go（computeScore 抽取）、go.mod/go.sum（sqlite 测试依赖）

## 5. 数据库变化
无（测试用 :memory: sqlite，不动 MySQL）。

## 6. API 变化
无。

## 7. 前端页面变化
无（前端以端到端验收与 HMR 编译零错误作为验证手段）。

## 8. Docker 变化
无。

## 9. 完整代码
见仓库。

## 10. 启动命令
```
cd backend
$env:GOPROXY='https://goproxy.cn,direct'
go test ./... -count=1
.\tools\phase13-regression.ps1   # 项目根目录，dev 栈运行中
```

## 11-12. 实测验收
**单元测试（go test ./...，全绿）**
| 包 | 覆盖 | 用例 |
|---|---|---|
| pkg/crypto | 82.1% | AES-GCM 加解密回环 / 随机 nonce / 错密钥 / 篡改 / 垃圾输入 |
| pkg/jwt | 87.5% | 签发解析回环 / 错密钥 / 过期 / 垃圾 token |
| pkg/redact | 76.2% | 敏感键掩码（含嵌套与大小写）/ 非敏感透传 |
| internal/pipeline | 4.9%（解析部分全路径） | 定义解析：合法/空/无步骤/坏 YAML/步骤校验（4 类错误） |
| internal/service | 2.8%（配额/安全纯逻辑全路径） | 单租户默认配额 / 组织配额覆盖与组织维度汇总 / CPU·内存限额 / 未配置回退 / computeScore / capDropAll |

**全栈回归（tools/phase13-regression.ps1，14/14 PASS）**
| # | 场景 | 结果 |
|---|---|---|
| R1 | 健康检查 | ✅ |
| R2 | admin 登录 | ✅ |
| R3 | 注册并登录新用户 | ✅ |
| R4 | 创建项目 | ✅ |
| R5 | ECS 创建 → running | ✅ |
| R6 | ECS stats + logs | ✅ |
| R7 | Web 终端 WebSocket（resize + `id` 回显） | ✅ |
| R8 | 停止 → stopped / 启动 → running | ✅ |
| R9 | 应用蓝绿部署 → success | ✅ |
| R10 | Pipeline 创建 → 运行 → success | ✅ |
| R11 | 监控总览（ecs=6 apps=3） | ✅ |
| R12 | 安全扫描（基线+镜像双报告） | ✅ |
| R13 | 计费总览 | ✅ |
| R14 | 资源清理（app/ecs/pipeline） | ✅ |

## 13. 浏览器测试
回归套件覆盖全部后端闭环；前端页面由 HMR 编译零错误 + 各阶段手工验证保障。

## 14. 常见问题 / 已知限制
- service/pipeline 包整体覆盖率低属预期：大部分代码依赖 MySQL/Redis/Docker 实环境，由全栈回归与各阶段验收套件覆盖；纯逻辑路径已 100% 单测。
- sqlite 与 MySQL 语义差异（空串撞唯一索引）已按 GORM 零值省略行为在测试中显式处理。
- 前端未引入 vitest：以端到端验收 + 生产构建零错误为准（引入成本高，Phase 14 视需要补充）。

## 15. 下一阶段
Phase 14 最终优化与验收：用户验收全闭环（注册→建项目→建 ECS→终端→部署→Pipeline→监控→回滚）一遍到底 + 界面打磨 + README/文档终稿。
