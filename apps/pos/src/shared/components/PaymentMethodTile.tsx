import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { Banknote, CreditCard, Ellipsis, Gift } from 'lucide-react';
import type { PaymentMethod } from '../types/pos';

interface PaymentMethodTileProps {
  method: PaymentMethod;
  active?: boolean;
  onClick?: () => void;
  label?: string;
}

const ICON: Record<PaymentMethod, LucideIcon> = {
  cash: Banknote,
  card: CreditCard,
  gift_card: Gift,
  other: Ellipsis,
};

const LABEL: Record<PaymentMethod, string> = {
  cash: 'Cash',
  card: 'Card',
  gift_card: 'Gift Card',
  other: 'Other',
};

const PaymentMethodTile: React.FC<PaymentMethodTileProps> = ({
  method,
  active,
  onClick,
  label,
}) => {
  const Icon = ICON[method];
  return (
    <button
      onClick={onClick}
      type="button"
      className={`flex flex-col items-center justify-center gap-2 p-5 rounded-xl border-2 transition min-h-[7rem] ${
        active
          ? 'bg-pos-text text-pos-bg border-pos-text'
          : 'bg-pos-card border-pos-border text-pos-text hover:border-white/25'
      }`}
    >
      <Icon className="h-7 w-7" aria-hidden="true" strokeWidth={2} />
      <span className="text-sm font-semibold">{label ?? LABEL[method]}</span>
    </button>
  );
};

export default PaymentMethodTile;
