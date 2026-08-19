# Spring Boot 项目接入 DxCloud CI/CD 完整指南

> 本文介绍如何把一个 Spring Boot 项目接入 DxCloud 的 Pipeline 流水线，实现 **git push 后自动构建镜像并蓝绿部署上线**。
>
> 文中所有 Pipeline YAML 示例均严格按照引擎实现（`backend/internal/pipeline/engine.go`）的 schema 编写，可直接粘贴到控制台使用。

---

## 目录

1. [原理与流程](#1-原理与流程)
2. [准备工作](#2-准备工作)
3. [Spring Boot 侧：添加 Dockerfile](#3-spring-boot-侧添加-dockerfile)
4. [完整 Pipeline YAML 示例](#4-完整-pipeline-yaml-示例)
5. [在平台 UI 中创建 Pipeline](#5-在平台-ui-中创建-pipeline)
6. [配置 GitHub Webhook（push 自动触发）](#6-配置-github-webhookpush-自动触发)
7. [验证与日常使用](#7-验证与日常使用)
8. [步骤类型与字段速查表](#8-步骤类型与字段速查表)
9. [常见问题](#9-常见问题)

---

## 1. 原理与流程

```text
开发者本地
   │  git push（推送到 GitHub）
   ▼
GitHub 仓库
   │  push 事件 → Webhook 回调（携带 X-Hub-Signature-256 HMAC 签名）
   ▼
DxCloud Webhook 接收端  POST /api/v1/webhooks/github/<hook_code>
   │  ① 校验 HMAC-SHA256 签名（secret 在平台生成）
   │  ② 分支过滤（如仅 main）
   ▼
创建 Pipeline 运行（入 Redis 队列 dx:pipe:queue，Worker 消费）
   ▼
Pipeline 引擎按顺序执行步骤（每步独立状态 + 实时日志）：
   │
   ├─ 步骤1 git            拉取你的 Spring Boot 仓库代码（隔离 Job 容器，clone 到 /workspace）
   ├─ 步骤2 shell          Maven 构建打包：mvn -DskipTests package（产出 target/*.jar）
   ├─ 步骤3 docker-build   构建镜像（平台服务端执行，BuildKit；构建上下文 = 本次运行的 workspace）
   ├─ 步骤4 docker-push    推送镜像到平台内置私有 Registry
   ├─ 步骤5 docker-deploy  蓝绿部署到指定应用（平台部署服务执行：
   │                       创建新版本容器 → 健康检查 → Traefik 流量切换 → 停旧容器）
   └─ 步骤6 wait-health    业务级健康探测（见 §4 中关于该步骤当前状态的说明）
   ▼
运行结束：success / failed / canceled，站内通知触发者
```

要点说明：

- **git / shell 步骤**在一次性隔离 Job 容器中执行（限制 CPU 2 核 / 内存 2048MB、非特权、**不挂载 docker.sock**），同一次运行的所有步骤共享一个 workspace 卷（挂载在容器内 `/workspace`），因此步骤 2 打包出的 `target/*.jar` 可被步骤 3 的镜像构建直接使用。
- **docker-build / docker-push / docker-deploy / wait-health 步骤**由平台服务端执行：构建走 BuildKit（架构设计中的 kaniko 无 socket 隔离构建为演进方向，二者都不把 Docker 权限下放给 Job 容器）；部署由平台部署服务完成蓝绿切换。
- **蓝绿部署自带健康检查**：`docker-deploy` 会等待新容器通过 HTTP/TCP 健康探测（默认最长 60 秒）后才切换流量，失败自动保留旧版本。

---

## 2. 准备工作

开始之前，请确认：

1. **平台已启动**：按 [GETTING-STARTED.md](../GETTING-STARTED.md) 完成 `docker compose up -d --build`，并能访问 http://localhost；
2. **Spring Boot 代码在 GitHub 仓库中**（公开仓库；私有仓库的限制见 [§9 常见问题](#9-常见问题)）；
3. **本地可构建**：项目能用 `mvn -DskipTests package` 成功打包出 `target/*.jar`（JDK 17）；
4. **可直接使用示例模板**：复制仓库中的 `examples/springboot/Dockerfile` 和 `examples/springboot/dxcloud-pipeline.yml` 到自己的 Spring Boot 仓库，替换其中的仓库地址、应用名即可。
4. **在平台创建应用**（`docker-deploy` 步骤需要引用一个已存在的应用名）：
   - 控制台左侧菜单 **「PaaS → 应用」** → **创建应用**；
   - 名称：如 `spring-demo`（记住这个名字，YAML 里要用）；
   - 端口：`8080`（Spring Boot 默认监听端口）；
   - 健康检查路径：引入了 Spring Boot Actuator 填 `/actuator/health`；没有则填 `/` 或留空（留空 = TCP 端口探测）；
   - 镜像一栏可留空或随意填写，Pipeline 部署时会覆盖为新构建的镜像。

---

## 3. Spring Boot 侧：添加 Dockerfile

在你的 Spring Boot 仓库**根目录**添加 `Dockerfile`（流水线第 3 步会用它构建镜像）：

```dockerfile
# 运行环境：Eclipse Temurin JRE 17
FROM eclipse-temurin:17-jre

WORKDIR /app

# 由流水线第 2 步 mvn package 产出（构建上下文 = 仓库根目录）
COPY target/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "/app/app.jar"]
```

再添加 `.dockerignore`（减小构建上下文，**注意不要忽略 target/**）：

```text
.git
.idea
*.iml
src
```

> 说明：`COPY target/*.jar` 依赖第 2 步 shell 在 workspace 中产出的 `target/` 目录；`docker-build` 的构建上下文就是整个 workspace（即仓库根目录），因此 Dockerfile 放在仓库根目录、路径写相对路径即可。

提交并推送这两个文件到你的 GitHub 仓库。

---

## 4. 完整 Pipeline YAML 示例

以下定义可直接粘贴到控制台「创建 Pipeline」的 YAML 输入框（请先替换 `git` 步骤的仓库地址、应用名等占位内容）：

```yaml
name: springboot-ci-cd
timeout: 2h
env:
  MAVEN_OPTS: "-Xmx1024m"

steps:
  # ---------- 步骤 1：拉取 Spring Boot 仓库代码 ----------
  - name: fetch-code
    type: git
    url: https://github.com/<你的用户名>/spring-boot-demo.git   # 替换为你的仓库地址
    branch: main

  # ---------- 步骤 2：Maven 构建打包（跳过测试） ----------
  - name: maven-package
    type: shell
    timeout: 30m
    script: |
      set -e
      # shell 步骤运行在固定的 alpine:3.20 隔离容器中（当前版本不支持为
      # shell 步骤指定自定义镜像），因此先安装 JDK17 与 Maven
      apk add --no-cache openjdk17 maven
      export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
      export PATH="$JAVA_HOME/bin:$PATH"
      java -version && mvn -v
      # 中国大陆网络环境建议配置阿里云 Maven 镜像，显著加快依赖下载
      mkdir -p ~/.m2
      cat > ~/.m2/settings.xml <<'EOF'
      <settings>
        <mirrors>
          <mirror>
            <id>aliyunmaven</id>
            <mirrorOf>central</mirrorOf>
            <url>https://maven.aliyun.com/repository/public</url>
          </mirror>
        </mirrors>
      </settings>
      EOF
      # 构建打包（跳过测试；如需跑测试去掉 -DskipTests 即可）
      mvn -B -DskipTests package
      ls -l target/

  # ---------- 步骤 3：构建镜像（服务端 BuildKit，上下文 = workspace） ----------
  - name: build-image
    type: docker-build
    dockerfile: Dockerfile
    tags:
      - host.docker.internal:15000/default/spring-demo:latest

  # ---------- 步骤 4：推送到平台内置私有 Registry ----------
  - name: push-image
    type: docker-push
    tags:
      - host.docker.internal:15000/default/spring-demo:latest

  # ---------- 步骤 5：蓝绿部署到平台应用 ----------
  - name: deploy
    type: docker-deploy
    application: spring-demo                                    # §2 中创建的应用名
    image: host.docker.internal:15000/default/spring-demo:latest

  # ---------- 步骤 6：业务健康探测 ----------
  - name: health
    type: wait-health
    url: http://spring-demo.你的域名/actuator/health            # 替换为应用实际访问地址
    allow_failure: true
```

### 逐段解释

| 字段 | 说明 |
|---|---|
| `name` / `timeout` | 流水线名称；全流程超时（Go duration 格式，如 `30m`、`2h`，缺省 2h） |
| `env` | 全局环境变量，会注入 git/shell 步骤的 Job 容器（示例中的 `MAVEN_OPTS` 限制 Maven 堆内存） |
| `git.url` / `git.branch` | 仓库地址（必填）与分支（缺省为仓库默认分支）；以 `git clone --depth 1` 浅克隆到 workspace |
| `shell.script` | 在 `/workspace` 下以 `sh -c` 执行的脚本（必填）；`set -e` 保证任一命令失败即步骤失败 |
| `docker-build.dockerfile` | Dockerfile 相对路径，缺省 `Dockerfile` |
| `docker-build.tags` / `docker-push.tags` | 镜像标签列表（必填）；推送目标即平台内置 Registry（宿主端口 15000） |
| `docker-deploy.application` / `.image` | 目标应用名（必须已存在）与部署镜像（必填） |
| `wait-health.url` | 部署完成后的健康探测地址 |
| `allow_failure` | 该步骤失败时标记为 skipped 并继续后续步骤（示例中用于 wait-health，原因见下） |
| 每步 `timeout` | 单步超时（缺省 30m，对 git/shell 的 Job 容器生效） |

### 关于 wait-health 步骤的重要说明

- `wait-health` 已在引擎 schema 白名单内（`url` 为其配置字段），YAML 校验可以通过；但**当前版本的运行时实现尚在完善中**，执行到该步骤可能返回「wait-health 将在后续版本支持」。
- 实际上 `docker-deploy` 蓝绿部署**已内置容器健康检查**（HTTP/TCP 探测通过后才切流量），因此省略 wait-health 不影响部署安全性。
- 若你希望现在就有一个显式的业务级探测，可用 shell 步骤轮询代替（把上面 health 步骤替换为）：

```yaml
  - name: health
    type: shell
    timeout: 5m
    script: |
      apk add --no-cache curl
      for i in $(seq 1 30); do
        code=$(curl -s -o /dev/null -w '%{http_code}' http://spring-demo.你的域名/actuator/health || true)
        echo "try $i -> $code"
        [ "$code" = "200" ] && exit 0
        sleep 5
      done
      echo "health check failed" && exit 1
```

### 进阶：使用提交号作为镜像标签

`tags` 与 `docker-deploy.image` 支持内置变量 `${COMMIT_SHA}`（Webhook 触发时为 push 的提交 SHA）与 `${REF}`（分支名），可实现每次提交一个不可变镜像版本：

```yaml
  - name: build-image
    type: docker-build
    dockerfile: Dockerfile
    tags:
      - host.docker.internal:15000/default/spring-demo:${COMMIT_SHA}
  - name: push-image
    type: docker-push
    tags:
      - host.docker.internal:15000/default/spring-demo:${COMMIT_SHA}
  - name: deploy
    type: docker-deploy
    application: spring-demo
    image: host.docker.internal:15000/default/spring-demo:${COMMIT_SHA}
```

> 注意：**手动点击「运行」触发时没有提交 SHA（变量会被替换为空串）**，因此这种写法仅适合 Webhook 触发场景；手动 + Webhook 混用时建议像主示例那样使用固定标签（如 `:latest`）。

---

## 5. 在平台 UI 中创建 Pipeline

1. 登录控制台（http://localhost，初始账号 `admin` / `Admin@123456`）；
2. 左侧菜单 **「DevOps → CI/CD Pipeline」** 进入流水线列表页；
3. 点击右上角 **「创建 Pipeline」** 按钮，在弹窗中填写：
   - **名称**（必填）：如 `springboot-ci-cd`；
   - **描述**（可选）：如「Spring Boot 演示应用自动构建部署」；
   - **定义 YAML**：粘贴 [§4](#4-完整-pipeline-yaml-示例) 的完整 YAML（记得替换仓库地址与应用名）；
4. 点击 **「创建」**。若 YAML 不符合 schema（如步骤缺 `name`、shell 缺 `script`、git 缺 `url`、类型不在白名单），会在此处直接给出中文校验错误，按提示修改即可；
5. 创建成功后，点击该行的 **「详情/运行」** 进入详情页，可先点 **「▶ 运行」** 手动触发一次，跳转到运行详情页实时查看每个步骤的状态与日志（Maven 下载依赖较慢时属正常现象）。

> 修改流水线：详情页点击 **「编辑定义」**，修改 YAML 后 **「保存」**，下次运行生效。

> 中国大陆网络环境下，示例已内置 Maven 阿里云镜像；进入「设置 → 区域与镜像源」选择中国大陆后，Pipeline 的 shell/git 步骤基础镜像也会自动使用加速源。

---

## 6. 配置 GitHub Webhook（push 自动触发）

### 6.1 在平台生成 Webhook 地址与 Secret

1. 进入该 Pipeline 的详情页，点击 **「Webhook」** 按钮展开面板；
2. 填写：
   - **Provider**：选择 `GitHub`；
   - **分支过滤**：如 `main`（只接收 main 分支的 push；支持 `release/*` 通配；留空 = 全部分支）；
   - **Secret**：留空自动生成，也可自填一个强口令；
3. 点击 **「创建」**，页面会显示：
   - **URL**：形如 `http://localhost/api/v1/webhooks/github/<hook_code>`；
   - **Secret**：**仅展示这一次，请立即复制保存**（后续要填到 GitHub）。

> 若平台部署在服务器/内网穿透域名上，请把 URL 中的 `localhost` 换成 GitHub 可访问到的平台地址（如 `http://你的域名/api/v1/webhooks/github/<hook_code>`）。

### 6.2 在 GitHub 仓库添加 Webhook

1. 打开你的 Spring Boot 仓库页面 → **Settings → Webhooks → Add webhook**（可能需要输入 GitHub 密码确认）；
2. 填写：
   - **Payload URL**：粘贴上一步平台生成的完整 URL；
   - **Content type**：选择 **application/json**（必须）；
   - **Secret Token**：粘贴平台生成的 Secret；
3. **Which events would you like to trigger this webhook?** 保持默认 **Just the push event**；
4. 勾选 **Active** → 点击 **Add webhook**；
5. GitHub 会立即发送一次 ping 测试，页面下方 **Recent Deliveries** 出现绿勾（200）即配置成功。若返回 401，检查 Secret 是否一致、Content type 是否为 application/json。

### 6.3 端到端验证

```bash
# 本地随便改一行代码（比如 README），提交并推送
git add .
git commit -m "ci: trigger pipeline"
git push origin main
```

回到平台 Pipeline 详情页，可看到一条 **触发方式 = webhook** 的新运行记录；全部步骤变绿后，到 **「PaaS → 应用 → spring-demo」** 查看部署版本，或 **「DevOps → 部署」** 查看本次蓝绿部署记录。

---

## 7. 验证与日常使用

- **查看运行与日志**：Pipeline 详情页「运行历史」→ 点「详情」，逐步查看实时日志（git clone、Maven 构建、镜像构建推送、部署输出）；
- **手动触发**：详情页「▶ 运行」（适合调试 YAML）；
- **取消运行**：运行详情页可取消排队中/运行中的任务；
- **回滚**：「PaaS → 应用」中选择历史版本重新部署即可（版本记录永不删除）；
- **查看推送的镜像**：「资源 → 镜像中心」或宿主访问 `http://127.0.0.1:15000/v2/_catalog`。

---

## 8. 步骤类型与字段速查表

引擎仅接受以下 6 种步骤类型（白名单），YAML 顶层字段为 `name` / `timeout` / `env` / `steps`：

| 步骤类型 | 必填字段 | 可选字段 | 执行方式 |
|---|---|---|---|
| `git` | `name`、`type`、`url` | `branch`、`timeout`、`allow_failure` | 隔离 Job 容器（alpine/git，`clone --depth 1` 到 /workspace） |
| `shell` | `name`、`type`、`script` | `timeout`、`allow_failure` | 隔离 Job 容器（固定 alpine:3.20，`sh -c`，工作目录 /workspace） |
| `docker-build` | `name`、`type`、`tags` | `dockerfile`（缺省 `Dockerfile`） | 平台服务端（BuildKit；上下文 = 本次运行 workspace） |
| `docker-push` | `name`、`type`、`tags` | — | 平台服务端（逐个推送 tags） |
| `docker-deploy` | `name`、`type`、`image`、`application` | — | 平台部署服务（蓝绿 + 健康检查 + Traefik 切换） |
| `wait-health` | `name`、`type`、`url` | `allow_failure` | schema 已支持；运行时实现完善中（见 §4 说明） |

补充：

- `tags` / `image` 中支持变量 `${COMMIT_SHA}`、`${REF}`（Webhook 触发时生效，手动触发时为空）；
- `timeout` 为 Go duration 格式（如 `90s`、`30m`、`2h`）：全流程缺省 2h，单步缺省 30m；
- `env` 中的键值对会注入 git/shell 步骤的 Job 容器，并额外注入 `PIPELINE_RUN_ID`；
- 除上表字段外请勿添加其他字段（如为 shell 步骤写 `image` 不会生效——当前版本 shell 固定使用 alpine:3.20 容器，自定义步骤镜像为后续版本能力）。

---

## 9. 常见问题

**Q1：git 步骤克隆私有仓库失败（403/认证失败）？**
当前 `git` 步骤没有独立的凭据字段。两种办法：a) 使用公开仓库；b) 在 `url` 中内嵌 GitHub Personal Access Token，如 `https://<PAT>@github.com/<用户>/<仓库>.git`（注意：该令牌会保存在 Pipeline 定义中，请控制流水线可见范围并定期轮换令牌）。

**Q2：Maven 下载依赖超时？**
已在示例 script 中内置阿里云镜像（`maven.aliyun.com`）。若仍慢，可适当调大该步骤 `timeout`（如 `45m`），首次运行会缓存依赖到镜像层之外，重复构建请以实际网络为准。

**Q3：为什么 shell 步骤不能直接用 `maven:3.9-eclipse-temurin-17` 镜像？**
当前引擎的 shell 步骤固定在 alpine:3.20 隔离容器中执行（V1 架构约束：不支持任意镜像步骤），因此示例通过 `apk add openjdk17 maven` 获得等价的 JDK17 + Maven 3.9 环境；平台后续支持自定义步骤镜像后，可直接切换为 `maven:3.9-eclipse-temurin-17` 而无需 apk 安装步骤。

**Q4：docker-build 报「需要 tags」或找不到 Dockerfile？**
`tags` 是 docker-build/docker-push 的必填字段；`dockerfile` 是相对 workspace（仓库根目录）的路径，请确认 Dockerfile 已提交到仓库根目录。

**Q5：docker-deploy 报 `application "xxx" not found`？**
`application` 必须是平台中已创建的应用名（区分大小写）。请先在「PaaS → 应用」创建，或检查拼写。

**Q6：部署成功但访问不到服务？**
应用默认只在平台内部网络可达。如需对外访问，请在「PaaS → 域名」为应用绑定域名（Traefik 路由），或使用应用绑定的端口映射；wait-health/健康探测的 URL 也应使用该实际可达地址。

**Q7：Webhook 推送后平台没有新运行？**
依次检查：GitHub Recent Deliveries 是否 200；平台 Webhook 分支过滤是否与实际推送分支一致（如过滤 `main` 但推的是 `master`）；平台地址是否可被 GitHub 访问（本机部署需内网穿透或公网地址）。
