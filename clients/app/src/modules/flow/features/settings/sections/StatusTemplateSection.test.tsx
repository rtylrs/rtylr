import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';
import { createEffectiveAccess } from '../../../../../../test/effectiveAccess';
import { initTestI18n } from '../../../shared/testing/i18n';
import type { ApiStatusTemplateItem } from '../../../shared/services/flowApi';
import StatusTemplateSection, { templateProblem } from './StatusTemplateSection';

const state = vi.hoisted(() => ({
  template: [] as ApiStatusTemplateItem[],
  save: vi.fn(),
  access: undefined as EffectiveAccess | undefined,
}));
vi.mock('../../projects/useStatuses', () => ({
  useStatusTemplate: () => ({ data: state.template, isPending: false, isError: false }),
  useSaveStatusTemplate: () => ({ mutate: state.save, isPending: false }),
}));
vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: (selector: (s: { user: { effectiveAccess?: EffectiveAccess } }) => unknown) =>
    selector({ user: { effectiveAccess: state.access } }),
}));

const TEMPLATE: ApiStatusTemplateItem[] = [
  { name: 'To Do', bucket: 'backlog', legacyStatus: 'todo', color: '#8b5cf6', sortOrder: 0 },
  { name: 'Doing', bucket: 'active', legacyStatus: 'in_progress', color: '#8b5cf6', sortOrder: 10 },
  { name: 'Done', bucket: 'done', legacyStatus: 'done', color: '#8b5cf6', sortOrder: 20 },
];

const names = (): string[] =>
  screen
    .getAllByRole('textbox', { name: 'Status name' })
    .map((el) => (el as HTMLInputElement).value);

beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  state.template = TEMPLATE;
  state.access = createEffectiveAccess('flow', { permissions: ['flow:team:manage'] });
});

describe('Work settings status template', () => {
  it('edits, reorders and saves the template, noting it only affects new projects', () => {
    render(<StatusTemplateSection />);
    expect(
      screen.getByText('Applies to new projects only. Existing projects keep their own statuses.')
    ).toBeInTheDocument();
    const save = screen.getByRole('button', { name: 'Save template' });
    expect(save).toBeDisabled();

    fireEvent.change(screen.getAllByRole('textbox', { name: 'Status name' })[1], {
      target: { value: 'Building' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Move Done up' }));
    fireEvent.change(screen.getByRole('textbox', { name: 'New status name' }), {
      target: { value: 'Blocked' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Add status' }));
    expect(names()).toEqual(['To Do', 'Done', 'Building', 'Blocked']);

    fireEvent.click(save);
    expect(state.save).toHaveBeenCalledWith(
      [
        expect.objectContaining({ name: 'To Do', bucket: 'backlog', sortOrder: 0 }),
        expect.objectContaining({ name: 'Done', bucket: 'done', sortOrder: 10 }),
        expect.objectContaining({ name: 'Building', bucket: 'active', sortOrder: 20 }),
        expect.objectContaining({ name: 'Blocked', bucket: 'active', sortOrder: 30 }),
      ],
      expect.any(Object)
    );
  });

  it('will not save a template without a done status', () => {
    render(<StatusTemplateSection />);
    fireEvent.click(screen.getByRole('button', { name: 'Remove Done' }));

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Include at least one status in the Done bucket.'
    );
    expect(screen.getByRole('button', { name: 'Save template' })).toBeDisabled();
  });

  it('is read-only without team management', () => {
    state.access = createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] });
    render(<StatusTemplateSection />);

    expect(
      screen.getByText('Only people who manage the Work team can change the template.')
    ).toBeInTheDocument();
    for (const input of screen.getAllByRole('textbox', { name: 'Status name' })) {
      expect(input).toBeDisabled();
    }
    expect(screen.queryByRole('button', { name: 'Save template' })).not.toBeInTheDocument();
    const firstRow = screen.getAllByRole('listitem')[0];
    expect(within(firstRow).getByRole('button', { name: 'Remove To Do' })).toBeDisabled();
  });
});

describe('templateProblem', () => {
  it('flags blank and duplicate names', () => {
    expect(templateProblem([{ name: ' ', bucket: 'done' }])).toBe(
      'flow.settings.statuses.errorName'
    );
    expect(
      templateProblem([
        { name: 'Done', bucket: 'done' },
        { name: 'done', bucket: 'active' },
      ])
    ).toBe('flow.settings.statuses.errorDuplicate');
    expect(templateProblem(TEMPLATE)).toBeNull();
  });
});
