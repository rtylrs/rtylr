import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { MessageSquare, Star } from 'lucide-react';
import { useListFeedback } from '../../shared/hooks/queries/useFeedback';
import { PageHeader, Card, EmptyState, SkeletonList } from '../../shared/components/ui';

const Feedback: React.FC = () => {
  const { data: fetched, isLoading, isError } = useListFeedback();
  const feedback = fetched ?? [];
  const metricsUnavailable = isLoading || isError;

  const npsResponses = feedback.filter((f) => f.type === 'nps');
  const reviews = feedback.filter((f) => f.type === 'review');

  const promoters = npsResponses.filter((f) => f.score >= 9).length;
  const detractors = npsResponses.filter((f) => f.score < 7).length;
  const npsTotal = npsResponses.length || 1;
  const nps = Math.round(((promoters - detractors) / npsTotal) * 100);
  const avgRating =
    reviews.length > 0 ? reviews.reduce((s, f) => s + f.score, 0) / reviews.length : 0;

  return (
    <>
      <HelmetContainer title="Feedback · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title="Feedback and NPS"
          subtitle={
            isLoading
              ? 'Loading responses…'
              : isError
                ? 'Feedback is unavailable'
                : feedback.length + ' responses'
          }
        />

        <div
          className="crm-stats-grid"
          style={{ display: 'grid', gap: '12px', marginBottom: '20px' }}
        >
          <Card>
            <div
              style={{
                fontSize: '10px',
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                fontWeight: 700,
              }}
            >
              NPS Score
            </div>
            <div
              style={{
                fontSize: '28px',
                fontWeight: 700,
                color: nps >= 50 ? '#10b981' : nps >= 0 ? '#f59e0b' : '#ef4444',
                marginTop: '4px',
              }}
            >
              {metricsUnavailable ? '—' : npsResponses.length > 0 ? nps : '—'}
            </div>
          </Card>
          <Card>
            <div
              style={{
                fontSize: '10px',
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                fontWeight: 700,
              }}
            >
              Avg Rating
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginTop: '4px' }}>
              <span style={{ fontSize: '28px', fontWeight: 700, color: 'var(--hub-text)' }}>
                {metricsUnavailable ? '—' : reviews.length > 0 ? avgRating.toFixed(1) : '—'}
              </span>
              <Star size={18} style={{ color: '#f59e0b', fill: '#f59e0b' }} />
            </div>
          </Card>
          <Card>
            <div
              style={{
                fontSize: '10px',
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                fontWeight: 700,
              }}
            >
              Promoters
            </div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: '#10b981', marginTop: '4px' }}>
              {metricsUnavailable ? '—' : promoters}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {metricsUnavailable
                ? 'Loading…'
                : npsResponses.length > 0
                  ? Math.round((promoters / npsTotal) * 100) + '% of NPS responses'
                  : 'No NPS responses'}
            </div>
          </Card>
          <Card>
            <div
              style={{
                fontSize: '10px',
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                fontWeight: 700,
              }}
            >
              Detractors
            </div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: '#ef4444', marginTop: '4px' }}>
              {metricsUnavailable ? '—' : detractors}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {metricsUnavailable
                ? 'Loading…'
                : npsResponses.length > 0
                  ? Math.round((detractors / npsTotal) * 100) + '% of NPS responses'
                  : 'No NPS responses'}
            </div>
          </Card>
        </div>

        <Card padded={false}>
          <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--hub-border)' }}>
            <h4
              style={{
                fontSize: '12px',
                fontWeight: 600,
                color: 'var(--hub-text)',
                margin: 0,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              Recent feedback
            </h4>
          </div>
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
          ) : isError ? (
            <EmptyState
              icon={MessageSquare}
              title="Unable to load feedback"
              description="Check your connection and try again."
            />
          ) : feedback.length === 0 ? (
            <EmptyState
              icon={MessageSquare}
              title="No feedback yet"
              description="Feedback will appear here once collected."
            />
          ) : (
            <div>
              {feedback.map((f, i) => (
                <div
                  key={f.id}
                  style={{
                    padding: '14px 16px',
                    borderBottom: i < feedback.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '10px',
                      marginBottom: '6px',
                    }}
                  >
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {f.customerName}
                    </div>
                    {f.type === 'review' ? (
                      <div style={{ display: 'flex', gap: '1px' }}>
                        {Array.from({ length: 5 }).map((_, idx) => (
                          <Star
                            key={idx}
                            size={12}
                            style={{
                              color: idx < f.score ? '#f59e0b' : 'var(--hub-text-dim)',
                              fill: idx < f.score ? '#f59e0b' : 'none',
                            }}
                          />
                        ))}
                      </div>
                    ) : (
                      <span
                        style={{
                          padding: '2px 8px',
                          borderRadius: '9999px',
                          fontSize: '10px',
                          fontWeight: 600,
                          background:
                            f.score >= 9
                              ? 'rgba(16,185,129,0.12)'
                              : f.score >= 7
                                ? 'rgba(245,158,11,0.12)'
                                : 'rgba(239,68,68,0.12)',
                          color: f.score >= 9 ? '#10b981' : f.score >= 7 ? '#f59e0b' : '#ef4444',
                        }}
                      >
                        NPS {f.score}
                      </span>
                    )}
                    <span
                      style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginLeft: 'auto' }}
                    >
                      {f.createdAt}
                    </span>
                  </div>
                  <p
                    style={{
                      fontSize: '13px',
                      color: 'var(--hub-text)',
                      margin: '4px 0 0',
                      lineHeight: 1.6,
                    }}
                  >
                    {f.comment}
                  </p>
                  <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginTop: '4px' }}>
                    via {f.source || 'unknown source'}
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </>
  );
};

export default Feedback;
