import React, { useState } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';
import { useUserStore } from '@voxire/auth';
import { useGetEmployee } from '../../shared/hooks/queries/useEmployees';
import { canViewCompensation } from '../../shared/lib/hrPermissions';
import DocumentVault from '../documents/DocumentVault';

type Tab = 'personal' | 'employment' | 'documents';

const EmployeeDetail = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams({ strict: false });
  const employeeId = id ? Number(id) : undefined;
  const { data: employee, isLoading } = useGetEmployee(employeeId);
  const [activeTab, setActiveTab] = useState<Tab>('personal');
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const showCompensation = canViewCompensation(role);

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
          type="button"
          onClick={() => navigate({ to: '/employees' })}
          className="rounded-lg px-4 py-2 text-sm font-medium"
          style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
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
            type="button"
            onClick={() => navigate({ to: '/employees' })}
            className="rounded-lg p-2 transition-colors"
            style={{ background: 'var(--hr-hover)', color: 'var(--hr-text-muted)' }}
            aria-label={t('common.back')}
          >
            <ArrowLeft size={16} />
          </button>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--hr-text)' }}>
              {employee.jobTitle}
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
              {infoRow(t('employees.form.jobTitle'), employee.jobTitle)}
              {infoRow(t('employees.form.currency'), employee.currency)}
            </div>
          )}

          {activeTab === 'employment' && (
            <div className="flex flex-col">
              {infoRow(
                t('employees.contractType'),
                t(`employees.contractTypes.${employee.contractType}` as const)
              )}
              {infoRow(t('employees.hireDate'), employee.hireDate)}
              {infoRow(t('employees.form.probationEndDate'), employee.probationEndDate)}
              {showCompensation &&
                infoRow(
                  t('employees.salary'),
                  employee.monthlySalary != null && employee.currency
                    ? `${employee.monthlySalary} ${employee.currency}/mo`
                    : employee.hourlyRate != null && employee.currency
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
