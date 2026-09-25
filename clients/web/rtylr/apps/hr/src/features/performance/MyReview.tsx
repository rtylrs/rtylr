import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { ClipboardList } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import type {
  ReviewGoalPayload,
  PerformanceReviewPayload,
} from '@voxire/voxie/hr/performance/v1/types';
import {
  useListReviews,
  useListGoals,
  useSubmitReview,
} from '../../shared/hooks/queries/usePerformance';
import { myReviewSchema } from '../../shared/schemas/performance';
import {
  SectionHeader,
  Card,
  Pill,
  Button,
  Label,
  Select,
  Textarea,
  Skeleton,
} from '../../shared/components/ui';

const STATUS_COLORS: Record<string, string> = {
  active: 'var(--hr-status-approved)',
  achieved: 'var(--hr-accent)',
  missed: 'var(--hr-error)',
};

function GoalContext({ goals }: { goals: ReviewGoalPayload[] }): React.JSX.Element {
  const { t } = useTranslation();

  if (goals.length === 0) return <></>;

  return (
    <div style={{ marginBottom: 20 }}>
      <h3 style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)', margin: '0 0 8px' }}>
        {t('performance.myReview.goalsContext')}
      </h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {goals.map((g) => (
          <Card key={g.id} style={{ padding: 10 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--hr-text)' }}>
                {g.title}
              </span>
              <Pill color={STATUS_COLORS[g.status ?? ''] ?? 'var(--hr-status-draft)'}>
                {g.status}
              </Pill>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

function SubmittedView({ review }: { review: PerformanceReviewPayload }): React.JSX.Element {
  const { t } = useTranslation();

  return (
    <Card>
      <h3 style={{ fontSize: 14, fontWeight: 600, color: 'var(--hr-text)', margin: '0 0 12px' }}>
        {t('performance.myReview.submittedTitle')}
      </h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div>
          <Label>{t('performance.myReview.rating')}</Label>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hr-text)' }}>{review.overallRating}</p>
        </div>
        <div>
          <Label>{t('performance.myReview.narrative')}</Label>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hr-text)', whiteSpace: 'pre-wrap' }}>
            {review.narrative}
          </p>
        </div>
      </div>
    </Card>
  );
}

const MyReview = (): React.JSX.Element => {
  const { t } = useTranslation();
  const submitMutation = useSubmitReview();

  const {
    data: reviewsData,
    isLoading: reviewsLoading,
    refetch: refetchReviews,
  } = useListReviews({
    status: undefined,
  });
  const reviews = reviewsData?.reviews ?? [];
  const selfReview = reviews.find((r) => r.reviewType === 'self');

  const { data: goalsData } = useListGoals(
    selfReview?.reviewCycleId ? { reviewCycleId: selfReview.reviewCycleId } : undefined
  );
  const goals = goalsData?.goals ?? [];

  const [rating, setRating] = useState('');
  const [narrative, setNarrative] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  /* Rating scale is config-driven; falls back to 1–5 until the backend
     exposes a performanceRatingScale field on HRCompanyConfigPayload. */
  const scaleOptions = ['1', '2', '3', '4', '5'];

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    if (!selfReview) return;

    const result = myReviewSchema.safeParse({ overallRating: rating, narrative });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = String(issue.path[0]);
        if (key === 'narrative') {
          fieldErrors[key] = t('performance.myReview.narrativeMinLength');
        } else {
          fieldErrors[key] = t('common.required');
        }
      }
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    submitMutation.mutate({
      reviewId: selfReview.id,
      overallRating: result.data.overallRating,
      narrative: result.data.narrative,
    });
  };

  if (reviewsLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={200} />
      </div>
    );
  }

  if (!selfReview) {
    return (
      <div>
        <SectionHeader title={t('performance.myReview.title')} />
        <EmptyState
          icon={ClipboardList}
          title={t('performance.myReview.emptyTitle')}
          description={t('performance.myReview.emptyDescription')}
          action={{ label: t('common.refresh'), onClick: () => refetchReviews() }}
        />
      </div>
    );
  }

  const isSubmitted =
    selfReview.status === 'submitted' ||
    selfReview.status === 'completed' ||
    selfReview.status === 'shared';

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.myReview'))} />
      <div>
        <SectionHeader
          title={t('performance.myReview.title')}
          subtitle={t('performance.myReview.subtitle')}
        />

        <GoalContext goals={goals} />

        {isSubmitted ? (
          <SubmittedView review={selfReview} />
        ) : (
          <Card>
            <form
              onSubmit={handleSubmit}
              style={{ display: 'flex', flexDirection: 'column', gap: 14 }}
            >
              <div>
                <Label htmlFor="review-rating">{t('performance.myReview.rating')}</Label>
                <Select
                  id="review-rating"
                  value={rating}
                  onChange={(e) => setRating(e.target.value)}
                >
                  <option value="">{t('performance.myReview.selectRating')}</option>
                  {scaleOptions.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt}
                    </option>
                  ))}
                </Select>
                {errors.overallRating && (
                  <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
                    {t('common.required')}
                  </p>
                )}
              </div>

              <div>
                <Label htmlFor="review-narrative">{t('performance.myReview.narrative')}</Label>
                <Textarea
                  id="review-narrative"
                  value={narrative}
                  onChange={(e) => setNarrative(e.target.value)}
                  placeholder={t('performance.myReview.narrativePlaceholder')}
                  style={{ minHeight: 140 }}
                />
                {errors.narrative && (
                  <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
                    {errors.narrative}
                  </p>
                )}
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <Button type="submit" variant="primary" disabled={submitMutation.isPending}>
                  {submitMutation.isPending
                    ? t('common.loading')
                    : t('performance.myReview.submit')}
                </Button>
              </div>
            </form>
          </Card>
        )}
      </div>
    </>
  );
};

export default MyReview;
