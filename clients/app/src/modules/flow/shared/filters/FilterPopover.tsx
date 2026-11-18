import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, Search } from 'lucide-react';
import { Avatar, Popover, PopoverContent, PopoverTrigger, Switch } from '@voxire/ui';
import { normalizeSearchText } from './taskFilters';
import {
  DUE_WINDOWS,
  hasDueFilter,
  toggleValue,
  type DueWindow,
  type FilterField,
  type TaskFilters,
} from './taskFilters';
import type { FilterFieldConfig, FilterOption } from './filterFields';

export const DUE_LABEL_KEY: Record<DueWindow, string> = {
  overdue: 'flow.filters.due.overdue',
  today: 'flow.filters.due.today',
  next_7: 'flow.filters.due.next7',
  next_30: 'flow.filters.due.next30',
  no_date: 'flow.filters.due.noDate',
};

interface FilterPopoverProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** The field the right pane shows; null shows the "pick a field" hint. */
  field: FilterField | null;
  onFieldChange: (field: FilterField) => void;
  fields: FilterFieldConfig[];
  showDue: boolean;
  filters: TaskFilters;
  onChange: (next: TaskFilters) => void;
  dim: boolean;
  onDimChange: (dim: boolean) => void;
  trigger: React.ReactNode;
}

/**
 * Jira's two-pane filter: fields on the left, the chosen field's values on the
 * right as a searchable checklist. Each field shows how many values it holds,
 * so the left pane alone tells you what is narrowing the list.
 */
export const FilterPopover: React.FC<FilterPopoverProps> = ({
  open,
  onOpenChange,
  field,
  onFieldChange,
  fields,
  showDue,
  filters,
  onChange,
  dim,
  onDimChange,
  trigger,
}) => {
  const { t, i18n } = useTranslation();
  const config = fields.find((f) => f.field === field);
  const selectedCount = (f: FilterField): number =>
    f === 'due'
      ? filters.due.length + (filters.dueFrom || filters.dueTo ? 1 : 0)
      : filters[f].length;

  const leftItems: { field: FilterField; label: string }[] = [
    ...fields.map((f) => ({ field: f.field as FilterField, label: f.label })),
    ...(showDue ? [{ field: 'due' as FilterField, label: t('flow.filters.fields.due') }] : []),
  ];

  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger asChild>{trigger}</PopoverTrigger>
      <PopoverContent
        align={i18n.dir() === 'rtl' ? 'end' : 'start'}
        aria-label={t('flow.filters.title')}
        className="w-[min(560px,calc(100vw-1.5rem))] overflow-hidden rounded-lg shadow-lg"
        style={{ background: 'var(--flow-surface)', border: '1px solid var(--flow-border)' }}
      >
        <div className="flex min-h-[260px] flex-col sm:flex-row">
          <ul
            role="tablist"
            aria-orientation="vertical"
            className="flex shrink-0 gap-0.5 overflow-x-auto p-2 sm:w-[180px] sm:flex-col sm:overflow-visible"
            style={{ borderInlineEnd: '1px solid var(--flow-border)' }}
          >
            {leftItems.map((item) => {
              const active = item.field === field;
              const count = selectedCount(item.field);
              return (
                <li key={item.field}>
                  <button
                    type="button"
                    role="tab"
                    aria-selected={active}
                    data-filter-field={item.field}
                    onClick={() => onFieldChange(item.field)}
                    className="flow-focus flex min-h-9 w-full items-center justify-between gap-2 whitespace-nowrap rounded-md px-2.5 text-start text-flow-meta font-medium"
                    style={
                      active
                        ? { background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }
                        : { color: 'var(--flow-text)' }
                    }
                  >
                    {item.label}
                    {count > 0 && (
                      <span
                        className="rounded-full px-1.5 text-flow-meta font-semibold"
                        style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
                      >
                        {count}
                      </span>
                    )}
                  </button>
                </li>
              );
            })}
          </ul>

          <div className="min-w-0 flex-1 p-2" role="tabpanel">
            {field === 'due' ? (
              <DuePane filters={filters} onChange={onChange} />
            ) : config ? (
              <OptionsPane
                key={config.field}
                config={config}
                selected={filters[config.field]}
                onToggle={(value) =>
                  onChange({
                    ...filters,
                    [config.field]: toggleValue(filters[config.field], value),
                  })
                }
                onClearField={() => onChange({ ...filters, [config.field]: [] })}
              />
            ) : (
              <p
                className="flex h-full items-center justify-center p-4 text-center text-flow-meta"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.filters.pickField')}
              </p>
            )}
          </div>
        </div>

        <div
          className="flex flex-wrap items-center justify-between gap-2 px-3 py-2"
          style={{ borderTop: '1px solid var(--flow-border)' }}
        >
          <label
            className="flex items-center gap-2 text-flow-meta"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <Switch
              checked={dim}
              onCheckedChange={onDimChange}
              aria-label={t('flow.filters.dim')}
            />
            {t('flow.filters.dim')}
          </label>
          <button
            type="button"
            disabled={leftItems.every((i) => selectedCount(i.field) === 0)}
            onClick={() =>
              onChange({
                ...filters,
                assignees: [],
                priorities: [],
                labels: [],
                due: [],
                dueFrom: undefined,
                dueTo: undefined,
                statuses: [],
                cycles: [],
                milestones: [],
                projects: [],
              })
            }
            className="flow-focus min-h-8 rounded-md px-2 text-flow-meta font-semibold disabled:opacity-40"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.filters.clearAll')}
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
};

const OptionsPane: React.FC<{
  config: FilterFieldConfig;
  selected: readonly string[];
  onToggle: (value: string) => void;
  onClearField: () => void;
}> = ({ config, selected, onToggle, onClearField }) => {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const shown = useMemo(() => {
    const q = normalizeSearchText(query.trim());
    return q === ''
      ? config.options
      : config.options.filter((o) => normalizeSearchText(o.label).includes(q));
  }, [config.options, query]);

  return (
    <div className="flex h-full flex-col">
      <label
        className="mb-1.5 flex min-h-9 items-center gap-2 rounded-md px-2"
        style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text-dim)' }}
      >
        <Search size={13} aria-hidden="true" />
        <input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t('flow.filters.searchField', { field: config.label })}
          aria-label={t('flow.filters.searchField', { field: config.label })}
          className="min-w-0 flex-1 bg-transparent text-flow-meta"
          style={{ color: 'var(--flow-text)' }}
        />
      </label>
      <ul className="max-h-[260px] flex-1 overflow-y-auto" aria-label={config.label}>
        {shown.length === 0 && (
          <li className="p-3 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.filters.noOptions')}
          </li>
        )}
        {shown.map((option) => (
          <li key={option.value}>
            <OptionRow
              option={option}
              checked={selected.includes(option.value)}
              onToggle={() => onToggle(option.value)}
            />
          </li>
        ))}
      </ul>
      <div className="flex items-center justify-between pt-1.5">
        <button
          type="button"
          disabled={selected.length === 0}
          onClick={onClearField}
          className="flow-focus min-h-8 rounded-md px-2 text-flow-meta font-medium disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.filters.clearField')}
        </button>
        <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.filters.shownOf', { shown: shown.length, total: config.options.length })}
        </span>
      </div>
    </div>
  );
};

const OptionRow: React.FC<{ option: FilterOption; checked: boolean; onToggle: () => void }> = ({
  option,
  checked,
  onToggle,
}) => (
  <button
    type="button"
    role="checkbox"
    aria-checked={checked}
    data-filter-option={option.value}
    onClick={onToggle}
    className="flow-focus flex min-h-9 w-full items-center gap-2 rounded-md px-2 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
    style={{ color: 'var(--flow-text)' }}
  >
    <span
      className="flex h-4 w-4 shrink-0 items-center justify-center rounded-[4px]"
      style={{
        border: `1px solid ${checked ? 'var(--flow-action-bg)' : 'var(--flow-text-dim)'}`,
        background: checked ? 'var(--flow-action-bg)' : 'transparent',
        color: 'var(--flow-on-accent)',
      }}
      aria-hidden="true"
    >
      {checked && <Check size={11} strokeWidth={3} />}
    </span>
    {option.initials !== undefined && <Avatar size="xs" initials={option.initials} aria-hidden />}
    {option.color && (
      <span
        className="h-2.5 w-2.5 shrink-0 rounded-full"
        style={{ background: option.color }}
        aria-hidden="true"
      />
    )}
    <span className="truncate">{option.label}</span>
  </button>
);

const DuePane: React.FC<{ filters: TaskFilters; onChange: (next: TaskFilters) => void }> = ({
  filters,
  onChange,
}) => {
  const { t } = useTranslation();
  const dateInput = (key: 'dueFrom' | 'dueTo', label: string): React.ReactNode => (
    <label
      className="flex flex-1 flex-col gap-1 text-flow-meta"
      style={{ color: 'var(--flow-text-muted)' }}
    >
      {label}
      <input
        type="date"
        value={filters[key] ?? ''}
        onChange={(e) => onChange({ ...filters, [key]: e.target.value || undefined })}
        className="min-h-9 rounded-md bg-transparent px-2 text-flow-meta"
        style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text)' }}
      />
    </label>
  );
  return (
    <div className="flex h-full flex-col">
      <ul aria-label={t('flow.filters.fields.due')}>
        {DUE_WINDOWS.map((window) => (
          <li key={window}>
            <OptionRow
              option={{ value: window, label: t(DUE_LABEL_KEY[window]) }}
              checked={filters.due.includes(window)}
              onToggle={() => onChange({ ...filters, due: toggleValue(filters.due, window) })}
            />
          </li>
        ))}
      </ul>
      <p
        className="mt-3 mb-1.5 px-2 text-flow-meta font-semibold"
        style={{ color: 'var(--flow-text)' }}
      >
        {t('flow.filters.due.custom')}
      </p>
      <div className="flex gap-2 px-2">
        {dateInput('dueFrom', t('flow.filters.due.from'))}
        {dateInput('dueTo', t('flow.filters.due.to'))}
      </div>
      <div className="pt-2">
        <button
          type="button"
          disabled={!hasDueFilter(filters)}
          onClick={() => onChange({ ...filters, due: [], dueFrom: undefined, dueTo: undefined })}
          className="flow-focus min-h-8 rounded-md px-2 text-flow-meta font-medium disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.filters.clearField')}
        </button>
      </div>
    </div>
  );
};

export default FilterPopover;
