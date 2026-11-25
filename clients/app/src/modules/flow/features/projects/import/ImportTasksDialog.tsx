import React, { useId, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { toast } from 'sonner';
import { Download, FileUp } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../../shared/components/ui/dialog';
import { Button } from '../../../../../shared/components/ui/button';
import { buildCsv, downloadCsv } from '../../../../../shared/lib/csv';
import { useImportTasks } from '../../tasks/useTasks';
import { useProjectMembers } from '../useProjectMembers';
import { useProjectStatuses } from '../useStatuses';
import { useFlowTeam } from '../../../shared/hooks/queries/useFlowTeam';
import type { ApiImportRowError } from '../../../shared/services/flowApi';
import {
  MAX_IMPORT_ROWS,
  SAMPLE_HEADERS,
  buildImportPreview,
  type ImportIssue,
  type ImportPreview,
} from './taskImport';

// The preview lists this many rows; the rest are imported all the same.
const PREVIEW_ROWS = 50;

function issueText(issue: ImportIssue, t: TFunction): string {
  return issue.kind === 'noTitle'
    ? t('flow.taskImport.issue.noTitle')
    : t(`flow.taskImport.issue.${issue.kind}`, { value: issue.value });
}

function downloadSample(t: TFunction): void {
  downloadCsv(
    'tasks-sample.csv',
    buildCsv([
      SAMPLE_HEADERS,
      [
        t('flow.taskImport.sample.title'),
        t('flow.taskImport.sample.description'),
        'name@company.com',
        '2026-12-01',
        'high',
        '2h',
        'To Do',
        t('flow.taskImport.sample.subtasks'),
      ],
    ])
  );
}

/**
 * Project List > Import (T-343): pick a CSV, check the preview, import. Every
 * problem is shown per row before anything is sent; the server checks again
 * and creates nothing if any row is refused.
 */
const ImportTasksDialog: React.FC<{
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}> = ({ projectId, open, onOpenChange }) => {
  const { t } = useTranslation();
  const fileId = useId();
  const { data: members = [] } = useProjectMembers(open ? projectId : undefined);
  const { data: team = [] } = useFlowTeam();
  const { data: statuses = [] } = useProjectStatuses(open ? projectId : '');
  const importTasks = useImportTasks();
  const [fileName, setFileName] = useState('');
  const [text, setText] = useState('');
  const [serverErrors, setServerErrors] = useState<ApiImportRowError[]>([]);

  // Project members first: their emails are the ones a project file names.
  // The business team covers people who are not in the project yet; assigning
  // them adds them, as it does from the task panel.
  const people = useMemo(() => {
    const seen = new Set<string>();
    return [
      ...members.map((m) => ({ id: String(m.userId), name: m.name, email: m.email })),
      ...team.map((m) => ({ id: m.id, name: m.name, email: m.email })),
    ].filter((p) => {
      const key = (p.email ?? '').toLowerCase();
      if (!key || seen.has(key)) return false;
      seen.add(key);
      return true;
    });
  }, [members, team]);

  const preview: ImportPreview | null = useMemo(
    () =>
      text
        ? buildImportPreview(text, {
            people,
            statuses: statuses.map((s) => ({ id: String(s.id), name: s.name })),
          })
        : null,
    [text, people, statuses]
  );
  const rowsWithIssues = preview?.rows.filter((r) => r.issues.length > 0).length ?? 0;
  const ready =
    preview !== null && !preview.missingTitle && preview.rows.length > 0 && rowsWithIssues === 0;

  const reset = () => {
    setFileName('');
    setText('');
    setServerErrors([]);
  };

  const readFile = async (file: File | undefined) => {
    setServerErrors([]);
    if (!file) return;
    setFileName(file.name);
    setText(await file.text());
  };

  const submit = () => {
    if (!preview || !ready) return;
    importTasks.mutate(
      {
        projectId,
        rows: preview.rows.map((row) => ({
          title: row.title,
          description: row.description,
          assigneeId: row.assigneeId,
          dueDate: row.dueDate,
          priority: row.priority,
          effortMinutes: row.effortMinutes,
          sectionId: row.sectionId,
          subtasks: row.subtasks.length > 0 ? row.subtasks : undefined,
        })),
      },
      {
        onSuccess: (res) => {
          const errors = res.errors ?? [];
          if (errors.length > 0) {
            setServerErrors(errors);
            return;
          }
          toast.success(t('flowToasts.task.imported', { count: res.created ?? 0 }));
          reset();
          onOpenChange(false);
        },
        onError: () => toast.error(t('flowToasts.task.importFailed')),
      }
    );
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) reset();
        onOpenChange(next);
      }}
    >
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t('flow.taskImport.title')}</DialogTitle>
          <DialogDescription>
            {t('flow.taskImport.description', { max: MAX_IMPORT_ROWS })}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 px-6">
          <div className="flex flex-wrap items-center gap-2">
            <label
              htmlFor={fileId}
              className="flow-focus inline-flex cursor-pointer items-center gap-1.5 rounded-sm px-3 py-1.5 text-flow-meta font-semibold"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              <FileUp size={12} aria-hidden="true" />
              {fileName ? t('flow.taskImport.chooseAnother') : t('flow.taskImport.choose')}
            </label>
            <input
              id={fileId}
              type="file"
              accept=".csv,text/csv"
              className="sr-only"
              onChange={(e) => {
                void readFile(e.target.files?.[0]);
                e.target.value = '';
              }}
            />
            {fileName && (
              <span className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
                {fileName}
              </span>
            )}
            <button
              type="button"
              onClick={() => downloadSample(t)}
              className="ms-auto inline-flex items-center gap-1 text-flow-meta hover:underline"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Download size={12} aria-hidden="true" />
              {t('flow.taskImport.sampleFile')}
            </button>
          </div>
          <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.taskImport.columns')}
          </p>

          {preview?.missingTitle && (
            <p role="alert" className="text-flow-meta" style={{ color: 'var(--flow-danger-fg)' }}>
              {t('flow.taskImport.noTitleColumn')}
            </p>
          )}
          {preview && !preview.missingTitle && (
            <>
              <p
                role="status"
                className="text-flow-meta"
                style={{
                  color: rowsWithIssues > 0 ? 'var(--flow-danger-fg)' : 'var(--flow-text-muted)',
                }}
              >
                {rowsWithIssues > 0
                  ? t('flow.taskImport.summaryIssues', {
                      count: preview.rows.length,
                      issues: rowsWithIssues,
                    })
                  : t('flow.taskImport.summary', { count: preview.rows.length })}
                {preview.truncated &&
                  ` ${t('flow.taskImport.truncated', { max: MAX_IMPORT_ROWS })}`}
              </p>
              <div
                className="max-h-[320px] overflow-auto rounded-md"
                style={{ border: '1px solid var(--flow-border)' }}
              >
                <table className="w-full text-flow-meta">
                  <thead
                    className="sticky top-0"
                    style={{ background: 'var(--flow-card)', color: 'var(--flow-text-muted)' }}
                  >
                    <tr>
                      <th className="px-2 py-1.5 text-start font-semibold">#</th>
                      <th className="px-2 py-1.5 text-start font-semibold">
                        {t('flow.taskImport.col.title')}
                      </th>
                      <th className="px-2 py-1.5 text-start font-semibold">
                        {t('flow.taskImport.col.assignee')}
                      </th>
                      <th className="px-2 py-1.5 text-start font-semibold">
                        {t('flow.taskImport.col.due')}
                      </th>
                      <th className="px-2 py-1.5 text-start font-semibold">
                        {t('flow.taskImport.col.section')}
                      </th>
                      <th className="px-2 py-1.5 text-start font-semibold">
                        {t('flow.taskImport.col.subtasks')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {preview.rows.slice(0, PREVIEW_ROWS).map((row) => (
                      <tr
                        key={row.line}
                        style={{
                          borderTop: '1px solid var(--flow-border)',
                          color: 'var(--flow-text)',
                        }}
                      >
                        <td
                          className="px-2 py-1.5 align-top"
                          style={{ color: 'var(--flow-text-dim)' }}
                        >
                          {row.line}
                        </td>
                        <td className="px-2 py-1.5 align-top" dir="auto">
                          {row.title}
                          {row.issues.map((issue, i) => (
                            <span
                              key={i}
                              className="block"
                              style={{ color: 'var(--flow-danger-fg)' }}
                            >
                              {issueText(issue, t)}
                            </span>
                          ))}
                        </td>
                        <td className="px-2 py-1.5 align-top">{row.assigneeLabel ?? ''}</td>
                        <td className="px-2 py-1.5 align-top">{row.dueDate ?? ''}</td>
                        <td className="px-2 py-1.5 align-top">{row.sectionLabel ?? ''}</td>
                        <td className="px-2 py-1.5 align-top">
                          {row.subtasks.length > 0 ? row.subtasks.length : ''}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {preview.rows.length > PREVIEW_ROWS && (
                <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.taskImport.more', { count: preview.rows.length - PREVIEW_ROWS })}
                </p>
              )}
            </>
          )}
          {serverErrors.length > 0 && (
            <ul
              role="alert"
              className="space-y-0.5 text-flow-meta"
              style={{ color: 'var(--flow-danger-fg)' }}
            >
              <li>{t('flow.taskImport.refused')}</li>
              {serverErrors.map((e) => (
                <li key={`${e.row}-${e.message}`}>
                  {t('flow.taskImport.rowError', { row: e.row ?? 0, message: e.message ?? '' })}
                </li>
              ))}
            </ul>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('flowCommon.cancel')}
          </Button>
          <Button disabled={!ready || importTasks.isPending} onClick={submit}>
            {importTasks.isPending
              ? t('flow.taskImport.importing')
              : t('flow.taskImport.import', { count: preview?.rows.length ?? 0 })}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ImportTasksDialog;
