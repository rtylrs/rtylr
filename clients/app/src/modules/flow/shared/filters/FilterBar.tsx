import React, { useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ListFilter, Search, UserRound, UserX, X } from 'lucide-react';
import { Avatar } from '@voxire/ui';
import { initialsFromName } from '../lib/memberUtils';
import {
  activeFieldCount,
  isOnlyMine,
  toggleOnlyMine,
  toggleValue,
  UNASSIGNED,
  type FilterField,
  type TaskFilters,
} from './taskFilters';
import { orderPeople, type FilterFieldConfig, type FilterPerson } from './filterFields';
import FilterPopover, { DUE_LABEL_KEY } from './FilterPopover';
import { useFilterShortcuts } from './useFilterShortcuts';
import type { TaskFilterState } from './useTaskFilters';

/** Avatars shown before the "+N" overflow. Selected people are always shown. */
const STRIP_SIZE = 6;

export interface FilterBarProps {
  state: TaskFilterState;
  currentUserId: string | null;
  /** Fields the Filter popover lists, in order. Due is added after them. */
  fields: FilterFieldConfig[];
  /** People for the avatar strip and "Only mine". Omit to leave both out. */
  people?: readonly FilterPerson[];
  showDue?: boolean;
  /** Matching tasks, for "X of Y tasks visible". */
  visible: number;
  total: number;
  /** Page-specific controls after the Filter button (e.g. "Show completed"). */
  extra?: React.ReactNode;
}

interface Chip {
  id: string;
  label: string;
  next: TaskFilters;
  /** A person chip; hidden where the avatar strip already shows the selection. */
  person?: boolean;
}

/**
 * The filter toolbar every Flow task list shares, modelled on Jira's board
 * header: a search box, an avatar per person (click to show their tasks,
 * several combine), "Only mine", a two-pane Filter popover, a chip for each
 * active value, and how many tasks are left visible.
 */
export const FilterBar: React.FC<FilterBarProps> = ({
  state,
  currentUserId,
  fields,
  people,
  showDue = true,
  visible,
  total,
  extra,
}) => {
  const { t } = useTranslation();
  const { filters, setFilters, clear, active, dim, setDim } = state;
  const [open, setOpen] = useState(false);
  const [field, setField] = useState<FilterField | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  const openPopover = (target?: FilterField): void => {
    setField(target ?? field ?? fields[0]?.field ?? (showDue ? 'due' : null));
    setOpen(true);
  };

  useFilterShortcuts({
    onOpenFilter: () => openPopover(),
    onToggleMine: people ? () => setFilters((f) => toggleOnlyMine(f, currentUserId)) : undefined,
    onClearAll: clear,
  });

  const ordered = useMemo(() => orderPeople(people ?? [], currentUserId), [people, currentUserId]);
  const strip = useMemo(() => {
    const selected = ordered.filter((p) => filters.assignees.includes(p.id));
    const rest = ordered.filter((p) => !filters.assignees.includes(p.id));
    return [...selected, ...rest].slice(0, Math.max(STRIP_SIZE, selected.length));
  }, [ordered, filters.assignees]);
  const overflow = ordered.length - strip.length;

  const chips = useMemo(() => buildChips(filters, fields, t), [filters, fields, t]);
  const stripShown = Boolean(people && ordered.length > 0);
  const chipHidden = (chip: Chip): boolean => Boolean(chip.person && stripShown);
  const count = activeFieldCount(filters);
  const mine = isOnlyMine(filters, currentUserId);

  return (
    <div
      className="flex shrink-0 flex-col gap-2 px-4 py-2 sm:px-5"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
      data-flow-filter-bar
    >
      <div className="flex flex-wrap items-center gap-2">
        <label
          className="flex min-h-9 w-full items-center gap-2 rounded-md px-2.5 sm:w-[220px] md:min-h-8"
          style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text-dim)' }}
        >
          <Search size={13} aria-hidden="true" />
          <input
            ref={searchRef}
            data-flow-filter-search
            value={filters.query}
            onChange={(e) => setFilters((f) => ({ ...f, query: e.target.value }))}
            onKeyDown={(e) => {
              if (e.key === 'Escape' && !e.shiftKey) {
                setFilters((f) => ({ ...f, query: '' }));
                searchRef.current?.blur();
              }
            }}
            placeholder={t('flow.filters.searchPlaceholder')}
            aria-label={t('flow.filters.searchLabel')}
            className="min-w-0 flex-1 bg-transparent text-flow-meta outline-none"
            style={{ color: 'var(--flow-text)' }}
          />
          <kbd
            className="hidden rounded px-1 text-[10px] font-semibold md:inline"
            style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text-dim)' }}
            aria-hidden="true"
          >
            /
          </kbd>
        </label>

        {people && ordered.length > 0 && (
          <div
            className="hidden items-center md:flex"
            role="group"
            aria-label={t('flow.filters.people')}
          >
            {strip.map((person, i) => {
              const pressed = filters.assignees.includes(person.id);
              const name =
                person.id === currentUserId
                  ? t('flow.filters.you', { name: person.name })
                  : person.name;
              return (
                <button
                  key={person.id}
                  type="button"
                  aria-pressed={pressed}
                  aria-label={t('flow.filters.showPerson', { name })}
                  title={name}
                  data-filter-person={person.id}
                  onClick={() =>
                    setFilters((f) => ({ ...f, assignees: toggleValue(f.assignees, person.id) }))
                  }
                  className="flow-focus relative rounded-full transition-transform hover:z-10 hover:-translate-y-0.5"
                  style={{
                    marginInlineStart: i === 0 ? 0 : 3,
                    boxShadow: pressed
                      ? '0 0 0 2px var(--flow-surface), 0 0 0 4px var(--flow-action-bg)'
                      : '0 0 0 2px var(--flow-surface)',
                    zIndex: pressed ? 5 : undefined,
                  }}
                >
                  <Avatar size="sm" initials={initialsFromName(person.name)} aria-hidden />
                </button>
              );
            })}
            {overflow > 0 && (
              <button
                type="button"
                onClick={() => openPopover('assignees')}
                aria-label={t('flow.filters.morePeople', { count: overflow })}
                className="flow-focus flex h-7 min-w-7 items-center justify-center rounded-full px-1.5 text-[11px] font-semibold"
                style={{
                  marginInlineStart: 3,
                  background: 'var(--flow-hover)',
                  color: 'var(--flow-text-muted)',
                  boxShadow: '0 0 0 2px var(--flow-surface)',
                }}
              >
                +{overflow}
              </button>
            )}
            <button
              type="button"
              aria-pressed={filters.assignees.includes(UNASSIGNED)}
              aria-label={t('flow.filters.showPerson', { name: t('flow.filters.unassigned') })}
              title={t('flow.filters.unassigned')}
              data-filter-person={UNASSIGNED}
              onClick={() =>
                setFilters((f) => ({ ...f, assignees: toggleValue(f.assignees, UNASSIGNED) }))
              }
              className="flow-focus ms-1.5 flex h-7 w-7 items-center justify-center rounded-full"
              style={{
                border: '1px dashed var(--flow-border)',
                color: 'var(--flow-text-dim)',
                boxShadow: filters.assignees.includes(UNASSIGNED)
                  ? '0 0 0 2px var(--flow-surface), 0 0 0 4px var(--flow-action-bg)'
                  : undefined,
              }}
            >
              <UserX size={13} aria-hidden="true" />
            </button>
          </div>
        )}

        {people && currentUserId !== null && (
          <button
            type="button"
            aria-pressed={mine}
            onClick={() => setFilters((f) => toggleOnlyMine(f, currentUserId))}
            title={t('flow.filters.onlyMineHint')}
            className="flow-focus flex min-h-9 items-center gap-1.5 rounded-md px-2.5 text-flow-meta font-medium md:min-h-8"
            style={
              mine
                ? {
                    background: 'var(--accent-muted)',
                    color: 'var(--accent-on-muted)',
                    border: '1px solid var(--accent-border)',
                  }
                : { color: 'var(--flow-text-muted)', border: '1px solid var(--flow-border)' }
            }
          >
            <UserRound size={13} aria-hidden="true" />
            {t('flow.filters.onlyMine')}
          </button>
        )}

        <FilterPopover
          open={open}
          onOpenChange={setOpen}
          field={field}
          onFieldChange={setField}
          fields={fields}
          showDue={showDue}
          filters={filters}
          onChange={(next) => setFilters(next)}
          dim={dim}
          onDimChange={setDim}
          trigger={
            <button
              type="button"
              title={t('flow.filters.filterHint')}
              onClick={() => (open ? setOpen(false) : openPopover())}
              className="flow-focus flex min-h-9 items-center gap-1.5 rounded-md px-2.5 text-flow-meta font-medium md:min-h-8"
              style={
                count > 0
                  ? {
                      background: 'var(--accent-muted)',
                      color: 'var(--accent-on-muted)',
                      border: '1px solid var(--accent-border)',
                    }
                  : { color: 'var(--flow-text-muted)', border: '1px solid var(--flow-border)' }
              }
            >
              <ListFilter size={13} aria-hidden="true" />
              {t('flow.filters.filter')}
              {count > 0 && (
                <span
                  className="rounded-full px-1.5 text-[11px] font-semibold"
                  style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
                >
                  {count}
                </span>
              )}
            </button>
          }
        />

        {extra}

        {active && (
          <>
            <button
              type="button"
              onClick={clear}
              title={t('flow.filters.clearHint')}
              className="flow-focus min-h-9 rounded-md px-2 text-flow-meta font-semibold md:min-h-8"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.filters.clearAll')}
            </button>
            <span
              className="ms-auto text-flow-meta"
              style={{ color: 'var(--flow-text-dim)' }}
              role="status"
              data-filter-visible-count
            >
              {t('flow.filters.visibleCount', { visible, total })}
            </span>
          </>
        )}
      </div>

      {chips.length > 0 && (
        <div
          className={[
            'flex flex-wrap items-center gap-1.5',
            chips.every(chipHidden) ? 'md:hidden' : '',
          ].join(' ')}
          aria-label={t('flow.filters.active')}
        >
          {chips.map((chip) => (
            <button
              key={chip.id}
              type="button"
              data-chip={chip.id}
              onClick={() => setFilters(chip.next)}
              aria-label={t('flow.filters.removeChip', { label: chip.label })}
              className={`${chipHidden(chip) ? 'md:hidden ' : ''}flow-focus inline-flex min-h-7 items-center gap-1.5 rounded-full px-2.5 text-flow-meta font-medium`}
              style={{
                background: 'var(--accent-muted)',
                border: '1px solid var(--accent-border)',
                color: 'var(--flow-text)',
              }}
            >
              {chip.label}
              <X size={11} aria-hidden="true" />
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

/**
 * One chip per selected value. The search text has no chip because its box
 * shows it. Person chips are flagged so the bar can hide them where the avatar
 * strip is on screen (selected people are always pulled into the strip).
 */
export function buildChips(
  filters: TaskFilters,
  fields: readonly FilterFieldConfig[],
  t: (key: string, options?: Record<string, unknown>) => string
): Chip[] {
  const chips: Chip[] = [];
  for (const config of fields) {
    const selected = filters[config.field];
    for (const value of selected) {
      const option = config.options.find((o) => o.value === value);
      chips.push({
        id: `${config.field}:${value}`,
        label: `${config.label}: ${option?.label ?? t('flow.filters.unknownValue')}`,
        next: { ...filters, [config.field]: selected.filter((v) => v !== value) },
        person: config.field === 'assignees',
      });
    }
  }
  const dueLabel = t('flow.filters.fields.due');
  for (const window of filters.due) {
    chips.push({
      id: `due:${window}`,
      label: `${dueLabel}: ${t(DUE_LABEL_KEY[window])}`,
      next: { ...filters, due: filters.due.filter((w) => w !== window) },
    });
  }
  if (filters.dueFrom || filters.dueTo) {
    chips.push({
      id: 'due:range',
      label: `${dueLabel}: ${t('flow.filters.due.range', {
        from: filters.dueFrom ?? '…',
        to: filters.dueTo ?? '…',
      })}`,
      next: { ...filters, dueFrom: undefined, dueTo: undefined },
    });
  }
  return chips;
}

export default FilterBar;
