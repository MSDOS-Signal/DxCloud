-- ============================================================
-- 000008_ops.sql  Phase 9：指标采样 / 操作日志 / 通知
-- ============================================================

CREATE TABLE IF NOT EXISTS metric_samples (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    kind        VARCHAR(8)   NOT NULL DEFAULT 'ecs' COMMENT 'ecs/app',
    ref_id      BIGINT UNSIGNED NOT NULL COMMENT '实例/部署 ID',
    ts          DATETIME(3)  NOT NULL,
    cpu_pct     DECIMAL(8, 2) NOT NULL DEFAULT 0,
    mem_used    BIGINT       NOT NULL DEFAULT 0,
    mem_limit   BIGINT       NOT NULL DEFAULT 0,
    net_rx      BIGINT       NOT NULL DEFAULT 0,
    net_tx      BIGINT       NOT NULL DEFAULT 0,
    disk_read   BIGINT       NOT NULL DEFAULT 0,
    disk_write  BIGINT       NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_metric_ref_ts (kind, ref_id, ts),
    KEY idx_metric_ts (ts)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '监控指标采样（分钟级）';

CREATE TABLE IF NOT EXISTS operation_logs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id      BIGINT UNSIGNED NULL,
    user_id     BIGINT UNSIGNED NULL,
    module      VARCHAR(32)  NOT NULL DEFAULT '',
    action      VARCHAR(64)  NOT NULL,
    target_type VARCHAR(32)  NOT NULL DEFAULT '',
    target_id   VARCHAR(64)  NOT NULL DEFAULT '',
    target_name VARCHAR(128) NOT NULL DEFAULT '',
    result      TINYINT      NOT NULL DEFAULT 1 COMMENT '1=成功 2=失败',
    duration_ms BIGINT       NOT NULL DEFAULT 0,
    ip          VARCHAR(64)  NOT NULL DEFAULT '',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_oplogs_user (user_id, created_at),
    KEY idx_oplogs_module (module, created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '操作日志（用户操作流水）';

CREATE TABLE IF NOT EXISTS notifications (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    org_id     BIGINT UNSIGNED NULL,
    type       VARCHAR(32)  NOT NULL DEFAULT 'system' COMMENT 'system/pipeline/deploy',
    title      VARCHAR(128) NOT NULL,
    content    VARCHAR(512) NOT NULL DEFAULT '',
    read_at    DATETIME(3)  NULL,
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_notifications_user (user_id, read_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '站内通知';
