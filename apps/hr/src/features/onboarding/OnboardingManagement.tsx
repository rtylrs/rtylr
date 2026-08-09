import React from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Rocket, AlertTriangle } from 'lucide-react';
import { useGetEmployeeChecklist } from '../../shared/hooks/queries/useOnboarding';
import type { EmployeeChecklistPayload } from '@voxire/voxie/hr/onboarding/v1/types';
import {
  SectionHeader,
  Card,
  EmptyState,
  Skeleton,
  Pill,
  Button,
  ACCENT,
} from '../../shared/components/ui';

const OnboardingManagement = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data, isLoading } = useGetEmployeeChecklist();
  const checklists = data?.checklists ?? [];

  const getProgress = (checklist: EmployeeChecklistPayload): number => {
    const items = checklist.items ?? [];
    const total = items.length;
    if (total === 0) return 0;
    const done = items.filter((i) => i.status === 'complete').length;
    return Math.round((done / total) * 100);
  };

  const getOverdueCount = (checklist: EmployeeChecklistPayload): number => {
    return (checklist.items ?? []).filter(
      (i) => i.status !== 'complete' && new Date(i.dueDate ?? 0).getTime() < Date.now()
    ).length;
  };

  return (
    <>
      <HelmetContainer title="Onboarding management · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('onboardingMgmt.management.title')}
          subtitle={t('onboardingMgmt.management.subtitle')}
          right={
            <Button variant="secondary" onClick={() => navigate({ to: '/onboarding/templates' })}>
              {t('onboardingMgmt.management.manageTemplates')}
            </Button>
          }
        />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Skeleton height={100} />
            <Skeleton height={100} />
          </div>
        ) : checklists.length === 0 ? (
          <EmptyState
            icon={<Rocket size={20} />}
            title={t('onboardingMgmt.management.emptyTitle')}
            description={t('onboardingMgmt.management.emptyDescription')}
          />
        ) : (
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            <div style={{ overflowX: 'auto' }}>
              <table
                style={{ width: '100%', minWidth: 600, borderCollapse: 'collapse', fontSize: 12 }}
              >
                <thead>
                  <tr style={{ background: 'var(--hr-hover)' }}>
                    {[
                      t('onboardingMgmt.management.table.employee'),
                      t('onboardingMgmt.management.table.progress'),
                      t('onboardingMgmt.management.table.overdue'),
                      t('onboardingMgmt.management.table.triggeredAt'),
                    ].map((h) => (
                      <th
                        key={h}
                        style={{
                          textAlign: 'left',
                          padding: '10px 12px',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                          fontWeight: 600,
                        }}
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {checklists.map((cl) => {
                    const pct = getProgress(cl);
                    const overdue = getOverdueCount(cl);
                    return (
                      <tr key={cl.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                        <td
                          style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                        >
                          {t('onboardingMgmt.management.employeeId', { id: cl.employeeId })}
                        </td>
                        <td style={{ padding: '10px 12px' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <div
                              style={{
                                flex: 1,
                                height: 4,
                                borderRadius: 999,
                                background: 'var(--hr-hover)',
                                maxWidth: 120,
                              }}
                            >
                              <div
                                style={{
                                  width: `${pct}%`,
                                  height: '100%',
                                  background: ACCENT,
                                  borderRadius: 999,
                                }}
                              />
                            </div>
                            <span
                              style={{ fontSize: 11, fontWeight: 600, color: 'var(--hr-text)' }}
                            >
                              {pct}%
                            </span>
                          </div>
                        </td>
                        <td style={{ padding: '10px 12px' }}>
                          {overdue > 0 ? (
                            <Pill
                              style={{
                                background:
                                  'color-mix(in srgb, var(--hr-status-overdue) 13%, transparent)',
                                color: 'var(--hr-status-overdue)',
                                border:
                                  '1px solid color-mix(in srgb, var(--hr-status-overdue) 20%, transparent)',
                              }}
                            >
                              <AlertTriangle size={10} style={{ marginRight: 4 }} />
                              {overdue}
                            </Pill>
                          ) : (
                            <span style={{ color: 'var(--hr-text-muted)' }}>-</span>
                          )}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {new Date(Number(cl.triggeredAt ?? 0) * 1000).toLocaleDateString()}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </Card>
        )}
      </div>
    </>
  );
};

export default OnboardingManagement;
