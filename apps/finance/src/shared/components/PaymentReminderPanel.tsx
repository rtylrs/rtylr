import React, { useState } from 'react';
import { Bell, BellRing } from 'lucide-react';
import { cardStyle, ACCENT, ghostBtn } from '../lib/tokens';

const DEFAULT_SCHEDULE = [
  { id: 'r1', label: '3 days before due', enabled: true, kind: 'pre' },
  { id: 'r2', label: 'On due date', enabled: true, kind: 'on' },
  { id: 'r3', label: '1 day overdue', enabled: true, kind: 'post' },
  { id: 'r4', label: '7 days overdue', enabled: false, kind: 'post' },
  { id: 'r5', label: '14 days overdue', enabled: false, kind: 'post' },
];

const PaymentReminderPanel: React.FC = () => {
  const [schedule, setSchedule] = useState(DEFAULT_SCHEDULE);

  const toggle = (id: string) =>
    setSchedule((prev) => prev.map((r) => (r.id === id ? { ...r, enabled: !r.enabled } : r)));

  return (
    <div style={{ ...cardStyle, padding: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
        <Bell size={14} style={{ color: ACCENT }} />
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, color: 'var(--hub-text)' }}>
          Payment Reminders
        </h3>
      </div>
      <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginBottom: 14 }}>
        Automated emails sent to clients when invoices are approaching or past their due date.
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
            <span style={{ fontSize: 13, color: 'var(--hub-text)' }}>{r.label}</span>
          </div>
          <button
            onClick={() => toggle(r.id)}
            style={{
              padding: '3px 10px',
              borderRadius: 20,
              fontSize: 11,
              border: 'none',
              cursor: 'pointer',
              background: r.enabled ? 'rgba(16,185,129,0.12)' : 'var(--hub-hover)',
              color: r.enabled ? ACCENT : 'var(--hub-text-muted)',
              fontWeight: 600,
            }}
          >
            {r.enabled ? 'ON' : 'OFF'}
          </button>
        </div>
      ))}
      <button
        style={{ ...ghostBtn, marginTop: 14 }}
        disabled
        title="Reminder schedule coming in a future update"
        onClick={undefined}
      >
        Save schedule
      </button>
    </div>
  );
};

export default PaymentReminderPanel;
