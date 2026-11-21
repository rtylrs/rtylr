import { describe, expect, it } from 'vitest';
import type { ApiField } from '../services/flowApi';
import {
  fieldIdFromKey,
  fieldKey,
  fieldValueParts,
  formatFieldValue,
  initialFieldValues,
  isValidFieldInput,
  missingRequiredFields,
  toFieldValueInputs,
} from './customFields';

function field(overrides: Partial<ApiField>): ApiField {
  return {
    id: '1',
    projectId: '12',
    name: 'Field',
    type: 'text',
    options: [],
    required: false,
    defaultValue: '',
    defaultValues: [],
    sortOrder: 0,
    valueCount: 0,
    ...overrides,
  };
}

const severity = field({
  id: '5',
  name: 'Severity',
  type: 'select',
  options: [
    { id: 'a', label: 'Low', color: '#0f0' },
    { id: 'b', label: 'High', color: '#f00' },
  ],
});

describe('custom field values', () => {
  it('treats an unchecked checkbox like no value', () => {
    expect(fieldValueParts({ value: 'false' })).toEqual([]);
    expect(fieldValueParts({ value: 'true' })).toEqual(['true']);
    expect(fieldValueParts({ values: ['1', '2'] })).toEqual(['1', '2']);
    expect(fieldValueParts(undefined)).toEqual([]);
  });

  it('starts a new task on each default and reports required fields without one', () => {
    const fields = [
      severity,
      field({ id: '6', name: 'Team', required: true }),
      field({ id: '7', name: 'Area', required: true, defaultValue: 'Ops' }),
      field({ id: '8', name: 'Owners', type: 'person', defaultValues: ['3'] }),
    ];
    const values = initialFieldValues(fields);
    expect(values).toEqual({ '7': { value: 'Ops' }, '8': { values: ['3'] } });
    expect(missingRequiredFields(fields, values).map((f) => f.name)).toEqual(['Team']);
    expect(missingRequiredFields(fields, { ...values, '6': { value: 'Core' } })).toEqual([]);
  });

  it('sends only set values, in the shape of each type', () => {
    const fields = [severity, field({ id: '8', type: 'multi_select' }), field({ id: '9' })];
    expect(
      toFieldValueInputs(fields, { '5': { value: 'b' }, '8': { values: ['x'] }, '9': { value: '' } })
    ).toEqual([
      { fieldId: '5', value: 'b' },
      { fieldId: '8', values: ['x'] },
    ]);
  });

  it('refuses numbers and links the server would refuse', () => {
    expect(isValidFieldInput('number', '12.5')).toBe(true);
    expect(isValidFieldInput('number', '1e3')).toBe(false);
    expect(isValidFieldInput('url', 'https://rtylr.com')).toBe(true);
    expect(isValidFieldInput('url', 'javascript:alert(1)')).toBe(false);
    expect(isValidFieldInput('url', '')).toBe(true);
  });

  it('reads a value as text by option label and person name', () => {
    const ctx = { personName: (id: string) => ({ '3': 'Ada' })[id], yes: 'Yes', no: 'No' };
    expect(formatFieldValue(severity, { value: 'b' }, ctx)).toBe('High');
    expect(formatFieldValue(field({ type: 'person' }), { values: ['3', '99'] }, ctx)).toBe('Ada');
    expect(formatFieldValue(field({ type: 'checkbox' }), { value: 'false' }, ctx)).toBe('No');
  });

  it('round-trips a field key', () => {
    expect(fieldIdFromKey(fieldKey('42'))).toBe('42');
    expect(fieldIdFromKey('priority')).toBeNull();
  });
});
