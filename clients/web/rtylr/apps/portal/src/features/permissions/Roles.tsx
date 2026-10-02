import { PageHeader } from '../../shared/components/dash/PageHeader';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { canEffectiveAccess } from '@voxire/permissions';
import { Archive, Check, Pencil, Plus, Shield } from 'lucide-react';
import {
  Button,
  Input,
  Select,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../shared/components/ui';
import { EmptyState, SkeletonList } from '../../shared/components/dash';
import {
  useArchivePermissionProfile,
  useCreatePermissionProfile,
  usePermissionCatalog,
  usePermissionProfiles,
  useUpdatePermissionProfile,
} from '../../shared/hooks/queries/usePermissions';
import type { PermissionProfile } from '../../shared/schemas/permissions';
import { useUserStore } from '../../shared/stores/userStore';

const Roles: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const search = useSearch({ from: '/business/roles' });
  const user = useUserStore((s) => s.user);
  const { data = [], isLoading, isError, refetch } = usePermissionProfiles();
  const { data: permissionCatalog } = usePermissionCatalog();
  const createRole = useCreatePermissionProfile();
  const updateRole = useUpdatePermissionProfile();
  const archiveRole = useArchivePermissionProfile();
  const [open, setOpen] = useState(false);
  const [roleName, setRoleName] = useState('');
  const [roleDescription, setRoleDescription] = useState('');
  const [roleAppId, setRoleAppId] = useState('dash');
  const [rolePermissions, setRolePermissions] = useState<string[]>([]);
  const [editingRole, setEditingRole] = useState<PermissionProfile | null>(null);

  const canCreateRole = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:create')
    : false;
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
      <HelmetContainer title={`${t('common.permissionProfiles.title')} · rtylr`} />
      <div className="portal-management-page space-y-5 max-w-5xl">
        <PageHeader
          title={t('common.permissionProfiles.title')}
          description={t('common.permissionProfiles.description')}
          actions={
            canCreateRole && (
              <Button onClick={openCreateRole}>
                <Plus />
                {t('common.permissionProfiles.new')}
              </Button>
            )
          }
        />

        {isError ? (
          <EmptyState
            icon={Shield}
            title={t('common.permissionProfiles.loadFailed')}
            action={{ label: t('common.refresh'), onClick: () => void refetch() }}
          />
        ) : isLoading ? (
          <SkeletonList rows={6} withAvatar={false} />
        ) : (
          <Table
            className="min-w-[720px] table-fixed"
            aria-label={t('common.permissionProfiles.title')}
          >
            <TableHeader>
              <TableRow>
                <TableHead className="w-[32%]">{t('common.permissionProfiles.name')}</TableHead>
                <TableHead className="w-[18%]">{t('common.roleManagement.application')}</TableHead>
                <TableHead className="w-[10%]">{t('common.permissionProfiles.members')}</TableHead>
                <TableHead className="w-[22%]">{t('common.roleManagement.permissions')}</TableHead>
                <TableHead className="w-[18%] text-end">
                  {t('common.permissionProfiles.actions')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-12 text-center text-[var(--dash-text-dim)]">
                    {t('common.permissionProfiles.empty')}
                  </TableCell>
                </TableRow>
              ) : (
                data.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="py-3">
                      <p className="font-medium break-words">{role.name}</p>
                      {role.description && (
                        <p className="mt-1 text-xs leading-relaxed text-[var(--dash-text-dim)] break-words">
                          {role.description}
                        </p>
                      )}
                    </TableCell>
                    <TableCell className="py-3 uppercase text-xs text-[var(--dash-text-dim)]">
                      {role.appId}
                    </TableCell>
                    <TableCell className="py-3 tabular-nums">{role.memberCount}</TableCell>
                    <TableCell className="py-3">
                      {role.permissions.length === 0 ? (
                        <span className="text-xs text-[var(--dash-text-dim)]">
                          {t('common.permissionProfiles.permissionCount', { count: 0 })}
                        </span>
                      ) : (
                        <details className="text-xs">
                          <summary className="cursor-pointer rounded-sm text-[var(--dash-text-dim)] hover:text-[var(--dash-text)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--dash-text)]">
                            {t('common.permissionProfiles.permissionCount', {
                              count: role.permissions.length,
                            })}
                          </summary>
                          <ul className="mt-2 space-y-1.5 text-[11px] text-[var(--dash-text-dim)]">
                            {role.permissions.map((permission) => (
                              <li key={permission} className="break-all font-mono">
                                {permission}
                              </li>
                            ))}
                          </ul>
                        </details>
                      )}
                    </TableCell>
                    <TableCell className="py-3">
                      <div className="flex justify-end gap-1">
                        {canEditRole && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            aria-label={`${t('common.edit')} ${role.name}`}
                            title={t('common.edit')}
                            className="px-2"
                            onClick={() => {
                              setEditingRole(role);
                              setRoleName(role.name);
                              setRoleDescription(role.description);
                              setRoleAppId(role.appId);
                              setRolePermissions(role.permissions);
                            }}
                          >
                            <Pencil size={16} />
                          </Button>
                        )}
                        {canArchiveRole && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            aria-label={`${t('common.archive')} ${role.name}`}
                            title={t('common.archive')}
                            className="px-2"
                            onClick={() => archiveRole.mutate(role.id)}
                            disabled={archiveRole.isPending}
                          >
                            <Archive size={16} />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        )}
      </div>

      <Dialog
        open={open}
        onOpenChange={(nextOpen) => (nextOpen ? setOpen(true) : closeCreateRole())}
      >
        <DialogContent className="max-w-lg">
          <form onSubmit={handleCreateRole}>
            <DialogHeader>
              <DialogTitle>{t('common.permissionProfiles.create')}</DialogTitle>
            </DialogHeader>
            <div className="grid gap-5 px-6 pb-6">
              <div className="space-y-2">
                <label
                  htmlFor="permission-profile-app"
                  className="block text-xs font-medium"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('common.roleManagement.application')}
                </label>
                <Select
                  id="permission-profile-app"
                  value={roleAppId}
                  onChange={(event) => {
                    setRoleAppId(event.target.value);
                    setRolePermissions([]);
                  }}
                  className="focus:ring-0"
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
                </Select>
              </div>
              <div className="space-y-2">
                <label
                  htmlFor="permission-profile-name"
                  className="block text-xs font-medium"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('common.permissionProfiles.name')}
                </label>
                <Input
                  id="permission-profile-name"
                  className="focus:ring-0"
                  value={roleName}
                  onChange={(e) => setRoleName(e.target.value)}
                  placeholder={t('common.permissionProfiles.namePlaceholder')}
                  autoFocus
                />
              </div>
              <fieldset className="min-w-0 space-y-2">
                <legend className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('common.roleManagement.permissions')}
                </legend>
                <div
                  className="max-h-56 overflow-y-auto overscroll-contain rounded-lg p-1"
                  style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
                >
                  {(permissionCatalog?.permissions ?? [])
                    .filter((permission) => permission.appId === roleAppId)
                    .map((permission) => {
                      const checked = rolePermissions.includes(permission.key);
                      return (
                        <button
                          type="button"
                          key={permission.key}
                          aria-pressed={checked}
                          onClick={() =>
                            setRolePermissions((current) =>
                              checked
                                ? current.filter((key) => key !== permission.key)
                                : [...current, permission.key]
                            )
                          }
                          className="flex min-h-12 w-full items-center gap-3 rounded-md px-3 py-2 text-start hover:bg-[var(--dash-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--dash-text)]"
                          style={{ color: 'var(--dash-text)' }}
                        >
                          <span
                            className="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
                            style={{
                              borderColor: checked
                                ? 'var(--dash-text)'
                                : 'var(--dash-border-focus)',
                              background: checked ? 'var(--dash-text)' : 'transparent',
                            }}
                          >
                            {checked && (
                              <Check size={11} style={{ color: 'var(--dash-surface)' }} />
                            )}
                          </span>
                          <span className="min-w-0">
                            <span className="block text-sm">
                              {permission.resource.replace(/:/g, ' / ')} · {permission.action}
                            </span>
                            <span
                              className="mt-0.5 block truncate text-[11px] font-mono"
                              style={{ color: 'var(--dash-text-dim)' }}
                            >
                              {permission.key}
                            </span>
                          </span>
                        </button>
                      );
                    })}
                </div>
              </fieldset>
              <div className="space-y-2">
                <label
                  htmlFor="permission-profile-description"
                  className="block text-xs font-medium"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('common.description')}
                </label>
                <Input
                  id="permission-profile-description"
                  className="focus:ring-0"
                  value={roleDescription}
                  onChange={(event) => setRoleDescription(event.target.value)}
                  placeholder={t('common.permissionProfiles.optionalDescription')}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeCreateRole}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!roleName.trim() || createRole.isPending}>
                {t('common.permissionProfiles.create')}
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
        <DialogContent className="max-w-lg">
          <form onSubmit={handleUpdateRole}>
            <DialogHeader>
              <DialogTitle>{t('common.permissionProfiles.edit')}</DialogTitle>
            </DialogHeader>
            <div className="grid gap-5 px-6 pb-6">
              <div className="space-y-2">
                <label
                  htmlFor="permission-profile-name"
                  className="block text-xs font-medium"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('common.permissionProfiles.name')}
                </label>
                <Input
                  id="permission-profile-name"
                  className="focus:ring-0"
                  value={roleName}
                  onChange={(event) => setRoleName(event.target.value)}
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <label
                  htmlFor="permission-profile-description"
                  className="block text-xs font-medium"
                  style={{ color: 'var(--dash-text-dim)' }}
                >
                  {t('common.description')}
                </label>
                <Input
                  id="permission-profile-description"
                  className="focus:ring-0"
                  value={roleDescription}
                  onChange={(event) => setRoleDescription(event.target.value)}
                  placeholder={t('common.permissionProfiles.optionalDescription')}
                />
              </div>
              <fieldset className="min-w-0 space-y-2">
                <legend className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('common.roleManagement.permissions')}
                </legend>
                <div
                  className="max-h-56 overflow-y-auto overscroll-contain rounded-lg p-1"
                  style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
                >
                  {(permissionCatalog?.permissions ?? [])
                    .filter((permission) => permission.appId === roleAppId)
                    .map((permission) => {
                      const checked = rolePermissions.includes(permission.key);
                      return (
                        <button
                          type="button"
                          key={permission.key}
                          aria-pressed={checked}
                          onClick={() =>
                            setRolePermissions((current) =>
                              checked
                                ? current.filter((key) => key !== permission.key)
                                : [...current, permission.key]
                            )
                          }
                          className="flex min-h-12 w-full items-center gap-3 rounded-md px-3 py-2 text-start hover:bg-[var(--dash-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--dash-text)]"
                          style={{ color: 'var(--dash-text)' }}
                        >
                          <span
                            className="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
                            style={{
                              borderColor: checked
                                ? 'var(--dash-text)'
                                : 'var(--dash-border-focus)',
                              background: checked ? 'var(--dash-text)' : 'transparent',
                            }}
                          >
                            {checked && (
                              <Check size={11} style={{ color: 'var(--dash-surface)' }} />
                            )}
                          </span>
                          <span className="min-w-0">
                            <span className="block text-sm">
                              {permission.resource.replace(/:/g, ' / ')} · {permission.action}
                            </span>
                            <span
                              className="mt-0.5 block truncate text-[11px] font-mono"
                              style={{ color: 'var(--dash-text-dim)' }}
                            >
                              {permission.key}
                            </span>
                          </span>
                        </button>
                      );
                    })}
                </div>
              </fieldset>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditingRole(null)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!roleName.trim() || updateRole.isPending}>
                {t('common.saveChanges')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default Roles;
