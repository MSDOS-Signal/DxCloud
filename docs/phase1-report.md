# Phase 1 报告：项目骨架

> 状态：已完成（本报告由 Phase 1 交付时生成）

## 1. 本阶段目标

搭起可一键启动的全栈骨架：Go 后端（Gin/GORM/zap/go-redis）+ Nuxt 3 SPA 前端（Naive UI/Tailwind/Pinia）+ MySQL 8 + Redis 7 + Registry + Traefik，`docker compose up -d` 即可运行，前后端 → 数据层全链路打通。

## 2. 架构变化

- 按 Phase 0 架构落地 Modular Monolith 包结构（internal/api|middleware|handler|config|database + pkg/logger|resp|errcode|redisx）。
- 新增 `internal/database`（连接 + 迁移执行器）、`migrations/`（embed + 编号 SQL，只增不改）。
- Redis 方案确认仅使用 3.0 兼容原语（LIST/ZSET/SET NX EX/HASH/PUBSUB），生产统一 redis:7-alpine。
- 开发模式确认：全栈容器化（用户选择），宿主机端口映射 13306/16379/15000 避开本机 MySQL/Redis。
- 环境适配：Docker Desktop 无法直连 Docker Hub，已配置 registry-mirrors（daocloud/1ms/xuanyuan/rat.dev/1panel）。

## 3. 新增文件

backend/: go.mod, cmd/server/main.go, internal/config/config.go, internal/api/router.go,
internal/middleware/{requestid,accesslog,recovery,cors}.go, internal/handler/health.go,
internal/database/{db,migrate}.go, pkg/{logger,resp,errcode,redisx}, migrations/{embed.go,000001_init.sql},
Dockerfile, Dockerfile.dev, .dockerignore

frontend/: package.json, nuxt.config.ts, tsconfig.json, tailwind.config.ts, app.vue,
plugins/naive.ts, stores/auth.ts, middleware/auth.global.ts, services/http.ts, types/index.ts,
layouts/{auth,default}.vue, pages/{index,login,dashboard/index,[...slug]}.vue,
Dockerfile, Dockerfile.dev, .dockerignore

根目录: docker-compose.yml, docker-compose.dev.yml, docker-compose.prod.yml,
deploy/traefik/{traefik.dev,traefik.prod}.yml, .env.example, .gitignore, README.md

## 4. 修改文件

无（本阶段全部为新增）。

## 5. 数据库变化

新增迁移 `backend/migrations/000001_init.sql`，由 `internal/database.Migrate` 按序执行并记录于 `schema_migrations`：
users / roles / permissions / user_roles / role_permissions / login_logs /
organizations / organization_members / projects / project_environments / system_settings

约定：InnoDB + utf8mb4；统一 id/created_at/updated_at/deleted_at；租户表自始含 org_id/project_id/owner_id（按 Phase 0 决策不后补）。

## 6. API 变化

- GET /healthz —— 运维探针（compose/traefik 健康检查用）
- GET /api/v1/health —— 业务探针（统一响应 + MySQL/Redis 状态）
- 其余 /api/* 统一返回 `{code:40401, message:"endpoint not implemented yet"}`（后续 Phase 替换为真实 handler）

## 7. 前端页面变化

- /login：登录页（Phase 1 占位登录，Phase 2 接真实认证）
- /dashboard：仪表盘（占位统计卡 + 全链路连通性探测卡）
- 控制台布局：左侧云厂商风格导航（ECS/镜像/网络/存储/应用/CI-CD/监控/日志/IAM/计费/设置）
- [...slug]：未交付模块占位页（随各 Phase 被真实页面替换）

## 8. Docker 变化

- 三套 Compose：docker-compose.yml（开发全栈，源码挂载热运行）/ docker-compose.dev.yml（仅基础设施，本机直跑）/ docker-compose.prod.yml（生产构建产物，强制 .env 密码）
- Traefik v3 边缘路由：`/api`、`/healthz`、`/ws` → cloud-api；其余 → cloud-web
- 后端容器独占 docker.sock；Traefik 只读挂载
- Docker Desktop daemon.json 增加 registry-mirrors（国内镜像加速）

## 9. 完整代码

见本仓库（backend/、frontend/、docker-compose*.yml 等，本报告不重复粘贴）。

## 10. 启动命令

```bash
cp .env.example .env        # 可选
docker compose up -d --build
docker compose ps
```

## 11. 测试方法

- go vet / go build（本机 Go 1.23 已通过）
- docker compose config -q × 3 套配置（已通过）
- 健康检查轮询 http://localhost/healthz

## 12. curl 测试

```bash
curl http://localhost/healthz
curl http://localhost/api/v1/health
curl -X POST http://localhost/api/v1/auth/login -H 'Content-Type: application/json' -d '{}'
```

## 13. 浏览器测试

1. http://localhost → 登录页
2. 任意账号密码登录 → Dashboard
3. Dashboard「系统连通性」卡片显示 code=0 / DB=up / Redis=up
4. 左侧导航点击任意模块 → 占位页
5. http://localhost:8080 → Traefik Dashboard

## 14. 常见问题

### 实测问题与修复记录（Windows + Docker Desktop 29.6 环境）

| 问题 | 原因 | 修复 |
|---|---|---|
| Docker Hub 直连拉取失败 | 国内网络无法访问 registry-1.docker.io | `~/.docker/daemon.json` 配置 registry-mirrors（hub.rat.dev / daocloud / 1ms.run / xuanyuan），`docker desktop restart` 生效；开代理时直连最快 |
| 镜像源大层失速 | 部分公共镜像源限速/断流 | 多源并行分摊 + 显式拉取打短名标签（`docker pull hub.rat.dev/library/mysql:8.0` → `docker tag` → `mysql:8.0`） |
| 前端构建 npm 崩溃（`edgesOut`） | npm arborist 在无 lockfile 大依赖树下崩溃 | 本机生成 `package-lock.json`，容器内改 `npm ci`（确定性安装） |
| Traefik 3.1 无法读取 docker.sock（`Error response from daemon:` 空错误） | Traefik 3.1 内置 Docker SDK 与 Docker 29.6 (API 1.55) 不兼容 | 升级 `traefik:v3`（新版 SDK） |
| Traefik 路由后端超时 | 后端容器同时在 infra/edge 双网络，Traefik 默认取第一个网络（infra）不可达 | traefik.yml 显式 `network: dxcloud_edge` |

- Docker Hub 拉取失败 → 已配置镜像加速；若某镜像源失效可编辑 `~/.docker/daemon.json` 后 `docker desktop restart`
- 端口冲突 → 宿主映射 13306/16379/15000/8080，.env 可改
- 后端起不来 → `docker compose logs backend`；确认 mysql/redis healthy
- 前端热更新失效 → CHOKIDAR_USEPOLLING=true 已内置

## 15. 下一阶段

Phase 2：认证 + RBAC —— 注册/登录/JWT/Refresh Token/退出/改密/会话管理、6 个默认角色与权限码种子数据、中间件鉴权链、登录日志与审计日志、前端登录接真实 API。
