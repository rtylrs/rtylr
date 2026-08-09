import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Shield, Smartphone, Laptop, LogOut, CheckCircle2 } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, SettingSection, SkeletonList } from '../../shared/components/dash';
import {
  useSecurity,
  useToggleTwoFactor,
  useRevokeSession,
  type Session,
} from '../../shared/hooks/queries/useSecurity';

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

const SessionRow: React.FC<{ session: Session; onRevoke: (id: string) => void }> = ({
  session,
  onRevoke,
}) => {
  const isMobile = /iphone|android|ipad/i.test(session.device);
  const Icon = isMobile ? Smartphone : Laptop;
  return (
    <div className="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)]">
      <div
        className="h-9 w-9 rounded-lg flex items-center justify-center shrink-0"
        style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
      >
        <Icon size={14} style={{ color: 'var(--dash-text-muted)' }} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {session.device}
          </p>
          {session.current && (
            <span
              className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase"
              style={{ background: 'rgba(16,185,129,0.15)', color: '#10b981' }}
            >
              <CheckCircle2 size={9} /> This device
            </span>
          )}
        </div>
        <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
          {session.browser} · {session.location} · {session.ip} ·{' '}
          {formatRelative(session.lastActive)}
        </p>
      </div>
      {!session.current && (
        <Button size="sm" variant="secondary" onClick={() => onRevoke(session.id)}>
          <LogOut size={12} />
          Sign out everywhere
        </Button>
      )}
    </div>
  );
};

const Security: React.FC = () => {
  const { data, isLoading } = useSecurity();
  const toggle2fa = useToggleTwoFactor();
  const revoke = useRevokeSession();

  if (isLoading || !data) {
    return (
      <div className="space-y-5 max-w-3xl">
        <SkeletonList rows={4} withAvatar={false} />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Security · rtylr" />
      <div className="space-y-5 max-w-3xl">
        <div className="flex items-center gap-3">
          <div
            className="h-9 w-9 rounded-xl flex items-center justify-center"
            style={{ background: 'rgba(239,68,68,0.15)' }}
          >
            <Shield size={16} style={{ color: '#ef4444' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              Security
            </h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              Two-factor authentication, sessions, and trusted devices.
            </p>
          </div>
        </div>

        <SettingSection
          title="Two-factor authentication"
          description="Add an extra layer of security to your account."
          footer={
            <Button
              size="sm"
              variant={data.twoFactorEnabled ? 'secondary' : 'default'}
              onClick={() => toggle2fa.mutate(!data.twoFactorEnabled)}
              disabled={toggle2fa.isPending}
            >
              {data.twoFactorEnabled ? 'Disable 2FA' : 'Enable 2FA'}
            </Button>
          }
        >
          <div className="flex items-center gap-3">
            <div
              className="h-10 w-10 rounded-lg flex items-center justify-center"
              style={{
                background: data.twoFactorEnabled
                  ? 'rgba(16,185,129,0.15)'
                  : 'rgba(239,68,68,0.15)',
              }}
            >
              <Shield size={16} style={{ color: data.twoFactorEnabled ? '#10b981' : '#ef4444' }} />
            </div>
            <div>
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {data.twoFactorEnabled ? 'Enabled' : 'Not enabled'}
              </p>
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {data.twoFactorEnabled
                  ? `Using ${data.twoFactorMethod === 'sms' ? 'SMS codes' : 'authenticator app'}`
                  : 'Protect your account against password theft.'}
              </p>
            </div>
          </div>
        </SettingSection>

        <SettingSection
          title="Active sessions"
          description="Devices currently signed in to your account."
        >
          <div className="-mx-5">
            {data.sessions.map((s) => (
              <SessionRow key={s.id} session={s} onRevoke={(id) => revoke.mutate(id)} />
            ))}
          </div>
        </SettingSection>

        <SettingSection title="Trusted devices" description="Devices that skip 2FA on login.">
          <div className="space-y-2">
            {data.trustedDevices.map((d) => (
              <div
                key={d.id}
                className="flex items-center gap-3 px-3 py-2 rounded-lg"
                style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
              >
                <Laptop size={13} style={{ color: 'var(--dash-text-dim)' }} />
                <div className="flex-1">
                  <p className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {d.name}
                  </p>
                  <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                    {d.os} · seen {formatRelative(d.lastSeen)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </SettingSection>
      </div>
    </>
  );
};

export default Security;
