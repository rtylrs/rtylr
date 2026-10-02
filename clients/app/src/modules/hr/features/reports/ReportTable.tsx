import React, { useRef, useCallback, useState } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useTranslation } from 'react-i18next';
import { Download, FileDown, FileSearch } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Button } from '../../shared/components/ui';
import type { ReportPayload } from '../../shared/schemas/reports';
import { toLocalIso } from '../../shared/lib/dateUtils';

interface ReportTableProps {
  report: ReportPayload;
  dateFrom?: string;
  dateTo?: string;
  onRefresh: () => void;
}

/* ── CSV helpers ──────────────────────────────────────────────── */

function escapeCsvValue(val: unknown): string {
  const str = val == null ? '' : String(val);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

function downloadCsv(columns: string[], rows: Record<string, unknown>[], filename: string): void {
  const header = columns.map(escapeCsvValue).join(',');
  const body = rows
    .map((row) => columns.map((col) => escapeCsvValue(row[col])).join(','))
    .join('\n');
  const csv = `${header}\n${body}`;
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

/* ── PDF export ───────────────────────────────────────────────────
 * @react-pdf/renderer is ~1.5 MB. We load it lazily only when the
 * user clicks "Export PDF" so it is excluded from the initial chunk.
 * ─────────────────────────────────────────────────────────────── */

async function generatePdfBlob(
  report: ReportPayload,
  companyName: string,
  reportTypeLabel: string,
  dateRangeLabel: string,
  generatedLabel: string
): Promise<Blob> {
  // Dynamic import, only fetched when called, not bundled with ReportTable
  const { Document, Page, Text, View, StyleSheet, pdf } = await import(
    /* webpackChunkName: "react-pdf-renderer" */
    '@react-pdf/renderer'
  );

  const PDF_COLORS = {
    text: '#1a1a1a',
    muted: '#6b7280',
    faint: '#9ca3af',
    border: '#e5e7eb',
    rowBorder: '#f3f4f6',
    cell: '#374151',
  } as const;

  const styles = StyleSheet.create({
    page: { padding: 40, fontSize: 9, fontFamily: 'Helvetica', color: PDF_COLORS.text },
    header: { marginBottom: 20 },
    companyName: { fontSize: 16, fontFamily: 'Helvetica-Bold', marginBottom: 4 },
    reportTitle: {
      fontSize: 12,
      fontFamily: 'Helvetica-Bold',
      color: PDF_COLORS.muted,
      marginBottom: 2,
    },
    dateRange: { fontSize: 9, color: PDF_COLORS.faint, marginBottom: 2 },
    generatedAt: { fontSize: 8, color: PDF_COLORS.faint },
    table: { marginTop: 10 },
    headerRow: {
      flexDirection: 'row',
      borderBottomWidth: 1,
      borderBottomColor: PDF_COLORS.border,
      paddingBottom: 4,
      marginBottom: 4,
    },
    headerCell: {
      flex: 1,
      fontSize: 8,
      fontFamily: 'Helvetica-Bold',
      color: PDF_COLORS.muted,
      textTransform: 'uppercase',
    },
    row: {
      flexDirection: 'row',
      paddingVertical: 3,
      borderBottomWidth: 0.5,
      borderBottomColor: PDF_COLORS.rowBorder,
    },
    cell: { flex: 1, fontSize: 8, color: PDF_COLORS.cell },
    footer: { marginTop: 20, fontSize: 7, color: PDF_COLORS.faint, textAlign: 'center' },
  });

  const doc = (
    <Document>
      <Page size="A4" style={styles.page}>
        <View style={styles.header}>
          <Text style={styles.companyName}>{companyName}</Text>
          <Text style={styles.reportTitle}>{reportTypeLabel}</Text>
          <Text style={styles.dateRange}>{dateRangeLabel}</Text>
          <Text style={styles.generatedAt}>{generatedLabel}</Text>
        </View>
        <View style={styles.table}>
          <View style={styles.headerRow}>
            {report.columns.map((col) => (
              <Text key={col} style={styles.headerCell}>
                {col}
              </Text>
            ))}
          </View>
          {report.rows.map((row, i) => (
            <View key={i} style={styles.row}>
              {report.columns.map((col) => (
                <Text key={col} style={styles.cell}>
                  {row[col] == null ? '' : String(row[col])}
                </Text>
              ))}
            </View>
          ))}
        </View>
        <Text style={styles.footer}>{generatedLabel}</Text>
      </Page>
    </Document>
  );

  return pdf(doc).toBlob();
}

/* ── Table constants ──────────────────────────────────────────── */

const ROW_HEIGHT = 36;
const HEADER_HEIGHT = 40;

/* ── ReportTable ──────────────────────────────────────────────── */

const ReportTable = ({
  report,
  dateFrom,
  dateTo,
  onRefresh,
}: ReportTableProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const parentRef = useRef<HTMLDivElement>(null);
  const [pdfLoading, setPdfLoading] = useState(false);

  const rowVirtualizer = useVirtualizer({
    count: report.rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 20,
  });

  const handleExportCsv = useCallback((): void => {
    const filename = `${report.reportType}_${toLocalIso(new Date(report.generatedAt * 1000))}.csv`;
    downloadCsv(report.columns, report.rows, filename);
  }, [report]);

  const handleExportPdf = useCallback(async (): Promise<void> => {
    setPdfLoading(true);
    try {
      const reportTypeLabel = t(`hrReports.reportTypes.${report.reportType}`, report.reportType);
      const companyName = t('hrReports.pdfCompany');
      const dateRangeLabel =
        dateFrom && dateTo ? t('hrReports.pdfDateRange', { from: dateFrom, to: dateTo }) : '';
      const generatedLabel = t('hrReports.pdfGenerated', {
        date: new Date().toLocaleDateString(),
      });

      // @react-pdf/renderer is loaded lazily here, not bundled with the initial chunk
      const blob = await generatePdfBlob(
        report,
        companyName,
        reportTypeLabel,
        dateRangeLabel,
        generatedLabel
      );

      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${report.reportType}_${toLocalIso(new Date(report.generatedAt * 1000))}.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setPdfLoading(false);
    }
  }, [report, dateFrom, dateTo, t]);

  if (report.rows.length === 0) {
    return (
      <EmptyState
        icon={FileSearch}
        title={t('hrReports.noResults')}
        action={{ label: t('hrCommon.refresh'), onClick: onRefresh }}
      />
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {/* Actions bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
          {t('hrReports.rowCount', { count: report.rows.length })}
        </span>
        <div style={{ display: 'flex', gap: 8 }}>
          <Button variant="secondary" icon={<Download size={12} />} onClick={handleExportCsv}>
            {t('hrReports.exportCsv')}
          </Button>
          <Button
            variant="secondary"
            icon={<FileDown size={12} />}
            onClick={handleExportPdf}
            disabled={pdfLoading}
          >
            {pdfLoading ? t('hrCommon.loading') : t('hrReports.exportPdf')}
          </Button>
        </div>
      </div>

      {/* Table */}
      <div
        style={{
          background: 'var(--hr-card)',
          border: '1px solid var(--hr-border)',
          borderRadius: 12,
          overflow: 'hidden',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            height: HEADER_HEIGHT,
            alignItems: 'center',
            borderBottom: '1px solid var(--hr-border)',
            background: 'var(--hr-hover)',
            paddingLeft: 16,
            paddingRight: 16,
          }}
        >
          {report.columns.map((col) => (
            <div
              key={col}
              style={{
                flex: 1,
                minWidth: 100,
                fontSize: 11,
                fontWeight: 600,
                color: 'var(--hr-text-muted)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
                paddingRight: 8,
              }}
            >
              {col}
            </div>
          ))}
        </div>

        {/* Virtualised rows */}
        <div
          ref={parentRef}
          style={{
            height: Math.min(report.rows.length * ROW_HEIGHT, 480),
            overflow: 'auto',
          }}
        >
          <div
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`,
              width: '100%',
              position: 'relative',
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = report.rows[virtualRow.index];
              return (
                <div
                  key={virtualRow.index}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: `${virtualRow.size}px`,
                    transform: `translateY(${virtualRow.start}px)`,
                    display: 'flex',
                    alignItems: 'center',
                    paddingLeft: 16,
                    paddingRight: 16,
                    borderBottom: '1px solid var(--hr-border)',
                    fontSize: 12,
                    color: 'var(--hr-text)',
                  }}
                >
                  {report.columns.map((col) => (
                    <div
                      key={col}
                      style={{
                        flex: 1,
                        minWidth: 100,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        paddingRight: 8,
                      }}
                    >
                      {row[col] == null ? '' : String(row[col])}
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ReportTable;
