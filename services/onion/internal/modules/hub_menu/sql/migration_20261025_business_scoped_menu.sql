-- Menu banners and the navigation menu belong to a business. owner_id goes
-- nullable so this build can stop writing it; a later release drops it.

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hub_menu_banners' AND column_name = 'owner_id' AND is_nullable = 'NO');
SET @stmt := IF(@apply, 'ALTER TABLE hub_menu_banners MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE hub_menu_stmt FROM @stmt;
EXECUTE hub_menu_stmt;
DEALLOCATE PREPARE hub_menu_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'hub_navigation_menus' AND column_name = 'owner_id' AND is_nullable = 'NO');
SET @stmt := IF(@apply, 'ALTER TABLE hub_navigation_menus MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE hub_menu_stmt FROM @stmt;
EXECUTE hub_menu_stmt;
DEALLOCATE PREPARE hub_menu_stmt;
