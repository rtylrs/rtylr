import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { AlertTriangle } from 'lucide-react';
import MoneyDisplay from './MoneyDisplay';
import { cardStyle, ACCENT } from '../lib/tokens';

export interface OverdueRow {
  id: string;
  client: string;
  amount: number | string;
  daysOverdue: number;
}

interface OverdueInvoicesWidgetProps {
  items: OverdueRow[];
}

const OverdueInvoicesWidget: React.FC<OverdueInvoicesWidgetProps> = ({ items }) => {
  const { t } = useTranslation();
  return (
    <div style={cardStyle}>
      <div style={{ padding: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
          <AlertTriangle size={14} style={{ color: 'var(--finance-danger)' }} />
          <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
            {t('finance.overdue.title')}
          </h3>
          <span
            style={{
              fontSize: 10,
              padding: '2px 8px',
              borderRadius: 10,
              background: 'var(--finance-danger-soft)',
              color: 'var(--finance-danger)',
              marginInlineStart: 'auto',
              fontWeight: 600,
            }}
          >
            {items.length}
          </span>
        </div>
        {items.length === 0 ? (
          <div
            style={{
              padding: 20,
              textAlign: 'center',
              color: 'var(--hub-text-muted)',
              fontSize: 12,
            }}
          >
            {t('finance.overdue.none')}
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {items.slice(0, 6).map((item) => (
              <Link
                key={item.id}
                to="/invoices/$id"
                params={{ id: String(item.id) }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '8px 10px',
                  borderRadius: 8,
                  background: 'var(--hub-hover)',
                  textDecoration: 'none',
                  color: 'var(--hub-text)',
                }}
              >
                <div>
                  <div style={{ fontSize: 12, fontWeight: 600 }}>
                    {item.id} <>&middot;</> {item.client}
                  </div>
                  <div style={{ fontSize: 10, color: 'var(--finance-danger)' }}>
                    {item.daysOverdue}d overdue
                  </div>
                </div>
                <span style={{ fontSize: 13, fontWeight: 700, color: ACCENT }}>
                  {typeof item.amount === 'number' ? (
                    <MoneyDisplay amount={item.amount} />
                  ) : (
                    item.amount
                  )}
                </span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default OverdueInvoicesWidget;
