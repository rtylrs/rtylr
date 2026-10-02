import React, { useEffect, useState } from 'react';
import { pageTitle } from '../../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import {
  useListRoles,
  useListPermissionCatalog,
  useCreateRole,
  useUpdateRole,
} from '../../../shared/hooks/queries/useRoles';
import type { Role, PermissionCatalogEntry } from '../../../shared/schemas/rolesPermissions';
import { Button, Card, SectionHeader } from '../../../shared/components/ui';

function groupByCategory(
  catalog: PermissionCatalogEntry[]
): Array<{ category: string; entries: PermissionCatalogEntry[] }> {
  const order: string[] = [];
  const groups = new Map<string, PermissionCatalogEntry[]>();
  for (const entry of catalog) {
    if (!groups.has(entry.category)) {
      order.push(entry.category);
      groups.set(entry.category, []);
    }
    groups.get(entry.category)?.push(entry);
  }
  return order.map((category) => ({ category, entries: groups.get(category) ?? [] }));
}

interface RoleCardProps {
  role: Role;
  catalog: PermissionCatalogEntry[];
  isExpanded: boolean;
  onToggleExpand: () => void;
}

const RoleCard = ({
  role,
  catalog,
  isExpanded,
  onToggleExpand,
}: RoleCardProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const updateRole = useUpdateRole();
  const [pending, setPending] = useState<Record<string, boolean>>(role.permissions);
  // The card is keyed by role.id, so a refetch (after any mutation invalidates
  // ['roles']) updates the prop without remounting. Re-seed the buffer then, or
  // a save would PUT the mount-time permissions and revert another admin's
  // change. React Query's structural sharing keeps the reference stable when
  // the refetched payload is unchanged, so in-progress edits survive.
  useEffect(() => {
    setPending(role.permissions);
  }, [role.permissions]);
  const grantedCount = Object.values(pending).filter(Boolean).length;
  const isDirty = JSON.stringify(pending) !== JSON.stringify(role.permissions);

  const togglePermission = (key: string): void => {
    setPending((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const save = (): void => {
    updateRole.mutate({
      id: role.id,
      name: role.name,
      description: role.description,
      permissions: JSON.stringify(pending),
    });
  };

  return (
    <Card>
      <button
        type="button"
        onClick={onToggleExpand}
        className="flex w-full min-h-[44px] items-center justify-between gap-3 px-1 text-left"
      >
        <div>
          <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {role.name}
          </span>
          <span
            className="ms-2 rounded px-1.5 py-0.5 text-xs"
            style={{
              background: role.isSystem ? 'var(--hr-hover)' : 'var(--accent)',
              color: role.isSystem ? 'var(--hr-text-dim)' : 'var(--hr-on-accent)',
            }}
          >
            {role.isSystem
              ? t('hrRolesPermissions.systemRoleBadge')
              : t('hrRolesPermissions.customRoleBadge')}
          </span>
        </div>
        <span className="text-xs" style={{ color: 'var(--hr-text-dim)' }}>
          {t('hrRolesPermissions.permissionsGrantedCount', {
            granted: grantedCount,
            total: catalog.length,
          })}
        </span>
      </button>

      {isExpanded && (
        <div className="mt-3 flex flex-col gap-4">
          {groupByCategory(catalog).map(({ category, entries }) => (
            <div key={category}>
              <p
                className="text-xs font-semibold uppercase"
                style={{ color: 'var(--hr-text-dim)' }}
              >
                {t(`hrRolesPermissions.categories.${category}`, category)}
              </p>
              <div className="mt-1.5 flex flex-col gap-2">
                {entries.map((entry) => (
                  <label key={entry.key} className="flex min-h-[44px] items-center gap-2">
                    <input
                      type="checkbox"
                      className="h-4 w-4"
                      checked={Boolean(pending[entry.key])}
                      onChange={() => togglePermission(entry.key)}
                      aria-label={t(`hrRolesPermissions.permissions.${entry.key}`, entry.label)}
                    />
                    <span className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
                      {t(`hrRolesPermissions.permissions.${entry.key}`, entry.label)}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          ))}

          <div className="flex items-center gap-3">
            <Button
              type="button"
              variant="primary"
              onClick={save}
              disabled={!isDirty || updateRole.isPending}
            >
              {updateRole.isPending ? t('hrRolesPermissions.saving') : t('hrRolesPermissions.save')}
            </Button>
            {isDirty && !updateRole.isPending && (
              <span className="text-xs" style={{ color: 'var(--hr-text-dim)' }}>
                {t('hrRolesPermissions.unsavedChanges')}
              </span>
            )}
          </div>
        </div>
      )}
    </Card>
  );
};

const RolesPermissions = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const { data: roles, isLoading: rolesLoading, isError: rolesError } = useListRoles();
  const {
    data: catalog,
    isLoading: catalogLoading,
    isError: catalogError,
  } = useListPermissionCatalog();
  const isLoading = rolesLoading || catalogLoading;
  const isError = rolesError || catalogError;
  const [expandedRoleId, setExpandedRoleId] = useState<number | null>(null);
  const [isAddingRole, setIsAddingRole] = useState(false);
  const [newRoleName, setNewRoleName] = useState('');
  const [newRoleDescription, setNewRoleDescription] = useState('');
  const createRole = useCreateRole();

  const submitNewRole = (event: React.FormEvent): void => {
    event.preventDefault();
    if (!newRoleName.trim()) return;
    createRole.mutate(
      {
        name: newRoleName.trim(),
        description: newRoleDescription.trim(),
        permissions: '{}',
      },
      {
        onSuccess: () => {
          setNewRoleName('');
          setNewRoleDescription('');
          setIsAddingRole(false);
        },
      }
    );
  };

  if (isLoading) {
    return (
      <div role="status" className="flex items-center justify-center py-20">
        <div
          className="h-8 w-8 rounded-full border-2 animate-spin"
          style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
        />
      </div>
    );
  }

  if (isError) {
    return (
      <p role="alert" className="py-10 text-center text-sm" style={{ color: 'var(--hr-text)' }}>
        {t('hrRolesPermissions.loadError')}
      </p>
    );
  }

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrNav.rolesPermissions'))} />
      <div className="flex flex-col gap-4">
        <SectionHeader
          title={t('hrRolesPermissions.title')}
          subtitle={t('hrRolesPermissions.subtitle')}
          right={
            <Button type="button" variant="secondary" onClick={() => setIsAddingRole(true)}>
              {t('hrRolesPermissions.addRole')}
            </Button>
          }
        />

        {isAddingRole && (
          <Card>
            <form onSubmit={submitNewRole} className="flex flex-col gap-3">
              <input
                type="text"
                value={newRoleName}
                onChange={(e) => setNewRoleName(e.target.value)}
                placeholder={t('hrRolesPermissions.newRoleNamePlaceholder')}
                className="min-h-[44px] rounded-lg px-3"
                style={{ border: '1px solid var(--hr-border)', fontSize: '16px' }}
                required
              />
              <input
                type="text"
                value={newRoleDescription}
                onChange={(e) => setNewRoleDescription(e.target.value)}
                placeholder={t('hrRolesPermissions.newRoleDescriptionPlaceholder')}
                className="min-h-[44px] rounded-lg px-3"
                style={{ border: '1px solid var(--hr-border)', fontSize: '16px' }}
              />
              <Button type="submit" variant="primary" disabled={createRole.isPending}>
                {t('hrRolesPermissions.createRole')}
              </Button>
            </form>
          </Card>
        )}

        {(roles ?? []).map((role) => (
          <RoleCard
            key={role.id}
            role={role}
            catalog={catalog ?? []}
            isExpanded={expandedRoleId === role.id}
            onToggleExpand={() => setExpandedRoleId(expandedRoleId === role.id ? null : role.id)}
          />
        ))}
      </div>
    </>
  );
};

export default RolesPermissions;
