import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React from 'react';
import { AlertCircle } from 'lucide-react';

import { useTranslation } from 'react-i18next';

/**
 * The third state, for screens that had only two.
 *
 * An empty list, a failed request and a loading list are three different
 * facts. With every request forced to 500, Flow told an owner "No tasks
 * assigned - Nothing is on your plate right now. Nice work!" and "All caught
 * up!", which is worse than showing nothing: it reports success while the
 * server is down.
 *
 * The copy is deliberately generic and shared, so adding this state to a
 * screen costs an import rather than three locale entries per screen. Screens
 * whose failure has a specific cause worth naming, Directory and Presence
 * mention permissions, keep their own wording.
 */
export const ListLoadFailed: React.FC<{ onRetry: () => void }> = ({ onRetry }) => {
  const { t } = useTranslation();
  return (
    <EmptyState
      icon={AlertCircle}
      title={t('flow.loadFailed.title')}
      description={t('flow.loadFailed.description')}
      action={{ label: t('flow.loadFailed.retry'), onClick: onRetry }}
    />
  );
};
