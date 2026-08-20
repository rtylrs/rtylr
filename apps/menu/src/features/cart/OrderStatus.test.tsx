import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '999' }),
  useRouterState: () => '',
  useNavigate: () => navigateSpy,
}));

const apiGet = vi.hoisted(() => vi.fn());
vi.mock('../../shared/lib/api', () => ({
  default: { get: apiGet },
}));

import OrderStatus from './OrderStatus';

describe('OrderStatus not-found state', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
  });

  it('shows the shared EmptyState and navigates back to the menu on the CTA', async () => {
    apiGet.mockRejectedValue({ response: { status: 404 } });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText('notFoundTitle')).toBeInTheDocument());
    expect(screen.getByText('notFoundDescription')).toBeInTheDocument();

    // The always-on "Order Another" CTA should not double up with the
    // EmptyState's own action once the order is confirmed not found.
    expect(screen.queryByText('Order Another')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'backToMenu' }));

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/' });
  });
});
