import { describe, expect, it } from 'vitest';
import { formatWorkEntryTarget } from './timeEntryLabel';

describe('formatWorkEntryTarget', () => {
  it('renders the identifier and title when the task has an identifier', () => {
    expect(
      formatWorkEntryTarget({ identifier: 'ENG-12', title: 'Fix the gate hinge' }, undefined)
    ).toBe('ENG-12 Fix the gate hinge');
  });

  it('falls back to the bare title when the task has no identifier', () => {
    expect(formatWorkEntryTarget({ title: 'Fix the gate hinge' }, { name: 'Site fitout' })).toBe(
      'Fix the gate hinge'
    );
  });

  it('falls back to the project name when there is no task', () => {
    expect(formatWorkEntryTarget(undefined, { name: 'Site fitout' })).toBe('Site fitout');
  });

  it('returns an empty string when neither task nor project is known', () => {
    expect(formatWorkEntryTarget(undefined, undefined)).toBe('');
  });
});
