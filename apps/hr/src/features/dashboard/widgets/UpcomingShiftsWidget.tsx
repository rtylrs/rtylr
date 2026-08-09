import React from 'react';
import { useTranslation } from 'react-i18next';
import { CalendarDays } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface ShiftEntry {
  id: number;
  name: string;
  start: string;
  end: string;
  role: string;
}

interface UpcomingShiftsWidgetProps {
  data: ShiftEntry[] | null;
  isLoading: boolean;
}

const UpcomingShiftsWidget = ({
  data,
  isLoading,
}: UpcomingShiftsWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <WidgetShell
      title={t('dashboard.widgets.upcomingShifts')}
      icon={<CalendarDays size={14} />}
      isLoading={isLoading}
      isEmpty={!data || data.length === 0}
      emptyMessage={t('dashboard.widgets.noShifts')}
    >
      {data && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {data.map((s) => (
            <div
              key={s.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '8px 10px',
                borderRadius: 8,
                background: 'var(--hr-hover)',
                gap: 8,
              }}
            >
              <div style={{ minWidth: 0 }}>
                <p
                  style={{
                    margin: 0,
                    fontSize: 12,
                    fontWeight: 600,
                    color: 'var(--hr-text)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {s.name}
                </p>
                <p style={{ margin: 0, fontSize: 10, color: 'var(--hr-text-muted)' }}>{s.role}</p>
              </div>
              <span
                style={{
                  fontSize: 11,
                  color: 'var(--hr-text-muted)',
                  flexShrink: 0,
                  whiteSpace: 'nowrap',
                }}
              >
                {s.start} – {s.end}
              </span>
            </div>
          ))}
        </div>
      )}
    </WidgetShell>
  );
};

export default UpcomingShiftsWidget;
