import { describe, expect, it } from 'vitest';
import type { TFunction } from 'i18next';
import { branchNameFor, templateFor } from './taskTemplates';

const t = ((key: string) => key) as unknown as TFunction;

describe('templateFor', () => {
  it("prefers the project's template, then the built-in bug one", () => {
    const own = [{ id: '1', projectId: '3', taskType: 'bug', body: '## Ours' }];
    expect(templateFor('bug', own, t)).toBe('## Ours');
    expect(templateFor('bug', [], t)).toBe('flow.taskTemplates.defaultBug');
    expect(templateFor('feature', [], t)).toBe('');
  });
});

describe('branchNameFor', () => {
  it('prefixes by type and slugs the title', () => {
    expect(
      branchNameFor({ identifier: 'IH-42', title: 'Login button: fix it!', taskType: 'bug' })
    ).toBe('fix/ih-42-login-button-fix-it');
    expect(branchNameFor({ identifier: 'T-7', title: 'Add CSV import', taskType: 'feature' })).toBe(
      'feat/t-7-add-csv-import'
    );
    expect(branchNameFor({ identifier: 'T-8', title: 'Tidy', taskType: undefined })).toBe(
      'task/t-8-tidy'
    );
  });

  it('keeps only the key for a title with no Latin letters', () => {
    expect(branchNameFor({ identifier: 'T-9', title: 'إصلاح الزر', taskType: 'bug' })).toBe(
      'fix/t-9'
    );
  });
});
