import React, { useState, useCallback } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import {
  FileText,
  Calendar,
  DollarSign,
  Clock,
  Users,
  TrendingDown,
  Timer,
  Shield,
  Building2,
  ClipboardList,
} from 'lucide-react';
import { useGenerateReport } from './useReports';
import ReportTable from './ReportTable';
import { REPORT_TYPES } from '../../shared/schemas/reports';
import type { ReportType, ReportPayload } from '../../shared/schemas/reports';
import { Button, Select, Label, Skeleton, useIsMobile } from '../../shared/components/ui';
import { toLocalIso, todayIso } from '../../shared/lib/dateUtils';

const REPORT_ICONS: Record<ReportType, React.ComponentType<{ size?: number }>> = {
  attendance_summary: Calendar,
  leave_balance: Clock,
  payroll_summary: DollarSign,
  overtime_report: Timer,
  headcount: Users,
  turnover: TrendingDown,
  timesheet_summary: ClipboardList,
  shift_coverage: Shield,
  department_cost: Building2,
  employee_directory: FileText,
};

function todayISO(): string {
  return todayIso();
}

function thirtyDaysAgoISO(): string {
  const d = new Date();
  d.setDate(d.getDate() - 30);
  return toLocalIso(d);
}

const Reports: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);

  const [selectedType, setSelectedType] = useState<ReportType>(REPORT_TYPES[0]);
  const [dateFrom, setDateFrom] = useState<string>(thirtyDaysAgoISO);
  const [dateTo, setDateTo] = useState<string>(todayISO);
  const [departmentId, setDepartmentId] = useState<string>('');
  const [locationId, setLocationId] = useState<string>('');
  // The payload and the filters it was generated with travel together: editing
  // the date inputs after Generate must not relabel the exported PDF.
  const [generated, setGenerated] = useState<{
    report: ReportPayload;
    dateFrom: string;
    dateTo: string;
  } | null>(null);

  const generateMutation = useGenerateReport();

  const handleGenerate = useCallback((): void => {
    generateMutation.mutate(
      {
        reportType: selectedType,
        dateFrom,
        dateTo,
        format: 'json',
        ...(departmentId ? { departmentId: Number(departmentId) } : {}),
        ...(locationId ? { locationId: Number(locationId) } : {}),
      },
      {
        onSuccess: (data) => {
          setGenerated({ report: data, dateFrom, dateTo });
          toast.success(t('reports.generated'));
        },
        onError: () => {
          toast.error(t('errors.generic'));
        },
      }
    );
  }, [selectedType, dateFrom, dateTo, departmentId, locationId, generateMutation, t]);

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.reports'))} />
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
        style={{ display: 'flex', flexDirection: 'column', gap: 24 }}
      >
        {/* Header */}
        <div>
          <h1 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hr-text)' }}>
            {t('reports.title')}
          </h1>
          <p style={{ margin: '4px 0 0', fontSize: 13, color: 'var(--hr-text-muted)' }}>
            {t('reports.subtitle')}
          </p>
        </div>

        {/* Main layout */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: isMobile ? '1fr' : '240px minmax(0, 1fr)',
            gap: 16,
            alignItems: 'start',
          }}
        >
          {/* Sidebar: report types */}
          <div
            style={{
              background: 'var(--hr-card)',
              border: '1px solid var(--hr-border)',
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                padding: '12px 16px',
                borderBottom: '1px solid var(--hr-border)',
              }}
            >
              <h2
                style={{
                  margin: 0,
                  fontSize: 12,
                  fontWeight: 600,
                  color: 'var(--hr-text-muted)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {t('reports.types')}
              </h2>
            </div>
            <nav style={{ display: 'flex', flexDirection: 'column', padding: 6, gap: 2 }}>
              {REPORT_TYPES.map((type) => {
                const Icon = REPORT_ICONS[type];
                const isActive = selectedType === type;
                return (
                  <button
                    type="button"
                    key={type}
                    onClick={() => {
                      setSelectedType(type);
                      setGenerated(null);
                    }}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: '8px 12px',
                      borderRadius: 8,
                      fontSize: 12,
                      fontWeight: 500,
                      textDecoration: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      textAlign: 'left',
                      background: isActive ? 'var(--hr-active)' : 'transparent',
                      color: isActive ? 'var(--hr-text)' : 'var(--hr-text-muted)',
                      transition: 'background 0.15s, color 0.15s',
                      width: '100%',
                    }}
                  >
                    <Icon size={14} />
                    <span>{t(`reports.reportTypes.${type}`)}</span>
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Main area */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {/* Filter panel */}
            <div
              style={{
                background: 'var(--hr-card)',
                border: '1px solid var(--hr-border)',
                borderRadius: 12,
                padding: 16,
              }}
            >
              <h3
                style={{
                  margin: '0 0 12px',
                  fontSize: 13,
                  fontWeight: 600,
                  color: 'var(--hr-text)',
                }}
              >
                {t('reports.filters')}
              </h3>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: isMobile ? '1fr' : 'repeat(auto-fit, minmax(160px, 1fr))',
                  gap: 12,
                  alignItems: 'end',
                }}
              >
                <div>
                  <Label htmlFor="report-from">{t('reports.dateFrom')}</Label>
                  <input
                    id="report-from"
                    type="date"
                    value={dateFrom}
                    onChange={(e) => setDateFrom(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 8,
                      border: '1px solid var(--hr-border)',
                      background: 'var(--hr-input-bg)',
                      color: 'var(--hr-text)',
                      fontSize: 13,
                      outline: 'none',
                    }}
                  />
                </div>
                <div>
                  <Label htmlFor="report-to">{t('reports.dateTo')}</Label>
                  <input
                    id="report-to"
                    type="date"
                    value={dateTo}
                    onChange={(e) => setDateTo(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 8,
                      border: '1px solid var(--hr-border)',
                      background: 'var(--hr-input-bg)',
                      color: 'var(--hr-text)',
                      fontSize: 13,
                      outline: 'none',
                    }}
                  />
                </div>
                <div>
                  <Label htmlFor="report-dept">{t('reports.department')}</Label>
                  <Select
                    id="report-dept"
                    value={departmentId}
                    onChange={(e) => setDepartmentId(e.target.value)}
                  >
                    <option value="">{t('reports.allDepartments')}</option>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="report-loc">{t('reports.location')}</Label>
                  <Select
                    id="report-loc"
                    value={locationId}
                    onChange={(e) => setLocationId(e.target.value)}
                  >
                    <option value="">{t('reports.allLocations')}</option>
                  </Select>
                </div>
                <div>
                  <Button
                    variant="primary"
                    onClick={handleGenerate}
                    disabled={generateMutation.isPending}
                    style={{ width: '100%', marginTop: 4 }}
                  >
                    {generateMutation.isPending ? t('common.loading') : t('reports.generate')}
                  </Button>
                </div>
              </div>
            </div>

            {/* Loading skeleton */}
            {generateMutation.isPending && (
              <div
                style={{
                  background: 'var(--hr-card)',
                  border: '1px solid var(--hr-border)',
                  borderRadius: 12,
                  padding: 20,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 12,
                }}
              >
                <Skeleton height={16} width="40%" />
                <Skeleton height={40} />
                <Skeleton height={36} />
                <Skeleton height={36} />
                <Skeleton height={36} />
              </div>
            )}

            {/* Results */}
            {generated && !generateMutation.isPending && (
              <ReportTable
                report={generated.report}
                dateFrom={generated.dateFrom}
                dateTo={generated.dateTo}
                onRefresh={handleGenerate}
              />
            )}

            {/* Empty state */}
            {!generated && !generateMutation.isPending && (
              <div
                style={{
                  padding: 48,
                  textAlign: 'center',
                  background: 'var(--hr-card)',
                  border: '1px dashed var(--hr-border)',
                  borderRadius: 12,
                }}
              >
                <div
                  style={{
                    display: 'inline-flex',
                    padding: 12,
                    borderRadius: 12,
                    background: 'var(--hr-hover)',
                    color: 'var(--hr-text-muted)',
                    marginBottom: 12,
                  }}
                >
                  <FileText size={24} />
                </div>
                <p style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hr-text)' }}>
                  {t('reports.emptyTitle')}
                </p>
                <p style={{ margin: '4px 0 0', fontSize: 12, color: 'var(--hr-text-muted)' }}>
                  {t('reports.emptyDescription')}
                </p>
              </div>
            )}
          </div>
        </div>
      </motion.div>
    </>
  );
};

export default Reports;
