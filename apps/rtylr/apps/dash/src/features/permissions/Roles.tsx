import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { canEffectiveAccess } from '@voxire/permissions';
import { Archive, Check, Pencil, Plus } from 'lucide-react';
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
import {
  useArchiveCustomRole,
  useCreateCustomRole,
  useCustomRoles,
  usePermissions,
  useUpdateCustomRole,
} from '../../shared/hooks/queries/usePermissions';
import type { CustomRole } from '../../shared/schemas/permissions';
import { useUserStore } from '../../shared/stores/userStore';

const Roles: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const search = useSearch({ from: '/business/roles' });
  const user = useUserStore((s) => s.user);
  const { data = [], isLoading } = useCustomRoles();
  const { data: permissionMatrix } = usePermissions();
  const createRole = useCreateCustomRole();
  const updateRole = useUpdateCustomRole();
  const archiveRole = useArchiveCustomRole();
  const [open, setOpen] = useState(false);
  const [roleName, setRoleName] = useState('');
  const [roleDescription, setRoleDescription] = useState('');
  const [roleAppId, setRoleAppId] = useState('dash');
  const [rolePermissions, setRolePermissions] = useState<string[]>([]);
  const [editingRole, setEditingRole] = useState<CustomRole | null>(null);

  const canCreateRole = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:create')
    : user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'owner';
  const canEditRole = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:edit')
    : canCreateRole;
  const canArchiveRole = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:archive')
    : canCreateRole;

  const openCreateRole = () => {
    setRoleName('');
    setRoleDescription('');
    setRoleAppId('dash');
    setRolePermissions([]);
    setOpen(true);
  };

  const closeCreateRole = () => {
    setOpen(false);
    if (search.create === 1) {
      void navigate({ to: '/roles', search: {}, replace: true });
    }
  };

  useEffect(() => {
    if (search.create === 1 && canCreateRole) {
      setRoleName('');
      setRoleDescription('');
      setRoleAppId('dash');
      setRolePermissions([]);
      setOpen(true);
    }
  }, [search.create, canCreateRole]);

  const handleCreateRole = (e: React.FormEvent) => {
    e.preventDefault();
    if (!canCreateRole) return;
    const name = roleName.trim();
    if (name) {
      createRole.mutate({
        name,
        description: roleDescription.trim() || undefined,
        appId: roleAppId,
        permissions: rolePermissions,
      });
      setRoleName('');
      setRoleDescription('');
      setOpen(false);
    }
  };

  const handleUpdateRole = (event: React.FormEvent) => {
    event.preventDefault();
    if (!editingRole || !canEditRole) return;
    const name = roleName.trim();
    if (!name) return;
    updateRole.mutate(
      {
        id: editingRole.id,
        name,
        description: roleDescription.trim() || undefined,
        permissions: rolePermissions,
      },
      { onSuccess: () => setEditingRole(null) }
    );
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
          {canCreateRole && (
            <Button onClick={openCreateRole}>
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
                  <span
                    className="rounded-full px-1.5 py-0.5 text-[9px] font-semibold uppercase"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
                  >
                    {role.appId}
                  </span>
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {role.name}
                  </p>
                </div>
                <p className="text-[10px] mb-2" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('common.roleManagement.managedInDash')}
                </p>
                <p className="text-[11px] mb-3" style={{ color: 'var(--dash-text-dim)' }}>
                  {role.memberCount} {role.memberCount === 1 ? 'member' : 'members'} ·{' '}
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
                {(canEditRole || canArchiveRole) && (
                  <div className="flex gap-2 mt-4">
                    {canEditRole && (
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                          setEditingRole(role);
                          setRoleName(role.name);
                          setRoleDescription(role.description);
                          setRoleAppId(role.appId);
                          setRolePermissions(role.permissions);
                        }}
                      >
                        <Pencil size={13} />
                        {t('common.edit')}
                      </Button>
                    )}
                    {canArchiveRole && (
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => archiveRole.mutate(role.id)}
                        disabled={archiveRole.isPending}
                      >
                        <Archive size={13} />
                        {t('common.archive')}
                      </Button>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <Dialog
        open={open}
        onOpenChange={(nextOpen) => (nextOpen ? setOpen(true) : closeCreateRole())}
      >
        <DialogContent className="max-w-md">
          <form onSubmit={handleCreateRole} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Create Custom Role</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.roleManagement.application')}
              </label>
              <select
                value={roleAppId}
                onChange={(event) => setRoleAppId(event.target.value)}
                className="w-full rounded-lg px-3 py-2.5 text-base"
                style={{
                  background: 'var(--dash-bg)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
              >
                {[
                  'dash',
                  'pos',
                  'flow',
                  'finance',
                  'erp',
                  'hr',
                  'crm',
                  'insights',
                  'menu',
                  'recruit',
                ].map((appId) => (
                  <option key={appId} value={appId}>
                    {appId.toUpperCase()}
                  </option>
                ))}
              </select>
            </div>
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
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.roleManagement.permissions')}
              </label>
              <div
                className="max-h-52 overflow-y-auto space-y-1 rounded-lg p-2"
                style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
              >
                {(permissionMatrix?.permissions ?? [])
                  .filter((permission) => permission.key.startsWith(`${roleAppId}:`))
                  .map((permission) => {
                    const checked = rolePermissions.includes(permission.key);
                    return (
                      <button
                        type="button"
                        key={permission.key}
                        onClick={() =>
                          setRolePermissions((current) =>
                            checked
                              ? current.filter((key) => key !== permission.key)
                              : [...current, permission.key]
                          )
                        }
                        className="flex min-h-11 w-full items-center gap-2 rounded-md px-2 text-left"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        <span
                          className="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
                          style={{
                            borderColor: checked ? 'var(--accent)' : 'var(--dash-border)',
                            background: checked ? 'var(--accent)' : 'transparent',
                          }}
                        >
                          {checked && <Check size={11} color="white" />}
                        </span>
                        <span className="min-w-0">
                          <span className="block text-xs">{permission.label}</span>
                          <span
                            className="block truncate text-[10px] font-mono"
                            style={{ color: 'var(--dash-text-dim)' }}
                          >
                            {permission.key}
                          </span>
                        </span>
                      </button>
                    );
                  })}
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.description')}
              </label>
              <Input
                value={roleDescription}
                onChange={(event) => setRoleDescription(event.target.value)}
                placeholder={t('common.optionalRoleDescription')}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeCreateRole}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!roleName.trim() || createRole.isPending}>
                {t('common.createRole')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog
        open={editingRole !== null}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) setEditingRole(null);
        }}
      >
        <DialogContent className="max-w-md">
          <form onSubmit={handleUpdateRole} className="space-y-4">
            <DialogHeader>
              <DialogTitle>{t('common.editCustomRole')}</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.roleName')}
              </label>
              <Input
                value={roleName}
                onChange={(event) => setRoleName(event.target.value)}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.description')}
              </label>
              <Input
                value={roleDescription}
                onChange={(event) => setRoleDescription(event.target.value)}
                placeholder={t('common.optionalRoleDescription')}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('common.roleManagement.permissions')}
              </label>
              <div
                className="max-h-52 overflow-y-auto space-y-1 rounded-lg p-2"
                style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
              >
                {(permissionMatrix?.permissions ?? [])
                  .filter((permission) => permission.key.startsWith(`${roleAppId}:`))
                  .map((permission) => {
                    const checked = rolePermissions.includes(permission.key);
                    return (
                      <button
                        type="button"
                        key={permission.key}
                        onClick={() =>
                          setRolePermissions((current) =>
                            checked
                              ? current.filter((key) => key !== permission.key)
                              : [...current, permission.key]
                          )
                        }
                        className="flex min-h-11 w-full items-center gap-2 rounded-md px-2 text-left"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        <span
                          className="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
                          style={{
                            borderColor: checked ? 'var(--accent)' : 'var(--dash-border)',
                            background: checked ? 'var(--accent)' : 'transparent',
                          }}
                        >
                          {checked && <Check size={11} color="white" />}
                        </span>
                        <span className="min-w-0">
                          <span className="block text-xs">{permission.label}</span>
                          <span
                            className="block truncate text-[10px] font-mono"
                            style={{ color: 'var(--dash-text-dim)' }}
                          >
                            {permission.key}
                          </span>
                        </span>
                      </button>
                    );
                  })}
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditingRole(null)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!roleName.trim() || updateRole.isPending}>
                {t('common.saveRole')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default Roles;
