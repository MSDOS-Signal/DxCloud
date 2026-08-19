# DxCloud（CloudECS）新手入门指南

> 本文面向第一次接触本项目、从 GitHub 下载源码的新手用户。全程只需要 Docker，无需安装 Go / Node.js / MySQL 等任何开发环境。跟着步骤做，10 分钟左右即可在本机跑起完整平台。

---

## 目录

1. [前置条件：安装 Docker Desktop](#1-前置条件安装-docker-desktop)
2. [中国大陆用户专项：加速镜像拉取](#2-中国大陆用户专项加速镜像拉取)
3. [一键启动平台](#3-一键启动平台)
4. [常见问题](#4-常见问题)
5. [下一步](#5-下一步)

---

## 1. 前置条件：安装 Docker Desktop

DxCloud 通过 Docker Compose 一键拉起全部服务（MySQL、Redis、私有 Registry、Traefik 网关、Go 后端、Nuxt 前端），因此**唯一的前置依赖就是 Docker**。

### 1.1 先检查电脑上是否已有 Docker

打开终端（Windows：PowerShell 或 CMD；macOS：终端），执行：

```bash
docker --version
```

- 正常输出类似 `Docker version 27.x.x, build xxxxxxx` → 已安装，可直接跳到 [第 3 节](#3-一键启动平台)。
- 如果报错：
  - Windows：`docker : 无法将 "docker" 项识别为 cmdlet、函数、脚本文件或可运行程序的名称`
  - macOS：`command not found: docker`

  说明尚未安装 Docker，请继续往下看。

### 1.2 下载 Docker Desktop

官方下载地址（Windows / macOS 均在此页面选择对应版本）：

> **https://www.docker.com/products/docker-desktop/**

1. 打开上面的链接，点击页面上的 **Download Docker Desktop** 按钮；
2. 网站会自动识别你的操作系统，选择对应版本：
   - **Windows**：下载 `Docker Desktop Installer.exe`（要求 Windows 10/11 64 位，并开启 WSL 2 或 Hyper-V）；
   - **macOS**：下载 `Docker.dmg`（Apple Silicon 芯片选 Apple Chip，Intel 芯片选 Intel Chip）。

### 1.3 安装

**Windows：**

1. 双击下载好的 `Docker Desktop Installer.exe`；
2. 勾选 **Use WSL 2 instead of Hyper-V**（推荐，默认即此项），点击 **OK**；
3. 等待安装完成，点击 **Close and restart** 重启电脑；
4. 重启后 Docker Desktop 会自动启动，首次运行可能提示同意服务条款，接受即可。

**macOS：**

1. 双击 `Docker.dmg`；
2. 将鲸鱼图标 **Docker** 拖入 **Applications** 文件夹；
3. 在「启动台」或「应用程序」中打开 Docker，按提示授予权限。

### 1.4 启动并确认 Docker 就绪

1. 启动 Docker Desktop 后，系统托盘（Windows 任务栏右下角 / macOS 顶部菜单栏）会出现一个**鲸鱼图标**；
2. 等待图标停止动画、状态显示 **Engine running**（引擎运行中），表示 Docker 已就绪；
3. 回到终端验证：

```bash
docker version
```

能同时看到 `Client` 和 `Server`（Daemon）两段版本信息即安装成功。再跑一个官方小镜像做最终验证（可选）：

```bash
docker run --rm hello-world
```

看到 `Hello from Docker!` 字样说明一切正常。

---

## 2. 中国大陆用户专项：加速镜像拉取

DxCloud 启动时需要从 Docker Hub 拉取 `mysql:8.0`、`redis:7-alpine`、`registry:2.8`、`traefik:v3`、`alpine:3.20` 等基础镜像。由于网络原因，中国大陆直连 Docker Hub 经常出现**拉取超时 / 速度极慢 / `toomanyrequests` 限流**。以下两种方案任选其一（推荐都配上）。

### 方案 a) 给 Docker 配置镜像加速器（registry-mirrors）

1. 打开 Docker Desktop → 右上角齿轮 **Settings（设置）** → 左侧 **Docker Engine**；
2. 在打开的 JSON 编辑框中，加入 `registry-mirrors` 字段（如果已有其他配置项，注意用逗号拼接，不要整体覆盖）：

```json
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1ms.run",
    "https://hub.rat.dev"
  ]
}
```

3. 点击 **Apply & Restart**，等待 Docker 重启完成；
4. 验证加速器已生效：

```bash
docker info
```

输出末尾出现 `Registry Mirrors:` 及你配置的地址即生效。

> **注意**：公共镜像加速源由第三方运营，**可能随时失效或限速**。上面的示例源是截至 2026 年仍常用的地址，仅供参考；若发现失效，可在本平台控制台 **「设置」→「区域与镜像源」** 中切换其他源，或自行搜索最新可用的加速源后替换上面的 URL。

### 方案 b) 使用平台内置的区域加速能力（推荐）

平台控制台内置了镜像拉取加速能力：

1. 登录控制台后，进入左侧菜单 **「系统 → 设置」** 页；
2. 在 **「区域与镜像源」** 中将区域选择为 **「中国大陆」**；
3. 之后平台内拉取官方镜像（如镜像中心拉取、Pipeline 构建拉取基础镜像）会**自动走加速前缀**，无需再手工改镜像地址；
4. 如果当前加速源传输失败，平台会自动依次尝试其他内置国内候选源，并在拉取日志中显示切换记录。

> 该能力对方案 a) 是互补关系：方案 a) 作用于本机 Docker 引擎（影响 `docker compose up` 时的基础镜像拉取），方案 b) 作用于平台内部的镜像操作。首次部署建议先配好方案 a)，再在控制台里开启方案 b)。

---

## 3. 一键启动平台

### 3.1 获取代码

```bash
# 克隆仓库（请将仓库地址替换为你实际获取项目的 GitHub 地址）
git clone https://github.com/<组织或用户名>/dxcloud.git

# 进入项目目录
cd dxcloud
```

> 没有安装 git 的话，也可以在 GitHub 仓库页面点击 **Code → Download ZIP**，解压后进入解压目录，效果相同。

### 3.2 构建并启动

Windows 用户推荐执行项目根目录的启动脚本：

```powershell
# 中国大陆：自动写入国内基础镜像加速配置
.\scripts\start.ps1 -Region cn

# Windows 也可以直接双击 start-cn.bat；命令行等价：
start-cn.bat

# 非中国大陆：使用 Docker Hub 官方源
.\scripts\start.ps1 -Region global

# Windows 也可以直接双击 start-global.bat；命令行等价：
start-global.bat
```

macOS / Linux 用户可使用：

```bash
chmod +x scripts/start.sh
./scripts/start.sh cn
```

脚本会检查 Docker 是否可用、复制 `.env.example`、按区域写入基础镜像变量，然后执行 `docker compose up -d --build`。如果暂时不想用脚本，也可按原来的方式启动：

```bash
# （可选）复制环境变量模板；不复制也可用内置开发默认值直接启动
cp .env.example .env

# 一键构建并启动全栈（MySQL / Redis / Registry / Traefik / 后端 / 前端）
docker compose up -d --build
```

首次执行会拉取基础镜像并构建前后端镜像，视网络情况约需 **5～20 分钟**，请耐心等待。看到一串 `Started` / `Healthy` 输出即完成。

### 3.3 验证服务状态

```bash
docker compose ps
```

预期 `dx-mysql`、`dx-redis`、`dx-registry`、`cloud-api`（后端）、`cloud-web`(前端)、`cloud-proxy`（Traefik）全部处于 **running / healthy** 状态。后端依赖 MySQL/Redis 健康检查，启动稍慢属正常现象，等待 1～2 分钟再查一次即可。

### 3.4 访问控制台

浏览器打开：

> **http://localhost**

使用初始管理员账号登录：

| 用户名 | 密码 |
|---|---|
| `admin` | `Admin@123456` |

> **安全提醒**：初始密码为全网公开的默认值，**登录后请立即修改**。（如需在首次启动前就自定义，可设置环境变量 `ADMIN_INIT_PASSWORD`。）

其他常用入口：

| 入口 | 地址 |
|---|---|
| Go API 健康检查 | http://localhost/healthz |
| 业务探针 | http://localhost/api/v1/health |
| Traefik Dashboard | http://localhost:8080 |

至此，平台已经完整跑起来了。

---

## 4. 常见问题

### Q1：端口冲突（80 / 8080 / 13306 / 16379 / 15000 被占用）

`docker compose up -d` 报 `port is already allocated` 时：

1. 复制环境变量模板：`cp .env.example .env`；
2. 编辑 `.env`，调整对应端口后重新 `docker compose up -d`：
   - 控制台 80 端口 → `TRAEFIK_HTTP_PORT`（例如改成 8081，之后访问 http://localhost:8081）；
   - Traefik Dashboard 8080 端口 → `TRAEFIK_DASHBOARD_PORT`；
   - MySQL / Redis / Registry 宿主映射端口 → `MYSQL_EXPOSE_PORT` / `REDIS_EXPOSE_PORT` / `REGISTRY_EXPOSE_PORT`（默认 13306 / 16379 / 15000，已刻意避开本机 3306 / 6379）。

### Q2：镜像拉取超时 / 非常慢 / 报 `toomanyrequests`、`unauthorized`

按优先级尝试：

1. **配置镜像加速器**：见 [第 2 节方案 a)](#方案-a-给-docker-配置镜像加速器registry-mirrors)；
2. **使用代理**：
   - Docker Desktop → **Settings → Resources → Proxies**，开启 Manual proxy configuration，填入你的 HTTP/HTTPS 代理地址（如 `http://127.0.0.1:7890`），Apply & Restart；
   - 或使用系统代理 / TUN 全局代理模式后重试；
3. 大镜像中途失速时，可显式从加速源拉取后打标签，例如：

```bash
docker pull hub.rat.dev/library/mysql:8.0
docker tag hub.rat.dev/library/mysql:8.0 mysql:8.0
```

### Q3：想清空数据从头再来

```bash
docker compose down -v
```

`-v` 会**删除所有数据卷**（MySQL 数据、Redis 数据、Registry 镜像、缓存等全部清空），然后重新 `docker compose up -d --build` 即可获得一套全新环境。**该操作不可恢复，请谨慎使用。**

### Q4：如何查看日志排查问题

```bash
docker compose logs -f backend      # 后端日志（最常用）
docker compose logs -f frontend     # 前端日志
docker compose logs -f mysql        # 数据库日志
docker compose logs -f traefik      # 网关日志
```

常见判断：若 `backend` 一直不 healthy，先 `docker compose ps` 确认 `mysql`、`redis` 已 healthy；后端启动依赖两者就绪。

### Q5：停止平台

```bash
docker compose down        # 停止并移除容器（数据卷保留，下次启动数据还在）
```

---

## 5. 下一步

- 想把一个 **Spring Boot 项目**接入平台 CI/CD（push 代码自动构建、自动部署）？请阅读：**[docs/springboot-cicd.md](docs/springboot-cicd.md)**
- 了解三种运行模式、目录结构、常用命令等更多细节：请阅读项目根目录 **[README.md](README.md)**
- 架构设计原理：`docs/phase0-architecture.md`
