import React, { useState } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';
import { useGetEmployee } from '../../shared/hooks/queries/useEmployees';
import DocumentVault from '../documents/DocumentVault';

type Tab = 'personal' | 'employment' | 'documents';

const EmployeeDetail = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams({ strict: false });
  const employeeId = id ? Number(id) : undefined;
  const { data: employee, isLoading } = useGetEmployee(employeeId);
  const [activeTab, setActiveTab] = useState<Tab>('personal');

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <div
          className="h-8 w-8 rounded-full border-2 animate-spin"
          style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
        />
      </div>
    );
  }

  if (!employee) {
    return (
      <div className="flex flex-col items-center gap-4 py-20">
        <p className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
          {t('errors.notFound')}
        </p>
        <button
          onClick={() => navigate({ to: '/employees' })}
          className="rounded-lg px-4 py-2 text-sm font-medium"
          style={{ background: 'var(--accent)', color: '#fff' }}
        >
          {t('common.back')}
        </button>
      </div>
    );
  }

  const tabStyle = (tab: Tab): React.CSSProperties => ({
    padding: '8px 16px',
    fontSize: '13px',
    fontWeight: 500,
    borderRadius: '8px',
    background: activeTab === tab ? 'var(--accent)' : 'transparent',
    color: activeTab === tab ? '#fff' : 'var(--hr-text-muted)',
    border: 'none',
    cursor: 'pointer',
  });

  const infoRow = (label: string, value: string | number | undefined): React.JSX.Element => (
    <div
      className="flex items-center justify-between py-2"
      style={{ borderBottom: '1px solid var(--hr-border)' }}
    >
      <span className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
        {label}
      </span>
      <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
        {value ?? '-'}
      </span>
    </div>
  );

  return (
    <>
      <HelmetContainer title="Employee detail · rtylr" />
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate({ to: '/employees' })}
            className="rounded-lg p-2 transition-colors"
            style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
            aria-label={t('common.back')}
          >
            <ArrowLeft size={16} />
          </button>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--hr-text)' }}>
              {employee.jobTitle || `${employee.firstName} ${employee.lastName}`}
            </h1>
            <span
              className="rounded-full px-2 py-0.5 text-xs font-medium"
              style={{
                background:
                  employee.status === 'active' ? 'rgba(34,197,94,0.15)' : 'var(--hr-hover)',
                color: employee.status === 'active' ? '#22c55e' : 'var(--hr-text-muted)',
              }}
            >
              {t(`employees.statuses.${employee.status}` as const)}
            </span>
          </div>
        </div>

        <div className="flex gap-1 rounded-lg p-1" style={{ background: 'var(--hr-hover)' }}>
          <button
            type="button"
            onClick={() => setActiveTab('personal')}
            style={tabStyle('personal')}
          >
            {t('employees.personalInfo')}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('employment')}
            style={tabStyle('employment')}
          >
            {t('employees.employmentInfo')}
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('documents')}
            style={tabStyle('documents')}
          >
            {t('employees.documentsTab')}
          </button>
        </div>

        <div
          className="rounded-xl p-6"
          style={{ background: 'var(--hr-card)', border: '1px solid var(--hr-border)' }}
        >
          {activeTab === 'personal' && (
            <div className="flex flex-col">
              {infoRow(t('employees.form.userId'), employee.userId)}
              {infoRow(
                t('employees.form.jobTitle'),
                employee.jobTitle || `${employee.firstName} ${employee.lastName}`
              )}
              {infoRow(t('employees.form.email'), employee.email)}
              {infoRow(t('employees.form.phone'), employee.phone)}
              {infoRow(t('employees.form.dateOfBirth'), employee.dateOfBirth)}
              {infoRow(t('employees.form.currency'), employee.currency)}
            </div>
          )}

          {activeTab === 'employment' && (
            <div className="flex flex-col">
              {infoRow(
                t('employees.contractType'),
                t(`employees.contractTypes.${employee.employmentType}` as const, {
                  defaultValue: employee.employmentType,
                })
              )}
              {infoRow(t('employees.hireDate'), employee.startDate)}
              {infoRow(t('employees.form.probationEndDate'), employee.probationEndDate)}
              {infoRow(
                t('employees.salary'),
                employee.salary > 0 && employee.currency
                  ? `${employee.salary} ${employee.currency}/mo`
                  : employee.hourlyRate > 0 && employee.currency
                    ? `${employee.hourlyRate} ${employee.currency}/hr`
                    : '-'
              )}
              {infoRow(t('employees.form.accountCreationMode'), employee.accountCreationMode)}
            </div>
          )}

          {activeTab === 'documents' && employeeId != null && (
            <DocumentVault employeeId={employeeId} />
          )}
        </div>
      </div>
    </>
  );
};

export default EmployeeDetail;
