import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Wallet } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListLeaveBalances, useListLeaveTypes } from '../../shared/hooks/queries/useLeave';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { Card, Skeleton } from '../../shared/components/ui';

const LeaveBalances = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const { data: myEmployeeId } = useMyEmployeeId();
  const employeeId = myEmployeeId !== undefined ? String(myEmployeeId) : undefined;
  const { data: balances, isLoading, refetch } = useListLeaveBalances(employeeId);
  const { data: leaveTypes } = useListLeaveTypes();

  const leaveTypesById = useMemo(() => {
    const map = new Map<string, string>();
    (leaveTypes ?? []).forEach((lt) => {
      if (lt.name) {
        map.set(String(lt.id), lt.name);
      }
    });
    return map;
  }, [leaveTypes]);

  if (isLoading) {
    return (
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
          gap: 12,
        }}
      >
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} height={120} />
        ))}
      </div>
    );
  }

  if (!balances || balances.length === 0) {
    return (
      <EmptyState
        icon={Wallet}
        title={t('hrLeave.balances.emptyTitle')}
        description={t('hrLeave.balances.emptyDescription')}
        action={{ label: t('hrCommon.refresh'), onClick: () => refetch() }}
      />
    );
  }

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
        gap: 12,
      }}
    >
      {balances.map((balance) => {
        const typeName =
          leaveTypesById.get(String(balance.leaveTypeId)) || `#${balance.leaveTypeId}`;
        return (
          <Card key={balance.leaveTypeId}>
            <div
              style={{
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--hr-text-muted)',
                marginBottom: 8,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              {typeName}
            </div>
            <div
              style={{ fontSize: 28, fontWeight: 700, color: 'var(--hr-text)', marginBottom: 12 }}
            >
              {balance.remainingDays}
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 500,
                  color: 'var(--hr-text-muted)',
                  marginLeft: 4,
                }}
              >
                {t('hrLeave.balances.daysLeft')}
              </span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
              <BalanceRow
                label={t('hrLeave.balances.entitled')}
                value={balance.entitledDays ?? 0}
                color="var(--hr-accent)"
              />
              <BalanceRow
                label={t('hrLeave.balances.used')}
                value={balance.usedDays ?? 0}
                color="var(--hr-status-rejected)"
              />
              <BalanceRow
                label={t('hrLeave.balances.pending')}
                value={balance.pendingDays ?? 0}
                color="var(--hr-status-submitted)"
              />
            </div>
          </Card>
        );
      })}
    </div>
  );
};

const BalanceRow = ({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color: string;
}): React.JSX.Element => (
  <div
    style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 11 }}
  >
    <span style={{ color: 'var(--hr-text-muted)' }}>{label}</span>
    <span style={{ fontWeight: 600, color }}>{value}</span>
  </div>
);

export default LeaveBalances;
