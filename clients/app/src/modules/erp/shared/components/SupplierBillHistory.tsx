import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronRight, Receipt } from 'lucide-react';

import { useSupplierBills } from '../hooks/queries/useSupplierBills';

interface SupplierBillHistoryProps {
  supplier: { id?: string };
}

/** FIN->ERP-01 (issue 844): bill history panel for one supplier. Hidden when
 *  Finance is absent or unreachable; never shows a zero-bill claim. */
const SupplierBillHistory: React.FC<SupplierBillHistoryProps> = ({ supplier }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const crmSupplierId = supplier.id;
  const { data: bills } = useSupplierBills(crmSupplierId);

  if (!bills || bills.length === 0) return null;

  return (
    <div style={{ marginTop: 8 }}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 4,
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--hub-text-muted)',
          fontSize: 12,
          padding: '2px 0',
        }}
      >
        {open ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
        <Receipt size={13} />
        {t('erp.supplierBills', { count: bills.length })}
      </button>
      {open && (
        <div style={{ paddingLeft: 20, paddingTop: 4 }}>
          {bills.slice(0, 10).map((bill) => (
            <div
              key={bill.id}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 12,
                padding: '3px 0',
              }}
            >
              <span
                style={{
                  fontWeight: bill.status === 'unpaid' ? 600 : 400,
                  color: bill.status === 'overdue' ? 'var(--erp-danger)' : undefined,
                }}
              >
                {bill.status}
              </span>
              <span>
                {bill.currency ?? ''}
                {(bill.amount ?? 0).toFixed(2)}
              </span>
            </div>
          ))}
          {bills.length > 10 && (
            <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>+{bills.length - 10}</div>
          )}
        </div>
      )}
    </div>
  );
};

export default SupplierBillHistory;
