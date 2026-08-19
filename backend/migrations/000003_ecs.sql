-- ============================================================
-- 000003_ecs.sql  Phase 3：ECS 云主机（双状态模型 + 事件）
-- ============================================================

CREATE TABLE IF NOT EXISTS ecs_instances (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    instance_no     VARCHAR(32)  NOT NULL COMMENT '对外实例 ID，形如 i-a1b2c3d4e5f60718',
    org_id          BIGINT UNSIGNED NULL,
    project_id      BIGINT UNSIGNED NULL,
    owner_id        BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(64)  NOT NULL,
    description     VARCHAR(255) NOT NULL DEFAULT '',
    image           VARCHAR(255) NOT NULL,
    cpu             DECIMAL(6, 2) NOT NULL DEFAULT 1 COMMENT '核数',
    memory_mb       INT          NOT NULL DEFAULT 512,
    disk_gb         INT          NOT NULL DEFAULT 10 COMMENT '逻辑磁盘配额 GB',
    network_id      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'Docker 网络 ID（Phase 5 接入）',
    fixed_ip        VARCHAR(64)  NOT NULL DEFAULT '',
    ports           TEXT         NULL COMMENT '端口映射 JSON',
    env             TEXT         NULL COMMENT '环境变量 JSON 数组',
    command         TEXT         NULL COMMENT '启动命令 JSON 数组',
    mounts          TEXT         NULL COMMENT '挂载 JSON（Phase 5）',
    restart_policy  VARCHAR(32)  NOT NULL DEFAULT 'no',
    readonly_rootfs TINYINT(1)   NOT NULL DEFAULT 0,
    desired_state   VARCHAR(16)  NOT NULL DEFAULT 'creating',
    observed_state  VARCHAR(16)  NOT NULL DEFAULT 'creating',
    container_id    VARCHAR(128) NULL,
    container_name  VARCHAR(128) NOT NULL DEFAULT '',
    last_error      TEXT         NULL,
    created_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at      DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_ecs_instance_no (instance_no),
    UNIQUE KEY uk_ecs_container_id (container_id),
    KEY idx_ecs_owner (owner_id),
    KEY idx_ecs_org_proj (org_id, project_id),
    KEY idx_ecs_state (desired_state, observed_state),
    KEY idx_ecs_name (name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'ECS 云主机（底层 Docker 容器）';

CREATE TABLE IF NOT EXISTS ecs_instance_events (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    instance_id BIGINT UNSIGNED NOT NULL,
    event_type  VARCHAR(32)  NOT NULL,
    level       VARCHAR(16)  NOT NULL DEFAULT 'info' COMMENT 'info/warn/error',
    message     VARCHAR(512) NOT NULL DEFAULT '',
    actor_id    BIGINT UNSIGNED NULL,
    request_id  VARCHAR(64)  NOT NULL DEFAULT '',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_ecs_events_instance (instance_id, created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'ECS 实例事件';
