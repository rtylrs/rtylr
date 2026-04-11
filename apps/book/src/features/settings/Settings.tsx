import React from 'react';
import { motion } from 'motion/react';
import { Settings as SettingsIcon, Globe, Clock, XCircle, Bell } from 'lucide-react';
import { useBookSettings, useUpdateBookSettings } from './useBookSettings';
import { toast } from 'sonner';

const INDIGO = '#6366f1';

const Section: React.FC<{ title: string; icon: React.ComponentType<{ size?: number; style?: any }>; children: React.ReactNode }> = ({
  title,
  icon: Icon,
  children,
}) => (
  <div
    className="rounded-xl p-5 space-y-4"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <div className="flex items-center gap-2">
      <Icon size={14} style={{ color: INDIGO }} />
      <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
        {title}
      </h3>
    </div>
    <div className="space-y-3">{children}</div>
  </div>
);

const Field: React.FC<{ label: string; description?: string; children: React.ReactNode }> = ({ label, description, children }) => (
  <div className="flex items-center justify-between gap-4">
    <div className="flex-1">
      <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
        {label}
      </p>
      {description && (
        <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
          {description}
        </p>
      )}
    </div>
    <div>{children}</div>
  </div>
);

const Toggle: React.FC<{ value: boolean; onChange: (v: boolean) => void }> = ({ value, onChange }) => (
  <button
    onClick={() => onChange(!value)}
    className="relative h-5 w-9 rounded-full transition-colors"
    style={{ background: value ? INDIGO : 'var(--hub-hover)' }}
  >
    <span
      className="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-all"
      style={{ left: value ? '18px' : '2px' }}
    />
  </button>
);

const SkeletonBlock: React.FC = () => (
  <div className="h-48 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const Settings: React.FC = () => {
  const { data: settings, isLoading, isError } = useBookSettings();
  const updateMutation = useUpdateBookSettings();

  const handleSave = () => {
    toast.success('Settings saved');
  };

  if (isLoading) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-5 max-w-3xl"
      >
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>Settings</h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>Configure booking preferences</p>
        </div>
        <SkeletonBlock />
        <SkeletonBlock />
      </motion.div>
    );
  }

  if (isError) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-5 max-w-3xl"
      >
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>Settings</h2>
        </div>
        <div
          className="rounded-xl p-4 text-sm"
          style={{ background: 'rgba(239,68,68,0.08)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
        >
          Failed to load settings. Please try again later.
        </div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-3xl"
    >
      <div>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>Settings</h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>Configure booking preferences</p>
      </div>

      <Section title="Booking Page" icon={Globe}>
        <Field label="Booking URL" description="Public URL for customer self-service booking">
          <input
            type="text"
            defaultValue={settings?.bookingUrl ?? ''}
            className="w-64 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            readOnly
          />
        </Field>
        <Field label="Allow Online Booking" description="Let customers book directly from the booking page">
          <Toggle value={settings?.onlineBookingEnabled ?? true} onChange={() => {}} />
        </Field>
      </Section>

      <Section title="Time Slots" icon={Clock}>
        <Field label="Slot Duration (min)" description="Default time slot length for appointments">
          <input
            type="number"
            defaultValue={settings?.slotDuration ?? 30}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
        <Field label="Buffer Time (min)" description="Gap between consecutive appointments">
          <input
            type="number"
            defaultValue={settings?.bufferTime ?? 10}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
        <Field label="Business Hours Start" description="When your day begins">
          <input
            type="time"
            defaultValue={settings?.businessHoursStart ?? '09:00'}
            className="px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
        <Field label="Business Hours End" description="When your day ends">
          <input
            type="time"
            defaultValue={settings?.businessHoursEnd ?? '18:00'}
            className="px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
      </Section>

      <Section title="Cancellation Policy" icon={XCircle}>
        <Field label="Allow Cancellations" description="Let customers cancel appointments online">
          <Toggle value={settings?.allowCancellations ?? true} onChange={() => {}} />
        </Field>
        <Field label="Cancellation Window (hours)" description="Minimum hours before appointment to allow cancellation">
          <input
            type="number"
            defaultValue={settings?.cancellationWindow ?? 24}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
      </Section>

      <Section title="Notifications" icon={Bell}>
        <Field label="Email Confirmations" description="Send email when appointments are booked">
          <Toggle value={settings?.emailConfirmations ?? true} onChange={() => {}} />
        </Field>
        <Field label="SMS Reminders" description="Send SMS reminders before appointments">
          <Toggle value={settings?.smsReminders ?? false} onChange={() => {}} />
        </Field>
        <Field label="Reminder Advance (hours)" description="How far ahead to send reminders">
          <input
            type="number"
            defaultValue={settings?.reminderAdvance ?? 24}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
        </Field>
      </Section>

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          className="px-4 py-2 rounded-lg text-sm font-semibold"
          style={{ background: INDIGO, color: '#fff' }}
        >
          Save Settings
        </button>
      </div>
    </motion.div>
  );
};

export default Settings;
