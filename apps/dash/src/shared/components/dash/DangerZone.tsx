import React from 'react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle } from 'lucide-react';
import { Button } from '../ui';

export interface DangerAction {
  label: string;
  description: string;
  buttonLabel: string;
  onClick: () => void;
}

export interface DangerZoneProps {
  actions: DangerAction[];
}

export const DangerZone: React.FC<DangerZoneProps> = ({ actions }) => {
  const { t } = useTranslation();
  return (
    <div
      className="rounded-xl"
      style={{
        background: 'var(--status-cancelled-bg)',
        border: '1px solid var(--status-cancelled)',
      }}
    >
      <div
        className="px-5 py-4 flex items-center gap-2"
        style={{ borderBottom: '1px solid var(--status-cancelled-bg)' }}
      >
        <AlertTriangle size={16} style={{ color: 'var(--status-cancelled)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--status-cancelled)' }}>
          {t('common.dangerZone')}
        </p>
      </div>
      <div className="divide-y" style={{ borderColor: 'var(--status-cancelled-bg)' }}>
        {actions.map((a, i) => (
          <div
            key={i}
            className="flex items-center justify-between gap-4 px-5 py-4"
            style={{ borderTop: i > 0 ? '1px solid var(--status-cancelled-bg)' : undefined }}
          >
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {a.label}
              </p>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
                {a.description}
              </p>
            </div>
            <Button variant="destructive" size="sm" onClick={a.onClick} className="shrink-0">
              {a.buttonLabel}
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DangerZone;
