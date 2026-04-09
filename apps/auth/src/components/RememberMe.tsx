import React from 'react';
import { Check } from 'lucide-react';

export interface RememberMeProps {
  checked: boolean;
  onChange: (v: boolean) => void;
  label?: string;
}

const RememberMe: React.FC<RememberMeProps> = ({
  checked,
  onChange,
  label = 'Remember me for 30 days',
}) => {
  return (
    <label className="inline-flex items-center gap-2 cursor-pointer select-none group text-xs text-white/55 hover:text-white/80 transition-colors">
      <span
        className={[
          'relative flex h-[16px] w-[16px] items-center justify-center rounded border transition-colors',
          checked
            ? 'bg-[#4583C4] border-[#4583C4]'
            : 'bg-white/[0.04] border-white/20 group-hover:border-white/40',
        ].join(' ')}
      >
        {checked && <Check className="h-2.5 w-2.5 text-white" strokeWidth={3} />}
      </span>
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="sr-only"
      />
      {label}
    </label>
  );
};

export default RememberMe;
