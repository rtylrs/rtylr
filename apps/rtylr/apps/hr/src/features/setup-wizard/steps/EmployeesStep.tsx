import React from 'react';
import { useTranslation } from 'react-i18next';
import { UserPlus, Upload } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListEmployees } from '../../../shared/hooks/queries/useEmployees';
import { toast } from 'sonner';
import StepContainer from '../components/StepContainer';

interface EmployeesStepProps {
  onNext: () => void;
  onBack: () => void;
  onAddEmployee: () => void;
  isSubmitting?: boolean;
  isLast?: boolean;
}

const EmployeesStep = ({
  onNext,
  onBack,
  onAddEmployee,
  isSubmitting,
  isLast,
}: EmployeesStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const { data } = useListEmployees({ limit: 50, offset: 0 });
  const employees = data?.employees ?? [];

  const handleCsvImport = (): void => {
    toast.info(t('setup.employees.csvComingSoon'));
  };

  return (
    <StepContainer
      title={t('setup.employees.title')}
      description={t('setup.employees.description')}
      onBack={onBack}
      onNext={onNext}
      isLast={isLast}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <button
            type="button"
            onClick={onAddEmployee}
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium"
            style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
          >
            <UserPlus size={14} />
            {t('setup.employees.addEmployee')}
          </button>
          <button
            type="button"
            onClick={handleCsvImport}
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium"
            style={{
              background: 'var(--hr-hover)',
              color: 'var(--hr-text-muted)',
              border: '1px solid var(--hr-border)',
            }}
          >
            <Upload size={14} />
            {t('setup.employees.importCsv')}
          </button>
        </div>

        {employees.length === 0 ? (
          <EmptyState
            icon={UserPlus}
            title={t('setup.employees.noEmployeesHint')}
            action={{ label: t('setup.employees.addEmployee'), onClick: onAddEmployee }}
          />
        ) : (
          <div
            className="rounded-lg overflow-hidden"
            style={{ border: '1px solid var(--hr-border)' }}
          >
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: 'var(--hr-hover)' }}>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('employees.jobTitle')}
                  </th>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('employees.contractType')}
                  </th>
                  <th
                    className="px-4 py-2 text-left font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('common.status')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {employees.map((emp) => (
                  <tr key={emp.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                    <td className="px-4 py-2" style={{ color: 'var(--hr-text)' }}>
                      {emp.jobTitle}
                    </td>
                    <td className="px-4 py-2" style={{ color: 'var(--hr-text-muted)' }}>
                      {emp.contractType}
                    </td>
                    <td className="px-4 py-2">
                      <span
                        className="rounded-full px-2 py-0.5 text-xs font-medium"
                        style={{
                          background:
                            emp.status === 'active' ? 'rgba(34,197,94,0.15)' : 'var(--hr-hover)',
                          color: emp.status === 'active' ? '#22c55e' : 'var(--hr-text-muted)',
                        }}
                      >
                        {emp.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </StepContainer>
  );
};

export default EmployeesStep;
