import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Users, Plus, Search, Trash2, ChevronRight } from 'lucide-react';
import { motion } from 'motion/react';
import { Button, Input } from '../../shared/components/ui';
import { ConfirmDialog, EmptyState, SkeletonList } from '../../shared/components/dash';
import { useGroups, useDeleteGroup, type Group } from '../../shared/hooks/queries/useGroups';

const APP_COLORS: Record<string, string> = {
  pos: '#4583C4',
  erp: '#10b981',
  hr: '#f59e0b',
  crm: '#8b5cf6',
  finance: '#ef4444',
  flow: '#64C5B9',
};

const Groups: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const { data: groups = [], isLoading } = useGroups();
  const deleteGroup = useDeleteGroup();
  const [query, setQuery] = useState('');
  const [confirmDelete, setConfirmDelete] = useState<Group | null>(null);

  const filtered = query
    ? groups.filter((g) => g.name.toLowerCase().includes(query.toLowerCase()))
    : groups;

  return (
    <div className="space-y-5">
      <motion.div
        initial={{ opacity: 0, y: -6 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center justify-between"
      >
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Groups
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            Bundle users into shared groups to control access across apps.
          </p>
        </div>
        <Button onClick={() => navigate('/team/groups/new')}>
          <Plus size={14} /> New group
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
          placeholder="Search groups..."
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
              title={query ? 'No groups match' : 'No groups yet'}
              description="Create a group to bundle users with shared app access."
              actionLabel="New group"
              onAction={() => navigate('/team/groups/new')}
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
              <div className="col-span-5">Group name</div>
              <div className="col-span-4">Apps</div>
              <div className="col-span-2">Members</div>
              <div className="col-span-1"></div>
            </div>

            {filtered.map((g) => (
              <div
                key={g.id}
                className="grid grid-cols-12 gap-3 items-center px-5 py-3.5 cursor-pointer hover:bg-[var(--dash-hover)] transition-colors"
                style={{ borderBottom: '1px solid var(--dash-border)' }}
                onClick={() => navigate(`/team/groups/${g.id}`)}
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
                      <p className="text-[11px] truncate" style={{ color: 'var(--dash-text-dim)' }}>
                        {g.description}
                      </p>
                    )}
                  </div>
                </div>

                {/* Apps */}
                <div className="col-span-4 flex items-center gap-1 flex-wrap">
                  {g.appPermissions.length === 0 ? (
                    <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                      None
                    </span>
                  ) : (
                    g.appPermissions.map((p) => (
                      <span
                        key={p.appId}
                        className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                        style={{
                          background: `${APP_COLORS[p.appId] ?? '#888'}18`,
                          color: APP_COLORS[p.appId] ?? '#888',
                        }}
                      >
                        {p.appId.toUpperCase()}
                      </span>
                    ))
                  )}
                </div>

                {/* Members */}
                <div className="col-span-2">
                  <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                    {g.memberCount}
                  </span>
                  <span className="text-[11px] ml-1" style={{ color: 'var(--dash-text-dim)' }}>
                    {g.memberCount === 1 ? 'user' : 'users'}
                  </span>
                </div>

                {/* Actions */}
                <div className="col-span-1 flex justify-end items-center gap-1">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setConfirmDelete(g);
                    }}
                    aria-label="Delete group"
                    className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    <Trash2 size={13} />
                  </button>
                  <ChevronRight size={14} style={{ color: 'var(--dash-text-dim)' }} />
                </div>
              </div>
            ))}
          </>
        )}
      </div>

      <ConfirmDialog
        open={!!confirmDelete}
        onOpenChange={(o) => !o && setConfirmDelete(null)}
        title={`Delete "${confirmDelete?.name}"?`}
        description="Members will lose access granted by this group. This cannot be undone."
        confirmLabel="Delete"
        destructive
        isLoading={deleteGroup.isPending}
        onConfirm={async () => {
          if (confirmDelete) await deleteGroup.mutateAsync(confirmDelete.id);
        }}
      />
    </div>
  );
};

export default Groups;
