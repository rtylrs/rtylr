import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useNavigate } from '@tanstack/react-router';
import {
  Send,
  MonitorSpeaker,
  Package,
  UserCheck,
  Users as UsersIcon,
  CreditCard,
  Layers,
  BarChart3,
  Utensils,
  BriefcaseBusiness,
  MapPin,
  Eye,
  EyeOff,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Input, Label } from '../../shared/components/ui';
import { useInviteMember } from '../../shared/hooks/queries/useTeam';
import { useGroups, useBulkInvite, type BulkResult } from '../../shared/hooks/queries/useGroups';
import { usePOSLocations, useAssignPOSLocation } from '../../shared/hooks/queries/usePOSAccess';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';
import { buildPermissionProfileAssignments } from './authorizationDraft';
import { usePermissionProfiles } from '../../shared/hooks/queries/usePermissions';
import { APP_NAMES } from '../../shared/config/appNames';

const ALL_APPS = [
  {
    key: 'pos',
    labelKey: APP_NAMES.pos.short,
    descKey: 'appNames.desc.pos',
    icon: MonitorSpeaker,
    color: 'var(--app-pos)',
  },
  {
    key: 'erp',
    labelKey: APP_NAMES.erp.short,
    descKey: 'appNames.desc.erp',
    icon: Package,
    color: 'var(--app-erp)',
  },
  {
    key: 'hr',
    labelKey: APP_NAMES.hr.short,
    descKey: 'appNames.desc.hr',
    icon: UserCheck,
    color: 'var(--app-hr)',
  },
  {
    key: 'crm',
    labelKey: APP_NAMES.crm.short,
    descKey: 'appNames.desc.crm',
    icon: UsersIcon,
    color: 'var(--app-crm)',
  },
  {
    key: 'finance',
    labelKey: APP_NAMES.finance.short,
    descKey: 'appNames.desc.finance',
    icon: CreditCard,
    color: 'var(--app-finance)',
  },
  {
    key: 'flow',
    labelKey: APP_NAMES.flow.short,
    descKey: 'appNames.desc.flow',
    icon: Layers,
    color: 'var(--app-flow)',
  },
  {
    key: 'insights',
    labelKey: APP_NAMES.insights.short,
    descKey: 'appNames.desc.insights',
    icon: BarChart3,
    color: 'var(--app-insights)',
  },
  {
    key: 'menu',
    labelKey: APP_NAMES.menu.short,
    descKey: 'appNames.desc.menu',
    icon: Utensils,
    color: 'var(--app-menu)',
  },
  {
    key: 'recruit',
    labelKey: APP_NAMES.recruit.short,
    descKey: 'appNames.desc.recruit',
    icon: BriefcaseBusiness,
    color: 'var(--app-recruit)',
  },
];

const MEMBER_TYPES = ['admin', 'member'] as const;

const STATUS_COLORS: Record<BulkResult['status'], string> = {
  invited: '#10b981',
  existing: '#f59e0b',
  error: '#ef4444',
};

const InviteUser: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const invite = useInviteMember();
  const bulkInvite = useBulkInvite();
  const assignLocation = useAssignPOSLocation();
  const { data: sub } = useMySubscription();
  const { data: groups = [] } = useGroups();
  const { data: customRoles = [] } = usePermissionProfiles();

  const [tab, setTab] = useState<'single' | 'bulk'>('single');

  // Single tab state
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [memberType, setMemberType] = useState<'member' | 'admin'>('member');
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [selectedLocations, setSelectedLocations] = useState<number[]>([]);
  const [authMethod, setAuthMethod] = useState<'invite' | 'manual'>('invite');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [singleGroupId, setSingleGroupId] = useState<string>('');
  const [singleStep, setSingleStep] = useState<0 | 1 | 2 | 3>(0);
  const [profileIdsByApp, setProfileIdsByApp] = useState<Record<string, string[]>>({});

  // Bulk tab state
  const [bulkEmailsRaw, setBulkEmailsRaw] = useState('');
  const [bulkMemberType, setBulkMemberType] = useState<'member' | 'admin'>('member');
  const [bulkApps, setBulkApps] = useState<string[]>([]);
  const [bulkLocations, setBulkLocations] = useState<number[]>([]);
  const [bulkGroupId, setBulkGroupId] = useState<string>('');
  const [bulkResults, setBulkResults] = useState<BulkResult[] | null>(null);

  // Location assignment only applies when the POS app is licensed.
  const posSelectedSingle = selectedApps.includes('pos');
  const posSelectedBulk = bulkApps.includes('pos');
  const { data: locations = [] } = usePOSLocations(posSelectedSingle || posSelectedBulk);

  const parsedEmails = bulkEmailsRaw
    .split('\n')
    .map((e) => e.trim())
    .filter((e) => e.length > 0 && e.includes('@'));

  const toggleApp = (key: string): void => {
    setSelectedApps((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
    );
  };

  const toggleBulkApp = (key: string): void => {
    setBulkApps((prev) => (prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]));
  };

  const toggleLocation = (id: number): void => {
    setSelectedLocations((prev) =>
      prev.includes(id) ? prev.filter((l) => l !== id) : [...prev, id]
    );
  };

  const toggleBulkLocation = (id: number): void => {
    setBulkLocations((prev) => (prev.includes(id) ? prev.filter((l) => l !== id) : [...prev, id]));
  };

  // Assigns the member to each chosen POS location. Best-effort per location so
  // one failure doesn't abandon the rest; failures surface as a toast.
  const assignLocations = async (userId: number, locationIds: number[]): Promise<void> => {
    if (!userId || locationIds.length === 0) return;
    const results = await Promise.allSettled(
      locationIds.map((locationId) =>
        assignLocation.mutateAsync({
          userId,
          locationId,
          canManagePosUsers: false,
          canManageTerminals: false,
          canViewReports: false,
        })
      )
    );
    if (results.some((r) => r.status === 'rejected')) {
      toast.error(t('team.inviteUser.locationAssignmentsFailed'));
    }
  };

  const submitSingle = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    if (!email.trim() || !name.trim()) {
      toast.error(t('team.inviteUser.emailAndNameRequired'));
      return;
    }
    if (authMethod === 'manual' && (!username.trim() || password.length < 8)) {
      toast.error(t('team.inviteUser.passwordTooShort'));
      return;
    }
    try {
      const res = await invite.mutateAsync({
        email,
        fullName: name,
        memberType,
        licensedAppIds: memberType === 'admin' ? [] : selectedApps,
        permissionProfileAssignments:
          memberType === 'admin'
            ? []
            : buildPermissionProfileAssignments(selectedApps, profileIdsByApp),
        ...(singleGroupId ? { groupIds: [Number(singleGroupId)] } : {}),
        provisioningMode: authMethod === 'manual' ? 'direct_credentials' : 'invitation',
        ...(authMethod === 'manual'
          ? { username: username.trim(), temporaryPassword: password }
          : {}),
      });
      if (memberType === 'member' && posSelectedSingle) {
        const userId = Number(res.userId || 0);
        await assignLocations(userId, selectedLocations);
      }
      navigate({ to: '/team' });
    } catch {
      // onError in hook handles toast
    }
  };

  const continueToAccess = (): void => {
    if (!name.trim() || !email.trim()) {
      toast.error(t('team.inviteUser.identityRequired'));
      return;
    }
    if (authMethod === 'manual' && (!username.trim() || password.length < 8)) {
      toast.error(t('team.inviteUser.credentialsRequired'));
      return;
    }
    setSingleStep(1);
  };

  const renderAppRoleSelectors = (): React.JSX.Element => (
    <div className="space-y-3">
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {t('team.inviteUser.permissionProfilesTitle')}
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.inviteUser.permissionProfilesDescription')}
        </p>
      </div>
      {selectedApps.map((appKey) => {
        const options = customRoles.filter((customRole) => customRole.appId === appKey);
        return (
          <div key={appKey} className="space-y-1.5">
            <Label htmlFor={`invite-app-role-${appKey}`}>{appKey.toUpperCase()}</Label>
            <select
              id={`invite-app-role-${appKey}`}
              value={profileIdsByApp[appKey]?.[0] ?? ''}
              onChange={(event) =>
                setProfileIdsByApp((current) => ({
                  ...current,
                  [appKey]: event.target.value ? [event.target.value] : [],
                }))
              }
              className="w-full rounded-lg px-3 py-2.5 text-base"
              style={{
                background: 'var(--dash-bg)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            >
              <option value="">{t('team.inviteUser.noPermissionProfile')}</option>
              {options.map((customRole) => (
                <option key={customRole.id} value={customRole.id}>
                  {customRole.name}
                </option>
              ))}
            </select>
          </div>
        );
      })}
    </div>
  );

  const submitBulk = async (): Promise<void> => {
    if (parsedEmails.length === 0) return;
    try {
      const res = await bulkInvite.mutateAsync({
        emails: parsedEmails,
        memberType: bulkMemberType,
        licensedAppIds: bulkMemberType === 'admin' ? [] : bulkApps,
        permissionProfileAssignments:
          bulkMemberType === 'admin' ? [] : buildPermissionProfileAssignments(bulkApps, {}),
        ...(bulkGroupId ? { groupIds: [Number(bulkGroupId)] } : {}),
      });
      setBulkResults(res.results);
      if (bulkMemberType === 'member' && posSelectedBulk && bulkLocations.length > 0) {
        await Promise.allSettled(
          res.results
            .filter((r) => r.userId && (r.status === 'invited' || r.status === 'existing'))
            .map((r) => assignLocations(Number(r.userId), bulkLocations))
        );
      }
    } catch {
      // onError in hook handles toast
    }
  };

  const renderMemberTypeSelector = (
    currentMemberType: 'member' | 'admin',
    onChange: (memberType: 'member' | 'admin') => void
  ): React.JSX.Element => (
    <div className="space-y-2">
      {MEMBER_TYPES.map((type) => (
        <button
          key={type}
          type="button"
          onClick={() => onChange(type)}
          className="flex items-center gap-3 w-full px-4 py-3 rounded-lg text-left transition-colors"
          style={{
            background: currentMemberType === type ? 'rgba(69,131,196,0.06)' : 'transparent',
            border: `1px solid ${
              currentMemberType === type ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'
            }`,
          }}
        >
          <div
            className="h-4 w-4 rounded-full border-2 shrink-0 flex items-center justify-center"
            style={{ borderColor: currentMemberType === type ? '#4583C4' : 'var(--dash-border)' }}
          >
            {currentMemberType === type && (
              <div className="h-2 w-2 rounded-full" style={{ background: '#4583C4' }} />
            )}
          </div>
          <div className="flex-1 min-w-0">
            <p
              className="text-sm font-medium"
              style={{ color: currentMemberType === type ? '#4583C4' : 'var(--dash-text)' }}
            >
              {t(`common.roleLabels.${type}`)}
            </p>
            <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
              {type === 'admin'
                ? t('team.inviteUser.adminAccessDescription')
                : t('team.inviteUser.memberAccessDescription')}
            </p>
          </div>
        </button>
      ))}
    </div>
  );

  const renderAppSelector = (
    currentApps: string[],
    toggle: (key: string) => void
  ): React.JSX.Element => (
    <div className="grid grid-cols-3 gap-2">
      {ALL_APPS.map((app) => {
        const Icon = app.icon;
        const checked = currentApps.includes(app.key);
        return (
          <button
            type="button"
            key={app.key}
            onClick={() => toggle(app.key)}
            className="flex flex-col items-center gap-2 py-4 rounded-xl transition-all"
            style={{
              background: checked ? `${app.color}08` : 'var(--dash-bg)',
              border: `1px solid ${checked ? `${app.color}40` : 'var(--dash-border)'}`,
            }}
          >
            <div
              className="h-10 w-10 rounded-xl flex items-center justify-center"
              style={{ background: `${app.color}12` }}
            >
              <Icon size={20} style={{ color: app.color }} />
            </div>
            <span
              className="text-xs font-semibold"
              style={{ color: checked ? app.color : 'var(--dash-text-muted)' }}
            >
              {t(app.labelKey)}
            </span>
            <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
              {t(app.descKey)}
            </span>
          </button>
        );
      })}
    </div>
  );

  const renderLocationSelector = (
    currentLocations: number[],
    toggle: (id: number) => void,
    onRemovePos: () => void
  ): React.JSX.Element => (
    <div className="space-y-2 pt-1">
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {t('team.inviteUser.posLocationsTitle')}
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.inviteUser.posLocationsMemberHint')}
        </p>
      </div>
      {locations.length === 0 ? (
        <EmptyState
          icon={MapPin}
          title={t('team.inviteUser.noPosLocationsTitle')}
          description={t('team.inviteUser.noPosLocationsDescription')}
          action={{ label: t('team.inviteUser.removePosAccess'), onClick: onRemovePos }}
          size="compact"
        />
      ) : (
        <div className="grid grid-cols-2 gap-2">
          {locations.map((loc) => {
            const checked = currentLocations.includes(loc.id);
            return (
              <button
                type="button"
                key={loc.id}
                onClick={() => toggle(loc.id)}
                className="flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-left transition-all"
                style={{
                  background: checked ? 'rgba(69,131,196,0.06)' : 'var(--dash-bg)',
                  border: `1px solid ${checked ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'}`,
                }}
              >
                <div
                  className="h-4 w-4 rounded border-2 shrink-0 flex items-center justify-center"
                  style={{ borderColor: checked ? '#4583C4' : 'var(--dash-border)' }}
                >
                  {checked && (
                    <div className="h-2 w-2 rounded-sm" style={{ background: '#4583C4' }} />
                  )}
                </div>
                <MapPin size={14} style={{ color: checked ? '#4583C4' : 'var(--dash-text-dim)' }} />
                <div className="flex-1 min-w-0">
                  <p
                    className="text-xs font-medium truncate"
                    style={{ color: checked ? '#4583C4' : 'var(--dash-text)' }}
                  >
                    {loc.name}
                  </p>
                  {loc.address && (
                    <p className="text-[10px] truncate" style={{ color: 'var(--dash-text-dim)' }}>
                      {loc.address}
                    </p>
                  )}
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );

  const renderGroupSelector = (value: string, onChange: (v: string) => void): React.JSX.Element => (
    <div
      className="rounded-xl p-5 space-y-3"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {t('team.inviteUser.groupTitle')}{' '}
          <span style={{ color: 'var(--dash-text-dim)', fontWeight: 400 }}>
            {t('team.inviteUser.groupOptional')}
          </span>
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.inviteUser.groupHint')}
        </p>
      </div>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-lg px-3 py-2 text-sm"
        style={{
          background: 'var(--dash-bg)',
          border: '1px solid var(--dash-border)',
          color: 'var(--dash-text)',
        }}
      >
        <option value="">{t('team.inviteUser.noGroup')}</option>
        {groups.map((g) => (
          <option key={g.id} value={String(g.id)}>
            {g.name}
          </option>
        ))}
      </select>
    </div>
  );

  return (
    <>
      <HelmetContainer title={`${t('common.nav.addMember')} · rtylr`} />
      <div className="space-y-6 max-w-2xl">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            {t('common.nav.addMember')}
          </h1>
          <p className="text-sm mt-1" style={{ color: 'var(--dash-text-muted)' }}>
            {t('team.inviteUser.subtitle')}
          </p>
        </div>

        {/* Tabs */}
        <div
          className="flex items-center gap-1"
          style={{ borderBottom: '1px solid var(--dash-border)' }}
        >
          {(['single', 'bulk'] as const).map((t) => (
            <button
              type="button"
              key={t}
              onClick={() => setTab(t)}
              className="min-h-11 px-4 py-2 text-sm font-medium capitalize transition-colors"
              style={{
                color: tab === t ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                borderBottom: tab === t ? '2px solid #4583C4' : '2px solid transparent',
                marginBottom: '-1px',
              }}
            >
              {t === 'single' ? 'Single' : 'Bulk'}
            </button>
          ))}
        </div>

        {tab === 'single' && (
          <div className="grid grid-cols-3 gap-2" aria-label={t('team.inviteUser.progressLabel')}>
            {[
              t('team.inviteUser.stepIdentity'),
              t('team.inviteUser.stepAccess'),
              t('team.inviteUser.stepReview'),
              t('team.inviteUser.stepActivate'),
            ].map((label, index) => (
              <button
                type="button"
                key={label}
                onClick={() => {
                  if (index <= singleStep) setSingleStep(index as 0 | 1 | 2 | 3);
                }}
                className="min-h-11 rounded-lg px-3 py-2 text-left text-xs font-semibold"
                style={{
                  background: index === singleStep ? 'var(--accent-muted)' : 'var(--dash-card)',
                  border: `1px solid ${index <= singleStep ? 'var(--accent-border)' : 'var(--dash-border)'}`,
                  color: index <= singleStep ? 'var(--accent-on-muted)' : 'var(--dash-text-dim)',
                }}
              >
                <span className="block text-[10px] uppercase tracking-wider">
                  {t('team.inviteUser.stepLabel', { number: index + 1 })}
                </span>
                {label}
              </button>
            ))}
          </div>
        )}

        {/* Single tab */}
        {tab === 'single' && (
          <form onSubmit={submitSingle} className="space-y-6">
            {singleStep === 0 && (
              <>
                {/* Personal details */}
                <div
                  className="rounded-xl p-5 space-y-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {t('team.inviteUser.personalDetails')}
                  </p>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-1.5">
                      <Label htmlFor="inv-name">{t('team.inviteUser.fullNameLabel')}</Label>
                      <Input
                        id="inv-name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder={t('team.inviteUser.fullNamePlaceholder')}
                        required
                      />
                    </div>
                    <div className="space-y-1.5">
                      <Label htmlFor="inv-email">{t('team.inviteUser.emailLabel')}</Label>
                      <Input
                        id="inv-email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="jane@example.com"
                        required
                      />
                    </div>
                  </div>
                </div>

                {/* Authentication */}
                <div
                  className="rounded-xl p-5 space-y-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {t('team.inviteUser.authenticationTitle')}
                  </p>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setAuthMethod('invite')}
                      className="flex-1 px-3 py-2.5 rounded-lg text-xs font-medium text-left transition-colors"
                      style={{
                        background:
                          authMethod === 'invite' ? 'rgba(69,131,196,0.08)' : 'var(--dash-bg)',
                        border: `1px solid ${
                          authMethod === 'invite' ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'
                        }`,
                        color: authMethod === 'invite' ? '#4583C4' : 'var(--dash-text-muted)',
                      }}
                    >
                      <p className="font-semibold">{t('team.inviteUser.sendInvitationEmail')}</p>
                      <p
                        className="text-[11px] mt-0.5 font-normal"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {t('team.inviteUser.sendInvitationEmailHint')}
                      </p>
                    </button>
                    <button
                      type="button"
                      onClick={() => setAuthMethod('manual')}
                      className="flex-1 px-3 py-2.5 rounded-lg text-xs font-medium text-left transition-colors"
                      style={{
                        background:
                          authMethod === 'manual' ? 'rgba(69,131,196,0.08)' : 'var(--dash-bg)',
                        border: `1px solid ${
                          authMethod === 'manual' ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'
                        }`,
                        color: authMethod === 'manual' ? '#4583C4' : 'var(--dash-text-muted)',
                      }}
                    >
                      <p className="font-semibold">{t('team.inviteUser.setPasswordManually')}</p>
                      <p
                        className="text-[11px] mt-0.5 font-normal"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {t('team.inviteUser.setPasswordManuallyHint')}
                      </p>
                    </button>
                  </div>

                  {authMethod === 'manual' && (
                    <div className="space-y-3">
                      <div className="space-y-1.5">
                        <Label htmlFor="inv-username">{t('team.inviteUser.usernameLabel')}</Label>
                        <Input
                          id="inv-username"
                          value={username}
                          onChange={(e) => setUsername(e.target.value)}
                          placeholder={t('team.inviteUser.usernamePlaceholder')}
                          required={authMethod === 'manual'}
                        />
                      </div>
                      <div className="space-y-1.5">
                        <Label htmlFor="inv-password">
                          {t('team.inviteUser.temporaryPasswordLabel')}
                        </Label>
                        <div className="relative">
                          <Input
                            id="inv-password"
                            type={showPassword ? 'text' : 'password'}
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder={t('team.inviteUser.passwordPlaceholder')}
                            required={authMethod === 'manual'}
                            minLength={8}
                          />
                          <button
                            type="button"
                            onClick={() => setShowPassword(!showPassword)}
                            className="absolute right-2.5 top-1/2 -translate-y-1/2 min-h-11 min-w-11 flex items-center justify-center"
                            style={{ color: 'var(--dash-text-dim)' }}
                            aria-label={
                              showPassword
                                ? t('team.inviteUser.hidePassword')
                                : t('team.inviteUser.showPassword')
                            }
                          >
                            {showPassword ? <EyeOff size={14} /> : <Eye size={14} />}
                          </button>
                        </div>
                      </div>
                      <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                        {t('team.inviteUser.temporaryPasswordHint')}
                      </p>
                    </div>
                  )}
                </div>
              </>
            )}

            {singleStep === 1 && (
              <>
                {/* Role */}
                <div
                  className="rounded-xl p-5 space-y-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {t('team.inviteUser.roleTitle')}
                    </p>
                    <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      {t('team.inviteUser.roleHint')}
                    </p>
                  </div>
                  {renderMemberTypeSelector(memberType, setMemberType)}
                </div>

                {/* Licenses */}
                <div
                  className="rounded-xl p-5 space-y-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {t('team.inviteUser.licensesTitle')}
                    </p>
                    <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      {t('team.inviteUser.licensesHint')}
                    </p>
                  </div>
                  {memberType === 'admin' ? (
                    <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                      {t('team.inviteUser.adminReceivesFullAccess')}
                    </p>
                  ) : (
                    renderAppSelector(selectedApps, toggleApp)
                  )}

                  {memberType === 'member' &&
                    posSelectedSingle &&
                    renderLocationSelector(selectedLocations, toggleLocation, () =>
                      toggleApp('pos')
                    )}

                  {memberType === 'member' && selectedApps.length > 0 && renderAppRoleSelectors()}

                  {/* Seat cost preview */}
                  {sub &&
                    sub.pricePerSeatCents > 0 &&
                    sub.seatCount >= sub.includedSeatLimit &&
                    (() => {
                      const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                        computeSeatProration(sub.pricePerSeatCents);
                      return (
                        <div
                          className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs"
                          style={{
                            background: 'var(--accent-muted)',
                            border: '1px solid var(--accent-border)',
                          }}
                        >
                          <CreditCard
                            size={12}
                            style={{ color: 'var(--accent)', flexShrink: 0, marginTop: 1 }}
                          />
                          <p style={{ color: 'var(--dash-text-muted)' }}>
                            Adding this member adds{' '}
                            <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                              +{centsToDisplay(proratedCents)}
                            </span>{' '}
                            to this month ({daysRemaining} of {daysInMonth} days remaining), then{' '}
                            <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                              +{centsToDisplay(sub.pricePerSeatCents)}/mo
                            </span>{' '}
                            from {nextMonthName} 1st.
                          </p>
                        </div>
                      );
                    })()}
                </div>

                {/* Group */}
                {renderGroupSelector(singleGroupId, setSingleGroupId)}
              </>
            )}

            {singleStep === 2 && (
              <div
                className="rounded-xl p-5 space-y-4"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              >
                <div>
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {t('team.inviteUser.reviewTitle')}
                  </p>
                  <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
                    {t('team.inviteUser.reviewDescription')}
                  </p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                  <div>
                    <strong>{t('team.inviteUser.reviewIdentity')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {name} · {email}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.reviewMemberType')}</strong>
                    <p className="capitalize" style={{ color: 'var(--dash-text-muted)' }}>
                      {memberType}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.reviewApps')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {memberType === 'admin'
                        ? t('team.inviteUser.adminReceivesFullAccess')
                        : selectedApps.length
                          ? selectedApps.join(', ')
                          : t('team.inviteUser.noApps')}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.reviewSignIn')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {authMethod === 'invite'
                        ? t('team.inviteUser.invitationMethod')
                        : t('team.inviteUser.directCredentialMethod')}
                    </p>
                  </div>
                </div>
                <div
                  className="rounded-lg px-3 py-2.5 text-xs"
                  style={{
                    background: 'var(--accent-muted)',
                    border: '1px solid var(--accent-border)',
                    color: 'var(--dash-text-muted)',
                  }}
                >
                  {t('team.inviteUser.activationGuard')}
                </div>
              </div>
            )}

            {singleStep === 3 && (
              <div
                className="rounded-xl p-5 space-y-4"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              >
                <div>
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {t('team.inviteUser.activationTitle')}
                  </p>
                  <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
                    {t('team.inviteUser.activationDescription')}
                  </p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                  <div>
                    <strong>{t('team.inviteUser.activationMember')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>{name}</p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.activationState')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {authMethod === 'invite'
                        ? t('team.inviteUser.activationPendingInvite')
                        : t('team.inviteUser.activationPendingVerification')}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.activationApps')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {memberType === 'admin'
                        ? t('team.inviteUser.adminReceivesFullAccess')
                        : selectedApps.length
                          ? selectedApps.join(', ')
                          : t('team.inviteUser.noApps')}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.activationOwnership')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {t('team.inviteUser.activationOwnershipValue')}
                    </p>
                  </div>
                </div>
                <div
                  className="rounded-lg px-3 py-2.5 text-xs"
                  style={{
                    background: 'var(--accent-muted)',
                    border: '1px solid var(--accent-border)',
                    color: 'var(--dash-text-muted)',
                  }}
                >
                  {t('team.inviteUser.activationGuard')}
                </div>
              </div>
            )}

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => navigate({ to: '/team' })}
                className="min-h-11 px-4 py-2 rounded-lg text-sm font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              {singleStep > 0 && (
                <button
                  type="button"
                  onClick={() => setSingleStep((singleStep - 1) as 0 | 1 | 2 | 3)}
                  className="min-h-11 px-4 py-2 rounded-lg text-sm font-medium"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  {t('team.inviteUser.backButton')}
                </button>
              )}
              {singleStep < 3 ? (
                <button
                  type="button"
                  onClick={() => {
                    if (singleStep === 0) {
                      continueToAccess();
                    } else {
                      setSingleStep((singleStep + 1) as 0 | 1 | 2 | 3);
                    }
                  }}
                  className="flex min-h-11 items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-40"
                  style={{ background: 'var(--accent)', color: 'var(--dash-on-accent)' }}
                >
                  {singleStep === 2
                    ? t('team.inviteUser.activateStepButton')
                    : t('team.inviteUser.nextButton')}
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={invite.isPending}
                  className="flex min-h-11 items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-40"
                  style={{ background: 'var(--accent)', color: 'var(--dash-on-accent)' }}
                >
                  <Send size={14} />
                  {authMethod === 'invite'
                    ? t('team.inviteUser.sendInvitation')
                    : t('team.inviteUser.createMember')}
                </button>
              )}
            </div>
          </form>
        )}

        {/* Bulk tab */}
        {tab === 'bulk' && (
          <div className="space-y-6">
            {/* Email list */}
            <div
              className="rounded-xl p-5 space-y-4"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <div>
                <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {t('team.inviteUser.bulkEmailsTitle')}
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('team.inviteUser.bulkEmailsHint')}
                </p>
              </div>
              <textarea
                value={bulkEmailsRaw}
                onChange={(e) => setBulkEmailsRaw(e.target.value)}
                placeholder="jane@example.com&#10;john@example.com&#10;alex@example.com"
                rows={6}
                className="w-full rounded-lg px-3 py-2.5 text-base resize-y"
                style={{
                  background: 'var(--dash-bg)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                  fontFamily: 'monospace',
                }}
              />
              {parsedEmails.length > 0 && (
                <p className="text-[11px]" style={{ color: '#10b981' }}>
                  {parsedEmails.length} valid email{parsedEmails.length !== 1 ? 's' : ''} detected
                </p>
              )}
            </div>

            {/* Role */}
            <div
              className="rounded-xl p-5 space-y-4"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <div>
                <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {t('team.inviteUser.roleTitle')}
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('team.inviteUser.appliedToAllInvited')}
                </p>
              </div>
              {renderMemberTypeSelector(bulkMemberType, setBulkMemberType)}
            </div>

            {/* Licenses */}
            <div
              className="rounded-xl p-5 space-y-4"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <div>
                <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {t('team.inviteUser.licensesTitle')}
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('team.inviteUser.appliedToAllInvited')}
                </p>
              </div>
              {bulkMemberType === 'admin' ? (
                <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                  {t('team.inviteUser.adminReceivesFullAccess')}
                </p>
              ) : (
                renderAppSelector(bulkApps, toggleBulkApp)
              )}

              {bulkMemberType === 'member' &&
                posSelectedBulk &&
                renderLocationSelector(bulkLocations, toggleBulkLocation, () =>
                  toggleBulkApp('pos')
                )}
            </div>

            {/* Group */}
            {renderGroupSelector(bulkGroupId, setBulkGroupId)}

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => navigate({ to: '/team' })}
                className="min-h-11 px-4 py-2 rounded-lg text-sm font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                disabled={bulkInvite.isPending || parsedEmails.length === 0}
                onClick={submitBulk}
                className="flex min-h-11 items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-40"
                style={{ background: 'var(--accent)', color: 'var(--dash-on-accent)' }}
              >
                <Send size={14} />
                {bulkInvite.isPending
                  ? t('team.inviteUser.bulkSending')
                  : t('team.inviteUser.bulkSend', { count: parsedEmails.length })}
              </button>
            </div>

            {/* Bulk results */}
            {bulkResults && (
              <div
                className="rounded-xl overflow-hidden"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              >
                <div
                  className="px-5 py-3 text-[10px] font-semibold uppercase tracking-wider grid grid-cols-12 gap-3"
                  style={{
                    color: 'var(--dash-text-dim)',
                    borderBottom: '1px solid var(--dash-border)',
                  }}
                >
                  <div className="col-span-6">{t('team.inviteUser.resultEmail')}</div>
                  <div className="col-span-3">{t('team.inviteUser.resultStatus')}</div>
                  <div className="col-span-3">{t('team.inviteUser.resultNote')}</div>
                </div>
                {bulkResults.map((r, idx) => (
                  <div
                    key={idx}
                    className="grid grid-cols-12 gap-3 px-5 py-3 items-center"
                    style={{ borderBottom: '1px solid var(--dash-border)' }}
                  >
                    <div
                      className="col-span-6 text-sm truncate"
                      style={{ color: 'var(--dash-text)' }}
                    >
                      {r.email}
                    </div>
                    <div className="col-span-3">
                      <span
                        className="text-[10px] px-1.5 py-0.5 rounded-md font-medium capitalize"
                        style={{
                          background: `${STATUS_COLORS[r.status]}18`,
                          color: STATUS_COLORS[r.status],
                        }}
                      >
                        {r.status}
                      </span>
                    </div>
                    <div
                      className="col-span-3 text-[11px] truncate"
                      style={{ color: 'var(--dash-text-dim)' }}
                    >
                      {r.message}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </>
  );
};

export default InviteUser;
