CREATE TABLE IF NOT EXISTS `user` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    `email`      VARCHAR(255)     NOT NULL COMMENT '註冊信箱',
    `password`   VARCHAR(255)     NOT NULL COMMENT '密碼',
    `status`     TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '註冊狀態 0:預設狀態, 1:完成信箱, 2:完成密碼設定, 3:綁定GA',
    `created_at` BIGINT           NOT NULL DEFAULT 0 COMMENT '建立時間',
    `updated_at` BIGINT           NOT NULL DEFAULT 0 COMMENT '修改時間',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用戶表';
