import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Users, Plus, Search, Trash2, ChevronRight } from 'lucide-react';
import { motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { Button, Input } from '../../shared/components/ui';
import { ConfirmDialog, SkeletonList } from '../../shared/components/dash';
import { useGroups, useDeleteGroup, type Group } from '../../shared/hooks/queries/useGroups';
import { APP_RESOURCES } from '../../shared/config/appResources';

/** Returns the unique app-level ids that have at least one active resourcePermission. */
const getActiveAppIds = (group: Group): string[] => {
  const appIds = new Set<string>();
  for (const p of group.resourcePermissions) {
    const appId = p.resourceKey.split('.')[0];
    if (appId) appIds.add(appId);
  }
  return Array.from(appIds);
};

/** Returns the label for an app id, falling back to the id uppercased. */
const appLabel = (appId: string): string =>
  APP_RESOURCES.find((a) => a.id === appId)?.label ?? appId.toUpperCase();

const Groups = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: groups = [], isLoading } = useGroups();
  const deleteGroup = useDeleteGroup();
  const [query, setQuery] = useState('');
  const [confirmDelete, setConfirmDelete] = useState<Group | null>(null);

  const filtered = query
    ? groups.filter((g) => g.name.toLowerCase().includes(query.toLowerCase()))
    : groups;

  return (
    <>
      <HelmetContainer title="Groups · rtylr" />
      <div className="space-y-5">
        <motion.div
          initial={{ opacity: 0, y: -6 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center justify-between"
        >
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('team.groups.title')}
            </h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              {t('team.groups.subtitle')}
            </p>
          </div>
          <Button onClick={() => navigate({ to: '/team/groups/new' })}>
            <Plus size={14} /> {t('team.groups.newGroup')}
          </Button>
        </motion.div>

        <div className="relative max-w-sm">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('team.groups.searchPlaceholder')}
            className="pl-9"
          />
        </div>

        <div
          className="rounded-xl"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isLoading ? (
            <div className="p-2">
              <SkeletonList rows={5} />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-4">
              <EmptyState
                icon={Users}
                title={query ? t('team.groups.noGroupsMatch') : t('team.groups.noGroupsTitle')}
                description={t('team.groups.noGroupsDescription')}
                action={{
                  label: t('team.groups.newGroup'),
                  onClick: () => navigate({ to: '/team/groups/new' }),
                }}
              />
            </div>
          ) : (
            <>
              {/* Header */}
              <div
                className="grid grid-cols-12 gap-3 px-5 py-2.5 text-[10px] font-semibold uppercase tracking-wider"
                style={{
                  color: 'var(--dash-text-dim)',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                <div className="col-span-5">{t('team.groups.columnName')}</div>
                <div className="col-span-4">{t('team.groups.columnSections')}</div>
                <div className="col-span-2">{t('team.groups.columnMembers')}</div>
                <div className="col-span-1"></div>
              </div>

              {filtered.map((g) => {
                const activeAppIds = getActiveAppIds(g);
                const sectionCount = g.resourcePermissions.length;
                return (
                  <div
                    key={g.id}
                    className="grid grid-cols-12 gap-3 items-center px-5 py-3.5 cursor-pointer hover:bg-[var(--dash-hover)] transition-colors"
                    style={{ borderBottom: '1px solid var(--dash-border)' }}
                    onClick={() => navigate({ to: `/team/groups/${g.id}` })}
                  >
                    {/* Name */}
                    <div className="col-span-5 flex items-center gap-3 min-w-0">
                      <div
                        className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0"
                        style={{
                          background: `${g.color}18`,
                          border: `1px solid ${g.color}30`,
                        }}
                      >
                        <Users size={14} style={{ color: g.color }} />
                      </div>
                      <div className="min-w-0">
                        <p
                          className="text-sm font-medium truncate"
                          style={{ color: 'var(--dash-text)' }}
                        >
                          {g.name}
                        </p>
                        {g.description && (
                          <p
                            className="text-[11px] truncate"
                            style={{ color: 'var(--dash-text-dim)' }}
                          >
                            {g.description}
                          </p>
                        )}
                      </div>
                    </div>

                    {/* Sections summary */}
                    <div className="col-span-4 flex items-center gap-1 flex-wrap">
                      {activeAppIds.length === 0 ? (
                        <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                          {t('team.groups.noneLabel')}
                        </span>
                      ) : (
                        <>
                          {activeAppIds.map((appId) => (
                            <span
                              key={appId}
                              className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                              style={{
                                background: `${g.color}18`,
                                color: g.color,
                              }}
                            >
                              {appLabel(appId)}
                            </span>
                          ))}
                          <span
                            className="text-[11px] ml-1"
                            style={{ color: 'var(--dash-text-dim)' }}
                          >
                            {t('groups.sectionsCount', { count: sectionCount })}
                          </span>
                        </>
                      )}
                    </div>

                    {/* Members */}
                    <div className="col-span-2">
                      <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                        {g.memberCount}
                      </span>
                      <span className="text-[11px] ml-1" style={{ color: 'var(--dash-text-dim)' }}>
                        {g.memberCount === 1
                          ? t('team.groups.userSingular')
                          : t('team.groups.userPlural')}
                      </span>
                    </div>

                    {/* Actions */}
                    <div className="col-span-1 flex justify-end items-center gap-1">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          setConfirmDelete(g);
                        }}
                        aria-label={t('team.groups.deleteConfirm')}
                        className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
                        style={{ color: 'var(--dash-text-muted)' }}
                      >
                        <Trash2 size={13} />
                      </button>
                      <ChevronRight size={14} style={{ color: 'var(--dash-text-dim)' }} />
                    </div>
                  </div>
                );
              })}
            </>
          )}
        </div>

        <ConfirmDialog
          open={!!confirmDelete}
          onOpenChange={(o) => !o && setConfirmDelete(null)}
          title={t('team.groups.deleteTitle', { name: confirmDelete?.name ?? '' })}
          description={t('team.groups.deleteDescription')}
          confirmLabel={t('team.groups.deleteConfirm')}
          destructive
          isLoading={deleteGroup.isPending}
          onConfirm={async () => {
            if (confirmDelete)
              try {
                await deleteGroup.mutateAsync(confirmDelete.id);
              } catch {
                // onError in hook handles toast
              }
          }}
        />
      </div>
    </>
  );
};

export default Groups;
