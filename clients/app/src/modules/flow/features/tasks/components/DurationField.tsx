import React from 'react';
import { parseDurationInput } from '../../../shared/lib/durationUtils';

interface DurationFieldProps {
  id: string;
  label: string;
  value: string;
  onChange: (next: string) => void;
  onSubmit?: () => void;
  onCancel?: () => void;
  placeholder?: string;
  /** Sentence explaining what the field accepts. Shown always, not only on error. */
  hint?: string;
  invalidMessage?: string;
  autoFocus?: boolean;
  disabled?: boolean;
}

/**
 * Text input for a duration, accepting `90m`, `1.5h` or `1:30`.
 *
 * A free-text field rather than a number input on purpose: people type the
 * shape they think in, and a stepper on minutes is unusable for `1:45`.
 *
 * The hint is permanent rather than error-triggered - a field with three valid
 * notations and no visible hint teaches nobody the other two.
 *
 * Base size is 16px so iOS does not zoom on focus; `md:` drops it to the dense
 * 11px the desktop rail uses.
 */
const DurationField = ({
  id,
  label,
  value,
  onChange,
  onSubmit,
  onCancel,
  placeholder,
  hint,
  invalidMessage,
  autoFocus,
  disabled,
}: DurationFieldProps): React.JSX.Element => {
  const touched = value.trim().length > 0;
  const invalid = touched && parseDurationInput(value) === null;
  const hintId = `${id}-hint`;
  const errorId = `${id}-error`;

  return (
    <div>
      <label
        htmlFor={id}
        className="mb-1 block text-flow-meta font-semibold uppercase tracking-wider"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {label}
      </label>
      <input
        id={id}
        type="text"
        inputMode="text"
        autoComplete="off"
        autoFocus={autoFocus}
        disabled={disabled}
        value={value}
        placeholder={placeholder}
        aria-invalid={invalid || undefined}
        aria-describedby={invalid && invalidMessage ? errorId : hint ? hintId : undefined}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && onSubmit) {
            e.preventDefault();
            onSubmit();
          }
          if (e.key === 'Escape' && onCancel) {
            e.preventDefault();
            onCancel();
          }
        }}
        className="flow-focus min-h-[44px] w-full rounded-sm px-2 py-2 text-flow-title md:min-h-0 md:py-1 md:text-flow-meta"
        style={{
          background: 'var(--flow-input-bg)',
          border: `1px solid ${invalid ? 'var(--flow-status-rejected-border)' : 'var(--flow-border)'}`,
          color: 'var(--flow-text)',
        }}
      />
      {invalid && invalidMessage ? (
        <p
          id={errorId}
          className="mt-1 text-flow-meta"
          style={{ color: 'var(--flow-status-rejected)' }}
        >
          {invalidMessage}
        </p>
      ) : (
        hint && (
          <p
            id={hintId}
            className="mt-1 text-flow-meta"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {hint}
          </p>
        )
      )}
    </div>
  );
};

export default DurationField;
