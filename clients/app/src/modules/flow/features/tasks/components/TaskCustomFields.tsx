import React from 'react';
import { useTranslation } from 'react-i18next';
import { SlidersHorizontal } from 'lucide-react';
import { toast } from 'sonner';
import type { ApiTaskFieldValue } from '../../../shared/services/flowApi';
import CustomFieldInput, {
  type CustomFieldPerson,
} from '../../../shared/components/CustomFieldInput';
import { fieldValueInput, isEmptyFieldValue } from '../../../shared/lib/customFields';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { useProjectFields, useSetTaskFieldValues } from '../../projects/useFields';

type RowProps = {
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  children: React.ReactNode;
};

interface TaskCustomFieldsProps {
  taskId: string;
  projectId: string;
  values: Record<string, ApiTaskFieldValue> | undefined;
  canEdit: boolean;
  people: readonly CustomFieldPerson[];
  /** The task panel's own property row, so a custom field lines up with the built-in ones. */
  Row: React.ComponentType<RowProps>;
}

/**
 * The task's custom fields, one property row each, in the project's order.
 * Each change is written on its own; a required field cannot be cleared.
 */
const TaskCustomFields: React.FC<TaskCustomFieldsProps> = ({
  taskId,
  projectId,
  values,
  canEdit,
  people,
  Row,
}) => {
  const { t } = useTranslation();
  const fieldsQuery = useProjectFields(projectId);
  const setValues = useSetTaskFieldValues(taskId);

  if (hasQueryFailed(fieldsQuery)) {
    return (
      <div role="alert" className="mb-5 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        <span>{t('flow.loadFailed.description')} </span>
        <button
          type="button"
          onClick={() => void fieldsQuery.refetch()}
          className="flow-focus underline"
          style={{ color: 'var(--flow-text)' }}
        >
          {t('flow.loadFailed.retry')}
        </button>
      </div>
    );
  }

  const fields = fieldsQuery.data ?? [];
  if (fields.length === 0) return null;

  return (
    <div data-task-custom-fields>
      {fields.map((field) => (
        <Row
          key={field.id}
          icon={SlidersHorizontal}
          label={field.required ? `${field.name} *` : field.name}
        >
          <CustomFieldInput
            field={field}
            value={values?.[field.id]}
            people={people}
            disabled={!canEdit || setValues.isPending}
            onChange={(next) => {
              if (field.required && isEmptyFieldValue(next)) {
                // The server refuses this too; saying so here keeps the old value on screen.
                toast.error(t('flow.fields.requiredMissing', { names: field.name }));
                return;
              }
              setValues.mutate([fieldValueInput(field, next)]);
            }}
          />
        </Row>
      ))}
    </div>
  );
};

export default TaskCustomFields;
