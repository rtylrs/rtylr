import React, { useRef, useState } from 'react';
import { ListChecks, Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import {
  useCreateWorkType,
  useDeleteWorkType,
  useUpdateWorkType,
  useWorkTypes,
} from '../../time/useWorkDay';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';

const inputClass =
  'flow-focus min-h-[44px] w-full rounded-sm px-2 py-2 text-title md:min-h-0 md:py-1.5 md:text-body';

function inputStyle(): React.CSSProperties {
  return {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border-control)',
    color: 'var(--flow-text)',
  };
}

/**
 * The list an employee picks from in Log Hours. Each entry's description is
 * what shows up there as the read-only "Task Description" once picked - this
 * is where that text is authored, not typed per entry.
 */
const WorkTypesSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const workTypesQuery = useWorkTypes();
  const { data, isLoading } = workTypesQuery;
  const createWorkType = useCreateWorkType();
  const updateWorkType = useUpdateWorkType();
  const deleteWorkType = useDeleteWorkType();

  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const nameInputRef = useRef<HTMLInputElement | null>(null);

  const workTypes = (data?.workTypes ?? []).filter((wt) => !wt.parentId);

  const addWorkType = (): void => {
    if (!newName.trim()) return;
    createWorkType.mutate(
      { name: newName.trim(), description: newDescription.trim() },
      {
        onSuccess: () => {
          setNewName('');
          setNewDescription('');
        },
      }
    );
  };

  const startEdit = (id: string, name: string, description: string): void => {
    setEditingId(id);
    setEditName(name);
    setEditDescription(description);
  };

  const saveEdit = (id: string): void => {
    if (!editName.trim()) return;
    updateWorkType.mutate(
      { id, name: editName.trim(), description: editDescription.trim(), archived: false },
      { onSuccess: () => setEditingId(null) }
    );
  };

  return (
    <section>
      <h2
        className="mb-3 text-meta font-semibold uppercase tracking-wider"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {t('flow.settings.workTypes.section')}
      </h2>
      <p className="mb-3 text-meta leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.settings.workTypes.explained')}
      </p>

      {isLoading ? (
        <p className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.workTypes.loading')}
        </p>
      ) : hasQueryFailed(workTypesQuery) ? (
        // A failed request emptied the list exactly as a first run does, so a
        // 500 invited the user to recreate work types they already had.
        <ListLoadFailed onRetry={() => void workTypesQuery.refetch()} />
      ) : workTypes.length === 0 ? (
        <EmptyState
          icon={ListChecks}
          title={t('flow.settings.workTypes.empty')}
          action={{
            label: t('flow.settings.workTypes.addFirstCta'),
            onClick: () => nameInputRef.current?.focus(),
          }}
          size="compact"
        />
      ) : (
        <ul className="space-y-1.5">
          {workTypes.map((wt) => {
            const id = wt.id ?? '';
            const isEditing = editingId === id;
            return (
              <li key={id} className="rounded-md p-2" style={{ background: 'var(--flow-hover)' }}>
                {isEditing ? (
                  <div className="flex flex-col gap-2">
                    <input
                      aria-label={t('flow.settings.workTypes.name')}
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      className={inputClass}
                      style={inputStyle()}
                    />
                    <input
                      aria-label={t('flow.settings.workTypes.description')}
                      value={editDescription}
                      onChange={(e) => setEditDescription(e.target.value)}
                      placeholder={t('flow.settings.workTypes.descriptionPlaceholder')}
                      className={inputClass}
                      style={inputStyle()}
                    />
                    <div className="flex justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => setEditingId(null)}
                        className="flow-focus min-h-[44px] rounded-sm px-3 text-body font-medium md:min-h-0 md:py-1.5"
                        style={{ color: 'var(--flow-text-muted)' }}
                      >
                        {t('flow.settings.workTypes.cancel')}
                      </button>
                      <button
                        type="button"
                        onClick={() => saveEdit(id)}
                        disabled={updateWorkType.isPending}
                        className="flow-focus min-h-[44px] rounded-sm px-3 text-body font-semibold disabled:opacity-40 md:min-h-0 md:py-1.5"
                        style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
                      >
                        {t('flow.settings.workTypes.save')}
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-start justify-between gap-3">
                    <button
                      type="button"
                      onClick={() => startEdit(id, wt.name ?? '', wt.description ?? '')}
                      className="flow-focus min-w-0 flex-1 rounded-sm px-1 py-0.5 text-left"
                    >
                      <p
                        className="truncate text-body font-medium"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {wt.name}
                      </p>
                      {wt.description && (
                        <p
                          className="mt-0.5 truncate text-meta"
                          style={{ color: 'var(--flow-text-muted)' }}
                        >
                          {wt.description}
                        </p>
                      )}
                    </button>
                    <button
                      type="button"
                      onClick={() => deleteWorkType.mutate(id)}
                      aria-label={t('flow.settings.workTypes.removeLabel', { name: wt.name })}
                      className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm md:h-8 md:w-8"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      <X size={14} aria-hidden="true" />
                    </button>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}

      <div className="mt-3 flex flex-col gap-2">
        <input
          ref={nameInputRef}
          aria-label={t('flow.settings.workTypes.name')}
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder={t('flow.settings.workTypes.namePlaceholder')}
          className={inputClass}
          style={inputStyle()}
        />
        <input
          aria-label={t('flow.settings.workTypes.description')}
          value={newDescription}
          onChange={(e) => setNewDescription(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault();
              addWorkType();
            }
          }}
          placeholder={t('flow.settings.workTypes.descriptionPlaceholder')}
          className={inputClass}
          style={inputStyle()}
        />
        <button
          type="button"
          onClick={addWorkType}
          disabled={!newName.trim() || createWorkType.isPending}
          className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 self-start rounded-sm px-3 text-body font-semibold disabled:opacity-40 md:min-h-0 md:py-2"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          <Plus size={14} aria-hidden="true" />
          {t('flow.settings.workTypes.add')}
        </button>
      </div>
    </section>
  );
};

export default WorkTypesSection;
