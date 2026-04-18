import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import {
  Bell,
  Palette,
  Keyboard,
  Sparkles,
  Info,
  User,
  Save,
  Tag,
  Trash2,
  Users,
  CreditCard,
  UserMinus,
  UserPlus,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import {
  useFlowTeam,
  useFlowInviteMember,
  useFlowDeactivateMember,
} from '../../shared/hooks/queries/useFlowTeam';
import {
  useFlowSubscription,
  useFlowBillingDetail,
  centsToDisplay,
  trialDaysLeft,
  formatPeriodRange,
  computeSeatProration,
} from '../../shared/hooks/queries/useFlowSubscription';
import appConfig from '../../app.json';
import { useThemeStore } from '../../shared/stores/themeStore';
import { cn } from '../../shared/lib/utils';
import { useLifePreferences } from '../../shared/lib/lifePreferences';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiUserSettings } from '../../shared/services/flowApi';
import { useUserStore } from '../../shared/stores/userStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useLabels, useCreateLabel, useDeleteLabel } from '../../features/tasks/useLabels';

const ACCENT = '#64C5B9';

interface SettingsSection {
  id: string;
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const SECTIONS: SettingsSection[] = [
  { id: 'profile', label: 'Profile', icon: User },
  { id: 'team', label: 'Team', icon: Users },
  { id: 'billing', label: 'Billing', icon: CreditCard },
  { id: 'notifications', label: 'Notifications', icon: Bell },
  { id: 'appearance', label: 'Appearance', icon: Palette },
  { id: 'labels', label: 'Labels', icon: Tag },
  { id: 'life', label: 'Life & rhythm', icon: Sparkles },
  { id: 'shortcuts', label: 'Keyboard Shortcuts', icon: Keyboard },
  { id: 'about', label: 'About', icon: Info },
];

const ProfileSection: React.FC = () => {
  const user = useUserStore((s) => s.user);

  const [name, setName] = useState(user?.name || '');
  const [email, setEmail] = useState(user?.email || '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');

  useEffect(() => {
    if (user) {
      setName(user.name || '');
      setEmail(user.email || '');
    }
  }, [user]);

  const updateProfile = useMutation({
    mutationFn: (data: { name?: string; email?: string }) => unwrap(flowApi.updateProfile(data)),
    onSuccess: () => toast.success('Profile updated'),
    onError: () => toast.error('Failed to update profile'),
  });

  const changePassword = useMutation({
    mutationFn: (data: { currentPassword: string; newPassword: string }) =>
      unwrap(flowApi.changePassword(data)),
    onSuccess: () => {
      toast.success('Password changed');
      setCurrentPassword('');
      setNewPassword('');
    },
    onError: () => toast.error('Failed to change password'),
  });

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    color: 'var(--flow-text)',
  };

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Profile
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Update your name and email address.
        </p>
      </div>

      <div className="space-y-4">
        <div>
          <label
            className="block text-[12px] font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
            style={inputStyle}
          />
        </div>
        <div>
          <label
            className="block text-[12px] font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
            style={inputStyle}
          />
        </div>
        <button
          type="button"
          disabled={updateProfile.isPending}
          onClick={() => updateProfile.mutate({ name, email })}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
          style={{ background: ACCENT, color: '#0a0a0a' }}
        >
          <Save size={12} />
          {updateProfile.isPending ? 'Saving…' : 'Save profile'}
        </button>
      </div>

      <div className="pt-6" style={{ borderTop: '1px solid var(--flow-border)' }}>
        <h3 className="text-[13px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Change password
        </h3>
        <p className="text-[12px] mb-4" style={{ color: 'var(--flow-text-muted)' }}>
          Leave blank to keep your current password.
        </p>
        <div className="space-y-4">
          <div>
            <label
              className="block text-[12px] font-medium mb-1.5"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Current password
            </label>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
              style={inputStyle}
            />
          </div>
          <div>
            <label
              className="block text-[12px] font-medium mb-1.5"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              New password
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
              style={inputStyle}
            />
          </div>
          <button
            type="button"
            disabled={changePassword.isPending || !currentPassword || !newPassword}
            onClick={() => changePassword.mutate({ currentPassword, newPassword })}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
          >
            {changePassword.isPending ? 'Changing…' : 'Change password'}
          </button>
        </div>
      </div>
    </div>
  );
};

const NotificationsSection: React.FC = () => {
  const { data: remote } = useQuery({
    queryKey: ['flow-settings'],
    queryFn: () => unwrap(flowApi.getUserSettings()),
  });

  const save = useMutation({
    mutationFn: (data: ApiUserSettings) => unwrap(flowApi.updateUserSettings(data)),
  });

  const [settings, setSettings] = useState<ApiUserSettings>({
    taskAssigned: true,
    taskComments: true,
    mentions: true,
    statusChanges: false,
    weeklyDigest: true,
    emailNotifications: false,
  });

  useEffect(() => {
    if (remote) setSettings(remote);
  }, [remote]);

  const toggle = (key: keyof ApiUserSettings) => {
    const next = { ...settings, [key]: !settings[key] };
    setSettings(next);
    save.mutate(next);
  };

  const items = [
    {
      key: 'taskAssigned' as const,
      label: 'Task assigned to me',
      desc: 'When someone assigns a task to you',
    },
    {
      key: 'taskComments' as const,
      label: 'Comments on my tasks',
      desc: 'When someone comments on your tasks',
    },
    { key: 'mentions' as const, label: 'Mentions', desc: 'When someone @mentions you' },
    {
      key: 'statusChanges' as const,
      label: 'Status changes',
      desc: 'When a task you follow changes status',
    },
    {
      key: 'weeklyDigest' as const,
      label: 'Weekly digest',
      desc: 'A weekly summary of your activity',
    },
    {
      key: 'emailNotifications' as const,
      label: 'Email notifications',
      desc: 'Receive notifications via email',
    },
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
        <label
          className="block text-[12px] font-medium mb-3"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          Theme
        </label>
        <div className="grid grid-cols-3 gap-3">
          {(
            [
              { value: 'light', label: 'Light', preview: '#f4f4f5' },
              { value: 'dark', label: 'Dark', preview: '#0a0a0a' },
              {
                value: 'system',
                label: 'System',
                preview: 'linear-gradient(135deg, #f4f4f5 50%, #0a0a0a 50%)',
              },
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
              <div className="w-full h-12 rounded-lg mb-2" style={{ background: preview }} />
              {label}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label
          className="block text-[12px] font-medium mb-1.5"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          Accent color
        </label>
        <div className="flex items-center gap-2">
          <span className="h-8 w-8 rounded-full" style={{ background: ACCENT }} />
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

const AboutSection: React.FC = () => (
  <div className="space-y-6">
    <div>
      <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
        About
      </h2>
      <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
        Application information.
      </p>
    </div>
    <div className="rounded-xl divide-y" style={{ border: '1px solid var(--flow-border)' }}>
      {[
        { label: 'Application', value: appConfig.name },
        { label: 'Version', value: `v${appConfig.version}` },
      ].map(({ label, value }) => (
        <div key={label} className="flex items-center justify-between px-4 py-3">
          <span className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
            {label}
          </span>
          <span className="text-[13px] font-medium font-mono" style={{ color: 'var(--flow-text)' }}>
            {value}
          </span>
        </div>
      ))}
    </div>
  </div>
);

const LABEL_COLORS = [
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#64C5B9',
];

const LabelsSection: React.FC = () => {
  const { activeWorkspaceId } = useFlowStore();
  const workspaceId = activeWorkspaceId ?? '';
  const { data: labels = [], isLoading } = useLabels(workspaceId || undefined);
  const createLabel = useCreateLabel();
  const deleteLabel = useDeleteLabel();

  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState(LABEL_COLORS[0]);

  const handleCreate = () => {
    const name = newName.trim();
    if (!name) return;
    if (!workspaceId) {
      toast.error('No active workspace');
      return;
    }
    createLabel.mutate(
      { workspace_id: Number(workspaceId), name, color: newColor },
      {
        onSuccess: () => {
          setNewName('');
          setNewColor(LABEL_COLORS[0]);
          toast.success('Label created');
        },
        onError: () => toast.error('Failed to create label'),
      }
    );
  };

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    color: 'var(--flow-text)',
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Labels
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Manage labels for your workspace. Labels can be applied to any task.
        </p>
      </div>

      {/* Create form */}
      <div
        className="rounded-xl p-4 space-y-3"
        style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
      >
        <h3 className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          New label
        </h3>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder="Label name"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            className="flex-1 rounded-lg px-3 py-2 text-[13px] outline-none"
            style={inputStyle}
          />
        </div>
        <div>
          <p className="text-[11px] mb-2" style={{ color: 'var(--flow-text-muted)' }}>
            Color
          </p>
          <div className="flex flex-wrap gap-2">
            {LABEL_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setNewColor(c)}
                className="h-6 w-6 rounded-full transition-all"
                style={{
                  background: c,
                  outline: newColor === c ? `2px solid ${c}` : 'none',
                  outlineOffset: 2,
                  boxShadow: newColor === c ? '0 0 0 2px var(--flow-bg)' : 'none',
                }}
                title={c}
              />
            ))}
          </div>
        </div>
        <button
          type="button"
          disabled={!newName.trim() || createLabel.isPending}
          onClick={handleCreate}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
          style={{ background: ACCENT, color: '#0a0a0a' }}
        >
          {createLabel.isPending ? 'Creating…' : 'Create label'}
        </button>
      </div>

      {/* Existing labels */}
      <div>
        <h3 className="text-[12px] font-semibold mb-3" style={{ color: 'var(--flow-text-muted)' }}>
          Workspace labels
        </h3>
        {isLoading ? (
          <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
            Loading…
          </p>
        ) : labels.length === 0 ? (
          <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
            No labels yet. Create one above.
          </p>
        ) : (
          <div
            className="rounded-xl divide-y divide-[var(--flow-border)]"
            style={{ border: '1px solid var(--flow-border)' }}
          >
            {labels.map((label) => (
              <div
                key={label.id}
                className="flex items-center justify-between px-4 py-3 hover:bg-[var(--flow-hover)] transition-colors"
              >
                <div className="flex items-center gap-3">
                  <span
                    className="h-3 w-3 rounded-full shrink-0"
                    style={{ background: label.color }}
                  />
                  <span className="text-[13px] font-medium" style={{ color: 'var(--flow-text)' }}>
                    {label.name}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    deleteLabel.mutate(label.id, {
                      onSuccess: () => toast.success('Label deleted'),
                      onError: () => toast.error('Failed to delete label'),
                    })
                  }
                  className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-dim)' }}
                  title="Delete label"
                >
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

// ── Team Section ─────────────────────────────────────────────────────────────

const TeamSection: React.FC = () => {
  const { data: members = [], isLoading } = useFlowTeam();
  const invite = useFlowInviteMember();
  const deactivate = useFlowDeactivateMember();
  const { data: sub } = useFlowSubscription();

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [showInvite, setShowInvite] = useState(false);

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    color: 'var(--flow-text)',
  };

  const handleInvite = (): void => {
    if (!email.trim() || !fullName.trim()) return;
    invite.mutate(
      { email: email.trim(), fullName: fullName.trim(), role: 'member' },
      {
        onSuccess: () => {
          setEmail('');
          setFullName('');
          setShowInvite(false);
        },
      }
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Team
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          People who have access to this Flow workspace.
        </p>
      </div>

      {/* Invite form */}
      {showInvite ? (
        <div
          className="rounded-xl p-4 space-y-3"
          style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
        >
          <h3 className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            Invite someone to Flow
          </h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                className="block text-[11px] font-medium mb-1"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                Full name
              </label>
              <input
                type="text"
                placeholder="Jane Smith"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
                style={inputStyle}
              />
            </div>
            <div>
              <label
                className="block text-[11px] font-medium mb-1"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                Email
              </label>
              <input
                type="email"
                placeholder="jane@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleInvite()}
                className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
                style={inputStyle}
              />
            </div>
          </div>
          {sub &&
            sub.pricePerSeatCents > 0 &&
            (() => {
              const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                computeSeatProration(sub.pricePerSeatCents);
              return (
                <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                  Adding this member costs{' '}
                  <span className="font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {centsToDisplay(proratedCents)}
                  </span>{' '}
                  this month ({daysRemaining}/{daysInMonth} days), then{' '}
                  <span className="font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {centsToDisplay(sub.pricePerSeatCents)}/mo
                  </span>{' '}
                  from {nextMonthName} 1st.
                </p>
              );
            })()}
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!email.trim() || !fullName.trim() || invite.isPending}
              onClick={handleInvite}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
              style={{ background: ACCENT, color: '#0a0a0a' }}
            >
              <UserPlus size={12} />
              {invite.isPending ? 'Sending…' : 'Send invite'}
            </button>
            <button
              type="button"
              onClick={() => setShowInvite(false)}
              className="px-3 py-1.5 rounded-lg text-[12px] font-medium"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => setShowInvite(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold"
          style={{ background: ACCENT, color: '#0a0a0a' }}
        >
          <UserPlus size={12} />
          Invite member
        </button>
      )}

      {/* Members list */}
      {isLoading ? (
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-14 rounded-xl animate-pulse"
              style={{ background: 'var(--flow-hover)' }}
            />
          ))}
        </div>
      ) : members.length === 0 ? (
        <div className="py-8 text-center" style={{ color: 'var(--flow-text-dim)' }}>
          <Users size={24} className="mx-auto mb-2 opacity-40" />
          <p className="text-[13px]">No team members yet.</p>
        </div>
      ) : (
        <div
          className="rounded-xl divide-y divide-[var(--flow-border)]"
          style={{ border: '1px solid var(--flow-border)' }}
        >
          {members.map((m) => (
            <div key={m.id} className="flex items-center gap-3 px-4 py-3">
              <div
                className="h-8 w-8 rounded-full flex items-center justify-center text-[12px] font-semibold shrink-0"
                style={{ background: ACCENT + '33', color: ACCENT }}
              >
                {m.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <p
                  className="text-[13px] font-medium truncate"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {m.name}
                </p>
                <p className="text-[11px] truncate" style={{ color: 'var(--flow-text-muted)' }}>
                  {m.email}
                </p>
              </div>
              <span
                className="text-[10px] font-semibold px-2 py-0.5 rounded-full shrink-0"
                style={{
                  background: 'var(--flow-hover)',
                  color: 'var(--flow-text-muted)',
                }}
              >
                {m.role}
              </span>
              <button
                type="button"
                aria-label={`Remove ${m.name}`}
                onClick={() => deactivate.mutate(m.id)}
                className="h-7 w-7 flex items-center justify-center rounded hover:bg-[var(--flow-hover)] transition-colors shrink-0"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <UserMinus size={13} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// ── Billing Section ───────────────────────────────────────────────────────────

const STATUS_ICON: Record<string, React.ComponentType<{ size?: number }>> = {
  active: CheckCircle2,
  trial: Clock,
  cancelled: XCircle,
  expired: AlertCircle,
};

const STATUS_COLOR: Record<string, string> = {
  active: 'var(--status-active, #16a34a)',
  trial: 'var(--status-trial, #d97706)',
  cancelled: 'var(--status-cancelled, #dc2626)',
  expired: 'var(--status-expired, #6b7280)',
};

const BillingSection: React.FC = () => {
  const { data: sub, isLoading: subLoading } = useFlowSubscription();
  const { data: billing, isLoading: billingLoading } = useFlowBillingDetail();

  const isLoading = subLoading || billingLoading;

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div
          className="h-6 w-32 rounded animate-pulse"
          style={{ background: 'var(--flow-hover)' }}
        />
        <div
          className="h-28 rounded-xl animate-pulse"
          style={{ background: 'var(--flow-hover)' }}
        />
        <div
          className="h-20 rounded-xl animate-pulse"
          style={{ background: 'var(--flow-hover)' }}
        />
      </div>
    );
  }

  const StatusIcon = sub ? (STATUS_ICON[sub.status] ?? AlertCircle) : null;
  const statusColor = sub ? (STATUS_COLOR[sub.status] ?? STATUS_COLOR.expired) : '';

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Billing
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Your Flow subscription and usage this month.
        </p>
      </div>

      {!sub ? (
        <div
          className="rounded-xl p-6 text-center space-y-3"
          style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
        >
          <CreditCard
            size={24}
            className="mx-auto opacity-30"
            style={{ color: 'var(--flow-text)' }}
          />
          <p className="text-[13px] font-medium" style={{ color: 'var(--flow-text)' }}>
            No active subscription
          </p>
          <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
            Contact your account owner or visit billing to activate Flow.
          </p>
        </div>
      ) : (
        <>
          {/* Plan overview */}
          <div
            className="rounded-xl p-4 space-y-3"
            style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
          >
            <div className="flex items-center justify-between">
              <p className="text-[14px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                {sub.planName}
              </p>
              {StatusIcon && (
                <div
                  className="flex items-center gap-1.5 text-[11px] font-semibold"
                  style={{ color: statusColor }}
                >
                  <StatusIcon size={12} />
                  <span className="capitalize">{sub.status}</span>
                </div>
              )}
            </div>

            <div className="flex items-end justify-between">
              <div>
                <p className="text-[22px] font-bold" style={{ color: 'var(--flow-text)' }}>
                  {centsToDisplay(billing?.totalCents ?? sub.seatCount * sub.pricePerSeatCents)}
                </p>
                <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                  {sub.seatCount} seat{sub.seatCount !== 1 ? 's' : ''} ×{' '}
                  {centsToDisplay(sub.pricePerSeatCents)}/seat · billed 1st of each month
                </p>
              </div>
              {sub.status === 'trial' && (
                <div
                  className="text-[11px] font-semibold px-2.5 py-1 rounded-lg shrink-0"
                  style={{
                    background: 'rgba(245,158,11,0.12)',
                    color: '#d97706',
                  }}
                >
                  {trialDaysLeft(sub.trialEndsAt)}d left in trial
                </div>
              )}
            </div>
          </div>

          {/* Current month line items */}
          {billing && billing.lineItems.length > 0 && (
            <div
              className="rounded-xl overflow-hidden"
              style={{ border: '1px solid var(--flow-border)' }}
            >
              <div
                className="px-4 py-2.5 text-[11px] font-semibold uppercase tracking-wider"
                style={{
                  background: 'var(--flow-hover)',
                  color: 'var(--flow-text-dim)',
                  borderBottom: '1px solid var(--flow-border)',
                }}
              >
                This month
              </div>
              {billing.lineItems.map((item, i) => (
                <div
                  key={i}
                  className="px-4 py-3 flex items-center justify-between gap-4"
                  style={{
                    borderBottom:
                      i < billing.lineItems.length - 1 ? '1px solid var(--flow-border)' : 'none',
                  }}
                >
                  <div>
                    <p className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
                      {item.seatCount} seat{item.seatCount !== 1 ? 's' : ''} ·{' '}
                      {centsToDisplay(item.pricePerSeatCents)}/seat/mo
                    </p>
                    <p
                      className="text-[11px] font-mono mt-0.5"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      {formatPeriodRange(item.periodStart, item.periodEnd)}
                      {!item.periodEnd && (
                        <span
                          className="ml-1.5 font-sans text-[10px] px-1.5 py-0.5 rounded-full"
                          style={{ background: 'rgba(245,158,11,0.12)', color: '#d97706' }}
                        >
                          in progress
                        </span>
                      )}
                    </p>
                  </div>
                  <p className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {centsToDisplay(item.amountCents)}
                  </p>
                </div>
              ))}
              <div
                className="px-4 py-3 flex items-center justify-between"
                style={{
                  background: 'var(--flow-hover)',
                  borderTop: '1px solid var(--flow-border)',
                }}
              >
                <p className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                  {sub.status === 'trial'
                    ? 'Estimated total (trial, not charged)'
                    : 'Estimated total'}
                </p>
                <p className="text-[15px] font-bold" style={{ color: 'var(--flow-text)' }}>
                  {centsToDisplay(billing.totalCents)}
                </p>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

// ── Section content map ───────────────────────────────────────────────────────

const SECTION_CONTENT: Record<string, React.FC> = {
  profile: ProfileSection,
  team: TeamSection,
  billing: BillingSection,
  notifications: NotificationsSection,
  appearance: AppearanceSection,
  labels: LabelsSection,
  life: LifeRhythmSection,
  shortcuts: ShortcutsSection,
  about: AboutSection,
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
    SECTION_CONTENT[activeSection] || (() => <PlaceholderSection id={activeSection} />);

  return (
    <div className="flex h-full min-h-0 flex-col md:flex-row">
      {/* Settings sidebar */}
      <div className="w-full shrink-0 overflow-x-auto overflow-y-hidden border-b border-[var(--flow-border)] md:w-[236px] md:border-b-0 md:border-r md:border-[var(--flow-border)] md:overflow-y-auto">
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
