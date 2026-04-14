import React from 'react';
import { useTranslation } from 'react-i18next';
import { Palmtree } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface LeaveBalance {
  type: string;
  used: number;
  entitled: number;
}

interface LeaveBalanceWidgetProps {
  data: LeaveBalance[] | null;
  isLoading: boolean;
}

const LeaveBalanceWidget = ({ data, isLoading }: LeaveBalanceWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <WidgetShell
      title={t('dashboard.widgets.leaveBalance')}
      icon={<Palmtree size={14} />}
      isLoading={isLoading}
      isEmpty={!data || data.length === 0}
      emptyMessage={t('dashboard.widgets.noLeaveTypes')}
    >
      {data && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {data.map((b) => {
            const remaining = b.entitled - b.used;
            const pct = b.entitled > 0 ? (b.used / b.entitled) * 100 : 0;
            return (
              <div key={b.type}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    fontSize: 12,
                    marginBottom: 4,
                  }}
                >
                  <span style={{ color: 'var(--hr-text)', fontWeight: 600 }}>{b.type}</span>
                  <span style={{ color: 'var(--hr-text-muted)' }}>
                    {remaining} {t('dashboard.widgets.daysLeft')}
                  </span>
                </div>
                <div
                  style={{
                    height: 4,
                    borderRadius: 2,
                    background: 'var(--hr-hover)',
                    overflow: 'hidden',
                  }}
                >
                  <div
                    style={{
                      height: '100%',
                      width: `${Math.min(pct, 100)}%`,
                      borderRadius: 2,
                      background: pct > 80 ? 'var(--hr-status-pending)' : 'var(--hr-accent)',
                      transition: 'width 0.3s ease',
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </WidgetShell>
  );
};

export default LeaveBalanceWidget;
