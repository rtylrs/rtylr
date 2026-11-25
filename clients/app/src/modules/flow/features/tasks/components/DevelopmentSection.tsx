import React from 'react';
import { useTranslation } from 'react-i18next';
import { Code2, Copy } from 'lucide-react';
import { toast } from 'sonner';
import { useTaskLinks } from '../useTaskLinks';

// Link kinds that describe code, builds or runtime errors for this task.
const CODE_LINK_KINDS = new Set([
  'github_pr',
  'github_issue',
  'gitlab_mr',
  'gitlab_issue',
  'vercel',
  'netlify',
  'sentry',
]);

/**
 * Development section of the task panel, shown only on developer projects.
 * Lists the task's code links; the dev fields are filled in by the
 * dev-planning work that follows.
 */
const DevelopmentSection = ({
  taskId,
  typeLabel,
  branchName,
}: {
  taskId: string;
  /** The task's type, translated. */
  typeLabel?: string;
  /** A suggested git branch for the task, prefixed by its type (T-380). */
  branchName?: string;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: links = [] } = useTaskLinks(taskId);
  const codeLinks = links.filter((link) => CODE_LINK_KINDS.has(link.kind ?? ''));
  const fields = [
    { key: 'type', label: t('flow.taskDetail.devType'), value: typeLabel },
    { key: 'points', label: t('flow.taskDetail.devPoints'), value: undefined },
    { key: 'branch', label: t('flow.taskDetail.devBranch'), value: branchName },
  ];
  return (
    <section className="mb-8" aria-labelledby="flow-task-development-heading">
      <div className="flex items-center gap-2 mb-3">
        <Code2 size={12} aria-hidden style={{ color: 'var(--flow-text-muted)' }} />
        <h3
          id="flow-task-development-heading"
          className="text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.taskDetail.development')}
        </h3>
        <span
          className="text-flow-meta px-1.5 rounded-sm"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {t('flowCommon.comingSoon')}
        </span>
      </div>
      <dl className="grid grid-cols-3 gap-2 mb-3">
        {fields.map((field) => (
          <div
            key={field.key}
            className="rounded-md px-3 py-2"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
            }}
          >
            <dt className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {field.label}
            </dt>
            <dd
              className="flex min-w-0 items-center gap-1 text-flow-meta"
              style={{ color: field.value ? 'var(--flow-text)' : 'var(--flow-text-dim)' }}
            >
              <span className={field.key === 'branch' ? 'truncate font-mono' : undefined} dir="ltr">
                {field.value ?? t('flow.taskDetail.devNotSet')}
              </span>
              {field.key === 'branch' && field.value && (
                <button
                  type="button"
                  aria-label={t('flow.taskDetail.copyBranch')}
                  title={t('flow.taskDetail.copyBranch')}
                  onClick={async () => {
                    try {
                      await navigator.clipboard.writeText(field.value ?? '');
                      toast.success(t('flow.taskDetail.branchCopied'));
                    } catch {
                      toast.error(t('flowErrors.generic'));
                    }
                  }}
                  className="flow-focus shrink-0 rounded-sm p-0.5 hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  <Copy size={11} aria-hidden="true" />
                </button>
              )}
            </dd>
          </div>
        ))}
      </dl>
      {codeLinks.length > 0 ? (
        <ul className="space-y-1">
          {codeLinks.map((link) => (
            <li key={link.id} className="flex items-baseline gap-2 text-flow-meta">
              <a
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className="min-w-0 truncate font-medium hover:underline"
                style={{ color: 'var(--flow-text)' }}
              >
                {link.title || link.ref || link.url}
              </a>
              {link.title && link.ref && (
                <span className="shrink-0" style={{ color: 'var(--flow-text-dim)' }}>
                  {link.ref}
                </span>
              )}
              {link.state && (
                <span className="shrink-0" style={{ color: 'var(--flow-text-dim)' }}>
                  {t(`flow.taskDetail.linkState.${link.state}`, { defaultValue: link.state })}
                </span>
              )}
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.taskDetail.developmentEmpty')}
        </p>
      )}
    </section>
  );
};

export default DevelopmentSection;
