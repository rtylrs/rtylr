import React from 'react';
import { X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { SavedView } from '../../shared/stores/flowStore';
import { priorityLabelKey } from '../../shared/lib/priorityLabels';
import { statusLabelKey } from '../../shared/lib/statusLabels';

type Filters = SavedView['filters'];

export interface ActiveFilterChip {
  /** Stable key for React and for tests. */
  id: string;
  /** Already-translated text. */
  label: string;
  /** The filters this view would have with this one chip removed. */
  next: Filters;
}

/**
 * Every active filter on a view, one chip each, with the filters that would
 * remain if it were removed.
 *
 * Exported and pure so the chip set can be asserted without rendering: the
 * bug this exists for is a view whose filters cannot be seen, and the
 * interesting part is which chips appear for a given filter object.
 */
export function activeFilterChips(
  filters: Filters,
  t: (key: string) => string
): ActiveFilterChip[] {
  const chips: ActiveFilterChip[] = [];

  for (const status of filters.status ?? []) {
    chips.push({
      id: `status:${status}`,
      label: t(statusLabelKey(status)),
      next: { ...filters, status: (filters.status ?? []).filter((s) => s !== status) },
    });
  }

  for (const priority of filters.priority ?? []) {
    chips.push({
      id: `priority:${priority}`,
      label: t(priorityLabelKey(priority)),
      next: { ...filters, priority: (filters.priority ?? []).filter((p) => p !== priority) },
    });
  }

  // These three are ids rather than names, and this screen has no project or
  // member list to resolve them against. A chip that says which dimension is
  // filtered, and can clear it, is worth more than one that shows a raw id.
  if (filters.assigneeId) {
    chips.push({
      id: 'assigneeId',
      label: t('flow.views.filters.assignee'),
      next: { ...filters, assigneeId: undefined },
    });
  }
  if (filters.projectId) {
    chips.push({
      id: 'projectId',
      label: t('flow.views.filters.project'),
      next: { ...filters, projectId: undefined },
    });
  }
  if ((filters.labels ?? []).length > 0) {
    chips.push({
      id: 'labels',
      label: t('flow.views.filters.labels'),
      next: { ...filters, labels: [] },
    });
  }

  return chips;
}

/** True when the view narrows anything at all. */
export function hasActiveFilters(filters: Filters | undefined): boolean {
  if (!filters) return false;
  return (
    (filters.status ?? []).length > 0 ||
    (filters.priority ?? []).length > 0 ||
    (filters.labels ?? []).length > 0 ||
    Boolean(filters.assigneeId) ||
    Boolean(filters.projectId)
  );
}

export const EMPTY_FILTERS: Filters = {
  status: [],
  priority: [],
  labels: [],
  assigneeId: undefined,
  projectId: undefined,
};

interface ViewFilterChipsProps {
  filters: Filters;
  disabled?: boolean;
  onChange: (next: Filters) => void;
}

/**
 * The toolbar a view's filters were missing entirely: they could not be seen,
 * changed or cleared, which also left the filtered-empty state with nothing
 * to point at.
 */
export const ViewFilterChips: React.FC<ViewFilterChipsProps> = ({
  filters,
  disabled = false,
  onChange,
}) => {
  const { t } = useTranslation();
  const chips = activeFilterChips(filters, t);
  if (chips.length === 0) return null;

  return (
    <div
      className="flex flex-wrap items-center gap-2 px-6 py-2.5"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.views.filters.heading')}
      </span>
      {chips.map((chip) => (
        <button
          key={chip.id}
          type="button"
          disabled={disabled}
          onClick={() => onChange(chip.next)}
          data-chip={chip.id}
          className="inline-flex min-h-8 items-center gap-1.5 rounded-full px-2.5 text-flow-meta font-medium disabled:opacity-50"
          style={{
            background: 'var(--accent-muted)',
            border: '1px solid var(--accent-border)',
            color: 'var(--flow-text)',
          }}
          aria-label={t('flow.views.filters.remove') + ': ' + chip.label}
        >
          {chip.label}
          <X size={11} aria-hidden />
        </button>
      ))}
      {chips.length > 1 && (
        <button
          type="button"
          disabled={disabled}
          onClick={() => onChange(EMPTY_FILTERS)}
          className="min-h-8 rounded-md px-2 text-flow-meta font-semibold disabled:opacity-50"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.views.filters.clearAll')}
        </button>
      )}
    </div>
  );
};

export default ViewFilterChips;
