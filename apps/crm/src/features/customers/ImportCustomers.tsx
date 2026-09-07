import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Upload, Check } from 'lucide-react';
import { useImportCustomerRows } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, PageHeader, Select } from '../../shared/components/ui';

const REQUIRED = ['name', 'email', 'phone'];

export function parseCsv(input: string): string[][] {
  const records: string[][] = [];
  let record: string[] = [];
  let field = '';
  let quoted = false;
  for (let i = 0; i < input.length; i += 1) {
    const char = input[i];
    const next = input[i + 1];
    if (char === '"') {
      if (quoted && next === '"') {
        field += '"';
        i += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (char === ',' && !quoted) {
      record.push(field.trim());
      field = '';
      continue;
    }
    if ((char === '\n' || char === '\r') && !quoted) {
      if (char === '\r' && next === '\n') i += 1;
      record.push(field.trim());
      field = '';
      if (record.some((value) => value !== '')) records.push(record);
      record = [];
      continue;
    }
    field += char;
  }
  if (field !== '' || record.length > 0) {
    record.push(field.trim());
    if (record.some((value) => value !== '')) records.push(record);
  }
  return records;
}

const ImportCustomers: React.FC = () => {
  const navigate = useNavigate();
  const importMut = useImportCustomerRows();
  const [rows, setRows] = useState<string[][]>([]);
  const [headers, setHeaders] = useState<string[]>([]);
  const [mapping, setMapping] = useState<Record<string, string>>({});
  const [dedupe, setDedupe] = useState(true);
  const [fileName, setFileName] = useState('');

  const handleFile = (f: File) => {
    setFileName(f.name);
    const reader = new FileReader();
    reader.onload = () => {
      const text = reader.result as string;
      const parsed = parseCsv(text);
      if (parsed.length === 0) return;
      const hdr = parsed[0];
      const body = parsed.slice(1);
      setHeaders(hdr);
      setRows(body);
      const autoMap: Record<string, string> = {};
      REQUIRED.forEach((req) => {
        const idx = hdr.findIndex((h) => h.toLowerCase().includes(req));
        if (idx >= 0) autoMap[req] = hdr[idx];
      });
      setMapping(autoMap);
    };
    reader.readAsText(f);
  };

  const handleImport = () => {
    const payload = rows.map((r) => {
      const o: Record<string, string> = {};
      REQUIRED.forEach((req) => {
        const col = mapping[req];
        if (col) {
          const idx = headers.indexOf(col);
          if (idx >= 0) o[req] = r[idx];
        }
      });
      // The backend reads first_name/last_name, not a single `name`. Split the
      // mapped name on the first whitespace so imported customers keep a name.
      const fullName = (o.name ?? '').trim();
      const firstSpace = fullName.indexOf(' ');
      const first_name = firstSpace === -1 ? fullName : fullName.slice(0, firstSpace);
      const last_name = firstSpace === -1 ? '' : fullName.slice(firstSpace + 1).trim();
      return { first_name, last_name, email: o.email ?? '', phone: o.phone ?? '' };
    });
    importMut.mutate(
      { customers: payload, dedupe },
      {
        onSuccess: () => {
          navigate({ to: '/customers' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Import customers · rtylr" />
      <div style={{ maxWidth: '820px' }}>
        <Breadcrumbs items={[{ label: 'Customers', to: '/customers' }, { label: 'Import' }]} />
        <PageHeader title="Import customers" subtitle="Upload a CSV and map the columns" />

        <Card style={{ marginBottom: '16px' }}>
          <label
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: '8px',
              padding: '32px',
              borderRadius: '12px',
              border: '1px dashed var(--hub-border)',
              background: 'var(--hub-hover)',
              cursor: 'pointer',
              textAlign: 'center',
            }}
          >
            <input
              type="file"
              accept=".csv"
              style={{ display: 'none' }}
              onChange={(e) => e.target.files?.[0] && handleFile(e.target.files[0])}
            />
            <div
              style={{
                width: '48px',
                height: '48px',
                borderRadius: '12px',
                background: 'rgba(59,130,246,0.12)',
                color: '#3b82f6',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Upload size={22} />
            </div>
            <div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)' }}>
              {fileName || 'Upload CSV file'}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--hub-text-muted)' }}>
              Click to browse. Your file should have headers in the first row.
            </div>
          </label>
        </Card>

        {headers.length > 0 && (
          <>
            <Card style={{ marginBottom: '16px' }}>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 12px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Column mapping
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {REQUIRED.map((req) => (
                  <div
                    key={req}
                    style={{
                      display: 'grid',
                      gridTemplateColumns: '120px 1fr',
                      gap: '12px',
                      alignItems: 'center',
                    }}
                  >
                    <div
                      style={{
                        fontSize: '12px',
                        color: 'var(--hub-text)',
                        fontWeight: 600,
                        textTransform: 'capitalize',
                      }}
                    >
                      {req}
                    </div>
                    <Select
                      value={mapping[req] || ''}
                      onChange={(e) => setMapping((m) => ({ ...m, [req]: e.target.value }))}
                    >
                      <option value="">Select column...</option>
                      {headers.map((h) => (
                        <option key={h} value={h}>
                          {h}
                        </option>
                      ))}
                    </Select>
                  </div>
                ))}
              </div>
              <label
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  marginTop: '14px',
                  fontSize: '12px',
                  color: 'var(--hub-text)',
                  cursor: 'pointer',
                }}
              >
                <input
                  type="checkbox"
                  checked={dedupe}
                  onChange={(e) => setDedupe(e.target.checked)}
                  style={{ accentColor: '#3b82f6' }}
                />
                Deduplicate by email
              </label>
            </Card>

            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 12px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Preview ({rows.length} rows)
              </h4>
              <div style={{ maxHeight: '240px', overflow: 'auto' }}>
                <table style={{ width: '100%', fontSize: '12px' }}>
                  <thead>
                    <tr>
                      {headers.map((h) => (
                        <th
                          key={h}
                          style={{
                            textAlign: 'left',
                            padding: '6px 8px',
                            color: 'var(--hub-text-dim)',
                            borderBottom: '1px solid var(--hub-border)',
                            fontWeight: 600,
                          }}
                        >
                          {h}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {rows.slice(0, 10).map((r, i) => (
                      <tr key={i} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                        {r.map((c, j) => (
                          <td key={j} style={{ padding: '6px 8px', color: 'var(--hub-text)' }}>
                            {c}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>

            <div
              style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '16px' }}
            >
              <Button variant="outline" onClick={() => navigate({ to: '/customers' })}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={handleImport}
                disabled={importMut.isPending || rows.length === 0}
              >
                <Check size={13} />
                {importMut.isPending ? 'Importing…' : `Import ${rows.length} customers`}
              </Button>
            </div>
          </>
        )}
      </div>
    </>
  );
};

export default ImportCustomers;
