import React, { useMemo } from 'react';
import { Network } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useStaff } from '../staff/useStaff';
import { Card, SectionHeader, EmptyState, Skeleton } from '../../shared/components/ui';
import OrgNode, { OrgTreeNode } from '../../shared/components/OrgNode';
import { StaffMember } from '../../shared/services/hrApi';

const OrgChart: React.FC = () => {
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
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader title="Org Chart" subtitle="Reporting hierarchy visualized" />
      {isLoading ? (
        <Skeleton height={260} />
      ) : !root ? (
        <EmptyState
          icon={<Network size={20} />}
          title="No employees"
          description="Add employees to see the org chart."
        />
      ) : (
        <Card style={{ overflowX: 'auto', padding: 32 }}>
          <div style={{ display: 'flex', justifyContent: 'center' }}>
            <OrgNode node={root} />
          </div>
        </Card>
      )}
    </div>
  );
};

export default OrgChart;
