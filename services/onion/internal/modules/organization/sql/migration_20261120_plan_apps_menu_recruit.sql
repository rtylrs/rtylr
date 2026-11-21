-- Menu and Recruit were in no plan, so the plan gate refused them to every
-- business, paying ones included. The paid tiers now carry both. shared/
-- subscriptions.sql seeds the same lists, but it is recorded as applied on
-- existing databases and never re-runs there, so the rows are patched here.
-- Each statement skips a plan that already has the app, so a fresh database
-- and a retried run are both no-ops.
UPDATE plans
SET apps = JSON_ARRAY_APPEND(apps, '$', 'menu')
WHERE slug IN ('pro', 'scale')
  AND NOT JSON_CONTAINS(apps, JSON_QUOTE('menu'));

UPDATE plans
SET apps = JSON_ARRAY_APPEND(apps, '$', 'recruit')
WHERE slug IN ('pro', 'scale')
  AND NOT JSON_CONTAINS(apps, JSON_QUOTE('recruit'));
