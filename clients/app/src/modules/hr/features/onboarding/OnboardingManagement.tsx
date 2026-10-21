import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Rocket, AlertTriangle } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListEmployeeChecklistSummaries } from '../../shared/hooks/queries/useOnboarding';
import { SectionHeader, Card, Skeleton, Pill, Button, ACCENT } from '../../shared/components/ui';

const PAGE_SIZE = 20;

const OnboardingManagement = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const [offset, setOffset] = useState(0);
  const { data, isLoading } = useListEmployeeChecklistSummaries({
    limit: PAGE_SIZE,
    offset,
  });
  const summaries = data?.summaries ?? [];
  const total = data?.total ?? 0;

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.onboardingManagement'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrOnboardingMgmt.management.title')}
          subtitle={t('hrOnboardingMgmt.management.subtitle')}
          right={
            <Button
              variant="secondary"
              onClick={() => navigate({ to: '/hr/onboarding/templates' })}
            >
              {t('hrOnboardingMgmt.management.manageTemplates')}
            </Button>
          }
        />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Skeleton height={100} />
            <Skeleton height={100} />
          </div>
        ) : summaries.length === 0 ? (
          <EmptyState
            icon={Rocket}
            title={t('hrOnboardingMgmt.management.emptyTitle')}
            description={t('hrOnboardingMgmt.management.emptyDescription')}
            action={{
              label: t('hrOnboardingMgmt.management.manageTemplates'),
              onClick: () => navigate({ to: '/hr/onboarding/templates' }),
            }}
          />
        ) : (
          <>
            <Card style={{ padding: 0, overflow: 'hidden' }}>
              <div style={{ overflowX: 'auto' }}>
                <table
                  style={{
                    width: '100%',
                    minWidth: 600,
                    borderCollapse: 'collapse',
                    fontSize: 12,
                  }}
                >
                  <thead>
                    <tr style={{ background: 'var(--hr-hover)' }}>
                      {[
                        t('hrOnboardingMgmt.management.table.employee'),
                        t('hrOnboardingMgmt.management.table.progress'),
                        t('hrOnboardingMgmt.management.table.overdue'),
                        t('hrOnboardingMgmt.management.table.triggeredAt'),
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
                    {summaries.map((summary) => (
                      <tr
                        key={summary.checklistId}
                        style={{ borderTop: '1px solid var(--hr-border)' }}
                      >
                        <td
                          style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                        >
                          {t('hrOnboardingMgmt.management.employeeId', { id: summary.employeeId })}
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
                                  width: `${summary.completionPercentage}%`,
                                  height: '100%',
                                  background: ACCENT,
                                  borderRadius: 999,
                                }}
                              />
                            </div>
                            <span
                              style={{ fontSize: 11, fontWeight: 600, color: 'var(--hr-text)' }}
                            >
                              {summary.completionPercentage}%
                            </span>
                          </div>
                        </td>
                        <td style={{ padding: '10px 12px' }}>
                          {summary.overdueItemsCount > 0 ? (
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
                              {summary.overdueItemsCount}
                            </Pill>
                          ) : (
                            <span style={{ color: 'var(--hr-text-muted)' }}>-</span>
                          )}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {new Date(summary.triggeredAt * 1000).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>

            {total > PAGE_SIZE && (
              <div
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 16 }}
              >
                <Button
                  variant="secondary"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                >
                  {t('hrCommon.back')}
                </Button>
                <span style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
                  {t('hrOnboardingMgmt.management.pagination.range', {
                    from: offset + 1,
                    to: Math.min(offset + PAGE_SIZE, total),
                    total,
                  })}
                </span>
                <Button
                  variant="secondary"
                  disabled={offset + PAGE_SIZE >= total}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                >
                  {t('hrCommon.next')}
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </>
  );
};

export default OnboardingManagement;
