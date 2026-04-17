import React from 'react';
import {
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
  ExternalLink,
} from 'lucide-react';
import { APP_URLS } from '@voxire/config';

const apps = [
  {
    id: 'pos',
    name: 'POS',
    icon: MonitorSpeaker,
    color: '#4583C4',
    href: APP_URLS.pos,
  },
  {
    id: 'erp',
    name: 'ERP',
    icon: Package,
    color: '#10b981',
    href: APP_URLS.erp,
  },
  {
    id: 'hr',
    name: 'HR',
    icon: UserCheck,
    color: '#f59e0b',
    href: APP_URLS.hr,
  },
  {
    id: 'crm',
    name: 'CRM',
    icon: Users,
    color: '#8b5cf6',
    href: APP_URLS.crm,
  },
  {
    id: 'finance',
    name: 'Finance',
    icon: CreditCard,
    color: '#ef4444',
    href: APP_URLS.finance,
  },
  {
    id: 'flow',
    name: 'Flow',
    icon: Layers,
    color: '#64C5B9',
    href: APP_URLS.flow,
  },
];

const AppsWidget: React.FC = () => {
  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3" style={{ borderBottom: '1px solid var(--dash-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Apps
        </p>
      </div>
      <div className="flex-1 grid grid-cols-6 gap-2 p-3">
        {apps.map((app) => {
          const Icon = app.icon;
          return (
            <a
              key={app.id}
              href={app.href}
              className="group flex flex-col items-center gap-2 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)]"
              style={{ textDecoration: 'none' }}
            >
              <div
                className="h-9 w-9 rounded-xl flex items-center justify-center"
                style={{ background: `${app.color}12` }}
              >
                <Icon size={18} style={{ color: app.color }} />
              </div>
              <span className="text-[11px] font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {app.name}
              </span>
            </a>
          );
        })}
      </div>
    </div>
  );
};

export default AppsWidget;
