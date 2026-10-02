import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { CheckSquare } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface ApprovalItem {
  type: string;
  count: number;
  path: string;
}

interface PendingApprovalsWidgetProps {
  data: ApprovalItem[] | null;
  isLoading: boolean;
}

const PendingApprovalsWidget = ({
  data,
  isLoading,
}: PendingApprovalsWidgetProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const total = data?.reduce((sum, a) => sum + a.count, 0) ?? 0;

  return (
    <WidgetShell
      title={t('hrDashboard.widgets.pendingApprovals')}
      icon={<CheckSquare size={14} />}
      isLoading={isLoading}
      isEmpty={!data || total === 0}
      emptyMessage={t('hrDashboard.widgets.noApprovals')}
    >
      {data && total > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {data
            .filter((a) => a.count > 0)
            .map((a) => (
              <Link
                key={a.type}
                to={a.path}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '8px 10px',
                  borderRadius: 8,
                  background: 'var(--hr-hover)',
                  textDecoration: 'none',
                  color: 'var(--hr-text)',
                  fontSize: 12,
                  fontWeight: 500,
                  transition: 'background 0.15s',
                }}
              >
                <span>{a.type}</span>
                <span
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minWidth: 20,
                    height: 20,
                    borderRadius: 10,
                    background: 'color-mix(in srgb, var(--hr-accent) 13%, transparent)',
                    color: 'var(--hr-accent)',
                    fontSize: 11,
                    fontWeight: 700,
                  }}
                >
                  {a.count}
                </span>
              </Link>
            ))}
        </div>
      )}
    </WidgetShell>
  );
};

export default PendingApprovalsWidget;
