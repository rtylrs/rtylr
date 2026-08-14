import React from 'react';
import { getCurrencySymbol, type Currency } from '../lib/money';
import { majorToMinor, minorToMajor } from '../utils/money';
import { useSettings } from '../lib/settings';
import { inputStyle } from '../lib/tokens';

interface CurrencyInputProps {
  value: number | string;
  onChange: (val: number) => void;
  currency?: Currency;
  placeholder?: string;
  disabled?: boolean;
  style?: React.CSSProperties;
}

const CurrencyInput: React.FC<CurrencyInputProps> = ({
  value,
  onChange,
  currency,
  placeholder = '0.00',
  disabled,
  style,
}) => {
  const settingsCurrency = useSettings((s) => s.currency);
  const c = currency ?? settingsCurrency;
  const symbol = getCurrencySymbol(c);

  const handle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/[^0-9.\-]/g, '');
    const n = parseFloat(raw);
    onChange(isNaN(n) ? 0 : majorToMinor(n, c));
  };

  return (
    <div style={{ position: 'relative', width: '100%' }}>
      <span
        style={{
          position: 'absolute',
          insetInlineStart: 12,
          top: '50%',
          transform: 'translateY(-50%)',
          fontSize: 13,
          color: 'var(--hub-text-dim)',
          pointerEvents: 'none',
          fontWeight: 600,
        }}
      >
        {symbol}
      </span>
      <input
        type="text"
        inputMode="decimal"
        value={value === 0 ? '' : String(minorToMajor(Number(value), c))}
        onChange={handle}
        placeholder={placeholder}
        disabled={disabled}
        style={{
          ...inputStyle,
          paddingInlineStart: 28,
          fontVariantNumeric: 'tabular-nums',
          ...style,
        }}
      />
    </div>
  );
};

export default CurrencyInput;
