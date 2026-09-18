import React from 'react';
import { useTranslation } from 'react-i18next';
import { Cake } from 'lucide-react';
import { Card, ACCENT } from './ui';
import { useBirthdays } from '../hooks/useHrExtras';

function daysUntil(monthDay: string) {
  const [m, d] = monthDay.split('-').map(Number);
  const now = new Date();
  let next = new Date(now.getFullYear(), m - 1, d);
  if (next < now) next = new Date(now.getFullYear() + 1, m - 1, d);
  return Math.ceil((next.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

const BirthdayWidget: React.FC = () => {
  const { t } = useTranslation();
  const { data } = useBirthdays();
  const upcoming = (data ?? [])
    .map((b) => ({ ...b, days: daysUntil(b.date) }))
    .sort((a, b) => a.days - b.days)
    .slice(0, 4);

  return (
    <Card>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 10 }}>
        <span style={{ padding: 6, borderRadius: 8, background: `${ACCENT}22`, color: ACCENT }}>
          <Cake size={14} />
        </span>
        <h3 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
          {t('dashboard.upcomingBirthdays')}
        </h3>
      </div>
      {upcoming.length === 0 ? (
        <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
          {t('dashboard.noBirthdays')}
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
          {upcoming.map((b) => (
            <li
              key={b.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                fontSize: 12,
              }}
            >
              <div>
                <div style={{ color: 'var(--hr-text)', fontWeight: 600 }}>{b.name}</div>
                {b.department && (
                  <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>{b.department}</div>
                )}
              </div>
              <span style={{ fontSize: 11, color: ACCENT, fontWeight: 600 }}>
                {b.days === 0 ? 'Today' : `in ${b.days}d`}
              </span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
};

export default BirthdayWidget;
