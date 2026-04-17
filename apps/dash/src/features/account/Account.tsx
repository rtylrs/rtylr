import React, { useState } from 'react';
import { UserCircle, Mail, KeyRound } from 'lucide-react';
import { toast } from 'sonner';
import { Button, Input, Label } from '../../shared/components/ui';
import { Breadcrumbs, DangerZone, SettingSection } from '../../shared/components/dash';
import { useUserStore } from '../../shared/stores/userStore';

const Account: React.FC = () => {
  const { user } = useUserStore();
  const [name, setName] = useState((user as any)?.name ?? '');
  const [email, setEmail] = useState((user as any)?.email ?? (user as any)?.emailAddress ?? '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  return (
    <div className="space-y-5 max-w-2xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(69,131,196,0.15)' }}
        >
          <UserCircle size={16} style={{ color: '#4583C4' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Account
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            Your personal owner account details.
          </p>
        </div>
      </div>

      <SettingSection
        title="Profile"
        description="Your name is visible to your team and customers."
        footer={
          <Button size="sm" onClick={() => toast.success('Profile saved')}>
            Save changes
          </Button>
        }
      >
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label>Full name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label>Display email</Label>
            <Input value={email} readOnly disabled />
          </div>
        </div>
      </SettingSection>

      <SettingSection
        title="Email address"
        description="Used for sign-in and recovery."
        footer={
          <Button
            size="sm"
            variant="secondary"
            onClick={() => toast.info('Verification email sent')}
          >
            Send verification
          </Button>
        }
      >
        <div
          className="flex items-center gap-3 px-3 py-2.5 rounded-lg"
          style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
        >
          <Mail size={14} style={{ color: 'var(--dash-text-dim)' }} />
          <span className="text-sm" style={{ color: 'var(--dash-text)' }}>
            {email || 'Not set'}
          </span>
          <span
            className="ml-auto px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase"
            style={{ background: 'rgba(16,185,129,0.15)', color: '#10b981' }}
          >
            Verified
          </span>
        </div>
      </SettingSection>

      <SettingSection
        title="Password"
        description="Choose a strong password you don't use anywhere else."
        footer={
          <Button
            size="sm"
            onClick={() => {
              if (newPassword !== confirmPassword) {
                toast.error('Passwords do not match');
                return;
              }
              toast.success('Password updated');
              setCurrentPassword('');
              setNewPassword('');
              setConfirmPassword('');
            }}
          >
            <KeyRound size={13} />
            Update password
          </Button>
        }
      >
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label>Current password</Label>
            <Input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>New password</Label>
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label>Confirm new password</Label>
              <Input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>
          </div>
        </div>
      </SettingSection>

      <DangerZone
        actions={[
          {
            label: 'Delete account',
            description: 'Permanently delete your account and all associated data.',
            buttonLabel: 'Delete account',
            onClick: () => toast.error('Contact support to delete your account'),
          },
        ]}
      />
    </div>
  );
};

export default Account;
