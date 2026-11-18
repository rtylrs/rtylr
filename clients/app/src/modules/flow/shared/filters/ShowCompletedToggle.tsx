import React from 'react';
import { useTranslation } from 'react-i18next';
import { CheckCheck } from 'lucide-react';

/** The "Show completed" pill a cross-project list puts in its filter bar. */
export const ShowCompletedToggle: React.FC<{
  pressed: boolean;
  onChange: (pressed: boolean) => void;
}> = ({ pressed, onChange }) => {
  const { t } = useTranslation();
  return (
    <button
      type="button"
      aria-pressed={pressed}
      onClick={() => onChange(!pressed)}
      className="flow-focus flex min-h-9 items-center gap-1.5 rounded-md px-2.5 text-flow-meta font-medium md:min-h-8"
      style={
        pressed
          ? {
              background: 'var(--accent-muted)',
              color: 'var(--accent-on-muted)',
              border: '1px solid var(--accent-border)',
            }
          : { color: 'var(--flow-text-muted)', border: '1px solid var(--flow-border)' }
      }
    >
      <CheckCheck size={13} aria-hidden="true" />
      {t('flow.filters.showCompleted')}
    </button>
  );
};

export default ShowCompletedToggle;
