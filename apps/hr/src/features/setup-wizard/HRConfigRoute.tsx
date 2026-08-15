import React from 'react';
import { Navigate, Outlet } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useGetHRConfig, useGetSetupStatus } from '../../shared/hooks/queries/useSetup';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';

function PageLoader(): React.JSX.Element {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
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
  const { t } = useTranslation();

  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20" role="alert">
      <p className="text-sm" style={{ color: 'var(--hr-text)' }}>
        {t('setup.statusError')}
      </p>
      <button
        type="button"
        className="min-h-[44px] rounded-lg px-4 py-2 text-sm font-medium"
        style={{ backgroundColor: 'var(--accent)', color: 'var(--hr-card-bg)' }}
        disabled={isRetrying}
        onClick={onRetry}
      >
        {t('common.retry')}
      </button>
    </div>
  );
}

const HRConfigRoute = (): React.JSX.Element => {
  const { data: config, isLoading, isError, isFetching, refetch } = useGetHRConfig();
  const {
    data: setupStatus,
    isLoading: isSetupLoading,
    isError: isSetupError,
    refetch: refetchSetup,
  } = useGetSetupStatus();
  const setConfig = useHRConfigStore((s) => s.setConfig);

  React.useEffect(() => {
    if (config) {
      setConfig(config);
    }
  }, [config, setConfig]);

  if (isLoading || isSetupLoading) {
    return <PageLoader />;
  }

  // A first-run tenant can legitimately have no company config yet. The
  // setup-status endpoint is the source of truth for that state; do not hide
  // the redirect behind a 404 from the config endpoint.
  if (isError && !isSetupError && setupStatus && !setupStatus.isComplete) {
    return <Navigate to="/setup" replace />;
  }

  if (isError || isSetupError) {
    return (
      <ConfigLoadError
        onRetry={() => {
          refetch();
          refetchSetup();
        }}
        isRetrying={isFetching}
      />
    );
  }

  if (!config) {
    return <Navigate to="/setup" replace />;
  }

  if (!setupStatus?.isComplete) {
    return <Navigate to="/setup" replace />;
  }

  return <Outlet />;
};

export default HRConfigRoute;
