import React, { useMemo } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { Network } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { useStaff } from '../staff/useStaff';
import { Card, SectionHeader, Skeleton } from '../../shared/components/ui';
import OrgNode, { OrgTreeNode } from '../../shared/components/OrgNode';
import { StaffMember } from '../../shared/services/hrApi';

const OrgChart: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data, isLoading } = useStaff();

  const root = useMemo<OrgTreeNode | null>(() => {
    const staff = data ?? [];
    if (staff.length === 0) return null;
    // Simple heuristic: first staff member is the CEO/root, split others into 2 branches.
    const sorted = [...staff].sort((a, b) => a.id - b.id);
    const [ceo, ...rest] = sorted;
    const managers = rest.slice(0, Math.min(2, rest.length));
    const reports = rest.slice(managers.length);
    const children: OrgTreeNode[] = managers.map((m, i) => ({
      staff: m,
      children: reports
        .filter((_, idx) => idx % Math.max(1, managers.length) === i)
        .map<OrgTreeNode>((r: StaffMember) => ({ staff: r, children: [] })),
    }));
    return { staff: ceo, children };
  }, [data]);

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.orgChart'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('departments.orgChart.title')}
          subtitle={t('departments.orgChart.subtitle')}
        />
        {isLoading ? (
          <Skeleton height={260} />
        ) : !root ? (
          <EmptyState
            icon={Network}
            title={t('departments.orgChart.emptyTitle')}
            description={t('departments.orgChart.emptyDescription')}
            action={{
              label: t('staff.inviteStaff'),
              onClick: () => navigate({ to: '/employees/new' }),
            }}
          />
        ) : (
          <Card style={{ overflowX: 'auto', padding: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'center' }}>
              <OrgNode node={root} />
            </div>
          </Card>
        )}
      </div>
    </>
  );
};

export default OrgChart;
