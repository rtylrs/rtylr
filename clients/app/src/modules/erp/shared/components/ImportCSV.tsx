import React, { useRef, useState } from 'react';
import { Upload, FileText, X } from 'lucide-react';
import { toast } from 'sonner';

interface ImportCSVProps {
  onImport: (rows: Array<Record<string, string>>) => void | number | Promise<void | number>;
  expectedColumns?: string[];
}

// Quote-aware so exports from ExportButton (which quotes fields containing
// commas, quotes or newlines) can be re-imported without shifting columns.
function parseCSV(text: string): Array<Record<string, string>> {
  const records: string[][] = [];
  let record: string[] = [];
  let field = '';
  let quoted = false;
  let fieldQuoted = false;

  const endField = (): void => {
    record.push(fieldQuoted ? field : field.trim());
    field = '';
    fieldQuoted = false;
  };

  const endRecord = (): void => {
    endField();
    if (record.some((value) => value !== '')) records.push(record);
    record = [];
  };

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    if (quoted) {
      if (char === '"') {
        if (text[i + 1] === '"') {
          field += '"';
          i += 1;
        } else {
          quoted = false;
        }
      } else {
        field += char;
      }
      continue;
    }
    if (char === '"' && field.trim() === '') {
      quoted = true;
      fieldQuoted = true;
      field = '';
      continue;
    }
    if (char === ',') {
      endField();
      continue;
    }
    if (char === '\n' || char === '\r') {
      if (char === '\r' && text[i + 1] === '\n') i += 1;
      endRecord();
      continue;
    }
    field += char;
  }
  if (field !== '' || record.length > 0) endRecord();

  if (records.length < 2) return [];
  const headers = records[0];
  return records.slice(1).map((values) => {
    const row: Record<string, string> = {};
    headers.forEach((h, i) => {
      row[h] = values[i] ?? '';
    });
    return row;
  });
}

const ImportCSV: React.FC<ImportCSVProps> = ({ onImport, expectedColumns }) => {
  const fileRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<Array<Record<string, string>> | null>(null);
  const [importing, setImporting] = useState(false);

  const handleFile = async (file: File) => {
    const text = await file.text();
    const rows = parseCSV(text);
    if (rows.length === 0) {
      toast.error('CSV is empty or invalid');
      return;
    }
    setPreview(rows);
  };

  const confirm = async (): Promise<void> => {
    if (!preview || importing) return;
    setImporting(true);
    try {
      const created = await onImport(preview);
      toast.success(`Imported ${created ?? preview.length} rows`);
      setPreview(null);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Import failed');
    } finally {
      setImporting(false);
    }
  };

  return (
    <div>
      <input
        ref={fileRef}
        data-testid="csv-file-input"
        type="file"
        accept=".csv"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
        }}
      />
      <button
        type="button"
        onClick={() => fileRef.current?.click()}
        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
        style={{
          background: 'var(--hub-hover)',
          color: 'var(--hub-text-muted)',
          border: '1px solid var(--hub-border)',
        }}
      >
        <Upload size={12} />
        Import CSV
      </button>
      {preview && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-6"
          style={{ background: 'var(--hub-overlay-heavy)' }}
        >
          <div
            className="w-full max-w-3xl rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center justify-between mb-3">
              <h3
                className="text-sm font-semibold flex items-center gap-2"
                style={{ color: 'var(--hub-text)' }}
              >
                <FileText size={14} /> Preview ({preview.length} rows)
              </h3>
              <button type="button" onClick={() => setPreview(null)}>
                <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
              </button>
            </div>
            <div
              className="overflow-auto max-h-64 rounded-lg"
              style={{ border: '1px solid var(--hub-border)' }}
            >
              <table className="w-full text-xs">
                <thead>
                  <tr style={{ background: 'var(--hub-hover)' }}>
                    {Object.keys(preview[0]).map((k) => (
                      <th
                        key={k}
                        className="p-2 text-left font-semibold"
                        style={{ color: 'var(--hub-text-muted)' }}
                      >
                        {k}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {preview.slice(0, 10).map((row, i) => (
                    <tr key={i} style={{ borderTop: '1px solid var(--hub-border)' }}>
                      {Object.values(row).map((v, j) => (
                        <td key={j} className="p-2" style={{ color: 'var(--hub-text)' }}>
                          {v}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {expectedColumns && (
              <p className="text-[10px] mt-2" style={{ color: 'var(--hub-text-dim)' }}>
                Expected columns: {expectedColumns.join(', ')}
              </p>
            )}
            <div className="flex gap-2 justify-end mt-4">
              <button
                type="button"
                onClick={() => setPreview(null)}
                className="px-3 py-1.5 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                disabled={importing}
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={confirm}
                className="px-3 py-1.5 rounded-lg text-xs font-semibold"
                style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
                disabled={importing}
              >
                {importing ? 'Importing…' : `Import ${preview.length} rows`}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ImportCSV;
