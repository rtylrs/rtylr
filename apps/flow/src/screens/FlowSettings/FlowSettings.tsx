import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Bell, Palette, Keyboard, Sparkles } from 'lucide-react';
import { useThemeStore } from '../../stores/themeStore';
import { cn } from '../../lib/utils';
import { useLifePreferences } from '../../lib/lifePreferences';

const ACCENT = '#64C5B9';

interface SettingsSection {
  id: string;
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const SECTIONS: SettingsSection[] = [
  { id: 'notifications', label: 'Notifications', icon: Bell },
  { id: 'appearance', label: 'Appearance', icon: Palette },
  { id: 'life', label: 'Life & rhythm', icon: Sparkles },
  { id: 'shortcuts', label: 'Keyboard Shortcuts', icon: Keyboard },
];

const NotificationsSection: React.FC = () => {
  const [settings, setSettings] = useState({
    taskAssigned: true,
    taskComments: true,
    mentions: true,
    statusChanges: false,
    weeklyDigest: true,
    emailNotifications: false,
  });

  const toggle = (key: keyof typeof settings) => {
    setSettings((s) => ({ ...s, [key]: !s[key] }));
  };

  const items = [
    { key: 'taskAssigned' as const,      label: 'Task assigned to me',    desc: 'When someone assigns a task to you' },
    { key: 'taskComments' as const,      label: 'Comments on my tasks',   desc: 'When someone comments on your tasks' },
    { key: 'mentions' as const,          label: 'Mentions',               desc: 'When someone @mentions you' },
    { key: 'statusChanges' as const,     label: 'Status changes',         desc: 'When a task you follow changes status' },
    { key: 'weeklyDigest' as const,      label: 'Weekly digest',          desc: 'A weekly summary of your activity' },
    { key: 'emailNotifications' as const,label: 'Email notifications',    desc: 'Receive notifications via email' },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Notifications
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Control what you get notified about.
        </p>
      </div>

      <div className="space-y-1">
        {items.map((item) => (
          <div
            key={item.key}
            className="flex items-center justify-between py-3 px-4 rounded-lg hover:bg-[var(--flow-hover)] transition-colors"
          >
            <div>
              <p className="text-[13px] font-medium" style={{ color: 'var(--flow-text)' }}>
                {item.label}
              </p>
              <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
                {item.desc}
              </p>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={settings[item.key]}
              onClick={() => toggle(item.key)}
              className={cn(
                'flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full p-0.5',
                'transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[#64C5B9]/45',
                'focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--flow-bg)]'
              )}
              style={{
                background: settings[item.key] ? ACCENT : 'var(--flow-hover)',
              }}
            >
              <span
                className={cn(
                  'pointer-events-none h-4 w-4 shrink-0 rounded-full bg-white shadow-sm',
                  'transition-[margin] duration-200 ease-out',
                  settings[item.key] ? 'ml-auto' : 'ml-0'
                )}
              />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
};

const LifeRhythmSection: React.FC = () => {
  const { prefs, setPrefs } = useLifePreferences();

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Life & rhythm
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Quiet hours affect in-app banners only for now; digest hour is saved for when scheduled
          reminders ship.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label
            className="block text-[12px] font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Quiet hours start
          </label>
          <input
            type="time"
            value={prefs.quietStart}
            onChange={(e) => setPrefs({ quietStart: e.target.value })}
            className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
          />
        </div>
        <div>
          <label
            className="block text-[12px] font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Quiet hours end
          </label>
          <input
            type="time"
            value={prefs.quietEnd}
            onChange={(e) => setPrefs({ quietEnd: e.target.value })}
            className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
          />
        </div>
      </div>

      <div>
        <label
          className="block text-[12px] font-medium mb-1.5"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          Preferred digest hour (local)
        </label>
        <select
          value={prefs.digestHour}
          onChange={(e) => setPrefs({ digestHour: Number(e.target.value) })}
          className="w-full max-w-[200px] rounded-lg px-3 py-2 text-[13px] outline-none"
          style={{
            background: 'var(--flow-input-bg)',
            border: '1px solid var(--flow-border)',
            color: 'var(--flow-text)',
          }}
        >
          {Array.from({ length: 24 }, (_, i) => (
            <option key={i} value={i}>
              {String(i).padStart(2, '0')}:00
            </option>
          ))}
        </select>
      </div>
    </div>
  );
};

const AppearanceSection: React.FC = () => {
  const { theme, setTheme } = useThemeStore();

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Appearance
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Customize the look and feel of Flow.
        </p>
      </div>

      <div>
        <label className="block text-[12px] font-medium mb-3" style={{ color: 'var(--flow-text-muted)' }}>
          Theme
        </label>
        <div className="grid grid-cols-3 gap-3">
          {(
            [
              { value: 'light', label: 'Light', preview: '#f4f4f5' },
              { value: 'dark',  label: 'Dark',  preview: '#0a0a0a' },
              { value: 'system',label: 'System', preview: 'linear-gradient(135deg, #f4f4f5 50%, #0a0a0a 50%)' },
            ] as const
          ).map(({ value, label, preview }) => (
            <button
              key={value}
              onClick={() => setTheme(value)}
              className="p-3 rounded-xl border-2 text-[12px] font-medium transition-all text-center"
              style={{
                borderColor: theme === value ? ACCENT : 'var(--flow-border)',
                background: 'var(--flow-card)',
                color: theme === value ? ACCENT : 'var(--flow-text-muted)',
              }}
            >
              <div
                className="w-full h-12 rounded-lg mb-2"
                style={{ background: preview }}
              />
              {label}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label className="block text-[12px] font-medium mb-1.5" style={{ color: 'var(--flow-text-muted)' }}>
          Accent color
        </label>
        <div className="flex items-center gap-2">
          <span
            className="h-8 w-8 rounded-full"
            style={{ background: ACCENT }}
          />
          <span className="text-[13px] font-mono" style={{ color: 'var(--flow-text)' }}>
            {ACCENT}
          </span>
          <span className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            (Flow teal, from Voxire logo)
          </span>
        </div>
      </div>
    </div>
  );
};

const ShortcutsSection: React.FC = () => {
  const shortcuts = [
    { keys: ['C'], desc: 'Create a new task' },
    { keys: ['/'], desc: 'Open command bar / search' },
    { keys: ['⌘', 'K'], desc: 'Open command bar' },
    { keys: ['G', 'I'], desc: 'Go to Inbox' },
    { keys: ['G', 'M'], desc: 'Go to My Issues' },
    { keys: ['G', 'F'], desc: 'Go to Today (focus)' },
    { keys: ['G', 'D'], desc: 'Go to Dashboard' },
    { keys: ['G', 'C'], desc: 'Go to Calendar' },
    { keys: ['G', 'T'], desc: 'Go to Timeline' },
    { keys: ['G', 'R'], desc: 'Go to Roadmap' },
    { keys: ['G', 'S'], desc: 'Go to Insights' },
    { keys: ['G', 'Y'], desc: 'Go to Cycles' },
    { keys: ['G', 'X'], desc: 'Go to Triage' },
    { keys: ['Esc'], desc: 'Close any modal' },
    { keys: ['⌘', '↵'], desc: 'Submit comment' },
  ];
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Keyboard Shortcuts
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Move fast through Flow without touching the mouse.
        </p>
      </div>
      <div className="space-y-1">
        {shortcuts.map((s, i) => (
          <div
            key={i}
            className="flex items-center justify-between py-2 px-3 rounded hover:bg-[var(--flow-hover)]"
          >
            <span className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
              {s.desc}
            </span>
            <div className="flex items-center gap-1">
              {s.keys.map((k, j) => (
                <kbd
                  key={j}
                  className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold"
                  style={{
                    background: 'var(--flow-hover)',
                    color: 'var(--flow-text)',
                    border: '1px solid var(--flow-border)',
                    minWidth: 22,
                    textAlign: 'center',
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
};

const SECTION_CONTENT: Record<string, React.FC> = {
  notifications: NotificationsSection,
  appearance: AppearanceSection,
  life: LifeRhythmSection,
  shortcuts: ShortcutsSection,
};

const PlaceholderSection: React.FC<{ id: string }> = ({ id }) => (
  <div className="flex flex-col items-center justify-center py-20 text-center">
    <p className="text-[14px] font-medium mb-1" style={{ color: 'var(--flow-text-muted)' }}>
      {SECTIONS.find((s) => s.id === id)?.label}
    </p>
    <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
      This section is coming soon.
    </p>
  </div>
);

const FlowSettings: React.FC = () => {
  const [activeSection, setActiveSection] = useState('notifications');

  const ActiveComponent =
    SECTION_CONTENT[activeSection] ||
    (() => <PlaceholderSection id={activeSection} />);

  return (
    <div className="flex h-full min-h-0 flex-col md:flex-row">
      {/* Settings sidebar */}
      <div
        className="w-full shrink-0 overflow-x-auto overflow-y-hidden border-b border-[var(--flow-border)] md:w-[236px] md:border-b-0 md:border-r md:border-[var(--flow-border)] md:overflow-y-auto"
      >
        <div className="flex flex-row gap-1 p-3 md:flex-col md:gap-0 md:p-4">
          <h1
            className="mb-3 hidden text-[13px] font-semibold md:mb-4 md:block"
            style={{ color: 'var(--flow-text)' }}
          >
            Settings
          </h1>
          <nav className="flex min-w-max flex-row gap-1 md:min-w-0 md:flex-col md:space-y-0.5">
            {SECTIONS.map((section) => {
              const isActive = activeSection === section.id;
              const Icon = section.icon;
              return (
                <button
                  key={section.id}
                  type="button"
                  onClick={() => setActiveSection(section.id)}
                  className="flex shrink-0 items-center gap-2 whitespace-nowrap rounded-md px-3 py-2 text-left text-[13px] transition-colors md:w-full md:min-w-0 md:whitespace-normal"
                  style={{
                    background: isActive ? 'var(--flow-active)' : 'transparent',
                    color: isActive ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                  }}
                >
                  <Icon size={14} className="shrink-0" />
                  <span className="flex-1 min-w-0 truncate">{section.label}</span>
                  {isActive && (
                    <span
                      className="h-1.5 w-1.5 rounded-full shrink-0"
                      style={{ background: ACCENT }}
                    />
                  )}
                </button>
              );
            })}
          </nav>
        </div>
      </div>

      {/* Settings content */}
      <div className="min-h-0 flex-1 overflow-y-auto">
        <motion.div
          key={activeSection}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.18 }}
          className="max-w-xl p-4 sm:p-8"
        >
          <ActiveComponent />
        </motion.div>
      </div>
    </div>
  );
};

export default FlowSettings;
