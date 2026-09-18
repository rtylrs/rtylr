import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  workScheduleSchema,
  type WorkScheduleFormData,
} from '../../../shared/schemas/wizardSchemas';
import StepContainer from '../components/StepContainer';

const DAYS = [
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
  'sunday',
] as const;

interface WorkScheduleStepProps {
  onNext: (data: WorkScheduleFormData) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const WorkScheduleStep = ({
  onNext,
  onBack,
  isSubmitting,
}: WorkScheduleStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [form, setForm] = useState<WorkScheduleFormData>({
    name: t('shifts.schedule.standardName'),
    workingDays: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
    startTime: '09:00',
    endTime: '17:00',
    breakMinutes: 60,
    contractedHoursPerWeek: 40,
    dailyOtThresholdMinutes: 480,
    weeklyOtThresholdMinutes: 2400,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const toggleDay = (day: string): void => {
    setForm((f) => ({
      ...f,
      workingDays: f.workingDays.includes(day)
        ? f.workingDays.filter((d) => d !== day)
        : [...f.workingDays, day],
    }));
  };

  const handleSubmit = (): void => {
    const result = workScheduleSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const field = issue.path[0];
        if (typeof field === 'string') {
          fieldErrors[field] = t('common.required');
        }
      }
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    onNext(result.data);
  };

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hr-border)',
    background: 'var(--hr-input-bg)',
    color: 'var(--hr-text)',
    fontSize: '16px',
    outline: 'none',
  };

  return (
    <StepContainer
      title={t('setup.workSchedule.title')}
      description={t('setup.workSchedule.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div>
          <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.workSchedule.scheduleName')}
          </label>
          <input
            type="text"
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            style={inputStyle}
          />
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.workSchedule.workingDays')}
          </label>
          <div className="flex flex-wrap gap-2">
            {DAYS.map((day) => (
              <button
                key={day}
                type="button"
                onClick={() => toggleDay(day)}
                className="rounded-lg px-3 py-2 text-xs font-medium transition-colors"
                style={{
                  background: form.workingDays.includes(day) ? 'var(--accent)' : 'var(--hr-hover)',
                  color: form.workingDays.includes(day) ? '#fff' : 'var(--hr-text-muted)',
                  border: `1px solid ${form.workingDays.includes(day) ? 'var(--accent)' : 'var(--hr-border)'}`,
                }}
              >
                {t(`setup.workSchedule.days.${day}`)}
              </button>
            ))}
          </div>
          {errors.workingDays && <p className="mt-1 text-xs text-red-500">{errors.workingDays}</p>}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.startTime')}
            </label>
            <input
              type="time"
              value={form.startTime}
              onChange={(e) => setForm((f) => ({ ...f, startTime: e.target.value }))}
              style={inputStyle}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.endTime')}
            </label>
            <input
              type="time"
              value={form.endTime}
              onChange={(e) => setForm((f) => ({ ...f, endTime: e.target.value }))}
              style={inputStyle}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.breakMinutes')}
            </label>
            <input
              type="number"
              min={0}
              value={form.breakMinutes}
              onChange={(e) => setForm((f) => ({ ...f, breakMinutes: Number(e.target.value) }))}
              style={inputStyle}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.weeklyHours')}
            </label>
            <input
              type="number"
              min={1}
              value={form.contractedHoursPerWeek}
              onChange={(e) =>
                setForm((f) => ({ ...f, contractedHoursPerWeek: Number(e.target.value) }))
              }
              style={inputStyle}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.dailyOt')}
            </label>
            <input
              type="number"
              min={0}
              value={form.dailyOtThresholdMinutes}
              onChange={(e) =>
                setForm((f) => ({ ...f, dailyOtThresholdMinutes: Number(e.target.value) }))
              }
              style={inputStyle}
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t('setup.workSchedule.weeklyOt')}
            </label>
            <input
              type="number"
              min={0}
              value={form.weeklyOtThresholdMinutes}
              onChange={(e) =>
                setForm((f) => ({ ...f, weeklyOtThresholdMinutes: Number(e.target.value) }))
              }
              style={inputStyle}
            />
          </div>
        </div>
      </div>
    </StepContainer>
  );
};

export default WorkScheduleStep;
