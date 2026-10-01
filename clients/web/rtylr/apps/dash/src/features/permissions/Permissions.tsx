import React from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { AlertTriangle, Shield } from 'lucide-react';
import { SkeletonList } from '../../shared/components/dash';
import { usePermissionCatalog } from '../../shared/hooks/queries/usePermissions';

const Permissions: React.FC = () => {
  const { t } = useTranslation();
  const { data, isLoading, isError, refetch } = usePermissionCatalog();

  if (isError) {
    return (
      <div className="max-w-5xl space-y-5">
        <EmptyState
          icon={Shield}
          title={t('common.permissionCatalog.loadErrorTitle')}
          description={t('common.permissionCatalog.loadError')}
          action={{ label: t('common.refresh'), onClick: () => void refetch() }}
        />
      </div>
    );
  }

  if (isLoading || !data) {
    return (
      <div className="max-w-5xl space-y-5">
        <SkeletonList rows={6} withAvatar={false} />
      </div>
    );
  }

  const byApp = data.permissions.reduce<Record<string, typeof data.permissions>>(
    (groups, entry) => {
      (groups[entry.appId] ??= []).push(entry);
      return groups;
    },
    {}
  );

  return (
    <>
      <HelmetContainer title={`${t('common.permissionCatalog.title')} · rtylr`} />
      <div className="max-w-5xl space-y-5">
        <header className="flex items-center gap-3">
          <div
            className="flex h-9 w-9 items-center justify-center rounded-xl"
            style={{ background: 'rgba(69,131,196,0.15)' }}
          >
            <Shield size={16} style={{ color: '#4583C4' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('common.permissionCatalog.title')}
            </h1>
            <p className="mt-0.5 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              {t('common.permissionCatalog.description')}
            </p>
          </div>
        </header>

        {Object.entries(byApp).map(([appId, permissions]) => (
          <section
            key={appId}
            aria-labelledby={`permission-app-${appId}`}
            className="overflow-hidden rounded-xl"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <h2
              id={`permission-app-${appId}`}
              className="border-b px-5 py-3 text-xs font-semibold uppercase tracking-wider"
              style={{
                borderColor: 'var(--dash-border)',
                background: 'var(--dash-hover)',
                color: 'var(--dash-text-muted)',
              }}
            >
              {appId}
            </h2>
            <ul className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
              {permissions.map((permission) => (
                <li
                  key={permission.key}
                  className="flex items-center justify-between gap-4 px-5 py-3"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <div className="min-w-0">
                    <p
                      className="truncate text-xs font-semibold"
                      style={{ color: 'var(--dash-text)' }}
                    >
                      {permission.resource.replace(/:/g, ' / ')} · {permission.action}
                    </p>
                    <p
                      className="truncate font-mono text-[10px]"
                      style={{ color: 'var(--dash-text-dim)' }}
                    >
                      {permission.key}
                    </p>
                  </div>
                  {(permission.isSensitive || permission.isHighRisk) && (
                    <span
                      className="inline-flex shrink-0 items-center gap-1 rounded-full px-2 py-1 text-[10px] font-medium"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                    >
                      <AlertTriangle size={11} />
                      {permission.isHighRisk
                        ? t('common.permissionProfiles.highImpact')
                        : t('common.permissionProfiles.sensitive')}
                    </span>
                  )}
                </li>
              ))}
            </ul>
          </section>
        ))}
      </div>
    </>
  );
};

export default Permissions;
