import React from 'react';
import { Navigate, Outlet } from '@tanstack/react-router';
import { useGetHRConfig } from '../../shared/hooks/queries/useSetup';
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

const HRConfigRoute = (): React.JSX.Element => {
  const { data: config, isLoading, isError } = useGetHRConfig();
  const setConfig = useHRConfigStore((s) => s.setConfig);

  React.useEffect(() => {
    if (config) {
      setConfig(config);
    }
  }, [config, setConfig]);

  if (isLoading) {
    return <PageLoader />;
  }

  if (isError || !config) {
    return <Navigate to="/setup" replace />;
  }

  if (!config.setupCompletedAt) {
    return <Navigate to="/setup" replace />;
  }

  return <Outlet />;
};

export default HRConfigRoute;
