import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll, beforeEach } from 'vitest';

/**
 * SectionsAdmin empty state - the old copy told the admin to "Click
 * Add section" without the empty state's own button doing anything.
 * Verifies the CTA now opens the same AddSectionPicker the header's
 * "Add section" button opens.
 *
 * i18next is deliberately not mocked, so the English below also proves the
 * keys resolve. See rtylr#1270 for what an identity `t` hides.
 */

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
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe('SectionsAdmin empty state', () => {
  it('shows the empty message and opens the section-kind picker from its own CTA', () => {
    render(<SectionsAdmin />);

    expect(
      screen.getByText('No sections yet. Add one to start building this page.')
    ).toBeInTheDocument();
    expect(screen.queryByText('Pick a section kind')).not.toBeInTheDocument();

    // With the copy resolving, the header button and the empty state's own CTA
    // read the same, which they did not while the CTA was rendering its key.
    // The point of this test is the second one, so clicking the header button
    // instead would let it pass without proving anything.
    const addButtons = screen.getAllByRole('button', { name: 'Add section' });
    expect(addButtons).toHaveLength(2);
    fireEvent.click(addButtons[1]);

    expect(screen.getByText('Pick a section kind')).toBeInTheDocument();
  });
});
