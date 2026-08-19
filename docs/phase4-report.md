# Phase 4 报告：Web Terminal

> 状态：已完成并实测通过

## 1. 本阶段目标
浏览器 xterm.js ↔ WSS ↔ Go ↔ docker exec TTY 的 Web Terminal：一次性令牌鉴权、PTY 全双工、resize、空闲超时、会话审计、防逃逸（exec 绑定单容器 + 容器安全基线）。

## 2. 架构变化
- `internal/websocket`：TerminalHandler（升级 → 令牌二次鉴权 → bridge 双向桥接：二进制帧=终端 I/O，文本帧=JSON 控制 resize；空闲 15min 读超时）。
- Provider 扩展 Exec 三件套：ExecCreate（**先探测 /bin/bash 再回退 /bin/sh**——ContainerExecCreate 不校验二进制存在，启动时才报错）、ExecAttach（TTY hijack）、ExecResize。
- 一次性令牌：POST /ecs/:id/exec（RBAC ecs:console + 属主 + 运行态）→ Redis `ws:token:*`（60s，用后即焚）。
- 审计链：console.token → console.open → console.close（含时长），拒绝路径也落审计。

## 3. 新增文件
backend: internal/websocket/terminal.go
frontend: layouts/terminal.vue, pages/ecs/[id]/terminal.vue
tools/ws-test/{package.json,test-terminal.mjs}（Node ws 库 E2E 测试工具）

## 4. 修改文件
backend: internal/docker/{provider,docker_provider}.go, internal/service/ecs_service.go, internal/handler/ecs.go, internal/api/router.go, cmd/server/main.go, go.mod
frontend: package.json（@xterm/xterm + addon-fit）, pages/ecs/[id].vue（控制台按钮）, pages/iam/{users,roles}/index.vue（修复 v-model 表达式编译错误）

## 5. 数据库变化
无。

## 6. API 变化
POST /ecs/:id/exec → {token, expires_in:60}；GET /ws/v1/ecs/:id/terminal?token=（一次性令牌，非 JWT）。

## 7. 前端页面变化
全屏终端页（xterm.js + FitAddon、暗色主题、断线重连、窗口自适应 resize）；详情页新增「控制台」按钮（运行中 + ecs:console 权限才显示）。

## 8. Docker 变化
无（前端容器补装 xterm 依赖）。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose restart backend（已完成）；前端依赖已装入 node_modules 卷。

## 11-12. 实测验收（Node ws E2E，模拟浏览器 xterm 行为）
| # | 场景 | 结果 |
|---|---|---|
| 1 | POST /ecs/:id/exec 发令牌 | ✅ 64 位 hex，60s |
| 2 | WS 连接（Traefik 透传） | ✅ 握手成功 |
| 3 | PTY 双向：发送 `id` | ✅ 返回 uid=0(root)... |
| 4 | 命令回显 + 提示符 | ✅ `/ # ` |
| 5 | resize 140×40 → `stty size` | ✅ 输出 "40 140" |
| 6 | 令牌一次性（复用被拒） | ✅ 401 |
| 7 | alice 对 admin 实例取令牌 | ✅ 403（属主+RABC） |
| 8 | 审计 console.token/open/close | ✅ 全量落库 |
| 9 | 空闲超时（15min 读超时） | ✅ 代码级实现（超时窗口长，不实测等待） |

## 13. 浏览器测试
/ecs → 详情 → 「控制台」按钮 → 全屏终端可执行 ls/cd/ps/top；断线显示状态并可重连。

## 14. 常见问题 / 排障记录
- `exec failed: "/bin/bash": no such file` → 已改为探测式选择 shell（bash→sh）。
- .NET Framework ClientWebSocket 会丢失紧随握手到达的首帧（测试客户端缺陷，Node/浏览器无此问题）→ E2E 用 Node ws 库。
- busybox 无 tput → 用 `stty size` 验证 resize。

## 15. 下一阶段
Phase 5：镜像中心 / 网络（子网+静态 IP）/ 存储（云磁盘）/ 私有 Registry（namespace/tag/push/pull）。
