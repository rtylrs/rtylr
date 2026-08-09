import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Shield } from 'lucide-react';
import { PermissionCheckbox, RoleBadge, SkeletonList } from '../../shared/components/dash';
import { usePermissions } from '../../shared/hooks/queries/usePermissions';

const Permissions: React.FC = () => {
  const { data, isLoading } = usePermissions();

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

        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <div
            className="grid items-center px-5 py-3"
            style={{
              gridTemplateColumns: `minmax(0, 1fr) repeat(${data.roles.length}, 88px)`,
              borderBottom: '1px solid var(--dash-border)',
              background: 'var(--dash-hover)',
            }}
          >
            <p
              className="text-[10px] font-semibold uppercase tracking-widest"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              Permission
            </p>
            {data.roles.map((role) => (
              <div key={role} className="flex justify-center">
                <RoleBadge role={role} />
              </div>
            ))}
          </div>

          {Object.entries(groups).map(([group, perms]) => (
            <div key={group}>
              <p
                className="px-5 py-2 text-[10px] font-semibold uppercase tracking-widest"
                style={{
                  color: 'var(--dash-text-dim)',
                  background: 'var(--dash-hover)',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                {group}
              </p>
              {perms.map((p) => (
                <div
                  key={p.key}
                  className="grid items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)]"
                  style={{
                    gridTemplateColumns: `minmax(0, 1fr) repeat(${data.roles.length}, 88px)`,
                    borderBottom: '1px solid var(--dash-border)',
                  }}
                >
                  <div className="min-w-0">
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
                    return (
                      <div key={role} className="flex justify-center">
                        <PermissionCheckbox checked={isOwner ? true : value} disabled />
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </>
  );
};

export default Permissions;
