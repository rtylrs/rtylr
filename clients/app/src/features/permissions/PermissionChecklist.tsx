import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, Search } from 'lucide-react';
import type { PermissionCatalog } from '../../shared/schemas/permissions';

type CatalogPermission = PermissionCatalog['permissions'][number];

interface PermissionChecklistProps {
  permissions: CatalogPermission[];
  selected: string[];
  onChange: (next: string[]) => void;
}

const labelOf = (permission: CatalogPermission) =>
  `${permission.resource.replace(/:/g, ' / ')} · ${permission.action}`;

/** The permission picker shared by the create and edit dialogs, with search. */
export function PermissionChecklist({ permissions, selected, onChange }: PermissionChecklistProps) {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const needle = query.trim().toLocaleLowerCase();
  const visible = needle
    ? permissions.filter((permission) =>
        `${labelOf(permission)} ${permission.key}`.toLocaleLowerCase().includes(needle)
      )
    : permissions;

  return (
    <fieldset className="min-w-0 space-y-2">
      <legend className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
        {t('common.roleManagement.permissions')}
      </legend>
      <div
        className="flex items-center gap-2 rounded-lg px-3"
        style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
      >
        <Search size={14} aria-hidden className="shrink-0" />
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={t('common.permissionProfiles.searchPermissions')}
          aria-label={t('common.permissionProfiles.searchPermissions')}
          className="h-9 w-full min-w-0 bg-transparent text-sm outline-none"
          style={{ color: 'var(--dash-text)' }}
        />
      </div>
      <div
        className="max-h-56 overflow-y-auto overscroll-contain rounded-lg p-1"
        style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
      >
        {visible.length === 0 ? (
          <p role="status" className="px-3 py-4 text-sm" style={{ color: 'var(--dash-text-dim)' }}>
            {t('common.permissionProfiles.noPermissionMatches', { query: query.trim() })}
          </p>
        ) : (
          visible.map((permission) => {
            const checked = selected.includes(permission.key);
            return (
              <button
                type="button"
                key={permission.key}
                aria-pressed={checked}
                onClick={() =>
                  onChange(
                    checked
                      ? selected.filter((key) => key !== permission.key)
                      : [...selected, permission.key]
                  )
                }
                className="flex min-h-12 w-full items-center gap-3 rounded-md px-3 py-2 text-start hover:bg-[var(--dash-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--dash-text)]"
                style={{ color: 'var(--dash-text)' }}
              >
                <span
                  className="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
                  style={{
                    borderColor: checked ? 'var(--dash-text)' : 'var(--dash-border-focus)',
                    background: checked ? 'var(--dash-text)' : 'transparent',
                  }}
                >
                  {checked && <Check size={11} style={{ color: 'var(--dash-surface)' }} />}
                </span>
                <span className="min-w-0">
                  <span className="block text-sm">{labelOf(permission)}</span>
                  <span
                    className="mt-0.5 block truncate text-[11px] font-mono"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    {permission.key}
                  </span>
                </span>
              </button>
            );
          })
        )}
      </div>
    </fieldset>
  );
}
