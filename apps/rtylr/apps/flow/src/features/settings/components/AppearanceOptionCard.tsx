import React from 'react';

/**
 * The chrome and the accessibility around one auditionable option.
 *
 * Deliberately does not know what it is previewing: the panel composes the
 * miniature (or the pair of them the `system` card needs) and passes it in as a
 * child. That keeps the selection state, the hit area and the audition wiring in
 * one place, rather than once per control group.
 *
 * `onAuditionStart` fires on pointer enter and on focus. Pointer LEAVE is
 * deliberately not handled here: clearing per card meant that travelling from
 * one card to its neighbour crossed the grid gap with nothing hovered, and the
 * preview snapped back to the applied option for those few pixels. Measured as
 * eight pixels of flashing between every pair of cards. The group container
 * clears instead, so moving within the grid never reverts and leaving it does.
 *
 * `onAuditionEnd` remains for blur, where focus genuinely does leave the control
 * rather than merely passing between two of them.
 *
 * `onAuditionEnd` fires on
 * leave and on blur. Neither writes anything: only `onSelect` commits. Keyboard
 * users get the same audition as pointer users because focus is wired alongside
 * hover, which is the whole reason both pairs exist rather than just the pointer
 * one.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * section 5.
 */

interface AppearanceOptionCardProps {
  /** Already translated. */
  label: string;
  /** Already translated. One line saying what the option is for. */
  description?: string;
  selected: boolean;
  onSelect: () => void;
  onAuditionStart: () => void;
  onAuditionEnd: () => void;
  children: React.ReactNode;
}

const AppearanceOptionCard = ({
  label,
  description,
  selected,
  onSelect,
  onAuditionStart,
  onAuditionEnd,
  children,
}: AppearanceOptionCardProps): React.JSX.Element => (
  <button
    type="button"
    aria-pressed={selected}
    onClick={onSelect}
    onPointerEnter={onAuditionStart}
    onFocus={onAuditionStart}
    onBlur={onAuditionEnd}
    className="flow-focus flex min-h-[44px] w-full flex-col gap-2 p-2 text-start transition-colors"
    style={{
      background: 'var(--flow-card)',
      border: `2px solid ${selected ? 'var(--accent)' : 'var(--flow-border)'}`,
      borderRadius: 'var(--flow-r)',
    }}
  >
    {children}
    <span className="flex min-w-0 flex-col gap-0.5">
      <span
        className="text-meta font-medium"
        style={{ color: selected ? 'var(--accent-fg)' : 'var(--flow-text)' }}
      >
        {label}
      </span>
      {description ? (
        <span className="text-meta leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
          {description}
        </span>
      ) : null}
    </span>
  </button>
);

export default AppearanceOptionCard;
