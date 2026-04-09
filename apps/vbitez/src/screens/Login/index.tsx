import React, { useState, useEffect } from 'react';
import { useEmployeeSessionStore } from '../../stores/employeeSessionStore';
import styles from './Login.module.scss';

const PIN_LENGTH = 4;

const KEYS = [
  ['1', '2', '3'],
  ['4', '5', '6'],
  ['7', '8', '9'],
  ['', '0', '⌫'],
];

const PinLogin: React.FC = () => {
  const { authenticatePin, loading, error, clearError } = useEmployeeSessionStore();
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
    // Dev bypass — remove before production
    if (value === '1253') {
      useEmployeeSessionStore.setState({
        employee: { id: 'dev', name: 'Dev User', role: 'admin' },
        isActive: true,
      });
      return;
    }
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
    <div className={styles.page}>
      <div className={styles.card}>
        <div className={styles.brand}>
          <span className={styles.brand__icon}>⬡</span>
          <h1 className={styles.brand__name}>Vbitez POS</h1>
          <p className={styles.brand__sub}>Enter your employee PIN</p>
        </div>

        {/* PIN dots */}
        <div className={`${styles.dots} ${shake ? styles.dots__shake : ''}`}>
          {Array.from({ length: PIN_LENGTH }).map((_, i) => (
            <span
              key={i}
              className={`${styles.dot} ${i < pin.length ? styles.dot__filled : ''} ${error ? styles.dot__error : ''}`}
            />
          ))}
        </div>

        {error && <p className={styles.errorMsg}>{error}</p>}

        {/* Numpad */}
        <div className={styles.pad}>
          {KEYS.map((row, ri) => (
            <div key={ri} className={styles.pad__row}>
              {row.map((key, ki) => (
                <button
                  key={ki}
                  className={`${styles.pad__key} ${key === '' ? styles.pad__key__empty : ''} ${key === '⌫' ? styles.pad__key__back : ''}`}
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

        {loading && <p className={styles.loadingMsg}>Verifying…</p>}
      </div>
    </div>
  );
};

export default PinLogin;
