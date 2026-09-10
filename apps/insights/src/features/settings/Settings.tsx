import React, { useState, useEffect } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion, type Variants } from 'motion/react';
import { Clock, Loader2, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import {
  useInsightsSettings,
  useUpdateInsightsSettings,
  useScheduledReports,
  useDeleteScheduledReport,
} from '../../shared/hooks/queries/useInsightsSettings';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Settings: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: settings, isLoading: settingsLoading, isError, refetch } = useInsightsSettings();
  const updateSettings = useUpdateInsightsSettings();
  const {
    data: scheduledRes,
    isLoading: schedLoading,
    refetch: refetchScheduled,
  } = useScheduledReports();
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
    updateSettings.mutate({ defaultCurrency, fiscalYearStart, timezone });
  };

  const handleDeleteScheduled = (id: string): void => {
    deleteScheduled.mutate(id);
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-3xl"
    >
      <HelmetContainer title={t('insights.settings.metaTitle')} />
      <motion.div variants={itemVariants}>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.settings.title')}
        </h1>
      </motion.div>

      <motion.div
        variants={itemVariants}
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

      <motion.div variants={itemVariants}>
        <button
          type="button"
          onClick={handleSave}
          disabled={updateSettings.isPending || settingsLoading}
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold min-h-[44px] disabled:opacity-50"
          style={{ background: 'var(--accent)', color: 'var(--hub-on-accent)' }}
          aria-label={t('insights.settings.save')}
        >
          {updateSettings.isPending && <Loader2 size={14} className="animate-spin" />}
          {updateSettings.isPending ? t('insights.settings.saving') : t('insights.settings.save')}
        </button>
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Clock size={14} style={{ color: 'var(--accent)' }} />
            <h2 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('insights.settings.scheduledReports.title')}
            </h2>
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
            icon={Clock}
            title={t('insights.settings.scheduledReports.empty.title')}
            description={t('insights.settings.scheduledReports.empty.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchScheduled() }}
          />
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
                  className="min-h-[44px] min-w-[44px] rounded flex items-center justify-center"
                  style={{ color: 'var(--hub-text-dim)' }}
                  aria-label={t(
                    'insights.settings.scheduledReports.delete',
                    'Delete scheduled report'
                  )}
                >
                  <Trash2 size={14} />
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
