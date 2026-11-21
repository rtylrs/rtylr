import React from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Layers, Palette, Shapes } from 'lucide-react';
import type { Task, TaskType } from '../../../shared/stores/flowStore';
import FlowSelect from '../../../shared/components/FlowSelect';
import { PROJECT_COLORS } from '../../../shared/lib/colorUtils';
import { TASK_TYPES, isEpic, parentCandidates } from '../../../shared/lib/taskHierarchy';
import { useSetTaskParent, useUpdateTask } from '../useTasks';

interface TaskHierarchyFieldsProps {
  task: Task;
  /** The business's tasks the panel already loads; parents are picked from them. */
  allTasks: readonly Task[];
  canEdit: boolean;
  /** Renders one labelled row of the properties rail. */
  Field: React.FC<{
    icon: React.ComponentType<{ size?: number }>;
    label: string;
    children: React.ReactNode;
  }>;
}

const dot = (color: string) => (
  <span
    aria-hidden="true"
    className="h-2 w-2 shrink-0 rounded-full"
    style={{ background: color }}
  />
);

/**
 * Type, epic colour and parent, for the task panel's rail (T-380, T-573).
 * The parent picker reuses the panel's task list, the same source as the
 * dependency picker, and offers only parents the server would accept.
 */
const TaskHierarchyFields: React.FC<TaskHierarchyFieldsProps> = ({
  task,
  allTasks,
  canEdit,
  Field,
}) => {
  const { t } = useTranslation();
  const updateTask = useUpdateTask();
  const setParent = useSetTaskParent();
  const epic = isEpic(task);
  const currentParentId = task.parentId ?? task.epicId ?? '';
  const candidates = parentCandidates(task, allTasks);
  const current = allTasks.find((other) => other.id === currentParentId);

  const parentOptions = [
    { value: '', label: t('flow.epics.noParent') },
    // The current parent stays listed even when the task list did not return it.
    ...(currentParentId && !candidates.some((c) => c.id === currentParentId)
      ? [
          {
            value: currentParentId,
            label: current
              ? `${current.identifier} ${current.title}`
              : (task.epicTitle ?? currentParentId),
          },
        ]
      : []),
    ...candidates.map((c) => ({
      value: c.id,
      label: `${c.identifier} ${c.title}`,
      icon: isEpic(c) ? <Layers size={11} aria-hidden="true" className="shrink-0" /> : undefined,
    })),
  ];

  return (
    <>
      <Field icon={Shapes} label={t('flow.taskType.label')}>
        <FlowSelect
          label={t('flow.taskType.label')}
          value={task.taskType ?? 'task'}
          options={TASK_TYPES.map((type) => ({ value: type, label: t(`flow.taskType.${type}`) }))}
          onChange={(value) =>
            updateTask.mutate({ id: task.id, data: { taskType: value as TaskType } })
          }
          disabled={!canEdit}
        />
      </Field>

      {epic && (
        <Field icon={Palette} label={t('flow.epics.color')}>
          <FlowSelect
            label={t('flow.epics.color')}
            value={task.epicColor ?? ''}
            leading={task.epicColor ? dot(task.epicColor) : undefined}
            options={[
              { value: '', label: t('flow.epics.noColor') },
              ...PROJECT_COLORS.map((color) => ({ value: color, label: color, icon: dot(color) })),
            ]}
            onChange={(value) => updateTask.mutate({ id: task.id, data: { epicColor: value } })}
            disabled={!canEdit}
          />
        </Field>
      )}

      {!epic && (
        <Field icon={Layers} label={t('flow.epics.parent')}>
          <FlowSelect
            label={t('flow.epics.changeParent')}
            value={currentParentId}
            options={parentOptions}
            onChange={(value) => {
              if (value === currentParentId) return;
              setParent.mutate(
                { taskId: task.id, parentId: value },
                { onSuccess: () => toast.success(t('flowToasts.task.parentChanged')) }
              );
            }}
            disabled={!canEdit || setParent.isPending}
            muted={!currentParentId}
          />
        </Field>
      )}
    </>
  );
};

export default TaskHierarchyFields;
