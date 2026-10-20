import React, { useId, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, ChevronDown, ListFilter, Search, X } from 'lucide-react';
import { Popover, PopoverClose, PopoverContent, PopoverTrigger } from '@voxire/ui';
import { BILLINGS, FOLLOW_UP_WINDOWS, type Presence } from '../data/model';
import { type LeadListFilters, useLeadFacets, useLeadImports, useReps } from '../data/queries';
import {
  activeLeadFilterGroups,
  clearLeadFilterGroup,
  type LeadFilterGroup,
  NO_LEAD_FILTERS,
} from '../data/leadFilters';

type Change = (change: Partial<LeadListFilters>) => void;
type Option = { value: string; label: string };

/** Filters whose value is a Presence, keyed to their label strings. */
const PRESENCE = {
  hasPhone: 'phone',
  hasEmail: 'email',
  hasWebsite: 'website',
  hasLinkedin: 'linkedin',
  hasCompany: 'company',
  hasValue: 'value',
  contacted: 'contacted',
} as const;
type PresenceField = keyof typeof PRESENCE;

const SET_GROUPS = ['reps', 'sources', 'industries', 'sizes', 'countries', 'cities'] as const;
type SetGroup = (typeof SET_GROUPS)[number];

/** The pickable values for each set filter, with readable labels. */
function useSetOptions(): Record<SetGroup, Option[]> {
  const { t } = useTranslation();
  const { data: facets } = useLeadFacets();
  const { data: reps = [] } = useReps();
  return useMemo(() => {
    const plain = (values?: string[]) => (values ?? []).map((value) => ({ value, label: value }));
    return {
      reps: [
        { value: '0', label: t('crmSales.filters.unassigned') },
        ...reps.map((rep) => ({ value: rep.id, label: rep.name })),
      ],
      sources: plain(facets?.sources),
      industries: plain(facets?.industries),
      sizes: plain(facets?.sizes),
      countries: plain(facets?.countries),
      cities: plain(facets?.cities),
    };
  }, [facets, reps, t]);
}

// ── Controls ────────────────────────────────────────────────────────────

/** Any / yes / no, as a radio group of segments. */
const PresenceControl: React.FC<{
  label: string;
  value: Presence;
  onChange: (value: Presence) => void;
  yes: string;
  no: string;
}> = ({ label, value, onChange, yes, no }) => {
  const { t } = useTranslation();
  const options: Array<[Presence, string]> = [
    ['', t('crmSales.filters.any')],
    ['yes', yes],
    ['no', no],
  ];
  return (
    <div className="sales-segmented" role="radiogroup" aria-label={label}>
      {options.map(([option, text]) => (
        <button
          key={option || 'any'}
          type="button"
          role="radio"
          aria-checked={value === option}
          onClick={() => onChange(option)}
        >
          {text}
        </button>
      ))}
    </div>
  );
};

/** A searchable checklist in its own popover, summarised on its trigger. */
const MultiSelect: React.FC<{
  label: string;
  options: Option[];
  value: string[];
  onChange: (value: string[]) => void;
}> = ({ label, options, value, onChange }) => {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const labels = new Map(options.map((option) => [option.value, option.label]));
  const chosen = value.map((item) => labels.get(item) ?? item);
  const summary =
    chosen.length === 0
      ? t('crmSales.filters.any')
      : chosen.length <= 2
        ? chosen.join(', ')
        : t('crmSales.filters.selected', { count: chosen.length });
  const needle = query.trim().toLocaleLowerCase();
  const shown = needle
    ? options.filter((option) => option.label.toLocaleLowerCase().includes(needle))
    : options;
  const toggle = (item: string) =>
    onChange(value.includes(item) ? value.filter((v) => v !== item) : [...value, item]);

  return (
    <Popover onOpenChange={(open) => !open && setQuery('')}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="sales-multi-trigger"
          aria-label={`${label}: ${summary}`}
          data-active={value.length > 0 || undefined}
        >
          <span>{summary}</span>
          <ChevronDown size={14} aria-hidden="true" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="crm-popover sales-multi-pop" aria-label={label}>
        {options.length > 6 && (
          <label className="sales-multi-search">
            <Search size={13} aria-hidden="true" />
            <input
              autoFocus
              value={query}
              placeholder={t('crmSales.filters.searchOptions')}
              aria-label={t('crmSales.filters.searchOptions')}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>
        )}
        <div className="sales-multi-list" role="group" aria-label={label}>
          {options.length === 0 ? (
            <p className="sales-multi-empty">{t('crmSales.filters.noOptions')}</p>
          ) : shown.length === 0 ? (
            <p className="sales-multi-empty">{t('crmSales.filters.noMatches')}</p>
          ) : (
            shown.map((option) => {
              const checked = value.includes(option.value);
              return (
                <button
                  key={option.value}
                  type="button"
                  role="checkbox"
                  aria-checked={checked}
                  className="sales-multi-option"
                  onClick={() => toggle(option.value)}
                >
                  <span className="sales-multi-box" aria-hidden="true">
                    {checked && <Check size={11} strokeWidth={3} />}
                  </span>
                  <span className="sales-multi-label">{option.label}</span>
                </button>
              );
            })
          )}
        </div>
        {value.length > 0 && (
          <div className="sales-multi-foot">
            <button type="button" className="sales-link-button" onClick={() => onChange([])}>
              {t('crmSales.filters.clearAll')}
            </button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
};

/** Two inputs bounding a value; the filter ignores a side until it parses. */
const RangeControl: React.FC<{
  label: string;
  type: 'number' | 'date';
  from: string;
  to: string;
  onChange: (from: string, to: string) => void;
}> = ({ label, type, from, to, onChange }) => {
  const { t } = useTranslation();
  const [low, high] =
    type === 'date'
      ? [t('crmSales.filters.from'), t('crmSales.filters.to')]
      : [t('crmSales.filters.min'), t('crmSales.filters.max')];
  const props = type === 'number' ? { inputMode: 'decimal' as const, min: 0 } : {};
  return (
    <div className="sales-range" role="group" aria-label={label}>
      <input
        className="sales-input"
        type={type}
        {...props}
        placeholder={low}
        aria-label={`${label} ${low}`}
        value={from}
        max={type === 'date' ? to || undefined : undefined}
        onChange={(event) => onChange(event.target.value, to)}
      />
      <span aria-hidden="true">–</span>
      <input
        className="sales-input"
        type={type}
        {...props}
        placeholder={high}
        aria-label={`${label} ${high}`}
        value={to}
        min={type === 'date' ? from || undefined : 0}
        onChange={(event) => onChange(from, event.target.value)}
      />
    </div>
  );
};

const Row: React.FC<{
  label: string;
  htmlFor?: string;
  children: React.ReactNode;
  wide?: boolean;
}> = ({ label, htmlFor, children, wide }) => (
  <div className={wide ? 'sales-filter-row is-wide' : 'sales-filter-row'}>
    {htmlFor ? <label htmlFor={htmlFor}>{label}</label> : <span>{label}</span>}
    <div className="sales-filter-control">{children}</div>
  </div>
);

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <section className="sales-filter-section">
    <h3>{title}</h3>
    {children}
  </section>
);

// ── Popover ─────────────────────────────────────────────────────────────

/** The Filter button and its popover of every lead filter. */
export const LeadFilterPopover: React.FC<{
  filters: LeadListFilters;
  onChange: Change;
  total?: number;
}> = ({ filters, onChange, total }) => {
  const { t, i18n } = useTranslation();
  const options = useSetOptions();
  const { data: imports = [] } = useLeadImports();
  const ids = { followUp: useId(), importId: useId() };
  const active = activeLeadFilterGroups(filters).length;
  const presence = (field: PresenceField) => (
    <Row label={t(`crmSales.filters.presence.${PRESENCE[field]}.label`)}>
      <PresenceControl
        label={t(`crmSales.filters.presence.${PRESENCE[field]}.label`)}
        value={filters[field]}
        onChange={(value) => onChange({ [field]: value })}
        yes={field === 'contacted' ? t('crmSales.filters.contactedYes') : t('crmSales.filters.has')}
        no={
          field === 'contacted' ? t('crmSales.filters.contactedNo') : t('crmSales.filters.missing')
        }
      />
    </Row>
  );
  const facet = (group: SetGroup, label: string) => (
    <Row label={label}>
      <MultiSelect
        label={label}
        options={options[group]}
        value={filters[group]}
        onChange={(value) => onChange({ [group]: value })}
      />
    </Row>
  );

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button type="button" className="sales-filter-button" data-active={active > 0 || undefined}>
          <ListFilter size={14} aria-hidden="true" />
          {t('crmSales.filters.button')}
          {active > 0 && <span className="sales-filter-count">{active}</span>}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align={i18n.dir() === 'rtl' ? 'start' : 'end'}
        className="crm-popover sales-filter-pop"
        aria-label={t('crmSales.filters.title')}
      >
        <header className="sales-filter-head">
          <h2>{t('crmSales.filters.title')}</h2>
          <div className="sales-filter-head-actions">
            {active > 0 && (
              <button
                type="button"
                className="sales-link-button"
                onClick={() => onChange(NO_LEAD_FILTERS)}
              >
                {t('crmSales.filters.clearAll')}
              </button>
            )}
            <PopoverClose className="sales-filter-done">
              {total === undefined
                ? t('crmSales.filters.title')
                : t('crmSales.filters.showResults', { count: total })}
            </PopoverClose>
            <PopoverClose className="sales-icon-button" aria-label={t('crmSales.filters.close')}>
              <X size={15} />
            </PopoverClose>
          </div>
        </header>

        <div className="sales-filter-body">
          <Section title={t('crmSales.filters.sections.contact')}>
            {presence('hasPhone')}
            {presence('hasEmail')}
            {presence('hasWebsite')}
            {presence('hasLinkedin')}
            {presence('hasCompany')}
          </Section>
          <Section title={t('crmSales.filters.sections.owner')}>
            {facet('reps', t('crmSales.fields.rep'))}
            {facet('sources', t('crmSales.fields.source'))}
            <Row label={t('crmSales.filters.import')} htmlFor={ids.importId}>
              <select
                id={ids.importId}
                className="sales-select"
                value={filters.importId ?? ''}
                onChange={(event) => onChange({ importId: event.target.value || undefined })}
              >
                <option value="">{t('crmSales.filters.anyImport')}</option>
                {imports.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.fileName}
                  </option>
                ))}
              </select>
            </Row>
          </Section>
          <Section title={t('crmSales.filters.sections.company')}>
            {facet('industries', t('crmSales.fields.industry'))}
            {facet('sizes', t('crmSales.fields.size'))}
            {facet('countries', t('crmSales.fields.country'))}
            {facet('cities', t('crmSales.fields.city'))}
          </Section>
          <Section title={t('crmSales.filters.sections.activity')}>
            <Row label={t('crmSales.fields.followUp')} htmlFor={ids.followUp}>
              <select
                id={ids.followUp}
                className="sales-select"
                value={filters.followUp}
                onChange={(event) =>
                  onChange({ followUp: event.target.value as LeadListFilters['followUp'] })
                }
              >
                <option value="">{t('crmSales.filters.anyFollowUp')}</option>
                {FOLLOW_UP_WINDOWS.map((window) => (
                  <option key={window} value={window}>
                    {t(`crmSales.followUpWindow.${window}`)}
                  </option>
                ))}
              </select>
            </Row>
            {presence('contacted')}
            <Row label={t('crmSales.filters.attempts')} wide>
              <RangeControl
                label={t('crmSales.filters.attempts')}
                type="number"
                from={filters.attemptsMin}
                to={filters.attemptsMax}
                onChange={(attemptsMin, attemptsMax) => onChange({ attemptsMin, attemptsMax })}
              />
            </Row>
            <Row label={t('crmSales.filters.lastActivity')} wide>
              <RangeControl
                label={t('crmSales.filters.lastActivity')}
                type="date"
                from={filters.activeFrom}
                to={filters.activeTo}
                onChange={(activeFrom, activeTo) => onChange({ activeFrom, activeTo })}
              />
            </Row>
          </Section>
          <Section title={t('crmSales.filters.sections.value')}>
            {presence('hasValue')}
            <Row label={t('crmSales.fields.estimatedValue')} wide>
              <RangeControl
                label={t('crmSales.fields.estimatedValue')}
                type="number"
                from={filters.valueMin}
                to={filters.valueMax}
                onChange={(valueMin, valueMax) => onChange({ valueMin, valueMax })}
              />
            </Row>
            <Row label={t('crmSales.fields.billing')} wide>
              <div
                className="sales-chip-row"
                role="group"
                aria-label={t('crmSales.fields.billing')}
              >
                {BILLINGS.map((billing) => {
                  const on = filters.billing.includes(billing);
                  return (
                    <button
                      key={billing}
                      type="button"
                      className="sales-chip"
                      aria-pressed={on}
                      onClick={() =>
                        onChange({
                          billing: on
                            ? filters.billing.filter((b) => b !== billing)
                            : [...filters.billing, billing],
                        })
                      }
                    >
                      {t(`crmSales.billing.${billing}`)}
                    </button>
                  );
                })}
              </div>
            </Row>
          </Section>
          <Section title={t('crmSales.filters.sections.dates')}>
            <Row label={t('crmSales.filters.added')} wide>
              <RangeControl
                label={t('crmSales.filters.added')}
                type="date"
                from={filters.createdFrom}
                to={filters.createdTo}
                onChange={(createdFrom, createdTo) => onChange({ createdFrom, createdTo })}
              />
            </Row>
          </Section>
        </div>
      </PopoverContent>
    </Popover>
  );
};

// ── Chips ───────────────────────────────────────────────────────────────

/** The active filters as removable chips, so the list says what it shows. */
export const LeadFilterChips: React.FC<{ filters: LeadListFilters; onChange: Change }> = ({
  filters,
  onChange,
}) => {
  const { t, i18n } = useTranslation();
  const options = useSetOptions();
  const { data: imports = [] } = useLeadImports();
  const number = (value: string) =>
    Number.isFinite(Number(value))
      ? new Intl.NumberFormat(i18n.language).format(Number(value))
      : value;
  const date = (value: string) =>
    new Intl.DateTimeFormat(i18n.language, { dateStyle: 'medium', timeZone: 'UTC' }).format(
      new Date(`${value}T00:00:00Z`)
    );
  const range = (from: string, to: string, kind: 'number' | 'date') => {
    const format = kind === 'date' ? date : number;
    if (from && to) return `${format(from)} – ${format(to)}`;
    return from
      ? t(`crmSales.filters.chip.${kind === 'date' ? 'since' : 'atLeast'}`, { value: format(from) })
      : t(`crmSales.filters.chip.${kind === 'date' ? 'until' : 'atMost'}`, { value: format(to) });
  };
  const list = (values: string[]) =>
    values.length <= 2
      ? values.join(', ')
      : t('crmSales.filters.chip.more', {
          values: values.slice(0, 2).join(', '),
          count: values.length - 2,
        });

  const label = (group: LeadFilterGroup): string => {
    if (group in PRESENCE) {
      const field = group as PresenceField;
      return t(`crmSales.filters.presence.${PRESENCE[field]}.${filters[field]}`);
    }
    if ((SET_GROUPS as readonly string[]).includes(group)) {
      const set = group as SetGroup;
      const labels = new Map(options[set].map((option) => [option.value, option.label]));
      return t(`crmSales.filters.chip.${set}`, {
        values: list(filters[set].map((value) => labels.get(value) ?? value)),
      });
    }
    switch (group) {
      case 'billing':
        return t('crmSales.filters.chip.billing', {
          values: list(filters.billing.map((billing) => t(`crmSales.billing.${billing}`))),
        });
      case 'importId':
        return t('crmSales.filters.chip.importId', {
          values: imports.find((item) => item.id === filters.importId)?.fileName ?? '…',
        });
      case 'followUp':
        return t('crmSales.filters.chip.followUp', {
          values: t(`crmSales.followUpWindow.${filters.followUp}`),
        });
      case 'attempts':
        return t('crmSales.filters.chip.attempts', {
          values: range(filters.attemptsMin, filters.attemptsMax, 'number'),
        });
      case 'value':
        return t('crmSales.filters.chip.value', {
          values: range(filters.valueMin, filters.valueMax, 'number'),
        });
      case 'created':
        return t('crmSales.filters.chip.created', {
          values: range(filters.createdFrom, filters.createdTo, 'date'),
        });
      case 'active':
        return t('crmSales.filters.chip.active', {
          values: range(filters.activeFrom, filters.activeTo, 'date'),
        });
    }
    return group;
  };

  const groups = activeLeadFilterGroups(filters);
  if (groups.length === 0) return null;
  return (
    <ul className="sales-filter-chips">
      {groups.map((group) => {
        const text = label(group);
        return (
          <li key={group} className="sales-filter-chip">
            <span>{text}</span>
            <button
              type="button"
              aria-label={t('crmSales.filters.chip.remove', { label: text })}
              onClick={() => onChange(clearLeadFilterGroup(group))}
            >
              <X size={12} />
            </button>
          </li>
        );
      })}
    </ul>
  );
};
