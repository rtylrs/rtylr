import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Target } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import type { ReviewGoalPayload } from '@voxire/voxie/hr/performance/v1/types';
import {
  useListGoals,
  useCreateGoal,
  useListReviewCycles,
} from '../../shared/hooks/queries/usePerformance';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { goalFormSchema } from '../../shared/schemas/performance';
import {
  SectionHeader,
  Card,
  Pill,
  Button,
  Input,
  Textarea,
  Label,
  Skeleton,
} from '../../shared/components/ui';

const STATUS_COLORS: Record<string, string> = {
  active: 'var(--hr-status-approved)',
  achieved: 'var(--hr-accent)',
  missed: 'var(--hr-error)',
};

function InlineGoalForm({
  onClose,
  employeeId,
  reviewCycleId,
}: {
  onClose: () => void;
  employeeId: number | undefined;
  reviewCycleId: number | undefined;
}): React.JSX.Element {
  const { t } = useTranslation();
  const createGoal = useCreateGoal();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [weight, setWeight] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    if (!employeeId || !reviewCycleId) return;

    const parsed = goalFormSchema.safeParse({
      title,
      description,
      weight: weight ? Number(weight) : undefined,
    });

    if (!parsed.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        const key = String(issue.path[0]);
        if (key === 'weight') {
          fieldErrors[key] = t('performance.goals.form.weightRange');
        } else {
          fieldErrors[key] = t('common.required');
        }
      }
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    createGoal.mutate(
      {
        employeeId,
        reviewCycleId,
        title: parsed.data.title,
        description: parsed.data.description,
        weight: parsed.data.weight,
      },
      {
        onSuccess: () => {
          setTitle('');
          setDescription('');
          setWeight('');
          onClose();
        },
      }
    );
  };

  return (
    <Card style={{ marginTop: 12 }}>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <Label htmlFor="goal-title">{t('performance.goals.form.title')}</Label>
          <Input
            id="goal-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={t('performance.goals.form.titlePlaceholder')}
          />
          {errors.title && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
              {t('common.required')}
            </p>
          )}
        </div>

        <div>
          <Label htmlFor="goal-desc">{t('performance.goals.form.description')}</Label>
          <Textarea
            id="goal-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('performance.goals.form.descriptionPlaceholder')}
            style={{ minHeight: 60 }}
          />
        </div>

        <div>
          <Label htmlFor="goal-weight">{t('performance.goals.form.weight')}</Label>
          <Input
            id="goal-weight"
            type="number"
            min={0}
            max={100}
            value={weight}
            onChange={(e) => setWeight(e.target.value)}
            placeholder={t('performance.goals.form.weightPlaceholder')}
          />
          {errors.weight && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>{errors.weight}</p>
          )}
        </div>

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
          <Button type="button" variant="ghost" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={createGoal.isPending || !employeeId || !reviewCycleId}
          >
            {createGoal.isPending ? t('common.loading') : t('common.add')}
          </Button>
        </div>
      </form>
    </Card>
  );
}

const GoalsList = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [formOpen, setFormOpen] = useState(false);
  const { data: myEmployee } = useMyEmployeeId();
  const { data: cycleData } = useListReviewCycles();
  const { data, isLoading } = useListGoals(
    myEmployee !== undefined ? { employeeId: myEmployee } : undefined
  );
  const goals = data?.goals ?? [];
  const latestCycleId = Number(cycleData?.cycles?.[0]?.id ?? 0) || undefined;

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.goals'))} />
      <div>
        <SectionHeader
          title={t('performance.goals.title')}
          subtitle={t('performance.goals.subtitle')}
          right={
            <Button variant="primary" onClick={() => setFormOpen(true)}>
              {t('performance.goals.addGoal')}
            </Button>
          }
        />

        {formOpen && (
          <InlineGoalForm
            onClose={() => setFormOpen(false)}
            employeeId={myEmployee}
            reviewCycleId={latestCycleId}
          />
        )}

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 16 }}>
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} height={90} />
            ))}
          </div>
        ) : goals.length === 0 ? (
          <EmptyState
            icon={Target}
            title={t('performance.goals.emptyTitle')}
            description={t('performance.goals.emptyDescription')}
            action={{ label: t('performance.goals.addGoal'), onClick: () => setFormOpen(true) }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 16 }}>
            {goals.map((goal: ReviewGoalPayload) => (
              <Card key={goal.id}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}
                >
                  <h4 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hr-text)' }}>
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
        )}
      </div>
    </>
  );
};

export default GoalsList;
