import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Package, Plus, CheckCircle, XCircle } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListEquipmentRequests,
  useApproveEquipment,
} from '../../shared/hooks/queries/useEquipment';
import {
  SectionHeader,
  Card,
  Skeleton,
  Pill,
  Button,
  Tabs,
  ACCENT,
} from '../../shared/components/ui';
import EquipmentRequestForm from './EquipmentRequestForm';

const STATUS_COLORS: Record<string, string> = {
  pending: 'var(--hr-status-pending)',
  approved: 'var(--hr-status-approved)',
  ordered: 'var(--hr-shift-blue)',
  delivered: 'var(--hr-accent)',
  rejected: 'var(--hr-error)',
};

const pillStyle = (cssVar: string): React.CSSProperties => ({
  background: `color-mix(in srgb, ${cssVar} 13%, transparent)`,
  color: cssVar,
  border: `1px solid color-mix(in srgb, ${cssVar} 20%, transparent)`,
});

const Equipment = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [tab, setTab] = useState<string>(t('hrEquipment.tabs.myRequests'));
  const [showForm, setShowForm] = useState(false);
  const { data, isLoading, refetch } = useListEquipmentRequests();
  const approve = useApproveEquipment();

  const tabs = [t('hrEquipment.tabs.myRequests'), t('hrEquipment.tabs.pendingApprovals')];
  const requests = data?.requests ?? [];

  const pendingRequests = requests.filter((r) => r.status === 'pending');
  const displayRequests = tab === tabs[1] ? pendingRequests : requests;

  const handleApproval = (id: string, approved: boolean): void => {
    approve.mutate({ equipmentRequestId: id, body: { approved } });
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrNav.equipment'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrEquipment.title')}
          subtitle={t('hrEquipment.subtitle')}
          right={
            <Button variant="primary" icon={<Plus size={14} />} onClick={() => setShowForm(true)}>
              {t('hrEquipment.newRequest')}
            </Button>
          }
        />

        <Tabs tabs={tabs} active={tab} onChange={setTab} />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Skeleton height={80} />
            <Skeleton height={80} />
            <Skeleton height={80} />
          </div>
        ) : displayRequests.length === 0 ? (
          <EmptyState
            icon={Package}
            title={t('hrEquipment.emptyTitle')}
            description={t('hrEquipment.emptyDescription')}
            action={
              tab === tabs[0]
                ? { label: t('hrEquipment.newRequest'), onClick: () => setShowForm(true) }
                : { label: t('hrCommon.refresh'), onClick: () => refetch() }
            }
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {displayRequests.map((req) => {
              const statusColor = STATUS_COLORS[req.status ?? ''] ?? 'var(--hr-status-draft)';
              return (
                <Card
                  key={req.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    flexWrap: 'wrap',
                  }}
                >
                  <div
                    style={{
                      width: 36,
                      height: 36,
                      borderRadius: 10,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: `color-mix(in srgb, ${ACCENT} 13%, transparent)`,
                      color: ACCENT,
                      flexShrink: 0,
                    }}
                  >
                    <Package size={16} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                      {req.equipmentType}
                    </div>
                    <div
                      style={{
                        fontSize: 11,
                        color: 'var(--hr-text-muted)',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {req.description}
                    </div>
                  </div>
                  <Pill style={pillStyle(statusColor)}>
                    {t(`hrEquipment.statuses.${req.status}` as const)}
                  </Pill>
                  {req.neededByDate && (
                    <span style={{ fontSize: 10, color: 'var(--hr-text-muted)', flexShrink: 0 }}>
                      {t('hrEquipment.neededBy', { date: req.neededByDate })}
                    </span>
                  )}
                  {tab === tabs[1] && req.status === 'pending' && (
                    <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                      <Button
                        variant="primary"
                        icon={<CheckCircle size={13} />}
                        onClick={() => handleApproval(req.id ?? '', true)}
                        style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
                      >
                        {t('hrEquipment.approve')}
                      </Button>
                      <Button
                        variant="danger"
                        icon={<XCircle size={13} />}
                        onClick={() => handleApproval(req.id ?? '', false)}
                        style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
                      >
                        {t('hrEquipment.reject')}
                      </Button>
                    </div>
                  )}
                </Card>
              );
            })}
          </div>
        )}

        {showForm && <EquipmentRequestForm onClose={() => setShowForm(false)} />}
      </div>
    </>
  );
};

export default Equipment;
