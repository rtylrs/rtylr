import { describe, expect, it } from 'vitest';
import type { Task } from '../../shared/stores/flowStore';
import type { ApiField } from '../../shared/services/flowApi';
import { NO_FIELD_LANE, fieldGroupOptions, fieldLanes, laneKeyFor } from './taskGroups';

const severity: ApiField = {
  id: '5',
  projectId: '12',
  name: 'Severity',
  type: 'select',
  options: [
    { id: 'a', label: 'Low', color: '#0f0' },
    { id: 'b', label: 'High', color: '#f00' },
  ],
  required: false,
  defaultValue: '',
  defaultValues: [],
  sortOrder: 0,
  valueCount: 0,
};

const task = (fieldValues?: Task['fieldValues']): Task =>
  ({ id: '1', priority: 'medium', fieldValues }) as Task;

describe('grouping the board by a custom field', () => {
  it('offers single selects only', () => {
    expect(
      fieldGroupOptions([severity, { ...severity, id: '6', name: 'Tags', type: 'multi_select' }])
    ).toEqual([{ key: 'field:5', label: 'Severity' }]);
  });

  it('puts each task in its option lane, or "No value"', () => {
    expect(laneKeyFor(task({ '5': { value: 'b' } }), 'field:5')).toBe('b');
    expect(laneKeyFor(task(), 'field:5')).toBe(NO_FIELD_LANE);
  });

  it('lists lanes in option order with "No value" last', () => {
    expect(fieldLanes(severity).map((l) => l.key)).toEqual(['a', 'b', NO_FIELD_LANE]);
  });
});
