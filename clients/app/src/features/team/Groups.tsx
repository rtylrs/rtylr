import { PageHeader } from '../../shared/components/dash/PageHeader';
import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Users, Plus, Search, Trash2, ChevronRight } from 'lucide-react';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import { Button, Input } from '../../shared/components/ui';
import { ConfirmDialog, SkeletonList } from '../../shared/components/dash';
import { useGroups, useDeleteGroup, type Group } from '../../shared/hooks/queries/useGroups';
import { APP_NAMES } from '../../shared/config/appNames';
import { NewGroupDialog } from './NewGroupDialog';

const appLabelKey = (appId: string): string =>
  APP_NAMES[appId as keyof typeof APP_NAMES]?.short ?? appId.toUpperCase();

const Groups = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const search = useSearch({ from: '/business/team/groups' });
  const creating = search.create === 1;
  const setCreating = (open: boolean): void => {
    void navigate({ to: '/team/groups', search: open ? { create: 1 } : {}, replace: !open });
  };
  const { data: groups = [], isLoading, isError, refetch } = useGroups();
  const deleteGroup = useDeleteGroup();
  const [query, setQuery] = useState('');
  const [confirmDelete, setConfirmDelete] = useState<Group | null>(null);

  const filtered = query
    ? groups.filter((g) => g.name.toLowerCase().includes(query.toLowerCase()))
    : groups;

  return (
    <>
      <HelmetContainer title={`${t('team.groups.title')} · rtylr`} />
      <NewGroupDialog open={creating} onOpenChange={setCreating} />
      <div className="portal-management-page space-y-6">
        <PageHeader
          title={t('team.groups.title')}
          description={t('team.groups.subtitle')}
          actions={
            <Button onClick={() => setCreating(true)}>
              <Plus /> {t('team.groups.newGroup')}
            </Button>
          }
        />

        <div className="relative max-w-sm">
          <Search
            size={14}
            className="absolute start-3 top-1/2 -translate-y-1/2 pointer-events-none"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('team.groups.searchPlaceholder')}
            aria-label={t('team.groups.searchPlaceholder')}
            className="ps-9"
          />
        </div>

        <div
          className="rounded-xl"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isError ? (
            <div className="p-4" role="alert">
              <EmptyState
                icon={Users}
                title={t('team.groups.loadFailed')}
                description={t('team.groups.loadFailedDescription')}
                action={{ label: t('common.refresh'), onClick: () => void refetch() }}
              />
            </div>
          ) : isLoading ? (
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
                  onClick: () => setCreating(true),
                }}
              />
            </div>
          ) : (
            <>
              {/* Header */}
              <div
                className="portal-list-heading grid grid-cols-12 min-w-[620px] gap-3 items-center px-5 py-3"
                style={{
                  color: 'var(--dash-text-dim)',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                <div className="col-span-5">{t('team.groups.columnName')}</div>
                <div className="col-span-4">{t('team.groups.columnProfiles')}</div>
                <div className="col-span-2">{t('team.groups.columnMembers')}</div>
                <div className="col-span-1"></div>
              </div>

              {filtered.map((g) => {
                const activeAppIds = g.licensedAppIds;
                const profileCount = g.permissionProfileAssignments.reduce(
                  (count, assignment) => count + assignment.permissionProfileIds.length,
                  0
                );
                return (
                  <div
                    key={g.id}
                    className="grid grid-cols-12 min-w-[620px] gap-3 items-center px-5 py-4.5 cursor-pointer hover:bg-[var(--dash-hover)] transition-colors"
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

                    {/* Licensed apps and permission profile count */}
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
                              {t(appLabelKey(appId))}
                            </span>
                          ))}
                          <span
                            className="text-[11px] ml-1"
                            style={{ color: 'var(--dash-text-dim)' }}
                          >
                            {t('team.groups.profilesCount', { count: profileCount })}
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
