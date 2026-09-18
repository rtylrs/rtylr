-- hr_time_entries.owner_id is the business identifier, matching business_id.
-- Flow historically wrote the resolved account owner here, which made
-- owner-scoped cross-module reads disagree with HR-created entries. The Flow
-- writer now uses business_id; normalize existing Flow rows idempotently.
UPDATE hr_time_entries
SET owner_id = business_id
WHERE source = 'flow'
  AND owner_id <> business_id;
