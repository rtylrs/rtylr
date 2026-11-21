import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronUp, Columns3, Download } from 'lucide-react';
import { Popover, PopoverContent, PopoverTrigger } from '@voxire/ui';
import {
  DEFAULT_LIST_PREFS,
  LIST_COLUMNS,
  moveListColumn,
  toggleListColumn,
  type ListColumnId,
  type ListColumnPrefs,
} from './listColumns';

const toolbarButton =
  'flow-focus flex h-8 items-center gap-1.5 rounded-md px-2.5 text-flow-meta font-medium transition-colors hover:bg-[var(--flow-hover)] disabled:opacity-50';
/** The view header is crowded below 2xl: the buttons drop to their icons there. */
const toolbarLabel = 'hidden 2xl:inline';
const toolbarButtonStyle: React.CSSProperties = {
  color: 'var(--flow-text-muted)',
  background: 'var(--flow-surface)',
  border: '1px solid var(--flow-border)',
};

/**
 * Which columns the List shows and in what order. The title always shows.
 * Custom fields are not offered yet: the task list does not carry their values.
 */
export function ListColumnPicker({
  prefs,
  onChange,
}: {
  prefs: ListColumnPrefs;
  onChange: (update: (prefs: ListColumnPrefs) => ListColumnPrefs) => void;
}) {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  // Shown columns in their order, then the rest in their usual order.
  const ordered: ListColumnId[] = [
    ...prefs.columns,
    ...LIST_COLUMNS.map((column) => column.id).filter((id) => !prefs.columns.includes(id)),
  ];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={t('flow.projectList.columns')}
          title={t('flow.projectList.columns')}
          className={toolbarButton}
          style={toolbarButtonStyle}
        >
          <Columns3 size={14} aria-hidden="true" />
          <span className={toolbarLabel}>{t('flow.projectList.columns')}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent
        align={i18n?.dir?.() === 'rtl' ? 'start' : 'end'}
        aria-label={t('flow.projectList.columns')}
        className="w-[260px] rounded-lg p-2 shadow-lg"
        style={{ background: 'var(--flow-surface)', border: '1px solid var(--flow-border)' }}
      >
        <ul className="flex flex-col gap-0.5">
          {ordered.map((id) => {
            const column = LIST_COLUMNS.find((c) => c.id === id)!;
            const label = t(column.labelKey);
            const shown = prefs.columns.includes(id);
            const index = prefs.columns.indexOf(id);
            return (
              <li
                key={id}
                className="flex items-center gap-2 rounded-md px-2 py-1 hover:bg-[var(--flow-hover)]"
              >
                <label
                  className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-flow-meta"
                  style={{ color: 'var(--flow-text)' }}
                >
                  <input
                    type="checkbox"
                    checked={shown}
                    disabled={id === 'title'}
                    onChange={() => onChange((current) => toggleListColumn(current, id))}
                    className="h-3.5 w-3.5 shrink-0 accent-[var(--accent)]"
                  />
                  <span className="truncate">{label}</span>
                </label>
                {shown && (
                  <span className="flex shrink-0 items-center">
                    <button
                      type="button"
                      disabled={index === 0}
                      aria-label={t('flow.projectList.moveColumnEarlier', { column: label })}
                      onClick={() => onChange((current) => moveListColumn(current, id, -1))}
                      className="flow-focus flex h-6 w-6 items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] disabled:opacity-30"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      <ChevronUp size={13} aria-hidden="true" />
                    </button>
                    <button
                      type="button"
                      disabled={index === prefs.columns.length - 1}
                      aria-label={t('flow.projectList.moveColumnLater', { column: label })}
                      onClick={() => onChange((current) => moveListColumn(current, id, 1))}
                      className="flow-focus flex h-6 w-6 items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] disabled:opacity-30"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      <ChevronDown size={13} aria-hidden="true" />
                    </button>
                  </span>
                )}
              </li>
            );
          })}
        </ul>
        <div className="mt-2 pt-2" style={{ borderTop: '1px solid var(--flow-border)' }}>
          <button
            type="button"
            onClick={() =>
              onChange((current) => ({
                ...current,
                columns: [...DEFAULT_LIST_PREFS.columns],
                widths: {},
                // A sort on a column the reset hides would order by something unseen.
                sort:
                  current.sort && DEFAULT_LIST_PREFS.columns.includes(current.sort.column)
                    ? current.sort
                    : null,
              }))
            }
            className="flow-focus w-full rounded-md px-2 py-1.5 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.projectList.resetColumns')}
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

/** Downloads the rows on screen (after the filters) as CSV, built in the browser. */
export function ExportCsvButton({
  disabled,
  onExport,
}: {
  disabled: boolean;
  onExport: () => void;
}) {
  const { t } = useTranslation();
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onExport}
      aria-label={t('flow.projectList.exportCsv')}
      title={t('flow.projectList.exportCsv')}
      className={toolbarButton}
      style={toolbarButtonStyle}
    >
      <Download size={14} aria-hidden="true" />
      <span className={toolbarLabel}>{t('flow.projectList.exportCsv')}</span>
    </button>
  );
}
