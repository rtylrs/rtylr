import React, { useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { Download, FileUp, Undo2 } from 'lucide-react';
import { toast } from 'sonner';
import { Breadcrumbs, Button, ConfirmModal } from '../../shared/components/ui';
import { parseCsv, toCsv } from '../../shared/lib/csv';
import {
  type ColumnMapping,
  guessMapping,
  IMPORT_FIELDS,
  type ImportFailure,
  type ImportField,
  type ImportProgress,
  mappingNamesLead,
  rowToLead,
  runLeadImport,
} from './data/leadImport';
import { dateOf, personName } from './data/model';
import { useLeadFilters } from './data/leadFilters';
import { importClient, salesErrorMessage, useLeadImports, useUndoImport } from './data/queries';
import { RepPicker } from './components/common';
import { BillingSelect, Field, FieldGrid, TextInput } from './components/fields';

const MAX_ROWS = 200_000;

interface ParsedFile {
  name: string;
  headers: string[];
  rows: string[][];
}

type Step = 'upload' | 'map' | 'running' | 'done';

function downloadCsv(name: string, content: string): void {
  const url = URL.createObjectURL(new Blob(['﻿', content], { type: 'text/csv;charset=utf-8' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  link.click();
  URL.revokeObjectURL(url);
}

const RecentImports: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const setFilters = useLeadFilters((state) => state.setFilters);
  const { data: imports = [] } = useLeadImports();
  const undo = useUndoImport();
  const [undoing, setUndoing] = useState<string | null>(null);
  if (imports.length === 0) return null;
  const format = (value?: number) => new Intl.NumberFormat(i18n.language).format(value ?? 0);
  return (
    <section className="sales-card" aria-label={t('crmSales.import.recent')}>
      <div style={{ padding: '12px 12px 0', fontWeight: 600, fontSize: 13 }}>
        {t('crmSales.import.recent')}
      </div>
      <div className="sales-table-wrap">
        <table className="sales-table">
          <thead>
            <tr>
              <th>{t('crmSales.import.file')}</th>
              <th>{t('crmSales.import.when')}</th>
              <th>{t('crmSales.import.counts.created')}</th>
              <th>{t('crmSales.import.counts.updated')}</th>
              <th>{t('crmSales.import.counts.skipped')}</th>
              <th>{t('crmSales.import.counts.failed')}</th>
              <th>{t('crmSales.fields.status')}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {imports.map((item) => (
              <tr
                key={item.id}
                onClick={() => {
                  setFilters({ importId: item.id, statuses: [] });
                  void navigate({ to: '/crm/sales/leads' });
                }}
              >
                <td className="sales-cell-main">{item.fileName || '—'}</td>
                <td className="sales-muted">
                  {dateOf(item.createdAt)?.toLocaleString(i18n.language, {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  })}
                </td>
                <td>{format(item.created)}</td>
                <td>{format(item.updated)}</td>
                <td>{format(item.skipped)}</td>
                <td>{format(item.failed)}</td>
                <td>
                  <span className="sales-badge">
                    {t(`crmSales.import.status.${item.status ?? 'running'}`)}
                  </span>
                </td>
                <td onClick={(event) => event.stopPropagation()}>
                  {item.status !== 'undone' && (
                    <Button variant="ghost" onClick={() => setUndoing(item.id ?? null)}>
                      <Undo2 size={13} />
                      {t('crmSales.import.undo')}
                    </Button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <ConfirmModal
        open={undoing !== null}
        onClose={() => setUndoing(null)}
        onConfirm={() =>
          undoing &&
          undo.mutate(undoing, {
            onSuccess: (result) =>
              toast.success(
                t('crmSales.import.undone', {
                  removed: result.removed ?? 0,
                  kept: result.kept ?? 0,
                })
              ),
          })
        }
        title={t('crmSales.import.undoTitle')}
        description={t('crmSales.import.undoHint')}
        confirmLabel={t('crmSales.import.undo')}
        variant="danger"
      />
    </section>
  );
};

const LeadImport: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const setFilters = useLeadFilters((state) => state.setFilters);
  const [step, setStep] = useState<Step>('upload');
  const [file, setFile] = useState<ParsedFile | null>(null);
  const [mapping, setMapping] = useState<ColumnMapping>({});
  const [duplicates, setDuplicates] = useState<'skip' | 'fill'>('skip');
  const [assignedTo, setAssignedTo] = useState<string | undefined>();
  const [source, setSource] = useState('');
  const [billing, setBilling] = useState('one_time');
  const [progress, setProgress] = useState<ImportProgress | null>(null);
  const [failures, setFailures] = useState<ImportFailure[]>([]);
  const [dragging, setDragging] = useState(false);
  const [readError, setReadError] = useState('');
  const abort = useRef<AbortController | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const format = (value?: number) => new Intl.NumberFormat(i18n.language).format(value ?? 0);

  const readFile = (picked: File): void => {
    setReadError('');
    if (/\.(xlsx|xls|numbers)$/i.test(picked.name)) {
      setReadError(t('crmSales.import.spreadsheetFormat'));
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const parsed = parseCsv(String(reader.result ?? ''));
      if (parsed.length < 2) {
        setReadError(t('crmSales.import.emptyFile'));
        return;
      }
      const [headers, ...rows] = parsed;
      if (rows.length > MAX_ROWS) {
        setReadError(t('crmSales.import.tooManyRows', { max: format(MAX_ROWS) }));
        return;
      }
      setFile({ name: picked.name, headers, rows });
      setMapping(guessMapping(headers));
      setSource(picked.name.replace(/\.[^.]+$/, ''));
      setStep('map');
    };
    reader.onerror = () => setReadError(t('crmSales.import.readFailed'));
    reader.readAsText(picked);
  };

  const preview = useMemo(
    () => (file ? file.rows.slice(0, 3).map((row) => rowToLead(row, mapping, billing)) : []),
    [file, mapping, billing]
  );

  const start = async (): Promise<void> => {
    if (!file) return;
    abort.current = new AbortController();
    setStep('running');
    setProgress({
      sent: 0,
      total: file.rows.length,
      created: 0,
      updated: 0,
      skipped: 0,
      failed: 0,
    });
    try {
      const result = await runLeadImport(
        file.rows,
        mapping,
        { fileName: file.name, duplicates, assignedTo, source: source.trim(), billing },
        {
          start: async (fileName, totalRows) =>
            (await importClient.startLeadImport({ fileName, totalRows })).import ?? {},
          send: (id, body) => importClient.importLeadRows({ id, body }),
          finish: (id) => importClient.finishLeadImport({ id }),
        },
        setProgress,
        abort.current.signal
      );
      setProgress(result.progress);
      setFailures(result.failures);
      setStep('done');
    } catch (error) {
      toast.error(salesErrorMessage(error, 'crmSales.errors.import'));
      setStep('done');
    } finally {
      void qc.invalidateQueries({ queryKey: ['crm-sales'] });
    }
  };

  const downloadFailures = (): void => {
    if (!file) return;
    const header = [...file.headers, t('crmSales.import.errorColumn')];
    const rows = failures.map((failure) => [...(file.rows[failure.line - 2] ?? []), failure.error]);
    downloadCsv(`${file.name.replace(/\.[^.]+$/, '')}-failed.csv`, toCsv([header, ...rows]));
  };

  const viewImported = (): void => {
    if (!progress?.importId) return;
    setFilters({ importId: progress.importId, statuses: [] });
    void navigate({ to: '/crm/sales/leads' });
  };

  const percent =
    progress && progress.total > 0 ? Math.round((progress.sent / progress.total) * 100) : 0;

  return (
    <div className="sales-page">
      <HelmetContainer title={`${t('crmSales.import.title')} · rtylr`} />
      <Breadcrumbs
        items={[
          { label: t('crmSales.leads.title'), to: '/crm/sales/leads' },
          { label: t('crmSales.import.title') },
        ]}
      />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.import.title')}</h1>
          <p>{t('crmSales.import.subtitle')}</p>
        </div>
      </header>

      {step === 'upload' && (
        <>
          <div
            className={`sales-dropzone${dragging ? ' is-over' : ''}`}
            role="button"
            tabIndex={0}
            onClick={() => input.current?.click()}
            onKeyDown={(event) => {
              if (event.key === 'Enter' || event.key === ' ') input.current?.click();
            }}
            onDragOver={(event) => {
              event.preventDefault();
              setDragging(true);
            }}
            onDragLeave={() => setDragging(false)}
            onDrop={(event) => {
              event.preventDefault();
              setDragging(false);
              const picked = event.dataTransfer.files[0];
              if (picked) readFile(picked);
            }}
          >
            <FileUp size={26} aria-hidden="true" />
            <strong style={{ color: 'var(--hub-text)' }}>{t('crmSales.import.drop')}</strong>
            <span style={{ fontSize: 13 }}>{t('crmSales.import.formats')}</span>
            <input
              ref={input}
              type="file"
              accept=".csv,.tsv,.txt,text/csv"
              className="sales-visually-hidden"
              aria-label={t('crmSales.import.choose')}
              onChange={(event) => {
                const picked = event.target.files?.[0];
                if (picked) readFile(picked);
                event.target.value = '';
              }}
            />
          </div>
          {readError && (
            <p role="alert" style={{ margin: 0, color: '#ef4444', fontSize: 13 }}>
              {readError}
            </p>
          )}
          <RecentImports />
        </>
      )}

      {step === 'map' && file && (
        <>
          <section className="sales-panel">
            <h2>
              {t('crmSales.import.mapTitle', {
                count: file.rows.length,
                rows: format(file.rows.length),
                file: file.name,
              })}
            </h2>
            <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
              {t('crmSales.import.mapHint')}
            </p>
            <div className="sales-mapping">
              {IMPORT_FIELDS.map((field: ImportField) => {
                const column = mapping[field];
                const sample =
                  column === undefined
                    ? ''
                    : file.rows
                        .slice(0, 3)
                        .map((row) => row[column])
                        .filter(Boolean)
                        .join(' · ');
                return (
                  <React.Fragment key={field}>
                    <label htmlFor={`map-${field}`}>{t(`crmSales.import.fields.${field}`)}</label>
                    <select
                      id={`map-${field}`}
                      className="sales-select"
                      style={{ maxWidth: 'none' }}
                      value={column === undefined ? '' : String(column)}
                      onChange={(event) =>
                        setMapping((current) => {
                          const next = { ...current };
                          if (event.target.value === '') delete next[field];
                          else next[field] = Number(event.target.value);
                          return next;
                        })
                      }
                    >
                      <option value="">{t('crmSales.import.skipColumn')}</option>
                      {file.headers.map((header, index) => (
                        <option key={index} value={index}>
                          {header || t('crmSales.import.column', { number: index + 1 })}
                        </option>
                      ))}
                    </select>
                    <span className="sales-sample">{sample}</span>
                  </React.Fragment>
                );
              })}
            </div>
          </section>

          <section className="sales-panel">
            <h2>{t('crmSales.import.options')}</h2>
            <FieldGrid>
              <Field
                label={t('crmSales.import.duplicates')}
                hint={t('crmSales.import.duplicatesHint')}
              >
                {(id) => (
                  <select
                    id={id}
                    className="sales-select"
                    style={{ maxWidth: 'none' }}
                    value={duplicates}
                    onChange={(event) => setDuplicates(event.target.value as 'skip' | 'fill')}
                  >
                    <option value="skip">{t('crmSales.import.skipDuplicates')}</option>
                    <option value="fill">{t('crmSales.import.fillDuplicates')}</option>
                  </select>
                )}
              </Field>
              <Field label={t('crmSales.import.assignTo')}>
                {(id) => (
                  <RepPicker
                    id={id}
                    label={t('crmSales.import.assignTo')}
                    value={assignedTo}
                    onChange={setAssignedTo}
                  />
                )}
              </Field>
              <Field label={t('crmSales.fields.source')} hint={t('crmSales.import.sourceHint')}>
                {(id) => <TextInput id={id} value={source} onChange={setSource} maxLength={100} />}
              </Field>
              <Field label={t('crmSales.import.valueBilling')}>
                {(id) => <BillingSelect id={id} value={billing} onChange={setBilling} />}
              </Field>
            </FieldGrid>
          </section>

          <section className="sales-card" aria-label={t('crmSales.import.preview')}>
            <div style={{ padding: '12px 12px 0', fontWeight: 600, fontSize: 13 }}>
              {t('crmSales.import.preview')}
            </div>
            <div className="sales-table-wrap">
              <table className="sales-table">
                <thead>
                  <tr>
                    <th>{t('crmSales.fields.person')}</th>
                    <th>{t('crmSales.fields.company')}</th>
                    <th>{t('crmSales.fields.phone')}</th>
                    <th>{t('crmSales.fields.email')}</th>
                  </tr>
                </thead>
                <tbody>
                  {preview.map((lead, index) => (
                    <tr key={index}>
                      <td>{personName(lead) || '—'}</td>
                      <td>{lead.companyName || '—'}</td>
                      <td>{lead.phone || '—'}</td>
                      <td>{lead.email || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>

          {!mappingNamesLead(mapping) && (
            <p role="alert" style={{ margin: 0, color: '#ef4444', fontSize: 13 }}>
              {t('crmSales.import.needsName')}
            </p>
          )}
          <div className="sales-actions">
            <Button variant="outline" onClick={() => setStep('upload')}>
              {t('crmSales.import.chooseAnother')}
            </Button>
            <Button onClick={() => void start()} disabled={!mappingNamesLead(mapping)}>
              {t('crmSales.import.start', {
                count: file.rows.length,
                rows: format(file.rows.length),
              })}
            </Button>
          </div>
        </>
      )}

      {(step === 'running' || step === 'done') && progress && (
        <section className="sales-panel" aria-live="polite">
          <h2>
            {step === 'running' ? t('crmSales.import.running') : t('crmSales.import.finished')}
          </h2>
          <div
            className="sales-progress"
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={percent}
            aria-label={t('crmSales.import.progress')}
          >
            <div style={{ width: `${percent}%` }} />
          </div>
          <p style={{ margin: 0, fontSize: 13 }}>
            {t('crmSales.import.sentOf', {
              sent: format(progress.sent),
              total: format(progress.total),
            })}
          </p>
          <div className="sales-stats">
            {(['created', 'updated', 'skipped', 'failed'] as const).map((key) => (
              <div
                key={key}
                className={`sales-stat${key === 'failed' && progress.failed > 0 ? ' is-warning' : ''}`}
              >
                <div className="sales-stat-label">{t(`crmSales.import.counts.${key}`)}</div>
                <div className="sales-stat-value">{format(progress[key])}</div>
              </div>
            ))}
          </div>
          <div className="sales-actions">
            {step === 'running' ? (
              <Button variant="outline" onClick={() => abort.current?.abort()}>
                {t('crmSales.import.stop')}
              </Button>
            ) : (
              <>
                {progress.importId && (
                  <Button onClick={viewImported}>{t('crmSales.import.viewLeads')}</Button>
                )}
                {failures.length > 0 && (
                  <Button variant="outline" onClick={downloadFailures}>
                    <Download size={13} />
                    {t('crmSales.import.downloadFailed', { count: failures.length })}
                  </Button>
                )}
                <Button
                  variant="ghost"
                  onClick={() => {
                    setFile(null);
                    setProgress(null);
                    setFailures([]);
                    setStep('upload');
                  }}
                >
                  {t('crmSales.import.another')}
                </Button>
              </>
            )}
          </div>
        </section>
      )}
    </div>
  );
};

export default LeadImport;
