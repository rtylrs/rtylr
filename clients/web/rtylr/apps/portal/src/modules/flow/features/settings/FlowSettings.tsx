import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useState, useEffect, useRef } from 'react';
import { motion } from 'motion/react';
import { Tag, Trash2 } from 'lucide-react';
import { getFlowSettingsSections } from '../../settingsNavigation';
import { HelmetContainer } from '@voxire/components';

import { useQuery, useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useFeatureFlag, FLAG_KEYS } from '@voxire/config';
import ModeToggle from '../../shared/components/ModeToggle';
import { cn } from '../../shared/lib/utils';
import { useLifePreferences } from '../../shared/lib/lifePreferences';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiUserSettings } from '../../shared/services/flowApi';
import { useUserStore as useAuthUserStore } from '../../../../shared/stores/userStore';
import { getRouteApi } from '@tanstack/react-router';
import { useLabels, useCreateLabel, useDeleteLabel } from '../../features/tasks/useLabels';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import GateNotice from '../../shared/components/GateNotice';
import TimeSection from './sections/TimeSection';
import BinSection from './sections/BinSection';
import SidebarSettingsSection from './sections/SidebarSettingsSection';
import { canFlowPermission } from '../../shared/lib/roleUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

export const NotificationsSection: React.FC = () => {
  const { t } = useTranslation();
  const { data: remote } = useQuery({
    queryKey: ['flow-settings'],
    queryFn: () => unwrap(flowApi.getUserSettings()),
  });

  const save = useMutation({
    mutationFn: (data: ApiUserSettings) => unwrap(flowApi.updateUserSettings(data)),
    onError: () => toast.error(t('flowToasts.notification.savePreferencesFailed')),
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
    // The save endpoint returns an empty body; only adopt complete settings.
    if (remote && typeof remote.taskAssigned === 'boolean') setSettings(remote);
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
      label: t('flow.settings.notifications.items.taskAssigned.label'),
      desc: t('flow.settings.notifications.items.taskAssigned.desc'),
    },
    {
      key: 'taskComments' as const,
      label: t('flow.settings.notifications.items.taskComments.label'),
      desc: t('flow.settings.notifications.items.taskComments.desc'),
    },
    {
      key: 'mentions' as const,
      label: t('flow.settings.notifications.items.mentions.label'),
      desc: t('flow.settings.notifications.items.mentions.desc'),
    },
    {
      key: 'statusChanges' as const,
      label: t('flow.settings.notifications.items.statusChanges.label'),
      desc: t('flow.settings.notifications.items.statusChanges.desc'),
    },
    {
      key: 'weeklyDigest' as const,
      label: t('flow.settings.notifications.items.weeklyDigest.label'),
      desc: t('flow.settings.notifications.items.weeklyDigest.desc'),
    },
    {
      key: 'emailNotifications' as const,
      label: t('flow.settings.notifications.items.emailNotifications.label'),
      desc: t('flow.settings.notifications.items.emailNotifications.desc'),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flowCommon.settings.notifications')}
        </h2>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.notifications.subtitle')}
        </p>
      </div>

      <div className="space-y-1">
        {items.map((item) => (
          <div
            key={item.key}
            className="flex items-center justify-between py-3 px-4 rounded-lg hover:bg-[var(--flow-hover)] transition-colors"
          >
            <div>
              <p className="text-flow-body font-medium" style={{ color: 'var(--flow-text)' }}>
                {item.label}
              </p>
              <p className="text-flow-meta mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
                {item.desc}
              </p>
            </div>
            <button
              type="button"
              role="switch"
              // The switch is the only control in the row, and the label beside
              // it is not associated with it, so without this a screen reader
              // announces six switches with no names at all. item.label is
              // already translated copy, so this needs no new key.
              aria-label={item.label}
              aria-checked={settings[item.key]}
              onClick={() => toggle(item.key)}
              className={cn(
                'flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full p-0.5',
                'transition-colors focus:focus-visible:ring-2 focus-visible:ring-[var(--accent-focus)]',
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
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flowCommon.settings.life')}
        </h2>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flowCommon.settings.quietHoursNote')}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label
            className="block text-flow-meta font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.settings.life.quietStart')}
          </label>
          <input
            type="time"
            value={prefs.quietStart}
            onChange={(e) => setPrefs({ quietStart: e.target.value })}
            className="w-full rounded-lg px-3 py-2 text-flow-body"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          />
        </div>
        <div>
          <label
            className="block text-flow-meta font-medium mb-1.5"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.settings.life.quietEnd')}
          </label>
          <input
            type="time"
            value={prefs.quietEnd}
            onChange={(e) => setPrefs({ quietEnd: e.target.value })}
            className="w-full rounded-lg px-3 py-2 text-flow-body"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          />
        </div>
      </div>

      <div>
        <label
          className="block text-flow-meta font-medium mb-1.5"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.life.digestHour')}
        </label>
        <select
          value={prefs.digestHour}
          onChange={(e) => setPrefs({ digestHour: Number(e.target.value) })}
          className="w-full max-w-[200px] rounded-lg px-3 py-2 text-flow-body"
          style={{
            background: 'var(--flow-input-bg)',
            border: '1px solid var(--flow-border-control)',
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

export const AppearanceSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  return (
    <section className="space-y-3">
      <h2 className="text-flow-title font-semibold">{t('flowCommon.settings.appearance')}</h2>
      <ModeToggle />
    </section>
  );
};

const ShortcutsSection: React.FC = () => {
  const { t } = useTranslation();
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const shortcuts = [
    { keys: ['C'], desc: t('flow.settings.shortcuts.items.createTask') },
    { keys: ['/'], desc: t('flow.settings.shortcuts.items.openSearch') },
    { keys: ['⌘', 'K'], desc: t('flow.settings.shortcuts.items.openCommandBar') },
    { keys: ['G', 'I'], desc: t('flow.settings.shortcuts.items.goToInbox') },
    { keys: ['G', 'M'], desc: t('flow.settings.shortcuts.items.goToMyIssues') },
    { keys: ['G', 'F'], desc: t('flow.settings.shortcuts.items.goToToday') },
    { keys: ['G', 'C'], desc: t('flow.settings.shortcuts.items.goToCalendar') },
    // Timeline/Roadmap/Insights are gated behind the same preview-routes flag
    // as their sidebar links and routes (see FlowLayout.tsx, router.tsx) - the
    // key presses still fire, but land on PreviewRouteUnavailable when the
    // flag is off, so the reference list should only claim them when the
    // sidebar itself does.
    ...(showPreviewRoutes
      ? [
          { keys: ['G', 'T'], desc: t('flow.settings.shortcuts.items.goToTimeline') },
          { keys: ['G', 'R'], desc: t('flow.settings.shortcuts.items.goToRoadmap') },
          { keys: ['G', 'S'], desc: t('flow.settings.shortcuts.items.goToInsights') },
        ]
      : []),
    { keys: ['G', 'Y'], desc: t('flow.settings.shortcuts.items.goToCycles') },
    { keys: ['G', 'X'], desc: t('flow.settings.shortcuts.items.goToTriage') },
    { keys: ['Esc'], desc: t('flow.settings.shortcuts.items.closeModal') },
    { keys: ['⌘', '↵'], desc: t('flow.settings.shortcuts.items.submitComment') },
  ];
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flowCommon.settings.shortcuts')}
        </h2>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.shortcuts.subtitle')}
        </p>
      </div>
      <div className="space-y-1">
        {shortcuts.map((s, i) => (
          <div
            key={i}
            className="flex items-center justify-between py-2 px-3 rounded-sm hover:bg-[var(--flow-hover)]"
          >
            <span className="text-flow-meta" style={{ color: 'var(--flow-text)' }}>
              {s.desc}
            </span>
            <div className="flex items-center gap-1">
              {s.keys.map((k, j) => (
                <kbd
                  key={j}
                  className="px-1.5 py-0.5 rounded-sm text-flow-meta font-mono font-semibold"
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

export const LabelsSection: React.FC = () => {
  const { t } = useTranslation();
  const labelsQuery = useLabels();
  const { data: labels = [], isLoading } = labelsQuery;
  const createLabel = useCreateLabel();
  const deleteLabel = useDeleteLabel();

  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState(LABEL_COLORS[0]);
  const customColorInputRef = useRef<HTMLInputElement | null>(null);
  const nameInputRef = useRef<HTMLInputElement | null>(null);
  const [pendingDeleteLabel, setPendingDeleteLabel] = useState<{
    id: number;
    name: string;
  } | null>(null);

  const handleCreate = () => {
    const name = newName.trim();
    if (!name) {
      toast.error(t('flowErrors.generic'));
      return;
    }
    createLabel.mutate(
      { name, color: newColor },
      {
        onSuccess: () => {
          setNewName('');
          setNewColor(LABEL_COLORS[0]);
          toast.success(t('flowToasts.label.created'));
        },
      }
    );
  };

  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border-control)',
    color: 'var(--flow-text)',
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flowCommon.settings.labels')}
        </h2>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.labels.subtitle')}
        </p>
      </div>

      {/* Create form */}
      <div
        className="rounded-xl p-4 space-y-3"
        style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
      >
        <h3 className="text-flow-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.settings.labels.newLabel')}
        </h3>
        <div className="flex gap-2">
          <input
            ref={nameInputRef}
            type="text"
            placeholder={t('flow.settings.labels.namePlaceholder')}
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            className="flex-1 rounded-lg px-3 py-2 text-flow-body"
            style={inputStyle}
          />
        </div>
        <div>
          <p className="text-flow-meta mb-2" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.settings.labels.color')}
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
                border: '1px solid var(--flow-border-control)',
              }}
              title={t('flow.settings.labels.chooseCustomColor')}
              aria-label={t('flow.settings.labels.chooseCustomColor')}
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
                aria-label={t('flow.settings.chooseColorLabel', { color: c })}
                aria-pressed={newColor === c}
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
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-flow-meta font-semibold disabled:opacity-40"
          style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
        >
          {createLabel.isPending
            ? t('flow.settings.labels.creating')
            : t('flow.settings.labels.createLabel')}
        </button>
      </div>

      {/* Existing labels */}
      <div>
        <h3
          className="text-flow-meta font-semibold mb-3"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.settings.labels.workspaceLabels')}
        </h3>
        {isLoading ? (
          <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flowCommon.loading')}
          </p>
        ) : hasQueryFailed(labelsQuery) ? (
          // A failed request emptied the list exactly as a first run does, so
          // a 500 invited the user to recreate labels they already had.
          <ListLoadFailed onRetry={() => void labelsQuery.refetch()} />
        ) : labels.length === 0 ? (
          <EmptyState
            icon={Tag}
            title={t('flow.settings.labels.emptyList')}
            action={{
              label: t('flow.settings.labels.newLabel'),
              onClick: () => nameInputRef.current?.focus(),
            }}
            size="compact"
          />
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
                  <span
                    className="text-flow-body font-medium"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {label.name}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => setPendingDeleteLabel({ id: label.id, name: label.name })}
                  className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-dim)' }}
                  title={t('flow.settings.labels.deleteLabel')}
                  aria-label={t('flow.settings.deleteLabelLabel')}
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
        title={t('flow.settings.labels.deleteLabel')}
        description={
          <>
            {t('flow.settings.labels.deleteConfirmPrefix')}{' '}
            <strong>{pendingDeleteLabel?.name}</strong>
            {t('flow.settings.labels.deleteConfirmSuffix')}
          </>
        }
        confirmLabel={t('flowCommon.delete')}
        pending={deleteLabel.isPending}
        onConfirm={() => {
          if (!pendingDeleteLabel) return;
          deleteLabel.mutate(pendingDeleteLabel.id, {
            onSuccess: () => {
              toast.success(t('flowToasts.label.deleted'));
              setPendingDeleteLabel(null);
            },
            onError: () => toast.error(t('flowToasts.label.deleteFailed')),
          });
        }}
      />
    </div>
  );
};

const SECTION_CONTENT: Record<string, React.FC> = {
  bin: BinSection,
  notifications: NotificationsSection,
  appearance: AppearanceSection,
  labels: LabelsSection,
  time: TimeSection,
  life: LifeRhythmSection,
  shortcuts: ShortcutsSection,
  sidebar: SidebarSettingsSection,
};

const settingsRouteApi = getRouteApi('/business/settings/flow');

const FlowSettings: React.FC = () => {
  const { t } = useTranslation();
  const effectiveAccess = useAuthUserStore((s) => s.user?.effectiveAccess);
  const sections = getFlowSettingsSections(effectiveAccess);
  const { section } = settingsRouteApi.useSearch();
  const activeSection = sections.some((item) => item.id === section) ? section! : 'notifications';
  const ActiveComponent = SECTION_CONTENT[activeSection];

  return (
    <div className="portal-flow-settings portal-app-settings">
      <HelmetContainer title={`${t('flowNav.settings')} · Flow`} />
      <div className="max-w-2xl">
        <ActiveComponent />
      </div>
    </div>
  );
};

export default FlowSettings;
