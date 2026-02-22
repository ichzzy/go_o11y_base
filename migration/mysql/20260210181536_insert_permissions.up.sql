-- 目錄
INSERT INTO `permission` (`parent_id`, `name`, `code`, `type`, `http_method`, `http_path`, `created_at`, `updated_at`) VALUES
(0, 'IAM', 'iam', 1, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

-- 選單
SET @id := (SELECT id FROM `permission` WHERE `code` = 'iam');
INSERT INTO `permission` (`parent_id`, `name`, `code`, `type`, `http_method`, `http_path`, `created_at`, `updated_at`) VALUES
(@id, '角色管理', 'iam:roles', 2, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '用戶管理', 'iam:users', 2, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

-- APIs
SET @id := (SELECT id FROM `permission` WHERE `code` = 'iam:roles');
INSERT INTO `permission` (`parent_id`, `name`, `code`, `type`, `http_method`, `http_path`, `created_at`, `updated_at`) VALUES
(@id, '取得角色清單', 'iam:roles:list', 3, 'GET', '/v1/roles', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '建立角色(包含權限)', 'iam:roles:create', 3, 'POST', '/v1/roles', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '更新角色(包含權限)', 'iam:roles:update', 3, 'PUT', '/v1/roles/:id', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '取得角色的權限清單', 'iam:roles:list-permissions', 3, 'GET', '/v1/roles/:roleID/permissions', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '取得權限清單', 'iam:roles:permissions', 3, 'GET', '/v1/permissions', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(@id, '取得當前用戶可視菜單', 'iam:roles:get-my-visible-menus', 3, 'GET', '/v1/me/visible-menus', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

SET @id := (SELECT id FROM `permission` WHERE `code` = 'iam:users');
INSERT INTO `permission` (`parent_id`, `name`, `code`, `type`, `http_method`, `http_path`, `created_at`, `updated_at`) VALUES
(@id, '創建用戶', 'iam:users:create', 3, 'POST', '/v1/users', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);