import React, { useState } from 'react';
import { motion } from 'framer-motion';
import {
  User,
  Bell,
  Palette,
  Shield,
  Globe,
  Keyboard,
  CreditCard,
  ChevronRight,
} from 'lucide-react';
import { useUserStore } from '../../stores/userStore';
import { useThemeStore } from '../../stores/themeStore';
import { getInitials } from '../../lib/utils';

const ACCENT = '#64C5B9';

interface SettingsSection {
  id: string;
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const SECTIONS: SettingsSection[] = [
  { id: 'profile',       label: 'Profile',          icon: User     },
  { id: 'notifications', label: 'Notifications',    icon: Bell     },
  { id: 'appearance',    label: 'Appearance',       icon: Palette  },
  { id: 'privacy',       label: 'Privacy & Security', icon: Shield },
  { id: 'integrations',  label: 'Integrations',     icon: Globe    },
  { id: 'shortcuts',     label: 'Keyboard Shortcuts', icon: Keyboard},
  { id: 'billing',       label: 'Billing',          icon: CreditCard},
];

const ProfileSection: React.FC = () => {
  const { user } = useUserStore();
  const userName = user?.name || 'User';
  const userEmail = user?.email || '';

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Profile
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Manage your personal information.
        </p>
      </div>

      {/* Avatar */}
      <div className="flex items-center gap-4">
        <span
          className="h-16 w-16 rounded-full flex items-center justify-center text-[20px] font-bold"
          style={{ background: `${ACCENT}22`, color: ACCENT }}
        >
          {getInitials(userName)}
        </span>
        <div>
          <p className="text-[14px] font-medium" style={{ color: 'var(--flow-text)' }}>
            {userName}
          </p>
          <p className="text-[12px] mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
            {userEmail}
          </p>
          <button
            className="mt-2 text-[12px] px-3 py-1.5 rounded-md transition-colors hover:bg-[var(--flow-hover)]"
            style={{ color: ACCENT, border: `1px solid ${ACCENT}44` }}
          >
            Change avatar
          </button>
        </div>
      </div>

      {/* Fields */}
      {[
        { label: 'Full name', value: userName, placeholder: 'Your name' },
        { label: 'Email address', value: userEmail, placeholder: 'your@email.com' },
        { label: 'Display name', value: userName.split(' ')[0], placeholder: 'Display name' },
      ].map((field) => (
        <div key={field.label}>
          <label
            className="block text-[12px] font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {field.label}
          </label>
          <input
            defaultValue={field.value}
            placeholder={field.placeholder}
            className="w-full px-3 py-2 rounded-lg text-[13px] outline-none transition-colors"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
            onFocus={(e) => {
              e.target.style.borderColor = `${ACCENT}66`;
            }}
            onBlur={(e) => {
              e.target.style.borderColor = 'var(--flow-border)';
            }}
          />
        </div>
      ))}

      <button
        className="px-4 py-2 rounded-lg text-[13px] font-medium transition-colors"
        style={{ background: ACCENT, color: '#0a0a0a' }}
      >
        Save changes
      </button>
    </div>
  );
};

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
              onClick={() => toggle(item.key)}
              className="relative h-5 w-9 rounded-full transition-colors shrink-0"
              style={{
                background: settings[item.key] ? ACCENT : 'var(--flow-hover)',
              }}
            >
              <span
                className="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform shadow-sm"
                style={{
                  transform: settings[item.key] ? 'translateX(16px)' : 'translateX(2px)',
                }}
              />
            </button>
          </div>
        ))}
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
  profile:       ProfileSection,
  notifications: NotificationsSection,
  appearance:    AppearanceSection,
  shortcuts:     ShortcutsSection,
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
  const [activeSection, setActiveSection] = useState('profile');

  const ActiveComponent =
    SECTION_CONTENT[activeSection] ||
    (() => <PlaceholderSection id={activeSection} />);

  return (
    <div className="flex h-full">
      {/* Settings sidebar */}
      <div
        className="w-52 shrink-0 overflow-y-auto"
        style={{ borderRight: '1px solid var(--flow-border)' }}
      >
        <div className="p-4">
          <h1
            className="text-[13px] font-semibold mb-4"
            style={{ color: 'var(--flow-text)' }}
          >
            Settings
          </h1>
          <nav className="space-y-0.5">
            {SECTIONS.map((section) => {
              const isActive = activeSection === section.id;
              const Icon = section.icon;
              return (
                <button
                  key={section.id}
                  onClick={() => setActiveSection(section.id)}
                  className="w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-[13px] text-left transition-colors"
                  style={{
                    background: isActive ? 'var(--flow-active)' : 'transparent',
                    color: isActive ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                  }}
                >
                  <Icon size={14} />
                  <span className="flex-1">{section.label}</span>
                  {isActive && (
                    <span
                      className="h-1.5 w-1.5 rounded-full"
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
      <div className="flex-1 overflow-y-auto">
        <motion.div
          key={activeSection}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.18 }}
          className="max-w-xl p-8"
        >
          <ActiveComponent />
        </motion.div>
      </div>
    </div>
  );
};

export default FlowSettings;
