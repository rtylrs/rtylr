-- sqlc stub for product_stock_total view (see erp/migrations/002_unified_stock.sql).
CREATE TABLE product_stock_total (
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    total_quantity DECIMAL(14,3) NOT NULL
);
