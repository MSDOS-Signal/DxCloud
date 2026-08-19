-- ============================================================
-- 000005_apps.sql  Phase 6：应用 / 版本 / 部署 / 域名
-- ============================================================

CREATE TABLE IF NOT EXISTS applications (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id           BIGINT UNSIGNED NULL,
    project_id       BIGINT UNSIGNED NULL,
    owner_id         BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name             VARCHAR(64)  NOT NULL,
    type             VARCHAR(32)  NOT NULL DEFAULT 'custom' COMMENT 'node/go/java/python/nginx/mysql/redis/postgres/custom',
    image            VARCHAR(255) NOT NULL DEFAULT '' COMMENT '当前/默认镜像（部署时可覆盖）',
    git_url          VARCHAR(512) NOT NULL DEFAULT '',
    git_branch       VARCHAR(64)  NOT NULL DEFAULT 'main',
    port             INT          NOT NULL DEFAULT 80,
    health_check_path VARCHAR(255) NOT NULL DEFAULT '',
    env              TEXT         NULL COMMENT '默认环境变量 JSON',
    domain           VARCHAR(255) NOT NULL DEFAULT '',
    active_deployment_id BIGINT UNSIGNED NULL,
    status           TINYINT      NOT NULL DEFAULT 1,
    created_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at       DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_apps_project_name (project_id, name),
    KEY idx_apps_project (project_id),
    KEY idx_apps_owner (owner_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '应用（PaaS）';

CREATE TABLE IF NOT EXISTS application_versions (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_id BIGINT UNSIGNED NOT NULL,
    version        VARCHAR(64)  NOT NULL,
    image_ref      VARCHAR(255) NOT NULL,
    commit_sha     VARCHAR(64)  NOT NULL DEFAULT '',
    status         VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active/superseded/failed',
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_app_versions (application_id, version),
    KEY idx_app_versions_app (application_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '应用版本';

CREATE TABLE IF NOT EXISTS deployments (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id         BIGINT UNSIGNED NULL,
    project_id     BIGINT UNSIGNED NULL,
    application_id BIGINT UNSIGNED NOT NULL,
    environment_id BIGINT UNSIGNED NULL,
    version_id     BIGINT UNSIGNED NULL,
    version        VARCHAR(64)  NOT NULL DEFAULT '',
    image_ref      VARCHAR(255) NOT NULL,
    strategy       VARCHAR(16)  NOT NULL DEFAULT 'blue-green',
    status         VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending/deploying/success/failed/rolled-back',
    health_status  VARCHAR(16)  NOT NULL DEFAULT '' COMMENT 'healthy/unhealthy',
    trigger_type    VARCHAR(16)  NOT NULL DEFAULT 'manual' COMMENT 'manual/webhook/pipeline',
    pipeline_run_id BIGINT UNSIGNED NULL,
    container_id   VARCHAR(128) NOT NULL DEFAULT '',
    container_name VARCHAR(128) NOT NULL DEFAULT '',
    config_json    TEXT         NULL COMMENT '容器配置快照（用于降级重建/回滚）',
    note           VARCHAR(255) NOT NULL DEFAULT '',
    started_at     DATETIME(3)  NULL,
    finished_at    DATETIME(3)  NULL,
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_deployments_app (application_id, status),
    KEY idx_deployments_project (project_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '部署记录';

CREATE TABLE IF NOT EXISTS domains (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id         BIGINT UNSIGNED NULL,
    project_id     BIGINT UNSIGNED NULL,
    application_id BIGINT UNSIGNED NULL,
    domain         VARCHAR(255) NOT NULL,
    target_port    INT          NOT NULL DEFAULT 80,
    tls            TINYINT(1)   NOT NULL DEFAULT 0,
    cert_id        BIGINT UNSIGNED NULL,
    status         TINYINT      NOT NULL DEFAULT 1,
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at     DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_domains (domain),
    KEY idx_domains_app (application_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '域名';
