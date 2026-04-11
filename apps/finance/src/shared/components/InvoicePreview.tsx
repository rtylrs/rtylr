import React from 'react';
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
}

const InvoicePreview: React.FC<InvoicePreviewProps> = ({
  invoiceNumber,
  client,
  issuedDate,
  dueDate,
  items,
  notes,
  discount = 0,
}) => {
  const settings = useSettings();
  const subtotal = items.reduce((s, it) => s + it.quantity * it.unitPrice, 0);
  const tax = items.reduce((s, it) => s + it.quantity * it.unitPrice * (it.taxRate / 100), 0);
  const total = subtotal + tax - discount;

  return (
    <div
      style={{
        background: '#fff',
        color: '#111',
        borderRadius: 12,
        padding: 32,
        fontFamily: 'system-ui, sans-serif',
        boxShadow: '0 8px 30px rgba(0,0,0,0.15)',
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
          <div style={{ fontSize: 11, color: '#666' }}>{settings.companyEmail}</div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: 24, fontWeight: 800, letterSpacing: '0.04em', color: '#111' }}>
            INVOICE
          </div>
          <div style={{ fontSize: 12, color: '#666' }}>#{invoiceNumber}</div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 24 }}>
        <div>
          <div
            style={{
              fontSize: 10,
              textTransform: 'uppercase',
              color: '#888',
              fontWeight: 600,
              marginBottom: 4,
            }}
          >
            Bill To
          </div>
          <div style={{ fontSize: 13, fontWeight: 600 }}>
            {client?.name ?? 'No client selected'}
          </div>
          <div style={{ fontSize: 11, color: '#666' }}>{client?.email}</div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: 11, color: '#666' }}>
            Issued: <strong>{issuedDate || '-'}</strong>
          </div>
          <div style={{ fontSize: 11, color: '#666' }}>
            Due: <strong>{dueDate || '-'}</strong>
          </div>
        </div>
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: 20 }}>
        <thead>
          <tr style={{ borderBottom: `2px solid ${settings.brandColor}` }}>
            <th
              style={{
                textAlign: 'left',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: '#666',
              }}
            >
              Description
            </th>
            <th
              style={{
                textAlign: 'right',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: '#666',
              }}
            >
              Qty
            </th>
            <th
              style={{
                textAlign: 'right',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: '#666',
              }}
            >
              Price
            </th>
            <th
              style={{
                textAlign: 'right',
                padding: '8px 4px',
                fontSize: 10,
                textTransform: 'uppercase',
                color: '#666',
              }}
            >
              Total
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((it) => (
            <tr key={it.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={{ padding: '10px 4px', fontSize: 12 }}>{it.description || 'Item'}</td>
              <td style={{ padding: '10px 4px', fontSize: 12, textAlign: 'right' }}>
                {it.quantity}
              </td>
              <td style={{ padding: '10px 4px', fontSize: 12, textAlign: 'right' }}>
                <MoneyDisplay amount={it.unitPrice} />
              </td>
              <td
                style={{ padding: '10px 4px', fontSize: 12, textAlign: 'right', fontWeight: 600 }}
              >
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
            <span style={{ color: '#666' }}>Subtotal</span>
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
            <span style={{ color: '#666' }}>Tax</span>
            <MoneyDisplay amount={tax} />
          </div>
          {discount > 0 && (
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '4px 0',
                fontSize: 12,
                color: '#a85',
              }}
            >
              <span>Discount</span>
              <span>
                -<MoneyDisplay amount={discount} />
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
            <span>Total</span>
            <MoneyDisplay amount={total} />
          </div>
        </div>
      </div>

      {notes && (
        <div
          style={{
            marginTop: 24,
            padding: 12,
            background: '#f7f7f7',
            borderRadius: 8,
            fontSize: 11,
            color: '#555',
          }}
        >
          <strong>Notes:</strong> {notes}
        </div>
      )}
    </div>
  );
};

export default InvoicePreview;
