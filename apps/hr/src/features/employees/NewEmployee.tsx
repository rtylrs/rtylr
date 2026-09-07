import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import { employeeSchema, type EmployeeFormData } from '../../shared/schemas/employeeSchema';
import {
  toCreateEmployeePayload,
  useCreateEmployee,
  useListEmployees,
} from '../../shared/hooks/queries/useEmployees';
import { useDepartments } from '../departments/useDepartments';
import CreateLocationForm from './CreateLocationForm';
import { todayIso } from '../../shared/lib/dateUtils';

interface NewEmployeeProps {
  onCancel?: () => void;
  isWizardMode?: boolean;
}

const NewEmployee = ({ onCancel, isWizardMode }: NewEmployeeProps): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const createEmployee = useCreateEmployee();
  const { data: existingEmployees } = useListEmployees({ limit: 1000, offset: 0 });
  const { data: departments } = useDepartments();
  const [createLocationOpen, setCreateLocationOpen] = useState(false);
  const [createdLocationName, setCreatedLocationName] = useState('');

  const [form, setForm] = useState<EmployeeFormData>({
    userId: 0,
    departmentId: 0,
    locationId: 0,
    firstName: '',
    lastName: '',
    email: '',
    contractType: 'full_time',
    status: 'active',
    hireDate: todayIso(),
    currency: 'USD',
    accountCreationMode: 'hr_created',
    jobTitle: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    const result = employeeSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const field = issue.path[0];
        if (typeof field === 'string') {
          fieldErrors[field] = issue.message;
        }
      }
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    const userIdExists =
      result.data.userId > 0 &&
      existingEmployees?.employees?.some((e) => e.userId === String(result.data.userId));
    if (userIdExists) {
      toast.error(t('employees.form.duplicateUserId'));
      return;
    }
    const payload = toCreateEmployeePayload(result.data);
    createEmployee.mutate(payload, {
      onSuccess: () => {
        if (isWizardMode && onCancel) {
          onCancel();
        } else {
          navigate({ to: '/employees' });
        }
      },
    });
  };

  const goBack = (): void => {
    if (isWizardMode && onCancel) {
      onCancel();
    } else {
      navigate({ to: '/employees' });
    }
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

  const field = (
    label: string,
    name: keyof EmployeeFormData,
    type: 'text' | 'number' | 'date' = 'text'
  ): React.JSX.Element => (
    <div>
      <label
        htmlFor={name}
        className="mb-1 block text-sm font-medium"
        style={{ color: 'var(--hr-text)' }}
      >
        {label}
      </label>
      <input
        id={name}
        type={type}
        value={form[name] as string | number}
        onChange={(e) => {
          const value = type === 'number' ? Number(e.target.value) : e.target.value;
          setForm((f) => ({ ...f, [name]: value }));
        }}
        style={inputStyle}
      />
      {errors[name] && <p className="mt-1 text-xs text-red-500">{errors[name]}</p>}
    </div>
  );

  return (
    <>
      <HelmetContainer title="New employee · rtylr" />
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={goBack}
            className="rounded-lg p-2 transition-colors"
            style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
            aria-label={t('common.back')}
          >
            <ArrowLeft size={16} />
          </button>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hr-text)' }}>
            {t('employees.form.title')}
          </h1>
        </div>

        <form
          onSubmit={handleSubmit}
          className="rounded-xl p-6"
          style={{ background: 'var(--hr-card)', border: '1px solid var(--hr-border)' }}
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {field(t('employees.form.firstName'), 'firstName')}
            {field(t('employees.form.lastName'), 'lastName')}
            {field(t('employees.form.email'), 'email')}
            {field(t('employees.form.jobTitle'), 'jobTitle')}
            {field(t('employees.form.userId'), 'userId', 'number')}

            <div>
              <label
                htmlFor="departmentId"
                className="mb-1 block text-sm font-medium"
                style={{ color: 'var(--hr-text)' }}
              >
                {t('employees.form.departmentId')}
              </label>
              <select
                id="departmentId"
                value={form.departmentId}
                onChange={(e) => setForm((f) => ({ ...f, departmentId: Number(e.target.value) }))}
                style={inputStyle}
              >
                <option value={0}>{t('employees.form.selectDepartment')}</option>
                {(departments ?? []).map((dept) => (
                  <option key={dept.id} value={dept.id}>
                    {dept.name}
                  </option>
                ))}
              </select>
              {errors.departmentId && (
                <p className="mt-1 text-xs text-red-500">{errors.departmentId}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="locationId"
                className="mb-1 block text-sm font-medium"
                style={{ color: 'var(--hr-text)' }}
              >
                {t('employees.form.locationId')}
              </label>
              <div className="flex items-center gap-2">
                <input
                  id="locationId"
                  type="number"
                  value={form.locationId}
                  onChange={(e) => {
                    setCreatedLocationName('');
                    setForm((f) => ({ ...f, locationId: Number(e.target.value) }));
                  }}
                  style={inputStyle}
                />
                <button
                  type="button"
                  onClick={() => setCreateLocationOpen(true)}
                  className="whitespace-nowrap rounded-lg px-3 py-2 text-xs font-semibold"
                  style={{
                    background: 'var(--hr-hover)',
                    color: 'var(--hr-text)',
                    border: '1px solid var(--hr-border)',
                  }}
                >
                  {t('employees.form.createLocation.trigger')}
                </button>
              </div>
              {createdLocationName && (
                <p className="mt-1 text-xs" style={{ color: 'var(--hr-text-muted)' }}>
                  {t('employees.form.createLocation.created', { name: createdLocationName })}
                </p>
              )}
              {errors.locationId && (
                <p className="mt-1 text-xs text-red-500">{errors.locationId}</p>
              )}
            </div>

            {field(t('employees.form.hireDate'), 'hireDate', 'date')}
            {field(t('employees.form.currency'), 'currency')}

            <div>
              <label
                htmlFor="contractType"
                className="mb-1 block text-sm font-medium"
                style={{ color: 'var(--hr-text)' }}
              >
                {t('employees.form.contractType')}
              </label>
              <select
                id="contractType"
                value={form.contractType}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    contractType: e.target.value as EmployeeFormData['contractType'],
                  }))
                }
                style={inputStyle}
              >
                <option value="full_time">{t('employees.contractTypes.full_time')}</option>
                <option value="part_time">{t('employees.contractTypes.part_time')}</option>
                <option value="freelancer">{t('employees.contractTypes.freelancer')}</option>
              </select>
            </div>

            <div>
              <label
                htmlFor="status"
                className="mb-1 block text-sm font-medium"
                style={{ color: 'var(--hr-text)' }}
              >
                {t('employees.form.status')}
              </label>
              <select
                id="status"
                value={form.status}
                onChange={(e) =>
                  setForm((f) => ({ ...f, status: e.target.value as EmployeeFormData['status'] }))
                }
                style={inputStyle}
              >
                <option value="active">{t('employees.statuses.active')}</option>
                <option value="probation">{t('employees.statuses.probation')}</option>
                <option value="inactive">{t('employees.statuses.inactive')}</option>
                <option value="terminated">{t('employees.statuses.terminated')}</option>
              </select>
            </div>

            <div>
              <label
                htmlFor="accountCreationMode"
                className="mb-1 block text-sm font-medium"
                style={{ color: 'var(--hr-text)' }}
              >
                {t('employees.form.accountCreationMode')}
              </label>
              <select
                id="accountCreationMode"
                value={form.accountCreationMode}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    accountCreationMode: e.target.value as EmployeeFormData['accountCreationMode'],
                  }))
                }
                style={inputStyle}
              >
                <option value="hr_created">{t('employees.form.hrCreated')}</option>
                <option value="invite_link">{t('employees.form.inviteLink')}</option>
              </select>
            </div>

            {field(t('employees.form.monthlySalary'), 'monthlySalary', 'number')}
            {field(t('employees.form.hourlyRate'), 'hourlyRate', 'number')}
          </div>

          <div className="mt-6 flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={goBack}
              className="rounded-lg px-4 py-2 text-sm font-medium"
              style={{
                background: 'var(--hr-hover)',
                color: 'var(--hr-text-muted)',
                border: '1px solid var(--hr-border)',
              }}
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={createEmployee.isPending}
              className="rounded-lg px-5 py-2 text-sm font-medium transition-opacity disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
            >
              {createEmployee.isPending ? t('common.loading') : t('employees.form.submit')}
            </button>
          </div>
        </form>
      </div>

      <CreateLocationForm
        open={createLocationOpen}
        onClose={() => setCreateLocationOpen(false)}
        onCreated={(location) => {
          setForm((f) => ({ ...f, locationId: location.id }));
          setCreatedLocationName(location.name);
        }}
      />
    </>
  );
};

export default NewEmployee;
