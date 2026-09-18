import React from 'react';
import { useTranslation } from 'react-i18next';
import { Clock } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface HoursThisWeekData {
  hoursLogged: number;
  target: number;
}

interface HoursThisWeekWidgetProps {
  data: HoursThisWeekData | null;
  isLoading: boolean;
}

const HoursThisWeekWidget = ({ data, isLoading }: HoursThisWeekWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();
  const pct = data && data.target > 0 ? Math.min((data.hoursLogged / data.target) * 100, 100) : 0;

  return (
    <WidgetShell
      title={t('dashboard.widgets.hoursThisWeek')}
      icon={<Clock size={14} />}
      isLoading={isLoading}
      isEmpty={!data}
      emptyMessage={t('dashboard.widgets.noData')}
    >
      {data && (
        <>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
            <span style={{ fontSize: 28, fontWeight: 700, color: 'var(--hr-text)' }}>
              {data.hoursLogged}
            </span>
            <span style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
              / {data.target} {t('dashboard.widgets.hours')}
            </span>
          </div>
          <div
            style={{
              height: 6,
              borderRadius: 3,
              background: 'var(--hr-hover)',
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                height: '100%',
                width: `${pct}%`,
                borderRadius: 3,
                background: pct >= 100 ? 'var(--hr-status-approved)' : 'var(--hr-accent)',
                transition: 'width 0.3s ease',
              }}
            />
          </div>
        </>
      )}
    </WidgetShell>
  );
};

export default HoursThisWeekWidget;
