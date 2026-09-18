import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Reorder from './Reorder';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useReorderSuggestions: vi.fn(),
  refetch: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('./useReorderSuggestions', () => ({
  useReorderSuggestions: mocks.useReorderSuggestions,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Reorder empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useReorderSuggestions.mockReturnValue({
      data: { suggestions: [] },
      isLoading: false,
      refetch: mocks.refetch,
    });
  });

  it('shows the illustration, empty copy, and refetches on CTA click', () => {
    render(<Reorder />);

    const title = screen.getByText('No suggestions');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('All stock levels are healthy')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when there are suggestions', () => {
    mocks.useReorderSuggestions.mockReturnValue({
      data: {
        suggestions: [
          {
            productId: 'p1',
            name: 'Widget',
            sku: 'W-1',
            currentStock: 2,
            avgDailySales: 1,
            suggestedQuantity: 10,
            leadTimeDays: 3,
            confidence: 'high',
          },
        ],
      },
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<Reorder />);

    expect(screen.queryByText('No suggestions')).not.toBeInTheDocument();
  });
});
