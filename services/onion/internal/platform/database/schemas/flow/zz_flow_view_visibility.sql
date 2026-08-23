-- S2: a saved view becomes private to its creator or shared business-wide.
--
-- Views migrated to the business-wide pool as shared, which the proposal chose
-- deliberately ("saved views stay shared business-wide at migration; a per-view
-- visibility toggle is reserved for later"). So 'shared' is the default and
-- every existing row keeps exactly the reach it has today. A default of
-- 'private' would silently hide every colleague's view.
--
-- Named as its own zz_ file rather than an edit to flow/schema.sql because
-- migrate.sh records applied files by name: editing an already-applied file is a
-- silent no-op. Guarded on INFORMATION_SCHEMA because zz_ files re-run on every
-- pass, and migrate.sh records a FAILED migration as applied behind a green
-- check, so an unguarded ALTER would fail silently and forever after the first
-- run.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_views'
     AND COLUMN_NAME = 'visibility') = 0,
  'ALTER TABLE flow_views ADD COLUMN visibility ENUM(''private'',''shared'') NOT NULL DEFAULT ''shared'' AFTER is_default',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- A private view is only ever read by its creator, so the useful index is
-- (business_id, visibility, created_by) rather than visibility alone.
SET @idx := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_views'
     AND INDEX_NAME = 'idx_flow_views_business_visibility') = 0,
  'ALTER TABLE flow_views ADD INDEX idx_flow_views_business_visibility (business_id, visibility, created_by)',
  'SELECT 1'
);
PREPARE stmt2 FROM @idx; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;
