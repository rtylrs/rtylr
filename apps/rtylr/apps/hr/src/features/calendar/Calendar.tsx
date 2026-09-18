import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useTimeOff } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Button, ACCENT, useIsMobile } from '../../shared/components/ui';
import { isHrRequestStatus } from '../../shared/lib/requestStatus';

function startOfMonth(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}
function endOfMonth(date: Date) {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0);
}
function formatMonth(date: Date) {
  return date.toLocaleString('en-US', { month: 'long', year: 'numeric' });
}
function dayStr(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')}`;
}

const CalendarScreen: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const [view, setView] = useState(new Date());
  const { data: timeOff = [] } = useTimeOff();

  const cells = useMemo(() => {
    const first = startOfMonth(view);
    const last = endOfMonth(view);
    const startDay = first.getDay(); // 0 = Sun
    const cells: Array<{ date?: Date; events: typeof timeOff }> = [];
    for (let i = 0; i < startDay; i++) cells.push({ events: [] });
    for (let d = 1; d <= last.getDate(); d++) {
      const date = new Date(view.getFullYear(), view.getMonth(), d);
      const str = dayStr(date);
      const events = timeOff.filter(
        (t) => isHrRequestStatus(t.status, 'approved') && str >= t.startDate && str <= t.endDate
      );
      cells.push({ date, events });
    }
    return cells;
  }, [view, timeOff]);

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.calendar'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('timeOff.calendar.title')}
          subtitle={t('timeOff.calendar.subtitle')}
          right={
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
              <Button
                icon={<ChevronLeft size={13} />}
                onClick={() => setView(new Date(view.getFullYear(), view.getMonth() - 1, 1))}
              />
              <span
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: 'var(--hr-text)',
                  minWidth: isMobile ? 'auto' : 140,
                  textAlign: 'center',
                }}
              >
                {formatMonth(view)}
              </span>
              <Button
                icon={<ChevronRight size={13} />}
                onClick={() => setView(new Date(view.getFullYear(), view.getMonth() + 1, 1))}
              />
            </div>
          }
        />
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ overflowX: 'auto' }}>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(7, minmax(96px, 1fr))',
                minWidth: 672,
                background: 'var(--hr-hover)',
              }}
            >
              {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((d) => (
                <div
                  key={d}
                  style={{
                    padding: '8px 12px',
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--hr-text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                  }}
                >
                  {d}
                </div>
              ))}
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(7, minmax(96px, 1fr))',
                minWidth: 672,
              }}
            >
              {cells.map((cell, idx) => (
                <div
                  key={idx}
                  style={{
                    minHeight: 100,
                    padding: 8,
                    borderTop: '1px solid var(--hr-border)',
                    borderLeft: idx % 7 === 0 ? 'none' : '1px solid var(--hr-border)',
                    background: cell.date ? 'var(--hr-card)' : 'var(--hr-hover)',
                  }}
                >
                  {cell.date && (
                    <>
                      <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', fontWeight: 600 }}>
                        {cell.date.getDate()}
                      </div>
                      <div
                        style={{ display: 'flex', flexDirection: 'column', gap: 2, marginTop: 4 }}
                      >
                        {cell.events.slice(0, 2).map((e) => (
                          <div
                            key={e.id}
                            title={`${e.staffName} - ${e.type}`}
                            style={{
                              fontSize: 9,
                              padding: '2px 4px',
                              borderRadius: 4,
                              background: `${ACCENT}22`,
                              color: ACCENT,
                              whiteSpace: 'nowrap',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                            }}
                          >
                            {e.staffName}
                          </div>
                        ))}
                        {cell.events.length > 2 && (
                          <div style={{ fontSize: 9, color: 'var(--hr-text-dim)' }}>
                            +{cell.events.length - 2}
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>
    </>
  );
};

export default CalendarScreen;
