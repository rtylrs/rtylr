import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { toast } from 'sonner';
import { initTestI18n } from '../../../shared/testing/i18n';
import type { ApiField } from '../../../shared/services/flowApi';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../../test/flowSelect';

const mocks = vi.hoisted(() => ({ fields: [] as ApiField[], save: vi.fn() }));
vi.mock('../../projects/useFields', () => ({
  useProjectFields: () => ({ data: mocks.fields, isError: false }),
  useSetTaskFieldValues: () => ({ mutate: mocks.save, isPending: false }),
}));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import TaskCustomFields from './TaskCustomFields';

const Row: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div>
    <span>{label}</span>
    {children}
  </div>
);

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
      options: [
        { id: 'a', label: 'Low', color: '#22c55e' },
        { id: 'b', label: 'High', color: '#ef4444' },
      ],
    }),
    field({ id: '6', name: 'Customer' }),
  ];
});

function renderFields() {
  render(
    <TaskCustomFields
      taskId="77"
      projectId="12"
      values={{ '5': { value: 'a' }, '6': { value: 'Acme' } }}
      canEdit
      people={[]}
      Row={Row}
    />
  );
}

describe('Custom fields in the task panel', () => {
  it('shows each value with the field name, required ones marked', () => {
    renderFields();
    expect(flowSelectTrigger('Severity')).toHaveAccessibleName('Severity: Low');
    expect(screen.getByText('Severity *')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'Customer' })).toHaveValue('Acme');
  });

  it('writes a changed value for that field alone', async () => {
    renderFields();
    await chooseFlowOption('Severity', 'High');
    expect(mocks.save).toHaveBeenCalledWith([{ fieldId: '5', value: 'b' }]);

    const customer = screen.getByRole('textbox', { name: 'Customer' });
    fireEvent.focus(customer);
    fireEvent.change(customer, { target: { value: 'Globex' } });
    fireEvent.blur(customer);
    expect(mocks.save).toHaveBeenCalledWith([{ fieldId: '6', value: 'Globex' }]);
  });

  it('refuses to clear a required field', async () => {
    renderFields();
    await chooseFlowOption('Severity', 'No value');
    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith('Fill in the required fields first: Severity')
    );
    expect(mocks.save).not.toHaveBeenCalled();
  });
});
