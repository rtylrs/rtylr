import React from 'react';
import { useTranslation } from 'react-i18next';
import { Input } from '../../shared/components/ui';
import { usePermissionProfiles } from '../../shared/hooks/queries/usePermissions';
import { useUserStore } from '../../shared/stores/userStore';
import { APP_NAMES } from '../../shared/config/appNames';
import { PermissionProfilePicker } from './PermissionProfilePicker';

const GROUP_APP_IDS = ['pos', 'erp', 'hr', 'crm', 'finance', 'flow', 'insights', 'menu', 'recruit'];

export const DEFAULT_GROUP_COLOR = '#6366f1';

const PRESET_COLORS = [
  DEFAULT_GROUP_COLOR,
  '#4583C4',
  '#10b981',
  '#f59e0b',
  '#8b5cf6',
  '#ef4444',
  '#64C5B9',
  '#ec4899',
  '#f97316',
];

interface GroupGeneralFieldsProps {
  name: string;
  description: string;
  color: string;
  onNameChange: (name: string) => void;
  onDescriptionChange: (description: string) => void;
  onColorChange: (color: string) => void;
}

export function GroupGeneralFields({
  name,
  description,
  color,
  onNameChange,
  onDescriptionChange,
  onColorChange,
}: GroupGeneralFieldsProps): React.JSX.Element {
  const { t } = useTranslation();
  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <label
            htmlFor="group-name"
            className="text-xs font-medium"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            {t('team.groupDetail.labelName')}
          </label>
          <Input
            id="group-name"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder={t('team.groupDetail.namePlaceholder')}
          />
        </div>
        <div className="space-y-1.5">
          <label
            htmlFor="group-description"
            className="text-xs font-medium"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            {t('team.groupDetail.labelDescription')}
          </label>
          <Input
            id="group-description"
            value={description}
            onChange={(e) => onDescriptionChange(e.target.value)}
            placeholder={t('team.groupDetail.descriptionPlaceholder')}
          />
        </div>
      </div>
      <div className="space-y-1.5">
        <p className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
          {t('team.groupDetail.labelColor')}
        </p>
        <div className="flex items-center gap-2 flex-wrap">
          {PRESET_COLORS.map((c) => (
            <button
              type="button"
              key={c}
              aria-label={t('team.groupDetail.colorAriaLabel', { color: c })}
              aria-pressed={color === c}
              onClick={() => onColorChange(c)}
              className="h-6 w-6 rounded-full transition-transform hover:scale-110"
              style={{
                background: c,
                outline: color === c ? `2px solid ${c}` : 'none',
                outlineOffset: '2px',
              }}
            />
          ))}
        </div>
      </div>
    </>
  );
}

interface GroupAppAccessProps {
  /** The group's colour, used to tint its licence checkboxes. */
  accentColor: string;
  licensedAppIds: string[];
  profileIdsByApp: Record<string, string[]>;
  onToggleApp: (appId: string) => void;
  onProfilesChange: (appId: string, profileIds: string[]) => void;
}

/**
 * Per-app licence and permission choices for a group. Lists only the apps the
 * business subscribes to, and the active permission profiles from the
 * Permissions page for each of them.
 */
export function GroupAppAccess({
  accentColor,
  licensedAppIds,
  profileIdsByApp,
  onToggleApp,
  onProfilesChange,
}: GroupAppAccessProps): React.JSX.Element {
  const { t } = useTranslation();
  const { data: permissionProfiles = [] } = usePermissionProfiles();
  const subscribedApps = useUserStore((state) => state.user?.effectiveAccess?.subscribedApps);
  // Keep any app the group already licenses visible so it can still be removed.
  const appIds = GROUP_APP_IDS.filter(
    (appId) =>
      !subscribedApps ||
      (subscribedApps as readonly string[]).includes(appId) ||
      licensedAppIds.includes(appId)
  );

  return (
    <>
      <div
        className="portal-list-heading grid grid-cols-12 min-w-[620px] gap-4 items-center px-6 py-3"
        style={{
          color: 'var(--dash-text-dim)',
          borderBottom: '1px solid var(--dash-border)',
        }}
      >
        <div className="col-span-4">{t('team.groupDetail.columnApp')}</div>
        <div className="col-span-2 text-center">{t('team.groupDetail.columnLicense')}</div>
        <div className="col-span-6">{t('team.groupDetail.columnPermissionProfile')}</div>
      </div>

      {appIds.map((appId) => {
        const appName = APP_NAMES[appId as keyof typeof APP_NAMES];
        const appProfiles = permissionProfiles.filter(
          (profile) => profile.appId === appId && !profile.isArchived
        );
        const licensed = licensedAppIds.includes(appId);
        return (
          <div
            key={appId}
            className="grid grid-cols-12 min-w-[620px] gap-4 items-center px-6 py-4 transition-colors hover:bg-[var(--dash-hover)]"
            style={{ borderBottom: '1px solid var(--dash-border)' }}
          >
            <div className="col-span-4">
              <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                {t(appName.short)}
              </span>
            </div>
            <div className="col-span-2 flex justify-center">
              <input
                type="checkbox"
                aria-label={t('team.groupDetail.licenseAppAriaLabel', {
                  app: t(appName.short),
                })}
                checked={licensed}
                onChange={() => onToggleApp(appId)}
                style={{ accentColor }}
              />
            </div>
            <div className="col-span-6">
              <PermissionProfilePicker
                label={t('team.groupDetail.permissionProfileForApp', {
                  app: t(appName.short),
                })}
                profiles={appProfiles}
                selectedIds={profileIdsByApp[appId] ?? []}
                disabled={!licensed}
                onChange={(ids) => onProfilesChange(appId, ids)}
              />
            </div>
          </div>
        );
      })}
    </>
  );
}
