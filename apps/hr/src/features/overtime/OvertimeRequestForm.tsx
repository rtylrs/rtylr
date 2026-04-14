import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { useCreateOvertimeRequest } from '../../shared/hooks/queries/useOvertime';
import { Card, Button, Label, Input, Textarea, Modal } from '../../shared/components/ui';

const overtimeFormSchema = z.object({
  date: z.string().min(1),
  plannedMinutes: z.number().min(1),
  reason: z.string().min(1),
});

interface Props {
  open: boolean;
  onClose: () => void;
}

const OvertimeRequestForm = ({ open, onClose }: Props): React.JSX.Element => {
  const { t } = useTranslation();
  const createRequest = useCreateOvertimeRequest();

  const [form, setForm] = useState({ date: '', plannedMinutes: 0, reason: '' });
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
    createRequest.mutate(result.data, {
      onSuccess: () => {
        setForm({ date: '', plannedMinutes: 0, reason: '' });
        onClose();
      },
    });
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
              style={errors.date ? { borderColor: '#ef4444' } : undefined}
            />
            {errors.date && (
              <p style={{ fontSize: 11, color: '#ef4444', margin: '4px 0 0' }}>{errors.date}</p>
            )}
          </div>

          <div>
            <Label>{t('overtime.form.plannedMinutes')}</Label>
            <Input
              type="number"
              min={1}
              value={form.plannedMinutes || ''}
              onChange={(e) => update('plannedMinutes', Number(e.target.value))}
              placeholder={t('overtime.form.minutesPlaceholder')}
              style={errors.plannedMinutes ? { borderColor: '#ef4444' } : undefined}
            />
            {errors.plannedMinutes && (
              <p style={{ fontSize: 11, color: '#ef4444', margin: '4px 0 0' }}>
                {errors.plannedMinutes}
              </p>
            )}
          </div>

          <div>
            <Label>{t('overtime.form.reason')}</Label>
            <Textarea
              value={form.reason}
              onChange={(e) => update('reason', e.target.value)}
              placeholder={t('overtime.form.reasonPlaceholder')}
              style={{ minHeight: 60, ...(errors.reason ? { borderColor: '#ef4444' } : {}) }}
            />
            {errors.reason && (
              <p style={{ fontSize: 11, color: '#ef4444', margin: '4px 0 0' }}>{errors.reason}</p>
            )}
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 4 }}>
            <Button type="button" variant="ghost" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" variant="primary" disabled={createRequest.isPending}>
              {createRequest.isPending ? t('common.loading') : t('overtime.form.submit')}
            </Button>
          </div>
        </div>
      </form>
    </Modal>
  );
};

export default OvertimeRequestForm;
