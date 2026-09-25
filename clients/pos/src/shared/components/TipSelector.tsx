import React, { useState } from 'react';
import { formatCurrency } from '../utils/format';
import { useSettingsStore } from '../stores/settingsStore';

interface TipSelectorProps {
  subtotal: number;
  onTipChange: (tip: number) => void;
}

const TipSelector: React.FC<TipSelectorProps> = ({ subtotal, onTipChange }) => {
  const presets = useSettingsStore((s) => s.tipPresets);
  const [selected, setSelected] = useState<number | null>(null);
  const [custom, setCustom] = useState<string>('');

  const pick = (pct: number) => {
    setSelected(pct);
    setCustom('');
    onTipChange(subtotal * pct);
  };

  const pickCustom = (v: string) => {
    setCustom(v);
    setSelected(null);
    const n = parseFloat(v);
    onTipChange(!isNaN(n) && n >= 0 ? n : 0);
  };

  return (
    <div className="space-y-2">
      <p className="text-xs font-semibold text-pos-muted uppercase tracking-wider">Tip</p>
      <div className="grid grid-cols-4 gap-2">
        {presets.map((pct) => (
          <button
            key={pct}
            type="button"
            onClick={() => pick(pct)}
            className={`h-14 rounded-lg font-semibold text-sm transition ${
              selected === pct
                ? 'bg-pos-text text-pos-bg'
                : 'bg-white/8 text-pos-text hover:bg-white/14'
            }`}
          >
            <div>{Math.round(pct * 100)}%</div>
            <div className="text-[10px] opacity-70">{formatCurrency(subtotal * pct)}</div>
          </button>
        ))}
      </div>
      <div className="flex items-center gap-2">
        <label className="text-xs text-pos-muted">Custom</label>
        <input
          type="number"
          inputMode="decimal"
          value={custom}
          onChange={(e) => pickCustom(e.target.value)}
          placeholder="0.00"
          className="flex-1 h-11 rounded-lg bg-white/8 border border-pos-border px-3 text-sm text-pos-text outline-none focus:border-white/25"
        />
      </div>
    </div>
  );
};

export default TipSelector;
