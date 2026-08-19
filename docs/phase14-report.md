# Phase 14 报告：最终优化与用户验收

> 状态：**项目完成**。用户验收全闭环（注册→建项目→建ECS→终端→部署→Pipeline→监控→回滚）10/10 PASS。

## 1. 本阶段目标
界面收尾、修复监控曲线空数据返回 null 的问题、编写最终验收套件（用户验收标准一遍到底）并实测通过。

## 2. 架构变化
- 修复：/monitor/series 无采样时返回 `[]` 而非 `null`（前端图表友好，避免空引用）。
- `pages/[...slug].vue` 占位页升级为 404 结果页（"页面不存在"）。
- tools/final-acceptance.ps1：最终验收套件（10 项，含蓝绿部署 ×2 + 回滚 + Web 终端 WebSocket + Pipeline + 监控曲线）。

## 3. 新增文件
tools/final-acceptance.ps1

## 4. 修改文件
backend/internal/handler/monitor.go（空数组序列化）、frontend/pages/[...slug].vue（404 页）

## 5. 数据库变化
无。

## 6. API 变化
/monitor/series：空结果 `data: []`（原为 `null`）。

## 7. 前端页面变化
未知路由 404 结果页；全站功能页面于 Phase 1-13 交付完毕。

## 8. Docker 变化
无。

## 9. 完整代码
见仓库。

## 10. 启动命令
```
docker compose up -d                          # 开发栈（一键）
docker compose -f docker-compose.prod.yml up -d --build   # 生产栈（HTTPS，先运行 tools/gencert.go）
.\tools\final-acceptance.ps1                  # 最终验收
```

## 11-12. 实测验收（用户验收标准，10/10 PASS）
| # | 环节 | 实测证据 |
|---|---|---|
| F1 | 注册并登录 | 新用户 final0819032218，token 签发 ✅ |
| F2 | 建项目 | proj=9 ✅ |
| F3 | 建 ECS | busybox 实例 id=23 → running ✅ |
| F4 | Web 终端 | WebSocket resize(40×140) + `id` 回显 ✅ |
| F5 | 部署 v1 | 蓝绿部署 dep=21 → success ✅ |
| F6 | 部署 v2 | 二次蓝绿切换 dep=22 → success ✅ |
| F7 | 回滚 | 回滚到 v1 → dep=23 success（note=rollback to ...v5.21）✅ |
| F8 | Pipeline | 2 步 shell 流水线 run=23 → success ✅ |
| F9 | 监控 | 总览 ecs=6/deploy_today=23 + 曲线 37 个采样点 ✅ |
| F10 | 清理 | app/ecs/pipeline 删除 ✅ |

## 13. 浏览器测试
http://localhost/（dev）与 https://localhost/（prod）全功能可用；前端 HMR 零编译错误；全部页面 200。

## 14. 常见问题 / 已知限制
- 平台以 Docker 引擎为唯一运行时，容器能力映射为云资源（无 K8s 编排，符合轻量 ECS 定位）。
- 镜像漏洞扫描未接入 trivy（网络受限），以策略扫描 + 私有 Registry 管控替代。
- 演示数据保留：admin/Admin@123456、alice/Alice12345、pipeline「ci-cd」、应用「pipe-app」（域名 pipe.localhost）、组织 A/B（含 bob 成员）。

## 15. 项目状态
**14 个 Phase 全部完成并实测通过。** 验收脚本：tools/phase10-acceptance.ps1（19 项）、phase11-acceptance.ps1（12 项）、phase12-acceptance.ps1（15 项）、phase13-regression.ps1（14 项）、final-acceptance.ps1（10 项），全部可重复执行。
