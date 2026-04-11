import React from 'react';

interface NumericKeypadProps {
  value: string;
  onChange: (value: string) => void;
  onConfirm?: () => void;
  showDecimal?: boolean;
  label?: string;
  maxLength?: number;
}

const KEYS: (string | number)[][] = [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9],
  ['.', 0, '⌫'],
];

const NumericKeypad: React.FC<NumericKeypadProps> = ({
  value,
  onChange,
  onConfirm,
  showDecimal = true,
  label,
  maxLength = 12,
}) => {
  const push = (k: string | number) => {
    if (k === '⌫') {
      onChange(value.slice(0, -1));
      return;
    }
    if (k === '.') {
      if (!showDecimal) return;
      if (value.includes('.')) return;
      onChange((value || '0') + '.');
      return;
    }
    if (value.length >= maxLength) return;
    onChange(value + String(k));
  };

  return (
    <div className="flex flex-col gap-3 p-4 rounded-xl bg-pos-card border border-pos-border">
      {label && <p className="text-xs font-semibold text-pos-muted">{label}</p>}
      <div className="text-3xl font-bold text-pos-text tabular-nums min-h-[2.5rem] text-right">
        {value || '0'}
      </div>
      <div className="grid grid-cols-3 gap-2">
        {KEYS.flat().map((k, idx) => (
          <button
            key={idx}
            type="button"
            onClick={() => push(k)}
            disabled={k === '.' && !showDecimal}
            className="h-14 text-xl font-semibold rounded-lg bg-white/8 hover:bg-white/14 active:scale-95 transition text-pos-text disabled:opacity-40"
          >
            {k}
          </button>
        ))}
      </div>
      {onConfirm && (
        <button
          type="button"
          onClick={onConfirm}
          className="h-12 rounded-lg bg-pos-text text-pos-bg font-bold hover:opacity-90 transition"
        >
          Confirm
        </button>
      )}
    </div>
  );
};

export default NumericKeypad;
