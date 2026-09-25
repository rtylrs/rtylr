import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { createCycleSchema } from '../../shared/schemas/performance';
import { useCreateReviewCycle } from '../../shared/hooks/queries/usePerformance';
import { Modal, Button, Input, Label } from '../../shared/components/ui';

interface CreateCycleFormProps {
  open: boolean;
  onClose: () => void;
}

const CreateCycleForm = ({ open, onClose }: CreateCycleFormProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const createCycle = useCreateReviewCycle();

  const [name, setName] = useState('');
  const [periodStart, setPeriodStart] = useState('');
  const [periodEnd, setPeriodEnd] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();

    const result = createCycleSchema.safeParse({ name, periodStart, periodEnd });
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
      },
      {
        onSuccess: () => {
          setName('');
          setPeriodStart('');
          setPeriodEnd('');
          onClose();
        },
        onError: () => toast.error(t('errors.createReviewCycle')),
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
