import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import OrgStructureStep from './OrgStructureStep';

describe('OrgStructureStep empty state', () => {
  const onNext = vi.fn();
  const onBack = vi.fn();

  beforeEach(() => {
    onNext.mockReset();
    onBack.mockReset();
  });

  it('shows the empty state and focuses the name input when the CTA is clicked', () => {
    render(<OrgStructureStep onNext={onNext} onBack={onBack} />);

    expect(screen.getByText('setup.orgStructure.noDepartments')).toBeInTheDocument();

    // Two elements share the accessible name "setup.orgStructure.addDepartment":
    // the row's add button carries it as an aria-label (visible text is
    // "common.add"), the EmptyState CTA carries it as its visible text.
    // getByText disambiguates on text content.
    fireEvent.click(screen.getByText('setup.orgStructure.addDepartment'));

    expect(screen.getByPlaceholderText('setup.orgStructure.departmentName')).toHaveFocus();
  });

  it('does not show the empty state once a department has been added', () => {
    render(<OrgStructureStep onNext={onNext} onBack={onBack} />);

    fireEvent.change(screen.getByPlaceholderText('setup.orgStructure.departmentName'), {
      target: { value: 'Kitchen' },
    });
    // The row's real add button is the first with this accessible name
    // (its aria-label); the EmptyState CTA - not yet clicked here - is
    // the second, appearing later in the DOM.
    fireEvent.click(screen.getAllByRole('button', { name: 'setup.orgStructure.addDepartment' })[0]);

    expect(screen.queryByText('setup.orgStructure.noDepartments')).not.toBeInTheDocument();
    expect(screen.getByText('Kitchen')).toBeInTheDocument();
  });
});
