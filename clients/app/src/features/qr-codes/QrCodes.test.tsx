import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

type QrTable = { id: number; name: string; capacity: number; status: string; qrCodeUrl: string };

const tablesState = vi.hoisted(() => ({
  data: [] as QrTable[],
  isLoading: false,
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: () => tablesState,
}));

import QrCodes from './QrCodes';

beforeEach(() => {
  vi.clearAllMocks();
  tablesState.data = [];
  tablesState.isLoading = false;
});

describe('QrCodes list: true-empty state', () => {
  it('shows the illustration, title and a create-qr-code call-to-action when there are no tables', () => {
    render(<QrCodes />);

    expect(screen.getByText('qrCodes.emptyNoTablesTitle')).toBeInTheDocument();
    expect(screen.getByText('qrCodes.emptyNoQrCodesYet')).toBeInTheDocument();
  });

  it('navigates to the create screen when the true-empty call-to-action is clicked', () => {
    render(<QrCodes />);

    fireEvent.click(screen.getByRole('button', { name: 'qrCodes.createQrCode' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/qr-codes/create' });
  });
});

describe('QrCodes list: search-empty state', () => {
  it('shows the search-empty title and a clear-search call-to-action when a search matches nothing', () => {
    tablesState.data = [
      { id: 1, name: 'Table 1', capacity: 4, status: 'active', qrCodeUrl: 'https://x/1' },
    ];
    render(<QrCodes />);

    fireEvent.change(screen.getByPlaceholderText(/Search/i), {
      target: { value: 'no such table' },
    });

    expect(screen.getByText('qrCodes.emptyTryAnotherSearch')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearSearch' })).toBeInTheDocument();
    expect(screen.queryByText('qrCodes.emptyNoQrCodesYet')).not.toBeInTheDocument();
  });

  it('clears the search query when the search-empty call-to-action is clicked', () => {
    tablesState.data = [
      { id: 1, name: 'Table 1', capacity: 4, status: 'active', qrCodeUrl: 'https://x/1' },
    ];
    render(<QrCodes />);

    const searchInput = screen.getByPlaceholderText(/Search/i);
    fireEvent.change(searchInput, { target: { value: 'no such table' } });
    expect(screen.getByText('qrCodes.emptyTryAnotherSearch')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearSearch' }));

    expect(screen.queryByText('qrCodes.emptyTryAnotherSearch')).not.toBeInTheDocument();
    expect(screen.getByText('Table 1')).toBeInTheDocument();
  });

  it('does not show the empty state when tables match', () => {
    tablesState.data = [
      { id: 1, name: 'Table 1', capacity: 4, status: 'active', qrCodeUrl: 'https://x/1' },
    ];
    render(<QrCodes />);

    expect(screen.queryByText('qrCodes.emptyNoTablesTitle')).not.toBeInTheDocument();
  });
});
