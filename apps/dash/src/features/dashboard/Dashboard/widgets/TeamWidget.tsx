import React from 'react';
import { Users, ArrowUpRight } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../../../shared/lib/api';

interface WidgetMember {
  id?: string;
  fullName?: string;
  email?: string;
  role?: string;
  isActive?: boolean;
}

const TeamWidget: React.FC = () => {
  const navigate = useNavigate();
  const { data: raw } = useQuery({
    queryKey: ['team-widget'],
    queryFn: async (): Promise<WidgetMember[]> => {
      try {
        const res = await unwrap(
          safe.get<{ users?: WidgetMember[] } | WidgetMember[]>('/api/hub/team/members/v1/members')
        );
        return (Array.isArray(res) ? res : res.users) ?? [];
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });
  const members: WidgetMember[] = Array.isArray(raw) ? (raw as WidgetMember[]) : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center justify-between px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <div className="flex items-center gap-2">
          <Users size={14} style={{ color: 'var(--dash-text-muted)' }} />
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            Team
          </p>
        </div>
        <button
          onClick={() => navigate({ to: '/team' })}
          className="flex items-center gap-1 text-xs font-medium"
          style={{ color: '#4583C4' }}
        >
          Manage <ArrowUpRight size={11} />
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {members.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              No team members yet
            </p>
            <button
              onClick={() => navigate({ to: '/team/invite' })}
              className="text-xs font-medium mt-2"
              style={{ color: '#4583C4' }}
            >
              Invite someone
            </button>
          </div>
        ) : (
          members.slice(0, 5).map((m) => (
            <div key={m.id} className="flex items-center gap-3 px-2 py-1.5">
              <div
                className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
              >
                {(m.fullName || m.email || '?').charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
                  {m.fullName || m.email}
                </p>
                <p className="text-[10px] truncate" style={{ color: 'var(--dash-text-dim)' }}>
                  {m.role}
                </p>
              </div>
              <span
                className={`h-2 w-2 rounded-full shrink-0`}
                style={{ background: m.isActive ? '#10b981' : '#ef4444' }}
              />
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default TeamWidget;
