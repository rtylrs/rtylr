import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { createCycleSchema } from '../../shared/schemas/performance';
import { useCreateReviewCycle } from '../../shared/hooks/queries/usePerformance';
import { Modal, Button, Input, Label, Checkbox } from '../../shared/components/ui';

interface CreateCycleFormProps {
  open: boolean;
  onClose: () => void;
}

const REVIEW_TYPES = ['self', 'manager', 'peer'] as const;

const CreateCycleForm = ({ open, onClose }: CreateCycleFormProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const createCycle = useCreateReviewCycle();

  const [name, setName] = useState('');
  const [periodStart, setPeriodStart] = useState('');
  const [periodEnd, setPeriodEnd] = useState('');
  const [reviewTypes, setReviewTypes] = useState<string[]>([]);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const toggleType = (type: string): void => {
    setReviewTypes((prev) =>
      prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type]
    );
  };

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();

    const result = createCycleSchema.safeParse({ name, periodStart, periodEnd, reviewTypes });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = String(issue.path[0]);
        fieldErrors[key] = issue.message;
      }
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    createCycle.mutate(
      {
        name: result.data.name,
        type: 'adhoc',
        periodStart: result.data.periodStart,
        periodEnd: result.data.periodEnd,
        reviewTypes: result.data.reviewTypes,
      },
      {
        onSuccess: () => {
          setName('');
          setPeriodStart('');
          setPeriodEnd('');
          setReviewTypes([]);
          onClose();
        },
        onError: () => toast.error('Failed to create review cycle'),
      }
    );
  };

  return (
    <Modal open={open} onClose={onClose} title={t('performance.createCycle.title')}>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div>
          <Label htmlFor="cycle-name">{t('performance.createCycle.name')}</Label>
          <Input
            id="cycle-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('performance.createCycle.namePlaceholder')}
          />
          {errors.name && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
              {t('common.required')}
            </p>
          )}
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
          <div>
            <Label htmlFor="period-start">{t('performance.createCycle.periodStart')}</Label>
            <Input
              id="period-start"
              type="date"
              value={periodStart}
              onChange={(e) => setPeriodStart(e.target.value)}
            />
            {errors.periodStart && (
              <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
                {t('common.required')}
              </p>
            )}
          </div>
          <div>
            <Label htmlFor="period-end">{t('performance.createCycle.periodEnd')}</Label>
            <Input
              id="period-end"
              type="date"
              value={periodEnd}
              onChange={(e) => setPeriodEnd(e.target.value)}
            />
            {errors.periodEnd && (
              <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
                {t('common.required')}
              </p>
            )}
          </div>
        </div>

        <div>
          <Label>{t('performance.createCycle.reviewTypes')}</Label>
          <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
            {REVIEW_TYPES.map((type) => (
              <Checkbox
                key={type}
                checked={reviewTypes.includes(type)}
                onChange={() => toggleType(type)}
                label={t(`performance.reviewTypes.${type}`)}
              />
            ))}
          </div>
          {errors.reviewTypes && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
              {t('performance.createCycle.reviewTypesRequired')}
            </p>
          )}
        </div>

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
          <Button type="button" variant="ghost" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" variant="primary" disabled={createCycle.isPending}>
            {createCycle.isPending ? t('common.loading') : t('performance.createCycle.submit')}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default CreateCycleForm;
