-- Backfill blank names and employment types on auto-provisioned hr_employees
-- and hr_staff rows from the linked users account. Only touches rows whose
-- names or employment_type are blank, so re-running is a no-op.

UPDATE hr_employees AS e
JOIN users AS u ON u.id = e.user_id
SET
    e.first_name = CASE
        WHEN TRIM(COALESCE(u.full_name, '')) <> '' THEN SUBSTRING_INDEX(TRIM(u.full_name), ' ', 1)
        ELSE SUBSTRING_INDEX(u.email, '@', 1)
    END,
    e.last_name = CASE
        WHEN LOCATE(' ', TRIM(COALESCE(u.full_name, ''))) > 0
            THEN TRIM(SUBSTRING(TRIM(u.full_name), LOCATE(' ', TRIM(u.full_name)) + 1))
        ELSE ''
    END,
    e.email = CASE
        WHEN TRIM(COALESCE(e.email, '')) = '' THEN u.email
        ELSE e.email
    END
WHERE e.user_id IS NOT NULL
  AND TRIM(COALESCE(e.first_name, '')) = ''
  AND TRIM(COALESCE(e.last_name, '')) = '';

UPDATE hr_employees
SET employment_type = 'full_time'
WHERE TRIM(COALESCE(employment_type, '')) = '';

UPDATE hr_staff AS s
JOIN users AS u ON u.id = s.user_id
SET
    s.first_name = CASE
        WHEN TRIM(COALESCE(u.full_name, '')) <> '' THEN SUBSTRING_INDEX(TRIM(u.full_name), ' ', 1)
        ELSE SUBSTRING_INDEX(u.email, '@', 1)
    END,
    s.last_name = CASE
        WHEN LOCATE(' ', TRIM(COALESCE(u.full_name, ''))) > 0
            THEN TRIM(SUBSTRING(TRIM(u.full_name), LOCATE(' ', TRIM(u.full_name)) + 1))
        ELSE ''
    END,
    s.email = CASE
        WHEN TRIM(COALESCE(s.email, '')) = '' THEN u.email
        ELSE s.email
    END
WHERE s.user_id IS NOT NULL
  AND TRIM(COALESCE(s.first_name, '')) = ''
  AND TRIM(COALESCE(s.last_name, '')) = '';
