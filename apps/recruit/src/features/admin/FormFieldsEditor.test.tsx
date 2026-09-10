import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * FormFieldsEditor empty state. Unlike the other admin screens there
 * was no single "Add question" action to reuse - the real add flow
 * is a row of per-kind buttons at the bottom of the screen. The CTA
 * invents a sensible default: create a short-text question, the same
 * request shape the "Short text" button already sends.
 */

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

beforeEach(() => {
  vi.clearAllMocks();
});

describe('FormFieldsEditor empty state', () => {
  it('shows the empty message and creates a default short-text question from its CTA', () => {
    render(<FormFieldsEditor careerId={42} />);

    expect(screen.getByText('common.emptyStates.questionsNone')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.emptyStates.addQuestion' }));

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
