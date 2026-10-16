-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

CREATE OR REPLACE VIEW ai_customers AS
SELECT id, owner_id, business_id, first_name, last_name, email, phone, city, country,
       loyalty_points, total_spent, visit_count, last_visit_at, loyalty_tier,
       is_active, created_at, updated_at
FROM crm_customers;
