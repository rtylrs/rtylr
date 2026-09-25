import React from 'react';

export const UNITS = [
  { value: 'kg', label: 'Kilograms (kg)' },
  { value: 'g', label: 'Grams (g)' },
  { value: 'L', label: 'Litres (L)' },
  { value: 'mL', label: 'Millilitres (mL)' },
  { value: 'pcs', label: 'Pieces (pcs)' },
  { value: 'boxes', label: 'Boxes' },
  { value: 'bags', label: 'Bags' },
  { value: 'cases', label: 'Cases' },
] as const;

interface UnitSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

export function convertUnits(value: number, from: string, to: string): number {
  const rates: Record<string, number> = { kg: 1000, g: 1, L: 1000, mL: 1 };
  if (rates[from] && rates[to]) {
    return (value * rates[from]) / rates[to];
  }
  return value;
}

const UnitSelector: React.FC<UnitSelectorProps> = ({ value, onChange }) => {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="px-3 py-2 rounded-lg text-sm outline-none"
      style={{
        background: 'var(--hub-input-bg)',
        border: '1px solid var(--hub-border)',
        color: 'var(--hub-text)',
      }}
    >
      {UNITS.map((u) => (
        <option key={u.value} value={u.value}>
          {u.label}
        </option>
      ))}
    </select>
  );
};

export default UnitSelector;
