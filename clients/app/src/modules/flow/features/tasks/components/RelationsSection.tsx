import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';
import FlowSelect from '../../../shared/components/FlowSelect';
import ConfirmDialog from '../../../shared/components/ConfirmDialog';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { normalizeSearchText } from '../../../shared/filters/taskFilters';
import { taskIdFromKey, useTaskSearch } from '../../../shared/hooks/queries/useTaskSearch';
import type {
  ApiRelatedTask,
  ApiRelationDirection,
  ApiRelationKey,
  ApiRelationType,
} from '../../../shared/services/flowApi';
import { useAddTaskRelation, useRemoveTaskRelation, useTaskRelations } from '../useTaskRelations';
import { useUpdateTaskStatus } from '../useTasks';

/** Each way a relation reads from this task's side, in the order groups are listed. */
export const RELATION_KINDS: {
  key: ApiRelationKey;
  type: ApiRelationType;
  direction: ApiRelationDirection;
}[] = [
  { key: 'blocks', type: 'blocks', direction: 'outward' },
  { key: 'blocked_by', type: 'blocks', direction: 'inward' },
  { key: 'relates_to', type: 'relates', direction: 'outward' },
  { key: 'duplicates', type: 'duplicates', direction: 'outward' },
  { key: 'duplicated_by', type: 'duplicates', direction: 'inward' },
  { key: 'clones', type: 'clones', direction: 'outward' },
  { key: 'cloned_by', type: 'clones', direction: 'inward' },
];

const MAX_RESULTS = 8;

const isClosed = (task: Pick<ApiRelatedTask, 'statusBucket'>) =>
  task.statusBucket === 'done' || task.statusBucket === 'cancelled';

interface RelationsSectionProps {
  taskId: string;
  identifier: string;
  canEdit: boolean;
  onOpenTask: (taskId: string) => void;
}

/**
 * The task panel's relations: what this task blocks or is blocked by, relates
 * to, duplicates or clones, read from either end. It took over the
 * Dependencies field, whose links are "blocked by" relations.
 */
const RelationsSection: React.FC<RelationsSectionProps> = ({
  taskId,
  identifier,
  canEdit,
  onOpenTask,
}) => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const relationsQuery = useTaskRelations(taskId);
  const groups = relationsQuery.data ?? [];
  const addRelation = useAddTaskRelation();
  const removeRelation = useRemoveTaskRelation();
  const closeTask = useUpdateTaskStatus();
  // "Blocked by" first: it is what the Dependencies field used to add.
  const [kind, setKind] = useState<ApiRelationKey>('blocked_by');
  const [query, setQuery] = useState('');
  const [closeOffer, setCloseOffer] = useState<{ taskId: string; identifier: string } | null>(null);

  const typing = query.trim() !== '';
  const search = useTaskSearch(query, canEdit && typing);
  const tokens = normalizeSearchText(query).split(/\s+/).filter(Boolean);
  const keyId = taskIdFromKey(query);
  const candidates = typing
    ? (search.data ?? [])
        .filter((task) => String(task.id) !== taskId && !task.archived)
        // Short queries come back as the whole task list; narrow it here.
        .filter((task) => {
          if (keyId) return true;
          const haystack = normalizeSearchText(`${task.identifier ?? ''} ${task.title}`);
          return tokens.every((token) => haystack.includes(token));
        })
        .slice(0, MAX_RESULTS)
    : [];

  const link = (otherTaskId: string, otherIdentifier: string) => {
    const chosen = RELATION_KINDS.find((option) => option.key === kind) ?? RELATION_KINDS[1]!;
    addRelation.mutate(
      { taskId, otherTaskId, type: chosen.type, direction: chosen.direction },
      {
        onSuccess: () => {
          setQuery('');
          // A duplicate is usually closed once it is marked, so offer to.
          if (chosen.key === 'duplicates') setCloseOffer({ taskId, identifier });
          if (chosen.key === 'duplicated_by') {
            setCloseOffer({ taskId: otherTaskId, identifier: otherIdentifier });
          }
        },
      }
    );
  };

  return (
    <div className="space-y-1.5">
      {hasQueryFailed(relationsQuery) && (
        <ListLoadFailed onRetry={() => void relationsQuery.refetch()} />
      )}
      {groups.map((group) => (
        <div key={group.relation} className="space-y-1">
          <p className="text-flow-meta font-medium" style={{ color: 'var(--flow-text-muted)' }}>
            {t(`flow.taskDetail.relationKinds.${group.relation}`)}
          </p>
          {group.tasks.map((related) => (
            <div key={related.relationId} className="flex items-center justify-between gap-1">
              <button
                type="button"
                onClick={() => onOpenTask(related.taskId)}
                className="text-flow-meta text-start truncate flex-1 hover:underline"
                style={{
                  color: isClosed(related) ? 'var(--flow-text-muted)' : 'var(--accent)',
                  textDecoration: isClosed(related) ? 'line-through' : undefined,
                }}
                title={related.projectName || undefined}
              >
                {related.identifier} · {related.title}
              </button>
              {canEdit && (
                <button
                  type="button"
                  className="text-flow-meta shrink-0"
                  style={{ color: 'var(--flow-text-dim)' }}
                  disabled={removeRelation.isPending}
                  aria-label={t('flow.taskDetail.removeRelation', {
                    identifier: related.identifier,
                  })}
                  onClick={() =>
                    removeRelation.mutate({
                      taskId,
                      relationId: related.relationId,
                      otherTaskId: related.taskId,
                    })
                  }
                >
                  {t('flow.taskDetail.remove')}
                </button>
              )}
            </div>
          ))}
        </div>
      ))}
      {canEdit && (
        <div className="space-y-1 pt-1">
          <FlowSelect
            label={t('flow.taskDetail.relationType')}
            value={kind}
            options={RELATION_KINDS.map((option) => ({
              value: option.key,
              label: t(`flow.taskDetail.relationKinds.${option.key}`),
            }))}
            onChange={(value) => setKind(value as ApiRelationKey)}
            muted
          />
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={t('flow.taskDetail.relationSearch')}
            aria-label={t('flow.taskDetail.relationSearch')}
            className="flow-focus w-full rounded-md px-2.5 py-1.5 text-flow-meta placeholder:opacity-60"
            style={{
              color: 'var(--flow-text)',
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
            }}
          />
          {typing && (
            <ul aria-label={t('flow.taskDetail.relationResults')} className="space-y-0.5">
              {candidates.map((candidate) => (
                <li key={candidate.id}>
                  <button
                    type="button"
                    disabled={addRelation.isPending}
                    onClick={() =>
                      link(String(candidate.id), candidate.identifier ?? String(candidate.id))
                    }
                    className="flow-focus w-full truncate rounded-sm px-1.5 py-1 text-start text-flow-meta hover:bg-[var(--flow-hover)] disabled:opacity-40"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {candidate.identifier} · {candidate.title}
                  </button>
                </li>
              ))}
              {!search.isFetching && candidates.length === 0 && (
                <li className="px-1.5 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.taskDetail.relationNoMatch')}
                </li>
              )}
            </ul>
          )}
        </div>
      )}
      <ConfirmDialog
        open={closeOffer !== null}
        onOpenChange={(open) => {
          if (!open) setCloseOffer(null);
        }}
        title={t('flow.taskDetail.closeDuplicateTitle', {
          identifier: closeOffer?.identifier ?? '',
        })}
        description={t('flow.taskDetail.closeDuplicateDescription')}
        confirmLabel={t('flow.taskDetail.closeDuplicateConfirm')}
        cancelLabel={t('flow.taskDetail.closeDuplicateKeepOpen')}
        pending={closeTask.isPending}
        onConfirm={() => {
          if (!closeOffer) return;
          closeTask.mutate(
            { id: closeOffer.taskId, status: 'cancelled' },
            {
              onSuccess: () => {
                void qc.invalidateQueries({ queryKey: ['task-relations'] });
                void qc.invalidateQueries({ queryKey: ['task'] });
                setCloseOffer(null);
              },
            }
          );
        }}
      />
    </div>
  );
};

export default RelationsSection;
