import React from 'react';

/**
 * The reduce-motion toggle.
 *
 * Local rather than `@voxire/ui`'s `Switch` for one measurable reason: that one
 * sizes its root to the track, `h-5 w-9`, so the whole control is 20px tall and
 * fails the 44px hit-area floor. Here the button is the hit area and the track is
 * a 26px child of it, which is the arrangement the design calls for.
 *
 * The knob is a real element, not an `::after`. That matters for the setting this
 * switch controls: the reduced-motion rule has to reach pseudo-elements
 * explicitly, so a knob built as one is exactly the control most likely to keep
 * animating after a user asks it to stop. A child element is covered by any
 * rule that covers the subtree.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * sections 5 and 6.
 */

interface AppearanceSwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  /** Already translated. The switch carries no visible text of its own. */
  label: string;
}

const AppearanceSwitch = ({
  checked,
  onChange,
  label,
}: AppearanceSwitchProps): React.JSX.Element => (
  <button
    type="button"
    role="switch"
    aria-checked={checked}
    aria-label={label}
    onClick={() => onChange(!checked)}
    className="flow-focus inline-flex min-h-[44px] min-w-[44px] shrink-0 items-center justify-center"
    style={{ borderRadius: 'var(--flow-r)' }}
  >
    <span
      aria-hidden="true"
      className="relative block h-[26px] w-[46px] rounded-full transition-colors"
      style={{
        background: checked ? 'var(--accent)' : 'var(--flow-input-bg)',
        border: `1px solid ${checked ? 'var(--accent-border)' : 'var(--flow-border)'}`,
      }}
    >
      <span
        className="absolute top-1/2 block h-5 w-5 -translate-y-1/2 rounded-full transition-[inset-inline-start] duration-150"
        style={{
          background: checked ? 'var(--flow-on-accent)' : 'var(--flow-text-muted)',
          insetInlineStart: checked ? '21px' : '2px',
        }}
      />
    </span>
  </button>
);

export default AppearanceSwitch;
