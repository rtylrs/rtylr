import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const templatesList: unknown[] = [];

vi.mock('../../shared/hooks/queries/useTemplates', () => ({
  useTemplates: () => ({ data: templatesList, isLoading: false }),
  useCreateTemplate: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

import Templates from './Templates';

function renderTemplates(): void {
  render(<Templates />);
}

beforeEach(() => {
  templatesList.length = 0;
});

describe('Templates empty state', () => {
  it('shows a "new template" call-to-action that opens the create-template form', () => {
    renderTemplates();

    expect(screen.getByText('crm.templates.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.templates.emptyDescription')).toBeInTheDocument();

    // Header and empty-state actions share one key now, so both carry the same
    // accessible name. Assert both exist rather than matching one by luck.
    const buttons = screen.getAllByRole('button', { name: 'crm.templates.new' });
    expect(buttons).toHaveLength(2);
    fireEvent.click(buttons[1]);

    expect(screen.getByRole('heading', { name: 'crm.templates.new' })).toBeInTheDocument();
  });
});
