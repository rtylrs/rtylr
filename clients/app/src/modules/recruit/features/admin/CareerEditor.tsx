import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogTitle } from '@voxire/ui';

import type { RecruitCareer } from '../../shared/lib/types';
import FormFieldsEditor from './FormFieldsEditor';

interface CareerFormBody {
  slug: string;
  title: string;
  department?: string;
  location?: string;
  employmentType?: string;
  workMode?: string;
  seniority?: string;
  details: string;
  requirements?: string;
  salaryMin?: number;
  salaryMax?: number;
  salaryCurrency?: string;
  isPublished: boolean;
}

interface Props {
  career?: RecruitCareer;
  pending?: boolean;
  onClose: () => void;
  onSave: (body: CareerFormBody) => void;
}

/**
 * CareerEditor - side drawer for create + edit. Controlled form; the
 * parent decides whether the payload is a create or an update based on
 * whether `career` was passed in.
 */
// The value is what the API stores, so it is fixed; only the label is
// translated. Exported so the locale suite can assert a label exists for each.
export const EMPLOYMENT_TYPES = [
  { value: 'Full-time', key: 'fullTime' },
  { value: 'Part-time', key: 'partTime' },
  { value: 'Contract', key: 'contract' },
  { value: 'Internship', key: 'internship' },
] as const;

export const WORK_MODES = [
  { value: 'Remote', key: 'remote' },
  { value: 'Hybrid', key: 'hybrid' },
  { value: 'Onsite', key: 'onsite' },
] as const;

const CareerEditor: React.FC<Props> = ({ career, pending = false, onClose, onSave }) => {
  const { t } = useTranslation('recruitCommon');
  const [title, setTitle] = useState(career?.title ?? '');
  const [slug, setSlug] = useState(career?.slug ?? '');
  const [department, setDepartment] = useState(career?.department ?? '');
  const [location, setLocation] = useState(career?.location ?? '');
  const [employmentType, setEmploymentType] = useState(career?.employmentType ?? 'Full-time');
  const [workMode, setWorkMode] = useState(career?.workMode ?? 'Remote');
  const [seniority, setSeniority] = useState(career?.seniority ?? '');
  const [details, setDetails] = useState(career?.details ?? '');
  const [requirements, setRequirements] = useState(career?.requirements ?? '');
  const [salaryMin, setSalaryMin] = useState(career?.salaryMin ? String(career.salaryMin) : '');
  const [salaryMax, setSalaryMax] = useState(career?.salaryMax ? String(career.salaryMax) : '');
  const [salaryCurrency, setSalaryCurrency] = useState(career?.salaryCurrency ?? 'USD');
  const [isPublished, setIsPublished] = useState(career?.isPublished ?? false);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (pending) return;
    onSave({
      slug,
      title,
      department,
      location,
      employmentType,
      workMode,
      seniority,
      details,
      requirements,
      salaryMin: salaryMin ? Number(salaryMin) : undefined,
      salaryMax: salaryMax ? Number(salaryMax) : undefined,
      salaryCurrency,
      isPublished,
    });
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent
        aria-describedby={undefined}
        closeLabel={t('recruitAdmin.careerEditor.cancel')}
        className="left-auto right-0 top-0 h-dvh max-w-2xl translate-x-0 translate-y-0 rounded-none p-0"
        style={{ background: 'var(--recruit-admin-surface)', color: 'var(--dash-text)' }}
      >
        <form
          onSubmit={onSubmit}
          onClick={(e) => e.stopPropagation()}
          className="flex h-full w-full max-w-2xl flex-col overflow-y-auto"
          style={{
            background: 'var(--recruit-admin-surface)',
            borderLeft: '1px solid var(--recruit-admin-border)',
          }}
        >
          <div
            className="sticky top-0 z-10 flex items-center justify-between border-b px-6 py-4"
            style={{
              borderColor: 'var(--recruit-admin-border)',
              background: 'var(--recruit-admin-surface)',
            }}
          >
            <DialogTitle className="text-lg font-semibold">
              {career
                ? t('recruitAdmin.careerEditor.editTitle')
                : t('recruitAdmin.careerEditor.newTitle')}
            </DialogTitle>
          </div>

          <div className="flex-1 space-y-5 px-6 py-6">
            <Row>
              <Field label={t('recruitAdmin.careerEditor.title')} required>
                <Input value={title} onChange={setTitle} required />
              </Field>
              <Field
                label={t('recruitAdmin.careerEditor.slug')}
                hint={t('recruitAdmin.careerEditor.slugHint')}
              >
                <Input value={slug} onChange={setSlug} />
              </Field>
            </Row>

            <Row>
              <Field label={t('recruitAdmin.careerEditor.department')}>
                <Input value={department} onChange={setDepartment} />
              </Field>
              <Field label={t('recruitAdmin.careerEditor.location')}>
                <Input value={location} onChange={setLocation} />
              </Field>
            </Row>

            <Row>
              <Field label={t('recruitAdmin.careerEditor.employmentTypeLabel')}>
                <Select
                  value={employmentType}
                  onChange={setEmploymentType}
                  options={EMPLOYMENT_TYPES.map((o) => ({
                    value: o.value,
                    label: t(`recruitAdmin.careerEditor.employmentType.${o.key}`),
                  }))}
                />
              </Field>
              <Field label={t('recruitAdmin.careerEditor.workModeLabel')}>
                <Select
                  value={workMode}
                  onChange={setWorkMode}
                  options={WORK_MODES.map((o) => ({
                    value: o.value,
                    label: t(`recruitAdmin.careerEditor.workMode.${o.key}`),
                  }))}
                />
              </Field>
            </Row>

            <Field label={t('recruitAdmin.careerEditor.seniority')}>
              <Input value={seniority} onChange={setSeniority} />
            </Field>

            <Field label={t('recruitAdmin.careerEditor.description')} required>
              <Textarea value={details} onChange={setDetails} rows={8} required />
            </Field>

            <Field label={t('recruitAdmin.careerEditor.requirements')}>
              <Textarea value={requirements} onChange={setRequirements} rows={5} />
            </Field>

            <div className="grid grid-cols-3 gap-4">
              <Field label={t('recruitAdmin.careerEditor.salaryMin')}>
                <Input value={salaryMin} onChange={setSalaryMin} type="number" />
              </Field>
              <Field label={t('recruitAdmin.careerEditor.salaryMax')}>
                <Input value={salaryMax} onChange={setSalaryMax} type="number" />
              </Field>
              <Field label={t('recruitAdmin.careerEditor.currency')}>
                <Input value={salaryCurrency} onChange={setSalaryCurrency} />
              </Field>
            </div>

            <label className="flex items-center gap-2 pt-2">
              <input
                type="checkbox"
                checked={isPublished}
                onChange={(e) => setIsPublished(e.target.checked)}
                className="h-4 w-4"
              />
              <span className="text-sm">{t('recruitAdmin.careerEditor.publish')}</span>
            </label>

            {/* Custom application questions - only editable after the
              role exists. For new roles, the admin saves first, reopens
              the editor, and then this block shows up. Keeps the
              form-field CRUD logic tied to a known career_id. */}
            {career && <FormFieldsEditor careerId={career.id} />}
          </div>

          <div
            className="sticky bottom-0 flex items-center justify-end gap-2 border-t px-6 py-4"
            style={{
              borderColor: 'var(--recruit-admin-border)',
              background: 'var(--recruit-admin-surface)',
            }}
          >
            <button
              type="button"
              onClick={onClose}
              className="rounded-md px-4 py-2 text-sm hover:bg-[var(--dash-hover)]"
            >
              {t('recruitAdmin.careerEditor.cancel')}
            </button>
            <button
              type="submit"
              disabled={pending}
              className="rounded-md bg-[var(--dash-hover)] px-4 py-2 text-sm font-semibold text-[var(--dash-text)] hover:bg-[var(--dash-active)]"
            >
              {career ? t('recruitAdmin.careerEditor.save') : t('recruitAdmin.careerEditor.create')}
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

const Row: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">{children}</div>
);

const Field: React.FC<{
  label: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}> = ({ label, required, hint, children }) => (
  <label className="block">
    <div className="mb-1.5 flex items-baseline justify-between">
      <span className="text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {hint && <span className="text-[10px] text-[var(--dash-text-dim)]">{hint}</span>}
    </div>
    {children}
  </label>
);

const Input: React.FC<{
  value: string;
  onChange: (v: string) => void;
  required?: boolean;
  type?: string;
}> = ({ value, onChange, required, type = 'text' }) => (
  <input
    type={type}
    value={value}
    onChange={(e) => onChange(e.target.value)}
    required={required}
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2.5 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  />
);

const Textarea: React.FC<{
  value: string;
  onChange: (v: string) => void;
  rows?: number;
  required?: boolean;
}> = ({ value, onChange, rows = 4, required }) => (
  <textarea
    value={value}
    onChange={(e) => onChange(e.target.value)}
    rows={rows}
    required={required}
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2.5 text-sm leading-relaxed focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  />
);

const Select: React.FC<{
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}> = ({ value, onChange, options }) => (
  <select
    value={value}
    onChange={(e) => onChange(e.target.value)}
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2.5 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
  >
    {options.map((o) => (
      <option key={o.value} value={o.value}>
        {o.label}
      </option>
    ))}
  </select>
);

export default CareerEditor;
