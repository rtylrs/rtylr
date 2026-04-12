import React, { useState } from 'react';
import { motion } from 'motion/react';
import { FileText, Download, FileSpreadsheet, Loader2 } from 'lucide-react';
import { useReportTemplates, useExportReport } from './useReports';
import { SkeletonRow } from '../../shared/components/Skeleton';
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

const Reports: React.FC = () => {
  const { data: templates, isLoading, isError, refetch } = useReportTemplates();
  const exportMutation = useExportReport();
  const [exportingId, setExportingId] = useState<string | null>(null);

  if (isError) {
    return <ErrorState message="Failed to load report templates." onRetry={refetch} />;
  }

  type TemplateRow = {
    id?: string | number;
    templateId?: string;
    name?: string;
    title?: string;
    description?: string;
  };
  const templateList = (templates ?? []) as TemplateRow[];

  const handleExport = async (templateId: string, format: 'csv' | 'pdf') => {
    setExportingId(templateId);
    try {
      const blob = await exportMutation.mutateAsync({ templateId, format });
      const url = window.URL.createObjectURL(blob instanceof Blob ? blob : new Blob([blob]));
      const a = document.createElement('a');
      a.href = url;
      a.download = `report-${templateId}.${format}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      toast.success('Report exported successfully');
    } catch {
      toast.error('Failed to export report');
    } finally {
      setExportingId(null);
    }
  };

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Reports
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Pre-built report templates with CSV and PDF export
        </p>
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <FileText size={14} style={{ color: ACCENT }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Available Reports
          </h3>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : templateList.length === 0 ? (
          <EmptyState
            title="No report templates"
            description="Report templates will appear here once configured."
          />
        ) : (
          <div className="space-y-1">
            {templateList.map((t) => {
              const id = String(t.id ?? t.templateId ?? '');
              const isExporting = exportingId === id;
              return (
                <div
                  key={id}
                  className="flex items-center justify-between px-3 py-3 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {t.name ?? t.title ?? 'Untitled Report'}
                    </p>
                    {t.description && (
                      <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
                        {t.description}
                      </p>
                    )}
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleExport(id, 'csv')}
                      disabled={isExporting}
                      className="flex items-center gap-1 px-2 py-1 rounded-md text-[10px] font-medium transition-colors"
                      style={{ background: 'var(--hub-active)', color: 'var(--hub-text-muted)' }}
                    >
                      {isExporting ? (
                        <Loader2 size={10} className="animate-spin" />
                      ) : (
                        <FileSpreadsheet size={10} />
                      )}
                      CSV
                    </button>
                    <button
                      onClick={() => handleExport(id, 'pdf')}
                      disabled={isExporting}
                      className="flex items-center gap-1 px-2 py-1 rounded-md text-[10px] font-medium transition-colors"
                      style={{ background: 'var(--hub-active)', color: 'var(--hub-text-muted)' }}
                    >
                      {isExporting ? (
                        <Loader2 size={10} className="animate-spin" />
                      ) : (
                        <Download size={10} />
                      )}
                      PDF
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Reports;
