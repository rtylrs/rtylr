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

    fireEvent.click(screen.getByRole('button', { name: 'crm.templates.new' }));

    expect(screen.getByText('New Template')).toBeInTheDocument();
  });
});
