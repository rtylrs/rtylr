import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { useCreateOvertimeRequest } from '../../shared/hooks/queries/useOvertime';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { Button, Label, Input, Textarea, Modal } from '../../shared/components/ui';

const overtimeFormSchema = z.object({
  date: z.string().min(1),
  plannedExtraMinutes: z.number().min(1),
  reason: z.string().min(1),
});

interface Props {
  open: boolean;
  onClose: () => void;
}

const OvertimeRequestForm = ({ open, onClose }: Props): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: myEmployeeId } = useMyEmployeeId();
  const employeeId = myEmployeeId !== undefined ? String(myEmployeeId) : undefined;
  const createRequest = useCreateOvertimeRequest();

  const [form, setForm] = useState({ date: '', plannedExtraMinutes: 0, reason: '' });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    const result = overtimeFormSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const field = issue.path[0] as string;
        fieldErrors[field] = t('common.required');
      });
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    createRequest.mutate(
      { employeeId: employeeId ?? '', ...result.data },
      {
        onSuccess: () => {
          setForm({ date: '', plannedExtraMinutes: 0, reason: '' });
          onClose();
        },
      }
    );
  };

  const update = (field: string, value: string | number): void => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  };

  return (
    <Modal open={open} onClose={onClose} title={t('overtime.form.title')}>
      <form onSubmit={handleSubmit}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <Label>{t('overtime.form.date')}</Label>
            <Input
              type="date"
              value={form.date}
              onChange={(e) => update('date', e.target.value)}
              style={errors.date ? { borderColor: 'var(--hr-error)' } : undefined}
            />
            {errors.date && (
              <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                {errors.date}
              </p>
            )}
          </div>

          <div>
            <Label>{t('overtime.form.plannedMinutes')}</Label>
            <Input
              type="number"
              min={1}
              value={form.plannedExtraMinutes || ''}
              onChange={(e) => update('plannedExtraMinutes', Number(e.target.value))}
              placeholder={t('overtime.form.minutesPlaceholder')}
              style={errors.plannedExtraMinutes ? { borderColor: 'var(--hr-error)' } : undefined}
            />
            {errors.plannedExtraMinutes && (
              <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                {errors.plannedExtraMinutes}
              </p>
            )}
          </div>

          <div>
            <Label>{t('overtime.form.reason')}</Label>
            <Textarea
              value={form.reason}
              onChange={(e) => update('reason', e.target.value)}
              placeholder={t('overtime.form.reasonPlaceholder')}
              style={{
                minHeight: 60,
                ...(errors.reason ? { borderColor: 'var(--hr-error)' } : {}),
              }}
            />
            {errors.reason && (
              <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                {errors.reason}
              </p>
            )}
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 4 }}>
            <Button type="button" variant="ghost" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={createRequest.isPending || !employeeId}
            >
              {createRequest.isPending ? t('common.loading') : t('overtime.form.submit')}
            </Button>
          </div>
        </div>
      </form>
    </Modal>
  );
};

export default OvertimeRequestForm;
