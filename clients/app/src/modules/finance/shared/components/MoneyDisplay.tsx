import React from 'react';
import { formatMoney, type Currency } from '../lib/money';
import { useSettings } from '../lib/settings';

interface MoneyDisplayProps {
  amount: number | string | null | undefined;
  currency?: Currency;
  fractionDigits?: number;
  className?: string;
  style?: React.CSSProperties;
  signed?: boolean;
}

const MoneyDisplay: React.FC<MoneyDisplayProps> = ({
  amount,
  currency,
  fractionDigits = 2,
  className,
  style,
  signed = false,
}) => {
  const settingsCurrency = useSettings((s) => s.currency);
  const c = currency ?? settingsCurrency;
  const n = typeof amount === 'string' ? parseFloat(amount) : (amount ?? 0);
  const text = formatMoney(n, c, fractionDigits);
  const sign = signed && (n as number) > 0 ? '+' : '';
  return (
    <span className={className} style={{ fontVariantNumeric: 'tabular-nums', ...style }}>
      {sign}
      {text}
    </span>
  );
};

export default MoneyDisplay;
