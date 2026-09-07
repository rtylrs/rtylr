import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion, type Variants } from 'motion/react';
import { FileText, Download, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useUserStore } from '@voxire/auth';
import { EmptyState } from '@voxire/ui';
import {
  useReportTemplates,
  useGenerateRevenueReport,
  useExportRevenueReport,
} from '../../shared/hooks/queries/useInsightsReports';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import { canInsightsPermission } from '../../shared/lib/insightsPermissions';

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Reports: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, isError, refetch } = useReportTemplates();
  const generateRevenue = useGenerateRevenueReport();
  const exportRevenue = useExportRevenueReport();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canExportReports = canInsightsPermission(role, 'insights:reports:export');

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const templates = data?.templates ?? [];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Reports · rtylr" />
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.reports.title')}
        </h1>
        <div className="flex items-center gap-2">
          {canExportReports && (
            <button
              type="button"
              onClick={() => exportRevenue.mutate({})}
              disabled={exportRevenue.isPending}
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium min-h-[44px]"
              style={{ background: 'var(--hub-active)', color: 'var(--hub-text)' }}
              aria-label={t('insights.reports.exportRevenue')}
            >
              {exportRevenue.isPending ? (
                <Loader2 size={14} className="animate-spin" />
              ) : (
                <Download size={14} />
              )}
              {t('insights.reports.exportRevenue')}
            </button>
          )}
          <button
            type="button"
            onClick={() => generateRevenue.mutate({})}
            disabled={generateRevenue.isPending}
            className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium min-h-[44px]"
            style={{ background: 'var(--accent)', color: 'var(--hub-on-accent)' }}
            aria-label={t('insights.reports.generateRevenue')}
          >
            {generateRevenue.isPending ? (
              <Loader2 size={14} className="animate-spin" />
            ) : (
              <FileText size={14} />
            )}
            {t('insights.reports.generateRevenue')}
          </button>
        </div>
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <FileText size={14} style={{ color: 'var(--accent)' }} />
          <h2 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.reports.templates')}
          </h2>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : templates.length === 0 ? (
          <EmptyState
            icon={FileText}
            title={t('insights.reports.empty.title')}
            description={t('insights.reports.empty.description')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <div className="space-y-1">
            {templates.map((tpl) => (
              <div
                key={tpl.id}
                className="flex items-center justify-between px-3 py-3 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div>
                  <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {tpl.name}
                  </p>
                  {tpl.description && (
                    <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
                      {tpl.description}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Reports;
