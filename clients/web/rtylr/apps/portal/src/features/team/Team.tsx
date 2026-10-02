import { PageHeader } from '../../shared/components/dash/PageHeader';
import React, { useMemo, useState } from 'react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import {
  Search,
  UserPlus,
  Users,
  MoreHorizontal,
  Mail,
  Shield,
  Pencil,
  UserX,
  FolderOpen,
  X,
  Bell,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import { Button, Input, Select } from '../../shared/components/ui';
import { SkeletonList, ConfirmDialog, RoleBadge } from '../../shared/components/dash';
import {
  useTeam,
  useDeactivateMember,
  useAccessRequests,
  useApproveAccessRequest,
  type TeamMember,
  type AccessRequest,
} from '../../shared/hooks/queries/useTeam';
import { useMySubscription } from '../../shared/hooks/queries/useSubscription';
import { APP_ICONS } from '../../shared/config/appIcons';
import { SeatSavingsNote } from '../../shared/components/SeatSavingsNote';
import { useGroups, useAddGroupMembers } from '../../shared/hooks/queries/useGroups';
import { mailtoFor } from '../../shared/utils/mailto';

const APP_META: Record<
  string,
  { icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }> }
> = {
  pos: { icon: APP_ICONS.pos },
  erp: { icon: APP_ICONS.erp },
  hr: { icon: APP_ICONS.hr },
  crm: { icon: APP_ICONS.crm },
  finance: { icon: APP_ICONS.finance },
  flow: { icon: APP_ICONS.flow },
  insights: { icon: APP_ICONS.insights },
  menu: { icon: APP_ICONS.menu },
  recruit: { icon: APP_ICONS.recruit },
};

const ITEMS_PER_PAGE = 10;
// Search and filters cover the fetched window; hasMore exposes truncation.
const FETCH_WINDOW_SIZE = 200;

const Team: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const {
    data: teamData,
    isLoading,
    isError,
    refetch,
  } = useTeam({
    page: 1,
    pageSize: FETCH_WINDOW_SIZE,
  });
  const members = useMemo(() => teamData?.members ?? [], [teamData]);
  const deactivate = useDeactivateMember();
  const [query, setQuery] = useState('');
  const [memberTypeFilter, setMemberTypeFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('active');
  const [confirm, setConfirm] = useState<TeamMember | null>(null);
  const { data: sub } = useMySubscription();

  const [menuOpen, setMenuOpen] = useState<string | null>(null);

  // Bulk selection state
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showAddToGroupModal, setShowAddToGroupModal] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string>('');
  const [bulkConfirmDeactivate, setBulkConfirmDeactivate] = useState(false);

  const { data: groups = [] } = useGroups();
  const { data: accessRequests = [] } = useAccessRequests();
  const approveAccessRequest = useApproveAccessRequest();

  const addGroupMembers = useAddGroupMembers(selectedGroupId || undefined);

  const filtered = useMemo(() => {
    return members.filter((m) => {
      const matchesQuery =
        !query.trim() ||
        (m.name || '').toLowerCase().includes(query.toLowerCase()) ||
        (m.email || '').toLowerCase().includes(query.toLowerCase());
      const memberType = m.isOwner ? 'owner' : m.memberType;
      const matchesMemberType = memberTypeFilter === 'all' || memberType === memberTypeFilter;
      const isActive = m.status === 'active' || m.isActive === true;
      const matchesStatus =
        statusFilter === 'all' || (statusFilter === 'active' ? isActive : !isActive);
      return matchesQuery && matchesMemberType && matchesStatus;
    });
  }, [members, query, memberTypeFilter, statusFilter]);

  const totalPages = Math.ceil(filtered.length / ITEMS_PER_PAGE);
  const paged = filtered.slice(page * ITEMS_PER_PAGE, (page + 1) * ITEMS_PER_PAGE);

  const allPageIds = paged.map((m) => m.id);
  const allPageSelected = allPageIds.length > 0 && allPageIds.every((id) => selected.has(id));

  const toggleRow = (id: string): void => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleAllPage = (): void => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (allPageSelected) {
        allPageIds.forEach((id) => next.delete(id));
      } else {
        allPageIds.forEach((id) => next.add(id));
      }
      return next;
    });
  };

  const clearSelection = (): void => setSelected(new Set());

  const handleAddToGroup = async (): Promise<void> => {
    if (!selectedGroupId) return;
    // Member ids are the wire's int64-as-string values; pass them through
    // untouched instead of Number()-filtering, which silently dropped any
    // id that was not purely numeric and still toasted success.
    const userIds = Array.from(selected);
    if (userIds.length === 0) return;
    try {
      await addGroupMembers.mutateAsync(userIds);
      clearSelection();
      setShowAddToGroupModal(false);
      setSelectedGroupId('');
    } catch {
      // onError in hook handles toast
    }
  };

  const handleBulkDeactivate = async (): Promise<void> => {
    try {
      const nonOwnerIds = Array.from(selected).filter((id) => {
        const mem = members.find((m) => m.id === id);
        return mem && !mem.isOwner;
      });
      if (nonOwnerIds.length === 0) {
        toast.error(t('team.list.ownersCannotBeDeactivated'));
        setBulkConfirmDeactivate(false);
        return;
      }
      for (const id of nonOwnerIds) {
        await deactivate.mutateAsync(id);
      }
      clearSelection();
      setBulkConfirmDeactivate(false);
    } catch {
      // onError in hook handles toast
    }
  };

  return (
    <>
      <HelmetContainer title={`${t('common.nav.team')} · rtylr`} />
      <div className="portal-management-page space-y-6">
        <PageHeader
          title={t('common.nav.team')}
          description={
            !isLoading && !isError
              ? t('team.list.memberCount', { count: members.length })
              : undefined
          }
          actions={
            <>
              <Button variant="outline" onClick={() => navigate({ to: '/team/groups' })}>
                <FolderOpen /> {t('team.groups.title')}
              </Button>
              <Button onClick={() => navigate({ to: '/team/invite' })}>
                <UserPlus /> {t('common.nav.addMember')}
              </Button>
            </>
          }
        />

        {/* Pending access requests */}
        {accessRequests.length > 0 && (
          <div
            className="rounded-xl border px-4 py-3 space-y-2"
            style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-hover)' }}
          >
            <div className="flex items-center gap-2 mb-1">
              <Bell size={13} style={{ color: 'var(--dash-text)' }} />
              <span className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                {accessRequests.length} pending access{' '}
                {accessRequests.length === 1 ? 'request' : 'requests'}
              </span>
            </div>
            {accessRequests.map((req: AccessRequest) => (
              <div
                key={req.id}
                className="flex items-center justify-between text-xs py-1.5 px-2 rounded-lg"
                style={{ background: 'var(--dash-hover)' }}
              >
                <div className="flex items-center gap-2">
                  <span
                    className="h-6 w-6 rounded-full flex items-center justify-center text-[10px] font-bold uppercase shrink-0"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
                  >
                    {(req.requesterName || req.requesterEmail).charAt(0)}
                  </span>
                  <div>
                    <span style={{ color: 'var(--dash-text)' }}>
                      {req.requesterName || req.requesterEmail}
                    </span>
                    <span style={{ color: 'var(--dash-text-dim)' }}> wants access to </span>
                    <span className="font-semibold uppercase" style={{ color: 'var(--dash-text)' }}>
                      {req.app}
                    </span>
                  </div>
                </div>
                <button
                  type="button"
                  disabled={approveAccessRequest.isPending}
                  onClick={() =>
                    approveAccessRequest.mutate({ id: req.id, userId: req.userId, app: req.app })
                  }
                  className="text-[11px] px-2.5 py-1 rounded-md transition-colors disabled:opacity-50"
                  style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
                >
                  {approveAccessRequest.isPending ? '...' : 'Grant'}
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Filters */}
        <div className="portal-toolbar">
          <div className="relative flex-1 min-w-48 max-w-md">
            <Search
              size={14}
              className="absolute start-3 top-1/2 -translate-y-1/2 pointer-events-none"
              style={{ color: 'var(--dash-text-dim)' }}
            />
            <Input
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setPage(0);
              }}
              placeholder={t('team.list.searchPlaceholder')}
              aria-label={t('team.list.searchPlaceholder')}
              className="ps-9"
            />
          </div>
          <div className="portal-team-filters">
            <Select
              aria-label={t('team.list.memberTypeFilter')}
              value={memberTypeFilter}
              onChange={(event) => {
                setMemberTypeFilter(event.target.value);
                setPage(0);
              }}
            >
              <option value="all">{t('team.list.allTypes')}</option>
              <option value="admin">{t('common.roleLabels.admin')}</option>
              <option value="member">{t('common.roleLabels.member')}</option>
            </Select>
            <Select
              aria-label={t('team.list.statusFilter')}
              value={statusFilter}
              onChange={(event) => {
                setStatusFilter(event.target.value);
                setPage(0);
              }}
            >
              <option value="active">{t('team.list.activeStatus')}</option>
              <option value="inactive">{t('team.list.inactiveStatus')}</option>
              <option value="all">{t('team.list.allStatuses')}</option>
            </Select>
          </div>
        </div>

        {/* Member list */}
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isError ? (
            <div className="p-4" role="alert">
              <EmptyState
                icon={Users}
                title={t('team.list.loadFailed')}
                description={t('team.list.loadFailedDescription')}
                action={{ label: t('common.refresh'), onClick: () => void refetch() }}
              />
            </div>
          ) : isLoading ? (
            <div className="p-2">
              <SkeletonList rows={6} />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-4">
              <EmptyState
                icon={Users}
                title={
                  query
                    ? t('team.list.noMembersMatchTitle')
                    : members.length > 0
                      ? t('team.list.noMembersForFilters')
                      : t('team.list.noMembersTitle')
                }
                description={
                  query
                    ? t('team.list.noMembersMatchDescription')
                    : members.length > 0
                      ? t('team.list.noMembersForFiltersDescription')
                      : t('team.list.noMembersDescription')
                }
                action={
                  query
                    ? { label: t('common.clearSearch'), onClick: () => setQuery('') }
                    : members.length > 0
                      ? {
                          label: t('team.list.clearFilters'),
                          onClick: () => {
                            setQuery('');
                            setMemberTypeFilter('all');
                            setStatusFilter('all');
                            setPage(0);
                          },
                        }
                      : {
                          label: t('common.nav.addMember'),
                          onClick: () => navigate({ to: '/team/invite' }),
                        }
                }
              />
            </div>
          ) : (
            <>
              <div className="overflow-x-auto min-w-full">
                <div className="min-w-[720px]">
                  {/* Header */}
                  <div
                    className="portal-list-heading grid grid-cols-12 gap-3 items-center px-5 py-3"
                    style={{
                      color: 'var(--dash-text-dim)',
                      borderBottom: '1px solid var(--dash-border)',
                    }}
                  >
                    <div className="col-span-1 flex items-center">
                      <input
                        type="checkbox"
                        aria-label={t('team.list.selectAllOnPage')}
                        checked={allPageSelected}
                        onChange={toggleAllPage}
                        style={{ accentColor: 'var(--dash-text)' }}
                      />
                    </div>
                    <div className="col-span-3">{t('team.list.colMember')}</div>
                    <div className="col-span-2">{t('team.list.colRole')}</div>
                    <div className="col-span-4">{t('team.list.colLicenses')}</div>
                    <div className="col-span-1">{t('team.list.colStatus')}</div>
                    <div className="col-span-1"></div>
                  </div>

                  {/* Rows */}
                  {paged.map((m) => {
                    const apps =
                      m.effectiveAccess?.apps
                        .filter((app) => app.licensed)
                        .map((app) => app.appId) ?? m.licensedAppIds;
                    const isChecked = selected.has(m.id);
                    const emailHref = mailtoFor(m.email);
                    return (
                      <div
                        key={m.id}
                        className="grid grid-cols-12 gap-3 items-center px-5 py-4 transition-colors hover:bg-[var(--dash-hover)] cursor-pointer"
                        style={{ borderBottom: '1px solid var(--dash-border)' }}
                        onClick={() => navigate({ to: `/team/${m.id}` })}
                      >
                        {/* Checkbox */}
                        <div className="col-span-1" onClick={(e) => e.stopPropagation()}>
                          <input
                            type="checkbox"
                            aria-label={t('team.list.selectMember', { name: m.name || m.email })}
                            checked={isChecked}
                            onChange={() => toggleRow(m.id)}
                            style={{ accentColor: 'var(--dash-text)' }}
                          />
                        </div>

                        {/* Name + email */}
                        <div className="col-span-3 flex items-center gap-3 min-w-0">
                          <div
                            className="h-8 w-8 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0"
                            style={{
                              background: 'var(--accent-muted)',
                              color: 'var(--accent-on-muted)',
                            }}
                          >
                            {(m.name || m.email || '?').charAt(0).toUpperCase()}
                          </div>
                          <div className="min-w-0">
                            <button
                              type="button"
                              className="block max-w-full truncate text-sm font-medium text-start hover:underline"
                              style={{ color: 'var(--dash-text)' }}
                              onClick={(event) => {
                                event.stopPropagation();
                                navigate({ to: `/team/${m.id}` });
                              }}
                            >
                              {m.name || m.email}
                            </button>
                            <p
                              className="text-[11px] truncate"
                              style={{ color: 'var(--dash-text-dim)' }}
                            >
                              {m.email}
                            </p>
                          </div>
                        </div>

                        {/* Role */}
                        <div className="col-span-2">
                          <RoleBadge role={m.isOwner ? 'owner' : m.memberType} />
                        </div>

                        {/* App licenses */}
                        <div className="col-span-4 flex items-center gap-1.5 flex-wrap">
                          {apps.length === 0 ? (
                            <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                              {t('team.list.noLicenses')}
                            </span>
                          ) : (
                            apps.map((appId: string) => {
                              const meta = APP_META[appId];
                              if (!meta)
                                return (
                                  <span
                                    key={appId}
                                    className="text-[10px] px-1.5 py-0.5 rounded"
                                    style={{
                                      background: 'var(--dash-hover)',
                                      color: 'var(--dash-text-muted)',
                                    }}
                                  >
                                    {appId}
                                  </span>
                                );
                              const Icon = meta.icon;
                              return (
                                <div
                                  key={appId}
                                  className="h-7 w-7 rounded-lg flex items-center justify-center"
                                  style={{ background: 'var(--dash-hover)' }}
                                  title={appId.toUpperCase()}
                                >
                                  <Icon size={14} style={{ color: 'var(--dash-text-dim)' }} />
                                </div>
                              );
                            })
                          )}
                        </div>

                        {/* Status */}
                        <div className="col-span-1 flex items-center gap-1.5">
                          <span
                            className="h-2 w-2 rounded-full inline-block shrink-0"
                            style={{
                              background:
                                m.status === 'active' || m.isActive ? '#10b981' : '#ef4444',
                            }}
                          />
                          <span
                            className="text-[11px] capitalize whitespace-nowrap"
                            style={{ color: 'var(--dash-text-muted)' }}
                          >
                            {t(
                              m.status === 'active' || m.isActive
                                ? 'team.list.activeStatus'
                                : 'team.list.inactiveStatus'
                            )}
                          </span>
                        </div>

                        {/* Actions */}
                        <div className="col-span-1 flex justify-end relative">
                          <button
                            type="button"
                            aria-label={t('team.list.memberActions')}
                            onClick={(e) => {
                              e.stopPropagation();
                              setMenuOpen(menuOpen === m.id ? null : m.id);
                            }}
                            className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)]"
                            style={{ color: 'var(--dash-text-muted)' }}
                          >
                            <MoreHorizontal size={14} />
                          </button>

                          {menuOpen === m.id && (
                            <div
                              className="absolute right-0 top-8 z-50 rounded-lg py-1 w-48 shadow-lg"
                              style={{
                                background: 'var(--dash-surface)',
                                border: '1px solid var(--dash-border)',
                              }}
                              onClick={(e) => e.stopPropagation()}
                            >
                              <button
                                type="button"
                                className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-start hover:bg-[var(--dash-hover)]"
                                style={{ color: 'var(--dash-text-muted)' }}
                                onClick={() => {
                                  navigate({ to: `/team/${m.id}` });
                                  setMenuOpen(null);
                                }}
                              >
                                <Pencil size={12} /> {t('team.list.editMember')}
                              </button>
                              {emailHref ? (
                                <a
                                  href={emailHref}
                                  className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-start hover:bg-[var(--dash-hover)]"
                                  style={{ color: 'var(--dash-text-muted)' }}
                                  onClick={() => setMenuOpen(null)}
                                >
                                  <Mail size={12} /> Send email
                                </a>
                              ) : (
                                <span
                                  className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-start"
                                  style={{ color: 'var(--dash-text-dim)' }}
                                >
                                  <Mail size={12} /> Send email
                                </span>
                              )}
                              <button
                                type="button"
                                className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-start hover:bg-[var(--dash-hover)]"
                                style={{ color: 'var(--dash-text-muted)' }}
                                onClick={() => {
                                  navigate({ to: `/team/${m.id}` });
                                  setMenuOpen(null);
                                }}
                              >
                                <Shield size={12} /> Change role
                              </button>
                              <div
                                className="my-1"
                                style={{ borderTop: '1px solid var(--dash-border)' }}
                              />
                              {!m.isOwner && (
                                <button
                                  type="button"
                                  className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-start hover:bg-[var(--dash-hover)]"
                                  style={{ color: '#ef4444' }}
                                  onClick={() => {
                                    setConfirm(m);
                                    setMenuOpen(null);
                                  }}
                                >
                                  <UserX size={12} /> {t('team.list.deactivateConfirm')}
                                </button>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Pagination */}
              {(page > 0 || totalPages > 1) && (
                <div className="flex items-center justify-between px-5 py-3">
                  <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                    {teamData?.hasMore
                      ? t('team.list.showingTruncated', { count: filtered.length })
                      : t(
                          filtered.length === 1
                            ? 'team.list.membersCountSingular'
                            : 'team.list.membersCountPlural',
                          { count: filtered.length }
                        )}
                  </p>
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={() => setPage(Math.max(0, page - 1))}
                      disabled={page === 0}
                      className="px-2.5 py-1 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                    >
                      {t('common.pagination.previous')}
                    </button>
                    <button
                      type="button"
                      onClick={() => setPage(page + 1)}
                      disabled={page >= totalPages - 1}
                      className="px-2.5 py-1 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                    >
                      {t('common.pagination.next')}
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        {/* Close menu on outside click */}
        {menuOpen && <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(null)} />}

        {/* Floating bulk action bar */}
        {selected.size > 0 && (
          <div
            className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 rounded-xl shadow-2xl"
            style={{ background: 'var(--dash-surface)', border: '1px solid var(--dash-border)' }}
          >
            <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
              {selected.size} selected
            </span>
            <div className="w-px h-4" style={{ background: 'var(--dash-border)' }} />
            <button
              type="button"
              onClick={() => setShowAddToGroupModal(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
              style={{
                background: 'var(--dash-hover)',
                color: 'var(--dash-text)',
                border: '1px solid var(--dash-border)',
              }}
            >
              <FolderOpen size={12} /> Add to group
            </button>
            <button
              type="button"
              onClick={() => setBulkConfirmDeactivate(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
              style={{
                background: 'rgba(239,68,68,0.08)',
                color: '#ef4444',
                border: '1px solid rgba(239,68,68,0.2)',
              }}
            >
              <UserX size={12} /> Deactivate
            </button>
            <button
              type="button"
              aria-label={t('team.list.clearSelection')}
              onClick={clearSelection}
              className="h-6 w-6 rounded-md flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <X size={12} />
            </button>
          </div>
        )}

        {/* Add to group modal */}
        {showAddToGroupModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/40"
              onClick={() => setShowAddToGroupModal(false)}
            />
            <div
              className="relative rounded-xl p-6 w-full max-w-sm space-y-4 shadow-2xl"
              style={{ background: 'var(--dash-surface)', border: '1px solid var(--dash-border)' }}
            >
              <h3 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {t('team.list.addToGroupTitle', { count: selected.size })}
              </h3>
              <div className="space-y-1.5">
                <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                  {t('team.list.selectGroupLabel')}
                </label>
                <select
                  value={selectedGroupId}
                  onChange={(e) => setSelectedGroupId(e.target.value)}
                  className="w-full rounded-lg px-3 py-2 text-sm"
                  style={{
                    background: 'var(--dash-bg)',
                    border: '1px solid var(--dash-border)',
                    color: 'var(--dash-text)',
                  }}
                >
                  <option value="">{t('team.list.selectGroupPlaceholder')}</option>
                  {groups.map((g) => (
                    <option key={g.id} value={String(g.id)}>
                      {g.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddToGroupModal(false);
                    setSelectedGroupId('');
                  }}
                  className="px-3 py-1.5 rounded-lg text-sm"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  {t('common.cancel')}
                </button>
                <Button
                  onClick={handleAddToGroup}
                  disabled={!selectedGroupId || addGroupMembers.isPending}
                >
                  {t('team.memberDetail.addToGroup')}
                </Button>
              </div>
            </div>
          </div>
        )}

        <ConfirmDialog
          open={bulkConfirmDeactivate}
          onOpenChange={(o) => !o && setBulkConfirmDeactivate(false)}
          title={t('team.list.bulkDeactivateTitle', { count: selected.size })}
          description={t('team.list.bulkDeactivateDescription')}
          confirmLabel={t('team.list.deactivateConfirm')}
          destructive
          isLoading={deactivate.isPending}
          onConfirm={handleBulkDeactivate}
        />

        <ConfirmDialog
          open={!!confirm}
          onOpenChange={(o) => !o && setConfirm(null)}
          title={t('team.list.deactivateTitle', {
            name: confirm?.name || confirm?.email || t('team.list.deactivateFallbackName'),
          })}
          description={
            <span className="flex flex-col gap-1.5">
              <span>{t('team.list.deactivateDescription')}</span>
              {sub && <SeatSavingsNote pricePerSeatCents={sub.pricePerSeatCents} />}
            </span>
          }
          confirmLabel={t('team.list.deactivateConfirm')}
          destructive
          isLoading={deactivate.isPending}
          onConfirm={async () => {
            if (confirm)
              try {
                await deactivate.mutateAsync(confirm.id);
              } catch {
                // onError in hook handles toast
              }
          }}
        />
      </div>
    </>
  );
};

export default Team;
