import React, { useState } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useStaff } from '../staff/useStaff';
import { useCreateReview } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Select,
  Textarea,
  Breadcrumbs,
  useIsMobile,
} from '../../shared/components/ui';
import RatingStars from '../../shared/components/RatingStars';
import { todayIso } from '../../shared/lib/dateUtils';

const ReviewForm: React.FC = () => {
  const isMobile = useIsMobile(900);
  const nav = useNavigate();
  const { data: staff = [] } = useStaff();
  const create = useCreateReview();

  const [staffId, setStaffId] = useState(0);
  const [reviewerId, setReviewerId] = useState(0);
  const [period, setPeriod] = useState('Q2 2026');
  const [goals, setGoals] = useState(4);
  const [values, setValues] = useState(4);
  const [skills, setSkills] = useState(4);
  const [notes, setNotes] = useState('');

  const overall = Math.round((goals + values + skills) / 3);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const who = staff.find((s) => s.id === staffId);
    const reviewer = staff.find((s) => s.id === reviewerId);
    if (!who || !reviewer) {
      toast.error('Pick employee and reviewer');
      return;
    }
    create.mutate(
      {
        staffId: who.id,
        staffName: who.name,
        reviewerId: reviewer.id,
        reviewer: reviewer.name,
        period,
        status: 'Submitted',
        overall,
        goals,
        values,
        skills,
        notes,
        date: todayIso(),
      },
      {
        onSuccess: () => {
          toast.success('Review created');
          nav({ to: '/reviews' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="New review · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 640 }}>
        <Breadcrumbs items={[{ label: 'Reviews', href: '/reviews' }, { label: 'New review' }]} />
        <SectionHeader title="New Performance Review" subtitle="Rate goals, values, and skills" />
        <Card>
          <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr',
                gap: 10,
              }}
            >
              <div>
                <Label>Employee</Label>
                <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
                  <option value={0}>Select</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label>Reviewer</Label>
                <Select value={reviewerId} onChange={(e) => setReviewerId(Number(e.target.value))}>
                  <option value={0}>Select</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
            </div>
            <div>
              <Label>Period</Label>
              <Input
                value={period}
                onChange={(e) => setPeriod(e.target.value)}
                placeholder="e.g. Q2 2026"
              />
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr 1fr',
                gap: 10,
              }}
            >
              <div>
                <Label>Goals</Label>
                <RatingStars value={goals} onChange={setGoals} />
              </div>
              <div>
                <Label>Values</Label>
                <RatingStars value={values} onChange={setValues} />
              </div>
              <div>
                <Label>Skills</Label>
                <RatingStars value={skills} onChange={setSkills} />
              </div>
            </div>
            <div>
              <Label>Notes</Label>
              <Textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Key feedback and next steps"
              />
            </div>
            <div
              style={{
                padding: 10,
                background: 'var(--hr-hover)',
                borderRadius: 8,
                fontSize: 12,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <span style={{ color: 'var(--hr-text-muted)' }}>Overall rating</span>
              <RatingStars value={overall} readOnly />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', flexWrap: 'wrap' }}>
              <Link to="/reviews">
                <Button type="button" variant="ghost">
                  Cancel
                </Button>
              </Link>
              <Button type="submit" variant="primary">
                Submit review
              </Button>
            </div>
          </form>
        </Card>
      </div>
    </>
  );
};

export default ReviewForm;
