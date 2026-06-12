import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import LocationSelect from './LocationSelect';

const mockPosService = vi.hoisted(() => ({
  getMyLocations: vi.fn(),
  getBranches: vi.fn(),
  getTerminals: vi.fn(),
  createTerminal: vi.fn(),
}));

const mockToast = vi.hoisted(() => ({
  error: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('sonner', () => ({
  toast: mockToast,
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getMyLocations: (...args: unknown[]) => mockPosService.getMyLocations(...args),
    getBranches: (...args: unknown[]) => mockPosService.getBranches(...args),
    getTerminals: (...args: unknown[]) => mockPosService.getTerminals(...args),
    createTerminal: (...args: unknown[]) => mockPosService.createTerminal(...args),
  },
}));

describe('LocationSelect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('shows a load error when assigned and fallback locations both fail', async () => {
    mockPosService.getMyLocations.mockRejectedValue(new Error('assigned failed'));
    mockPosService.getBranches.mockRejectedValue(new Error('branches failed'));

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(
      await screen.findByText('Unable to load locations. Check your connection and try again.'),
    ).toBeTruthy();
  });

  it('shows a terminal load error when terminals cannot be loaded for a location', async () => {
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    mockPosService.getTerminals.mockRejectedValue(new Error('terminals failed'));

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.click(await screen.findByText('Main'));

    expect(
      await screen.findByText('Unable to load terminals. You can retry or register a new terminal.'),
    ).toBeTruthy();
  });

  it('reports invalid terminal create responses and does not complete setup', async () => {
    const onComplete = vi.fn();
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    mockPosService.getTerminals.mockResolvedValue({ terminals: [] });
    mockPosService.createTerminal.mockResolvedValue({ id: 'draft' });

    render(<LocationSelect onComplete={onComplete} />);

    fireEvent.click(await screen.findByText('Main'));
    fireEvent.change(await screen.findByPlaceholderText('Terminal name (e.g. Register 1)'), {
      target: { value: 'Register 1' },
    });
    fireEvent.click(screen.getByText('Add'));

    await waitFor(() =>
      expect(mockToast.error).toHaveBeenCalledWith(
        'Terminal registration did not return a valid terminal ID.',
      ),
    );
    expect(onComplete).not.toHaveBeenCalled();
  });
});
