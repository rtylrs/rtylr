import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Send } from 'lucide-react';
import { Button, Input, Label, Select } from '../../../components/ui';
import { Breadcrumbs, PermissionCheckbox, SettingSection } from '../../../components/hub';
import { useInviteMember, type TeamMember } from '../../../hooks/queries/useTeam';
import { useApps } from '../../../hooks/queries/useApps';

const InviteUser: React.FC = () => {
  const navigate = useNavigate();
  const invite = useInviteMember();
  const { data: apps = [] } = useApps();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<TeamMember['role']>('staff');
  const [selectedApps, setSelectedApps] = useState<string[]>(['pos']);

  const toggleApp = (key: string) => {
    setSelectedApps((prev) => (prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]));
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !name.trim()) return;
    await invite.mutateAsync({ email, fullName: name, role, apps: selectedApps });
    navigate('/team');
  };

  return (
    <div className="space-y-5 max-w-4xl">

      <div>
        <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Invite member</h1>
        <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
          We will email an invitation link. They can join by setting a password.
        </p>
      </div>

      <form onSubmit={submit} className="space-y-5">
        <SettingSection title="Personal details">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="inv-name">Full name</Label>
              <Input id="inv-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Jane Doe" required />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="inv-email">Email address</Label>
              <Input id="inv-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="jane@example.com" required />
            </div>
          </div>
        </SettingSection>

        <SettingSection title="Role" description="Controls default permissions. You can override permissions later.">
          <Select value={role} onChange={(e) => setRole(e.target.value as TeamMember['role'])}>
            <option value="admin">Admin - full access except billing</option>
            <option value="manager">Manager - team and operations</option>
            <option value="staff">Staff - app access only</option>
            <option value="viewer">Viewer - read only</option>
          </Select>
        </SettingSection>

        <SettingSection title="App access" description="Choose which apps this member can launch.">
          <div className="grid grid-cols-2 gap-3">
            {apps.filter((a) => a.status !== 'coming_soon').map((app) => {
              const Icon = app.icon;
              const checked = selectedApps.includes(app.key);
              return (
                <button
                  type="button"
                  key={app.key}
                  onClick={() => toggleApp(app.key)}
                  className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors"
                  style={{
                    background: checked ? 'rgba(69,131,196,0.08)' : 'var(--hub-hover)',
                    border: `1px solid ${checked ? 'rgba(69,131,196,0.3)' : 'var(--hub-border)'}`,
                  }}
                >
                  <div className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0" style={{ background: app.bg }}>
                    <Icon size={14} style={{ color: app.color }} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-semibold truncate" style={{ color: 'var(--hub-text)' }}>{app.label}</p>
                    <p className="text-[11px] truncate" style={{ color: 'var(--hub-text-dim)' }}>{app.description}</p>
                  </div>
                  <PermissionCheckbox checked={checked} />
                </button>
              );
            })}
          </div>
        </SettingSection>

        <div className="flex items-center justify-end gap-2">
          <Button type="button" variant="secondary" onClick={() => navigate('/team')}>Cancel</Button>
          <Button type="submit" disabled={invite.isPending}>
            <Send size={14} />
            Send invitation
          </Button>
        </div>
      </form>
    </div>
  );
};

export default InviteUser;
