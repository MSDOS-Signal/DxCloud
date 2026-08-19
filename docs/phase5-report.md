# Phase 5 报告：镜像中心 / 网络 / 存储 / Registry

> 状态：已完成并实测通过

## 1. 本阶段目标
镜像中心（拉取/删除/打标签）、云网络（子网+静态 IP+连接管理）、云磁盘（挂载/卸载/数据保留）、私有 Registry（目录/tags/拉取/删 tag）。

## 2. 架构变化
- Provider 扩展为聚合 `Provider` 接口：ImageProvider（List/Remove/Tag/Inspect/Pull）、NetworkProvider（Create/Inspect/List/Connect/Disconnect）、StorageProvider（Create/Remove/Inspect/List）；ECS 创建支持 networkingConfig 静态 IP。
- 镜像拉取为**异步任务**（DB 状态机 pulling→ready/failed，前端轮询），避免大镜像阻塞 HTTP。
- 云磁盘挂载 = 容器重建（Docker 限制），数据卷保留；挂载中删除磁盘被拒（使用中检查）。
- Registry 双地址：REST 内部地址（registry:5000）+ 引擎地址（REGISTRY_ENGINE_URL，daemon 在宿主机 VM 无法解析 compose 服务名）；daemon.json 增 insecure-registries 白名单。
- 修复：DockerNetID 列名映射、IP 从自定义网络端点提取、registry URL scheme 归一化、删 tag 改 POST（Gin :param 无法匹配含斜杠仓库名）。

## 3. 新增文件
backend: migrations/000004_infra.sql, internal/model/infra.go, internal/repository/infra_repo.go, internal/service/infra_service.go, internal/docker/infra_provider.go, internal/handler/infra.go
frontend: pages/{images,networks,volumes,registries}/index.vue

## 4. 修改文件
backend: internal/docker/{provider,docker_provider}.go, internal/service/{ecs_service,ecs_req}.go, internal/repository/ecs_repo.go, internal/handler/ecs.go, internal/dto/ecs.go, internal/config/config.go, internal/database/seed.go, internal/api/router.go
frontend: types/index.ts, pages/ecs/create.vue, pages/registries/index.vue
compose/env: docker-compose{.yml,.prod.yml}, .env.example（REGISTRY_ENGINE_URL、prod registry 端口绑定）
Docker Desktop: daemon.json 增 insecure-registries

## 5. 数据库变化
000004_infra.sql：docker_images / docker_networks / docker_volumes / registries / registry_repositories；种子：内置 Registry。

## 6. API 变化
GET/POST /images、/images/:id/{delete,tag}、POST /images/pull（异步）
GET/POST /networks、GET/DELETE /networks/:id、POST /networks/:id/{connect,disconnect}
GET/POST /volumes、DELETE /volumes/:id、POST /ecs/:id/volumes/{attach,detach}
GET /registries、GET /registries/:id/repositories、POST /registries/:id/repositories/{pull,delete-tag}

## 7. 前端页面变化
镜像中心（异步拉取+轮询+打标签/删除）、网络（创建/详情连接关系/连接容器静态 IP/删除保护）、云磁盘（创建/删除）、Registry（仓库目录/tags/拉取/删 tag）；ECS 创建页新增网络+静态 IP+磁盘挂载行。

## 8. Docker 变化
daemon.json 增 insecure-registries（host.docker.internal:15000 等）；prod compose registry 绑定 127.0.0.1:15000。

## 9. 完整代码
见仓库。

## 10. 启动命令
docker compose up -d backend（已完成）；Docker Desktop 需重启一次以加载 insecure-registries。

## 11-12. 实测验收（19 项全过）
| # | 场景 | 结果 |
|---|---|---|
| 1-3 | 异步拉取 hello-world / 列表 / 打标签 | ✅ ready/15KB/registry:5000/default/hello:v1 |
| 4 | 创建网络 net-a（10.30.0.0/24） | ✅ |
| 5 | ECS 静态 IP 10.30.0.10 | ✅ docker 实际 IP=10.30.0.10 |
| 6 | 第二容器 DHCP 加入 | ✅ 10.30.0.2，详情见连接关系 |
| 7 | 有容器时删网络 | ✅ 400 resource in use |
| 8 | 断开+清理 | ✅ |
| 9-11 | 磁盘创建/挂载/写入 | ✅ hello-volume 写入成功 |
| 12 | 挂载第二磁盘（重建） | ✅ mounts=2，**数据保留** |
| 13-14 | 卸载 / 挂载中删除拒绝 | ✅ |
| 15 | 清理 | ✅ |
| 16-19 | 私有 Registry：push→目录[v1]→API 拉取到引擎 ready→删 tag | ✅ 全闭环 |

## 13. 浏览器测试
/images /networks /volumes /registries /ecs/create 全部 200 无编译错误。

## 14. 常见问题 / 已知限制
- 引擎视角 registry 地址：Docker Desktop 用 host.docker.internal:15000（daemon.json 白名单），Linux 生产用 127.0.0.1:15000。
- 磁盘容量为软配额；挂载变更重建容器（数据保留）。
- docker push 到私有仓可由宿主 CLI/kaniko 完成（Phase 8 接 Pipeline）。

## 15. 下一阶段
Phase 6：应用 / 项目 / 环境 / 域名 / 部署（蓝绿 + 健康检查 + 回滚 + Traefik 域名路由）。
