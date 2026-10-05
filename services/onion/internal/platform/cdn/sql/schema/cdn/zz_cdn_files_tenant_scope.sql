-- Tenant boundary for cdn_files.
--
-- cdn_files recorded no owner, so every stored asset was global: there was no
-- column on which a tenant filter could even be expressed, GET /files/{uuid}
-- served any file to anyone, POST /sign minted a delivery signature for any
-- file_uuid, and PURGE /files/{uuid} deleted any file.
--
-- business_id is nullable on purpose. Rows that existed before this migration
-- cannot be attributed to a tenant, and the store treats a NULL business_id as
-- belonging to no tenant at all — it matches no tenant-scoped read, write or
-- delete, and is_public defaults to 0 so it is not anonymously readable
-- either. Such rows are inert until an operator attributes them:
--
--   UPDATE cdn_files SET business_id = <business id> WHERE file_uuid = '<uuid>';
--
-- That is a deliberate availability break: every pre-existing asset stops
-- being publicly readable on the deploy that applies this file. The report at
-- the bottom prints how many rows need attribution so the size of the
-- remediation is visible during the migration run rather than discovered by
-- users. New uploads carry the resolved tenant from the request principal.

ALTER TABLE cdn_files
    ADD COLUMN business_id BIGINT      NULL DEFAULT NULL AFTER file_uuid,
    ADD COLUMN is_public   BOOLEAN     NOT NULL DEFAULT FALSE AFTER original_filename;

-- Every tenant-scoped lookup filters on (business_id, file_uuid); the
-- delivery-path lookup for public assets filters on is_public alone.
CREATE INDEX idx_cdn_files_business_id_uuid ON cdn_files (business_id, file_uuid);
CREATE INDEX idx_cdn_files_is_public        ON cdn_files (is_public, file_uuid);

SELECT COUNT(*) AS cdn_files_pending_tenant_attribution
  FROM cdn_files
 WHERE business_id IS NULL;
