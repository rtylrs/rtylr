import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

vi.mock('@voxire/config', () => ({
  TIMEZONES: [{ value: 'UTC', label: 'UTC' }],
}));

import LocationsStep from './LocationsStep';

describe('LocationsStep empty state', () => {
  const onNext = vi.fn();
  const onBack = vi.fn();

  beforeEach(() => {
    onNext.mockReset();
    onBack.mockReset();
  });

  it('shows the empty state and focuses the name input when the CTA is clicked', () => {
    render(<LocationsStep onNext={onNext} onBack={onBack} />);

    expect(screen.getByText('hrSetup.locations.noLocations')).toBeInTheDocument();

    // Two buttons share the accessible name "hrSetup.locations.addLocation" here:
    // the row's "+" button carries it as an aria-label, the EmptyState CTA
    // carries it as its visible text. getByText disambiguates on text content.
    fireEvent.click(screen.getByText('hrSetup.locations.addLocation'));

    expect(screen.getByPlaceholderText('hrSetup.locations.locationName')).toHaveFocus();
  });

  it('does not show the empty state once a location has been added', () => {
    render(<LocationsStep onNext={onNext} onBack={onBack} />);

    fireEvent.change(screen.getByPlaceholderText('hrSetup.locations.locationName'), {
      target: { value: 'HQ' },
    });
    fireEvent.change(screen.getByPlaceholderText('hrSetup.locations.locationAddress'), {
      target: { value: '123 Main St' },
    });
    // The row's real "+" button is the first button with this accessible
    // name (its aria-label); the EmptyState CTA - not yet clicked here -
    // would be the second, appearing later in the DOM.
    fireEvent.click(screen.getAllByRole('button', { name: 'hrSetup.locations.addLocation' })[0]);

    expect(screen.queryByText('hrSetup.locations.noLocations')).not.toBeInTheDocument();
    expect(screen.getByText('HQ')).toBeInTheDocument();
  });
});
