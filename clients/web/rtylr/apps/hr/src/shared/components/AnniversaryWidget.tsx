import React from 'react';
import { useTranslation } from 'react-i18next';
import { Trophy } from 'lucide-react';
import { Card } from './ui';
import { useAnniversaries } from '../hooks/useHrExtras';

function daysUntil(monthDay: string) {
  const [m, d] = monthDay.split('-').map(Number);
  const now = new Date();
  // Compare calendar days, not timestamps: `now` carries the current time of
  // day, so an anniversary dated today would compare as past and roll to next
  // year instead of rendering 'Today'.
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  let next = new Date(now.getFullYear(), m - 1, d);
  if (next < today) next = new Date(now.getFullYear() + 1, m - 1, d);
  return Math.round((next.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
}

const AnniversaryWidget: React.FC = () => {
  const { t } = useTranslation();
  const { data } = useAnniversaries();
  const upcoming = (data ?? [])
    .map((a) => ({ ...a, days: daysUntil(a.date) }))
    .sort((a, b) => a.days - b.days)
    .slice(0, 4);

  return (
    <Card>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
        <span
          style={{
            padding: 6,
            borderRadius: 8,
            background: 'rgba(245,158,11,0.15)',
            color: '#f59e0b',
          }}
        >
          <Trophy size={14} />
        </span>
        <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
          {t('dashboard.workAnniversaries')}
        </h3>
      </div>
      {upcoming.length === 0 ? (
        <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
          {t('dashboard.nothingUpcoming')}
        </p>
      ) : (
        <ul
          style={{
            listStyle: 'none',
            padding: 0,
            margin: 0,
            display: 'flex',
            flexDirection: 'column',
            gap: 6,
          }}
        >
          {upcoming.map((a) => (
            <li
              key={a.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                fontSize: 12,
              }}
            >
              <div>
                <div style={{ color: 'var(--hr-text)', fontWeight: 600 }}>{a.name}</div>
                <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>
                  {a.years} year milestone
                </div>
              </div>
              <span style={{ fontSize: 11, color: '#f59e0b', fontWeight: 600 }}>
                {a.days === 0 ? 'Today' : `in ${a.days}d`}
              </span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
};

export default AnniversaryWidget;
