import React from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../../shared/testing/i18n';
import type { ApiField } from '../../../shared/services/flowApi';

const mocks = vi.hoisted(() => ({
  fields: [] as ApiField[],
  create: vi.fn(),
  update: vi.fn(),
  reorder: vi.fn(),
  remove: vi.fn(),
}));

vi.mock('../useFields', () => ({
  useProjectFields: () => ({ data: mocks.fields, isLoading: false, isError: false }),
  useCreateField: () => ({ mutate: mocks.create, isPending: false }),
  useUpdateField: () => ({ mutate: mocks.update, isPending: false }),
  useReorderFields: () => ({ mutate: mocks.reorder, isPending: false }),
  useDeleteField: () => ({ mutate: mocks.remove, isPending: false }),
}));
vi.mock('../useProjectMembers', () => ({ useProjectMembers: () => ({ data: [] }) }));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import ProjectFieldsSettings from './ProjectFieldsSettings';

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

beforeEach(async () => {
  vi.clearAllMocks();
  await initTestI18n();
  mocks.fields = [
    field({
      id: '5',
      name: 'Severity',
      type: 'select',
      required: true,
      valueCount: 3,
      options: [{ id: 'a', label: 'Low', color: '#22c55e' }],
    }),
    field({ id: '6', name: 'Customer', valueCount: 0 }),
  ];
});

describe('Project Settings > Fields', () => {
  it('lists the fields with their type, required flag and how many tasks hold a value', () => {
    render(<ProjectFieldsSettings projectId="12" canManage />);
    const row = screen.getByText('Severity').closest('li') as HTMLElement;
    expect(within(row).getByText('Required')).toBeInTheDocument();
    expect(within(row).getByText(/Single select · 3 tasks have a value/)).toBeInTheDocument();
  });

  it('says how many tasks lose a value before deleting', async () => {
    render(<ProjectFieldsSettings projectId="12" canManage />);
    fireEvent.click(screen.getByRole('button', { name: 'Delete Severity' }));
    expect(
      await screen.findByText(
        '3 tasks have a value for this field. Those values will be cleared. The tasks themselves stay.'
      )
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Delete field' }));
    await waitFor(() => expect(mocks.remove).toHaveBeenCalledWith('5', expect.anything()));
  });

  it('moves a field with the whole new order', () => {
    render(<ProjectFieldsSettings projectId="12" canManage />);
    fireEvent.click(screen.getByRole('button', { name: 'Move Customer up' }));
    expect(mocks.reorder).toHaveBeenCalledWith(['6', '5']);
  });

  it('keeps a renamed choice under its id so values survive', async () => {
    render(<ProjectFieldsSettings projectId="12" canManage />);
    fireEvent.click(screen.getByRole('button', { name: 'Edit Severity' }));
    const choice = await screen.findByRole('textbox', { name: 'Choice 1' });
    fireEvent.change(choice, { target: { value: 'Minor' } });
    fireEvent.change(screen.getByRole('textbox', { name: 'Field name' }), {
      target: { value: 'Impact' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(mocks.update).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '5',
        name: 'Impact',
        options: [{ id: 'a', label: 'Minor', color: '#22c55e' }],
        required: true,
      }),
      expect.anything()
    );
  });

  it('is read-only with the reason for someone who is not a project admin', () => {
    render(<ProjectFieldsSettings projectId="12" canManage={false} />);
    expect(screen.getByText('Only project admins can manage fields.')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add field' })).toBeNull();
    expect(screen.getByRole('button', { name: 'Delete Severity' })).toBeDisabled();
  });
});
