import type {
  ApiField,
  ApiFieldType,
  ApiTaskFieldValue,
  ApiTaskFieldValueInput,
} from '../services/flowApi';

/**
 * Custom fields (T-579): a project defines fields on the server and every task
 * holds one value per field. Everything here is pure so the rules the server
 * enforces (required, defaults, value shapes) can be checked before a request.
 */

export const FIELD_TYPES: readonly ApiFieldType[] = [
  'text',
  'number',
  'date',
  'select',
  'multi_select',
  'person',
  'url',
  'checkbox',
];

/** Types whose value is a list (option ids or user ids). */
export function isMultiField(type: ApiFieldType): boolean {
  return type === 'multi_select' || type === 'person';
}

export function hasOptions(type: ApiFieldType): boolean {
  return type === 'select' || type === 'multi_select';
}

/** Types the filter popover can list as a checklist of values. */
export function isListFilterable(type: ApiFieldType): boolean {
  return hasOptions(type) || type === 'person' || type === 'checkbox';
}

export function isEmptyFieldValue(value: ApiTaskFieldValue | undefined): boolean {
  return !value || ((value.value ?? '') === '' && (value.values ?? []).length === 0);
}

/**
 * The value's parts as a flat list: one entry for a single-value field, every
 * entry for a multi one. An unchecked checkbox has none, because to anyone
 * filtering or grouping, unchecked and never set are the same thing (the
 * server's checkbox filter treats them the same way).
 */
export function fieldValueParts(value: ApiTaskFieldValue | undefined): string[] {
  if (!value) return [];
  if (value.values && value.values.length > 0) return value.values;
  if (!value.value || value.value === 'false') return [];
  return [value.value];
}

/** The value a new task gets for a field it was not given. */
export function defaultFieldValue(field: ApiField): ApiTaskFieldValue | undefined {
  if (isMultiField(field.type)) {
    return field.defaultValues.length > 0 ? { values: field.defaultValues } : undefined;
  }
  return field.defaultValue ? { value: field.defaultValue } : undefined;
}

/** The starting values of a new-task form: every field's default. */
export function initialFieldValues(fields: readonly ApiField[]): Record<string, ApiTaskFieldValue> {
  const out: Record<string, ApiTaskFieldValue> = {};
  for (const field of fields) {
    const value = defaultFieldValue(field);
    if (value) out[field.id] = value;
  }
  return out;
}

/** Required fields a new task would be created without, in field order. */
export function missingRequiredFields(
  fields: readonly ApiField[],
  values: Readonly<Record<string, ApiTaskFieldValue>>
): ApiField[] {
  return fields.filter(
    (f) => f.required && isEmptyFieldValue(values[f.id]) && !defaultFieldValue(f)
  );
}

/** The createTask / setTaskFieldValues payload for the values that are set. */
export function toFieldValueInputs(
  fields: readonly ApiField[],
  values: Readonly<Record<string, ApiTaskFieldValue>>
): ApiTaskFieldValueInput[] {
  const out: ApiTaskFieldValueInput[] = [];
  for (const field of fields) {
    const value = values[field.id];
    if (isEmptyFieldValue(value)) continue;
    out.push(
      isMultiField(field.type)
        ? { fieldId: field.id, values: value?.values ?? [] }
        : { fieldId: field.id, value: value?.value ?? '' }
    );
  }
  return out;
}

/** One write for one field; an empty value clears it. */
export function fieldValueInput(field: ApiField, value: ApiTaskFieldValue): ApiTaskFieldValueInput {
  return isMultiField(field.type)
    ? { fieldId: field.id, values: value.values ?? [] }
    : { fieldId: field.id, value: value.value ?? '' };
}

const NUMBER = /^-?\d{1,15}(\.\d{1,6})?$/;

/** Whether a typed value is one the server will accept, so the input can refuse it first. */
export function isValidFieldInput(type: ApiFieldType, raw: string): boolean {
  const value = raw.trim();
  if (value === '') return true;
  if (type === 'number') return NUMBER.test(value);
  if (type === 'url') {
    try {
      const url = new URL(value);
      return (url.protocol === 'http:' || url.protocol === 'https:') && url.host !== '';
    } catch {
      return false;
    }
  }
  return true;
}

/** Field id in a filter or group-by key: `field:<id>`. */
export const FIELD_KEY_PREFIX = 'field:';

export function fieldKey(fieldId: string): `field:${string}` {
  return `${FIELD_KEY_PREFIX}${fieldId}`;
}

export function fieldIdFromKey(key: string): string | null {
  return key.startsWith(FIELD_KEY_PREFIX) ? key.slice(FIELD_KEY_PREFIX.length) || null : null;
}

/**
 * The text a value reads as on a card or in a chip. People and options are
 * looked up by id; an id that no longer resolves is left out rather than shown
 * raw.
 */
export function formatFieldValue(
  field: ApiField,
  value: ApiTaskFieldValue | undefined,
  ctx: { personName: (id: string) => string | undefined; yes: string; no: string }
): string {
  if (isEmptyFieldValue(value)) return '';
  switch (field.type) {
    case 'select':
    case 'multi_select':
      return fieldValueParts(value)
        .map((id) => field.options.find((o) => o.id === id)?.label)
        .filter(Boolean)
        .join(', ');
    case 'person':
      return fieldValueParts(value)
        .map((id) => ctx.personName(id))
        .filter(Boolean)
        .join(', ');
    case 'checkbox':
      return value?.value === 'true' ? ctx.yes : ctx.no;
    default:
      return value?.value ?? '';
  }
}
