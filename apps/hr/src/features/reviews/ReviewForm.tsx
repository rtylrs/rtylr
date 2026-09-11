import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
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
      toast.error(t('reviews.errorPickBoth'));
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
          toast.success(t('reviews.created'));
          nav({ to: '/reviews' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title={t('reviews.pageTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 640 }}>
        <Breadcrumbs items={[{ label: 'Reviews', href: '/reviews' }, { label: 'New review' }]} />
        <SectionHeader title={t('reviews.formTitle')} subtitle={t('reviews.formSubtitle')} />
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
                <Label>{t('reviews.employee')}</Label>
                <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
                  <option value={0}>{t('reviews.select')}</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label>{t('reviews.reviewer')}</Label>
                <Select value={reviewerId} onChange={(e) => setReviewerId(Number(e.target.value))}>
                  <option value={0}>{t('reviews.select')}</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
            </div>
            <div>
              <Label>{t('reviews.period')}</Label>
              <Input
                value={period}
                onChange={(e) => setPeriod(e.target.value)}
                placeholder={t('reviews.periodPlaceholder')}
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
                <Label>{t('reviews.goals')}</Label>
                <RatingStars value={goals} onChange={setGoals} />
              </div>
              <div>
                <Label>{t('reviews.values')}</Label>
                <RatingStars value={values} onChange={setValues} />
              </div>
              <div>
                <Label>{t('reviews.skills')}</Label>
                <RatingStars value={skills} onChange={setSkills} />
              </div>
            </div>
            <div>
              <Label>{t('reviews.notes')}</Label>
              <Textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder={t('reviews.notesPlaceholder')}
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
              <span style={{ color: 'var(--hr-text-muted)' }}>{t('reviews.overallRating')}</span>
              <RatingStars value={overall} readOnly />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', flexWrap: 'wrap' }}>
              <Link to="/reviews">
                <Button type="button" variant="ghost">
                  {t('common.cancel')}
                </Button>
              </Link>
              <Button type="submit" variant="primary">
                {t('reviews.form.submit')}
              </Button>
            </div>
          </form>
        </Card>
      </div>
    </>
  );
};

export default ReviewForm;
