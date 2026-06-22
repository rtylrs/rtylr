-- Adds indexed search support for staff lists.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE hr_staff ADD FULLTEXT INDEX idx_hr_staff_search (first_name, last_name, email, role);
