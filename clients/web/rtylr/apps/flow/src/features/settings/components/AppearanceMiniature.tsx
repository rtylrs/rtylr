import React from 'react';
import type { AppearanceTokens } from '@voxire/utils';
import { cn } from '../../../shared/lib/utils';

/**
 * The small render that sits inside a mode or palette card.
 *
 * A card has to answer "what will this look like" and a flat rectangle of the
 * background colour cannot: it is the same shape whichever palette it paints, so
 * the only thing it distinguishes is light from dark. This renders the four
 * pieces that actually differ between palettes: a rail on its own surface, body
 * text, a card floating above the ground, and the accent.
 *
 * Every colour comes from the tokens passed in, applied to this element so the
 * children resolve them by inheritance. Nothing here writes to documentElement,
 * which is what keeps it from fighting the provider.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * section 5.
 */

/**
 * CSS custom properties are valid in a style object but absent from
 * `React.CSSProperties`, which indexes only the known properties. The cast lives
 * here, once and explained, instead of at every call site.
 */
export function tokenStyle(
  tokens: AppearanceTokens,
  extra?: React.CSSProperties
): React.CSSProperties {
  return { ...tokens, ...extra } as React.CSSProperties;
}

interface AppearanceMiniatureProps {
  tokens: AppearanceTokens;
  className?: string;
}

const AppearanceMiniature = ({
  tokens,
  className,
}: AppearanceMiniatureProps): React.JSX.Element => (
  // Decorative: every card already carries its own name and description, so a
  // screen reader gains nothing from the shapes and loses the label to noise.
  <div
    aria-hidden="true"
    className={cn('flex h-14 w-full overflow-hidden', className)}
    style={tokenStyle(tokens, {
      background: 'var(--flow-bg)',
      borderRadius: 'var(--flow-r)',
      border: '1px solid var(--flow-border)',
    })}
  >
    <div
      className="flex w-4 shrink-0 flex-col justify-center gap-1 px-1"
      style={{
        background: 'var(--flow-surface)',
        borderInlineEnd: '1px solid var(--flow-border)',
      }}
    >
      <span className="block h-1 rounded-full" style={{ background: 'var(--accent)' }} />
      <span className="block h-1 rounded-full" style={{ background: 'var(--flow-text-dim)' }} />
      <span className="block h-1 rounded-full" style={{ background: 'var(--flow-text-dim)' }} />
    </div>

    <div className="flex min-w-0 flex-1 flex-col justify-center gap-1.5 px-2">
      <span className="block h-1.5 w-3/5 rounded-full" style={{ background: 'var(--flow-text)' }} />
      <div
        className="flex items-center gap-1.5 px-1.5 py-1"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          borderRadius: 'var(--flow-r)',
        }}
      >
        <span
          className="block h-1.5 w-1.5 shrink-0 rounded-full"
          style={{ background: 'var(--accent)' }}
        />
        <span
          className="block h-1 min-w-0 flex-1 rounded-full"
          style={{ background: 'var(--flow-text-muted)' }}
        />
      </div>
    </div>
  </div>
);

export default AppearanceMiniature;
