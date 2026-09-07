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
import { useInviteMember, type TeamMember } from '../../shared/hooks/queries/useTeam';
import { useGroups, useBulkInvite, type BulkResult } from '../../shared/hooks/queries/useGroups';
import { usePOSLocations, useAssignPOSLocation } from '../../shared/hooks/queries/usePOSAccess';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';
import { isAddonEnabled, useAddons } from '../../shared/hooks/queries/useAddons';
import { buildAppRoleAssignments, dashRoleToLegacyRole } from './authorizationDraft';
import { useCustomRoles } from '../../shared/hooks/queries/usePermissions';

const ALL_APPS = [
  { key: 'pos', label: 'POS', desc: 'Point of sale', icon: MonitorSpeaker, color: '#4583C4' },
  { key: 'erp', label: 'ERP', desc: 'Inventory and supply', icon: Package, color: '#10b981' },
  { key: 'hr', label: 'HR', desc: 'Staff and payroll', icon: UserCheck, color: '#f59e0b' },
  { key: 'crm', label: 'CRM', desc: 'Customers and loyalty', icon: UsersIcon, color: '#8b5cf6' },
  {
    key: 'finance',
    label: 'Finance',
    desc: 'Invoices and expenses',
    icon: CreditCard,
    color: '#ef4444',
  },
  { key: 'flow', label: 'Flow', desc: 'Projects and tasks', icon: Layers, color: '#64C5B9' },
  {
    key: 'insights',
    label: 'Insights',
    desc: 'Reports and analytics',
    icon: BarChart3,
    color: '#ec4899',
  },
  { key: 'menu', label: 'Menu', desc: 'Menus and ordering', icon: Utensils, color: '#f97316' },
  {
    key: 'recruit',
    label: 'Recruit',
    desc: 'Jobs and candidates',
    icon: BriefcaseBusiness,
    color: '#06b6d4',
  },
];

const ROLES = [
  { id: 'admin', label: 'Admin', desc: 'Manage business access and assigned applications' },
  { id: 'manager', label: 'Manager', desc: 'Manage assigned work within delegated scope' },
  { id: 'member', label: 'Member', desc: 'Use assigned applications and operational resources' },
  { id: 'viewer', label: 'Viewer', desc: 'Read-only access to assigned applications' },
];

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
  const { data: customRoles = [] } = useCustomRoles();

  const [tab, setTab] = useState<'single' | 'bulk'>('single');

  // Single tab state
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('member');
  const [selectedApps, setSelectedApps] = useState<string[]>([]);
  const [selectedLocations, setSelectedLocations] = useState<number[]>([]);
  const [authMethod, setAuthMethod] = useState<'invite' | 'manual'>('invite');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [singleGroupId, setSingleGroupId] = useState<string>('');
  const [singleStep, setSingleStep] = useState<0 | 1 | 2 | 3>(0);
  const [roleIdsByApp, setRoleIdsByApp] = useState<Record<string, string[]>>({});

  // Bulk tab state
  const [bulkEmailsRaw, setBulkEmailsRaw] = useState('');
  const [bulkRole, setBulkRole] = useState<string>('member');
  const [bulkApps, setBulkApps] = useState<string[]>([]);
  const [bulkLocations, setBulkLocations] = useState<number[]>([]);
  const [bulkGroupId, setBulkGroupId] = useState<string>('');
  const [bulkResults, setBulkResults] = useState<BulkResult[] | null>(null);

  // Location assignment only applies when the POS app is licensed.
  const posSelectedSingle = selectedApps.includes('pos');
  const posSelectedBulk = bulkApps.includes('pos');
  const { data: locations = [] } = usePOSLocations(posSelectedSingle || posSelectedBulk);

  const { data: addonsData = [] } = useAddons();
  const menuAddonEnabled = isAddonEnabled(addonsData, 'menu');

  const allRoles = [
    ...ROLES,
    ...(menuAddonEnabled
      ? [
          {
            id: 'menu_manager',
            label: 'Menu Manager',
            desc: 'Manage menus, QR codes, and storefront design',
          },
        ]
      : []),
  ];

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
  const assignLocations = async (
    userId: number,
    locationIds: number[],
    _memberRole: string
  ): Promise<void> => {
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
      toast.error('Member invited, but some location assignments failed');
    }
  };

  const submitSingle = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    if (!email.trim() || !name.trim()) {
      toast.error('Email and name are required');
      return;
    }
    if (authMethod === 'manual' && (!username.trim() || password.length < 8)) {
      toast.error('Password must be at least 8 characters');
      return;
    }
    try {
      const res = await invite.mutateAsync({
        email,
        fullName: name,
        role: dashRoleToLegacyRole(role) as TeamMember['role'],
        apps: selectedApps,
        dashRole: role,
        appAssignments: buildAppRoleAssignments(
          selectedApps,
          singleGroupId ? Number(singleGroupId) : undefined,
          roleIdsByApp
        ),
        ...(singleGroupId ? { groupIds: [Number(singleGroupId)] } : {}),
        provisioningMode: authMethod === 'manual' ? 'direct_credentials' : 'invitation',
        ...(authMethod === 'manual'
          ? { username: username.trim(), temporaryPassword: password }
          : {}),
      });
      if (posSelectedSingle) {
        const userId = Number(res.userId || 0);
        await assignLocations(userId, selectedLocations, role);
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
          {t('team.inviteUser.appRolesTitle')}
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.inviteUser.appRolesDescription')}
        </p>
      </div>
      {selectedApps.map((appKey) => {
        const options = customRoles.filter((customRole) => customRole.appId === appKey);
        return (
          <div key={appKey} className="space-y-1.5">
            <Label htmlFor={`invite-app-role-${appKey}`}>{appKey.toUpperCase()}</Label>
            <select
              id={`invite-app-role-${appKey}`}
              value={roleIdsByApp[appKey]?.[0] ?? ''}
              onChange={(event) =>
                setRoleIdsByApp((current) => ({
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
              <option value="">{t('team.inviteUser.defaultAppRole')}</option>
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
        apps: bulkApps,
        dashRole: bulkRole,
        role: dashRoleToLegacyRole(bulkRole) as TeamMember['role'],
        appAssignments: buildAppRoleAssignments(
          bulkApps,
          bulkGroupId ? Number(bulkGroupId) : undefined
        ),
        ...(bulkGroupId ? { groupIds: [Number(bulkGroupId)] } : {}),
        ...(bulkGroupId ? { groupId: Number(bulkGroupId) } : {}),
      });
      setBulkResults(res.results);
      if (posSelectedBulk && bulkLocations.length > 0) {
        await Promise.allSettled(
          res.results
            .filter((r) => r.userId && (r.status === 'invited' || r.status === 'existing'))
            .map((r) => assignLocations(Number(r.userId), bulkLocations, bulkRole))
        );
      }
    } catch {
      // onError in hook handles toast
    }
  };

  const renderRoleSelector = (
    currentRole: string,
    onChange: (r: string) => void
  ): React.JSX.Element => (
    <div className="space-y-2">
      {allRoles.map((r) => (
        <button
          key={r.id}
          type="button"
          onClick={() => onChange(r.id)}
          className="flex items-center gap-3 w-full px-4 py-3 rounded-lg text-left transition-colors"
          style={{
            background: currentRole === r.id ? 'rgba(69,131,196,0.06)' : 'transparent',
            border: `1px solid ${
              currentRole === r.id ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'
            }`,
          }}
        >
          <div
            className="h-4 w-4 rounded-full border-2 shrink-0 flex items-center justify-center"
            style={{ borderColor: currentRole === r.id ? '#4583C4' : 'var(--dash-border)' }}
          >
            {currentRole === r.id && (
              <div className="h-2 w-2 rounded-full" style={{ background: '#4583C4' }} />
            )}
          </div>
          <div className="flex-1 min-w-0">
            <p
              className="text-sm font-medium"
              style={{ color: currentRole === r.id ? '#4583C4' : 'var(--dash-text)' }}
            >
              {r.label}
            </p>
            <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
              {r.desc}
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
              {app.label}
            </span>
            <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
              {app.desc}
            </span>
          </button>
        );
      })}
    </div>
  );

  const renderLocationSelector = (
    currentLocations: number[],
    toggle: (id: number) => void,
    currentRole: string,
    onRemovePos: () => void
  ): React.JSX.Element => (
    <div className="space-y-2 pt-1">
      <div>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          POS locations
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {currentRole === 'manager'
            ? 'Locations this manager runs within the assigned POS scope'
            : 'Locations this member can work at'}
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
          Group <span style={{ color: 'var(--dash-text-dim)', fontWeight: 400 }}>(optional)</span>
        </p>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          Add this member to an existing group
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
        <option value="">No group</option>
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
            Add a new member to your organization.
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
                    Personal details
                  </p>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-1.5">
                      <Label htmlFor="inv-name">Full name</Label>
                      <Input
                        id="inv-name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="Jane Doe"
                        required
                      />
                    </div>
                    <div className="space-y-1.5">
                      <Label htmlFor="inv-email">Email address</Label>
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
                    Authentication
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
                      <p className="font-semibold">Send invitation email</p>
                      <p
                        className="text-[11px] mt-0.5 font-normal"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        Member sets their own password
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
                      <p className="font-semibold">Set password manually</p>
                      <p
                        className="text-[11px] mt-0.5 font-normal"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        You share credentials directly
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
                      Role
                    </p>
                    <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      Controls what this member can see and do
                    </p>
                  </div>
                  {renderRoleSelector(role, setRole)}
                </div>

                {/* Licenses */}
                <div
                  className="rounded-xl p-5 space-y-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      Assign licenses
                    </p>
                    <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      Choose which apps this member can access
                    </p>
                  </div>
                  {renderAppSelector(selectedApps, toggleApp)}

                  {posSelectedSingle &&
                    renderLocationSelector(selectedLocations, toggleLocation, role, () =>
                      toggleApp('pos')
                    )}

                  {selectedApps.length > 0 && renderAppRoleSelectors()}

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
                    <strong>{t('team.inviteUser.reviewDashRole')}</strong>
                    <p className="capitalize" style={{ color: 'var(--dash-text-muted)' }}>
                      {role}
                    </p>
                  </div>
                  <div>
                    <strong>{t('team.inviteUser.reviewApps')}</strong>
                    <p style={{ color: 'var(--dash-text-muted)' }}>
                      {selectedApps.length ? selectedApps.join(', ') : t('team.inviteUser.noApps')}
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
                      {selectedApps.length ? selectedApps.join(', ') : t('team.inviteUser.noApps')}
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
                Cancel
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
                  Email addresses
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  Paste one email per line
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
                  Role
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  Applied to all invited members
                </p>
              </div>
              {renderRoleSelector(bulkRole, setBulkRole)}
            </div>

            {/* Licenses */}
            <div
              className="rounded-xl p-5 space-y-4"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <div>
                <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  Assign licenses
                </p>
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  Applied to all invited members
                </p>
              </div>
              {renderAppSelector(bulkApps, toggleBulkApp)}

              {posSelectedBulk &&
                renderLocationSelector(bulkLocations, toggleBulkLocation, bulkRole, () =>
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
                Cancel
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
                  ? 'Sending...'
                  : `Send ${parsedEmails.length > 0 ? parsedEmails.length : ''} invitation${parsedEmails.length !== 1 ? 's' : ''}`}
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
                  <div className="col-span-6">Email</div>
                  <div className="col-span-3">Status</div>
                  <div className="col-span-3">Note</div>
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
