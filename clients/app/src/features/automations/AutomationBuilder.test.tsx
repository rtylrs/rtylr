import path from 'node:path';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  search: {} as { id?: number },
  automation: {} as { data?: unknown; isLoading?: boolean; isError?: boolean },
  updateMutate: vi.fn(),
  createMutate: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  useSearch: () => mocks.search,
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../modules/crm/shared/hooks/queries/useAutomations', () => ({
  useGetAutomation: () => mocks.automation,
  useCreateAutomation: () => ({ mutate: mocks.createMutate, isPending: false }),
  useUpdateAutomation: () => ({ mutate: mocks.updateMutate, isPending: false }),
  triggerTypeFor: (value: string) => value,
  actionTypeFor: (value: unknown) => value,
}));

import AutomationBuilder from './AutomationBuilder';

const nameInput = () => screen.queryByPlaceholderText('Welcome series');

const existingAutomation = {
  id: 5,
  name: 'Welcome series',
  description: 'New customer · 1 action(s)',
  trigger: 'New customer',
  triggerType: 'customer_signup',
  triggerConfig: { sendDelayMinutes: 30 },
  actions: [
    { type: 'send_email', subject: 'Hello', body: 'Welcome!', template_id: 'existing-template' },
  ],
  status: 'active' as const,
  isActive: true,
  enrolled: 3,
  completed: 0,
  createdAt: '2026-01-01',
  nodes: [
    {
      id: 'n1',
      type: 'trigger',
      label: 'New customer',
      automationType: 'customer_signup',
      x: 0,
      y: 0,
    },
    {
      id: 'n2',
      type: 'action',
      label: 'Send email',
      automationType: 'send_email',
      x: 0,
      y: 80,
    },
  ],
};

beforeEach(async () => {
  await initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'), 'en', [
    path.resolve(import.meta.dirname, '../../modules/crm/locales'),
  ]);
  mocks.navigate.mockClear();
  mocks.updateMutate.mockClear();
  mocks.createMutate.mockClear();
  mocks.search = { id: 5 };
  mocks.automation = { data: undefined, isLoading: false, isError: false };
});

describe('AutomationBuilder edit load gate (B25-R06)', () => {
  it('does not render the default form while the automation is loading', () => {
    mocks.automation = { data: undefined, isLoading: true };

    render(<AutomationBuilder />);

    expect(nameInput()).toBeNull();
    expect(screen.queryByRole('button', { name: 'Save workflow' })).not.toBeInTheDocument();
    expect(mocks.updateMutate).not.toHaveBeenCalled();
  });

  it('shows an error instead of an editable default form when the load failed', () => {
    mocks.automation = { data: undefined, isLoading: false, isError: true };

    render(<AutomationBuilder />);

    expect(
      screen.getByText('Unable to load this automation. Check your connection and try again.')
    ).toBeInTheDocument();
    expect(nameInput()).toBeNull();
    expect(screen.queryByRole('button', { name: 'Save workflow' })).not.toBeInTheDocument();
  });

  it('saves a loaded edit with its trigger, actions and run state intact', () => {
    mocks.automation = { data: existingAutomation, isLoading: false, isError: false };

    render(<AutomationBuilder />);

    expect(nameInput()).toHaveValue('Welcome series');
    fireEvent.click(screen.getByRole('button', { name: 'Save workflow' }));

    expect(mocks.updateMutate).toHaveBeenCalledWith(
      {
        id: 5,
        data: expect.objectContaining({
          name: 'Welcome series',
          triggerType: 'customer_signup',
          actions: [
            {
              type: 'send_email',
              subject: 'Hello',
              body: 'Welcome!',
              template_id: 'existing-template',
            },
          ],
          isActive: true,
          status: 'active',
          triggerConfig: { sendDelayMinutes: 30 },
        }),
      },
      expect.any(Object)
    );
  });

  it('still builds a create form when there is no id', () => {
    mocks.search = {};
    mocks.automation = { data: undefined, isLoading: false, isError: false };

    render(<AutomationBuilder />);

    const input = nameInput();
    expect(input).not.toBeNull();
    fireEvent.change(input as HTMLElement, { target: { value: 'Birthday flow' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send email' }));
    fireEvent.change(screen.getByLabelText('Email subject'), { target: { value: 'Thanks' } });
    fireEvent.change(screen.getByLabelText('Email message'), { target: { value: 'Thank you!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save workflow' }));

    expect(mocks.createMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Birthday flow',
        triggerType: 'order_created',
        isActive: false,
      }),
      expect.any(Object)
    );
    expect(mocks.updateMutate).not.toHaveBeenCalled();
  });
});

it('requires an action and nonblank email content before saving a new workflow', () => {
  mocks.search = {};
  render(<AutomationBuilder />);
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: '   ' } });
  expect(screen.getByRole('button', { name: 'Save workflow' })).toBeDisabled();
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Order thanks' } });
  expect(screen.getByRole('button', { name: 'Save workflow' })).toBeDisabled();
  fireEvent.click(screen.getByRole('button', { name: 'Send email' }));
  fireEvent.change(screen.getByLabelText('Email subject'), { target: { value: 'Thanks' } });
  expect(screen.getByRole('button', { name: 'Save workflow' })).toBeDisabled();
  fireEvent.change(screen.getByLabelText('Email message'), {
    target: { value: 'Thanks for ordering.' },
  });
  expect(screen.getByRole('button', { name: 'Save workflow' })).toBeEnabled();
  fireEvent.click(screen.getByRole('button', { name: 'Remove action 1' }));
  expect(screen.getByRole('button', { name: 'Save workflow' })).toBeDisabled();
});

it('keeps action configuration when changing the trigger and does not overwrite local edits on refetch', () => {
  mocks.automation = { data: existingAutomation };
  const { rerender } = render(<AutomationBuilder />);
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Changed locally' } });
  fireEvent.change(screen.getByLabelText('When'), { target: { value: 'order_created' } });
  mocks.automation = { data: { ...existingAutomation, name: 'Server refetch' } };
  rerender(<AutomationBuilder />);
  expect(screen.getByLabelText('Name')).toHaveValue('Changed locally');
  expect(screen.getByLabelText('Email message')).toHaveValue('Welcome!');
  fireEvent.click(screen.getByRole('button', { name: 'Save workflow' }));
  expect(mocks.updateMutate).toHaveBeenCalledWith(
    expect.objectContaining({
      data: expect.objectContaining({
        name: 'Changed locally',
        triggerType: 'order_created',
        actions: existingAutomation.actions,
      }),
    }),
    expect.any(Object)
  );
});
