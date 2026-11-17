import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { toast } from 'sonner';
import { GitBranch, Lock } from 'lucide-react';
import GateNotice from '../../../shared/components/GateNotice';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import {
  useGitHubConnection,
  useInstallationRepositories,
  useProjectRepositories,
  useSetProjectRepositories,
  type GitHubRepository,
} from '../../../../../shared/hooks/queries/useGitHub';

const GATE_ID = 'flow-project-repositories-gate';
const FILTER_THRESHOLD = 8;

interface ProjectRepositoriesPanelProps {
  projectId: string;
  /** Whether the caller holds the project-admin gate the server requires. */
  canAdminister: boolean;
}

const muted = { color: 'var(--flow-text-muted)' };
const card = { background: 'var(--flow-input-bg)', border: '1px solid var(--flow-border-control)' };

/**
 * Project Settings > Integrations > Code: the GitHub repositories this project
 * uses, picked from the business's GitHub App installation.
 */
const ProjectRepositoriesPanel = ({
  projectId,
  canAdminister,
}: ProjectRepositoriesPanelProps): React.JSX.Element => {
  const { t } = useTranslation();
  const connection = useGitHubConnection();
  const connected = !!connection.data?.configured && !!connection.data.connected;
  const selected = useProjectRepositories(projectId, connected);
  const available = useInstallationRepositories(projectId, connected && canAdminister);
  const save = useSetProjectRepositories(projectId);
  const [draft, setDraft] = React.useState<Set<string> | null>(null);
  const [filter, setFilter] = React.useState('');

  const savedIds = React.useMemo(
    () => new Set((selected.data ?? []).map((r) => r.id)),
    [selected.data]
  );
  const picked = draft ?? savedIds;
  const dirty =
    draft !== null && (draft.size !== savedIds.size || [...draft].some((id) => !savedIds.has(id)));

  // Picked repositories the installation no longer lists stay visible, so an
  // admin can still remove them.
  const options = React.useMemo(() => {
    const byId = new Map<string, GitHubRepository>();
    for (const r of available.data ?? []) byId.set(r.id, r);
    for (const r of selected.data ?? []) if (!byId.has(r.id)) byId.set(r.id, r);
    return [...byId.values()].sort((a, b) => a.fullName.localeCompare(b.fullName));
  }, [available.data, selected.data]);
  const query = filter.trim().toLocaleLowerCase();
  const visible = query
    ? options.filter((r) => r.fullName.toLocaleLowerCase().includes(query))
    : options;

  const toggle = (id: string) => {
    const next = new Set(picked);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setDraft(next);
  };

  const submit = () =>
    save.mutate([...picked], {
      onSuccess: () => {
        setDraft(null);
        toast.success(t('flow.projectSettings.code.saved'));
      },
      onError: () => toast.error(t('flow.projectSettings.code.saveError')),
    });

  let content: React.ReactNode;
  if (connection.isLoading) {
    content = (
      <p role="status" className="text-flow-meta" style={muted}>
        {t('flow.projectSettings.code.checking')}
      </p>
    );
  } else if (connection.isError) {
    content = <ListLoadFailed onRetry={() => void connection.refetch()} />;
  } else if (!connected) {
    content = (
      <div className="rounded-md px-3 py-2.5 space-y-1" style={card}>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text)' }}>
          {t('flow.projectSettings.code.notConnected')}
        </p>
        <Link
          to="/extensions"
          className="text-flow-meta underline underline-offset-2"
          style={{ color: 'var(--accent)' }}
        >
          {t('flow.projectSettings.code.connectLink')}
        </Link>
      </div>
    );
  } else {
    const health = connection.data?.status ?? 'unknown';
    content = (
      <div className="space-y-2">
        <p className="text-flow-meta break-words" style={muted}>
          {t('flow.projectSettings.code.connectedTo', {
            account: connection.data?.accountLogin,
          })}{' '}
          <span
            data-health={health}
            style={{ color: health === 'active' ? 'var(--flow-text)' : 'var(--flow-warn-fg)' }}
          >
            {t(`flow.projectSettings.code.health.${health}`)}
          </span>
        </p>
        <GateNotice
          id={GATE_ID}
          reason={canAdminister ? '' : t('flow.projectSettings.code.gateReason')}
        />
        {canAdminister ? (
          <RepositoryPicker
            loading={available.isLoading || selected.isLoading}
            failed={available.isError || selected.isError}
            onRetry={() => {
              void available.refetch();
              void selected.refetch();
            }}
            options={options}
            visible={visible}
            picked={picked}
            onToggle={toggle}
            filter={filter}
            onFilter={setFilter}
            disabled={save.isPending}
          />
        ) : (
          <SelectedList
            loading={selected.isLoading}
            failed={selected.isError}
            onRetry={() => void selected.refetch()}
            repositories={selected.data ?? []}
          />
        )}
        {canAdminister && (
          <button
            type="button"
            className="px-3 py-1.5 rounded-sm text-flow-meta font-semibold disabled:opacity-40"
            style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            disabled={!dirty || save.isPending}
            onClick={submit}
          >
            {save.isPending
              ? t('flow.projectSettings.code.saving')
              : t('flow.projectSettings.code.save')}
          </button>
        )}
      </div>
    );
  }

  return (
    <section aria-labelledby="flow-project-code-heading" className="space-y-2">
      <h2
        id="flow-project-code-heading"
        className="flex items-center gap-1.5 text-flow-meta font-medium"
        style={{ color: 'var(--flow-text)' }}
      >
        <GitBranch size={14} aria-hidden="true" style={muted} />
        {t('flow.projectSettings.code.heading')}
      </h2>
      <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.projectSettings.code.description')}
      </p>
      {content}
    </section>
  );
};

interface RepositoryPickerProps {
  loading: boolean;
  failed: boolean;
  onRetry: () => void;
  options: GitHubRepository[];
  visible: GitHubRepository[];
  picked: Set<string>;
  onToggle: (id: string) => void;
  filter: string;
  onFilter: (value: string) => void;
  disabled: boolean;
}

const RepositoryPicker = ({
  loading,
  failed,
  onRetry,
  options,
  visible,
  picked,
  onToggle,
  filter,
  onFilter,
  disabled,
}: RepositoryPickerProps): React.JSX.Element => {
  const { t } = useTranslation();
  if (loading) {
    return (
      <p role="status" className="text-flow-meta" style={muted}>
        {t('flow.projectSettings.code.loadingRepositories')}
      </p>
    );
  }
  if (failed) return <ListLoadFailed onRetry={onRetry} />;
  if (options.length === 0) {
    return (
      <p className="text-flow-meta" style={muted}>
        {t('flow.projectSettings.code.noneAvailable')}
      </p>
    );
  }
  return (
    <fieldset className="space-y-2 min-w-0" disabled={disabled}>
      <legend className="sr-only">{t('flow.projectSettings.code.pickerLabel')}</legend>
      {options.length > FILTER_THRESHOLD && (
        <input
          type="search"
          value={filter}
          onChange={(event) => onFilter(event.target.value)}
          placeholder={t('flow.projectSettings.code.filter')}
          aria-label={t('flow.projectSettings.code.filter')}
          className="w-full rounded-md px-3 py-1.5 text-flow-meta"
          style={{ ...card, color: 'var(--flow-text)' }}
        />
      )}
      <ul className="space-y-1 max-h-80 overflow-y-auto">
        {visible.map((repo) => (
          <li key={repo.id}>
            <label
              className="flex min-h-[40px] cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-flow-meta min-w-0"
              style={{ ...card, color: 'var(--flow-text)' }}
            >
              <input
                type="checkbox"
                checked={picked.has(repo.id)}
                onChange={() => onToggle(repo.id)}
              />
              <span className="min-w-0 flex-1 break-all">{repo.fullName}</span>
              {repo.private && (
                <span className="flex shrink-0 items-center gap-1" style={muted}>
                  <Lock size={11} aria-hidden="true" />
                  {t('flow.projectSettings.code.private')}
                </span>
              )}
            </label>
          </li>
        ))}
      </ul>
      {visible.length === 0 && (
        <p className="text-flow-meta" style={muted}>
          {t('flow.projectSettings.code.noMatches')}
        </p>
      )}
    </fieldset>
  );
};

interface SelectedListProps {
  loading: boolean;
  failed: boolean;
  onRetry: () => void;
  repositories: GitHubRepository[];
}

const SelectedList = ({
  loading,
  failed,
  onRetry,
  repositories,
}: SelectedListProps): React.JSX.Element => {
  const { t } = useTranslation();
  if (loading) {
    return (
      <p role="status" className="text-flow-meta" style={muted}>
        {t('flow.projectSettings.code.loadingRepositories')}
      </p>
    );
  }
  if (failed) return <ListLoadFailed onRetry={onRetry} />;
  if (repositories.length === 0) {
    return (
      <p className="text-flow-meta" style={muted}>
        {t('flow.projectSettings.code.none')}
      </p>
    );
  }
  return (
    <ul className="space-y-1" aria-label={t('flow.projectSettings.code.pickerLabel')}>
      {repositories.map((repo) => (
        <li
          key={repo.id}
          className="flex items-center gap-2 rounded-md px-3 py-2 text-flow-meta min-w-0"
          style={{ ...card, color: 'var(--flow-text)' }}
        >
          <span className="min-w-0 flex-1 break-all">{repo.fullName}</span>
          {repo.private && (
            <span className="flex shrink-0 items-center gap-1" style={muted}>
              <Lock size={11} aria-hidden="true" />
              {t('flow.projectSettings.code.private')}
            </span>
          )}
        </li>
      ))}
    </ul>
  );
};

export default ProjectRepositoriesPanel;
