-- 用戶初始數據
INSERT INTO `user` (`id`, `email`, `password`, `status`, `created_at`, `updated_at`)
VALUES
(1, 'admin', '$2a$10$EB4BSA4XsPnCzI2WdQ4YZuMJy2nC947.hlZrE33P/w9PP/ZfkOwkC', 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);