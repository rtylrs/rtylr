import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { FileText, Download, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useReportTemplates,
  useGenerateRevenueReport,
  useRevenueCSVExportURL,
} from '../../shared/hooks/queries/useInsightsReports';
import { SkeletonRow } from '../../shared/components/Skeleton';
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

const Reports: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, isError, refetch } = useReportTemplates();
  const generateRevenue = useGenerateRevenueReport();
  const csvExportUrl = useRevenueCSVExportURL();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const templates = data?.templates ?? [];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Reports · rtylr" />
      <motion.div variants={itemVariants as any} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.reports.title')}
        </h2>
        <div className="flex items-center gap-2">
          <a
            href={csvExportUrl}
            download
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1 px-3 py-1.5 rounded-md text-xs font-medium"
            style={{ background: 'var(--hub-active)', color: 'var(--hub-text)' }}
            aria-label={t('insights.reports.exportRevenue')}
          >
            <Download size={12} />
            {t('insights.reports.exportRevenue')}
          </a>
          <button
            type="button"
            onClick={() =>
              generateRevenue.mutate(
                {},
                {
                  onSuccess: () => toast.success('Report generated'),
                  onError: () => toast.error('Failed to generate report'),
                }
              )
            }
            disabled={generateRevenue.isPending}
            className="flex items-center gap-1 px-3 py-1.5 rounded-md text-xs font-medium"
            style={{ background: 'var(--accent)', color: 'var(--accent-fg)' }}
            aria-label={t('insights.reports.generateRevenue')}
          >
            {generateRevenue.isPending ? (
              <Loader2 size={12} className="animate-spin" />
            ) : (
              <FileText size={12} />
            )}
            {t('insights.reports.generateRevenue')}
          </button>
        </div>
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <FileText size={14} style={{ color: 'var(--accent)' }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.reports.templates')}
          </h3>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : templates.length === 0 ? (
          <EmptyState title="" description={t('insights.reports.empty')} />
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
