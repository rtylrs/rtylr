import React, { useEffect, useState } from 'react';
import { X } from 'lucide-react';

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
  onClose: () => void;
  onSave: (body: CareerFormBody) => void;
}

/**
 * CareerEditor - side drawer for create + edit. Controlled form; the
 * parent decides whether the payload is a create or an update based on
 * whether `career` was passed in.
 */
const CareerEditor: React.FC<Props> = ({ career, onClose, onSave }) => {
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

  // Escape closes the drawer.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
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
    <div className="fixed inset-0 z-50 flex justify-end bg-black/60" onClick={onClose}>
      <form
        onSubmit={onSubmit}
        onClick={(e) => e.stopPropagation()}
        className="flex h-full w-full max-w-2xl flex-col overflow-y-auto"
        style={{ background: '#0f0f0f', borderLeft: '1px solid #1f1f1f' }}
      >
        <div
          className="sticky top-0 z-10 flex items-center justify-between border-b px-6 py-4"
          style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
        >
          <h2 className="text-lg font-semibold">{career ? 'Edit role' : 'New role'}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1.5 text-white/60 hover:bg-white/10 hover:text-white"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="flex-1 space-y-5 px-6 py-6">
          <Row>
            <Field label="Title" required>
              <Input value={title} onChange={setTitle} required />
            </Field>
            <Field label="URL slug" hint="Auto-generated from title if left empty">
              <Input value={slug} onChange={setSlug} />
            </Field>
          </Row>

          <Row>
            <Field label="Department">
              <Input value={department} onChange={setDepartment} />
            </Field>
            <Field label="Location">
              <Input value={location} onChange={setLocation} />
            </Field>
          </Row>

          <Row>
            <Field label="Employment type">
              <Select
                value={employmentType}
                onChange={setEmploymentType}
                options={['Full-time', 'Part-time', 'Contract', 'Internship']}
              />
            </Field>
            <Field label="Work mode">
              <Select
                value={workMode}
                onChange={setWorkMode}
                options={['Remote', 'Hybrid', 'Onsite']}
              />
            </Field>
          </Row>

          <Field label="Seniority">
            <Input value={seniority} onChange={setSeniority} />
          </Field>

          <Field label="Role description" required>
            <Textarea value={details} onChange={setDetails} rows={8} required />
          </Field>

          <Field label="Requirements">
            <Textarea value={requirements} onChange={setRequirements} rows={5} />
          </Field>

          <div className="grid grid-cols-3 gap-4">
            <Field label="Salary min">
              <Input value={salaryMin} onChange={setSalaryMin} type="number" />
            </Field>
            <Field label="Salary max">
              <Input value={salaryMax} onChange={setSalaryMax} type="number" />
            </Field>
            <Field label="Currency">
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
            <span className="text-sm">Publish this role to the careers page</span>
          </label>

          {/* Custom application questions - only editable after the
              role exists. For new roles, the admin saves first, reopens
              the editor, and then this block shows up. Keeps the
              form-field CRUD logic tied to a known career_id. */}
          {career && <FormFieldsEditor careerId={career.id} />}
        </div>

        <div
          className="sticky bottom-0 flex items-center justify-end gap-2 border-t px-6 py-4"
          style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
        >
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-4 py-2 text-sm hover:bg-white/5"
          >
            Cancel
          </button>
          <button
            type="submit"
            className="rounded-md bg-white px-4 py-2 text-sm font-semibold text-black hover:bg-white/90"
          >
            {career ? 'Save changes' : 'Create role'}
          </button>
        </div>
      </form>
    </div>
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
      <span className="text-xs font-medium uppercase tracking-wider text-white/70">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {hint && <span className="text-[10px] text-white/40">{hint}</span>}
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
    className="w-full rounded-md border bg-black/40 px-3 py-2.5 text-sm outline-none focus:border-white/40"
    style={{ borderColor: '#2a2a2a' }}
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
    className="w-full rounded-md border bg-black/40 px-3 py-2.5 text-sm leading-relaxed outline-none focus:border-white/40"
    style={{ borderColor: '#2a2a2a' }}
  />
);

const Select: React.FC<{
  value: string;
  onChange: (v: string) => void;
  options: string[];
}> = ({ value, onChange, options }) => (
  <select
    value={value}
    onChange={(e) => onChange(e.target.value)}
    className="w-full rounded-md border bg-black/40 px-3 py-2.5 text-sm outline-none focus:border-white/40"
    style={{ borderColor: '#2a2a2a' }}
  >
    {options.map((o) => (
      <option key={o} value={o}>
        {o}
      </option>
    ))}
  </select>
);

export default CareerEditor;
