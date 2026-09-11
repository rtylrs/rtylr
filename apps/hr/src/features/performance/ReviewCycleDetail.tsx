import React, { useCallback, useMemo, useRef, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { Share2, ClipboardList, Target } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import type {
  PerformanceReviewPayload,
  ReviewGoalPayload,
} from '@voxire/voxie/hr/performance/v1/types';
import {
  useGetReviewCycle,
  useListReviews,
  useListGoals,
  useShareReview,
} from '../../shared/hooks/queries/usePerformance';
import { SectionHeader, Tabs, Pill, Button, Card, Skeleton } from '../../shared/components/ui';

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-status-draft)',
  submitted: 'var(--hr-status-submitted)',
  completed: 'var(--hr-status-approved)',
  shared: 'var(--hr-accent)',
  active: 'var(--hr-status-approved)',
  achieved: 'var(--hr-accent)',
  missed: 'var(--hr-error)',
};

const REVIEW_TYPE_COLORS: Record<string, string> = {
  self: 'var(--hr-shift-blue)',
  manager: 'var(--hr-status-submitted)',
  peer: 'var(--hr-status-approved)',
};

/* ── Virtualised reviews table ── */

function ReviewsTab({ cycleId }: { cycleId: number }): React.JSX.Element {
  const { t } = useTranslation();
  const [statusFilter, setStatusFilter] = useState<string>('');
  const shareMutation = useShareReview();

  const params = useMemo(
    () => ({ reviewCycleId: cycleId, ...(statusFilter ? { status: statusFilter } : {}) }),
    [cycleId, statusFilter]
  );
  const { data, isLoading, refetch } = useListReviews(params);
  const reviews = data?.reviews ?? [];

  const parentRef = useRef<HTMLDivElement>(null);
  const virtualizer = useVirtualizer({
    count: reviews.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 52,
    overscan: 10,
  });

  const handleShare = useCallback(
    (reviewId: number): void => {
      shareMutation.mutate(
        { reviewId },
        {
          onSuccess: () => toast.success('Review shared with employee'),
          onError: () => toast.error('Failed to share review'),
        }
      );
    },
    [shareMutation]
  );

  const canShare = useCallback(
    (r: PerformanceReviewPayload): boolean => r.status === 'submitted' && !r.sharedWithEmployeeAt,
    []
  );

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} height={48} />
        ))}
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12, flexWrap: 'wrap' }}>
        {['', 'draft', 'submitted', 'completed', 'shared'].map((s) => (
          <Button
            key={s}
            variant={statusFilter === s ? 'primary' : 'ghost'}
            onClick={() => setStatusFilter(s)}
            style={{ fontSize: 11 }}
          >
            {s === ''
              ? t('common.filter') + ': ' + t('performance.filters.all')
              : t(`performance.statuses.${s}`)}
          </Button>
        ))}
      </div>

      {reviews.length === 0 ? (
        <EmptyState
          icon={ClipboardList}
          title={t('performance.reviews.emptyTitle')}
          description={t('performance.reviews.emptyDescription')}
          action={
            statusFilter !== ''
              ? { label: t('common.clearFilters'), onClick: () => setStatusFilter('') }
              : { label: t('common.refresh'), onClick: () => refetch() }
          }
        />
      ) : (
        <div
          ref={parentRef}
          style={{
            maxHeight: 480,
            overflowY: 'auto',
            border: '1px solid var(--hr-border)',
            borderRadius: 10,
            background: 'var(--hr-card)',
          }}
        >
          {/* Header row */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr 100px 80px 80px 80px',
              gap: 8,
              padding: '10px 14px',
              fontSize: 10,
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--hr-text-dim)',
              borderBottom: '1px solid var(--hr-border)',
              position: 'sticky',
              top: 0,
              background: 'var(--hr-card)',
              zIndex: 1,
            }}
          >
            <span>{t('performance.reviews.reviewee')}</span>
            <span>{t('performance.reviews.reviewer')}</span>
            <span>{t('performance.reviews.type')}</span>
            <span>{t('common.status')}</span>
            <span>{t('performance.reviews.rating')}</span>
            <span>{t('common.actions')}</span>
          </div>

          {/* Virtualised rows */}
          <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
            {virtualizer.getVirtualItems().map((virtualRow) => {
              const review = reviews[virtualRow.index];
              return (
                <div
                  key={review.id}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: virtualRow.size,
                    transform: `translateY(${virtualRow.start}px)`,
                    display: 'grid',
                    gridTemplateColumns: '1fr 1fr 100px 80px 80px 80px',
                    gap: 8,
                    padding: '0 14px',
                    alignItems: 'center',
                    fontSize: 12,
                    color: 'var(--hr-text)',
                    borderBottom: '1px solid var(--hr-border)',
                  }}
                >
                  <span>#{review.employeeId}</span>
                  <span>#{review.reviewerEmployeeId}</span>
                  <Pill
                    color={REVIEW_TYPE_COLORS[review.reviewType ?? ''] ?? 'var(--hr-status-draft)'}
                  >
                    {t(`performance.reviewTypes.${review.reviewType}`)}
                  </Pill>
                  <Pill color={STATUS_COLORS[review.status ?? ''] ?? 'var(--hr-status-draft)'}>
                    {t(`performance.statuses.${review.status}`)}
                  </Pill>
                  <span style={{ color: 'var(--hr-text-muted)' }}>
                    {review.overallRating ?? '-'}
                  </span>
                  <span>
                    {canShare(review) && (
                      <Button
                        variant="ghost"
                        onClick={() => handleShare(Number(review.id ?? 0))}
                        aria-label={t('performance.reviews.share')}
                        style={{ padding: '4px 8px' }}
                      >
                        <Share2 size={13} />
                      </Button>
                    )}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

/* ── Goals tab ── */

function GoalsTab({ cycleId }: { cycleId: number }): React.JSX.Element {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useListGoals({ reviewCycleId: String(cycleId) });
  const goals = data?.goals ?? [];

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} height={80} />
        ))}
      </div>
    );
  }

  if (goals.length === 0) {
    return (
      <EmptyState
        icon={Target}
        title={t('performance.goals.emptyTitle')}
        description={t('performance.goals.emptyDescription')}
        action={{ label: t('common.refresh'), onClick: () => refetch() }}
      />
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      {goals.map((goal: ReviewGoalPayload) => (
        <Card key={goal.id}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <h4 style={{ margin: 0, fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
              {goal.title}
            </h4>
            <Pill color={STATUS_COLORS[goal.status ?? ''] ?? 'var(--hr-status-draft)'}>
              {goal.status}
            </Pill>
          </div>
          {goal.description && (
            <p style={{ margin: '6px 0 0', fontSize: 12, color: 'var(--hr-text-muted)' }}>
              {goal.description}
            </p>
          )}
          {goal.weight != null && (
            <p style={{ margin: '4px 0 0', fontSize: 11, color: 'var(--hr-text-dim)' }}>
              {t('performance.goals.weight')}: {goal.weight}%
            </p>
          )}
        </Card>
      ))}
    </div>
  );
}

/* ── Main component ── */

const ReviewCycleDetail = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { id } = useParams({ strict: false });
  const cycleId = id ? Number(id) : undefined;
  const { data: cycle, isLoading } = useGetReviewCycle(cycleId);

  const reviewsLabel = t('performance.tabs.reviews');
  const goalsLabel = t('performance.tabs.goals');
  const [activeTab, setActiveTab] = useState(reviewsLabel);

  if (isLoading || !cycle) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={240} />
        <Skeleton height={16} width={180} />
        <Skeleton height={300} />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={pageTitle(t('common.pageTitles.reviewCycle'))} />
      <div>
        <SectionHeader
          title={cycle.name ?? ''}
          subtitle={`${cycle.periodStart}, ${cycle.periodEnd}`}
          right={
            <Pill color={STATUS_COLORS[cycle.status ?? ''] ?? 'var(--hr-status-draft)'}>
              {cycle.status}
            </Pill>
          }
        />

        <Tabs tabs={[reviewsLabel, goalsLabel]} active={activeTab} onChange={setActiveTab} />

        {activeTab === reviewsLabel ? (
          <ReviewsTab cycleId={cycleId ?? 0} />
        ) : (
          <GoalsTab cycleId={cycleId ?? 0} />
        )}
      </div>
    </>
  );
};

export default ReviewCycleDetail;
