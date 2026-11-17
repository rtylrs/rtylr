import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { GitBranch } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../shared/components/ui';
import {
  useCompleteGitHubInstall,
  useDisconnectGitHub,
  useGitHubConnection,
  useStartGitHubInstall,
} from '../../shared/hooks/queries/useGitHub';

/**
 * Dash > Extensions card for the business's GitHub App installation. GitHub
 * sends the admin back to /extensions with installation_id and state, which
 * this card hands to the server once and then strips from the URL.
 */
export default function GitHubConnection() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const connection = useGitHubConnection();
  const start = useStartGitHubInstall();
  const complete = useCompleteGitHubInstall();
  const disconnect = useDisconnectGitHub();
  const [confirming, setConfirming] = useState(false);
  const handledReturn = useRef(false);
  const completeInstall = complete.mutate;
  const refetchConnection = connection.refetch;

  useEffect(() => {
    if (handledReturn.current) return;
    const params = new URLSearchParams(window.location.search);
    const installationId = params.get('installation_id');
    if (!installationId) return;
    handledReturn.current = true;
    const state = params.get('state');
    void navigate({ to: '/extensions', replace: true });
    // A change made from GitHub's own settings page comes back without a
    // state; there is nothing to complete, only the health to re-read.
    if (!state) {
      void refetchConnection();
      return;
    }
    completeInstall(
      { installationId, state },
      {
        onSuccess: () => toast.success(t('extensions.github.connected')),
        onError: (error) =>
          toast.error(getOneKitErrorMessage(error, t('extensions.github.connectError'))),
      }
    );
  }, [completeInstall, navigate, refetchConnection, t]);

  const data = connection.data;
  const busy = start.isPending || complete.isPending || disconnect.isPending;

  const connect = () =>
    start.mutate(undefined, {
      onSuccess: (url) => {
        if (url) window.location.assign(url);
      },
      onError: (error) =>
        toast.error(getOneKitErrorMessage(error, t('extensions.github.connectError'))),
    });

  const runDisconnect = () =>
    disconnect.mutate(undefined, {
      onSuccess: (uninstalled) => {
        setConfirming(false);
        if (uninstalled) toast.success(t('extensions.github.disconnected'));
        else toast.warning(t('extensions.github.disconnectedManual'));
      },
      onError: () => toast.error(t('extensions.github.disconnectError')),
    });

  let body: string;
  if (connection.isLoading || complete.isPending) body = t('extensions.github.checking');
  else if (connection.isError) body = t('extensions.github.loadError');
  else if (!data?.configured) body = t('extensions.github.notConfigured');
  else if (data.connected)
    body = t('extensions.github.connectedTo', {
      account: data.accountLogin,
      type: t(
        data.accountType === 'User'
          ? 'extensions.github.accountUser'
          : 'extensions.github.accountOrganization'
      ),
    });
  else body = t('extensions.github.description');

  return (
    <section className="portal-extensions-github" aria-labelledby="portal-extensions-github-title">
      <article className="portal-extensions-item">
        <div className="portal-extensions-item-top">
          <span className="portal-extensions-icon">
            <GitBranch size={23} strokeWidth={1.5} aria-hidden />
          </span>
          <span className="portal-extensions-category">{t('extensions.github.category')}</span>
          {data?.connected && (
            <span className="portal-extensions-status" data-health={data.status}>
              {t(`extensions.github.health.${data.status}`)}
            </span>
          )}
        </div>
        <h2 id="portal-extensions-github-title">GitHub</h2>
        <p role={connection.isError ? 'alert' : undefined}>{body}</p>
        <div className="portal-extensions-item-footer">
          <span className="portal-extensions-price">{t('extensions.github.scope')}</span>
          <div className="portal-extensions-item-actions">
            {connection.isError && (
              <Button size="sm" variant="ghost" onClick={() => void connection.refetch()}>
                {t('extensions.retry')}
              </Button>
            )}
            {data?.connected ? (
              <Button
                size="sm"
                variant="secondary"
                disabled={!data.canManage || busy}
                onClick={() => setConfirming(true)}
              >
                {t('extensions.github.disconnect')}
              </Button>
            ) : (
              <Button
                size="sm"
                variant="secondary"
                disabled={!data?.configured || !data.canManage || busy}
                onClick={connect}
              >
                {start.isPending ? t('extensions.working') : t('extensions.github.connect')}
              </Button>
            )}
          </div>
        </div>
      </article>
      <Dialog open={confirming} onOpenChange={(open) => !busy && setConfirming(open)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('extensions.github.disconnectTitle')}</DialogTitle>
            <DialogDescription>{t('extensions.github.disconnectDescription')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" disabled={busy} onClick={() => setConfirming(false)}>
              {t('extensions.cancel')}
            </Button>
            <Button disabled={busy} onClick={runDisconnect}>
              {disconnect.isPending ? t('extensions.working') : t('extensions.github.disconnect')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}
