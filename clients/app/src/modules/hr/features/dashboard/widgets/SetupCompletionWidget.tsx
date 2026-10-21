import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { Rocket } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface SetupStep {
  key: string;
  label: string;
  completed: boolean;
}

interface SetupCompletionWidgetProps {
  data: SetupStep[] | null;
  isLoading: boolean;
}

const SetupCompletionWidget = ({
  data,
  isLoading,
}: SetupCompletionWidgetProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const completed = data?.filter((s) => s.completed).length ?? 0;
  const total = data?.length ?? 0;
  const allDone = total > 0 && completed === total;

  if (allDone) return <></>;

  const pct = total > 0 ? (completed / total) * 100 : 0;

  return (
    <WidgetShell
      title={t('hrDashboard.widgets.setupCompletion')}
      icon={<Rocket size={14} />}
      isLoading={isLoading}
      isEmpty={!data || total === 0}
      emptyMessage={t('hrDashboard.widgets.noData')}
    >
      {data && total > 0 && (
        <>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginBottom: 4 }}>
            <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
              {completed}/{total}
            </span>
            <span style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
              {t('hrDashboard.widgets.stepsCompleted')}
            </span>
          </div>
          <div
            style={{
              height: 6,
              borderRadius: 3,
              background: 'var(--hr-hover)',
              overflow: 'hidden',
              marginBottom: 8,
            }}
          >
            <div
              style={{
                height: '100%',
                width: `${pct}%`,
                borderRadius: 3,
                background: 'var(--hr-accent)',
                transition: 'width 0.3s ease',
              }}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {data
              .filter((s) => !s.completed)
              .slice(0, 3)
              .map((s) => (
                <Link
                  key={s.key}
                  to="/hr/setup"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    padding: '6px 8px',
                    borderRadius: 6,
                    background: 'var(--hr-hover)',
                    textDecoration: 'none',
                    color: 'var(--hr-text)',
                    fontSize: 11,
                    fontWeight: 500,
                  }}
                >
                  <span
                    style={{
                      height: 14,
                      width: 14,
                      borderRadius: 4,
                      border: '1.5px solid var(--hr-border)',
                      flexShrink: 0,
                    }}
                  />
                  {s.label}
                </Link>
              ))}
          </div>
        </>
      )}
    </WidgetShell>
  );
};

export default SetupCompletionWidget;
