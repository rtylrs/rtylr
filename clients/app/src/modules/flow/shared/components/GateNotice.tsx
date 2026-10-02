import React from 'react';
import { Lock } from 'lucide-react';
import { cn } from '../lib/utils';

interface GateNoticeProps {
  /** Already-translated sentence explaining why the adjacent control cannot be used. */
  reason: string;
  /** Pass the same value to the control's `aria-describedby`. */
  id?: string;
  className?: string;
}

/**
 * The reason line that sits beside a disabled control.
 *
 * Root CLAUDE.md rule 12: a disabled control states its reason, or is absent.
 * One component so the eleven gated controls in Flow phrase it the same way -
 * and so the reason is never the thing that gets dropped under deadline.
 *
 * Returns null on an empty reason, which lets callers pass a value that may or
 * may not resolve to a gate without wrapping every use in a conditional.
 */
const GateNotice = ({ reason, id, className }: GateNoticeProps): React.JSX.Element | null => {
  if (!reason) return null;

  return (
    <p
      id={id}
      className={cn('flex items-start gap-1.5 text-flow-meta leading-snug', className)}
      style={{ color: 'var(--flow-text-muted)' }}
    >
      <Lock size={11} aria-hidden="true" className="mt-px shrink-0" />
      <span>{reason}</span>
    </p>
  );
};

export default GateNotice;
