import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Bell, BellRing } from 'lucide-react';
import { cardStyle, ACCENT, ghostBtn } from '../lib/tokens';

// Each row carries the key and the number, so a language decides both the
// wording and where the number sits. The day counts were literal strings
// starting with a digit, which the literal guard does not look for.
const DEFAULT_SCHEDULE = [
  { id: 'r1', labelKey: 'finance.reminders.beforeDue', days: 3, enabled: true, kind: 'pre' },
  { id: 'r2', labelKey: 'finance.reminders.onDueDate', enabled: true, kind: 'on' },
  { id: 'r3', labelKey: 'finance.reminders.overdue', days: 1, enabled: true, kind: 'post' },
  { id: 'r4', labelKey: 'finance.reminders.overdue', days: 7, enabled: false, kind: 'post' },
  { id: 'r5', labelKey: 'finance.reminders.overdue', days: 14, enabled: false, kind: 'post' },
];

const PaymentReminderPanel: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const [schedule, setSchedule] = useState(DEFAULT_SCHEDULE);

  const toggle = (id: string) =>
    setSchedule((prev) => prev.map((r) => (r.id === id ? { ...r, enabled: !r.enabled } : r)));

  return (
    <div style={{ ...cardStyle, padding: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
        <Bell size={14} style={{ color: ACCENT }} />
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, color: 'var(--hub-text)' }}>
          {t('finance.reminders.title')}
        </h3>
      </div>
      <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginBottom: 14 }}>
        {t('finance.reminders.description')}
      </p>
      {schedule.map((r) => (
        <div
          key={r.id}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '10px 0',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <BellRing size={13} style={{ color: r.enabled ? ACCENT : 'var(--hub-text-dim)' }} />
            <span style={{ fontSize: 13, color: 'var(--hub-text)' }}>
              {r.days === undefined ? t(r.labelKey) : t(r.labelKey, { count: r.days })}
            </span>
          </div>
          <button
            type="button"
            onClick={() => toggle(r.id)}
            style={{
              padding: '3px 10px',
              borderRadius: 20,
              fontSize: 11,
              border: 'none',
              cursor: 'pointer',
              background: r.enabled ? 'var(--finance-accent-soft)' : 'var(--hub-hover)',
              color: r.enabled ? ACCENT : 'var(--hub-text-muted)',
              fontWeight: 600,
            }}
          >
            {r.enabled ? 'ON' : 'OFF'}
          </button>
        </div>
      ))}
      <button
        type="button"
        style={{ ...ghostBtn, marginTop: 14 }}
        disabled
        title={t('finance.reminders.comingSoon')}
        onClick={undefined}
      >
        {t('finance.reminders.save')}
      </button>
    </div>
  );
};

export default PaymentReminderPanel;
