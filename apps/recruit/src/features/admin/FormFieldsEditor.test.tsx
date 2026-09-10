import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll, beforeEach } from 'vitest';

/**
 * FormFieldsEditor empty state. Unlike the other admin screens there
 * was no single "Add question" action to reuse - the real add flow
 * is a row of per-kind buttons at the bottom of the screen. The CTA
 * invents a sensible default: create a short-text question, the same
 * request shape the "Short text" button already sends.
 */

// i18next is deliberately not mocked. The seed label this test asserts on is
// persisted and then shown to the candidate, so an identity `t` would have the
// test agreeing that a dotted path is the right thing to store.

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const createMutate = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useAdmin', () => ({
  useAdminFormFields: () => ({ data: { fields: [] }, isLoading: false }),
  useCreateFormField: () => ({ mutate: createMutate }),
  useUpdateFormField: () => ({ mutate: vi.fn() }),
  useDeleteFormField: () => ({ mutate: vi.fn() }),
}));

import FormFieldsEditor from './FormFieldsEditor';
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe('FormFieldsEditor empty state', () => {
  it('shows the empty message and creates a default short-text question from its CTA', () => {
    render(<FormFieldsEditor careerId={42} />);

    expect(
      screen.getByText('No custom questions yet. Add one to start collecting more from applicants.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Add question' }));

    expect(createMutate).toHaveBeenCalledTimes(1);
    const [body] = createMutate.mock.calls[0];
    expect(body).toMatchObject({
      careerId: 42,
      kind: 'short_text',
      label: 'New question',
      required: false,
    });
  });
});
