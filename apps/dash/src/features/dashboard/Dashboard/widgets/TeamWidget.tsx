import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { EmptyState } from '@voxire/ui';
import { useTeam } from '../../../../shared/hooks/queries/useTeam';
import WidgetHeader from './WidgetHeader';

const TeamWidget: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data } = useTeam({ pageSize: 50 });
  const members = data?.members ?? [];

  return (
    <div className="flex flex-col h-full">
      <WidgetHeader
        icon={Users}
        title="Team"
        action={{ label: 'Manage', onClick: () => navigate({ to: '/team' }) }}
      />
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {members.length === 0 ? (
          <EmptyState
            icon={Users}
            title={t('dashboard.teamWidget.emptyTitle')}
            action={{
              label: t('dashboard.teamWidget.addMember'),
              onClick: () => navigate({ to: '/team/invite' }),
            }}
            size="compact"
          />
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
