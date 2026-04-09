import React from 'react';
import { Check } from 'lucide-react';
import { APP_URLS } from '@voxire/config';

export interface TermsCheckboxProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: React.ReactNode;
}

const TermsCheckbox: React.FC<TermsCheckboxProps> = ({
  checked,
  onChange,
  label,
}) => {
  return (
    <label className="flex items-start gap-2.5 cursor-pointer select-none group">
      <span
        className={[
          'relative flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-md border transition-colors mt-0.5',
          checked
            ? 'bg-[#4583C4] border-[#4583C4]'
            : 'bg-white/[0.04] border-white/20 group-hover:border-white/40',
        ].join(' ')}
      >
        {checked && <Check className="h-3 w-3 text-white" strokeWidth={3} />}
      </span>
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="sr-only"
      />
      <span className="text-xs text-white/60 leading-relaxed">
        {label ?? (
          <>
            I agree to the{' '}
            <a
              href={`${APP_URLS.landing}/legal/terms`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-white/85 hover:text-[#64C5B9] underline underline-offset-2"
            >
              Terms of Service
            </a>{' '}
            and{' '}
            <a
              href={`${APP_URLS.landing}/legal/privacy-policy`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-white/85 hover:text-[#64C5B9] underline underline-offset-2"
            >
              Privacy Policy
            </a>
            .
          </>
        )}
      </span>
    </label>
  );
};

export default TermsCheckbox;
