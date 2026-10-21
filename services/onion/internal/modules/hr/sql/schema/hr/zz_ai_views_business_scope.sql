-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

CREATE OR REPLACE VIEW ai_employees AS
SELECT id, owner_id, business_id, first_name, last_name, email, department_id, location_id,
       employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, created_at, updated_at
FROM hr_employees;
