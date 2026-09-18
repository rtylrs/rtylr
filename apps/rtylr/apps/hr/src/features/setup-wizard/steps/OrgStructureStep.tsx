import React, { useRef, useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2, Building2 } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import StepContainer from '../components/StepContainer';

interface DepartmentRow {
  name: string;
}

interface OrgStructureStepProps {
  onNext: (departments: DepartmentRow[]) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const OrgStructureStep = ({
  onNext,
  onBack,
  isSubmitting,
}: OrgStructureStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [departments, setDepartments] = useState<DepartmentRow[]>([]);
  const [newName, setNewName] = useState('');
  const nameInputRef = useRef<HTMLInputElement>(null);

  const addDepartment = (): void => {
    const trimmed = newName.trim();
    if (!trimmed) {
      toast.error(t('setup.org.departmentNameRequired'));
      return;
    }
    setDepartments((prev) => [...prev, { name: trimmed }]);
    setNewName('');
  };

  const removeDepartment = (index: number): void => {
    setDepartments((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = (): void => {
    onNext(departments);
  };

  const inputStyle: React.CSSProperties = {
    flex: 1,
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
      title={t('setup.orgStructure.title')}
      description={t('setup.orgStructure.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <input
            ref={nameInputRef}
            type="text"
            placeholder={t('setup.orgStructure.departmentName')}
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') addDepartment();
            }}
            style={inputStyle}
          />
          <button
            type="button"
            onClick={addDepartment}
            className="flex items-center gap-1 rounded-lg px-3 py-2 text-sm font-medium"
            style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
            aria-label={t('setup.orgStructure.addDepartment')}
          >
            <Plus size={14} />
            {t('common.add')}
          </button>
        </div>

        {departments.length === 0 ? (
          <EmptyState
            icon={Building2}
            title={t('setup.orgStructure.noDepartments')}
            action={{
              label: t('setup.orgStructure.addDepartment'),
              onClick: () => nameInputRef.current?.focus(),
            }}
          />
        ) : (
          <div className="flex flex-col gap-2">
            {departments.map((dept, index) => (
              <div
                key={index}
                className="flex items-center justify-between rounded-lg px-4 py-3"
                style={{ background: 'var(--hr-hover)', border: '1px solid var(--hr-border)' }}
              >
                <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
                  {dept.name}
                </span>
                <button
                  type="button"
                  onClick={() => removeDepartment(index)}
                  className="rounded p-1 transition-colors hover:bg-red-500/10"
                  style={{ color: '#ef4444' }}
                  aria-label={t('common.delete')}
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </StepContainer>
  );
};

export default OrgStructureStep;
