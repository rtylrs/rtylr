import React from 'react';
import { Download } from 'lucide-react';
import { toast } from 'sonner';

interface ExportButtonProps {
  data: Array<Record<string, any>>;
  filename?: string;
  columns?: string[];
}

function toCSV(data: Array<Record<string, any>>, columns?: string[]): string {
  if (data.length === 0) return '';
  const keys = columns && columns.length ? columns : Object.keys(data[0]);
  const escape = (v: unknown) => {
    if (v == null) return '';
    const s = String(v);
    if (s.includes(',') || s.includes('"') || s.includes('\n')) {
      return `"${s.replace(/"/g, '""')}"`;
    }
    return s;
  };
  const header = keys.join(',');
  const rows = data.map((row) => keys.map((k) => escape(row[k])).join(','));
  return [header, ...rows].join('\n');
}

const ExportButton: React.FC<ExportButtonProps> = ({ data, filename = 'export.csv', columns }) => {
  const handleExport = () => {
    if (data.length === 0) {
      toast.error('No data to export');
      return;
    }
    const csv = toCSV(data, columns);
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
    toast.success(`Exported ${data.length} rows`);
  };

  return (
    <button
      onClick={handleExport}
      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
      style={{
        background: 'var(--hub-hover)',
        color: 'var(--hub-text-muted)',
        border: '1px solid var(--hub-border)',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.color = 'var(--hub-text)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.color = 'var(--hub-text-muted)';
      }}
    >
      <Download size={12} />
      Export CSV
    </button>
  );
};

export default ExportButton;
