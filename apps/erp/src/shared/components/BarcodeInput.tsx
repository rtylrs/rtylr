import React, { useEffect, useRef, useState } from 'react';
import { ScanLine } from 'lucide-react';

interface BarcodeInputProps {
  onScan: (code: string) => void;
  placeholder?: string;
  autoFocus?: boolean;
}

const BarcodeInput: React.FC<BarcodeInputProps> = ({
  onScan,
  placeholder = 'Scan or type barcode…',
  autoFocus = true,
}) => {
  const [value, setValue] = useState('');
  const ref = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (autoFocus) ref.current?.focus();
  }, [autoFocus]);

  const submit = () => {
    if (!value.trim()) return;
    onScan(value.trim());
    setValue('');
    ref.current?.focus();
  };

  return (
    <div
      className="flex items-center gap-2 rounded-lg px-3 py-2"
      style={{
        background: 'var(--hub-input-bg)',
        border: '1px solid var(--hub-border)',
      }}
    >
      <ScanLine size={14} style={{ color: '#f59e0b' }} />
      <input
        ref={ref}
        type="text"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') submit();
        }}
        placeholder={placeholder}
        className="flex-1 bg-transparent outline-none text-sm"
        style={{ color: 'var(--hub-text)' }}
      />
      <button
        onClick={submit}
        className="text-xs font-semibold px-2 py-1 rounded-md"
        style={{ background: 'rgba(245,158,11,0.15)', color: '#f59e0b' }}
      >
        Scan
      </button>
    </div>
  );
};

export default BarcodeInput;
