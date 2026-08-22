import React, { useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Upload, Check } from 'lucide-react';
import { useImportCustomerRows } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, PageHeader, Select } from '../../shared/components/ui';

export const IMPORT_FIELD_KEYS = [
  'name',
  'first_name',
  'last_name',
  'email',
  'phone',
  'date_of_birth',
  'gender',
  'address',
  'city',
  'country',
  'tags',
  'notes',
] as const;

type ImportField = (typeof IMPORT_FIELD_KEYS)[number];

const IMPORT_FIELD_LABELS: Record<ImportField, string> = {
  name: 'Full name',
  first_name: 'First name',
  last_name: 'Last name',
  email: 'Email',
  phone: 'Phone',
  date_of_birth: 'Date of birth',
  gender: 'Gender',
  address: 'Address',
  city: 'City',
  country: 'Country',
  tags: 'Tags',
  notes: 'Notes',
};

const HEADER_ALIASES: Record<ImportField, string[]> = {
  name: ['name', 'full_name', 'fullname', 'customer_name'],
  first_name: ['first_name', 'firstname', 'given_name'],
  last_name: ['last_name', 'lastname', 'surname', 'family_name'],
  email: ['email', 'email_address', 'emailaddress'],
  phone: ['phone', 'phone_number', 'phonenumber', 'mobile', 'mobile_number'],
  date_of_birth: ['date_of_birth', 'dateofbirth', 'birthday', 'birth_date'],
  gender: ['gender', 'sex'],
  address: ['address', 'street', 'street_address'],
  city: ['city', 'town'],
  country: ['country', 'country_code'],
  tags: ['tags', 'tag', 'labels'],
  notes: ['notes', 'note', 'comments', 'comment'],
};

function normalizeHeader(header: string): string {
  return header
    .replace(/^\uFEFF/, '')
    .trim()
    .toLowerCase()
    .replace(/[\s-]+/g, '_');
}

const DUPLICATE_COLUMN_PREFIX = '__crm_import_column__';

/** Keeps duplicate CSV headers addressable without changing the mapping shape. */
export function importColumnValue(headers: string[], index: number): string {
  const header = headers[index] ?? '';
  return headers.indexOf(header) === headers.lastIndexOf(header)
    ? header
    : `${DUPLICATE_COLUMN_PREFIX}${index}:${header}`;
}

function importColumnIndex(headers: string[], column: string): number {
  if (column.startsWith(DUPLICATE_COLUMN_PREFIX)) {
    const separatorIndex = column.indexOf(':', DUPLICATE_COLUMN_PREFIX.length);
    const index = Number(column.slice(DUPLICATE_COLUMN_PREFIX.length, separatorIndex));
    const header = column.slice(separatorIndex + 1);
    if (
      separatorIndex >= 0 &&
      Number.isInteger(index) &&
      index >= 0 &&
      index < headers.length &&
      headers[index] === header
    ) {
      return index;
    }
  }
  return headers.indexOf(column);
}

function findHeaderIndex(
  headers: string[],
  aliases: string[],
  excludedIndices: Set<number>
): number {
  const normalized = headers.map(normalizeHeader);
  const exact = normalized.findIndex(
    (header, index) => !excludedIndices.has(index) && aliases.includes(header)
  );
  if (exact >= 0) return exact;
  return normalized.findIndex(
    (header, index) =>
      !excludedIndices.has(index) &&
      aliases.some((alias) => alias.length > 3 && header.includes(alias))
  );
}

export function autoMapHeaders(headers: string[]): Record<ImportField, string> {
  const mapped = {} as Record<ImportField, string>;
  const usedIndices = new Set<number>();

  // Resolve exact aliases for every field before fuzzy matching. This keeps a
  // generic alias such as "name" from claiming a more specific column such
  // as "first_name".
  IMPORT_FIELD_KEYS.forEach((field) => {
    const normalized = headers.map(normalizeHeader);
    const index = normalized.findIndex(
      (header, headerIndex) =>
        !usedIndices.has(headerIndex) && HEADER_ALIASES[field].includes(header)
    );
    if (index >= 0) {
      mapped[field] = importColumnValue(headers, index);
      usedIndices.add(index);
    }
  });

  IMPORT_FIELD_KEYS.forEach((field) => {
    if (mapped[field]) return;
    const index = findHeaderIndex(
      headers,
      HEADER_ALIASES[field].filter((alias) => alias !== 'name'),
      usedIndices
    );
    if (index >= 0) {
      mapped[field] = importColumnValue(headers, index);
      usedIndices.add(index);
    }
  });

  return mapped;
}

function mappedValue(
  row: string[],
  headers: string[],
  mapping: Partial<Record<ImportField, string>>,
  field: ImportField
): string {
  const column = mapping[field];
  if (!column) return '';
  const index = importColumnIndex(headers, column);
  return index >= 0 ? (row[index] ?? '').trim() : '';
}

export function toCustomerImportRow(
  row: string[],
  headers: string[],
  mapping: Partial<Record<ImportField, string>>
): Record<string, unknown> {
  const fullName = mappedValue(row, headers, mapping, 'name');
  const firstNameFromFullName = fullName.split(/\s+/).filter(Boolean);
  const firstName =
    mappedValue(row, headers, mapping, 'first_name') || firstNameFromFullName.shift() || '';
  const lastName =
    mappedValue(row, headers, mapping, 'last_name') || firstNameFromFullName.join(' ');
  const customer: Record<string, unknown> = {
    first_name: firstName,
    last_name: lastName,
    email: mappedValue(row, headers, mapping, 'email'),
    phone: mappedValue(row, headers, mapping, 'phone'),
  };

  (['date_of_birth', 'gender', 'address', 'city', 'country', 'notes'] as const).forEach((field) => {
    const value = mappedValue(row, headers, mapping, field);
    if (value) customer[field] = value;
  });

  const tags = mappedValue(row, headers, mapping, 'tags');
  if (tags) {
    customer.tags = tags
      .split(/[;|]/)
      .map((tag) => tag.trim())
      .filter(Boolean);
  }

  return customer;
}

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
  const [mapping, setMapping] = useState<Partial<Record<ImportField, string>>>({});
  const [dedupe, setDedupe] = useState(true);
  const [fileName, setFileName] = useState('');
  const [fileError, setFileError] = useState('');
  const fileReadId = useRef(0);

  const clearParsedFile = (): void => {
    setHeaders([]);
    setRows([]);
    setMapping({});
  };

  const handleFile = (f: File) => {
    const currentReadId = fileReadId.current + 1;
    fileReadId.current = currentReadId;
    setFileName(f.name);
    setFileError('');
    if (f.size === 0) {
      clearParsedFile();
      setFileError('The selected file is empty. Choose a CSV with headers and rows.');
      return;
    }
    const reader = new FileReader();
    reader.onerror = () => {
      if (currentReadId !== fileReadId.current) return;
      clearParsedFile();
      setFileError('Unable to read this file. Check the file and try again.');
    };
    reader.onload = () => {
      if (currentReadId !== fileReadId.current) return;
      if (typeof reader.result !== 'string') {
        clearParsedFile();
        setFileError('Unable to read this file. Check the file and try again.');
        return;
      }
      const text = reader.result;
      const parsed = parseCsv(text);
      if (parsed.length === 0) {
        clearParsedFile();
        setFileError('No CSV data was found. Choose a file with headers and rows.');
        return;
      }
      const hdr = parsed[0];
      const body = parsed.slice(1);
      setHeaders(hdr);
      setRows(body);
      setMapping(autoMapHeaders(hdr));
      if (body.length === 0) {
        setFileError('Add at least one data row below the header before importing.');
      }
    };
    reader.readAsText(f);
  };

  const handleImport = () => {
    const payload = rows.map((row) => toCustomerImportRow(row, headers, mapping));
    const invalidRows = payload.filter(
      (customer) => typeof customer.first_name !== 'string' || !customer.first_name.trim()
    ).length;
    if (!mapping.name && !mapping.first_name) {
      toast.error('Map a full name or first name column before importing');
      return;
    }
    if (invalidRows > 0) {
      toast.error(`${invalidRows} row${invalidRows === 1 ? '' : 's'} is missing a first name`);
      return;
    }
    importMut.mutate(
      { customers: payload, dedupe },
      {
        onSuccess: () => {
          navigate({ to: '/customers' });
        },
      }
    );
  };

  const mappedRows = rows.map((row) => toCustomerImportRow(row, headers, mapping));
  const invalidRowCount = mappedRows.filter(
    (customer) => typeof customer.first_name !== 'string' || !customer.first_name.trim()
  ).length;
  const hasNameMapping = Boolean(mapping.name || mapping.first_name);

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
              onChange={(e) => {
                const file = e.currentTarget.files?.[0];
                if (file) handleFile(file);
                e.currentTarget.value = '';
              }}
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
          {fileError && (
            <p role="alert" style={{ margin: '12px 0 0', fontSize: '12px', color: '#ef4444' }}>
              {fileError}
            </p>
          )}
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
                {IMPORT_FIELD_KEYS.map((field) => (
                  <div
                    key={field}
                    className="crm-import-mapping"
                    style={{ display: 'grid', gap: '12px', alignItems: 'center' }}
                  >
                    <div
                      style={{
                        fontSize: '12px',
                        color: 'var(--hub-text)',
                        fontWeight: 600,
                        textTransform: 'capitalize',
                      }}
                    >
                      {IMPORT_FIELD_LABELS[field]}
                    </div>
                    <Select
                      aria-label={`Map ${IMPORT_FIELD_LABELS[field]}`}
                      value={mapping[field] || ''}
                      onChange={(e) => setMapping((m) => ({ ...m, [field]: e.target.value }))}
                    >
                      <option value="">Select column...</option>
                      {headers.map((h, index) => (
                        <option key={`${h}-${index}`} value={importColumnValue(headers, index)}>
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
              {!hasNameMapping && (
                <p style={{ fontSize: '12px', color: '#f59e0b', margin: '12px 0 0' }}>
                  Map either Full name or First name to enable import.
                </p>
              )}
              {hasNameMapping && invalidRowCount > 0 && (
                <p style={{ fontSize: '12px', color: '#f59e0b', margin: '12px 0 0' }}>
                  {invalidRowCount} row{invalidRowCount === 1 ? '' : 's'} has no first name.
                </p>
              )}
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
                      {headers.map((h, index) => (
                        <th
                          key={`${h}-${index}`}
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
                disabled={
                  importMut.isPending || rows.length === 0 || !hasNameMapping || invalidRowCount > 0
                }
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
