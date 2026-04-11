import React from 'react';
import { useNavigate } from 'react-router-dom';
import { UserPlus, Package, FileText, BarChart3 } from 'lucide-react';

const actions = [
  { label: 'Invite team member', path: '/team/invite', icon: UserPlus },
  { label: 'Add product', path: '/apps', icon: Package },
  { label: 'Create invoice', path: '/apps', icon: FileText },
  { label: 'View analytics', path: '/analytics', icon: BarChart3 },
];

const QuickActionsWidget: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3" style={{ borderBottom: '1px solid var(--hub-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          Quick actions
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-2 space-y-0.5">
        {actions.map((a) => {
          const Icon = a.icon;
          return (
            <button
              key={a.label}
              onClick={() => navigate(a.path)}
              className="flex items-center gap-2.5 w-full px-3 py-2.5 rounded-lg text-xs font-medium transition-colors text-left hover:bg-[var(--hub-hover)]"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <Icon size={14} className="shrink-0" />
              {a.label}
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default QuickActionsWidget;
