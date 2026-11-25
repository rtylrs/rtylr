import React from 'react';
import { useTranslation } from 'react-i18next';
import { Code2, CheckCircle2, Circle, Loader2, MinusCircle, XCircle } from 'lucide-react';
import { useTaskLinks } from '../useTaskLinks';
import { useTaskPullRequests, type TaskPullRequest } from '../useTaskPullRequests';
import ChecksDot from '../../../shared/components/ChecksDot';

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
const DevelopmentSection = ({ taskId }: { taskId: string }): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: links = [] } = useTaskLinks(taskId);
  const codeLinks = links.filter((link) => CODE_LINK_KINDS.has(link.kind ?? ''));
  const hasPullRequests = codeLinks.some((link) => link.kind === 'github_pr');
  const { data: pullRequests = [] } = useTaskPullRequests(taskId, hasPullRequests);
  const pullRequestByRef = new Map(pullRequests.map((pr) => [pr.ref ?? '', pr]));
  const fields = [
    { key: 'type', label: t('flow.taskDetail.devType') },
    { key: 'points', label: t('flow.taskDetail.devPoints') },
    { key: 'branch', label: t('flow.taskDetail.devBranch') },
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
            <dd className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.taskDetail.devNotSet')}
            </dd>
          </div>
        ))}
      </dl>
      {codeLinks.length > 0 ? (
        <ul className="space-y-1">
          {codeLinks.map((link) => (
            <li key={link.id} className="flex flex-wrap items-baseline gap-x-2 text-flow-meta">
              <a
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className="min-w-0 truncate font-medium hover:underline"
                style={{ color: 'var(--flow-text)' }}
              >
                {link.title || pullRequestByRef.get(link.ref ?? '')?.title || link.ref || link.url}
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
              {link.kind === 'github_pr' && (
                <PullRequestChecks pr={pullRequestByRef.get(link.ref ?? '')} />
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

type CheckLook = { icon: typeof Circle; color: string; key: string };

function checkLook(status?: string, conclusion?: string): CheckLook {
  if (status !== 'completed') {
    return { icon: Loader2, color: 'var(--flow-warn-fg)', key: 'running' };
  }
  switch (conclusion) {
    case 'success':
      return { icon: CheckCircle2, color: 'var(--color-success)', key: 'passed' };
    case 'neutral':
    case 'skipped':
    case 'stale':
      return { icon: MinusCircle, color: 'var(--flow-text-dim)', key: 'skipped' };
    default:
      return { icon: XCircle, color: 'var(--flow-danger-fg)', key: 'failed' };
  }
}

/**
 * The CI checks on a linked pull request's head commit (T-376): overall state,
 * then each check with a link to its log.
 */
const PullRequestChecks = ({ pr }: { pr?: TaskPullRequest }): React.JSX.Element | null => {
  const { t } = useTranslation();
  if (!pr) return null;
  if (!pr.known) {
    return (
      <p className="w-full ps-4 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.checks.waiting')}
      </p>
    );
  }
  const checks = pr.checks ?? [];
  return (
    <div className="w-full ps-4">
      <p
        className="flex items-center gap-1.5 text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <ChecksDot state={pr.checksState} />
        {t(`flow.checks.summary.${pr.checksState || 'none'}`, { count: checks.length })}
      </p>
      {checks.length > 0 && (
        <ul className="mt-0.5 space-y-0.5">
          {checks.map((check) => {
            const look = checkLook(check.status, check.conclusion);
            const Icon = look.icon;
            return (
              <li key={check.name} className="flex items-center gap-1.5 text-flow-meta">
                <Icon size={11} aria-hidden="true" style={{ color: look.color }} />
                <span className="sr-only">{t(`flow.checks.check.${look.key}`)}</span>
                {check.detailsUrl ? (
                  <a
                    href={check.detailsUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="truncate hover:underline"
                    style={{ color: 'var(--flow-text)' }}
                    title={t('flow.checks.openLog')}
                  >
                    {check.name}
                  </a>
                ) : (
                  <span className="truncate" style={{ color: 'var(--flow-text)' }}>
                    {check.name}
                  </span>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};

export default DevelopmentSection;
