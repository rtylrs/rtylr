import React from 'react';
import { CARET_SHORTCUTS } from '../../shared/lib/caret/keymap';
import { PageHeader } from './components';

/** Reference page listing every Caret keyboard shortcut. */
const CaretShortcutsPage: React.FC = () => (
  <div className="space-y-4 max-w-2xl">
    <PageHeader title="Caret shortcuts" subtitle="Keyboard combos that drive Caret." />
    <div
      style={{
        background: 'var(--dash-surface)',
        border: '1px solid var(--dash-border)',
        borderRadius: 10,
        overflow: 'hidden',
      }}
    >
      <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
        {CARET_SHORTCUTS.map((s, idx) => (
          <li
            key={s.id}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '10px 14px',
              borderTop: idx === 0 ? 'none' : '1px solid var(--dash-border)',
            }}
          >
            <span style={{ fontSize: 13 }}>{s.description}</span>
            <span style={{ display: 'inline-flex', gap: 4 }}>
              {s.keys.map((k, i) => (
                <kbd
                  key={i}
                  style={{
                    fontFamily: 'inherit',
                    fontSize: 11,
                    padding: '2px 6px',
                    borderRadius: 4,
                    background: 'var(--dash-active)',
                    color: 'var(--dash-text)',
                    border: '1px solid var(--dash-border)',
                  }}
                >
                  {k}
                </kbd>
              ))}
            </span>
          </li>
        ))}
      </ul>
    </div>
  </div>
);

export default CaretShortcutsPage;
