CREATE TABLE IF NOT EXISTS `role` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`       VARCHAR(64)     NOT NULL COMMENT '角色名稱',
    `created_at` BIGINT          NOT NULL DEFAULT 0 COMMENT '創建時間',
    `updated_at` BIGINT          NOT NULL DEFAULT 0 COMMENT '更新時間',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_role_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `permission` (
    `id`          BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    `parent_id`   BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '父級ID',
    `name`        VARCHAR(64)      NOT NULL COMMENT '權限名稱',
    `code`        VARCHAR(64)      NOT NULL COMMENT '權限識別碼 (前端權限指令)',
    `type`        TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '權限類型;1:目錄,2:選單,3:API',
    `http_method` VARCHAR(16)      NOT NULL DEFAULT '' COMMENT 'HTTP方法 (僅API類型需要)',
    `http_path`   VARCHAR(255)     NOT NULL DEFAULT '' COMMENT 'HTTP路徑 (僅API類型需要)',
    `created_at`  BIGINT           NOT NULL DEFAULT 0 COMMENT '創建時間',
    `updated_at`  BIGINT           NOT NULL DEFAULT 0 COMMENT '更新時間',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_permission_code` (`code`),
    INDEX `idx_permission_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='權限表';

CREATE TABLE IF NOT EXISTS `role_permission` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `role_id`       BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '權限ID',
    `created_at`    BIGINT          NOT NULL DEFAULT 0 COMMENT '創建時間',
    `updated_at`    BIGINT          NOT NULL DEFAULT 0 COMMENT '更新時間',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_role_permission` (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色權限表';

CREATE TABLE IF NOT EXISTS `user_role` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用戶ID',
    `role_id`    BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `created_at` BIGINT          NOT NULL DEFAULT 0 COMMENT '創建時間',
    `updated_at` BIGINT          NOT NULL DEFAULT 0 COMMENT '更新時間',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_user_role` (`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用戶角色表';
