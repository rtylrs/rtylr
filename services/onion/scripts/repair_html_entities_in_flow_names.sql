-- One-off repair for names stored with HTML entities already encoded, e.g. a
-- workspace called "QuillForge Publishing & Learning" persisted as
-- "QuillForge Publishing &amp; Learning". The frontend renders names as JSX text
-- nodes, which never HTML-decode, so the entity shows literally in the sidebar
-- switcher, the dashboard greeting and the switcher button's accessible name.
--
-- This is data, not code. tryrtylr/rtylr#784 records the negative result on both
-- sides: no frontend write path encodes (workspace creation is a plain JSON POST
-- with no escaping step) and no backend write path does either (`grep -rn
-- 'Escape|Sanitiz|html\.' internal/modules/workspaces/` returns nothing, and
-- onion's single &-to-&amp; encoder cannot write). So nothing in the product can
-- produce or reproduce these rows going forward; only rows written by whatever
-- did it originally need fixing.
--
-- Run the SELECT sections first and review the affected rows. Only run the
-- UPDATE sections after that review - do not run this file unattended against
-- production.
--
-- The UPDATEs decode `&amp;` LAST, and that order matters. Verified against a
-- seeded local database rather than reasoned about:
--
--   'Publishing &amp; Learning'      -> 'Publishing & Learning'      (the bug)
--   'Tricky &amp;lt;tag&amp;gt;'     -> 'Tricky &lt;tag&gt;'         (NOT '<tag>')
--   'Double &amp;amp; Encoded'       -> 'Double &amp; Encoded'
--   'Legitimately Clean Name'        -> unchanged
--
-- The second line is the reason for the order: decoding `&amp;` first would turn
-- '&amp;lt;' into '&lt;', and the next statement would decode that to '<',
-- inventing a character the original name never contained.
--
-- The third line is a limitation, not a bug: each pass decodes exactly one
-- level. A double-encoded value needs the file run again. Re-run the dry runs
-- after any repair and repeat until they return zero rows.

-- ── Dry run: workspace names containing an HTML entity ─────────────────────
SELECT id, owner_id, business_id, name
FROM flow_workspaces
WHERE name LIKE '%&amp;%' OR name LIKE '%&lt;%' OR name LIKE '%&gt;%'
   OR name LIKE '%&quot;%' OR name LIKE '%&#39;%' OR name LIKE '%&apos;%'
ORDER BY id;

-- ── Dry run: project names ─────────────────────────────────────────────────
SELECT id, owner_id, business_id, name
FROM flow_projects
WHERE name LIKE '%&amp;%' OR name LIKE '%&lt;%' OR name LIKE '%&gt;%'
   OR name LIKE '%&quot;%' OR name LIKE '%&#39;%' OR name LIKE '%&apos;%'
ORDER BY id;

-- ── Dry run: task titles ───────────────────────────────────────────────────
SELECT id, owner_id, business_id, title
FROM flow_tasks
WHERE title LIKE '%&amp;%' OR title LIKE '%&lt;%' OR title LIKE '%&gt;%'
   OR title LIKE '%&quot;%' OR title LIKE '%&#39;%' OR title LIKE '%&apos;%'
ORDER BY id;

-- ── Repair: workspace names (run only after reviewing above) ───────────────
-- `&amp;` is decoded last on purpose: see the ordering note in the header.
-- UPDATE flow_workspaces
-- SET name = REPLACE(name, '&apos;', '''')
-- WHERE name LIKE '%&apos;%';
-- UPDATE flow_workspaces
-- SET name = REPLACE(name, '&#39;', '''')
-- WHERE name LIKE '%&#39;%';
-- UPDATE flow_workspaces
-- SET name = REPLACE(name, '&quot;', '"')
-- WHERE name LIKE '%&quot;%';
-- UPDATE flow_workspaces
-- SET name = REPLACE(REPLACE(name, '&lt;', '<'), '&gt;', '>')
-- WHERE name LIKE '%&lt;%' OR name LIKE '%&gt;%';
-- UPDATE flow_workspaces
-- SET name = REPLACE(name, '&amp;', '&')
-- WHERE name LIKE '%&amp;%';

-- ── Repair: project names (run only after reviewing above) ─────────────────
-- UPDATE flow_projects
-- SET name = REPLACE(name, '&apos;', '''')
-- WHERE name LIKE '%&apos;%';
-- UPDATE flow_projects
-- SET name = REPLACE(name, '&#39;', '''')
-- WHERE name LIKE '%&#39;%';
-- UPDATE flow_projects
-- SET name = REPLACE(name, '&quot;', '"')
-- WHERE name LIKE '%&quot;%';
-- UPDATE flow_projects
-- SET name = REPLACE(REPLACE(name, '&lt;', '<'), '&gt;', '>')
-- WHERE name LIKE '%&lt;%' OR name LIKE '%&gt;%';
-- UPDATE flow_projects
-- SET name = REPLACE(name, '&amp;', '&')
-- WHERE name LIKE '%&amp;%';

-- ── Repair: task titles (run only after reviewing above) ──────────────────
-- UPDATE flow_tasks
-- SET title = REPLACE(title, '&apos;', '''')
-- WHERE title LIKE '%&apos;%';
-- UPDATE flow_tasks
-- SET title = REPLACE(title, '&#39;', '''')
-- WHERE title LIKE '%&#39;%';
-- UPDATE flow_tasks
-- SET title = REPLACE(title, '&quot;', '"')
-- WHERE title LIKE '%&quot;%';
-- UPDATE flow_tasks
-- SET title = REPLACE(REPLACE(title, '&lt;', '<'), '&gt;', '>')
-- WHERE title LIKE '%&lt;%' OR title LIKE '%&gt;%';
-- UPDATE flow_tasks
-- SET title = REPLACE(title, '&amp;', '&')
-- WHERE title LIKE '%&amp;%';

-- ── Verify: re-run the dry runs above. All three should return zero rows. ──
