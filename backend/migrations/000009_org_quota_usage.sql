-- ============================================================
-- 000009_org_quota_usage.sql  Phase 10：配额 / 用量（组织与成员表 Phase 1 已有）
-- ============================================================

CREATE TABLE IF NOT EXISTS resource_quotas (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id        BIGINT UNSIGNED NOT NULL,
    project_id    BIGINT UNSIGNED NULL,
    resource_type VARCHAR(32)  NOT NULL COMMENT 'ecs_count/cpu/memory/storage/network/pipeline',
    limit_value   BIGINT       NOT NULL DEFAULT 0,
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_resource_quotas (org_id, project_id, resource_type),
    KEY idx_resource_quotas_org (org_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '资源配额';

CREATE TABLE IF NOT EXISTS resource_usage (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id        BIGINT UNSIGNED NOT NULL,
    project_id    BIGINT UNSIGNED NULL,
    resource_type VARCHAR(32)  NOT NULL COMMENT 'cpu_hour/mem_gb_hour/disk_gb_hour',
    used_value    DECIMAL(18, 4) NOT NULL DEFAULT 0,
    period        DATETIME(3)  NOT NULL COMMENT '所属小时（整点）',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_resource_usage_org (org_id, period),
    KEY idx_resource_usage_period (period)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '资源用量（虚拟计费）';
