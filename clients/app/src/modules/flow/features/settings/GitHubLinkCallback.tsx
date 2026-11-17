import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useCompleteGitHubLink } from '../../../../shared/hooks/queries/useGitHub';

/**
 * /flow/github/callback, the GitHub App's user authorization Callback URL.
 * GitHub returns here with code and state (or error when the member cancels).
 * The code is handed to the server once and stripped from the URL, then the
 * member lands on Settings > Flow > Connected accounts.
 */
const GitHubLinkCallback = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const complete = useCompleteGitHubLink();
  const handled = useRef(false);
  const completeLink = complete.mutate;

  useEffect(() => {
    if (handled.current) return;
    handled.current = true;
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    void navigate({ to: '/flow/github/callback', replace: true });
    // Leave the callback first so the settings modal opens over Flow rather
    // than over this page.
    const finish = async () => {
      await navigate({ to: '/flow', replace: true });
      await navigate({ to: '/settings/flow', search: { section: 'accounts' } });
    };
    if (!code || !state) {
      toast.error(
        t(
          params.get('error') === 'access_denied'
            ? 'flow.settings.accounts.github.cancelled'
            : 'flow.settings.accounts.github.linkError'
        )
      );
      void finish();
      return;
    }
    completeLink(
      { code, state },
      {
        onSuccess: () => toast.success(t('flow.settings.accounts.github.linked')),
        onError: (error) =>
          toast.error(getOneKitErrorMessage(error, t('flow.settings.accounts.github.linkError'))),
        onSettled: () => void finish(),
      }
    );
  }, [completeLink, navigate, t]);

  return (
    <p role="status" className="p-8 text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
      {t('flow.settings.accounts.github.linking')}
    </p>
  );
};

export default GitHubLinkCallback;
