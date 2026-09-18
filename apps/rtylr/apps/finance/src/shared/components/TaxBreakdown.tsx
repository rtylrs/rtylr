import React from 'react';
import { useTranslation } from 'react-i18next';
import MoneyDisplay from './MoneyDisplay';
import type { LineItem } from './LineItemRow';

interface TaxBreakdownProps {
  items: LineItem[];
  discount?: number;
}

interface TaxBucket {
  rate: number;
  taxable: number;
  tax: number;
}

function bucketize(items: LineItem[]): TaxBucket[] {
  const map = new Map<number, TaxBucket>();
  for (const it of items) {
    const sub = it.quantity * it.unitPrice;
    const b = map.get(it.taxRate) ?? { rate: it.taxRate, taxable: 0, tax: 0 };
    b.taxable += sub;
    b.tax += sub * (it.taxRate / 100);
    map.set(it.taxRate, b);
  }
  return Array.from(map.values()).sort((a, b) => a.rate - b.rate);
}

const TaxBreakdown: React.FC<TaxBreakdownProps> = ({ items, discount = 0 }) => {
  const { t } = useTranslation();
  const buckets = bucketize(items);
  const subtotal = items.reduce((s, it) => s + it.quantity * it.unitPrice, 0);
  const totalTax = buckets.reduce((s, b) => s + b.tax, 0);
  const total = subtotal + totalTax - discount;

  return (
    <div
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: 10,
        padding: 14,
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: 13,
          color: 'var(--hub-text-muted)',
        }}
      >
        <span>{t('common.subtotal')}</span>
        <MoneyDisplay amount={subtotal} />
      </div>
      {buckets.map((b) => (
        <div
          key={b.rate}
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            fontSize: 12,
            color: 'var(--hub-text-dim)',
          }}
        >
          <span>
            Tax ({b.rate.toFixed(1)}% of <MoneyDisplay amount={b.taxable} />)
          </span>
          <MoneyDisplay amount={b.tax} />
        </div>
      ))}
      {discount > 0 && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            fontSize: 13,
            color: 'var(--finance-warning)',
          }}
        >
          <span>{t('common.discount')}</span>
          <span>
            -<MoneyDisplay amount={discount} />
          </span>
        </div>
      )}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          paddingTop: 8,
          borderTop: '1px solid var(--hub-border)',
          fontSize: 15,
          fontWeight: 700,
          color: 'var(--hub-text)',
        }}
      >
        <span>{t('common.total')}</span>
        <MoneyDisplay amount={total} />
      </div>
    </div>
  );
};

export default TaxBreakdown;
