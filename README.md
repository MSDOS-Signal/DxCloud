# 多晓云 DxCloud — 做懂你心的云枢

**基于 Docker 的一体化云平台控制台** · IaaS + PaaS + SaaS + CI/CD + AI 助手

一个模拟公有云 ECS 控制台体验的轻量级云平台，所有资源均为**真实的 Docker 容器 / 网络 / 存储**，非模拟数据。

---

## 目录

- [1. 项目简介](#1-项目简介)
- [2. 核心特性](#2-核心特性)
- [3. 系统架构](#3-系统架构)
- [4. 技术栈](#4-技术栈)
- [5. 功能模块总览](#5-功能模块总览)
- [6. 界面展示](#6-界面展示)
- [7. 快速开始](#7-快速开始)
- [8. 环境变量配置](#8-环境变量配置)
- [9. 目录结构](#9-目录结构)
- [10. 代码规模统计](#10-代码规模统计)
- [11. 角色权限说明](#11-角色权限说明)
- [12. AI 智能助手「多晓」](#12-ai-智能助手多晓)
- [13. 常见问题](#13-常见问题)

---

## 1. 项目简介

**多晓云 DxCloud** 是一个基于 Docker 的轻量级一体化云平台，模拟公有云 ECS 控制台的完整使用体验。名字取自 **Dx = DuoXiao（多晓）**，寓意「通晓云上万事」。

平台整合了云服务全栈能力：

| 层级 | 能力 |
|------|------|
| **IaaS** | ECS 云主机（真实 Docker 容器）、容器实例、镜像中心、自定义网络、云磁盘 |
| **PaaS** | 应用托管、项目工作区、域名管理、镜像仓库 |
| **SaaS** | 开箱即用的应用模板一键部署 |
| **DevOps** | 自研轻量级 CI/CD 流水线引擎（Git 拉取 → 构建 → 测试 → 镜像 → 部署 → 健康检查 → 发布） |
| **运营** | 监控中心、操作审计、日志中心、计费中心、组织管理、IAM 权限体系 |
| **AI** | 全局悬浮 AI 助手「多晓」，结合平台知识库提供全模块答疑 |

**设计理念**：
- 🎨 腾讯云风格 UI：蓝色主色调、深浅色双主题、粒子动效、ECharts 数据可视化
- 🔐 企业级安全：JWT 认证、RBAC 三级权限（用户/角色/权限点）、多租户数据隔离、密码加密存储
- 🇨🇳 国内环境友好：Docker 镜像源可配置国内加速、操作提示全中文、兼容境内 API

---

## 2. 核心特性

### 真实资源调度
- ECS 实例 = 真实 Docker 容器（支持启动/停止/重启/删除/重建）
- 网络子网 = 真实 Docker bridge 网络（自定义子网网段）
- 云磁盘 = 真实 Docker volume（支持挂载/卸载）
- 镜像中心 = 真实拉取 Docker Hub / 镜像仓库

### Web VNC 终端
浏览器内直接进入容器终端，支持交互式命令操作，无需 SSH 客户端。

### 自研 CI/CD 引擎
```
Git Clone → Build → Test → Docker Build → Push Registry → Deploy → Health Check → Release
```
- YAML 声明式流水线配置
- 实时日志流（WebSocket 推送）
- 流水线执行历史与状态机管理

### AI 智能助手
- 全局悬浮球，可鼠标拖动，位置持久化
- 流式打字机输出（SSE）
- 内置平台全模块知识库，结合当前页面语境回答
- 深浅色主题自适应

---

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     浏览器（用户端）                          │
│         Nuxt 3 SPA + Naive UI + ECharts + AI 助手            │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS / WSS
┌────────────────────────▼────────────────────────────────────┐
│                    cloud-web（前端容器）                      │
│              Nuxt 3 SSR + Nginx 反向代理 :80                 │
└────────────────────────┬────────────────────────────────────┘
                         │ /api 反代
┌────────────────────────▼────────────────────────────────────┐
│                    cloud-api（后端容器）                      │
│      Go + Gin + GORM + JWT + RBAC + 限流 + 审计 + SSE        │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌───────────────┐   │
│  │ ECS 模块 │ │ PaaS 模块│ │ CI/CD 引擎 │ │  AI 代理(智谱) │   │
│  └────┬─────┘ └────┬─────┘ └─────┬─────┘ └───────┬───────┘   │
└───────┼────────────┼─────────────┼───────────────┼───────────┘
        │            │             │               │
   ┌────▼────────────▼─────────────▼───────┐  ┌───▼────────┐
   │        Docker Engine（宿主机）          │  │ 智谱 GLM    │
   │   容器 / 网络 / 卷 / 镜像（真实资源）     │  │ glm-4-flash│
   └───────────────────────────────────────┘  └────────────┘
        │                    │
   ┌────▼─────┐         ┌────▼─────┐
   │ dx-mysql │         │ dx-redis │
   │ 业务数据  │         │ 会话/缓存 │
   └──────────┘         └──────────┘
```

---

## 4. 技术栈

### 后端（Go）

| 技术 | 用途 |
|------|------|
| Go 1.22 | 开发语言 |
| Gin | HTTP 框架 |
| GORM | ORM（MySQL） |
| go-redis | Redis 客户端（会话管理/缓存/限流） |
| golang-jwt | JWT 签发与校验 |
| docker/docker client | Docker API 调用 |
| gorilla/websocket | 终端 WebSocket、日志实时推送 |
| bcrypt | 密码加密存储 |
| zap | 结构化日志 |

### 前端（Nuxt 3）

| 技术 | 用途 |
|------|------|
| Nuxt 3 / Vue 3 | 应用框架（SSR + SPA 混合） |
| TypeScript | 类型安全 |
| Naive UI | 组件库 |
| Pinia | 状态管理 |
| ECharts | 数据可视化（环形图/折线图/仪表盘） |
| Tailwind CSS | 原子化样式 |
| vue-i18n | 国际化预留 |

### 基础设施

| 技术 | 用途 |
|------|------|
| Docker Compose | 多容器编排 |
| MySQL 8 | 业务数据库（15 张表） |
| Redis 7 | 会话 / 缓存 / 限流计数 |
| Nginx | 前端静态资源 + API 反代 |
| 智谱 GLM (glm-4-flash) | AI 助手大模型（免费） |

---

## 5. 功能模块总览

| 模块 | 路由 | 功能 |
|------|------|------|
| 总览仪表盘 | `/dashboard` | 资源统计卡、CPU/内存实时曲线、最近操作 |
| ECS 云主机 | `/ecs` | 创建/启停/删除实例、端口映射、VNC 终端、监控 |
| 容器实例 | `/containers` | 底层容器列表、状态管理、日志查看 |
| 镜像中心 | `/images` | 镜像列表、拉取、清理、国内源加速 |
| 网络 | `/networks` | 自定义 Docker 网络创建、子网分配 |
| 存储 | `/volumes` | 云磁盘创建、挂载/卸载、容量统计 |
| 应用中心 | `/apps` | PaaS 应用部署、模板市场、应用详情 |
| 项目 | `/projects` | 项目工作区、成员协作 |
| 域名 | `/domains` | 域名绑定与解析管理 |
| 镜像仓库 | `/registries` | 私有镜像仓库配置 |
| 流水线 | `/pipelines` | CI/CD 流水线定义、执行历史 |
| 部署记录 | `/deployments` | 应用部署历史与回滚 |
| 监控中心 | `/monitor` | 全局 CPU/内存/网络实时监控（30s 自动刷新） |
| 日志中心 | `/logs` | 操作日志、系统日志检索 |
| 计费中心 | `/billing` | 资源计费统计、消费趋势图表 |
| 安全中心 | `/security` | 安全概览、审计日志、敏感操作记录 |
| IAM 用户 | `/iam/users` | 用户管理、角色分配 |
| IAM 角色 | `/iam/roles` | 角色管理、权限点绑定 |
| IAM 权限 | `/iam/permissions` | 权限点注册与查询 |
| 组织管理 | `/orgs` | 多租户组织、成员邀请 |
| 个人信息 | `/profile` | 头像上传（本地/预设/URL）、昵称、改密、会话管理 |
| 设置 | `/settings` | 深浅色主题切换、系统信息 |
| AI 助手 | 全局悬浮 | 全模块答疑、流式对话、可拖动 |

---

## 6. 界面展示

### 6.1 登录与注册

![登录页](run_img/01_login.png)

![注册页](run_img/02_register.png)

### 6.2 总览仪表盘

![Dashboard](run_img/03_dashboard.png)

### 6.3 ECS 云主机

![ECS 列表](run_img/04_ecs.png)

![ECS 实例详情](run_img/24_ecs_detail.png)

### 6.4 容器与镜像

![容器实例](run_img/05_containers.png)

![镜像中心](run_img/06_images.png)

### 6.5 网络与存储

![网络管理](run_img/07_networks.png)

![存储管理](run_img/08_volumes.png)

### 6.6 PaaS 应用

![应用中心](run_img/09_apps.png)

![项目管理](run_img/10_projects.png)

![域名管理](run_img/11_domains.png)

### 6.7 CI/CD 流水线

![流水线](run_img/12_pipelines.png)

![部署记录](run_img/13_deployments.png)

### 6.8 运维监控

![监控中心](run_img/14_monitor.png)

![日志中心](run_img/15_logs.png)

### 6.9 计费与组织

![计费中心](run_img/17_billing.png)

![组织管理](run_img/16_orgs.png)

### 6.10 安全与权限

![安全中心](run_img/18_security.png)

![IAM 用户](run_img/19_iam_users.png)

![IAM 角色](run_img/20_iam_roles.png)

![IAM 权限](run_img/21_iam_permissions.png)

### 6.11 个人信息与设置

![个人信息](run_img/25_profile.png)

![设置](run_img/22_settings.png)

### 6.12 AI 智能助手

浅色模式：

![AI 助手](run_img/23_ai_chat.png)

深色模式：

![深色 AI 助手](run_img/26_dark_ai_chat.png)

### 6.13 深色主题全站适配

![深色个人信息页](run_img/27_dark_profile.png)

---

## 7. 快速开始

### 7.1 环境要求

- Docker 20.10+ 与 Docker Compose v2
- 4GB+ 可用内存
- 现代浏览器（Chrome / Edge / Firefox）

### 7.2 一键部署

```bash
# 1. 克隆项目
git clone <项目地址>
cd dxcloud

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env，填写 MySQL 密码、JWT 密钥、智谱 API Key（可选）

# 3. 启动全部服务
docker compose up -d

# 4. 初始化数据库（首次）
docker exec dx-mysql mysql -uroot -p<密码> < /path/to/init.sql
# 或直接使用自动迁移，首次启动 cloud-api 自动建表

# 5. 访问
# 浏览器打开 http://localhost
```

### 7.3 默认账号

| 账号 | 密码 | 角色 |
|------|------|------|
| `admin` | `Admin@123456` | 超级管理员（superadmin） |

> 首次登录后请在「个人信息」页修改默认密码。

### 7.4 首次操作流程（新用户视角）

1. **注册账号**：登录页点击「注册」，填写用户名/邮箱/密码
2. **登录**：使用注册的账号进入控制台
3. **创建 ECS 实例**：计算 → ECS 云主机 → 创建（选镜像、规格、端口映射）
4. **访问实例**：实例卡片复制访问地址 `http://localhost:<映射端口>`
5. **进入终端**：实例详情 → 终端（Web VNC）
6. **部署应用**：PaaS → 应用中心 → 从模板部署
7. **配置流水线**：DevOps → 流水线 → 新建 → 关联 Git 仓库
8. **查看监控**：运维 → 监控中心

---

## 8. 环境变量配置

`.env` 文件（docker-compose 读取）：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MYSQL_ROOT_PASSWORD` | MySQL root 密码 | 必填 |
| `MYSQL_DATABASE` | 数据库名 | `dxcloud` |
| `REDIS_PASSWORD` | Redis 密码 | 空 |
| `JWT_SECRET` | JWT 签名密钥 | 必填（生产环境） |
| `JWT_EXPIRE_HOURS` | 会话有效期（小时） | `72` |
| `ZHIPU_API_KEY` | 智谱 AI API Key | 空（禁用 AI 助手） |
| `ZHIPU_BASE_URL` | 智谱 API 地址 | `https://open.bigmodel.cn/api/paas/v4` |
| `ZHIPU_MODEL` | 模型名 | `glm-4-flash` |
| `DOCKER_REGISTRY_MIRROR` | Docker 镜像加速源 | 空 |

---

## 9. 目录结构

```
dxcloud/
├── backend/                    # Go 后端
│   ├── cmd/server/             # 程序入口 main.go
│   ├── internal/
│   │   ├── api/                # 路由注册
│   │   ├── config/             # 配置加载
│   │   ├── database/           # MySQL/Redis 连接
│   │   ├── docker/             # Docker 资源调度封装
│   │   ├── dto/                # 传输对象
│   │   ├── handler/            # HTTP 处理器（15 个）
│   │   ├── iam/                # 认证/权限/会话核心
│   │   ├── middleware/         # JWT/限流/审计/CORS 等
│   │   ├── model/              # 数据模型（11 个）
│   │   ├── pipeline/           # CI/CD 流水线引擎
│   │   ├── repository/         # 数据访问层（11 个）
│   │   ├── runner/             # 容器任务执行器
│   │   ├── scheduler/          # 定时任务
│   │   ├── service/            # 业务逻辑（14 个）
│   │   └── websocket/          # 终端/日志 WS 服务
│   ├── pkg/                    # 工具包（jwt/crypto/redisx...）
│   └── migrations/             # SQL 迁移脚本
├── frontend/                   # Nuxt 3 前端
│   ├── components/             # 组件（图表/AI 助手/品牌等 12 个）
│   ├── composables/            # 组合函数（useEcharts/useCursor）
│   ├── layouts/                # 布局（default/auth）
│   ├── pages/                  # 页面（30 个路由）
│   ├── plugins/                # Naive UI 注册
│   ├── services/               # HTTP 客户端封装
│   ├── stores/                 # Pinia 状态（auth/org/theme）
│   └── types/                  # TS 类型定义
├── run_img/                    # 界面运行截图（27 张）
├── docker-compose.yml          # 容器编排
├── .env.example                # 环境变量模板
└── README.md                   # 本文档
```

---

## 10. 代码规模统计

| 类别 | 文件数 | 代码行数 |
|------|--------|----------|
| 后端 Go（.go） | 92 | 12,897 |
| 数据库脚本（.sql） | 15 | 1,703 |
| 前端页面/组件（.vue） | 48 | 9,846 |
| 前端脚本（.ts） | 16 | 1,193 |
| 全局样式（.css） | 1 | 559 |
| **合计** | **172** | **26,198** |

后端核心构成：

| 模块 | 文件数 | 行数 |
|------|--------|------|
| service（业务逻辑） | 14 | 3,798 |
| handler（HTTP 处理） | 15 | 2,696 |
| repository（数据访问） | 11 | 1,294 |
| docker（容器调度） | 3 | 1,012 |
| pipeline（CI/CD 引擎） | 2 | 746 |
| iam（认证权限） | 3 | 622 |
| model（数据模型） | 11 | 582 |

---

## 11. 角色权限说明

| 角色 | 能力 |
|------|------|
| **superadmin（超级管理员）** | 全部功能：IAM 用户/角色/权限管理、组织管理、全平台资源管理、审计日志 |
| **admin（管理员）** | 平台资源管理（ECS/应用/流水线/监控/计费），不能管理 IAM |
| **user（普通用户）** | 仅管理**自己创建**的资源（多租户数据隔离），不能访问 IAM/组织/审计 |

**数据隔离**：所有业务表带 `user_id` 字段，普通用户查询自动过滤；管理员按权限点控制。

---

## 12. AI 智能助手「多晓」

与平台同名的 AI 助手「多晓」（DuoXiao），右下角悬浮球点击即开：

- **知识库答疑**：内置 13 个功能模块操作指引，如「容器 IP 不能直接访问，需走端口映射」
- **页面语境感知**：在镜像中心页提问时优先返回镜像相关操作
- **流式输出**：SSE 逐字渲染，打字机效果
- **可拖动悬浮球**：Pointer Events 实现，位置存 localStorage
- **Markdown 渲染**：代码块/粗体/列表，XSS 已过滤
- **主题自适应**：深浅色实时切换

后端代理智谱 GLM（`glm-4-flash` 免费模型），API Key 仅存于服务端，前端零暴露；登录保护 + 30 次/分钟限流 + 审计日志。

---

## 13. 常见问题

**Q1：为什么访问不了容器 IP（如 10.8.8.2）？**

容器 IP 是 Docker 内部虚拟网络地址，宿主机没有路由。请使用**端口映射**访问：实例创建时配置端口映射（如 18888→80），浏览器访问 `http://localhost:18888`。

**Q2：镜像拉取慢/失败？**

配置国内镜像加速源：`.env` 中设置 `DOCKER_REGISTRY_MIRROR=https://docker.m.daocloud.io`（或阿里云个人加速地址），重启 cloud-api。

**Q3：前端样式 404 / MIME type 错误？**

开发模式下不要在宿主机运行 `npm run build`（会覆盖容器内 dev 产物）。类型检查用 `npx nuxi typecheck`。

**Q4：忘记 admin 密码？**

```bash
docker exec dx-mysql mysql -uroot -p<密码> dxcloud \
  -e "UPDATE users SET password='<bcrypt哈希>' WHERE username='admin';"
```

**Q5：AI 助手没反应？**

检查 `.env` 中 `ZHIPU_API_KEY` 是否已配置，并 `docker restart cloud-api`。

---

**多晓云 DxCloud** · 做懂你心的云枢
