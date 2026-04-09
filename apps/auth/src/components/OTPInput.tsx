import React, { useEffect, useRef, useState } from 'react';

export interface OTPInputProps {
  length?: number;
  value: string;
  onChange: (v: string) => void;
  onComplete?: (v: string) => void;
  autoFocus?: boolean;
  disabled?: boolean;
}

/**
 * Multi-cell OTP input with auto-advance, backspace navigation,
 * and paste support.
 */
const OTPInput: React.FC<OTPInputProps> = ({
  length = 6,
  value,
  onChange,
  onComplete,
  autoFocus = true,
  disabled,
}) => {
  const refs = useRef<Array<HTMLInputElement | null>>([]);
  const [focused, setFocused] = useState<number | null>(null);

  useEffect(() => {
    if (autoFocus && refs.current[0]) refs.current[0].focus();
  }, [autoFocus]);

  const chars = value.padEnd(length, ' ').split('').slice(0, length);

  const updateAt = (index: number, char: string) => {
    const arr = value.split('');
    while (arr.length < length) arr.push('');
    arr[index] = char;
    const next = arr.join('').replace(/\s/g, '').slice(0, length);
    onChange(next);
    if (next.length === length && onComplete) onComplete(next);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>, i: number) => {
    const char = e.target.value.slice(-1);
    if (!/^\d?$/.test(char)) return;
    updateAt(i, char);
    if (char && i < length - 1) refs.current[i + 1]?.focus();
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>, i: number) => {
    if (e.key === 'Backspace') {
      if (!chars[i].trim() && i > 0) {
        refs.current[i - 1]?.focus();
        updateAt(i - 1, '');
        e.preventDefault();
      } else {
        updateAt(i, '');
      }
    } else if (e.key === 'ArrowLeft' && i > 0) {
      refs.current[i - 1]?.focus();
    } else if (e.key === 'ArrowRight' && i < length - 1) {
      refs.current[i + 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const text = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, length);
    if (!text) return;
    onChange(text);
    if (text.length === length && onComplete) onComplete(text);
    refs.current[Math.min(text.length, length - 1)]?.focus();
  };

  return (
    <div className="flex items-center justify-between gap-2 sm:gap-3">
      {Array.from({ length }).map((_, i) => (
        <input
          key={i}
          ref={(el) => (refs.current[i] = el)}
          type="text"
          inputMode="numeric"
          pattern="\d*"
          maxLength={1}
          autoComplete="one-time-code"
          disabled={disabled}
          value={chars[i].trim()}
          onChange={(e) => handleChange(e, i)}
          onKeyDown={(e) => handleKeyDown(e, i)}
          onPaste={handlePaste}
          onFocus={() => setFocused(i)}
          onBlur={() => setFocused(null)}
          aria-label={`Digit ${i + 1}`}
          className={[
            'h-14 w-full rounded-xl text-center text-xl font-semibold text-white',
            'bg-white/[0.04] border transition-all duration-150 outline-none',
            focused === i
              ? 'border-[#4583C4] ring-2 ring-[#4583C4]/25 bg-white/[0.06]'
              : 'border-white/10 hover:border-white/20',
          ].join(' ')}
        />
      ))}
    </div>
  );
};

export default OTPInput;
