import React from 'react';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Plus, ClipboardList } from 'lucide-react';
import { useReviews } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  EmptyState,
  Skeleton,
  Pill,
} from '../../shared/components/ui';
import RatingStars from '../../shared/components/RatingStars';

const STATUS_COLORS: Record<string, string> = {
  Draft: '#a1a1aa',
  'In Progress': '#f59e0b',
  Submitted: '#3b82f6',
  Completed: '#10b981',
};

const Reviews: React.FC = () => {
  const { data, isLoading } = useReviews();
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title="Performance Reviews"
        subtitle="Track review cycles and ratings"
        right={
          <Link to="/reviews/new" style={{ textDecoration: 'none' }}>
            <Button variant="primary" icon={<Plus size={13} />}>
              New review
            </Button>
          </Link>
        }
      />
      {isLoading ? (
        <Skeleton height={200} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon={<ClipboardList size={20} />}
          title="No reviews yet"
          description="Start a review cycle to see results here."
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
                  <span style={{ color: 'var(--hr-text-muted)' }}>Goals</span>{' '}
                  <strong style={{ color: 'var(--hr-text)' }}>{r.goals}/5</strong>
                </div>
                <div>
                  <span style={{ color: 'var(--hr-text-muted)' }}>Values</span>{' '}
                  <strong style={{ color: 'var(--hr-text)' }}>{r.values}/5</strong>
                </div>
                <div>
                  <span style={{ color: 'var(--hr-text-muted)' }}>Skills</span>{' '}
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
  );
};

export default Reviews;
