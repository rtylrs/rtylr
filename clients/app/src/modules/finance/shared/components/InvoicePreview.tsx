import React from 'react';
import { useTranslation } from 'react-i18next';
import MoneyDisplay from './MoneyDisplay';
import type { LineItem } from './LineItemRow';
import type { Client } from './ClientSelect';
import { useSettings } from '../lib/settings';

interface InvoicePreviewProps {
  invoiceNumber: string;
  client: Client | null;
  issuedDate: string;
  dueDate: string;
  items: LineItem[];
  notes?: string;
  discount?: number;
  discountType?: string | null;
  discountAmount?: number | null;
  subtotal?: number | null;
  tax?: number | null;
  total?: number | null;
}

const InvoicePreview: React.FC<InvoicePreviewProps> = ({
  invoiceNumber,
  client,
  issuedDate,
  dueDate,
  items,
  notes,
  discount = 0,
  discountType,
  discountAmount,
  subtotal: persistedSubtotal,
  tax: persistedTax,
  total: persistedTotal,
}) => {
  const { t } = useTranslation('financeCommon');
  const settings = useSettings();
  const calculatedSubtotal = items.reduce((s, it) => s + it.quantity * it.unitPrice, 0);
  const calculatedTax = items.reduce(
    (s, it) => s + it.quantity * it.unitPrice * (it.taxRate / 100),
    0
  );
  const subtotal = persistedSubtotal ?? calculatedSubtotal;
  const tax = persistedTax ?? calculatedTax;
  const calculatedDiscount =
    discountType === 'percentage'
      ? (subtotal * Math.max(discount, 0)) / 100
      : Math.max(discount, 0);
  const appliedDiscount = Math.max(discountAmount ?? calculatedDiscount, 0);
  const total = persistedTotal ?? subtotal + tax - appliedDiscount;

  return (
    <div
      style={{
        background: 'var(--finance-invoice-surface)',
        color: 'var(--finance-invoice-text)',
        borderRadius: 12,
        padding: 32,
        fontFamily: 'system-ui, sans-serif',
        boxShadow: '0 8px 30px var(--finance-overlay-subtle)',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          marginBottom: 28,
        }}
      >
        <div>
          <div style={{ fontSize: 20, fontWeight: 800, color: settings.brandColor }}>
            {settings.companyName}
          </div>
          <div style={{ fontSize: 11, color: 'var(--finance-invoice-muted)' }}>
            {settings.companyEmail}
          </div>
        </div>
        <div style={{ textAlign: 'end' }}>
          <div
            style={{
              fontSize: 24,
              fontWeight: 800,
              letterSpacing: '0.04em',
              color: 'var(--finance-invoice-text)',
            }}
          >
            {t('financeInvoices.preview.heading')}
          </div>
          <div style={{ fontSize: 12, color: 'var(--finance-invoice-muted)' }}>
            #{invoiceNumber}
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 24 }}>
        <div>
          <div
            style={{
              fontSize: 10,
              textTransform: 'uppercase',
              color: 'var(--finance-invoice-label)',
              fontWeight: 600,
              marginBottom: 4,
            }}
          >
            {t('financeInvoices.preview.billTo')}
          </div>
          <div style={{ fontSize: 13, fontWeight: 600 }}>
            {client?.name ?? 'No client selected'}
          </div>
          <div style={{ fontSize: 11, color: 'var(--finance-invoice-muted)' }}>{client?.email}</div>
        </div>
        <div style={{ textAlign: 'end' }}>
          <div style={{ fontSize: 11, color: 'var(--finance-invoice-muted)' }}>
            Issued: <strong>{issuedDate || '-'}</strong>
          </div>
          <div style={{ fontSize: 11, color: 'var(--finance-invoice-muted)' }}>
            Due: <strong>{dueDate || '-'}</strong>
          </div>
        </div>
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: 20 }}>
        <thead>
          <tr style={{ borderBottom: `2px solid ${settings.brandColor}` }}>
            <th
              style={{
                textAlign: 'start',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: 'var(--finance-invoice-muted)',
              }}
            >
              {t('financeInvoices.preview.description')}
            </th>
            <th
              style={{
                textAlign: 'end',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: 'var(--finance-invoice-muted)',
              }}
            >
              {t('financeInvoices.preview.qty')}
            </th>
            <th
              style={{
                textAlign: 'end',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: 'var(--finance-invoice-muted)',
              }}
            >
              {t('financeInvoices.preview.price')}
            </th>
            <th
              style={{
                textAlign: 'end',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: 'var(--finance-invoice-muted)',
              }}
            >
              {t('financeInvoices.preview.total')}
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((it) => (
            <tr key={it.id} style={{ borderBottom: '1px solid var(--finance-invoice-border)' }}>
              <td style={{ padding: '10px 4px', fontSize: 12 }}>{it.description || 'Item'}</td>
              <td style={{ padding: '10px 4px', fontSize: 12, textAlign: 'end' }}>{it.quantity}</td>
              <td style={{ padding: '10px 4px', fontSize: 12, textAlign: 'end' }}>
                <MoneyDisplay amount={it.unitPrice} />
              </td>
              <td style={{ padding: '10px 4px', fontSize: 12, textAlign: 'end', fontWeight: 600 }}>
                <MoneyDisplay amount={it.quantity * it.unitPrice} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <div style={{ width: 260 }}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '4px 0',
              fontSize: 12,
            }}
          >
            <span style={{ color: 'var(--finance-invoice-muted)' }}>
              {t('financeInvoices.preview.subtotal')}
            </span>
            <MoneyDisplay amount={subtotal} />
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '4px 0',
              fontSize: 12,
            }}
          >
            <span style={{ color: 'var(--finance-invoice-muted)' }}>
              {t('financeInvoices.preview.tax')}
            </span>
            <MoneyDisplay amount={tax} />
          </div>
          {appliedDiscount > 0 && (
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '4px 0',
                fontSize: 12,
                color: 'var(--finance-invoice-tax)',
              }}
            >
              <span>{t('financeInvoices.preview.discount')}</span>
              <span>
                -<MoneyDisplay amount={appliedDiscount} />
              </span>
            </div>
          )}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '10px 0 4px',
              borderTop: `2px solid ${settings.brandColor}`,
              fontSize: 15,
              fontWeight: 800,
              marginTop: 6,
            }}
          >
            <span>{t('financeInvoices.preview.total')}</span>
            <MoneyDisplay amount={total} />
          </div>
        </div>
      </div>

      {notes && (
        <div
          style={{
            marginTop: 24,
            padding: 12,
            background: 'var(--finance-invoice-footer)',
            borderRadius: 8,
            fontSize: 11,
            color: 'var(--finance-invoice-footer-text)',
          }}
        >
          <strong>{t('financeInvoices.preview.notes')}</strong> {notes}
        </div>
      )}
    </div>
  );
};

export default InvoicePreview;
