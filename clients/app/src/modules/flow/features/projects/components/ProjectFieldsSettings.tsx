import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronUp, Pencil, Plus, SlidersHorizontal, X } from 'lucide-react';
import { EmptyState } from '../../../../../shared/components/dash/EmptyState';
import GateNotice from '../../../shared/components/GateNotice';
import ConfirmDialog from '../../../shared/components/ConfirmDialog';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import type { ApiField } from '../../../shared/services/flowApi';
import { movedStatusIds } from '../statusUtils';
import {
  useCreateField,
  useDeleteField,
  useProjectFields,
  useReorderFields,
  useUpdateField,
} from '../useFields';
import { useProjectMembers } from '../useProjectMembers';
import FieldEditorDialog from './FieldEditorDialog';

interface ProjectFieldsSettingsProps {
  projectId: string;
  /** Project admin. Everyone else sees the list read-only, with the reason. */
  canManage: boolean;
}

const iconButton =
  'flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8';

/**
 * Project Settings > Fields: the project's custom fields in order. Admins add,
 * edit, reorder and delete them; a delete first says how many tasks will lose
 * a value.
 */
const ProjectFieldsSettings: React.FC<ProjectFieldsSettingsProps> = ({ projectId, canManage }) => {
  const { t } = useTranslation();
  const fieldsQuery = useProjectFields(projectId);
  const { data: members = [] } = useProjectMembers(projectId);
  const createField = useCreateField(projectId);
  const updateField = useUpdateField(projectId);
  const reorderFields = useReorderFields(projectId);
  const deleteField = useDeleteField(projectId);

  const [editing, setEditing] = useState<ApiField | 'new' | null>(null);
  const [pendingDelete, setPendingDelete] = useState<ApiField | null>(null);

  if (hasQueryFailed(fieldsQuery)) {
    return <ListLoadFailed onRetry={() => void fieldsQuery.refetch()} />;
  }

  const fields = fieldsQuery.data ?? [];
  const people = members.map((m) => ({ id: String(m.userId), name: m.name }));

  const move = (index: number, direction: -1 | 1): void => {
    const ids = movedStatusIds(fields, index, index + direction);
    if (ids) reorderFields.mutate(ids);
  };

  return (
    <div className="space-y-3" data-project-fields>
      {!canManage && <GateNotice reason={t('flow.fields.gateReason')} />}
      <div className="flex items-center justify-between">
        <p
          className="text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.fields.count', { count: fields.length })}
        </p>
        {canManage && (
          <button
            type="button"
            onClick={() => setEditing('new')}
            className="flex items-center gap-1 text-flow-meta px-2 py-1 rounded-sm"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
          >
            <Plus size={10} aria-hidden="true" /> {t('flow.fields.addField')}
          </button>
        )}
      </div>

      {fieldsQuery.isLoading ? null : fields.length === 0 ? (
        <EmptyState
          icon={SlidersHorizontal}
          title={t('flow.fields.emptyTitle')}
          description={t('flow.fields.emptyDescription')}
          action={
            canManage
              ? { label: t('flow.fields.addField'), onClick: () => setEditing('new') }
              : undefined
          }
          size="compact"
        />
      ) : (
        <ul className="space-y-1.5">
          {fields.map((field, index) => (
            <li
              key={field.id}
              className="flex items-center gap-2 rounded-md p-2"
              style={{ background: 'var(--flow-hover)' }}
              data-field-row={field.id}
            >
              <div className="min-w-0 flex-1">
                <p
                  className="truncate text-flow-body font-medium"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {field.name}
                  {field.required && (
                    <span
                      className="ms-1.5 text-flow-meta"
                      style={{ color: 'var(--flow-warn-fg)' }}
                    >
                      {t('flow.fields.requiredBadge')}
                    </span>
                  )}
                </p>
                <p className="truncate text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t(`flow.fields.types.${field.type}`)}
                  {' · '}
                  {t('flow.fields.valueCount', { count: field.valueCount })}
                </p>
              </div>
              <button
                type="button"
                onClick={() => move(index, -1)}
                disabled={!canManage || index === 0 || reorderFields.isPending}
                aria-label={t('flow.fields.moveUp', { name: field.name })}
                className={iconButton}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <ChevronUp size={14} aria-hidden="true" />
              </button>
              <button
                type="button"
                onClick={() => move(index, 1)}
                disabled={!canManage || index === fields.length - 1 || reorderFields.isPending}
                aria-label={t('flow.fields.moveDown', { name: field.name })}
                className={iconButton}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <ChevronDown size={14} aria-hidden="true" />
              </button>
              <button
                type="button"
                onClick={() => setEditing(field)}
                disabled={!canManage}
                aria-label={t('flow.fields.edit', { name: field.name })}
                className={iconButton}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <Pencil size={13} aria-hidden="true" />
              </button>
              <button
                type="button"
                onClick={() => setPendingDelete(field)}
                disabled={!canManage}
                aria-label={t('flow.fields.delete', { name: field.name })}
                className={iconButton}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <X size={14} aria-hidden="true" />
              </button>
            </li>
          ))}
        </ul>
      )}

      {canManage && (
        <FieldEditorDialog
          open={editing !== null}
          onOpenChange={(open) => !open && setEditing(null)}
          field={editing && editing !== 'new' ? editing : undefined}
          people={people}
          pending={createField.isPending || updateField.isPending}
          onSubmit={({ type, ...data }) => {
            const done = { onSuccess: () => setEditing(null) };
            if (editing && editing !== 'new') updateField.mutate({ id: editing.id, ...data }, done);
            else createField.mutate({ type, ...data }, done);
          }}
        />
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => !open && setPendingDelete(null)}
        title={t('flow.fields.deleteTitle', { name: pendingDelete?.name ?? '' })}
        description={
          pendingDelete && pendingDelete.valueCount > 0
            ? t('flow.fields.deleteWithValues', { count: pendingDelete.valueCount })
            : t('flow.fields.deleteNoValues')
        }
        confirmLabel={t('flow.fields.deleteConfirm')}
        pending={deleteField.isPending}
        onConfirm={() => {
          if (!pendingDelete) return;
          deleteField.mutate(pendingDelete.id, { onSettled: () => setPendingDelete(null) });
        }}
      />
    </div>
  );
};

export default ProjectFieldsSettings;
