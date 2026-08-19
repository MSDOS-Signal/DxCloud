-- ============================================================
-- 000004_infra.sql  Phase 5：镜像 / 网络 / 存储 / Registry
-- ============================================================

CREATE TABLE IF NOT EXISTS docker_images (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id           BIGINT UNSIGNED NULL,
    project_id       BIGINT UNSIGNED NULL,
    repo             VARCHAR(255) NOT NULL COMMENT 'repository 名（含 registry 前缀）',
    tag              VARCHAR(128) NOT NULL DEFAULT 'latest',
    image_id         VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'digest/sha',
    size_bytes       BIGINT       NOT NULL DEFAULT 0,
    docker_created_at DATETIME(3)  NULL,
    source           VARCHAR(16)  NOT NULL DEFAULT 'pull' COMMENT 'pull/build/import',
    status           VARCHAR(16)  NOT NULL DEFAULT 'ready' COMMENT 'pulling/ready/failed',
    pull_error       VARCHAR(512) NOT NULL DEFAULT '',
    created_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at       DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_docker_images (repo, tag),
    KEY idx_docker_images_org (org_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Docker 镜像（镜像中心）';

CREATE TABLE IF NOT EXISTS docker_networks (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id            BIGINT UNSIGNED NULL,
    project_id        BIGINT UNSIGNED NULL,
    owner_id          BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name              VARCHAR(64)  NOT NULL COMMENT '云网络名（对外）',
    docker_name       VARCHAR(64)  NOT NULL COMMENT 'Docker 网络名 dxn-xxxxxx',
    docker_network_id VARCHAR(128) NOT NULL DEFAULT '',
    driver            VARCHAR(32)  NOT NULL DEFAULT 'bridge',
    subnet            VARCHAR(64)  NOT NULL DEFAULT '',
    gateway           VARCHAR(64)  NOT NULL DEFAULT '',
    ip_range          VARCHAR(64)  NOT NULL DEFAULT '',
    internal          TINYINT(1)   NOT NULL DEFAULT 0,
    created_at        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at        DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_docker_networks_name (name),
    UNIQUE KEY uk_docker_networks_docker (docker_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '云网络（底层 Docker Network）';

CREATE TABLE IF NOT EXISTS docker_volumes (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id      BIGINT UNSIGNED NULL,
    project_id  BIGINT UNSIGNED NULL,
    owner_id    BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name        VARCHAR(64)  NOT NULL COMMENT '云磁盘名（对外）',
    docker_name VARCHAR(64)  NOT NULL COMMENT 'Docker 卷名 dxv-xxxxxx',
    driver      VARCHAR(32)  NOT NULL DEFAULT 'local',
    mountpoint  VARCHAR(255) NOT NULL DEFAULT '',
    capacity_gb INT          NOT NULL DEFAULT 10 COMMENT '软配额 GB',
    used_mb     BIGINT       NOT NULL DEFAULT 0,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_docker_volumes_name (name),
    UNIQUE KEY uk_docker_volumes_docker (docker_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '云磁盘（底层 Docker Volume）';

CREATE TABLE IF NOT EXISTS registries (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id      BIGINT UNSIGNED NULL,
    name        VARCHAR(64)  NOT NULL,
    url         VARCHAR(255) NOT NULL,
    username    VARCHAR(128) NOT NULL DEFAULT '',
    password_enc VARCHAR(255) NOT NULL DEFAULT '',
    type        VARCHAR(16)  NOT NULL DEFAULT 'self' COMMENT 'self/docker-hub/other',
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_registries_org_name (org_id, name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '镜像仓库源';

CREATE TABLE IF NOT EXISTS registry_repositories (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    registry_id BIGINT UNSIGNED NOT NULL,
    org_id      BIGINT UNSIGNED NULL,
    project_id  BIGINT UNSIGNED NULL,
    namespace   VARCHAR(128) NOT NULL DEFAULT '',
    name        VARCHAR(255) NOT NULL,
    visibility  VARCHAR(16)  NOT NULL DEFAULT 'private' COMMENT 'private/public',
    pull_count  BIGINT       NOT NULL DEFAULT 0,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_registry_repos (registry_id, namespace, name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '仓库（namespace/name）';
