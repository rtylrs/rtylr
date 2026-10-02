import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { BarChart3 } from 'lucide-react';
import { ErrorBoundary, EmptyState } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import {
  SectionHeader,
  Button,
  Card,
  Pill,
  Skeleton,
  ConfirmDialog,
} from '../../shared/components/ui';
import { useListReviewCycles, useOpenReviewCycle } from '../../shared/hooks/queries/usePerformance';
import CreateCycleForm from './CreateCycleForm';
import type { ReviewCyclePayload } from '@voxire/voxie/hr/performance/v1/types';

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-status-draft)',
  open: 'var(--hr-status-approved)',
  in_review: 'var(--hr-status-submitted)',
  completed: 'var(--hr-accent)',
};

function CycleCard({ cycle }: { cycle: ReviewCyclePayload }): React.JSX.Element {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const openCycle = useOpenReviewCycle();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const handleOpen = (): void => {
    openCycle.mutate({ cycleId: cycle.id }, { onSuccess: () => setConfirmOpen(false) });
  };

  return (
    <>
      <Card
        style={{ cursor: 'pointer' }}
        onClick={() =>
          navigate({ to: '/hr/performance/cycles/$id', params: { id: String(cycle.id) } })
        }
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hr-text)' }}>
            {cycle.name}
          </h3>
          <Pill color={STATUS_COLORS[cycle.status ?? ''] ?? 'var(--hr-status-draft)'}>
            {cycle.status}
          </Pill>
        </div>

        <p style={{ margin: '8px 0 0', fontSize: 12, color: 'var(--hr-text-muted)' }}>
          {cycle.periodStart}, {cycle.periodEnd}
        </p>

        {cycle.status === 'draft' && (
          <Button
            variant="primary"
            style={{ marginTop: 12 }}
            onClick={(e) => {
              e.stopPropagation();
              setConfirmOpen(true);
            }}
          >
            {t('hrPerformance.openCycle')}
          </Button>
        )}
      </Card>

      <ConfirmDialog
        open={confirmOpen}
        title={t('hrPerformance.confirmOpen.title')}
        description={t('hrPerformance.confirmOpen.description')}
        confirmLabel={t('hrPerformance.confirmOpen.confirm')}
        variant="danger"
        onConfirm={handleOpen}
        onCancel={() => setConfirmOpen(false)}
      />
    </>
  );
}

const PerformanceInner = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [createOpen, setCreateOpen] = useState(false);
  const { data, isLoading } = useListReviewCycles();

  const cycles = data?.cycles ?? [];

  return (
    <div>
      <SectionHeader
        title={t('hrPerformance.title')}
        subtitle={t('hrPerformance.subtitle')}
        right={
          <Button variant="primary" onClick={() => setCreateOpen(true)}>
            {t('hrPerformance.newCycle')}
          </Button>
        }
      />

      {isLoading ? (
        <div
          style={{
            display: 'grid',
            gap: 12,
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            marginTop: 16,
          }}
        >
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} height={120} />
          ))}
        </div>
      ) : cycles.length === 0 ? (
        <EmptyState
          icon={BarChart3}
          title={t('hrPerformance.emptyTitle')}
          description={t('hrPerformance.emptyDescription')}
          action={{ label: t('hrPerformance.newCycle'), onClick: () => setCreateOpen(true) }}
        />
      ) : (
        <div
          style={{
            display: 'grid',
            gap: 12,
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            marginTop: 16,
          }}
        >
          {cycles.map((cycle) => (
            <CycleCard key={cycle.id} cycle={cycle} />
          ))}
        </div>
      )}

      <CreateCycleForm open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  );
};

const Performance = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  return (
    <>
      <HelmetContainer title={pageTitle(t('hrNav.performance'))} />
      <ErrorBoundary>
        <PerformanceInner />
      </ErrorBoundary>
    </>
  );
};

export default Performance;
