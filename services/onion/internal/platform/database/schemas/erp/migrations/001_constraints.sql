-- ERP integrity + performance migration (2026-06).
-- Adds indexes on owner_id / foreign-key columns, unique constraints on
-- (owner_id, sku) / (owner_id, barcode) / (owner_id, po_number), and the
-- foreign keys that were missing from the original ERP schema. Applied
-- directly to the live DB (near-empty, pre-client). Kept here (outside the
-- sqlc-scanned schema dir) as the reproducible record of what was applied.

-- ── Indexes ──────────────────────────────────────────────────────────────
ALTER TABLE erp_categories             ADD INDEX idx_erp_categories_owner (owner_id);
ALTER TABLE erp_suppliers              ADD INDEX idx_erp_suppliers_owner (owner_id);
ALTER TABLE erp_products               ADD INDEX idx_erp_products_owner (owner_id),
                                       ADD INDEX idx_erp_products_category (category_id),
                                       ADD INDEX idx_erp_products_supplier (supplier_id);
ALTER TABLE erp_inventory_transactions ADD INDEX idx_erp_invtx_owner (owner_id),
                                       ADD INDEX idx_erp_invtx_product (product_id),
                                       ADD INDEX idx_erp_invtx_ref (reference_type, reference_id);
ALTER TABLE erp_purchase_orders        ADD INDEX idx_erp_po_owner (owner_id),
                                       ADD INDEX idx_erp_po_supplier (supplier_id);
ALTER TABLE erp_purchase_order_items   ADD INDEX idx_erp_poi_po (purchase_order_id),
                                       ADD INDEX idx_erp_poi_product (product_id);
ALTER TABLE erp_product_variants       ADD INDEX idx_erp_variants_owner (owner_id),
                                       ADD INDEX idx_erp_variants_product (product_id);
ALTER TABLE erp_warehouses             ADD INDEX idx_erp_warehouses_owner (owner_id);
ALTER TABLE erp_brands                 ADD INDEX idx_erp_brands_owner (owner_id);
ALTER TABLE erp_stock_takes            ADD INDEX idx_erp_st_owner (owner_id),
                                       ADD INDEX idx_erp_st_warehouse (warehouse_id);
ALTER TABLE erp_stock_take_items       ADD INDEX idx_erp_sti_take (stock_take_id),
                                       ADD INDEX idx_erp_sti_product (product_id);
ALTER TABLE erp_stock_transfers        ADD INDEX idx_erp_str_owner (owner_id),
                                       ADD INDEX idx_erp_str_from (from_warehouse_id),
                                       ADD INDEX idx_erp_str_to (to_warehouse_id);
ALTER TABLE erp_stock_transfer_items   ADD INDEX idx_erp_stri_transfer (transfer_id),
                                       ADD INDEX idx_erp_stri_product (product_id);

-- ── Unique constraints (per tenant; NULLs are allowed/distinct in MySQL) ──
ALTER TABLE erp_products        ADD UNIQUE KEY uq_erp_products_owner_sku (owner_id, sku),
                                ADD UNIQUE KEY uq_erp_products_owner_barcode (owner_id, barcode);
ALTER TABLE erp_purchase_orders ADD UNIQUE KEY uq_erp_po_owner_number (owner_id, po_number);

-- ── Foreign keys ─────────────────────────────────────────────────────────
ALTER TABLE erp_products
  ADD CONSTRAINT fk_erp_products_category FOREIGN KEY (category_id) REFERENCES erp_categories(id) ON DELETE SET NULL,
  ADD CONSTRAINT fk_erp_products_supplier FOREIGN KEY (supplier_id) REFERENCES erp_suppliers(id) ON DELETE SET NULL;
ALTER TABLE erp_purchase_orders
  ADD CONSTRAINT fk_erp_po_supplier FOREIGN KEY (supplier_id) REFERENCES erp_suppliers(id);
ALTER TABLE erp_purchase_order_items
  ADD CONSTRAINT fk_erp_poi_po FOREIGN KEY (purchase_order_id) REFERENCES erp_purchase_orders(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_erp_poi_product FOREIGN KEY (product_id) REFERENCES erp_products(id);
ALTER TABLE erp_inventory_transactions
  ADD CONSTRAINT fk_erp_invtx_product FOREIGN KEY (product_id) REFERENCES erp_products(id);
ALTER TABLE erp_product_variants
  ADD CONSTRAINT fk_erp_variants_product FOREIGN KEY (product_id) REFERENCES erp_products(id) ON DELETE CASCADE;
ALTER TABLE erp_stock_takes
  ADD CONSTRAINT fk_erp_st_warehouse FOREIGN KEY (warehouse_id) REFERENCES erp_warehouses(id) ON DELETE SET NULL;
ALTER TABLE erp_stock_take_items
  ADD CONSTRAINT fk_erp_sti_take FOREIGN KEY (stock_take_id) REFERENCES erp_stock_takes(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_erp_sti_product FOREIGN KEY (product_id) REFERENCES erp_products(id);
ALTER TABLE erp_stock_transfers
  ADD CONSTRAINT fk_erp_str_from FOREIGN KEY (from_warehouse_id) REFERENCES erp_warehouses(id),
  ADD CONSTRAINT fk_erp_str_to FOREIGN KEY (to_warehouse_id) REFERENCES erp_warehouses(id);
ALTER TABLE erp_stock_transfer_items
  ADD CONSTRAINT fk_erp_stri_transfer FOREIGN KEY (transfer_id) REFERENCES erp_stock_transfers(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_erp_stri_product FOREIGN KEY (product_id) REFERENCES erp_products(id);
