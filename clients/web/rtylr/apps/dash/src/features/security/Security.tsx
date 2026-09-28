import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { QRCode } from 'react-qrcode-logo';
import { Shield, Smartphone, Laptop, LogOut, CheckCircle2, Copy, Check } from 'lucide-react';
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
import { SettingSection, SkeletonList, ConfirmDialog } from '../../shared/components/dash';
import {
  useSecurity,
  useToggleTwoFactor,
  useUpdateOrganizationMFAEnforcement,
  useBeginMFAEnrollment,
  useCompleteMFAEnrollment,
  useVerifyMFA,
  useMFARecoveryCode,
  useRevokeAllSessions,
  type Session,
} from '../../shared/hooks/queries/useSecurity';
import { useUserStore } from '../../shared/stores/userStore';

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

const SessionRow: React.FC<{ session: Session }> = ({ session }) => {
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
    </div>
  );
};

const Security: React.FC = () => {
  const { t } = useTranslation('common');
  const { data, isLoading, isError, refetch } = useSecurity();
  const currentUser = useUserStore((state) => state.user);
  const setUser = useUserStore((state) => state.setUser);
  const toggle2fa = useToggleTwoFactor();
  const updateOrganizationMFA = useUpdateOrganizationMFAEnforcement();
  const beginMFA = useBeginMFAEnrollment();
  const completeMFA = useCompleteMFAEnrollment();
  const verifyMFA = useVerifyMFA();
  const recoveryMFA = useMFARecoveryCode();
  const revokeAllSessions = useRevokeAllSessions();

  const [setupModalOpen, setSetupModalOpen] = useState(false);
  const [disableConfirmOpen, setDisableConfirmOpen] = useState(false);
  const [recoveryModalOpen, setRecoveryModalOpen] = useState(false);
  const [step, setStep] = useState<'qr' | 'verify' | 'backup'>('qr');
  const [verificationCode, setVerificationCode] = useState('');
  const [challengeCode, setChallengeCode] = useState('');
  const [recoveryCode, setRecoveryCode] = useState('');
  const [enrollment, setEnrollment] = useState<{
    secret: string;
    otpauthUri: string;
  } | null>(null);
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [copiedSecret, setCopiedSecret] = useState(false);
  const copyResetTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    return () => {
      if (copyResetTimer.current) clearTimeout(copyResetTimer.current);
    };
  }, []);

  const handleCopySecret = async () => {
    if (!enrollment?.secret) return;
    try {
      await navigator.clipboard.writeText(enrollment.secret);
      setCopiedSecret(true);
      if (copyResetTimer.current) clearTimeout(copyResetTimer.current);
      copyResetTimer.current = setTimeout(() => setCopiedSecret(false), 2000);
    } catch {
      toast.error(t('common.copyFailed', 'Failed to copy to clipboard'));
    }
  };
  const mfaEnabled = data?.mfaEnrolled || data?.twoFactorEnabled;
  const mfaButtonLabel = data?.mfaRequired
    ? mfaEnabled && data.organizationMfaEnforced
      ? t('common.securityPolicy.requiredByOrganization')
      : mfaEnabled
        ? 'Disable MFA'
        : t('common.securityPolicy.setupRequired')
    : mfaEnabled
      ? 'Disable MFA'
      : 'Enable MFA';
  const needsMFAChallenge =
    data?.mfaRequired && data.mfaEnrolled && currentUser?.provisioningStatus?.mfaVerified !== true;

  const markMFAVerified = (enrolled = true) => {
    if (!currentUser) return;
    setUser({
      ...currentUser,
      provisioningStatus: currentUser.provisioningStatus
        ? { ...currentUser.provisioningStatus, mfaEnrolled: enrolled, mfaVerified: true }
        : undefined,
    });
  };

  const handleStartEnable = async () => {
    setStep('qr');
    setVerificationCode('');
    setRecoveryCodes([]);
    try {
      const nextEnrollment = await beginMFA.mutateAsync();
      setEnrollment(nextEnrollment);
      setSetupModalOpen(true);
    } catch {
      toast.error('Failed to start MFA enrollment.');
    }
  };

  const handleVerifyCode = async () => {
    if (verificationCode.length !== 6) {
      toast.error('Please enter a 6-digit verification code.');
      return;
    }
    try {
      const codes = await completeMFA.mutateAsync(verificationCode);
      setRecoveryCodes(codes);
      setStep('backup');
      markMFAVerified();
      toast.success('MFA verified and enabled successfully.');
    } catch {
      toast.error('The MFA code is invalid or enrollment has expired.');
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

  const handleRecoveryCode = async () => {
    if (!recoveryCode.trim()) {
      toast.error('Enter a recovery code.');
      return;
    }
    try {
      await recoveryMFA.mutateAsync(recoveryCode.trim());
      setRecoveryCode('');
      setRecoveryModalOpen(false);
      markMFAVerified();
      toast.success('Recovery code accepted for this session.');
    } catch {
      toast.error('The recovery code is invalid or has already been used.');
    }
  };

  const handleMFAChallenge = async () => {
    if (challengeCode.length !== 6) {
      toast.error(t('common.mfaChallenge.invalidCode'));
      return;
    }
    try {
      await verifyMFA.mutateAsync(challengeCode);
      setChallengeCode('');
      markMFAVerified();
      toast.success(t('common.mfaChallenge.verified'));
    } catch {
      toast.error(t('common.mfaChallenge.invalidCode'));
    }
  };

  if (isError) {
    return (
      <div className="space-y-5 max-w-3xl">
        <div
          className="rounded-xl p-8 flex flex-col items-center justify-center gap-3"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <Shield size={28} style={{ color: 'var(--dash-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
            Failed to load security settings.
          </p>
          <Button size="sm" variant="secondary" onClick={() => void refetch()}>
            Try again
          </Button>
        </div>
      </div>
    );
  }

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
          description={
            data.mfaRequired
              ? data.organizationMfaEnforced
                ? t('common.securityPolicy.organizationRequiredDescription')
                : t('common.securityPolicy.accountRequiredDescription')
              : 'Add an extra layer of security to your account.'
          }
          footer={
            <Button
              size="sm"
              variant={mfaEnabled ? 'secondary' : 'default'}
              onClick={() => {
                if (mfaEnabled) {
                  setDisableConfirmOpen(true);
                } else {
                  handleStartEnable();
                }
              }}
              disabled={
                (mfaEnabled && data.organizationMfaEnforced) ||
                toggle2fa.isPending ||
                beginMFA.isPending
              }
            >
              {mfaButtonLabel}
            </Button>
          }
        >
          <div className="flex items-center gap-3">
            <div
              className="h-10 w-10 rounded-lg flex items-center justify-center"
              style={{
                background: mfaEnabled ? 'rgba(16,185,129,0.15)' : 'rgba(239,68,68,0.15)',
              }}
            >
              <Shield size={16} style={{ color: mfaEnabled ? '#10b981' : '#ef4444' }} />
            </div>
            <div>
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {mfaEnabled
                  ? 'Enabled'
                  : data.organizationMfaEnforced
                    ? t('common.securityPolicy.enrollmentRequired')
                    : 'Not enabled'}
              </p>
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {mfaEnabled
                  ? 'Using an authenticator app with one-time recovery codes'
                  : data.mfaRequired
                    ? data.organizationMfaEnforced
                      ? t('common.securityPolicy.organizationRequiredDescription')
                      : t('common.securityPolicy.accountRequiredDescription')
                    : 'Protect your account against password theft.'}
              </p>
            </div>
          </div>
        </SettingSection>

        {data.canManageOrganizationMfa && (
          <SettingSection
            title={t('common.securityPolicy.organizationTitle')}
            description={t('common.securityPolicy.organizationDescription')}
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                {data.organizationMfaEnforced
                  ? t('common.securityPolicy.organizationEnabled')
                  : t('common.securityPolicy.organizationDisabled')}
              </p>
              <Button
                size="sm"
                variant={data.organizationMfaEnforced ? 'secondary' : 'default'}
                onClick={() => updateOrganizationMFA.mutate(!data.organizationMfaEnforced)}
                disabled={updateOrganizationMFA.isPending}
              >
                {data.organizationMfaEnforced
                  ? t('common.securityPolicy.stopRequiring')
                  : t('common.securityPolicy.requireForEveryone')}
              </Button>
            </div>
          </SettingSection>
        )}

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
                  {enrollment?.otpauthUri ? (
                    <QRCode value={enrollment.otpauthUri} size={176} quietZone={8} />
                  ) : (
                    <div className="h-36 w-36 flex items-center justify-center">
                      <span className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                        Preparing secure enrollment…
                      </span>
                    </div>
                  )}
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
                      value={enrollment?.secret ?? ''}
                      readOnly
                      className="font-mono text-xs"
                    />
                    <Button size="sm" variant="outline" onClick={() => void handleCopySecret()}>
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
                  {recoveryCodes.map((code) => (
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
                  <Button onClick={handleVerifyCode} disabled={completeMFA.isPending}>
                    {completeMFA.isPending ? 'Verifying...' : 'Verify & Enable'}
                  </Button>
                </div>
              )}
              {step === 'backup' && <Button onClick={handleFinishSetup}>Done</Button>}
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {needsMFAChallenge && (
          <SettingSection
            title={t('common.mfaChallenge.title')}
            description={t('common.mfaChallenge.description')}
            footer={
              <Button size="sm" onClick={handleMFAChallenge} disabled={verifyMFA.isPending}>
                {verifyMFA.isPending
                  ? t('common.mfaChallenge.verifying')
                  : t('common.mfaChallenge.verify')}
              </Button>
            }
          >
            <label
              htmlFor="mfa-challenge-code"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              {t('common.mfaChallenge.codeLabel')}
            </label>
            <Input
              id="mfa-challenge-code"
              type="text"
              inputMode="numeric"
              maxLength={6}
              placeholder={t('common.mfaChallenge.codePlaceholder')}
              value={challengeCode}
              onChange={(event) => setChallengeCode(event.target.value.replace(/\D/g, ''))}
              className="mt-2 font-mono text-center tracking-widest text-lg"
              autoFocus
            />
          </SettingSection>
        )}

        {mfaEnabled && (
          <SettingSection
            title="Recovery codes"
            description="Use one single-use code when your authenticator is unavailable."
            footer={
              <Button size="sm" variant="secondary" onClick={() => setRecoveryModalOpen(true)}>
                Use recovery code
              </Button>
            }
          >
            <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              Each code can be used once and is invalidated immediately after use.
            </p>
          </SettingSection>
        )}

        <Dialog open={recoveryModalOpen} onOpenChange={setRecoveryModalOpen}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Use a recovery code</DialogTitle>
              <DialogDescription>
                Enter one of the single-use recovery codes saved during MFA enrollment.
              </DialogDescription>
            </DialogHeader>
            <div className="py-3">
              <label htmlFor="mfa-recovery-code" className="text-xs font-semibold">
                Recovery code
              </label>
              <Input
                id="mfa-recovery-code"
                value={recoveryCode}
                onChange={(event) => setRecoveryCode(event.target.value)}
                autoFocus
                className="mt-2 font-mono"
              />
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setRecoveryModalOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleRecoveryCode} disabled={recoveryMFA.isPending}>
                {recoveryMFA.isPending ? 'Verifying...' : 'Verify code'}
              </Button>
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
          footer={
            <Button
              size="sm"
              variant="secondary"
              onClick={() => revokeAllSessions.mutate()}
              disabled={revokeAllSessions.isPending}
            >
              <LogOut size={12} />
              Sign out all sessions
            </Button>
          }
        >
          <div className="-mx-5">
            {data.sessions.map((s) => (
              <SessionRow key={s.id} session={s} />
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
