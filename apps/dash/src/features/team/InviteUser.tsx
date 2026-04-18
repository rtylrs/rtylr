import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
import { Button, Input, Label, Select } from '../../shared/components/ui';
import { SettingSection } from '../../shared/components/dash';
import api from '../../shared/lib/api';
import { useInviteMember, type TeamMember } from '../../shared/hooks/queries/useTeam';
import { useMySubscription, centsToDisplay } from '../../shared/hooks/queries/useSubscription';

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

const InviteUser: React.FC = () => {
  const navigate = useNavigate();
  const invite = useInviteMember();
  const { data: sub } = useMySubscription();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<TeamMember['role']>('staff');
  const [selectedApps, setSelectedApps] = useState<string[]>(['pos']);
  const { data: addonsData } = useQuery<Array<{ id: string; isEnabled: boolean }>>({
    queryKey: ['addons'],
    queryFn: async () => {
      const res = await api.get('/addons');
      return res.data?.addons ?? [];
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a) => a.id === 'menu' && a.isEnabled);

  const [authMethod, setAuthMethod] = useState<'invite' | 'manual'>('invite');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const toggleApp = (key: string) => {
    setSelectedApps((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
    );
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !name.trim()) return;
    await invite.mutateAsync({ email, fullName: name, role, apps: selectedApps });
    navigate('/team');
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          Invite member
        </h1>
        <p className="text-sm mt-1" style={{ color: 'var(--dash-text-muted)' }}>
          Add a new member to your organization.
        </p>
      </div>

      <form onSubmit={submit} className="space-y-6">
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
                background: authMethod === 'invite' ? 'rgba(69,131,196,0.08)' : 'var(--dash-bg)',
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
                background: authMethod === 'manual' ? 'rgba(69,131,196,0.08)' : 'var(--dash-bg)',
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
          <div className="space-y-2">
            {[
              {
                id: 'admin',
                label: 'Admin',
                desc: 'Full access to all apps and settings, except billing',
              },
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
              ...(menuAddonEnabled
                ? [
                    {
                      id: 'menu_manager',
                      label: 'Menu Manager',
                      desc: 'Manage menus, QR codes, and storefront design',
                    },
                  ]
                : []),
              { id: 'staff', label: 'Staff', desc: 'Use assigned apps only, no admin access' },
              { id: 'viewer', label: 'Viewer', desc: 'Read-only access to assigned apps' },
            ].map((r) => (
              <button
                key={r.id}
                type="button"
                onClick={() => setRole(r.id as TeamMember['role'])}
                className="flex items-center gap-3 w-full px-4 py-3 rounded-lg text-left transition-colors"
                style={{
                  background: role === r.id ? 'rgba(69,131,196,0.06)' : 'transparent',
                  border: `1px solid ${
                    role === r.id ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'
                  }`,
                }}
              >
                <div
                  className="h-4 w-4 rounded-full border-2 shrink-0 flex items-center justify-center"
                  style={{ borderColor: role === r.id ? '#4583C4' : 'var(--dash-border)' }}
                >
                  {role === r.id && (
                    <div className="h-2 w-2 rounded-full" style={{ background: '#4583C4' }} />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p
                    className="text-sm font-medium"
                    style={{ color: role === r.id ? '#4583C4' : 'var(--dash-text)' }}
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
          <div className="grid grid-cols-3 gap-2">
            {ALL_APPS.map((app) => {
              const Icon = app.icon;
              const checked = selectedApps.includes(app.key);
              return (
                <button
                  type="button"
                  key={app.key}
                  onClick={() => toggleApp(app.key)}
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

          {/* Seat cost preview */}
          {sub && sub.pricePerSeatCents > 0 && (
            <div
              className="flex items-center gap-2 px-3 py-2.5 rounded-lg text-xs"
              style={{
                background: 'var(--accent-muted)',
                border: '1px solid var(--accent-border)',
              }}
            >
              <CreditCard size={12} style={{ color: 'var(--accent)', flexShrink: 0 }} />
              <p style={{ color: 'var(--dash-text-muted)' }}>
                Adding this member will add{' '}
                <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                  +{centsToDisplay(sub.pricePerSeatCents)}/mo
                </span>{' '}
                to your bill · applied at the next billing cycle.
              </p>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate('/team')}
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
    </div>
  );
};

export default InviteUser;
