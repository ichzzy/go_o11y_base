-- 角色表初始數據
INSERT INTO `role` (`id`, `name`, `created_at`, `updated_at`) VALUES
(1, 'admin', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

-- 權限表初始數據 (type=3 為 API)
INSERT INTO `permission` (`id`, `parent_id`, `name`, `code`, `type`, `http_method`, `http_path`, `created_at`, `updated_at`) VALUES
(1, 0, '所有權限', 'all', 3, '*', '*', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

-- 角色權限綁定
INSERT INTO `role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`) VALUES
(1, 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000); -- 賦予 admin 全部權限

-- 用戶角色關係
INSERT INTO `user_role` (`user_id`, `role_id`, `created_at`, `updated_at`) VALUES
(1, 1, UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000); -- 綁定 admin 的角色