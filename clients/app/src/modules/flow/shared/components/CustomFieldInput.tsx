import React from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { DatePicker, Switch } from '@voxire/ui';
import type { ApiField, ApiTaskFieldValue } from '../services/flowApi';
import { isValidFieldInput } from '../lib/customFields';
import { getInitials } from '../lib/utils';
import FlowSelect, { FlowMultiSelect, type FlowSelectOption } from './FlowSelect';
import TaskTextField from './TaskTextField';
import AssigneeAvatar from './AssigneeAvatar';

export interface CustomFieldPerson {
  id: string;
  name: string;
}

interface CustomFieldInputProps {
  field: ApiField;
  value: ApiTaskFieldValue | undefined;
  /** Called with the new value; an empty value clears the field. */
  onChange: (value: ApiTaskFieldValue) => void;
  /** People a person field can name: the project's members. */
  people: readonly CustomFieldPerson[];
  disabled?: boolean;
}

const inputClass =
  'w-full rounded-md px-2.5 py-1.5 text-flow-meta placeholder:opacity-50 disabled:opacity-70';
const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

function optionDot(color: string): React.ReactNode {
  return (
    <span
      aria-hidden="true"
      className="h-2 w-2 shrink-0 rounded-full"
      style={{ background: color || 'var(--flow-text-dim)' }}
    />
  );
}

/**
 * The editor for one custom field value, by type. It is built from the inputs
 * the task panel already uses (TaskTextField, FlowSelect, the date picker and
 * Switch), so a custom field edits exactly like a built-in property, in the
 * task panel and in the New task form alike.
 */
export const CustomFieldInput: React.FC<CustomFieldInputProps> = ({
  field,
  value,
  onChange,
  people,
  disabled = false,
}) => {
  const { t, i18n } = useTranslation();
  const label = field.name;

  const saveText = (raw: string): void => {
    const next = raw.trim();
    if (!isValidFieldInput(field.type, next)) {
      toast.error(
        t(field.type === 'url' ? 'flow.fields.invalidUrl' : 'flow.fields.invalidNumber', {
          name: field.name,
        })
      );
      return;
    }
    onChange({ value: next });
  };

  switch (field.type) {
    case 'text':
    case 'url':
    case 'number':
      return (
        <TaskTextField
          key={`${field.id}-${value?.value ?? ''}`}
          type={field.type === 'number' ? 'number' : field.type === 'url' ? 'url' : 'text'}
          step={field.type === 'number' ? 'any' : undefined}
          value={value?.value ?? ''}
          onSave={saveText}
          placeholder={t(`flow.fields.placeholder.${field.type}`)}
          aria-label={label}
          disabled={disabled}
          className={inputClass}
          style={inputStyle}
        />
      );
    case 'date':
      return (
        <DatePicker
          value={value?.value ?? ''}
          onChange={(date) => onChange({ value: date })}
          locale={i18n.language}
          disabled={disabled}
          clearable
          aria-label={label}
          className="h-9 min-h-0 py-1.5 text-flow-meta"
          labels={{
            placeholder: t('flow.fields.placeholder.date'),
            previousMonth: t('flowCommon.datePicker.previousMonth'),
            nextMonth: t('flowCommon.datePicker.nextMonth'),
          }}
        />
      );
    case 'checkbox':
      return (
        <Switch
          checked={value?.value === 'true'}
          onCheckedChange={(checked) => onChange({ value: checked ? 'true' : 'false' })}
          disabled={disabled}
          aria-label={label}
        />
      );
    case 'select':
      return (
        <FlowSelect
          label={label}
          value={value?.value ?? ''}
          options={[
            { value: '', label: t('flow.fields.noValue') },
            ...field.options.map((o) => ({
              value: o.id,
              label: o.label,
              icon: optionDot(o.color),
            })),
          ]}
          onChange={(next) => onChange({ value: next })}
          disabled={disabled}
          muted={!value?.value}
        />
      );
    case 'multi_select':
      return (
        <FlowMultiSelect
          label={label}
          values={value?.values ?? []}
          options={field.options.map((o) => ({
            value: o.id,
            label: o.label,
            icon: optionDot(o.color),
          }))}
          onChange={(values) => onChange({ values })}
          placeholder={t('flow.fields.noValue')}
          disabled={disabled}
        />
      );
    case 'person': {
      const picked = value?.values ?? [];
      // Someone picked earlier who has since left the project stays listed,
      // so the value can still be seen and removed.
      const options: FlowSelectOption[] = [
        ...people.map((p) => ({ value: p.id, label: p.name })),
        ...picked
          .filter((id) => !people.some((p) => p.id === id))
          .map((id) => ({ value: id, label: t('flow.fields.formerMember') })),
      ].map((o) => ({
        ...o,
        icon: (
          <AssigneeAvatar name={o.label} initials={getInitials(o.label)} seed={o.value} size={16} />
        ),
      }));
      return (
        <FlowMultiSelect
          label={label}
          values={picked}
          options={options}
          onChange={(values) => onChange({ values })}
          placeholder={t('flow.fields.noValue')}
          disabled={disabled}
        />
      );
    }
  }
};

export default CustomFieldInput;
