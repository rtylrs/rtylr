import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Mail, Server, Bell } from 'lucide-react';
import { toast } from 'sonner';
import { useMutation } from '@tanstack/react-query';
import { unwrap, getOneKitErrorMessage } from '@voxire/api';
import { rest } from '../../../../../shared/lib/clients';
import { useUpdateBusiness } from '../../../../../shared/hooks/queries/useBusiness';
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
}

const ToggleSwitch: React.FC<ToggleSwitchProps> = ({ checked, onChange, id, label }) => (
  <button
    id={id}
    type="button"
    role="switch"
    aria-checked={checked}
    aria-label={label || 'Toggle switch'}
    onClick={() => onChange(!checked)}
    className="relative shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#4583C4]/50 rounded-full min-h-[44px] min-w-[44px] inline-flex items-center justify-center p-1"
  >
    <span
      className="inline-block rounded-full"
      style={{
        width: 36,
        height: 20,
        background: checked ? '#4583C4' : 'var(--dash-border)',
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
            background: active ? '#4583C4' : 'transparent',
            color: active ? '#fff' : 'var(--dash-text-muted)',
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
}

const NotificationRow: React.FC<NotificationRowProps> = ({
  id,
  label,
  description,
  checked,
  onChange,
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
    <ToggleSwitch id={id} checked={checked} onChange={onChange} />
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

interface NotificationState {
  newOrder: boolean;
  statusUpdates: boolean;
  weeklySummary: boolean;
  feedback: boolean;
}

const Email: React.FC = () => {
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

  // Notifications
  const [notifications, setNotifications] = useState<NotificationState>({
    newOrder: true,
    statusUpdates: true,
    weeklySummary: false,
    feedback: false,
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
      toast.error('Sender email is required');
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
      toast.success(`SMTP connection to ${smtp.host}:${smtp.port || 587} verified successfully`);
    },
    onError: (err) => {
      toast.error(getOneKitErrorMessage(err, 'Failed to test SMTP connection'));
    },
  });

  const handleTestSmtp = (): void => {
    if (!smtp.host || !smtp.username) {
      toast.error('SMTP host and user are required');
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
      toast.error('SMTP host is required');
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

  const toggleNotification = (key: keyof NotificationState) => (next: boolean) => {
    const nextState = { ...notifications, [key]: next };
    setNotifications(nextState);
    updateBusiness.mutate({
      data: { notificationPreferences: nextState },
    });
  };

  return (
    <>
      <HelmetContainer title="Email settings · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Email Settings
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Configure how emails are sent to your customers.
          </p>
        </div>

        {/* Section 1, Sender Identity */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Mail size={15} style={{ color: 'var(--dash-text-dim)' }} />
              <CardTitle>Sender Identity</CardTitle>
            </div>
            <CardDescription>Control how your business appears in outgoing emails.</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="senderName">Sender Name</Label>
              <Input
                id="senderName"
                name="senderName"
                placeholder="The Golden Fork"
                value={sender.senderName}
                onChange={handleSenderChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="fromEmail">From Email</Label>
              <Input
                id="fromEmail"
                name="fromEmail"
                type="email"
                placeholder="hello@mybusiness.com"
                value={sender.fromEmail}
                onChange={handleSenderChange}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="replyTo">Reply-to Email</Label>
              <Input
                id="replyTo"
                name="replyTo"
                type="email"
                placeholder="support@mybusiness.com"
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
              <CardTitle>SMTP Configuration</CardTitle>
            </div>
            <CardDescription>
              Optional: use your own SMTP server for full control over email delivery.
            </CardDescription>
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
                    <Label htmlFor="smtpHost">SMTP Host</Label>
                    <Input
                      id="smtpHost"
                      name="host"
                      placeholder="smtp.example.com"
                      value={smtp.host}
                      onChange={handleSmtpChange}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="smtpPort">Port</Label>
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
                  <Label htmlFor="smtpUsername">Username</Label>
                  <Input
                    id="smtpUsername"
                    name="username"
                    placeholder="smtp-user@example.com"
                    value={smtp.username}
                    onChange={handleSmtpChange}
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="smtpPassword">Password</Label>
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
                  <Label>Encryption</Label>
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
              <CardTitle>Email Notifications</CardTitle>
            </div>
            <CardDescription>Choose which events trigger an email notification.</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 divide-y" style={{ borderColor: 'var(--dash-border)' }}>
            <NotificationRow
              id="notif-new-order"
              label="New order received"
              description="Get an email each time a customer places an order."
              checked={notifications.newOrder}
              onChange={toggleNotification('newOrder')}
            />
            <NotificationRow
              id="notif-status"
              label="Order status updates"
              description="Notify customers when their order status changes."
              checked={notifications.statusUpdates}
              onChange={toggleNotification('statusUpdates')}
            />
            <NotificationRow
              id="notif-weekly"
              label="Weekly summary"
              description="A weekly digest of orders, revenue, and trends."
              checked={notifications.weeklySummary}
              onChange={toggleNotification('weeklySummary')}
            />
            <NotificationRow
              id="notif-feedback"
              label="Customer feedback"
              description="Receive an email when a customer submits a review."
              checked={notifications.feedback}
              onChange={toggleNotification('feedback')}
            />
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Email;
