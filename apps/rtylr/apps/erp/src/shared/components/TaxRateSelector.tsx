import React from 'react';
import { useSettings } from '../lib/settingsStore';
import { useErpSettings } from '../hooks/queries/useErpSettings';
import { formatCurrency } from '../lib/format';

interface TaxRateSelectorProps {
  price: number;
  rate?: number;
  onRateChange?: (rate: number) => void;
  inclusive?: boolean;
  onInclusiveToggle?: (v: boolean) => void;
}

const TAX_RATES = [0, 5, 8.25, 10, 15, 20];

const TaxRateSelector: React.FC<TaxRateSelectorProps> = ({
  price,
  rate,
  onRateChange,
  inclusive = false,
  onInclusiveToggle,
}) => {
  const localDefaultRate = useSettings((s) => s.taxRate);
  const { data: serverSettings } = useErpSettings();
  const defaultRate = serverSettings?.taxRate ?? localDefaultRate;
  const effectiveRate = rate ?? defaultRate;

  const netPrice = inclusive ? price / (1 + effectiveRate / 100) : price;
  const taxAmount = netPrice * (effectiveRate / 100);
  const grossPrice = netPrice + taxAmount;

  return (
    <div
      className="rounded-lg p-3 space-y-2"
      style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
    >
      <div className="flex items-center gap-2">
        <label className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Tax Rate
        </label>
        <select
          value={effectiveRate}
          onChange={(e) => onRateChange?.(parseFloat(e.target.value))}
          className="flex-1 px-2 py-1 rounded-md text-xs outline-none"
          style={{
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
          }}
        >
          {TAX_RATES.map((r) => (
            <option key={r} value={r}>
              {r}%
            </option>
          ))}
        </select>
      </div>
      {onInclusiveToggle && (
        <label
          className="flex items-center gap-2 text-xs cursor-pointer"
          style={{ color: 'var(--hub-text-muted)' }}
        >
          <input
            type="checkbox"
            checked={inclusive}
            onChange={(e) => onInclusiveToggle(e.target.checked)}
          />
          Tax inclusive
        </label>
      )}
      <div className="space-y-1 pt-2" style={{ borderTop: '1px solid var(--hub-border)' }}>
        <div className="flex justify-between text-xs">
          <span style={{ color: 'var(--hub-text-muted)' }}>Net</span>
          <span style={{ color: 'var(--hub-text)' }}>{formatCurrency(netPrice)}</span>
        </div>
        <div className="flex justify-between text-xs">
          <span style={{ color: 'var(--hub-text-muted)' }}>Tax</span>
          <span style={{ color: 'var(--hub-text)' }}>{formatCurrency(taxAmount)}</span>
        </div>
        <div className="flex justify-between text-sm font-semibold">
          <span style={{ color: 'var(--hub-text)' }}>Total</span>
          <span style={{ color: 'var(--erp-accent)' }}>{formatCurrency(grossPrice)}</span>
        </div>
      </div>
    </div>
  );
};

export default TaxRateSelector;
