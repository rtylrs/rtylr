import React, { useState, useEffect, useMemo, useRef } from 'react';
import { motion } from 'motion/react';
import {
  Bell,
  Palette,
  Keyboard,
  Sparkles,
  Tag,
  Trash2,
  Users,
  UserMinus,
  UserPlus,
  LayoutDashboard,
  Plus,
  Clock,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useQuery, useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  useFlowTeam,
  useFlowGrantAccess,
  useFlowDeactivateMember,
} from '../../shared/hooks/queries/useFlowTeam';
import {
  useFlowSubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useFlowSubscription';
import { useThemeStore } from '../../shared/stores/themeStore';
import { cn } from '../../shared/lib/utils';
import { useLifePreferences } from '../../shared/lib/lifePreferences';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiUserSettings } from '../../shared/services/flowApi';
import { getCookie } from '@voxire/auth';
import { getRouteApi } from '@tanstack/react-router';
import type { SettingsSectionId } from '../../shared/schemas/settings';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useLabels, useCreateLabel, useDeleteLabel } from '../../features/tasks/useLabels';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import TimeSection from './sections/TimeSection';
import {
  useWorkspaces,
  useCreateWorkspace,
} from '../../features/workspace-dashboard/useWorkspaces';

interface SettingsSection {
  id: string;
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

function buildSections(
  t: (key: string) => string,
  canManageWorkspaces: boolean
): SettingsSection[] {
  const sections: SettingsSection[] = [
    { id: 'team', label: t('common.settings.team'), icon: Users },
    { id: 'notifications', label: t('common.settings.notifications'), icon: Bell },
    // Admin/owner only: rates and the approval lock change what the whole
    // workspace can edit. Gated on the same role check as Workspaces.
    ...(canManageWorkspaces
      ? [{ id: 'time', label: t('flow.settings.time.section'), icon: Clock }]
      : []),
    { id: 'appearance', label: t('common.settings.appearance'), icon: Palette },
    { id: 'labels', label: t('common.settings.labels'), icon: Tag },
    { id: 'life', label: t('common.settings.life'), icon: Sparkles },
    { id: 'shortcuts', label: t('common.settings.shortcuts'), icon: Keyboard },
  ];
  if (canManageWorkspaces) {
    sections.unshift({ id: 'workspaces', label: 'Workspaces', icon: LayoutDashboard });
  }
  return sections;
}

const WorkspacesSection: React.FC = () => {
  const { t } = useTranslation();
  const { data: workspaces = [], isLoading } = useWorkspaces();
  const createWorkspace = useCreateWorkspace();
  const [showCreate, setShowCreate] = useState(false);
  const [wsName, setWsName] = useState('');
  const [wsColor, setWsColor] = useState('#64C5B9');

  const WORKSPACE_COLORS = [
    '#64C5B9',
    '#3b82f6',
    '#8b5cf6',
    '#f59e0b',
    '#ef4444',
    '#22c55e',
    '#ec4899',
    '#f97316',
  ];

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    color: 'var(--flow-text)',
  };

  const handleCreate = (): void => {
    const name = wsName.trim();
    if (!name) {
      toast.error(t('toasts.workspace.nameRequired'));
      return;
    }
    createWorkspace.mutate(
      { name, color: wsColor },
      {
        onSuccess: () => {
          toast.success(t('toasts.workspace.created'));
          setWsName('');
          setWsColor('#64C5B9');
          setShowCreate(false);
        },
        onError: () => toast.error(t('toasts.workspace.createFailed')),
      }
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          Workspaces
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          Manage your Flow workspaces.
        </p>
      </div>

      {showCreate ? (
        <div
          className="rounded-xl p-4 space-y-3"
          style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
        >
          <h3 className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            New workspace
          </h3>
          <input
            type="text"
            placeholder={t('flow.settings.workspaceNamePlaceholder')}
            value={wsName}
            onChange={(e) => setWsName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
            style={inputStyle}
          />
          <div>
            <p className="text-[11px] mb-2" style={{ color: 'var(--flow-text-muted)' }}>
              Color
            </p>
            <div className="flex gap-2">
              {WORKSPACE_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setWsColor(c)}
                  className="h-6 w-6 rounded-full transition-all"
                  style={{
                    background: c,
                    outline: wsColor === c ? `2px solid ${c}` : 'none',
                    outlineOffset: 2,
                    boxShadow: wsColor === c ? '0 0 0 2px var(--flow-bg)' : 'none',
                  }}
                />
              ))}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!wsName.trim() || createWorkspace.isPending}
              onClick={handleCreate}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              <Plus size={12} />
              {createWorkspace.isPending ? 'Creating…' : 'Create workspace'}
            </button>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              className="px-3 py-1.5 rounded-lg text-[12px]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold"
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          <Plus size={12} />
          New workspace
        </button>
      )}

      {isLoading ? (
        <div className="space-y-2">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="h-12 rounded-xl animate-pulse"
              style={{ background: 'var(--flow-hover)' }}
            />
          ))}
        </div>
      ) : (
        <div
          className="rounded-xl divide-y divide-[var(--flow-border)]"
          style={{ border: '1px solid var(--flow-border)' }}
        >
          {workspaces.map((ws) => (
            <div key={ws.id} className="flex items-center gap-3 px-4 py-3">
              <span
                className="h-3 w-3 rounded-full shrink-0"
                style={{ background: ws.color || 'var(--accent)' }}
              />
              <p className="text-[13px] font-medium flex-1" style={{ color: 'var(--flow-text)' }}>
                {ws.name}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

const NotificationsSection: React.FC = () => {
  const { t } = useTranslation();
  const { data: remote } = useQuery({
    queryKey: ['flow-settings'],
    queryFn: () => unwrap(flowApi.getUserSettings()),
  });

  const save = useMutation({
    mutationFn: (data: ApiUserSettings) => unwrap(flowApi.updateUserSettings(data)),
    onError: () => toast.error(t('toasts.notification.savePreferencesFailed')),
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

  const toggle = (key: keyof ApiUserSettings): void => {
    const prev = { ...settings };
    const next = { ...settings, [key]: !settings[key] };
    setSettings(next);
    save.mutate(next, { onError: () => setSettings(prev) });
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
                background: settings[item.key] ? 'var(--accent)' : 'var(--flow-hover)',
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
  const { t } = useTranslation();
  const { prefs, setPrefs } = useLifePreferences();

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('common.settings.life')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('common.settings.quietHoursNote')}
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
                borderColor: theme === value ? 'var(--accent)' : 'var(--flow-border)',
                background: 'var(--flow-card)',
                color: theme === value ? 'var(--accent)' : 'var(--flow-text-muted)',
              }}
            >
              <div className="w-full h-12 rounded-lg mb-2" style={{ background: preview }} />
              {label}
            </button>
          ))}
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
  const { t } = useTranslation();
  const { activeWorkspaceId } = useFlowStore();
  const workspaceId = activeWorkspaceId ?? '';
  const { data: labels = [], isLoading } = useLabels(workspaceId || undefined);
  const createLabel = useCreateLabel();
  const deleteLabel = useDeleteLabel();

  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState(LABEL_COLORS[0]);
  const customColorInputRef = useRef<HTMLInputElement | null>(null);
  const [pendingDeleteLabel, setPendingDeleteLabel] = useState<{
    id: number;
    name: string;
  } | null>(null);

  const handleCreate = () => {
    const name = newName.trim();
    if (!name) {
      toast.error(t('errors.generic'));
      return;
    }
    if (!workspaceId) {
      toast.error(t('toasts.workspace.noActive'));
      return;
    }
    createLabel.mutate(
      { workspace_id: workspaceId, name, color: newColor },
      {
        onSuccess: () => {
          setNewName('');
          setNewColor(LABEL_COLORS[0]);
          toast.success(t('toasts.label.created'));
        },
        onError: () => toast.error(t('toasts.label.createFailed')),
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
            placeholder={t('flow.settings.labelNamePlaceholder')}
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
            <button
              type="button"
              onClick={() => customColorInputRef.current?.click()}
              className="relative h-6 w-6 rounded-full transition-all overflow-hidden"
              style={{
                background: `linear-gradient(135deg, ${newColor} 0%, ${newColor} 45%, #ffffff 45%, #ffffff 55%, #111111 55%, #111111 100%)`,
                outline: !LABEL_COLORS.includes(newColor) ? `2px solid ${newColor}` : 'none',
                outlineOffset: 2,
                boxShadow: !LABEL_COLORS.includes(newColor) ? '0 0 0 2px var(--flow-bg)' : 'none',
                border: '1px solid var(--flow-border)',
              }}
              title="Choose custom color"
              aria-label="Choose custom color"
            >
              <input
                ref={customColorInputRef}
                type="color"
                value={newColor}
                onChange={(e) => setNewColor(e.target.value)}
                className="sr-only"
                tabIndex={-1}
                aria-hidden="true"
              />
            </button>
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
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
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
                  onClick={() => setPendingDeleteLabel({ id: label.id, name: label.name })}
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
      <ConfirmDialog
        open={pendingDeleteLabel !== null}
        onOpenChange={(o) => {
          if (!o) setPendingDeleteLabel(null);
        }}
        title="Delete label"
        description={
          <>
            Delete <strong>{pendingDeleteLabel?.name}</strong>? It is removed from every task it is
            attached to.
          </>
        }
        confirmLabel="Delete"
        pending={deleteLabel.isPending}
        onConfirm={() => {
          if (!pendingDeleteLabel) return;
          deleteLabel.mutate(pendingDeleteLabel.id, {
            onSuccess: () => {
              toast.success(t('toasts.label.deleted'));
              setPendingDeleteLabel(null);
            },
            onError: () => toast.error(t('toasts.label.deleteFailed')),
          });
        }}
      />
    </div>
  );
};

// ── Team Section ─────────────────────────────────────────────────────────────

const TeamSection: React.FC = () => {
  const { data: allMembers = [], isLoading } = useFlowTeam();
  const grant = useFlowGrantAccess();
  const deactivate = useFlowDeactivateMember();
  const { data: sub } = useFlowSubscription();

  // The /team/members endpoint returns the entire business team. Flow
  // is one of several apps a member can have access to; we split here
  // so the visible list and the grant picker stay in sync.
  const flowMembers = allMembers.filter((m) => m.apps?.includes('flow'));
  const candidates = allMembers.filter((m) => !m.apps?.includes('flow'));

  const [selectedUserId, setSelectedUserId] = useState('');
  const [showInvite, setShowInvite] = useState(false);

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    color: 'var(--flow-text)',
  };

  const handleGrant = (): void => {
    const member = candidates.find((m) => String(m.id) === selectedUserId);
    if (!member) return;
    grant.mutate(
      { userId: member.id, apps: member.apps ?? [], role: member.role ?? 'member' },
      {
        onSuccess: () => {
          setSelectedUserId('');
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

      {/* Grant Flow access to an existing business member */}
      {showInvite ? (
        <div
          className="rounded-xl p-4 space-y-3"
          style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
        >
          <div>
            <h3 className="text-[13px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
              Grant Flow access
            </h3>
            <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
              Only people who are already part of your business can be added here. Add new people
              from the Team page first.
            </p>
          </div>
          <div>
            <label
              className="block text-[11px] font-medium mb-1"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Team member
            </label>
            {candidates.length === 0 ? (
              <p
                className="rounded-lg px-3 py-2 text-[12px]"
                style={{ ...inputStyle, color: 'var(--flow-text-muted)' }}
              >
                Everyone in your business already has Flow access.
              </p>
            ) : (
              <select
                value={selectedUserId}
                onChange={(e) => setSelectedUserId(e.target.value)}
                className="w-full rounded-lg px-3 py-2 text-[13px] outline-none"
                style={inputStyle}
              >
                <option value="">Select a team member…</option>
                {candidates.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name === m.email ? m.email : `${m.name} - ${m.email}`}
                  </option>
                ))}
              </select>
            )}
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
              disabled={!selectedUserId || grant.isPending}
              onClick={handleGrant}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              <UserPlus size={12} />
              {grant.isPending ? 'Granting…' : 'Grant access'}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowInvite(false);
                setSelectedUserId('');
              }}
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
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          <UserPlus size={12} />
          Add member
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
      ) : flowMembers.length === 0 ? (
        <div className="py-8 text-center" style={{ color: 'var(--flow-text-dim)' }}>
          <Users size={24} className="mx-auto mb-2 opacity-40" />
          <p className="text-[13px]">No team members yet.</p>
        </div>
      ) : (
        <div
          className="rounded-xl divide-y divide-[var(--flow-border)]"
          style={{ border: '1px solid var(--flow-border)' }}
        >
          {flowMembers.map((m) => (
            <div key={m.id} className="flex items-center gap-3 px-4 py-3">
              <div
                className="h-8 w-8 rounded-full flex items-center justify-center text-[12px] font-semibold shrink-0"
                style={{ background: 'var(--accent-border)', color: 'var(--accent)' }}
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

// ── Section content map ───────────────────────────────────────────────────────

const SECTION_CONTENT: Record<string, React.FC> = {
  workspaces: WorkspacesSection,
  team: TeamSection,
  notifications: NotificationsSection,
  appearance: AppearanceSection,
  labels: LabelsSection,
  time: TimeSection,
  life: LifeRhythmSection,
  shortcuts: ShortcutsSection,
};

const PlaceholderSection: React.FC<{ id: string }> = ({ id }) => {
  const { t } = useTranslation();
  const sections = buildSections(t, false);
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <p className="text-[14px] font-medium mb-1" style={{ color: 'var(--flow-text-muted)' }}>
        {sections.find((s) => s.id === id)?.label}
      </p>
      <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
        This section is not yet available.
      </p>
    </div>
  );
};

function getRoleFromToken(token: string | null): string {
  if (!token) return '';
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return (payload.role as string) || '';
  } catch {
    return '';
  }
}

const MANAGE_WORKSPACE_ROLES = new Set(['owner', 'admin', 'manager']);

// getRouteApi takes a route id, which for a top-level route is its path.
const settingsRouteApi = getRouteApi('/settings');

const FlowSettings: React.FC = () => {
  const { t } = useTranslation();
  const token = getCookie('token');
  const role = useMemo(() => getRoleFromToken(token), [token]);
  const canManageWorkspaces = MANAGE_WORKSPACE_ROLES.has(role);
  // ?section=time makes a section linkable, so the sidebar can point at one.
  const { section } = settingsRouteApi.useSearch();
  const [activeSection, setActiveSection] = useState<SettingsSectionId | string>(
    section ?? (canManageWorkspaces ? 'workspaces' : 'team')
  );
  useEffect(() => {
    if (section) setActiveSection(section);
  }, [section]);
  const sections = buildSections(t, canManageWorkspaces);

  const ActiveComponent =
    SECTION_CONTENT[activeSection] || (() => <PlaceholderSection id={activeSection} />);

  return (
    <div className="flex h-full min-h-0 flex-col md:flex-row">
      <HelmetContainer title="Settings · rtylr" />
      {/* Settings sidebar */}
      <div className="w-full shrink-0 overflow-x-auto overflow-y-hidden border-b border-[var(--flow-border)] md:w-[236px] md:border-b-0 md:border-r md:border-[var(--flow-border)] md:overflow-y-auto">
        <div className="flex flex-row gap-1 p-3 md:flex-col md:gap-0 md:p-4">
          <h1
            className="mb-3 hidden text-[13px] font-semibold md:mb-4 md:block"
            style={{ color: 'var(--flow-text)' }}
          >
            {t('nav.settings')}
          </h1>
          <nav className="flex min-w-max flex-row gap-1 md:min-w-0 md:flex-col md:space-y-0.5">
            {sections.map((section) => {
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
                      style={{ background: 'var(--accent)' }}
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
