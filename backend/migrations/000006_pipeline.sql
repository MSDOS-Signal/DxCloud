-- ============================================================
-- 000006_pipeline.sql  Phase 7：Pipeline 引擎
-- ============================================================

CREATE TABLE IF NOT EXISTS pipelines (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id      BIGINT UNSIGNED NULL,
    project_id  BIGINT UNSIGNED NULL,
    owner_id    BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name        VARCHAR(64)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    definition  MEDIUMTEXT   NULL COMMENT 'Pipeline YAML 定义',
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)  NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_pipelines_name (name),
    KEY idx_pipelines_project (project_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Pipeline 定义';

CREATE TABLE IF NOT EXISTS pipeline_steps (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    pipeline_id BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    type        VARCHAR(32)  NOT NULL,
    seq         INT          NOT NULL DEFAULT 0,
    config_json TEXT         NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_pipeline_steps (pipeline_id, seq),
    KEY idx_pipeline_steps_pipe (pipeline_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Pipeline 步骤（解析快照）';

CREATE TABLE IF NOT EXISTS pipeline_runs (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    pipeline_id  BIGINT UNSIGNED NOT NULL,
    run_no       INT          NOT NULL DEFAULT 1,
    trigger_type VARCHAR(16)  NOT NULL DEFAULT 'manual' COMMENT 'manual/webhook',
    ref          VARCHAR(64)  NOT NULL DEFAULT '',
    commit_sha   VARCHAR(64)  NOT NULL DEFAULT '',
    status       VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/canceled',
    started_at   DATETIME(3)  NULL,
    finished_at  DATETIME(3)  NULL,
    duration_ms  BIGINT       NOT NULL DEFAULT 0,
    triggered_by BIGINT UNSIGNED NULL,
    created_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_pipeline_runs_pipe (pipeline_id, status),
    KEY idx_pipeline_runs_status (status)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Pipeline 执行记录';

CREATE TABLE IF NOT EXISTS pipeline_job_runs (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    pipeline_run_id BIGINT UNSIGNED NOT NULL,
    step_id         BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name            VARCHAR(64)  NOT NULL,
    type            VARCHAR(32)  NOT NULL,
    status          VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/skipped',
    exit_code       INT          NOT NULL DEFAULT 0,
    container_id    VARCHAR(128) NOT NULL DEFAULT '',
    log_path        VARCHAR(255) NOT NULL DEFAULT '',
    started_at      DATETIME(3)  NULL,
    finished_at     DATETIME(3)  NULL,
    created_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_pipeline_jobs_run (pipeline_run_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Pipeline 步骤任务';
