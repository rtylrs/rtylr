import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Search,
  UserPlus,
  Users,
  MoreHorizontal,
  Mail,
  Shield,
  Pencil,
  UserX,
  MonitorSpeaker,
  Package,
  UserCheck,
  CreditCard,
  Layers,
} from 'lucide-react';
import { motion } from 'motion/react';
import { Button, Input } from '../../shared/components/ui';
import api from '../../shared/lib/api';
import { EmptyState, SkeletonList, ConfirmDialog, RoleBadge } from '../../shared/components/dash';
import { useTeam, useDeactivateMember, type TeamMember } from '../../shared/hooks/queries/useTeam';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';

const APP_ICONS: Record<
  string,
  { icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>; color: string }
> = {
  pos: { icon: MonitorSpeaker, color: '#4583C4' },
  erp: { icon: Package, color: '#10b981' },
  hr: { icon: UserCheck, color: '#f59e0b' },
  crm: { icon: Users, color: '#8b5cf6' },
  finance: { icon: CreditCard, color: '#ef4444' },
  flow: { icon: Layers, color: '#64C5B9' },
};

const ITEMS_PER_PAGE = 10;

const Team: React.FC = () => {
  const navigate = useNavigate();
  const { data: membersRaw = [], isLoading } = useTeam();
  const members = Array.isArray(membersRaw) ? membersRaw : [];
  const deactivate = useDeactivateMember();
  const [query, setQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [confirm, setConfirm] = useState<TeamMember | null>(null);
  const { data: sub } = useMySubscription();
  const { data: addonsData } = useQuery<Array<{ id: string; isEnabled: boolean }>>({
    queryKey: ['addons'],
    queryFn: async () => {
      const res = await api.get('/addons');
      return res.data?.addons ?? [];
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a) => a.id === 'menu' && a.isEnabled);

  const [menuOpen, setMenuOpen] = useState<string | null>(null);
  const [page, setPage] = useState(0);

  const filtered = useMemo(() => {
    return members.filter((m) => {
      const matchesQuery =
        !query.trim() ||
        (m.name || '').toLowerCase().includes(query.toLowerCase()) ||
        (m.email || '').toLowerCase().includes(query.toLowerCase());
      const matchesRole = roleFilter === 'all' || m.role === roleFilter;
      return matchesQuery && matchesRole;
    });
  }, [members, query, roleFilter]);

  const totalPages = Math.ceil(filtered.length / ITEMS_PER_PAGE);
  const paged = filtered.slice(page * ITEMS_PER_PAGE, (page + 1) * ITEMS_PER_PAGE);

  return (
    <div className="space-y-5">
      <motion.div
        initial={{ opacity: 0, y: -6 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center justify-between"
      >
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Team
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {members.length} {members.length === 1 ? 'member' : 'members'}
          </p>
        </div>
        <Button onClick={() => navigate('/team/invite')}>
          <UserPlus size={14} />
          Invite member
        </Button>
      </motion.div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setPage(0);
            }}
            placeholder="Search by name or email..."
            className="pl-9"
          />
        </div>
        <div className="flex items-center gap-1">
          {[
            { id: 'all', label: 'All' },
            { id: 'admin', label: 'Admin' },
            { id: 'operator', label: 'Operator' },
            { id: 'billing', label: 'Billing' },
            ...(menuAddonEnabled ? [{ id: 'menu_manager', label: 'Menu' }] : []),
            { id: 'staff', label: 'Staff' },
            { id: 'viewer', label: 'Viewer' },
          ].map((r) => (
            <button
              key={r.id}
              onClick={() => {
                setRoleFilter(r.id);
                setPage(0);
              }}
              className="px-2.5 py-1 rounded-md text-[11px] font-medium transition-colors"
              style={{
                background: roleFilter === r.id ? 'rgba(69,131,196,0.12)' : 'transparent',
                color: roleFilter === r.id ? '#4583C4' : 'var(--dash-text-muted)',
              }}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      {/* Member list */}
      <div
        className="rounded-xl"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        {isLoading ? (
          <div className="p-2">
            <SkeletonList rows={6} />
          </div>
        ) : filtered.length === 0 ? (
          <div className="p-4">
            <EmptyState
              icon={Users}
              title={query ? 'No members match your search' : 'No team members yet'}
              description="Invite your first team member to get started."
              actionLabel="Invite member"
              onAction={() => navigate('/team/invite')}
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
              <div className="col-span-4">Member</div>
              <div className="col-span-2">Role</div>
              <div className="col-span-4">Licenses</div>
              <div className="col-span-1">Status</div>
              <div className="col-span-1"></div>
            </div>

            {/* Rows */}
            {paged.map((m) => {
              const apps = m.apps ?? [];
              return (
                <div
                  key={m.id}
                  className="grid grid-cols-12 gap-3 items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)] cursor-pointer"
                  style={{ borderBottom: '1px solid var(--dash-border)' }}
                  onClick={() => navigate(`/team/${m.id}`)}
                >
                  {/* Name + email */}
                  <div className="col-span-4 flex items-center gap-3 min-w-0">
                    <div
                      className="h-8 w-8 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0"
                      style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
                    >
                      {(m.name || m.email || '?').charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p
                        className="text-sm font-medium truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {m.name || m.email}
                      </p>
                      <p className="text-[11px] truncate" style={{ color: 'var(--dash-text-dim)' }}>
                        {m.email}
                      </p>
                    </div>
                  </div>

                  {/* Role */}
                  <div className="col-span-2">
                    <RoleBadge role={m.role} />
                  </div>

                  {/* App licenses */}
                  <div className="col-span-4 flex items-center gap-1.5 flex-wrap">
                    {apps.length === 0 ? (
                      <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                        No licenses
                      </span>
                    ) : (
                      apps.map((appId: string) => {
                        const meta = APP_ICONS[appId];
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
                            style={{ background: `${meta.color}12` }}
                            title={appId.toUpperCase()}
                          >
                            <Icon size={14} style={{ color: meta.color }} />
                          </div>
                        );
                      })
                    )}
                  </div>

                  {/* Status */}
                  <div className="col-span-1">
                    <span
                      className="h-2 w-2 rounded-full inline-block"
                      style={{
                        background: m.status === 'active' || m.isActive ? '#10b981' : '#ef4444',
                      }}
                    />
                  </div>

                  {/* Actions */}
                  <div className="col-span-1 flex justify-end relative">
                    <button
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
                          className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-left hover:bg-[var(--dash-hover)]"
                          style={{ color: 'var(--dash-text-muted)' }}
                          onClick={() => {
                            navigate(`/team/${m.id}`);
                            setMenuOpen(null);
                          }}
                        >
                          <Pencil size={12} /> Edit member
                        </button>
                        <button
                          className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-left hover:bg-[var(--dash-hover)]"
                          style={{ color: 'var(--dash-text-muted)' }}
                          onClick={() => {
                            window.location.href = `mailto:${m.email}`;
                            setMenuOpen(null);
                          }}
                        >
                          <Mail size={12} /> Send email
                        </button>
                        <button
                          className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-left hover:bg-[var(--dash-hover)]"
                          style={{ color: 'var(--dash-text-muted)' }}
                          onClick={() => {
                            navigate(`/team/${m.id}`);
                            setMenuOpen(null);
                          }}
                        >
                          <Shield size={12} /> Change role
                        </button>
                        <div
                          className="my-1"
                          style={{ borderTop: '1px solid var(--dash-border)' }}
                        />
                        <button
                          className="flex items-center gap-2 w-full px-3 py-2 text-xs font-medium text-left hover:bg-[var(--dash-hover)]"
                          style={{ color: '#ef4444' }}
                          onClick={() => {
                            setConfirm(m);
                            setMenuOpen(null);
                          }}
                        >
                          <UserX size={12} /> Deactivate
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between px-5 py-3">
                <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                  Showing {page * ITEMS_PER_PAGE + 1}-
                  {Math.min((page + 1) * ITEMS_PER_PAGE, filtered.length)} of {filtered.length}
                </p>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setPage(Math.max(0, page - 1))}
                    disabled={page === 0}
                    className="px-2.5 py-1 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                  >
                    Prev
                  </button>
                  <button
                    onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
                    disabled={page >= totalPages - 1}
                    className="px-2.5 py-1 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Close menu on outside click */}
      {menuOpen && <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(null)} />}

      <ConfirmDialog
        open={!!confirm}
        onOpenChange={(o) => !o && setConfirm(null)}
        title={`Deactivate ${confirm?.name || confirm?.email || 'member'}?`}
        description={
          <span className="flex flex-col gap-1.5">
            <span>
              The member will lose access to all apps immediately. You can reactivate them later.
            </span>
            {sub &&
              sub.pricePerSeatCents > 0 &&
              (() => {
                const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                  computeSeatProration(sub.pricePerSeatCents);
                return (
                  <span style={{ color: 'var(--dash-text-muted)', fontSize: '12px' }}>
                    Saves{' '}
                    <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {centsToDisplay(proratedCents)}
                    </span>{' '}
                    this month ({daysRemaining}/{daysInMonth} days), then{' '}
                    <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {centsToDisplay(sub.pricePerSeatCents)}/mo
                    </span>{' '}
                    from {nextMonthName} 1st.
                  </span>
                );
              })()}
          </span>
        }
        confirmLabel="Deactivate"
        destructive
        isLoading={deactivate.isPending}
        onConfirm={async () => {
          if (confirm) await deactivate.mutateAsync(confirm.id);
        }}
      />
    </div>
  );
};

export default Team;
