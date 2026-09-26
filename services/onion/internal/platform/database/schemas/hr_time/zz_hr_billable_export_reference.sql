
  

UPDATE hr_time_entries
SET export_batch_id = billable_invoice_id,
    billable_invoice_id = NULL
WHERE billable_invoice_id >= 1000000000000
  AND export_batch_id IS NULL;
