import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface HeadcountData {
  total: number;
  fullTime: number;
  partTime: number;
  contractors: number;
  trend: number;
}

interface HeadcountWidgetProps {
  data: HeadcountData | null;
  isLoading: boolean;
}

const HeadcountWidget = ({ data, isLoading }: HeadcountWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <WidgetShell
      title={t('dashboard.widgets.headcount')}
      icon={<Users size={14} />}
      isLoading={isLoading}
      isEmpty={!data}
      emptyMessage={t('dashboard.widgets.noData')}
    >
      {data && (
        <>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
            <span style={{ fontSize: 28, fontWeight: 700, color: 'var(--hr-text)' }}>
              {data.total}
            </span>
            {data.trend !== 0 && (
              <span
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  color: data.trend > 0 ? 'var(--hr-status-approved)' : 'var(--hr-error)',
                }}
              >
                {data.trend > 0 ? '+' : ''}
                {data.trend} {t('dashboard.widgets.thisMonth')}
              </span>
            )}
          </div>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            {[
              { label: t('dashboard.widgets.fullTime'), value: data.fullTime },
              { label: t('dashboard.widgets.partTime'), value: data.partTime },
              { label: t('dashboard.widgets.contractors'), value: data.contractors },
            ].map((item) => (
              <div key={item.label}>
                <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
                  {item.value}
                </div>
                <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>{item.label}</div>
              </div>
            ))}
          </div>
        </>
      )}
    </WidgetShell>
  );
};

export default HeadcountWidget;
