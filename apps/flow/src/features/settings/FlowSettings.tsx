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
  LayoutDashboard,
  Clock,
  PanelLeft,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useQuery, useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useFeatureFlag, FLAG_KEYS } from '@voxire/config';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import {
  ACCENT_DEFAULT,
  ACCENT_SWATCHES,
  appearanceTokens,
  isValidHex,
  normalizeHex,
  palettesForMode,
  resolveMode,
  resolvePalette,
} from '@voxire/utils';
import type {
  AppearancePrefs,
  AppearanceTokens,
  Corners,
  Density,
  TextSize,
  ThemeMode,
} from '@voxire/utils';
import { useAppearanceStore } from '../../shared/stores/appearanceStore';
import { cn } from '../../shared/lib/utils';
import { useLifePreferences } from '../../shared/lib/lifePreferences';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiUserSettings } from '../../shared/services/flowApi';
import { useUserStore as useAuthUserStore } from '@voxire/auth';
import { getRouteApi } from '@tanstack/react-router';
import type { SettingsSectionId } from '../../shared/schemas/settings';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useLabels, useCreateLabel, useDeleteLabel } from '../../features/tasks/useLabels';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import GateNotice from '../../shared/components/GateNotice';
import AppearanceMiniature from './components/AppearanceMiniature';
import AppearanceOptionCard from './components/AppearanceOptionCard';
import AppearancePreview from './components/AppearancePreview';
import AppearanceSwitch from './components/AppearanceSwitch';
import SegmentedControl from './components/SegmentedControl';
import { useSystemDark } from './components/useSystemDark';
import TimeSection from './sections/TimeSection';
import BinSection from './sections/BinSection';
import SidebarSettingsSection from './sections/SidebarSettingsSection';
import { canFlowPermission } from '../../shared/lib/roleUtils';

interface SettingsSection {
  id: string;
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

function buildSections(
  t: (key: string) => string,
  canManageWorkspaces: boolean,
  canManageTeam: boolean,
  canManageBin: boolean
): SettingsSection[] {
  const sections: SettingsSection[] = [
    // Owner/admin only: grants and revokes Flow access for other business
    // members, gated the same way canInviteMembers gates it everywhere else.
    ...(canManageTeam ? [{ id: 'team', label: t('common.settings.team'), icon: Users }] : []),
    { id: 'notifications', label: t('common.settings.notifications'), icon: Bell },
    // Admin/owner only: rates and the approval lock change what the whole
    // workspace can edit. Gated on the same role check as Workspaces.
    ...(canManageWorkspaces
      ? [{ id: 'time', label: t('flow.settings.time.section'), icon: Clock }]
      : []),
    { id: 'appearance', label: t('common.settings.appearance'), icon: Palette },
    { id: 'sidebar', label: t('flow.sidebar.settingsSection'), icon: PanelLeft },
    { id: 'labels', label: t('common.settings.labels'), icon: Tag },
    { id: 'life', label: t('common.settings.life'), icon: Sparkles },
    { id: 'shortcuts', label: t('common.settings.shortcuts'), icon: Keyboard },
    // Business-wide recycle bin. Gated on the central Flow settings
    // permission, which is also enforced by the server.
    ...(canManageBin ? [{ id: 'bin', label: t('flow.settings.bin.section'), icon: Trash2 }] : []),
  ];
  if (canManageWorkspaces) {
    sections.unshift({
      id: 'workspaces',
      label: t('flow.settings.workspaces.section'),
      icon: LayoutDashboard,
    });
  }
  return sections;
}

export const NotificationsSection: React.FC = () => {
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
    // Adopt the server's settings only when they actually carry the booleans.
    // The cache entry for ['flow-settings'] used to be overwritten with the
    // save endpoint's empty `200 {}` response, and `{}` is truthy, so this
    // wiped all six switches. The next toggle then spread that empty object
    // and sent a one-field body, which the server reads as "the other five
    // are false" - one click silently cleared five preferences (#1063).
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
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('common.settings.notifications')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
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
              // The switch is the only control in the row, and the label beside
              // it is not associated with it, so without this a screen reader
              // announces six switches with no names at all. item.label is
              // already translated copy, so this needs no new key.
              aria-label={item.label}
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
            {t('flow.settings.life.quietStart')}
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
            {t('flow.settings.life.quietEnd')}
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
          {t('flow.settings.life.digestHour')}
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

// ── Appearance ────────────────────────────────────────────────────────────────

const MODE_OPTIONS: readonly ThemeMode[] = ['light', 'dark', 'system'];
const DENSITY_OPTIONS: readonly Density[] = ['compact', 'cozy', 'roomy'];
const TEXT_SIZE_OPTIONS: readonly TextSize[] = ['s', 'm', 'l'];
const CORNER_OPTIONS: readonly Corners[] = ['sharp', 'soft', 'round'];

/** id the disabled text-size control points its aria-describedby at. */
const TEXT_SIZE_GATE_ID = 'appearance-text-size-gate';

/** A colour input rejects a value it cannot parse, so it never receives one. */
function safeSwatch(value: string): string {
  return isValidHex(value) ? normalizeHex(value) : ACCENT_DEFAULT;
}

/**
 * The Appearance panel: seven controls, one of them deliberately disabled.
 *
 * Every control commits on click, straight to the store, and the provider in
 * `@voxire/ui` applies the result. There is no Apply button and no draft copy of
 * the preferences: `audition` holds only what is being hovered or focused, it
 * feeds the preview pane and nothing else, and it is cleared on leave and blur.
 *
 * Text size ships disabled with its reason stated. 618 of Flow's 637 font-size
 * declarations are px literals, so no custom property reaches them and a control
 * that moved the other 3% would be a lie. The concept is real, which is why it
 * is disabled rather than absent.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * sections 5 to 7.
 */
export const AppearanceSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const {
    mode,
    palette,
    accent,
    density,
    textSize,
    corners,
    reduceMotion,
    setMode,
    setPalette,
    setAccent,
    setDensity,
    setTextSize,
    setCorners,
    setReduceMotion,
  } = useAppearanceStore();

  const prefersDark = useSystemDark();
  // Read from the store rather than a query hook: the workspaces API was stubbed
  // out when Flow moved to the v2 contracts, so the store is the only remaining
  // source and it is what FlowLayout resolves the same colour from.
  const { activeWorkspaceId, workspaces } = useFlowStore();

  // The accent the derivation falls back to while `accent` is null, which is
  // what "a personal accent overrides the workspace one" means in practice.
  const workspaceAccent = safeSwatch(
    workspaces.find((w) => w.id === activeWorkspaceId)?.color || ACCENT_DEFAULT
  );

  const prefs = useMemo<AppearancePrefs>(
    () => ({ mode, palette, accent, density, textSize, corners, reduceMotion }),
    [mode, palette, accent, density, textSize, corners, reduceMotion]
  );

  // The audition. Never written to the store, never persisted, cleared the
  // moment the pointer or the focus leaves the option that set it.
  const [audition, setAudition] = useState<Partial<AppearancePrefs> | undefined>(undefined);
  const clearAudition = (): void => setAudition(undefined);

  const resolvedMode = resolveMode(mode, prefersDark);
  const presets = palettesForMode(resolvedMode);
  const activePalette = resolvePalette(palette, resolvedMode);

  const tokensFor = (override: Partial<AppearancePrefs>): AppearanceTokens => {
    const next = { ...prefs, ...override };
    return appearanceTokens(next, {
      resolvedMode: resolveMode(next.mode, prefersDark),
      workspaceAccent,
    });
  };

  const appliedAccent = safeSwatch(accent ?? workspaceAccent);
  const [hexDraft, setHexDraft] = useState<string | null>(null);
  const hexValue = hexDraft ?? appliedAccent;
  const hexInvalid = hexDraft !== null && !isValidHex(hexDraft);

  /** Writes only a parseable value, so a half-typed hex never reaches the store. */
  const onHexChange = (raw: string): void => {
    setHexDraft(raw);
    if (isValidHex(raw)) setAccent(normalizeHex(raw));
  };

  const commitAccent = (hex: string | null): void => {
    setHexDraft(null);
    setAccent(hex);
  };

  const labelClass = 'mb-3 block text-[12px] font-medium';
  const labelStyle = { color: 'var(--flow-text-muted)' };
  const inputStyle = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border)',
    borderRadius: 'var(--flow-r)',
    color: 'var(--flow-text)',
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="mb-1 text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('common.settings.appearance')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.appearance.subtitle')}
        </p>
        {/* --flow-text-muted, not --flow-text-dim. axe measured dim at 2.23:1
            here, against the 4.5 AA floor for body text, and it fails in both
            modes (2.25 dark, 2.35 light) because it is a 0.28 alpha decorative
            token. This line is the disclosure that makes the settings-seam rule
            satisfiable, so it is the one string on the screen that must not be
            hard to read. Muted measures 6.15 dark and 5.69 light. */}
        <p className="mt-1 text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.appearance.perDevice')}
        </p>
      </div>

      {/* Sticky so an option lower down can still be auditioned against it. */}
      <div
        className="sticky top-0 z-10 space-y-2 pb-3"
        style={{ background: 'var(--flow-bg)', borderBlockEnd: '1px solid var(--flow-border)' }}
      >
        <p className="text-[11px] font-medium" style={labelStyle}>
          {t('flow.settings.appearance.preview.label')}
        </p>
        <AppearancePreview
          prefs={prefs}
          previewOverride={audition}
          workspaceAccent={workspaceAccent}
          prefersDark={prefersDark}
        />
      </div>

      <div>
        <p id="appearance-mode-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.mode')}
        </p>
        <div
          className="grid grid-cols-3 gap-2"
          role="group"
          aria-labelledby="appearance-mode-label"
          onPointerLeave={clearAudition}
        >
          {MODE_OPTIONS.map((value) => (
            <AppearanceOptionCard
              key={value}
              label={t(`flow.shell.${value}`)}
              selected={mode === value}
              onSelect={() => setMode(value)}
              onAuditionStart={() => setAudition({ mode: value })}
              onAuditionEnd={clearAudition}
            >
              {value === 'system' ? (
                // Two miniatures rather than one, because what `system` shows is
                // exactly "whichever of these the OS is asking for".
                <span className="flex w-full gap-1">
                  <AppearanceMiniature
                    tokens={tokensFor({ mode: 'light' })}
                    className="min-w-0 flex-1"
                  />
                  <AppearanceMiniature
                    tokens={tokensFor({ mode: 'dark' })}
                    className="min-w-0 flex-1"
                  />
                </span>
              ) : (
                <AppearanceMiniature tokens={tokensFor({ mode: value })} />
              )}
            </AppearanceOptionCard>
          ))}
        </div>
      </div>

      <div>
        <p id="appearance-palette-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.palette')}
        </p>
        <div
          className="grid grid-cols-1 gap-2 sm:grid-cols-2"
          role="group"
          aria-labelledby="appearance-palette-label"
          // Clearing here rather than per card: the gap between two cards is
          // unhovered space, and clearing on each card's leave made the preview
          // flash back to the applied palette while crossing it.
          onPointerLeave={clearAudition}
        >
          {presets.map((preset) => (
            <AppearanceOptionCard
              key={preset.id}
              label={t(preset.labelKey)}
              description={t(`flow.settings.appearance.paletteDescriptions.${preset.id}`)}
              selected={activePalette.id === preset.id}
              onSelect={() => setPalette(preset.id)}
              onAuditionStart={() => setAudition({ palette: preset.id })}
              onAuditionEnd={clearAudition}
            >
              <AppearanceMiniature tokens={tokensFor({ palette: preset.id })} />
            </AppearanceOptionCard>
          ))}
        </div>
      </div>

      <div>
        <p id="appearance-accent-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.accent')}
        </p>

        {/* Capped columns, not 1fr: at full panel width 1fr stretches a swatch to
            178px, which reads as a colour bar rather than a swatch. */}
        <div
          className="grid gap-2"
          style={{ gridTemplateColumns: 'repeat(5, minmax(0, 56px))' }}
          role="group"
          aria-labelledby="appearance-accent-label"
          onPointerLeave={clearAudition}
        >
          {ACCENT_SWATCHES.map((swatch) => {
            const selected = accent !== null && normalizeHex(accent) === swatch.hex;
            return (
              <button
                key={swatch.id}
                type="button"
                aria-label={t(swatch.labelKey)}
                aria-pressed={selected}
                onClick={() => commitAccent(swatch.hex)}
                onPointerEnter={() => setAudition({ accent: swatch.hex })}
                onFocus={() => setAudition({ accent: swatch.hex })}
                onBlur={clearAudition}
                className="flow-focus h-[44px] w-full"
                style={{
                  // The one place a literal colour is right: a swatch that does
                  // not paint its own colour is not a preview of anything. The
                  // value is data from ACCENT_SWATCHES, never a typed literal.
                  background: swatch.hex,
                  border: `2px solid ${selected ? 'var(--flow-text)' : 'var(--flow-border)'}`,
                  borderRadius: 'var(--flow-r)',
                }}
              />
            );
          })}
        </div>

        <div className="mt-3 flex flex-wrap items-end gap-2">
          <div className="min-w-0 flex-1">
            <label
              htmlFor="appearance-accent-hex"
              className="mb-1.5 block text-[11px]"
              style={labelStyle}
            >
              {t('flow.settings.appearance.accentHex')}
            </label>
            <input
              id="appearance-accent-hex"
              type="text"
              // dir="ltr" because the value is a hex literal, not prose.
              //
              // Without it the field inherits the document direction, and under
              // RTL the leading '#' is a neutral character that resolves to the
              // visual right: "#f56d8f" renders as "f56d8f#". Confirmed in the
              // browser with dir=rtl on the document before this was added.
              //
              // The surrounding labels and descriptions are deliberately left to
              // inherit, since those are prose and should follow the reader's
              // direction. Only the value is pinned.
              dir="ltr"
              spellCheck={false}
              autoComplete="off"
              value={hexValue}
              aria-invalid={hexInvalid}
              aria-describedby={hexInvalid ? 'appearance-accent-hex-error' : undefined}
              onChange={(e) => onHexChange(e.target.value)}
              /* 16px minimum, or iOS auto-zooms the whole panel on focus. */
              className="flow-focus min-h-[44px] w-full px-3 text-[16px] outline-none"
              style={inputStyle}
            />
          </div>
          <label htmlFor="appearance-accent-picker" className="sr-only">
            {t('flow.settings.appearance.accentPicker')}
          </label>
          <input
            id="appearance-accent-picker"
            type="color"
            value={appliedAccent}
            onChange={(e) => commitAccent(normalizeHex(e.target.value))}
            className="flow-focus h-[44px] w-[44px] shrink-0 cursor-pointer bg-transparent p-0.5"
            style={{ border: '1px solid var(--flow-border)', borderRadius: 'var(--flow-r)' }}
          />
        </div>

        {hexInvalid && (
          <p
            id="appearance-accent-hex-error"
            className="mt-1.5 text-[11px]"
            style={{ color: 'var(--flow-danger-fg)' }}
          >
            {t('flow.settings.appearance.accentInvalid')}
          </p>
        )}

        <button
          type="button"
          aria-pressed={accent === null}
          onClick={() => commitAccent(null)}
          onPointerEnter={() => setAudition({ accent: null })}
          onPointerLeave={clearAudition}
          onFocus={() => setAudition({ accent: null })}
          onBlur={clearAudition}
          className="flow-focus mt-3 inline-flex min-h-[44px] items-center gap-2 px-3 text-[13px] font-medium"
          style={{
            background: accent === null ? 'var(--accent-muted)' : 'var(--flow-card)',
            border: `1px solid ${accent === null ? 'var(--accent-border)' : 'var(--flow-border)'}`,
            borderRadius: 'var(--flow-r)',
            color: accent === null ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)',
          }}
        >
          <span
            aria-hidden="true"
            className="block h-4 w-4 shrink-0 rounded-full"
            style={{ background: workspaceAccent }}
          />
          {t('flow.settings.appearance.accentReset')}
        </button>
      </div>

      <div>
        <p id="appearance-density-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.density')}
        </p>
        <SegmentedControl
          labelledBy="appearance-density-label"
          value={density}
          options={DENSITY_OPTIONS.map((value) => ({
            value,
            label: t(`flow.settings.appearance.density.${value}`),
          }))}
          onChange={setDensity}
          onAudition={(value) => setAudition(value === null ? undefined : { density: value })}
        />
      </div>

      <div>
        <p id="appearance-textSize-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.textSize')}
        </p>
        <SegmentedControl
          labelledBy="appearance-textSize-label"
          value={textSize}
          options={TEXT_SIZE_OPTIONS.map((value) => ({
            value,
            label: t(`flow.settings.appearance.textSize.${value}`),
          }))}
          onChange={setTextSize}
          disabled
          describedBy={TEXT_SIZE_GATE_ID}
        />
        <GateNotice
          id={TEXT_SIZE_GATE_ID}
          className="mt-2"
          reason={t('flow.settings.appearance.textSizeGate')}
        />
      </div>

      <div>
        <p id="appearance-corners-label" className={labelClass} style={labelStyle}>
          {t('flow.settings.appearance.groups.corners')}
        </p>
        <SegmentedControl
          labelledBy="appearance-corners-label"
          value={corners}
          options={CORNER_OPTIONS.map((value) => ({
            value,
            label: t(`flow.settings.appearance.corners.${value}`),
          }))}
          onChange={setCorners}
          onAudition={(value) => setAudition(value === null ? undefined : { corners: value })}
        />
      </div>

      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[13px] font-medium" style={{ color: 'var(--flow-text)' }}>
            {t('flow.settings.appearance.groups.motion')}
          </p>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.settings.appearance.motionHint')}
          </p>
        </div>
        <AppearanceSwitch
          checked={reduceMotion}
          onChange={setReduceMotion}
          label={t('flow.settings.appearance.groups.motion')}
        />
      </div>
    </div>
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
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('common.settings.shortcuts')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.shortcuts.subtitle')}
        </p>
      </div>
      <div className="space-y-1">
        {shortcuts.map((s, i) => (
          <div
            key={i}
            className="flex items-center justify-between py-2 px-3 rounded-sm hover:bg-[var(--flow-hover)]"
          >
            <span className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
              {s.desc}
            </span>
            <div className="flex items-center gap-1">
              {s.keys.map((k, j) => (
                <kbd
                  key={j}
                  className="px-1.5 py-0.5 rounded-sm text-[10px] font-mono font-semibold"
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
  const { data: labels = [], isLoading } = useLabels();
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
      toast.error(t('errors.generic'));
      return;
    }
    createLabel.mutate(
      { name, color: newColor },
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
          {t('common.settings.labels')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.labels.subtitle')}
        </p>
      </div>

      {/* Create form */}
      <div
        className="rounded-xl p-4 space-y-3"
        style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
      >
        <h3 className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
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
            className="flex-1 rounded-lg px-3 py-2 text-[13px] outline-none"
            style={inputStyle}
          />
        </div>
        <div>
          <p className="text-[11px] mb-2" style={{ color: 'var(--flow-text-muted)' }}>
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
                border: '1px solid var(--flow-border)',
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
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          {createLabel.isPending
            ? t('flow.settings.labels.creating')
            : t('flow.settings.labels.createLabel')}
        </button>
      </div>

      {/* Existing labels */}
      <div>
        <h3 className="text-[12px] font-semibold mb-3" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.labels.workspaceLabels')}
        </h3>
        {isLoading ? (
          <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
            {t('common.loading')}
          </p>
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
                  <span className="text-[13px] font-medium" style={{ color: 'var(--flow-text)' }}>
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
        confirmLabel={t('common.delete')}
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

// Exported so its own central permission gate can be unit-tested directly,
// independent of the
// buildSections-level gate that normally keeps a non-privileged role from
// ever mounting this component at all.
export const TeamSection: React.FC = () => {
  const { t } = useTranslation();
  const { data: allMembers = [], isLoading, refetch: refetchMembers } = useFlowTeam();

  // The /team/members endpoint returns the entire business team. Flow only
  // displays the licensed members here. Role and license changes are owned by
  // Dash and are not mutated from this surface.
  const flowMembers = allMembers.filter((m) => m.apps?.includes('flow'));

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.settings.team.section')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.team.subtitle')}
        </p>
      </div>

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
        <EmptyState
          icon={Users}
          title={t('flow.settings.team.noMembersYet')}
          action={{
            label: t('flow.settings.team.refreshCta'),
            onClick: () => void refetchMembers(),
          }}
          size="compact"
        />
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
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// ── Section content map ───────────────────────────────────────────────────────

const SECTION_CONTENT: Record<string, React.FC> = {
  bin: BinSection,
  team: TeamSection,
  notifications: NotificationsSection,
  appearance: AppearanceSection,
  labels: LabelsSection,
  time: TimeSection,
  life: LifeRhythmSection,
  shortcuts: ShortcutsSection,
  sidebar: SidebarSettingsSection,
};

const PlaceholderSection: React.FC<{ id: string }> = ({ id }) => {
  const { t } = useTranslation();
  const sections = buildSections(t, false, false, false);
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <p className="text-[14px] font-medium mb-1" style={{ color: 'var(--flow-text-muted)' }}>
        {sections.find((s) => s.id === id)?.label}
      </p>
      <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.settings.sectionNotAvailable')}
      </p>
    </div>
  );
};

// getRouteApi takes a route id, which for a top-level route is its path.
const settingsRouteApi = getRouteApi('/settings');

const FlowSettings: React.FC = () => {
  const { t } = useTranslation();
  const effectiveAccess = useAuthUserStore((s) => s.effectiveAccess);
  const legacyRole = useAuthUserStore((s) => s.user?.role);
  const canManageWorkspaces = canFlowPermission(
    legacyRole,
    'flow:settings:manage',
    effectiveAccess
  );
  const canManageTeam = canFlowPermission(legacyRole, 'flow:team:manage', effectiveAccess);
  const canManageBin = canManageWorkspaces;
  const sections = useMemo(
    () => buildSections(t, canManageWorkspaces, canManageTeam, canManageBin),
    [t, canManageWorkspaces, canManageTeam, canManageBin]
  );
  const isAccessibleSection = (id: string): boolean => sections.some((s) => s.id === id);
  // ?section=time makes a section linkable, so the sidebar can point at one -
  // but the search param is user-controlled, so it must be checked against
  // the same gate that hides the tab, not rendered on faith (#793).
  const { section } = settingsRouteApi.useSearch();
  const [activeSection, setActiveSection] = useState<SettingsSectionId | string>(() =>
    section && isAccessibleSection(section) ? section : (sections[0]?.id ?? 'notifications')
  );
  useEffect(() => {
    if (section && isAccessibleSection(section)) {
      setActiveSection(section);
    }
    // If `section` names an id this role can't reach, deliberately fall
    // through without updating state - activeSection stays at whatever
    // already-valid value it held (the initializer above already resolved
    // that for first render).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [section]);

  const ActiveComponent =
    SECTION_CONTENT[activeSection] || (() => <PlaceholderSection id={activeSection} />);

  return (
    <div className="flex h-full min-h-0 flex-col md:flex-row">
      <HelmetContainer title={`${t('nav.settings')} · Flow`} />
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
