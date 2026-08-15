import { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import {
  Shield,
  Smartphone,
  Laptop,
  LogOut,
  CheckCircle2,
  QrCode,
  Key,
  Copy,
  Check,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  Button,
  Input,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../shared/components/ui';
import {
  Breadcrumbs,
  SettingSection,
  SkeletonList,
  ConfirmDialog,
} from '../../shared/components/dash';
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
          Revoke session
        </Button>
      )}
    </div>
  );
};

const Security: React.FC = () => {
  const { data, isLoading } = useSecurity();
  const toggle2fa = useToggleTwoFactor();
  const revoke = useRevokeSession();

  const [setupModalOpen, setSetupModalOpen] = useState(false);
  const [disableConfirmOpen, setDisableConfirmOpen] = useState(false);
  const [step, setStep] = useState<'qr' | 'verify' | 'backup'>('qr');
  const [verificationCode, setVerificationCode] = useState('');
  const [copiedSecret, setCopiedSecret] = useState(false);

  const sampleSecret = 'JBSWY3DPEHPK3PXP';
  const sampleBackupCodes = [
    'a1b2-c3d4-e5f6',
    'f6e5-d4c3-b2a1',
    '9876-5432-10fe',
    'fe01-2345-6789',
    '4321-8765-dcba',
    'abcd-1234-ef56',
  ];

  const handleStartEnable = () => {
    setStep('qr');
    setVerificationCode('');
    setSetupModalOpen(true);
  };

  const handleVerifyCode = async () => {
    if (verificationCode.length !== 6) {
      toast.error('Please enter a 6-digit verification code.');
      return;
    }
    try {
      await toggle2fa.mutateAsync(true);
      setStep('backup');
      toast.success('2FA verified and enabled successfully!');
    } catch {
      toast.error('Failed to enable 2FA.');
    }
  };

  const handleFinishSetup = () => {
    setSetupModalOpen(false);
  };

  const handleDisableConfirm = async () => {
    try {
      await toggle2fa.mutateAsync(false);
      setDisableConfirmOpen(false);
      toast.success('2FA disabled');
    } catch {
      toast.error('Failed to disable 2FA');
    }
  };

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
              onClick={() => {
                if (data.twoFactorEnabled) {
                  setDisableConfirmOpen(true);
                } else {
                  handleStartEnable();
                }
              }}
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

        {/* 2FA Setup Modal */}
        <Dialog open={setupModalOpen} onOpenChange={(open) => !open && setSetupModalOpen(false)}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Set up two-factor authentication</DialogTitle>
              <DialogDescription>
                {step === 'qr' &&
                  'Scan the QR code or enter the secret key into your authenticator app (e.g. Google Authenticator or 1Password).'}
                {step === 'verify' && 'Enter the 6-digit code generated by your authenticator app.'}
                {step === 'backup' &&
                  'Save these recovery codes in a safe place. They allow access if you lose your phone.'}
              </DialogDescription>
            </DialogHeader>

            {step === 'qr' && (
              <div className="space-y-4 py-2">
                <div
                  className="flex flex-col items-center justify-center p-6 border rounded-xl"
                  style={{ background: 'var(--dash-bg)', borderColor: 'var(--dash-border)' }}
                >
                  <div className="h-36 w-36 bg-white p-2 rounded-lg flex items-center justify-center shadow-sm">
                    <QrCode size={120} className="text-black" />
                  </div>
                  <p className="text-xs mt-3 font-medium" style={{ color: 'var(--dash-text-dim)' }}>
                    Scan QR code with your authenticator app
                  </p>
                </div>
                <div className="space-y-1">
                  <label
                    htmlFor="totp-secret"
                    className="text-xs font-semibold"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    Secret Key
                  </label>
                  <div className="flex gap-2">
                    <Input
                      id="totp-secret"
                      value={sampleSecret}
                      readOnly
                      className="font-mono text-xs"
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        navigator.clipboard.writeText(sampleSecret);
                        setCopiedSecret(true);
                        setTimeout(() => setCopiedSecret(false), 2000);
                      }}
                    >
                      {copiedSecret ? <Check size={14} /> : <Copy size={14} />}
                    </Button>
                  </div>
                </div>
              </div>
            )}

            {step === 'verify' && (
              <div className="space-y-3 py-3">
                <label
                  htmlFor="totp-code"
                  className="text-xs font-semibold"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  Verification Code
                </label>
                <Input
                  id="totp-code"
                  type="text"
                  maxLength={6}
                  placeholder="123456"
                  value={verificationCode}
                  onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, ''))}
                  className="font-mono text-center tracking-widest text-lg"
                  autoFocus
                />
              </div>
            )}

            {step === 'backup' && (
              <div className="space-y-3 py-2">
                <div
                  className="grid grid-cols-2 gap-2 p-3 border rounded-xl font-mono text-xs text-center"
                  style={{ background: 'var(--dash-bg)', borderColor: 'var(--dash-border)' }}
                >
                  {sampleBackupCodes.map((code) => (
                    <div
                      key={code}
                      className="p-1.5 rounded"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text)' }}
                    >
                      {code}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <DialogFooter>
              {step === 'qr' && (
                <Button onClick={() => setStep('verify')}>Next: Verify Code</Button>
              )}
              {step === 'verify' && (
                <div className="flex gap-2 w-full justify-end">
                  <Button variant="outline" onClick={() => setStep('qr')}>
                    Back
                  </Button>
                  <Button onClick={handleVerifyCode} disabled={toggle2fa.isPending}>
                    {toggle2fa.isPending ? 'Verifying...' : 'Verify & Enable'}
                  </Button>
                </div>
              )}
              {step === 'backup' && <Button onClick={handleFinishSetup}>Done</Button>}
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <ConfirmDialog
          open={disableConfirmOpen}
          onOpenChange={(o) => !o && setDisableConfirmOpen(false)}
          title="Disable two-factor authentication?"
          description="Your account will no longer require a verification code when logging in. This reduces your account security."
          confirmLabel="Disable 2FA"
          destructive
          isLoading={toggle2fa.isPending}
          onConfirm={handleDisableConfirm}
        />

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
