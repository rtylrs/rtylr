import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Reorder from './Reorder';

const mocks = vi.hoisted(() => ({
  useReorderSuggestions: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

    const title = screen.getByText('reorder.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('reorder.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'reorder.refresh' }));

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

    expect(screen.queryByText('reorder.empty.title')).not.toBeInTheDocument();
  });
});
