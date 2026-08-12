import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Plus } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, SkeletonGrid } from '../../shared/components/dash';
import { useCustomRoles, useCreateCustomRole } from '../../shared/hooks/queries/usePermissions';

const Roles: React.FC = () => {
  const { data = [], isLoading } = useCustomRoles();
  const createRole = useCreateCustomRole();

  const handleCreateRole = () => {
    const name = window.prompt('Enter new role name:');
    if (name && name.trim()) {
      createRole.mutate({ name: name.trim() });
    }
  };

  return (
    <>
      <HelmetContainer title="Roles · rtylr" />
      <div className="space-y-5 max-w-5xl">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              Custom roles
            </h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              Build purpose-built roles for your business. Assign them to members in Team.
            </p>
          </div>
          <Button onClick={handleCreateRole}>
            <Plus size={14} />
            New role
          </Button>
        </div>

        {isLoading ? (
          <SkeletonGrid cols={3} rows={2} />
        ) : (
          <div className="grid grid-cols-3 gap-4">
            {data.map((role) => (
              <div
                key={role.id}
                className="rounded-xl p-5"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              >
                <div className="flex items-center gap-2 mb-3">
                  <span className="h-2.5 w-2.5 rounded-full" style={{ background: role.color }} />
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {role.name}
                  </p>
                </div>
                <p className="text-[11px] mb-3" style={{ color: 'var(--dash-text-dim)' }}>
                  {role.members} {role.members === 1 ? 'member' : 'members'} ·{' '}
                  {role.permissions.length} permissions
                </p>
                <div className="flex flex-wrap gap-1">
                  {role.permissions.slice(0, 4).map((p) => (
                    <span
                      key={p}
                      className="px-1.5 py-0.5 rounded text-[10px] font-mono"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                    >
                      {p}
                    </span>
                  ))}
                  {role.permissions.length > 4 && (
                    <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                      +{role.permissions.length - 4}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default Roles;
