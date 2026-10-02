import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const workTypesState = vi.hoisted(() => ({
  data: { workTypes: [] as Array<Record<string, unknown>> },
}));
const createMutate = vi.hoisted(() => vi.fn());

// Same reason as TimeSection.test.tsx: this hook module reaches the real
// i18next singleton for toasts, so it is mocked directly rather than relying
// on the react-i18next mock above to cover it.
vi.mock('../../time/useWorkDay', () => ({
  useWorkTypes: () => ({ data: workTypesState.data, isLoading: false }),
  useCreateWorkType: () => ({ mutate: createMutate, isPending: false }),
  useUpdateWorkType: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteWorkType: () => ({ mutate: vi.fn(), isPending: false }),
}));

import WorkTypesSection from './WorkTypesSection';

describe('WorkTypesSection empty state', () => {
  it('shows the empty title with an Add CTA that focuses the name input', () => {
    workTypesState.data = { workTypes: [] };
    render(<WorkTypesSection />);

    expect(screen.getByText('flow.settings.workTypes.empty')).toBeInTheDocument();
    const button = screen.getByRole('button', { name: 'flow.settings.workTypes.addFirstCta' });
    expect(button).toBeEnabled();

    const input = screen.getByLabelText('flow.settings.workTypes.name');
    fireEvent.click(button);
    expect(input).toHaveFocus();
  });
});

describe('WorkTypesSection with data', () => {
  it('renders each work type instead of the empty state', () => {
    workTypesState.data = {
      workTypes: [{ id: '1', name: 'Development', description: 'Coding work' }],
    };
    render(<WorkTypesSection />);

    expect(screen.queryByText('flow.settings.workTypes.empty')).not.toBeInTheDocument();
    expect(screen.getByText('Development')).toBeInTheDocument();
  });
});
