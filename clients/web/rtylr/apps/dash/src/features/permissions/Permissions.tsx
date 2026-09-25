import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { canEffectiveAccess } from '@voxire/permissions';
import { Shield, Lock } from 'lucide-react';
import { PermissionCheckbox, RoleBadge, SkeletonList } from '../../shared/components/dash';
import { usePermissions, useUpdatePermissions } from '../../shared/hooks/queries/usePermissions';
import { useUserStore } from '../../shared/stores/userStore';

const Permissions: React.FC = () => {
  const user = useUserStore((s) => s.user);
  const { data, isLoading, isError, refetch } = usePermissions();
  const updatePermissions = useUpdatePermissions();

  const canEditRoles = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:edit') ||
      canEffectiveAccess(user.effectiveAccess, 'hub:roles:create')
    : user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'owner';

  // A failed request settles with isLoading false and data undefined, which
  // the guard below reads as loading and leaves the skeleton on screen.
  if (isError) {
    return (
      <div className="space-y-5 max-w-5xl">
        <EmptyState
          icon={Shield}
          title="Couldn't load permissions"
          description="The permission matrix could not be loaded."
          action={{ label: 'Refresh', onClick: () => void refetch() }}
        />
      </div>
    );
  }

  if (isLoading || !data) {
    return (
      <div className="space-y-5 max-w-5xl">
        <SkeletonList rows={6} withAvatar={false} />
      </div>
    );
  }

  const groups = data.permissions.reduce<Record<string, typeof data.permissions>>((acc, p) => {
    (acc[p.group] ??= []).push(p);
    return acc;
  }, {});

  return (
    <>
      <HelmetContainer title="Permissions · rtylr" />
      <div className="space-y-5 max-w-5xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <Shield size={16} style={{ color: '#4583C4' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Permissions
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                Configure what each role can do across the workspace.
              </p>
            </div>
          </div>
          {!canEditRoles && (
            <div
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
              style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
            >
              <Lock size={13} />
              Read-only mode
            </div>
          )}
        </div>

        <div
          role="table"
          aria-label="Permission Matrix"
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <div
            role="row"
            className="grid items-center px-5 py-3 font-medium text-xs border-b"
            style={{
              gridTemplateColumns: `minmax(0, 1fr) repeat(${data.roles.length}, 88px)`,
              borderColor: 'var(--dash-border)',
              background: 'var(--dash-hover)',
            }}
          >
            <div role="columnheader" style={{ color: 'var(--dash-text-muted)' }}>
              Permission
            </div>
            {data.roles.map((role) => (
              <div key={role} role="columnheader" className="flex justify-center">
                <RoleBadge role={role} />
              </div>
            ))}
          </div>

          <div role="rowgroup">
            {Object.entries(groups).map(([group, perms]) => (
              <div key={group}>
                <div
                  className="px-5 py-2 text-[11px] font-semibold uppercase tracking-wider"
                  style={{
                    background: 'var(--dash-hover)',
                    color: 'var(--dash-text-dim)',
                    borderBottom: '1px solid var(--dash-border)',
                  }}
                >
                  {group}
                </div>
                {perms.map((p) => (
                  <div
                    key={p.key}
                    role="row"
                    className="grid items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)]"
                    style={{
                      gridTemplateColumns: `minmax(0, 1fr) repeat(${data.roles.length}, 88px)`,
                      borderBottom: '1px solid var(--dash-border)',
                    }}
                  >
                    <div role="cell" className="min-w-0">
                      <p
                        className="text-xs font-semibold truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {p.label}
                      </p>
                      <p
                        className="text-[10px] font-mono truncate"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {p.key}
                      </p>
                    </div>
                    {data.roles.map((role) => {
                      const value = data.values[role]?.[p.key] ?? false;
                      const isOwner = role === 'owner';
                      const isDisabled = isOwner || !canEditRoles;
                      return (
                        <div key={role} role="cell" className="flex justify-center">
                          <PermissionCheckbox
                            checked={isOwner ? true : value}
                            disabled={isDisabled}
                            ariaLabel={`${p.label} for ${role}`}
                            onChange={(checked) => {
                              if (!isDisabled) {
                                updatePermissions.mutate({
                                  role,
                                  key: p.key,
                                  enabled: checked,
                                });
                              }
                            }}
                          />
                        </div>
                      );
                    })}
                  </div>
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
};

export default Permissions;
