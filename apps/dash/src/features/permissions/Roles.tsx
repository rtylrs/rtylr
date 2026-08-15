import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Plus } from 'lucide-react';
import {
  Button,
  Input,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../../shared/components/ui';
import { SkeletonGrid } from '../../shared/components/dash';
import { useCustomRoles, useCreateCustomRole } from '../../shared/hooks/queries/usePermissions';
import { useUserStore } from '../../shared/stores/userStore';

const Roles: React.FC = () => {
  const user = useUserStore((s) => s.user);
  const { data = [], isLoading } = useCustomRoles();
  const createRole = useCreateCustomRole();
  const [open, setOpen] = useState(false);
  const [roleName, setRoleName] = useState('');

  const isAdminOrOwner =
    user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'owner';

  const handleCreateRole = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isAdminOrOwner) return;
    const name = roleName.trim();
    if (name) {
      createRole.mutate({ name });
      setRoleName('');
      setOpen(false);
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
          {isAdminOrOwner && (
            <Button onClick={() => setOpen(true)}>
              <Plus size={14} />
              New role
            </Button>
          )}
        </div>

        {isLoading ? (
          <SkeletonGrid cols={3} rows={2} />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
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

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-md">
          <form onSubmit={handleCreateRole} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Create Custom Role</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                Role Name
              </label>
              <Input
                value={roleName}
                onChange={(e) => setRoleName(e.target.value)}
                placeholder="e.g. Shift Manager"
                autoFocus
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={!roleName.trim() || createRole.isPending}>
                Create Role
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default Roles;
