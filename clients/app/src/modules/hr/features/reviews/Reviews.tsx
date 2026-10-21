import React from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Plus, ClipboardList } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useReviews } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Button, Skeleton, Pill } from '../../shared/components/ui';
import RatingStars from '../../shared/components/RatingStars';

const STATUS_COLORS: Record<string, string> = {
  Draft: '#a1a1aa',
  'In Progress': '#f59e0b',
  Submitted: '#3b82f6',
  Completed: '#10b981',
};

const Reviews: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const { data, isLoading } = useReviews();
  return (
    <>
      <HelmetContainer title={t('hrReviews.page.metaTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrReviews.page.title')}
          subtitle={t('hrReviews.page.subtitle')}
          right={
            <Link to="/hr/reviews/new" style={{ textDecoration: 'none' }}>
              <Button variant="primary" icon={<Plus size={13} />}>
                {t('hrReviews.page.newReview')}
              </Button>
            </Link>
          }
        />
        {isLoading ? (
          <Skeleton height={200} />
        ) : !data || data.length === 0 ? (
          <EmptyState
            icon={ClipboardList}
            title={t('hrReviews.emptyTitle')}
            description={t('hrReviews.emptyDescription')}
            action={{
              label: t('hrReviews.newReview'),
              onClick: () => navigate({ to: '/hr/reviews/new' }),
            }}
          />
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
              gap: 12,
            }}
          >
            {data.map((r) => (
              <Card key={r.id} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div
                  style={{ display: 'flex', alignItems: 'start', justifyContent: 'space-between' }}
                >
                  <div>
                    <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {r.staffName}
                    </div>
                    <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>
                      {r.period} • reviewed by {r.reviewer}
                    </div>
                  </div>
                  <Pill color={STATUS_COLORS[r.status]}>{r.status}</Pill>
                </div>
                <div>
                  <RatingStars value={r.overall} readOnly size={16} />
                </div>
                <div
                  style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(80px, 1fr))',
                    gap: 6,
                    fontSize: 10,
                  }}
                >
                  <div>
                    <span style={{ color: 'var(--hr-text-muted)' }}>
                      {t('hrReviews.page.goals')}
                    </span>{' '}
                    <strong style={{ color: 'var(--hr-text)' }}>{r.goals}/5</strong>
                  </div>
                  <div>
                    <span style={{ color: 'var(--hr-text-muted)' }}>
                      {t('hrReviews.page.values')}
                    </span>{' '}
                    <strong style={{ color: 'var(--hr-text)' }}>{r.values}/5</strong>
                  </div>
                  <div>
                    <span style={{ color: 'var(--hr-text-muted)' }}>
                      {t('hrReviews.page.skills')}
                    </span>{' '}
                    <strong style={{ color: 'var(--hr-text)' }}>{r.skills}/5</strong>
                  </div>
                </div>
                {r.notes && (
                  <p
                    style={{
                      margin: 0,
                      fontSize: 11,
                      color: 'var(--hr-text-muted)',
                      fontStyle: 'italic',
                    }}
                  >
                    "{r.notes}"
                  </p>
                )}
              </Card>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default Reviews;
