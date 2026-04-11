import React from 'react';
import { Sparkles } from 'lucide-react';
import { Breadcrumbs } from "../../shared/components/hub";

interface Entry {
  version: string;
  date: string;
  highlights: { type: 'new' | 'improved' | 'fixed'; text: string }[];
}

const ENTRIES: Entry[] = [
  {
    version: '2026.04.0',
    date: '2026-04-09',
    highlights: [
      { type: 'new',      text: 'Hub command palette (cmd+k) with keyboard shortcuts' },
      { type: 'new',      text: 'Unified Team, Permissions, Roles, Apps screens' },
      { type: 'new',      text: 'API Keys, Webhooks, and Audit Log' },
      { type: 'new',      text: 'Integrations directory with Stripe, Slack, Twilio and more' },
      { type: 'improved', text: 'Onboarding checklist and welcome banner on the dashboard' },
      { type: 'improved', text: 'Notification center with read/unread tracking' },
    ],
  },
  {
    version: '2026.03.2',
    date: '2026-03-28',
    highlights: [
      { type: 'improved', text: 'Faster dashboard loads on slower networks' },
      { type: 'fixed',    text: 'Theme switcher briefly flashing the wrong color' },
    ],
  },
  {
    version: '2026.03.1',
    date: '2026-03-15',
    highlights: [
      { type: 'new',      text: 'Analytics KPIs on the dashboard' },
      { type: 'improved', text: 'Sidebar business context pill' },
    ],
  },
];

const TAG_COLORS: Record<Entry['highlights'][number]['type'], { bg: string; fg: string }> = {
  new:      { bg: 'rgba(69,131,196,0.15)', fg: '#4583C4' },
  improved: { bg: 'rgba(139,92,246,0.15)', fg: '#8b5cf6' },
  fixed:    { bg: 'rgba(16,185,129,0.15)', fg: '#10b981' },
};

const Changelog: React.FC = () => {
  return (
    <div className="space-y-5 max-w-3xl">
      
      <div className="flex items-center gap-3">
        <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(139,92,246,0.15)' }}>
          <Sparkles size={16} style={{ color: '#8b5cf6' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>What&apos;s new</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Product updates and improvements.
          </p>
        </div>
      </div>

      <div className="space-y-4">
        {ENTRIES.map((entry) => (
          <div
            key={entry.version}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>v{entry.version}</p>
              <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>{entry.date}</p>
            </div>
            <ul className="space-y-2">
              {entry.highlights.map((h, i) => {
                const color = TAG_COLORS[h.type];
                return (
                  <li key={i} className="flex items-start gap-2">
                    <span
                      className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase tracking-wider shrink-0 mt-0.5"
                      style={{ background: color.bg, color: color.fg }}
                    >
                      {h.type}
                    </span>
                    <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{h.text}</p>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Changelog;
