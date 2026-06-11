import React from 'react';

const SHORTCUTS = [
  { keys: ['⌘', 'K'], desc: 'Toggle Spotlight search' },
  { keys: ['Shift', 'M'], desc: 'Switch to Website mode' },
  { keys: ['Shift', 'X'], desc: 'Close all windows' },
  { keys: ['Shift', 'W'], desc: 'Close focused window' },
  { keys: ['Shift', 'H'], desc: 'Minimize focused window' },
  { keys: ['⌘', '/'], desc: 'Open Keyboard Shortcuts' },
  { keys: ['Right-click'], desc: 'Desktop context menu' },
];

export default function ShortcutsWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        overflowY: 'auto',
        background: '#151515',
        color: '#e8e6e1',
        fontFamily: "'Inter', system-ui, sans-serif",
        padding: '32px 36px',
      }}
    >
      <div
        style={{
          fontSize: 11,
          fontWeight: 600,
          letterSpacing: '0.1em',
          textTransform: 'uppercase',
          color: 'rgba(255,255,255,0.35)',
          marginBottom: 20,
        }}
      >
        Keyboard Shortcuts
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {SHORTCUTS.map((s) => (
          <div
            key={s.desc}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '10px 14px',
              background: 'rgba(255,255,255,0.04)',
              border: '1px solid rgba(255,255,255,0.07)',
              borderRadius: 8,
            }}
          >
            <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.65)' }}>{s.desc}</span>
            <div style={{ display: 'flex', gap: 4 }}>
              {s.keys.map((k) => (
                <kbd
                  key={k}
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '3px 8px',
                    background: 'rgba(255,255,255,0.08)',
                    border: '1px solid rgba(255,255,255,0.15)',
                    borderRadius: 5,
                    color: '#e8e6e1',
                    fontFamily: 'inherit',
                  }}
                >
                  {k}
                </kbd>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
