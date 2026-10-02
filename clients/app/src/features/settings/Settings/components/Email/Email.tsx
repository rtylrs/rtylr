import { PageHeader } from '../../../../../shared/components/dash/PageHeader';
import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Mail, Server, Bell } from 'lucide-react';
import { toast } from 'sonner';
import { useMutation } from '@tanstack/react-query';
import { unwrap, getOneKitErrorMessage } from '@voxire/api';
import { rest } from '../../../../../shared/lib/clients';
import { cn } from '../../../../../shared/lib/utils';
import { useUpdateBusiness } from '../../../../../shared/hooks/queries/useBusiness';
import UnavailableNotice from '../../../../../shared/components/dash/UnavailableNotice';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
  Separator,
} from '../../../../../shared/components/ui';

// ─── Toggle switch ────────────────────────────────────────────────────────────
interface ToggleSwitchProps {
  checked: boolean;
  onChange: (next: boolean) => void;
  id?: string;
  label?: string;
  disabled?: boolean;
}

const ToggleSwitch: React.FC<ToggleSwitchProps> = ({ checked, onChange, id, label, disabled }) => (
  <button
    id={id}
    type="button"
    role="switch"
    aria-checked={checked}
    aria-label={label || 'Toggle switch'}
    disabled={disabled}
    onClick={() => onChange(!checked)}
    className={cn(
      'relative shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--dash-text)]/50 rounded-full min-h-[44px] min-w-[44px] inline-flex items-center justify-center p-1',
      disabled && 'opacity-50 cursor-not-allowed'
    )}
  >
    <span
      className="inline-block rounded-full"
      style={{
        width: 36,
        height: 20,
        background: checked ? 'var(--dash-text)' : 'var(--dash-border)',
        transition: 'background 0.2s',
        position: 'relative',
      }}
    >
      <span
        className="absolute top-0.5 rounded-full bg-white transition-all"
        style={{
          left: checked ? 18 : 2,
          width: 16,
          height: 16,
          boxShadow: '0 1px 3px rgba(0,0,0,0.25)',
        }}
      />
    </span>
  </button>
);

// ─── Segment button group (for encryption selector) ──────────────────────────
type Encryption = 'None' | 'TLS' | 'SSL';

interface SegmentGroupProps {
  options: Encryption[];
  value: Encryption;
  onChange: (v: Encryption) => void;
}

const SegmentGroup: React.FC<SegmentGroupProps> = ({ options, value, onChange }) => (
  <div
    className="flex rounded-lg overflow-hidden"
    style={{ border: '1px solid var(--dash-border)', background: 'var(--dash-input-bg)' }}
  >
    {options.map((opt) => {
      const active = opt === value;
      return (
        <button
          key={opt}
          type="button"
          onClick={() => onChange(opt)}
          className="flex-1 py-1.5 text-xs font-medium transition-colors"
          style={{
            background: active ? 'var(--dash-active)' : 'transparent',
            color: active ? 'var(--dash-text)' : 'var(--dash-text-muted)',
          }}
        >
          {opt}
        </button>
      );
    })}
  </div>
);

// ─── Notification toggle row ──────────────────────────────────────────────────
interface NotificationRowProps {
  id: string;
  label: string;
  description: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
}

const NotificationRow: React.FC<NotificationRowProps> = ({
  id,
  label,
  description,
  checked,
  onChange,
  disabled,
}) => (
  <div className="flex items-center justify-between gap-4 py-3">
    <div className="flex-1 min-w-0">
      <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
        {label}
      </p>
      <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
        {description}
      </p>
    </div>
    <ToggleSwitch id={id} checked={checked} onChange={onChange} disabled={disabled} />
  </div>
);

// ─── Main component ───────────────────────────────────────────────────────────
const ENCRYPTION_OPTIONS: Encryption[] = ['None', 'TLS', 'SSL'];

interface SenderForm {
  senderName: string;
  fromEmail: string;
  replyTo: string;
}

interface SmtpForm {
  host: string;
  port: string;
  username: string;
  password: string;
  encryption: Encryption;
}

/** Mirrors the UI's intended defaults while the prefs cannot be persisted. */
const NOTIFICATION_ROWS = [
  {
    id: 'notif-new-order',
    labelKey: 'settings.email.newOrder',
    descriptionKey: 'settings.email.newOrderDescription',
    checked: true,
  },
  {
    id: 'notif-status',
    labelKey: 'settings.email.orderStatus',
    descriptionKey: 'settings.email.orderStatusDescription',
    checked: true,
  },
  {
    id: 'notif-weekly',
    labelKey: 'settings.email.weeklySummary',
    descriptionKey: 'settings.email.weeklySummaryDescription',
    checked: false,
  },
  {
    id: 'notif-feedback',
    labelKey: 'settings.email.feedback',
    descriptionKey: 'settings.email.feedbackDescription',
    checked: false,
  },
] as const;

const Email: React.FC = () => {
  const { t } = useTranslation();
  // Sender identity
  const [sender, setSender] = useState<SenderForm>({
    senderName: '',
    fromEmail: '',
    replyTo: '',
  });
  // SMTP
  const [useSmtp, setUseSmtp] = useState(false);
  const [smtp, setSmtp] = useState<SmtpForm>({
    host: '',
    port: '587',
    username: '',
    password: '',
    encryption: 'TLS',
  });

  const handleSenderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setSender((prev) => ({ ...prev, [name]: value }));
  };

  const handleSmtpChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setSmtp((prev) => ({ ...prev, [name]: value }));
  };

  const updateBusiness = useUpdateBusiness();

  const handleSaveSender = (): void => {
    if (!sender.fromEmail) {
      toast.error(t('settings.email.senderRequired'));
      return;
    }
    updateBusiness.mutate({
      data: {
        senderName: sender.senderName,
        senderEmail: sender.fromEmail,
        replyTo: sender.replyTo,
      },
    });
  };

  const testSmtpMutation = useMutation({
    mutationFn: (data: { host: string; port: number; username: string }) =>
      // hub/settings/compat/v1 has no contract in voxie (#894).
      unwrap(rest.post('/hub/settings/email/test', data)),
    onSuccess: () => {
      toast.success(t('settings.email.smtpVerified', { host: smtp.host, port: smtp.port || 587 }));
    },
    onError: (err) => {
      toast.error(getOneKitErrorMessage(err, t('settings.email.smtpTestFailed')));
    },
  });

  const handleTestSmtp = (): void => {
    if (!smtp.host || !smtp.username) {
      toast.error(t('settings.email.smtpHostUserRequired'));
      return;
    }
    testSmtpMutation.mutate({
      host: smtp.host,
      port: Number(smtp.port) || 587,
      username: smtp.username,
    });
  };

  const handleSaveSmtp = (): void => {
    if (!smtp.host) {
      toast.error(t('settings.email.smtpHostRequired'));
      return;
    }
    updateBusiness.mutate({
      data: {
        smtpHost: smtp.host,
        smtpPort: Number(smtp.port) || 587,
        smtpUser: smtp.username,
        smtpPassword: smtp.password,
        smtpFrom: sender.fromEmail,
      },
    });
  };

  return (
    <>
      <HelmetContainer title={t('settings.email.pageTitle')} />
      <div className="portal-management-page space-y-6">
        <PageHeader
          title={t('common.nav.emailSettings')}
          description={t('common.management.emailDescription')}
        />

        {/* Section 1, Sender Identity */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Mail size={15} style={{ color: 'var(--dash-text-dim)' }} />
              <CardTitle>{t('settings.email.senderTitle')}</CardTitle>
            </div>
            <CardDescription>{t('settings.email.senderDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="senderName">{t('settings.email.senderName')}</Label>
              <Input
                id="senderName"
                name="senderName"
                placeholder={t('settings.email.senderNamePlaceholder')}
                value={sender.senderName}
                onChange={handleSenderChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="fromEmail">{t('settings.email.fromEmail')}</Label>
              <Input
                id="fromEmail"
                name="fromEmail"
                type="email"
                placeholder={t('settings.email.fromEmailPlaceholder')}
                value={sender.fromEmail}
                onChange={handleSenderChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="replyTo">{t('settings.email.replyTo')}</Label>
              <Input
                id="replyTo"
                name="replyTo"
                type="email"
                placeholder={t('settings.email.replyToPlaceholder')}
                value={sender.replyTo}
                onChange={handleSenderChange}
              />
            </div>
            <div className="flex justify-end pt-1">
              <Button size="sm" onClick={handleSaveSender}>
                Save changes
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Section 2, SMTP Configuration */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Server size={15} style={{ color: 'var(--dash-text-dim)' }} />
              <CardTitle>{t('settings.email.smtpTitle')}</CardTitle>
            </div>
            <CardDescription>{t('settings.email.smtpDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            {/* Toggle row */}
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                  Use custom SMTP server
                </p>
                <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  Send emails through your own mail server
                </p>
              </div>
              <ToggleSwitch id="smtp-toggle" checked={useSmtp} onChange={setUseSmtp} />
            </div>

            {useSmtp && (
              <>
                <Separator />

                <div className="grid grid-cols-3 gap-3">
                  {/* SMTP Host spans 2 cols */}
                  <div className="col-span-2 space-y-1.5">
                    <Label htmlFor="smtpHost">{t('settings.email.smtpHost')}</Label>
                    <Input
                      id="smtpHost"
                      name="host"
                      placeholder={t('settings.email.smtpHostPlaceholder')}
                      value={smtp.host}
                      onChange={handleSmtpChange}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="smtpPort">{t('settings.email.port')}</Label>
                    <Input
                      id="smtpPort"
                      name="port"
                      type="number"
                      placeholder="587"
                      value={smtp.port}
                      onChange={handleSmtpChange}
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="smtpUsername">{t('settings.email.username')}</Label>
                  <Input
                    id="smtpUsername"
                    name="username"
                    placeholder={t('settings.email.usernamePlaceholder')}
                    value={smtp.username}
                    onChange={handleSmtpChange}
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="smtpPassword">{t('settings.email.password')}</Label>
                  <Input
                    id="smtpPassword"
                    name="password"
                    type="password"
                    placeholder="••••••••"
                    value={smtp.password}
                    onChange={handleSmtpChange}
                  />
                </div>

                <div className="space-y-1.5">
                  <Label>{t('settings.email.encryption')}</Label>
                  <SegmentGroup
                    options={ENCRYPTION_OPTIONS}
                    value={smtp.encryption}
                    onChange={(v) => setSmtp((prev) => ({ ...prev, encryption: v }))}
                  />
                </div>

                <div className="flex items-center gap-2 justify-end pt-1">
                  <Button variant="outline" size="sm" onClick={handleTestSmtp}>
                    Test Connection
                  </Button>
                  <Button size="sm" onClick={handleSaveSmtp}>
                    Save changes
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Section 3, Email Notifications */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bell size={15} style={{ color: 'var(--dash-text-dim)' }} />
              <CardTitle>{t('settings.email.notificationsTitle')}</CardTitle>
            </div>
            <CardDescription>{t('settings.email.notificationsDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-3">
            {/* The config contract has no per-event preference fields, so a
                toggle here cannot persist anything. Saying so is honest until
                a notification-preferences endpoint exists. */}
            <UnavailableNotice>{t('settings.email.notificationsUnavailable')}</UnavailableNotice>
            <div className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
              {NOTIFICATION_ROWS.map((row) => (
                <NotificationRow
                  key={row.id}
                  id={row.id}
                  label={t(row.labelKey)}
                  description={t(row.descriptionKey)}
                  checked={row.checked}
                  onChange={() => undefined}
                  disabled
                />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Email;
