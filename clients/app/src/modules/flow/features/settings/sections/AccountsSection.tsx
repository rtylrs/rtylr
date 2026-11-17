import React from 'react';
import { useTranslation } from 'react-i18next';
import { GitBranch } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import {
  useGitHubConnection,
  useGitHubLink,
  useStartGitHubLink,
  useUnlinkGitHub,
} from '../../../../../shared/hooks/queries/useGitHub';

const muted = { color: 'var(--flow-text-muted)' };

/**
 * Settings > Flow > Connected accounts: the member's own GitHub account, so
 * commits and pull requests can be matched to them. The GitHub row shows only
 * once the business has connected GitHub in Extensions.
 */
const AccountsSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const connection = useGitHubConnection();
  const connected = !!connection.data?.configured && !!connection.data.connected;

  let content: React.ReactNode;
  if (connection.isLoading) {
    content = (
      <p role="status" className="text-flow-meta" style={muted}>
        {t('flow.settings.accounts.github.checking')}
      </p>
    );
  } else if (connection.isError) {
    content = <ListLoadFailed onRetry={() => void connection.refetch()} />;
  } else if (!connected) {
    content = (
      <p className="text-flow-meta" style={muted}>
        {t('flow.settings.accounts.notConnected')}
      </p>
    );
  } else {
    content = <GitHubAccountRow linkingEnabled={!!connection.data?.linkingEnabled} />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.settings.accounts.section')}
        </h2>
        <p className="text-flow-meta" style={muted}>
          {t('flow.settings.accounts.subtitle')}
        </p>
      </div>
      {content}
    </div>
  );
};

const GitHubAccountRow = ({ linkingEnabled }: { linkingEnabled: boolean }): React.JSX.Element => {
  const { t } = useTranslation();
  const link = useGitHubLink(true);
  const start = useStartGitHubLink();
  const unlink = useUnlinkGitHub();
  const busy = start.isPending || unlink.isPending;
  const linked = !!link.data?.linked;

  const startLink = () =>
    start.mutate(undefined, {
      onSuccess: (url) => {
        if (url) window.location.assign(url);
      },
      onError: (error) =>
        toast.error(getOneKitErrorMessage(error, t('flow.settings.accounts.github.linkError'))),
    });

  const removeLink = () =>
    unlink.mutate(undefined, {
      onSuccess: () => toast.success(t('flow.settings.accounts.github.unlinked')),
      onError: () => toast.error(t('flow.settings.accounts.github.unlinkError')),
    });

  let detail: string;
  if (link.isLoading) detail = t('flow.settings.accounts.github.checking');
  else if (linked)
    detail = t('flow.settings.accounts.github.linkedAs', { login: link.data?.login });
  else if (!linkingEnabled) detail = t('flow.settings.accounts.github.linkingOff');
  else detail = t('flow.settings.accounts.github.notLinked');

  if (link.isError) return <ListLoadFailed onRetry={() => void link.refetch()} />;

  return (
    <div
      className="flex flex-wrap items-center justify-between gap-3 rounded-xl px-4 py-3"
      style={{ border: '1px solid var(--flow-border)', background: 'var(--flow-card)' }}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <GitBranch size={18} aria-hidden="true" style={muted} />
        <div className="min-w-0">
          <p className="text-flow-body font-medium" style={{ color: 'var(--flow-text)' }}>
            {t('flow.settings.accounts.github.label')}
          </p>
          <p className="text-flow-meta mt-0.5 break-all" style={muted}>
            {detail}
          </p>
        </div>
      </div>
      {linked ? (
        <button
          type="button"
          onClick={removeLink}
          disabled={busy}
          className="shrink-0 rounded-lg px-3 py-1.5 text-flow-meta font-semibold disabled:opacity-40"
          style={{ border: '1px solid var(--flow-border-control)', color: 'var(--flow-text)' }}
        >
          {unlink.isPending
            ? t('flow.settings.accounts.github.working')
            : t('flow.settings.accounts.github.unlink')}
        </button>
      ) : (
        <button
          type="button"
          onClick={startLink}
          disabled={busy || link.isLoading || !linkingEnabled}
          className="shrink-0 rounded-lg px-3 py-1.5 text-flow-meta font-semibold disabled:opacity-40"
          style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
        >
          {start.isPending
            ? t('flow.settings.accounts.github.working')
            : t('flow.settings.accounts.github.link')}
        </button>
      )}
    </div>
  );
};

export default AccountsSection;
