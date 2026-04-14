import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Package, Plus, CheckCircle, XCircle } from 'lucide-react';
import {
  useListEquipmentRequests,
  useApproveEquipment,
} from '../../shared/hooks/queries/useEquipment';
import {
  SectionHeader,
  Card,
  EmptyState,
  Skeleton,
  Pill,
  Button,
  Tabs,
  ACCENT,
} from '../../shared/components/ui';
import EquipmentRequestForm from './EquipmentRequestForm';

const URGENCY_COLORS: Record<string, string> = {
  low: '#10b981',
  medium: '#f59e0b',
  high: '#ef4444',
};

const STATUS_COLORS: Record<string, string> = {
  pending: '#f59e0b',
  approved: '#10b981',
  ordered: '#3b82f6',
  delivered: '#8b5cf6',
  rejected: '#ef4444',
};

const Equipment = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [tab, setTab] = useState<string>(t('equipment.tabs.myRequests'));
  const [showForm, setShowForm] = useState(false);
  const { data, isLoading } = useListEquipmentRequests();
  const approve = useApproveEquipment();

  const tabs = [t('equipment.tabs.myRequests'), t('equipment.tabs.pendingApprovals')];
  const requests = data?.requests ?? [];

  const pendingRequests = requests.filter((r) => r.status === 'pending');
  const displayRequests = tab === tabs[1] ? pendingRequests : requests;

  const handleApproval = (id: number, approved: boolean): void => {
    approve.mutate({ equipmentRequestId: id, approved });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('equipment.title')}
        subtitle={t('equipment.subtitle')}
        right={
          <Button variant="primary" icon={<Plus size={14} />} onClick={() => setShowForm(true)}>
            {t('equipment.newRequest')}
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
          icon={<Package size={20} />}
          title={t('equipment.emptyTitle')}
          description={t('equipment.emptyDescription')}
          action={
            tab === tabs[0] ? (
              <Button variant="primary" icon={<Plus size={14} />} onClick={() => setShowForm(true)}>
                {t('equipment.newRequest')}
              </Button>
            ) : undefined
          }
        />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {displayRequests.map((req) => (
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
                  background: `${ACCENT}22`,
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
              <Pill color={URGENCY_COLORS[req.urgency] ?? '#a1a1aa'}>
                {t(`equipment.urgencies.${req.urgency}` as const)}
              </Pill>
              <Pill color={STATUS_COLORS[req.status] ?? '#a1a1aa'}>
                {t(`equipment.statuses.${req.status}` as const)}
              </Pill>
              {req.neededByDate && (
                <span style={{ fontSize: 10, color: 'var(--hr-text-muted)', flexShrink: 0 }}>
                  {t('equipment.neededBy', { date: req.neededByDate })}
                </span>
              )}
              {tab === tabs[1] && req.status === 'pending' && (
                <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                  <Button
                    variant="primary"
                    icon={<CheckCircle size={13} />}
                    onClick={() => handleApproval(req.id, true)}
                    style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
                  >
                    {t('equipment.approve')}
                  </Button>
                  <Button
                    variant="danger"
                    icon={<XCircle size={13} />}
                    onClick={() => handleApproval(req.id, false)}
                    style={{ minHeight: 30, padding: '4px 10px', fontSize: 11 }}
                  >
                    {t('equipment.reject')}
                  </Button>
                </div>
              )}
            </Card>
          ))}
        </div>
      )}

      {showForm && <EquipmentRequestForm onClose={() => setShowForm(false)} />}
    </div>
  );
};

export default Equipment;
