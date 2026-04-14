import React, { useRef, useCallback } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useTranslation } from 'react-i18next';
import { Download } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import type { ReportPayload } from '../../shared/schemas/reports';

interface ReportTableProps {
  report: ReportPayload;
}

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

const ROW_HEIGHT = 36;
const HEADER_HEIGHT = 40;

const ReportTable = ({ report }: ReportTableProps): React.JSX.Element => {
  const { t } = useTranslation();
  const parentRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: report.rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 20,
  });

  const handleExportCsv = useCallback((): void => {
    const filename = `${report.reportType}_${new Date(report.generatedAt * 1000).toISOString().slice(0, 10)}.csv`;
    downloadCsv(report.columns, report.rows, filename);
  }, [report]);

  if (report.rows.length === 0) {
    return (
      <div
        style={{
          padding: 32,
          textAlign: 'center',
          background: 'var(--hr-card)',
          border: '1px dashed var(--hr-border)',
          borderRadius: 12,
        }}
      >
        <p style={{ margin: 0, fontSize: 13, color: 'var(--hr-text-muted)' }}>
          {t('reports.noResults')}
        </p>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {/* Actions bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
          {t('reports.rowCount', { count: report.rows.length })}
        </span>
        <div style={{ display: 'flex', gap: 8 }}>
          <Button variant="secondary" icon={<Download size={12} />} onClick={handleExportCsv}>
            {t('reports.exportCsv')}
          </Button>
          <Button variant="ghost" disabled>
            {t('reports.exportPdf')}
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
