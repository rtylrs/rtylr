import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Bug,
  CircleDot,
  FileText,
  GitMerge,
  GitPullRequest,
  Link2,
  PenTool,
  Rocket,
  X,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { isWebUrl, useAddTaskLink, useRemoveTaskLink, useTaskLinks } from '../useTaskLinks';
import type { TaskLink } from '../useTaskLinks';

const KIND_ICON: Record<string, LucideIcon> = {
  github_pr: GitPullRequest,
  gitlab_mr: GitPullRequest,
  github_issue: CircleDot,
  gitlab_issue: CircleDot,
  figma: PenTool,
  vercel: Rocket,
  netlify: Rocket,
  sentry: Bug,
  google_docs: FileText,
};

const STATE_COLOR: Record<string, string> = {
  open: 'var(--flow-text)',
  draft: 'var(--flow-status-draft)',
  merged: 'var(--accent-on-muted)',
  closed: 'var(--flow-text-dim)',
};

interface LinksSectionProps {
  taskId: string;
  canEdit: boolean;
  peopleNames: Record<string, string>;
}

/** Links pasted onto a task. Works in every project, with or without an integration. */
const LinksSection = ({ taskId, canEdit, peopleNames }: LinksSectionProps): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const { data: links = [] } = useTaskLinks(taskId);
  const addLink = useAddTaskLink();
  const removeLink = useRemoveTaskLink();
  const [draft, setDraft] = useState('');
  const [invalid, setInvalid] = useState(false);

  const submit = () => {
    const url = draft.trim();
    if (!url || addLink.isPending) return;
    if (!isWebUrl(url)) {
      setInvalid(true);
      return;
    }
    addLink.mutate({ taskId, url }, { onSuccess: () => setDraft('') });
  };

  const label = (link: TaskLink) => link.title || (link.kind === 'generic' ? link.url : link.ref);

  return (
    <section className="mb-8" aria-labelledby="flow-task-links-heading">
      <h3
        id="flow-task-links-heading"
        className="text-flow-meta font-semibold uppercase tracking-wider mb-3"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {t('flow.taskDetail.links')}
      </h3>

      {links.length > 0 && (
        <ul className="space-y-1.5 mb-2">
          {links.map((link) => {
            const Icon = KIND_ICON[link.kind ?? ''] ?? Link2;
            const StateIcon = link.state === 'merged' ? GitMerge : null;
            const adder =
              peopleNames[String(link.addedBy)] ?? t('flow.taskDetail.linkAddedBySomeone');
            const when = link.createdAt
              ? new Date(link.createdAt).toLocaleDateString(i18n.language, {
                  day: 'numeric',
                  month: 'short',
                })
              : '';
            return (
              <li
                key={link.id}
                className="group flex items-start gap-2.5 rounded-md px-3 py-2"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                }}
              >
                <Icon
                  size={14}
                  aria-hidden
                  className="mt-0.5 shrink-0"
                  style={{ color: 'var(--flow-text-muted)' }}
                />
                <div className="min-w-0 flex-1">
                  <a
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block truncate text-flow-meta font-medium hover:underline"
                    style={{ color: 'var(--flow-text)' }}
                    title={link.url}
                  >
                    {label(link)}
                  </a>
                  <p
                    className="flex flex-wrap items-center gap-x-1.5 text-flow-meta"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {link.title && link.ref && <span className="truncate">{link.ref}</span>}
                    {link.state && (
                      <span
                        className="inline-flex items-center gap-1 rounded-sm px-1"
                        style={{ background: 'var(--flow-hover)', color: STATE_COLOR[link.state] }}
                      >
                        {StateIcon && <StateIcon size={10} aria-hidden />}
                        {t(`flow.taskDetail.linkState.${link.state}`, { defaultValue: link.state })}
                      </span>
                    )}
                    <span>{t('flow.taskDetail.linkAddedBy', { name: adder, date: when })}</span>
                  </p>
                </div>
                {canEdit && (
                  <button
                    type="button"
                    onClick={() => removeLink.mutate({ taskId, linkId: String(link.id) })}
                    disabled={removeLink.isPending}
                    aria-label={t('flow.taskDetail.removeLink', { name: label(link) })}
                    className="rounded-sm p-1 opacity-60 hover:opacity-100 hover:bg-[var(--flow-hover)] focus-visible:opacity-100"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <X size={12} aria-hidden />
                  </button>
                )}
              </li>
            );
          })}
        </ul>
      )}

      {canEdit && (
        <form
          className="flex gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            submit();
          }}
        >
          <input
            type="text"
            inputMode="url"
            autoComplete="off"
            value={draft}
            onChange={(event) => {
              setDraft(event.target.value);
              setInvalid(false);
            }}
            placeholder={t('flow.taskDetail.addLinkPlaceholder')}
            aria-label={t('flow.taskDetail.addLink')}
            aria-invalid={invalid || undefined}
            aria-describedby={invalid ? 'flow-task-link-error' : undefined}
            className="min-w-0 flex-1 rounded-md px-3 py-1.5 text-flow-meta"
            style={{
              background: 'var(--flow-input-bg)',
              border: `1px solid ${invalid ? 'var(--flow-danger-fg)' : 'var(--flow-border-control)'}`,
              color: 'var(--flow-text)',
            }}
          />
          <button
            type="submit"
            disabled={!draft.trim() || addLink.isPending}
            className="rounded-md px-3 py-1.5 text-flow-meta font-medium disabled:opacity-50"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
          >
            {addLink.isPending ? t('flow.taskDetail.addingLink') : t('flow.taskDetail.addLink')}
          </button>
        </form>
      )}
      {invalid && (
        <p
          id="flow-task-link-error"
          className="mt-1 text-flow-meta"
          style={{ color: 'var(--flow-danger-fg)' }}
        >
          {t('flow.taskDetail.invalidLink')}
        </p>
      )}
      {!canEdit && links.length === 0 && (
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.taskDetail.noLinks')}
        </p>
      )}
    </section>
  );
};

export default LinksSection;
