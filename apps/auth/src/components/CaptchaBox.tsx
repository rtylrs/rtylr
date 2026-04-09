import React, { useState } from 'react';
import { Shield, Loader2, Check } from 'lucide-react';

export interface CaptchaBoxProps {
  onVerify?: (token: string) => void;
}

/**
 * Placeholder captcha box. Not a real captcha, but mimics the hCaptcha /
 * reCAPTCHA visual pattern so the flow works end-to-end.
 */
const CaptchaBox: React.FC<CaptchaBoxProps> = ({ onVerify }) => {
  const [state, setState] = useState<'idle' | 'loading' | 'verified'>('idle');

  const handleClick = () => {
    if (state !== 'idle') return;
    setState('loading');
    setTimeout(() => {
      setState('verified');
      onVerify?.(`mock-token-${Date.now()}`);
    }, 900);
  };

  return (
    <div className="flex items-center gap-3 rounded-xl border border-white/10 bg-white/[0.04] px-4 py-3">
      <button
        type="button"
        onClick={handleClick}
        aria-label="I'm not a robot"
        className={[
          'relative flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded border transition-colors',
          state === 'verified'
            ? 'bg-[#64C5B9] border-[#64C5B9]'
            : 'bg-white/[0.06] border-white/25 hover:border-white/40',
        ].join(' ')}
      >
        {state === 'verified' && <Check className="h-3.5 w-3.5 text-[#05061a]" strokeWidth={3} />}
        {state === 'loading' && (
          <Loader2 className="h-3.5 w-3.5 text-white/70 animate-spin" />
        )}
      </button>
      <div className="flex-1">
        <p className="text-xs text-white/70 font-medium">I'm not a robot</p>
        <p className="text-[10px] text-white/40">Verify your humanity</p>
      </div>
      <Shield className="h-5 w-5 text-white/25" />
    </div>
  );
};

export default CaptchaBox;
