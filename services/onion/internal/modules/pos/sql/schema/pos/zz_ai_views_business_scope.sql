-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

CREATE OR REPLACE VIEW ai_pos_orders AS
SELECT id, owner_id, business_id, order_number, type, status, customer_id, customer_name,
       staff_id, location_id, subtotal, discount_amount, tax_amount, total,
       tip_amount, payment_method, payment_status, created_at, paid_at
FROM pos_orders;
