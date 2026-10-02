import React from 'react';
import { Navigate, Outlet } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useGetHRConfig } from '../../shared/hooks/queries/useSetup';
import { isHRSetupComplete } from '../../shared/lib/setupCompletion';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';

function PageLoader(): React.JSX.Element {
  return (
    <div
      className="flex min-h-48 items-center justify-center"
      style={{ background: 'var(--hr-bg)' }}
    >
      <div
        className="h-8 w-8 rounded-full border-2 animate-spin"
        style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
      />
    </div>
  );
}

function ConfigLoadError({
  onRetry,
  isRetrying,
}: {
  onRetry: () => void;
  isRetrying: boolean;
}): React.JSX.Element {
  const { t } = useTranslation('hrCommon');

  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20" role="alert">
      <p className="text-sm" style={{ color: 'var(--hr-text)' }}>
        {t('hrSetup.statusError')}
      </p>
      <button
        type="button"
        className="min-h-[44px] rounded-lg px-4 py-2 text-sm font-medium"
        style={{ backgroundColor: 'var(--accent)', color: 'var(--hr-on-accent)' }}
        disabled={isRetrying}
        onClick={onRetry}
      >
        {t('hrCommon.retry')}
      </button>
    </div>
  );
}

const HRConfigRoute = (): React.JSX.Element => {
  const { data: config, isLoading, isError, isFetching, refetch } = useGetHRConfig();
  const setConfig = useHRConfigStore((s) => s.setConfig);
  const storedConfig = useHRConfigStore((s) => s.config);

  React.useEffect(() => {
    if (config) {
      setConfig(config);
    }
  }, [config, setConfig]);

  if (isLoading) {
    return <PageLoader />;
  }

  if (isError) {
    return <ConfigLoadError onRetry={() => refetch()} isRetrying={isFetching} />;
  }

  // A missing config and an unfinalised one both mean "run the wizard", and the
  // shared predicate already folds the null case in.
  if (!isHRSetupComplete(config)) {
    return <Navigate to="/hr/setup" replace />;
  }

  if (storedConfig !== config) return <PageLoader />;
  return <Outlet />;
};

export default HRConfigRoute;
