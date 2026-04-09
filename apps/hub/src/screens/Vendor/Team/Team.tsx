import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { UserPlus, Search, Users } from 'lucide-react';
import { motion } from 'framer-motion';
import { Button, Input } from '../../../components/ui';
import {
  Breadcrumbs,
  EmptyState,
  SkeletonList,
  TeamMemberRow,
  ConfirmDialog,
} from '../../../components/hub';
import { useTeam, useDeactivateMember, type TeamMember } from '../../../hooks/queries/useTeam';

const Team: React.FC = () => {
  const navigate = useNavigate();
  const { data: members = [], isLoading } = useTeam();
  const deactivate = useDeactivateMember();
  const [query, setQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [confirm, setConfirm] = useState<TeamMember | null>(null);

  const filtered = useMemo(() => {
    return members.filter((m) => {
      const matchesQuery =
        !query.trim() ||
        m.name.toLowerCase().includes(query.toLowerCase()) ||
        m.email.toLowerCase().includes(query.toLowerCase());
      const matchesRole = roleFilter === 'all' || m.role === roleFilter;
      return matchesQuery && matchesRole;
    });
  }, [members, query, roleFilter]);

  return (
    <div className="space-y-5 max-w-6xl mx-auto">
      <Breadcrumbs />
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Team</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            {members.length} {members.length === 1 ? 'member' : 'members'} across your workspace
          </p>
        </div>
        <Button onClick={() => navigate('/team/invite')}>
          <UserPlus size={14} />
          Invite member
        </Button>
      </motion.div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" style={{ color: 'var(--hub-text-dim)' }} />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search members..."
            className="pl-9"
          />
        </div>
        <div className="flex items-center gap-1">
          {['all', 'owner', 'admin', 'manager', 'staff', 'viewer'].map((r) => (
            <button
              key={r}
              onClick={() => setRoleFilter(r)}
              className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
              style={{
                background: roleFilter === r ? 'rgba(69,131,196,0.15)' : 'var(--hub-hover)',
                color:      roleFilter === r ? '#4583C4' : 'var(--hub-text-muted)',
                border:     roleFilter === r ? '1px solid rgba(69,131,196,0.3)' : '1px solid var(--hub-border)',
              }}
            >
              {r}
            </button>
          ))}
        </div>
      </div>

      <div className="rounded-xl overflow-hidden" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
        {isLoading ? (
          <div className="p-2"><SkeletonList rows={6} /></div>
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
          <div className="p-2">
            {filtered.map((m) => (
              <TeamMemberRow key={m.id} member={m} onMenu={setConfirm} />
            ))}
          </div>
        )}
      </div>

      <ConfirmDialog
        open={!!confirm}
        onOpenChange={(o) => !o && setConfirm(null)}
        title={`Deactivate ${confirm?.name ?? 'member'}?`}
        description="The member will lose access to all apps immediately. You can reactivate them later."
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
