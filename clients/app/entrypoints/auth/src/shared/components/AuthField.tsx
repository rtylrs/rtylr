import React, { useState } from 'react';
import { AlertCircle, CheckCircle2 } from 'lucide-react';

/**
 * The auth app's form primitives. Every screen used to repeat the same
 * input class string (13 copies) and drift: a different focus colour here,
 * a missing disabled style there. These carry the one look, matching the
 * marketing site: soft rounded fields, an ink focus ring, quiet errors.
 */

const inputClass = (invalid: boolean, extra = ''): string =>
  [
    'w-full h-11 rounded-[10px] border bg-white px-3.5 text-base md:text-sm text-[#111]',
    'placeholder:text-[#9a9a9a] outline-none transition-[border-color,box-shadow] duration-150',
    'disabled:bg-[#f5f5f5] disabled:text-[#999]',
    invalid
      ? 'border-red-300 focus:border-red-400 focus:ring-4 focus:ring-red-100'
      : 'border-black/10 hover:border-black/20 focus:border-[#111] focus:ring-4 focus:ring-black/[0.06]',
    extra,
  ].join(' ');

/** The field style on its own, for inputs that can't use AuthField. */
export const authInputClass = inputClass(false);

interface FieldProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'id'> {
  id: string;
  label: React.ReactNode;
  /** Shown under the field and linked via aria-describedby. */
  error?: string | null;
  /** Content rendered to the right of the label (e.g. a "Forgot?" link). */
  labelAside?: React.ReactNode;
}

const FieldShell: React.FC<{
  id: string;
  label: React.ReactNode;
  labelAside?: React.ReactNode;
  error?: string | null;
  children: React.ReactNode;
}> = ({ id, label, labelAside, error, children }) => (
  <div className="flex flex-col gap-1.5">
    <div className="flex items-center justify-between">
      <label htmlFor={id} className="text-[13px] font-medium text-[#111]">
        {label}
      </label>
      {labelAside}
    </div>
    {children}
    {error && (
      <p id={`${id}-error`} className="text-xs text-red-600" role="alert">
        {error}
      </p>
    )}
  </div>
);

export const AuthField: React.FC<FieldProps> = ({
  id,
  label,
  error,
  labelAside,
  className,
  ...input
}) => (
  <FieldShell id={id} label={label} labelAside={labelAside} error={error}>
    <input
      id={id}
      aria-invalid={!!error || undefined}
      aria-describedby={error ? `${id}-error` : undefined}
      className={inputClass(!!error, className)}
      {...input}
    />
  </FieldShell>
);

interface PasswordFieldProps extends Omit<FieldProps, 'type'> {
  showLabel?: string;
  hideLabel?: string;
}

/** Password input with a show/hide toggle. Labels default to the old literals. */
export const AuthPasswordField: React.FC<PasswordFieldProps> = ({
  id,
  label,
  error,
  labelAside,
  className,
  showLabel = 'Show',
  hideLabel = 'Hide',
  ...input
}) => {
  const [visible, setVisible] = useState(false);
  return (
    <FieldShell id={id} label={label} labelAside={labelAside} error={error}>
      <div className="relative">
        <input
          id={id}
          type={visible ? 'text' : 'password'}
          aria-invalid={!!error || undefined}
          aria-describedby={error ? `${id}-error` : undefined}
          className={inputClass(!!error, `pr-14 ${className ?? ''}`)}
          {...input}
        />
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          aria-label={visible ? `${hideLabel} password` : `${showLabel} password`}
          className="absolute right-1.5 top-1/2 -translate-y-1/2 h-8 rounded-md px-2.5 text-xs font-medium text-[#777] hover:bg-black/[0.04] hover:text-[#111] transition-colors"
        >
          {visible ? hideLabel : showLabel}
        </button>
      </div>
    </FieldShell>
  );
};

/** Inline status message for a form: an error, or a success confirmation. */
export const AuthAlert: React.FC<{ tone?: 'error' | 'success'; children: React.ReactNode }> = ({
  tone = 'error',
  children,
}) => (
  <div
    className={[
      'flex items-start gap-2.5 rounded-[10px] px-3.5 py-3',
      tone === 'error' ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-800',
    ].join(' ')}
    role="alert"
    aria-live="polite"
  >
    {tone === 'error' ? (
      <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
    ) : (
      <CheckCircle2 className="h-4 w-4 mt-0.5 shrink-0" />
    )}
    <p className="text-[13px] leading-relaxed">{children}</p>
  </div>
);

/** Thin rule with a centered word, between the form and sign-in providers. */
export const AuthDivider: React.FC<{ label: string }> = ({ label }) => (
  <div className="flex items-center gap-3 text-xs text-[#999]" aria-hidden="true">
    <span className="h-px flex-1 bg-black/[0.08]" />
    {label}
    <span className="h-px flex-1 bg-black/[0.08]" />
  </div>
);

/** Text link styled for the auth screens (ink, underline on hover). */
export const authLinkClass =
  'font-medium text-[#111] underline decoration-black/25 underline-offset-[3px] hover:decoration-black/70 transition-colors';
