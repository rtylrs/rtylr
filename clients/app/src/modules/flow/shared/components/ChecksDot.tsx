import React from 'react';
import { useTranslation } from 'react-i18next';
import { CheckCircle2, Loader2, XCircle } from 'lucide-react';

export type ChecksState = 'failing' | 'running' | 'passing';

const LOOK: Record<ChecksState, { icon: typeof CheckCircle2; color: string }> = {
  failing: { icon: XCircle, color: 'var(--flow-danger-fg)' },
  running: { icon: Loader2, color: 'var(--flow-warn-fg)' },
  passing: { icon: CheckCircle2, color: 'var(--color-success)' },
};

/**
 * The CI state of a task's open pull requests on its card and rows (T-376).
 * Renders nothing when no PR has checks.
 */
const ChecksDot: React.FC<{ state?: string; size?: number }> = ({ state, size = 12 }) => {
  const { t } = useTranslation();
  if (state !== 'failing' && state !== 'running' && state !== 'passing') return null;
  const { icon: Icon, color } = LOOK[state];
  const label = t(`flow.checks.state.${state}`);
  return (
    <span className="inline-flex shrink-0 items-center" role="img" aria-label={label} title={label}>
      <Icon size={size} aria-hidden="true" style={{ color }} />
    </span>
  );
};

export default ChecksDot;
