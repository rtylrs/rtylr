import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { MessageSquare, Star } from 'lucide-react';
import { useListFeedback } from '../../shared/hooks/queries/useFeedback';
import { PageHeader, Card, EmptyState, SkeletonList } from '../../shared/components/ui';

const Feedback: React.FC = () => {
  const { data: fetched, isLoading } = useListFeedback();
  const feedback = fetched ?? [];

  const promoters = feedback.filter((f) => f.nps >= 9).length;
  const passives = feedback.filter((f) => f.nps >= 7 && f.nps < 9).length;
  const detractors = feedback.filter((f) => f.nps < 7).length;
  const total = feedback.length || 1;
  const nps = Math.round(((promoters - detractors) / total) * 100);
  const avgRating =
    feedback.length > 0 ? feedback.reduce((s, f) => s + f.rating, 0) / feedback.length : 0;

  return (
    <>
      <HelmetContainer title="Feedback · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader title="Feedback and NPS" subtitle={`${feedback.length} responses`} />

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: '12px',
            marginBottom: '20px',
          }}
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
              {nps}
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
                {avgRating.toFixed(1)}
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
              {promoters}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {Math.round((promoters / total) * 100)}% of responses
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
              {detractors}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {Math.round((detractors / total) * 100)}% of responses
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
              Recent reviews
            </h4>
          </div>
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
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
                      {f.customer}
                    </div>
                    <div style={{ display: 'flex', gap: '1px' }}>
                      {Array.from({ length: 5 }).map((_, idx) => (
                        <Star
                          key={idx}
                          size={12}
                          style={{
                            color: idx < f.rating ? '#f59e0b' : 'var(--hub-text-dim)',
                            fill: idx < f.rating ? '#f59e0b' : 'none',
                          }}
                        />
                      ))}
                    </div>
                    <span
                      style={{
                        padding: '2px 8px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        fontWeight: 600,
                        background:
                          f.nps >= 9
                            ? 'rgba(16,185,129,0.12)'
                            : f.nps >= 7
                              ? 'rgba(245,158,11,0.12)'
                              : 'rgba(239,68,68,0.12)',
                        color: f.nps >= 9 ? '#10b981' : f.nps >= 7 ? '#f59e0b' : '#ef4444',
                      }}
                    >
                      NPS {f.nps}
                    </span>
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
                    via {f.source}
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
