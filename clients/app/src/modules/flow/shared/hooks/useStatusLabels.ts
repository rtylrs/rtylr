import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../stores/flowStore';
import { STATUS_LABEL_KEYS } from '../lib/constantsFlow';

/**
 * Translated display label for every task status, resolved from the single
 * canonical STATUS_LABEL_KEYS map (see constantsFlow.ts). Any component that
 * renders a status label - the task detail status picker, a project's List
 * view groups, the StatusIcon tooltip - should read from this hook instead
 * of hardcoding its own copy, so the label can't drift between views again
 * (issue #756).
 */
export function useStatusLabels(): Record<Task['status'], string> {
  const { t } = useTranslation();

  return useMemo(() => {
    const entries = Object.entries(STATUS_LABEL_KEYS) as [Task['status'], string][];
    return entries.reduce(
      (labels, [status, key]) => {
        labels[status] = t(key);
        return labels;
      },
      {} as Record<Task['status'], string>
    );
  }, [t]);
}
