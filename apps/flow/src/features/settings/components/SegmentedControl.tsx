import React from 'react';
import { cn } from '../../../shared/lib/utils';

/**
 * A one-of-three chooser for the shape preferences.
 *
 * Buttons rather than radios, matching the rest of Flow's settings surfaces, so
 * selection is announced through `aria-pressed` and the group carries the name
 * of the preference it sets. The name comes from the visible heading through
 * `aria-labelledby` rather than being repeated in an `aria-label`, which is what
 * makes each button's own short label ("Compact") intelligible out of context
 * without announcing the preference's name twice.
 *
 * The audition callbacks are wired to hover and focus alike, and neither
 * commits: only `onChange` does. A disabled control gets no audition wiring at
 * all, because auditioning a value the user cannot select would preview a state
 * that is unreachable.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * section 5.
 */

export interface SegmentedOption<T extends string> {
  value: T;
  /** Already translated. */
  label: string;
}

interface SegmentedControlProps<T extends string> {
  /** id of the visible heading that names the preference this group sets. */
  labelledBy: string;
  options: readonly SegmentedOption<T>[];
  value: T;
  onChange: (value: T) => void;
  /** Called with an option while it is hovered or focused, and null on leave. */
  onAudition?: (value: T | null) => void;
  disabled?: boolean;
  /** id of the GateNotice that explains why a disabled control is disabled. */
  describedBy?: string;
  className?: string;
}

function SegmentedControl<T extends string>({
  labelledBy,
  options,
  value,
  onChange,
  onAudition,
  disabled = false,
  describedBy,
  className,
}: SegmentedControlProps<T>): React.JSX.Element {
  return (
    <div
      role="group"
      aria-labelledby={labelledBy}
      className={cn('grid gap-2', className)}
      style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}
    >
      {options.map((option) => {
        const selected = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            disabled={disabled}
            aria-pressed={selected}
            aria-describedby={describedBy}
            onClick={() => onChange(option.value)}
            onPointerEnter={disabled ? undefined : () => onAudition?.(option.value)}
            onPointerLeave={disabled ? undefined : () => onAudition?.(null)}
            onFocus={disabled ? undefined : () => onAudition?.(option.value)}
            onBlur={disabled ? undefined : () => onAudition?.(null)}
            className="flow-focus inline-flex min-h-[44px] items-center justify-center px-3 text-center text-[13px] font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-60"
            style={{
              background: selected ? 'var(--accent-muted)' : 'var(--flow-card)',
              border: `1px solid ${selected ? 'var(--accent-border)' : 'var(--flow-border)'}`,
              borderRadius: 'var(--flow-r)',
              color: selected ? 'var(--accent-fg)' : 'var(--flow-text-muted)',
            }}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

export default SegmentedControl;
