import React from 'react';

interface QuantityStepperProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  size?: 'sm' | 'md' | 'lg';
}

const QuantityStepper: React.FC<QuantityStepperProps> = ({
  value,
  onChange,
  min = 0,
  max = 999,
  size = 'md',
}) => {
  const cls =
    size === 'lg' ? 'w-11 h-11 text-lg' : size === 'sm' ? 'w-7 h-7 text-xs' : 'w-9 h-9 text-sm';
  const dec = () => onChange(Math.max(min, value - 1));
  const inc = () => onChange(Math.min(max, value + 1));
  return (
    <div className="inline-flex items-center gap-1">
      <button
        onClick={dec}
        className={`${cls} flex items-center justify-center rounded-lg bg-white/10 text-pos-text hover:bg-white/20 active:scale-95 transition`}
        aria-label="Decrease"
      >
        −
      </button>
      <span className="min-w-[2ch] text-center text-sm font-semibold text-pos-text tabular-nums">
        {value}
      </span>
      <button
        onClick={inc}
        className={`${cls} flex items-center justify-center rounded-lg bg-white/10 text-pos-text hover:bg-white/20 active:scale-95 transition`}
        aria-label="Increase"
      >
        +
      </button>
    </div>
  );
};

export default QuantityStepper;
