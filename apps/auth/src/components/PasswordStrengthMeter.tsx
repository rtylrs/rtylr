import React, { useMemo } from 'react';

export interface PasswordStrengthMeterProps {
  password: string;
  showLabel?: boolean;
}

type Strength = 0 | 1 | 2 | 3 | 4;

export const scorePassword = (pw: string): Strength => {
  if (!pw) return 0;
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  if (/[A-Z]/.test(pw) && /[a-z]/.test(pw)) score++;
  if (/\d/.test(pw)) score++;
  if (/[^A-Za-z0-9]/.test(pw)) score++;
  return Math.min(score, 4) as Strength;
};

const LABELS = ['Too short', 'Weak', 'Fair', 'Good', 'Strong'];
const COLORS = [
  'bg-white/10',
  'bg-red-400/80',
  'bg-amber-400/80',
  'bg-sky-400/80',
  'bg-[#64C5B9]',
];

const PasswordStrengthMeter: React.FC<PasswordStrengthMeterProps> = ({
  password,
  showLabel = true,
}) => {
  const score = useMemo(() => scorePassword(password), [password]);

  return (
    <div className="flex flex-col gap-2">
      <div className="flex gap-1.5">
        {[0, 1, 2, 3].map((i) => (
          <div
            key={i}
            className={[
              'h-1.5 flex-1 rounded-full transition-colors duration-300',
              i < score ? COLORS[score] : 'bg-white/5',
            ].join(' ')}
          />
        ))}
      </div>
      {showLabel && (
        <p className="text-[11px] text-white/45">
          Password strength:{' '}
          <span
            className={[
              'font-medium',
              score >= 3 ? 'text-[#64C5B9]' : score === 2 ? 'text-amber-300' : 'text-white/60',
            ].join(' ')}
          >
            {LABELS[score]}
          </span>
        </p>
      )}
    </div>
  );
};

export default PasswordStrengthMeter;
