-- cross-schema read from pos, finance, crm, erp, and hr — source timestamps version cached insights.
-- name: GetInsightsGeneration :one
-- A generation token changes whenever any source table used by Insights changes.
-- The token is read before constructing a cached key, so writes naturally
-- move readers to a new namespace without requiring every mutating module to
-- know the Insights cache layout.
SELECT CAST(GREATEST(
  COALESCE((SELECT MAX(o.updated_at) FROM pos_orders o WHERE o.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(r.refunded_at) FROM pos_refunds r WHERE r.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(i.updated_at) FROM finance_invoices i WHERE i.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(e.updated_at) FROM finance_expenses e WHERE e.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(c.updated_at) FROM crm_customers c WHERE c.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(p.updated_at) FROM erp_products p WHERE p.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(e.updated_at) FROM hr_employees e WHERE e.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(l.updated_at) FROM hr_leave_requests l WHERE l.business_id = ?), '1000-01-01'),
  COALESCE((SELECT MAX(a.updated_at) FROM hr_shift_assignments a WHERE a.business_id = ?), '1000-01-01')
) AS DATETIME) AS generation;
