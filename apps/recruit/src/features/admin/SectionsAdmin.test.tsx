import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * SectionsAdmin empty state - the old copy told the admin to "Click
 * Add section" without the empty state's own button doing anything.
 * Verifies the CTA now opens the same AddSectionPicker the header's
 * "Add section" button opens.
 */

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}));

// The mocked query result must be referentially stable across
// re-renders - SectionsAdmin mirrors data.sections into local state
// via a useEffect keyed on that reference, so a fresh array literal
// returned on every call would re-fire the effect every render and
// spin forever.
const sectionsQueryResult = { data: { sections: [] as unknown[] }, isLoading: false };

vi.mock('../../shared/hooks/queries/useAdmin', () => ({
  useAdminSections: () => sectionsQueryResult,
  useCreateSection: () => ({ mutate: vi.fn() }),
  useUpdateSection: () => ({ mutate: vi.fn() }),
  useDeleteSection: () => ({ mutate: vi.fn() }),
  useReorderSections: () => ({ mutate: vi.fn() }),
}));

import SectionsAdmin from './SectionsAdmin';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('SectionsAdmin empty state', () => {
  it('shows the empty message and opens the section-kind picker from its own CTA', () => {
    render(<SectionsAdmin />);

    expect(screen.getByText('common.emptyStates.sectionsNone')).toBeInTheDocument();
    expect(screen.queryByText('Pick a section kind')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.emptyStates.addSection' }));

    expect(screen.getByText('Pick a section kind')).toBeInTheDocument();
  });
});
