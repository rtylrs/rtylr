import React, { useState, useEffect } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { Clock, Loader2, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useInsightsSettings,
  useUpdateInsightsSettings,
  useScheduledReports,
  useDeleteScheduledReport,
} from '../../shared/hooks/queries/useInsightsSettings';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Settings: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: settings, isLoading: settingsLoading, isError, refetch } = useInsightsSettings();
  const updateSettings = useUpdateInsightsSettings();
  const { data: scheduledRes, isLoading: schedLoading } = useScheduledReports();
  const deleteScheduled = useDeleteScheduledReport();

  const [defaultCurrency, setDefaultCurrency] = useState<string>('');
  const [fiscalYearStart, setFiscalYearStart] = useState<string>('');
  const [timezone, setTimezone] = useState<string>('');

  useEffect(() => {
    if (settings) {
      setDefaultCurrency(settings.defaultCurrency);
      setFiscalYearStart(settings.fiscalYearStart);
      setTimezone(settings.timezone);
    }
  }, [settings]);

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const scheduledList = scheduledRes?.scheduledReports ?? [];

  const handleSave = (): void => {
    updateSettings.mutate(
      { defaultCurrency, fiscalYearStart, timezone },
      {
        onSuccess: () => toast.success('Settings saved'),
        onError: () => toast.error('Failed to save settings'),
      }
    );
  };

  const handleDeleteScheduled = (id: string): void => {
    deleteScheduled.mutate(id, {
      onSuccess: () => toast.success('Scheduled report deleted'),
      onError: () => toast.error('Failed to delete scheduled report'),
    });
  };

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-3xl"
    >
      <HelmetContainer title="Settings · rtylr" />
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.settings.title')}
        </h2>
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5 space-y-4"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        {settingsLoading ? (
          <CardSkeleton />
        ) : (
          <>
            <div>
              <label
                htmlFor="insights-default-currency"
                className="text-xs font-medium"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                {t('insights.settings.defaultCurrency')}
              </label>
              <input
                id="insights-default-currency"
                type="text"
                value={defaultCurrency}
                onChange={(e) => setDefaultCurrency(e.target.value)}
                className="mt-1 w-full px-3 py-2 rounded-lg text-base outline-none"
                style={{
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                  fontSize: '16px',
                }}
              />
            </div>
            <div>
              <label
                htmlFor="insights-fiscal-year"
                className="text-xs font-medium"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                {t('insights.settings.fiscalYearStart')}
              </label>
              <input
                id="insights-fiscal-year"
                type="text"
                value={fiscalYearStart}
                onChange={(e) => setFiscalYearStart(e.target.value)}
                placeholder="MM-DD"
                className="mt-1 w-full px-3 py-2 rounded-lg outline-none"
                style={{
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                  fontSize: '16px',
                }}
              />
            </div>
            <div>
              <label
                htmlFor="insights-timezone"
                className="text-xs font-medium"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                {t('insights.settings.timezone')}
              </label>
              <input
                id="insights-timezone"
                type="text"
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                className="mt-1 w-full px-3 py-2 rounded-lg outline-none"
                style={{
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                  fontSize: '16px',
                }}
              />
            </div>
          </>
        )}
      </motion.div>

      <motion.div variants={itemVariants as any}>
        <button
          type="button"
          onClick={handleSave}
          disabled={updateSettings.isPending}
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold min-h-[44px]"
          style={{ background: 'var(--accent)', color: 'var(--accent-fg)' }}
          aria-label={t('insights.settings.saved')}
        >
          {updateSettings.isPending && <Loader2 size={14} className="animate-spin" />}
          {t('insights.settings.saved')}
        </button>
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Clock size={14} style={{ color: 'var(--accent)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('insights.settings.scheduledReports.title')}
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
          <EmptyState title="" description={t('insights.settings.scheduledReports.empty')} />
        ) : (
          <div className="space-y-1">
            {scheduledList.map((s) => (
              <div
                key={s.id}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div>
                  <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {s.templateId}
                  </p>
                  <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                    {s.cadence} · {s.channel} · {s.recipient}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => handleDeleteScheduled(s.id)}
                  className="h-11 w-11 rounded flex items-center justify-center"
                  style={{ color: 'var(--hub-text-dim)' }}
                  aria-label={t('insights.settings.scheduledReports.delete')}
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
