import React, { useState } from 'react';
import { motion } from 'motion/react';
import {
  Settings as SettingsIcon,
  Clock,
  Mail,
  Calendar,
  Loader2,
  Plus,
  Trash2,
} from 'lucide-react';
import {
  useInsightsSettings,
  useUpdateInsightsSettings,
  useScheduledReports,
  useCreateScheduledReport,
  useDeleteScheduledReport,
} from './useSettings';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';
import { toast } from 'sonner';

const ACCENT = '#6366f1';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const dateRangeOptions = [
  { label: 'Last 7 Days', value: '7d' },
  { label: 'Last 30 Days', value: '30d' },
  { label: 'Last 90 Days', value: '90d' },
  { label: 'Year to Date', value: 'ytd' },
  { label: 'Last 12 Months', value: '12m' },
];

const Settings: React.FC = () => {
  const { data: settings, isLoading: settingsLoading, isError, refetch } = useInsightsSettings();
  const updateSettings = useUpdateInsightsSettings();
  const { data: scheduled, isLoading: schedLoading } = useScheduledReports();
  const createScheduled = useCreateScheduledReport();
  const deleteScheduled = useDeleteScheduledReport();

  const [defaultRange, setDefaultRange] = useState<string>(settings?.defaultDateRange ?? '30d');
  const [emailRecipients, setEmailRecipients] = useState<string>(
    settings?.emailRecipients?.join(', ') ?? ''
  );

  if (isError) {
    return <ErrorState message="Failed to load settings." onRetry={refetch} />;
  }

  const scheduledList: any[] = scheduled?.data ?? scheduled ?? [];

  const handleSave = async () => {
    try {
      await updateSettings.mutateAsync({
        defaultDateRange: defaultRange,
        emailRecipients: emailRecipients
          .split(',')
          .map((e: string) => e.trim())
          .filter(Boolean),
      });
      toast.success('Settings saved');
    } catch {
      toast.error('Failed to save settings');
    }
  };

  const handleDeleteScheduled = async (id: string) => {
    try {
      await deleteScheduled.mutateAsync(id);
      toast.success('Scheduled report removed');
    } catch {
      toast.error('Failed to remove scheduled report');
    }
  };

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-3xl"
    >
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Settings
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Report scheduling, email recipients, and defaults
        </p>
      </motion.div>

      {/* Default Date Range */}
      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <Calendar size={14} style={{ color: ACCENT }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Default Date Range
          </h3>
        </div>
        {settingsLoading ? (
          <CardSkeleton />
        ) : (
          <div className="flex flex-wrap gap-2">
            {dateRangeOptions.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setDefaultRange(opt.value)}
                className="px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
                style={{
                  background: defaultRange === opt.value ? ACCENT : 'var(--hub-hover)',
                  color: defaultRange === opt.value ? '#fff' : 'var(--hub-text-muted)',
                }}
              >
                {opt.label}
              </button>
            ))}
          </div>
        )}
      </motion.div>

      {/* Email Recipients */}
      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <Mail size={14} style={{ color: ACCENT }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Email Recipients
          </h3>
        </div>
        {settingsLoading ? (
          <SkeletonRow cols={1} />
        ) : (
          <div>
            <input
              type="text"
              value={emailRecipients}
              onChange={(e) => setEmailRecipients(e.target.value)}
              placeholder="email1@example.com, email2@example.com"
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              Comma-separated list of email addresses
            </p>
          </div>
        )}
      </motion.div>

      {/* Save Button */}
      <motion.div variants={itemVariants as any}>
        <button
          onClick={handleSave}
          disabled={updateSettings.isPending}
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold transition-colors"
          style={{ background: ACCENT, color: '#fff' }}
        >
          {updateSettings.isPending && <Loader2 size={14} className="animate-spin" />}
          Save Settings
        </button>
      </motion.div>

      {/* Scheduled Reports */}
      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Clock size={14} style={{ color: ACCENT }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Scheduled Reports
            </h3>
          </div>
        </div>
        {schedLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : scheduledList.length === 0 ? (
          <EmptyState
            title="No scheduled reports"
            description="Scheduled reports will appear here once configured."
          />
        ) : (
          <div className="space-y-1">
            {scheduledList.map((s: any) => (
              <div
                key={s.id}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div>
                  <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {s.name ?? s.reportName ?? 'Untitled'}
                  </p>
                  <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                    {s.schedule ?? s.frequency ?? 'Manual'}
                  </p>
                </div>
                <button
                  onClick={() => handleDeleteScheduled(s.id)}
                  className="h-6 w-6 rounded flex items-center justify-center transition-colors"
                  style={{ color: 'var(--hub-text-dim)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.color = '#f87171';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.color = 'var(--hub-text-dim)';
                  }}
                >
                  <Trash2 size={12} />
                </button>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Settings;
