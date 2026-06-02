import React from 'react';

const SERVICES = [
  { name: 'API', group: 'Core' },
  { name: 'Web App', group: 'Core' },
  { name: 'Auth', group: 'Core' },
  { name: 'POS', group: 'Apps' },
  { name: 'ERP', group: 'Apps' },
  { name: 'CRM', group: 'Apps' },
  { name: 'HR', group: 'Apps' },
  { name: 'Finance', group: 'Apps' },
  { name: 'Flow', group: 'Apps' },
  { name: 'Payments', group: 'Integrations' },
  { name: 'Webhooks', group: 'Integrations' },
];

const DOT = {
  width: 8,
  height: 8,
  borderRadius: '50%',
  background: '#22c55e',
  flexShrink: 0,
} as const;

export default function StatusWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        background: '#0f0f0f',
        color: '#e5e5e5',
        fontFamily: 'system-ui, sans-serif',
        fontSize: 13,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '20px 24px 16px',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 4 }}>
          <div style={DOT} />
          <span style={{ fontWeight: 600, fontSize: 15, color: '#fff' }}>
            All Systems Operational
          </span>
        </div>
        <div style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>Last checked just now</div>
      </div>

      {/* Services */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '12px 0' }}>
        {(['Core', 'Apps', 'Integrations'] as const).map((group) => (
          <div key={group} style={{ marginBottom: 4 }}>
            <div
              style={{
                padding: '6px 24px',
                fontSize: 11,
                fontWeight: 600,
                color: 'rgba(255,255,255,0.3)',
                letterSpacing: '0.06em',
                textTransform: 'uppercase',
              }}
            >
              {group}
            </div>
            {SERVICES.filter((s) => s.group === group).map((s) => (
              <div
                key={s.name}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '8px 24px',
                  borderBottom: '1px solid rgba(255,255,255,0.04)',
                }}
              >
                <span style={{ color: 'rgba(255,255,255,0.75)' }}>{s.name}</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={DOT} />
                  <span style={{ color: '#22c55e', fontSize: 12 }}>Operational</span>
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>

      {/* Footer */}
      <div
        style={{
          padding: '12px 24px',
          borderTop: '1px solid rgba(255,255,255,0.06)',
          display: 'flex',
          justifyContent: 'flex-end',
        }}
      >
        <a
          href="https://status.rtylr.com"
          target="_blank"
          rel="noopener noreferrer"
          style={{
            fontSize: 12,
            color: 'rgba(255,255,255,0.5)',
            textDecoration: 'none',
            padding: '6px 12px',
            border: '1px solid rgba(255,255,255,0.12)',
            borderRadius: 6,
            transition: 'color 0.15s, border-color 0.15s',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLAnchorElement).style.color = '#fff';
            (e.currentTarget as HTMLAnchorElement).style.borderColor = 'rgba(255,255,255,0.3)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLAnchorElement).style.color = 'rgba(255,255,255,0.5)';
            (e.currentTarget as HTMLAnchorElement).style.borderColor = 'rgba(255,255,255,0.12)';
          }}
        >
          Open status.rtylr.com
        </a>
      </div>
    </div>
  );
}
