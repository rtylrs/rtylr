import React, { useEffect, useRef, useState } from 'react';

interface BarcodeScannerInputProps {
  onScan: (code: string) => void;
  autoFocus?: boolean;
  placeholder?: string;
}

// Most USB barcode scanners emit keystrokes followed by Enter.
// This input auto-focuses and forwards the typed value to onScan on Enter.
const BarcodeScannerInput: React.FC<BarcodeScannerInputProps> = ({
  onScan,
  autoFocus = true,
  placeholder = 'Scan barcode...',
}) => {
  const ref = useRef<HTMLInputElement>(null);
  const [value, setValue] = useState('');

  useEffect(() => {
    if (autoFocus) ref.current?.focus();
    const refocus = () => ref.current?.focus();
    window.addEventListener('click', refocus);
    return () => window.removeEventListener('click', refocus);
  }, [autoFocus]);

  const submit = () => {
    const v = value.trim();
    if (v) onScan(v);
    setValue('');
  };

  return (
    <input
      ref={ref}
      type="text"
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onKeyDown={(e) => e.key === 'Enter' && submit()}
      placeholder={placeholder}
      className="h-11 w-full rounded-lg bg-white/6 border border-pos-border px-3 text-sm text-pos-text outline-none focus:border-white/25"
      aria-label="Barcode scanner input"
    />
  );
};

export default BarcodeScannerInput;
