import React, { useState } from 'react';

interface PinPadProps {
  length?: number;
  onComplete: (pin: string) => void;
  label?: string;
  error?: string | null;
}

const PinPad: React.FC<PinPadProps> = ({ length = 4, onComplete, label, error }) => {
  const [pin, setPin] = useState('');

  const handleKey = (k: string) => {
    if (k === 'C') {
      setPin('');
      return;
    }
    if (k === '⌫') {
      setPin((p) => p.slice(0, -1));
      return;
    }
    if (pin.length >= length) return;
    const next = pin + k;
    setPin(next);
    if (next.length === length) {
      onComplete(next);
      // Clear after brief delay so consumer can act
      setTimeout(() => setPin(''), 400);
    }
  };

  const keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', 'C', '0', '⌫'];
  return (
    <div className="flex flex-col items-center gap-4 p-5 rounded-2xl bg-pos-card border border-pos-border">
      {label && <p className="text-sm font-semibold text-pos-text">{label}</p>}
      <div className="flex gap-2" aria-label="PIN entry">
        {Array.from({ length }).map((_, i) => (
          <div
            key={i}
            className={`w-4 h-4 rounded-full border border-pos-border ${
              i < pin.length ? 'bg-pos-text' : 'bg-transparent'
            }`}
          />
        ))}
      </div>
      {error && <p className="text-xs text-destructive">{error}</p>}
      <div className="grid grid-cols-3 gap-2 w-full max-w-xs">
        {keys.map((k) => (
          <button
            key={k}
            type="button"
            onClick={() => handleKey(k)}
            className="h-14 rounded-xl bg-white/10 text-pos-text text-lg font-semibold hover:bg-white/20 active:scale-95 transition"
          >
            {k}
          </button>
        ))}
      </div>
    </div>
  );
};

export default PinPad;
