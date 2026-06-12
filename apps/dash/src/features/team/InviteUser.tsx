import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import {
  Send,
  MonitorSpeaker,
  Package,
  UserCheck,
  Users as UsersIcon,
  CreditCard,
  Layers,
  Eye,
  EyeOff,
} from 'lucide-react';
import { Button, Input, Label } from '../../shared/components/ui';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';
import { useInviteMember, type TeamMember } from '../../shared/hooks/queries/useTeam';
import { useGroups, useBulkInvite, type BulkResult } from '../../shared/hooks/queries/useGroups';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';

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
];

const ROLES = [
  { id: 'admin', label: 'Admin', desc: 'Full access to all apps and settings, except billing' },
  {
    id: 'operator',
    label: 'Operator',
    desc: 'Manage day-to-day operations across assigned apps',
  },
  {
    id: 'billing',
    label: 'Billing',
    desc: 'Access to billing, invoices, and payment settings only',
  },
  { id: 'staff', label: 'Staff', desc: 'Use assigned apps only, no admin access' },
  { id: 'viewer', label: 'Viewer', desc: 'Read-only access to assigned apps' },
];

const STATUS_COLORS: Record<BulkResult['status'], string> = {
  invited: '#10b981',
  existing: '#f59e0b',
  error: '#ef4444',
};

const InviteUser: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const invite = useInviteMember();
  const bulkInvite = useBulkInvite();
  const { data: sub } = useMySubscription();
  const { data: groups = [] } = useGroups();

  const [tab, setTab] = useState<'single' | 'bulk'>('single');

  // Single tab state
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<TeamMember['role']>('staff');
  const [selectedApps, setSelectedApps] = useState<string[]>(['pos']);
  const [authMethod, setAuthMethod] = useState<'invite' | 'manual'>('invite');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [singleGroupId, setSingleGroupId] = useState<string>('');

  // Bulk tab state
  const [bulkEmailsRaw, setBulkEmailsRaw] = useState('');
  const [bulkRole, setBulkRole] = useState<string>('staff');
  const [bulkApps, setBulkApps] = useState<string[]>(['pos']);
  const [bulkGroupId, setBulkGroupId] = useState<string>('');
  const [bulkResults, setBulkResults] = useState<BulkResult[] | null>(null);

  const { data: addonsData } = useQuery<Array<{ id: string; isEnabled: boolean }>>({
    queryKey: ['addons'],
    queryFn: async () => {
      try {
        const res = await unwrap(
          safe.get<{ addons?: Array<{ id: string; isEnabled: boolean }> }>('/addons')
        );
        return res.addons ?? [];
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a) => a.id === 'menu' && a.isEnabled);

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

  const submitSingle = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    if (!email.trim() || !name.trim()) {
      toast.error('Email and name are required');
      return;
    }
    try {
      await invite.mutateAsync({
        email,
        fullName: name,
        role,
        apps: selectedApps,
        ...(singleGroupId ? { groupId: Number(singleGroupId) } : {}),
      });
      navigate({ to: '/team' });
    } catch {
      // onError in hook handles toast
    }
  };

  const submitBulk = async (): Promise<void> => {
    if (parsedEmails.length === 0) return;
    try {
      const res = await bulkInvite.mutateAsync({
        emails: parsedEmails,
        role: bulkRole,
        apps: bulkApps,
        ...(bulkGroupId ? { groupId: Number(bulkGroupId) } : {}),
      });
      setBulkResults(res.results);
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
      <HelmetContainer title="Invite member · rtylr" />
      <div className="space-y-6 max-w-2xl">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Invite member
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
              key={t}
              onClick={() => setTab(t)}
              className="px-4 py-2 text-sm font-medium capitalize transition-colors"
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

        {/* Single tab */}
        {tab === 'single' && (
          <form onSubmit={submitSingle} className="space-y-6">
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
                <div className="space-y-1.5">
                  <Label htmlFor="inv-password">Password</Label>
                  <div className="relative">
                    <Input
                      id="inv-password"
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      placeholder="At least 8 characters"
                      required={authMethod === 'manual'}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2"
                      style={{ color: 'var(--dash-text-dim)' }}
                      aria-label={showPassword ? 'Hide password' : 'Show password'}
                    >
                      {showPassword ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                  </div>
                </div>
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
                  Controls what this member can see and do
                </p>
              </div>
              {renderRoleSelector(role, (r) => setRole(r as TeamMember['role']))}
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

              {/* Seat cost preview */}
              {sub &&
                sub.pricePerSeatCents > 0 &&
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

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => navigate({ to: '/team' })}
                className="px-4 py-2 rounded-lg text-sm font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={invite.isPending || !email.trim() || !name.trim()}
                className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-40"
                style={{ background: '#4583C4', color: '#fff' }}
              >
                <Send size={14} />
                {authMethod === 'invite' ? 'Send invitation' : 'Create member'}
              </button>
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
                className="w-full rounded-lg px-3 py-2.5 text-sm resize-y"
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
            </div>

            {/* Group */}
            {renderGroupSelector(bulkGroupId, setBulkGroupId)}

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => navigate({ to: '/team' })}
                className="px-4 py-2 rounded-lg text-sm font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={bulkInvite.isPending || parsedEmails.length === 0}
                onClick={submitBulk}
                className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-40"
                style={{ background: '#4583C4', color: '#fff' }}
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
