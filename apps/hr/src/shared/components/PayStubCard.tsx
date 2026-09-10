import React from 'react';
import { useTranslation } from 'react-i18next';
import { DollarSign } from 'lucide-react';
import { PayrollRow } from '../services/hrApi';
import { Card, Pill, ACCENT } from './ui';
import { initials, paletteFor } from './EmployeeCard';
import { formatCurrency } from '../lib/hrNumberUtils';

interface Props {
  row: PayrollRow;
  onPay?: (id: number) => void;
  currency?: string;
}

const PayStubCard: React.FC<Props> = ({ row, onPay, currency }) => {
  const { t } = useTranslation();
  const { bg, fg } = paletteFor(row.id);
  const net = row.basePay + row.overtime - row.deductions;
  return (
    <Card style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span
          style={{
            display: 'flex',
            width: 36,
            height: 36,
            borderRadius: 10,
            background: bg,
            color: fg,
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 700,
            fontSize: 12,
          }}
        >
          {initials(row.name)}
        </span>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>{row.name}</div>
          <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>{row.role}</div>
        </div>
        <Pill color={row.status === 'Paid' ? '#10b981' : '#f59e0b'}>{row.status ?? 'Pending'}</Pill>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 11 }}>
        <div>
          <div style={{ color: 'var(--hr-text-muted)' }}>{t('payroll.stub.hours')}</div>
          <div style={{ fontWeight: 600, color: 'var(--hr-text)' }}>{row.hoursWorked}h</div>
        </div>
        <div>
          <div style={{ color: 'var(--hr-text-muted)' }}>{t('payroll.stub.base')}</div>
          <div style={{ fontWeight: 600, color: 'var(--hr-text)' }}>
            {formatCurrency(row.basePay, currency)}
          </div>
        </div>
        <div>
          <div style={{ color: 'var(--hr-text-muted)' }}>{t('payroll.stub.overtime')}</div>
          <div style={{ fontWeight: 600, color: 'var(--hr-text)' }}>
            {formatCurrency(row.overtime, currency)}
          </div>
        </div>
        <div>
          <div style={{ color: 'var(--hr-text-muted)' }}>{t('payroll.stub.deductions')}</div>
          <div style={{ fontWeight: 600, color: '#ef4444' }}>
            {formatCurrency(row.deductions, currency)}
          </div>
        </div>
      </div>

      <div
        style={{
          borderTop: '1px solid var(--hr-border)',
          paddingTop: 8,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div>
          <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>
            {t('payroll.stub.netPay')}
          </div>
          <div style={{ fontSize: 15, fontWeight: 700, color: ACCENT }}>
            {formatCurrency(net, currency)}
          </div>
        </div>
        {row.status !== 'Paid' && onPay && (
          <button
            type="button"
            onClick={() => onPay(row.id)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 10px',
              borderRadius: 8,
              border: `1px solid ${ACCENT}55`,
              background: `${ACCENT}22`,
              color: ACCENT,
              fontSize: 11,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <DollarSign size={12} /> Pay
          </button>
        )}
      </div>
    </Card>
  );
};

export default PayStubCard;
