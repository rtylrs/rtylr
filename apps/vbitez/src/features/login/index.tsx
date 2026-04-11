import React, { useState, useEffect } from 'react';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useThemeStore } from '../../shared/stores/themeStore';

const PIN_LENGTH = 4;

const KEYS = [
  ['1', '2', '3'],
  ['4', '5', '6'],
  ['7', '8', '9'],
  ['', '0', '⌫'],
];

const PinLogin: React.FC = () => {
  const { authenticatePin, loading, error, clearError } = useEmployeeSessionStore();
  const theme = useThemeStore((s) => s.theme);
  const [pin, setPin] = useState('');
  const [shake, setShake] = useState(false);

  // Auto-submit when PIN is full
  useEffect(() => {
    if (pin.length === PIN_LENGTH) {
      handleSubmit(pin);
    }
  }, [pin]);

  // Keyboard support
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key >= '0' && e.key <= '9') pushDigit(e.key);
      if (e.key === 'Backspace') deleteDigit();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [pin]);

  const pushDigit = (d: string) => {
    clearError();
    setPin((p) => (p.length < PIN_LENGTH ? p + d : p));
  };

  const deleteDigit = () => {
    clearError();
    setPin((p) => p.slice(0, -1));
  };

  const handleSubmit = async (value: string) => {
    try {
      await authenticatePin(value);
    } catch {
      setShake(true);
      setPin('');
      setTimeout(() => setShake(false), 500);
    }
  };

  const handleKey = (key: string) => {
    if (key === '⌫') deleteDigit();
    else if (key !== '') pushDigit(key);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-pos-bg">
      <div className="w-full max-w-xs bg-pos-card border border-pos-border rounded-2xl p-8 flex flex-col items-center gap-6">
        <div className="flex flex-col items-center gap-1 text-center">
          <img
            src={theme === 'dark'
              ? 'https://cdn.voxire.dev/web_content/rtylr/logo-dark.webp'
              : 'https://cdn.voxire.dev/web_content/rtylr/logo-light.webp'}
            alt="rtylr"
            className="h-8 w-auto"
          />
          <h1 className="text-lg font-bold text-pos-text">Vbitez POS</h1>
          <p className="text-xs text-pos-muted">Enter your employee PIN</p>
        </div>

        {/* PIN dots */}
        <div className={`flex gap-3 ${shake ? 'animate-[shake_0.5s_ease-in-out]' : ''}`}>
          {Array.from({ length: PIN_LENGTH }).map((_, i) => (
            <span
              key={i}
              className={`w-3 h-3 rounded-full border-2 transition-colors ${
                error
                  ? 'border-destructive bg-destructive/30'
                  : i < pin.length
                  ? 'border-pos-text bg-pos-text'
                  : 'border-pos-muted bg-transparent'
              }`}
            />
          ))}
        </div>

        {error && <p className="text-xs text-destructive text-center">{error}</p>}

        {/* Numpad */}
        <div className="flex flex-col gap-2 w-full">
          {KEYS.map((row, ri) => (
            <div key={ri} className="flex gap-2 justify-center">
              {row.map((key, ki) => (
                <button
                  key={ki}
                  className={`w-16 h-14 rounded-xl text-base font-semibold transition-colors ${
                    key === ''
                      ? 'invisible'
                      : key === '⌫'
                      ? 'bg-transparent text-pos-muted hover:bg-white/8'
                      : 'bg-white/8 text-pos-text hover:bg-white/14 active:scale-95'
                  }`}
                  onClick={() => handleKey(key)}
                  disabled={loading || key === ''}
                  aria-label={key === '⌫' ? 'Delete' : key}
                >
                  {key}
                </button>
              ))}
            </div>
          ))}
        </div>

        {loading && <p className="text-xs text-pos-muted">Verifying…</p>}
      </div>
    </div>
  );
};

export default PinLogin;
