import React, { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Download, FileText, Upload, UserPlus } from 'lucide-react';
import {
  type CreateEmployeePayload,
  useCreateEmployee,
  useListEmployees,
} from '../../../shared/hooks/queries/useEmployees';
import { Button, Modal } from '../../../shared/components/ui';
import StepContainer from '../components/StepContainer';

interface EmployeesStepProps {
  onNext: () => void;
  onBack: () => void;
  onAddEmployee: () => void;
  isSubmitting?: boolean;
  isLast?: boolean;
}

interface CsvPreviewRow {
  line: number;
  payload?: CreateEmployeePayload;
  error?: string;
}

const parseCsv = (input: string): Record<string, string>[] => {
  const text = input.replace(/^\uFEFF/, '');
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = '';
  let quoted = false;

  for (let index = 0; index < text.length; index += 1) {
    const character = text[index];
    if (quoted) {
      if (character === '"' && text[index + 1] === '"') {
        cell += '"';
        index += 1;
      } else if (character === '"') {
        quoted = false;
      } else {
        cell += character;
      }
    } else if (character === '"') {
      quoted = true;
    } else if (character === ',') {
      row.push(cell.trim());
      cell = '';
    } else if (character === '\n' || character === '\r') {
      if (character === '\r' && text[index + 1] === '\n') index += 1;
      row.push(cell.trim());
      if (row.some((value) => value !== '')) rows.push(row);
      row = [];
      cell = '';
    } else {
      cell += character;
    }
  }

  row.push(cell.trim());
  if (row.some((value) => value !== '')) rows.push(row);
  if (rows.length < 2) return [];

  const headers = rows[0].map((header) =>
    header.trim().toLowerCase().replace(/[\s-]+/g, '_')
  );
  return rows.slice(1).map((values) =>
    Object.fromEntries(headers.map((header, index) => [header, values[index] ?? '']))
  );
};

const parsePositiveId = (row: Record<string, string>, field: string): number => {
  const value = Number(row[field]);
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error(`${field} must be a positive integer`);
  }
  return value;
};

const parseOptionalNumber = (row: Record<string, string>, field: string): number | undefined => {
  if (!row[field]) return undefined;
  const value = Number(row[field]);
  if (!Number.isFinite(value) || value < 0) throw new Error(`${field} must be a non-negative number`);
  return value;
};

const toEmployeePayload = (row: Record<string, string>): CreateEmployeePayload => {
  const email = row.email?.trim() ?? '';
  if (!email || !email.includes('@')) throw new Error('email is required and must be valid');
  const firstName = row.first_name?.trim() ?? '';
  const lastName = row.last_name?.trim() ?? '';
  if (!firstName || !lastName) throw new Error('first_name and last_name are required');

  const status = row.status || 'active';
  if (!['active', 'probation', 'inactive', 'terminated'].includes(status)) {
    throw new Error('status must be active, probation, inactive, or terminated');
  }

  return {
    userId: parsePositiveId(row, 'user_id'),
    departmentId: parsePositiveId(row, 'department_id'),
    locationId: parsePositiveId(row, 'location_id'),
    workScheduleId: parseOptionalNumber(row, 'work_schedule_id'),
    managerId: parseOptionalNumber(row, 'manager_id'),
    firstName,
    lastName,
    email,
    employmentType: row.employment_type || 'full_time',
    status,
    jobTitle: row.job_title || 'Employee',
    currency: (row.currency || 'USD').toUpperCase(),
    accountCreationMode: row.account_creation_mode || 'hr_created',
    salary: parseOptionalNumber(row, 'salary'),
    hourlyRate: parseOptionalNumber(row, 'hourly_rate'),
    overtimeRate: parseOptionalNumber(row, 'overtime_rate'),
    startDate: row.start_date || new Date().toISOString().slice(0, 10),
    dateOfBirth: row.date_of_birth || undefined,
    probationEndDate: row.probation_end_date || undefined,
    endDate: row.end_date || undefined,
    terminationDate: row.termination_date || undefined,
  };
};

const CSV_TEMPLATE =
  'user_id,department_id,location_id,first_name,last_name,email,employment_type,status,job_title,currency,account_creation_mode,start_date,date_of_birth,work_schedule_id,manager_id,salary,hourly_rate,overtime_rate,probation_end_date,end_date,termination_date\n' +
  '123,1,1,Jane,Doe,jane@example.com,full_time,active,Engineer,USD,hr_created,2026-01-15,,,,,,,,';

const EmployeesStep = ({
  onNext,
  onBack,
  onAddEmployee,
  isSubmitting,
  isLast,
}: EmployeesStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const { data } = useListEmployees({ limit: 50, offset: 0 });
  const createEmployee = useCreateEmployee();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [csvOpen, setCsvOpen] = useState(false);
  const [csvRows, setCsvRows] = useState<CsvPreviewRow[]>([]);
  const [csvError, setCsvError] = useState('');
  const [csvImporting, setCsvImporting] = useState(false);
  const [csvResult, setCsvResult] = useState('');
  const employees = data?.employees ?? [];

  const handleCsvFile = async (event: React.ChangeEvent<HTMLInputElement>): Promise<void> => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    setCsvError('');
    setCsvResult('');
    try {
      const records = parseCsv(await file.text());
      if (records.length === 0) {
        setCsvError(t('setup.employees.csvNoRows'));
        setCsvRows([]);
      } else {
        setCsvRows(
          records.map((record, index) => {
            try {
              return { line: index + 2, payload: toEmployeePayload(record) };
            } catch (error) {
              return {
                line: index + 2,
                error: error instanceof Error ? error.message : t('setup.employees.csvReadError'),
              };
            }
          })
        );
      }
      setCsvOpen(true);
    } catch {
      setCsvError(t('setup.employees.csvReadError'));
      setCsvRows([]);
      setCsvOpen(true);
    }
  };

  const downloadCsvTemplate = (): void => {
    const blob = new Blob([CSV_TEMPLATE], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = 'employee-import-template.csv';
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const handleCsvSubmit = async (): Promise<void> => {
    if (csvRows.length === 0 || csvRows.some((row) => !row.payload)) return;
    setCsvImporting(true);
    let imported = 0;
    for (const row of csvRows) {
      if (!row.payload) continue;
      try {
        await createEmployee.mutateAsync(row.payload);
        imported += 1;
      } catch {
        // Keep processing independent rows and report the count to the admin.
      }
    }
    setCsvImporting(false);
    setCsvResult(t('setup.employees.csvImported', { count: imported }));
    if (imported === csvRows.length) setCsvOpen(false);
  };

  return (
    <StepContainer
      title={t('setup.employees.title')}
      description={t('setup.employees.description')}
      onBack={onBack}
      onNext={onNext}
      isLast={isLast}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <button
            type="button"
            onClick={onAddEmployee}
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium"
            style={{ background: 'var(--accent)', color: '#fff' }}
          >
            <UserPlus size={14} />
            {t('setup.employees.addEmployee')}
          </button>
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium"
            style={{
              background: 'var(--hr-hover)',
              color: 'var(--hr-text-muted)',
              border: '1px solid var(--hr-border)',
            }}
          >
            <Upload size={14} />
            {t('setup.employees.importCsv')}
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv,text/csv"
            onChange={handleCsvFile}
            hidden
          />
        </div>

        {employees.length === 0 ? (
          <p className="py-8 text-center text-sm" style={{ color: 'var(--hr-text-dim)' }}>
            {t('setup.employees.noEmployeesHint')}
          </p>
        ) : (
          <div
            className="rounded-lg overflow-hidden"
            style={{ border: '1px solid var(--hr-border)' }}
          >
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: 'var(--hr-hover)' }}>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('employees.jobTitle')}
                  </th>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('employees.contractType')}
                  </th>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('common.status')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {employees.map((emp) => (
                  <tr key={emp.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                    <td className="px-4 py-2" style={{ color: 'var(--hr-text)' }}>
                      {emp.jobTitle}
                    </td>
                    <td className="px-4 py-2" style={{ color: 'var(--hr-text-muted)' }}>
                      {emp.employmentType}
                    </td>
                    <td className="px-4 py-2">
                      <span
                        className="rounded-full px-2 py-0.5 text-xs font-medium"
                        style={{
                          background:
                            emp.status === 'active' ? 'rgba(34,197,94,0.15)' : 'var(--hr-hover)',
                          color: emp.status === 'active' ? '#22c55e' : 'var(--hr-text-muted)',
                        }}
                      >
                        {emp.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal
        open={csvOpen}
        onClose={() => {
          if (!csvImporting) setCsvOpen(false);
        }}
        title={t('setup.employees.csvModalTitle')}
        width={720}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <p style={{ margin: 0, color: 'var(--hr-text-muted)', fontSize: 12 }}>
            {t('setup.employees.csvDescription')}
          </p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Button type="button" variant="ghost" onClick={downloadCsvTemplate} icon={<Download size={13} />}>
              {t('setup.employees.csvDownloadTemplate')}
            </Button>
            <Button type="button" variant="secondary" onClick={() => fileInputRef.current?.click()} icon={<Upload size={13} />}>
              {t('setup.employees.importCsv')}
            </Button>
          </div>
          {csvError && <p style={{ margin: 0, color: 'var(--hr-error)', fontSize: 12 }}>{csvError}</p>}
          {csvRows.length > 0 && (
            <>
              <p style={{ margin: 0, color: 'var(--hr-text-muted)', fontSize: 12 }}>
                {t('setup.employees.csvRowsReady', { count: csvRows.length })}
              </p>
              <div style={{ maxHeight: 260, overflow: 'auto', border: '1px solid var(--hr-border)', borderRadius: 8 }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
                  <thead>
                    <tr style={{ background: 'var(--hr-hover)' }}>
                      <th style={{ padding: 8, textAlign: 'left' }}>{t('setup.employees.csvLine')}</th>
                      <th style={{ padding: 8, textAlign: 'left' }}>{t('common.name')}</th>
                      <th style={{ padding: 8, textAlign: 'left' }}>{t('common.email')}</th>
                      <th style={{ padding: 8, textAlign: 'left' }}>{t('common.status')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {csvRows.map((row) => (
                      <tr key={row.line} style={{ borderTop: '1px solid var(--hr-border)' }}>
                        <td style={{ padding: 8 }}>{row.line}</td>
                        <td style={{ padding: 8 }}>{row.payload ? `${row.payload.firstName} ${row.payload.lastName}` : '-'}</td>
                        <td style={{ padding: 8 }}>{row.payload?.email ?? '-'}</td>
                        <td style={{ padding: 8, color: row.error ? 'var(--hr-error)' : 'var(--hr-status-approved)' }}>
                          {row.error ?? t('setup.employees.csvReady')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {csvRows.some((row) => row.error) && (
                <p style={{ margin: 0, color: 'var(--hr-error)', fontSize: 12 }}>
                  {t('setup.employees.csvInvalidRows')}
                </p>
              )}
            </>
          )}
          {csvResult && <p style={{ margin: 0, color: 'var(--hr-status-approved)', fontSize: 12 }}>{csvResult}</p>}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" onClick={() => setCsvOpen(false)} disabled={csvImporting}>
              {t('common.close')}
            </Button>
            <Button
              type="button"
              variant="primary"
              onClick={handleCsvSubmit}
              disabled={csvImporting || csvRows.length === 0 || csvRows.some((row) => !row.payload)}
              icon={<FileText size={13} />}
            >
              {csvImporting ? t('setup.employees.csvImporting') : t('setup.employees.csvImport')}
            </Button>
          </div>
        </div>
      </Modal>
    </StepContainer>
  );
};

export default EmployeesStep;
