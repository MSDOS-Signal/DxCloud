# Phase 15 报告：控制台体验增强（通知中心 2.0 / 创建页重构 / 空间说明 / 报错中文化）

> 状态：已完成并实测通过

## 1. 本阶段目标
通知中心全模块闭环（带跳转链接）、创建 ECS 页面重构（全宽分步 + 镜像中心联动 + 规格预设）、删除确认框错误处理修复、空间（默认/组织）语义说明、配额与端口报错中文化。

## 2. 架构变化
- 通知系统升级：
  - `notifications` 表新增 `link` 字段（迁移 000014），通知可点击跳转对应页面。
  - `internal/service/notify.go` 新增统一 `notify()` 助手（失败不影响主流程）。
  - 通知埋点补齐：ECS 创建/启停/重启/删除成功与失败、镜像拉取成功/失败、应用部署成功/失败/删除、Pipeline 运行结果（中文状态 + 耗时 + 跳转运行详情）。
- 删除实例：Docker 容器移除失败不再静默（记 warn 日志 + 实例事件 + 通知），避免容器残留无人知晓。
- 配额报错中文化并携带具体数值（实例数/CPU/内存上限、剩余量），仍以 `%w` 包装 `ErrQuotaExceed` 保持 `errors.Is` 识别；端口冲突报错区分「运行中容器占用」与「已被其他实例分配」。
- 前端通知中心 2.0（顶栏）：类型图标（ECS/镜像/部署/流水线/安全分色）、相对时间、未读高亮、点击标记已读并跳转 `link`、全部已读按钮、空态引导文案；30s 轮询保留。
- 顶栏空间选择器加悬停说明：默认空间（单租户、全局可见）与组织空间（资源隔离、独立配额与虚拟余额）的用途与切换行为。
- 创建 ECS 页面重构：全宽双栏（左分步表单 + 右粘性配置清单）；6 个分区（基础配置/镜像/规格/网络端口/存储/高级）；镜像区直接展示镜像中心 ready 镜像（含大小）可选，未拉取镜像给出警告与去镜像中心引导；规格预设卡片（轻量/标准/均衡/进阶）+ 自定义微调；右侧实时配置清单 + 虚拟计费预估（CPU ¥0.1/核时 + 内存 ¥0.05/GB时 + 磁盘 ¥0.01/GB时，与后端计费口径一致）。
- 删除确认框（ECS 页 + 容器页）：异步删除失败时弹错误消息且对话框不关闭（可重试），修复此前 Promise 错误被吞的问题。

## 3. 新增文件
backend: migrations/000014_notification_link.sql, internal/service/notify.go
frontend: （无新增页面，create.vue 全量重写）

## 4. 修改文件
backend: internal/model/ops.go（Notification.Link）、internal/pipeline/engine.go（finishRun 通知带链接/中文/耗时）、internal/service/ecs_service.go（错误中文化 + 删除容器错误处理 + 通知埋点）、internal/service/infra_service.go（doPull 带用户上下文 + 通知）、internal/service/app_service.go（部署/删除通知）、internal/service/org_service.go（配额报错中文化）
frontend: pages/ecs/create.vue（重构）、pages/ecs/index.vue、pages/containers/index.vue（删除错误处理）、layouts/default.vue（通知中心 2.0 + 空间说明）、assets/css/main.css（通知面板样式，popover raw 模式 teleport 到 body 需全局样式）

## 5. 数据库变化
000014_notification_link.sql：`ALTER TABLE notifications ADD COLUMN link VARCHAR(255) NULL AFTER content;`

## 6. API 变化
无新端点；GET /notifications 返回体新增 `link` 字段（向后兼容，老通知 link 为空时前端不显示跳转）。

## 7. 前端页面变化
- /ecs/create：全宽分步双栏布局、镜像中心联动、规格预设、配置清单与费用预估。
- 顶栏通知：分色图标 + 跳转 + 全部已读 + 空态。
- /ecs、/containers：删除失败错误可见。

## 8. Docker 变化
无（后端需重建镜像使迁移与代码生效）。

## 9-10. 启动命令
`docker compose up -d --build backend`（迁移随启动自动执行）。

## 11-12. 实测验收
| # | 场景 | 结果 |
|---|---|---|
| 1 | migrations 000014 执行 | ✅ notifications.link 字段存在 |
| 2 | 创建实例成功 → 顶栏通知 → 点击跳转 /ecs/:id | ✅ |
| 3 | 删除实例 → 通知「实例已删除」；容器移除失败时事件+通知 | ✅ |
| 4 | 镜像中心拉取（中国大陆加速源）→ 成功/失败通知 + 跳转 /images | ✅ |
| 5 | 应用部署成功/失败 → 通知 + 跳转 /apps/:id | ✅ |
| 6 | Pipeline 运行 → 通知（中文状态 + 耗时）→ 跳转 /pipeline-runs/:id | ✅ |
| 7 | 配额超限创建 → 中文报错（含数值） | ✅ |
| 8 | 创建页选镜像中心 ready 镜像 → 提示「已就绪」；选未拉取镜像 → 警告引导 | ✅ |
| 9 | 删除确认框失败场景（端口/网络错误注入）→ 错误提示 + 对话框保留 | ✅ |
| 10 | 前端 `npm run build`、后端 `go build ./...` + service/pipeline 单测 | ✅ 全绿 |

## 13. 浏览器测试
/ecs/create 全宽分步布局正常（1440px 双栏、<1280px 单栏）、通知面板浅色/深色两种模式均正常。

## 14. 常见问题 / 已知限制
- 通知中心单次拉取最近 10 条（顶栏下拉），更多历史待「通知中心独立页面」承接（后续）。
- 通知跳转目标目前覆盖 ECS/镜像/应用/流水线四类；安全告警通知随安全中心事件接入后补充。
- 磁盘仍为逻辑配额（Phase 11 既有限制）。

## 15. 下一阶段
可选增强：独立通知中心页面（全量历史 + 按类型筛选）、通知偏好设置（站内/邮件渠道预留）。
