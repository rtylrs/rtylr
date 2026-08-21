import React from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Clock, Check } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useTimesheets, useApproveTimesheet } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Button, Skeleton } from '../../shared/components/ui';

const DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;

const Timesheets: React.FC = () => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useTimesheets();
  const approve = useApproveTimesheet();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title="Timesheets" subtitle="Review clocked hours per week" />
      {isLoading ? (
        <Skeleton height={200} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon={Clock}
          title={t('timesheets.emptyTitle')}
          description={t('timesheets.managerViewEmptyDescription')}
          action={{ label: t('common.refresh'), onClick: () => refetch() }}
        />
      ) : (
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
            <thead>
              <tr style={{ background: 'var(--hr-hover)' }}>
                <th
                  style={{
                    textAlign: 'left',
                    padding: '10px 12px',
                    color: 'var(--hr-text-muted)',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    fontSize: 10,
                  }}
                >
                  Employee
                </th>
                <th
                  style={{
                    textAlign: 'left',
                    padding: '10px 12px',
                    color: 'var(--hr-text-muted)',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    fontSize: 10,
                  }}
                >
                  Week of
                </th>
                {DAYS.map((d) => (
                  <th
                    key={d}
                    style={{
                      textAlign: 'center',
                      padding: '10px 8px',
                      color: 'var(--hr-text-muted)',
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      fontSize: 10,
                    }}
                  >
                    {d.toUpperCase()}
                  </th>
                ))}
                <th
                  style={{
                    textAlign: 'right',
                    padding: '10px 12px',
                    color: 'var(--hr-text-muted)',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    fontSize: 10,
                  }}
                >
                  Total
                </th>
                <th
                  style={{
                    textAlign: 'right',
                    padding: '10px 12px',
                    color: 'var(--hr-text-muted)',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    fontSize: 10,
                  }}
                ></th>
              </tr>
            </thead>
            <tbody>
              {data.map((row) => {
                const total = DAYS.reduce((sum, d) => sum + (row.hours[d] ?? 0), 0);
                return (
                  <tr key={row.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                    <td style={{ padding: '12px', color: 'var(--hr-text)', fontWeight: 600 }}>
                      {row.staffName}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--hr-text-muted)' }}>{row.weekOf}</td>
                    {DAYS.map((d) => (
                      <td
                        key={d}
                        style={{
                          padding: '12px 8px',
                          textAlign: 'center',
                          color: 'var(--hr-text)',
                        }}
                      >
                        {row.hours[d] || '-'}
                      </td>
                    ))}
                    <td
                      style={{
                        padding: '12px',
                        textAlign: 'right',
                        color: '#8b5cf6',
                        fontWeight: 700,
                      }}
                    >
                      {total}h
                    </td>
                    <td style={{ padding: '12px', textAlign: 'right' }}>
                      {row.approved ? (
                        <span
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 4,
                            padding: '4px 10px',
                            borderRadius: 6,
                            background: 'rgba(16,185,129,0.12)',
                            color: '#10b981',
                            fontSize: 10,
                            fontWeight: 600,
                          }}
                        >
                          <Check size={11} /> Approved
                        </span>
                      ) : (
                        <Button
                          variant="primary"
                          style={{ padding: '4px 10px', fontSize: 10 }}
                          onClick={() =>
                            approve.mutate(row.id, {
                              onSuccess: () => toast.success('Timesheet approved'),
                            })
                          }
                        >
                          Approve
                        </Button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
};

export default Timesheets;
