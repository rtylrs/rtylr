import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import LocationSelect from './LocationSelect';

const mockPosService = vi.hoisted(() => ({
  getMyLocations: vi.fn(),
  getBranches: vi.fn(),
  getTerminals: vi.fn(),
  createTerminal: vi.fn(),
  createBranch: vi.fn(),
}));

const mockToast = vi.hoisted(() => ({
  error: vi.fn(),
}));

const mockUser = vi.hoisted(() => ({ role: 'owner' as string | null }));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (s: { role: string | null }) => unknown) =>
    selector({ role: mockUser.role }),
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
    createBranch: (...args: unknown[]) => mockPosService.createBranch(...args),
  },
}));

describe('LocationSelect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUser.role = 'owner';
  });

  afterEach(() => {
    cleanup();
  });

  it('shows a load error when assigned and owned locations both fail', async () => {
    mockPosService.getMyLocations.mockRejectedValue(new Error('assigned failed'));
    mockPosService.getBranches.mockRejectedValue(new Error('branches failed'));

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(
      await screen.findByText('Unable to load locations. Check your connection and try again.'),
    ).toBeTruthy();
  });

  it('shows owned locations when the assigned list is empty', async () => {
    mockPosService.getMyLocations.mockResolvedValue({ locations: [] });
    mockPosService.getBranches.mockResolvedValue({
      locations: [
        { id: 3, name: 'HQ' },
        { id: 4, name: 'Annex' },
      ],
    });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('HQ')).toBeTruthy();
    expect(screen.getByText('Annex')).toBeTruthy();
  });

  it('merges assigned and owned locations without duplicates', async () => {
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [{ id: 1, name: 'Main' }],
    });
    mockPosService.getBranches.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('Uptown')).toBeTruthy();
    expect(screen.getAllByText('Main')).toHaveLength(1);
  });

  it('shows the first-location setup form when no locations exist anywhere', async () => {
    mockPosService.getMyLocations.mockResolvedValue({ locations: [] });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(
      await screen.findByPlaceholderText('Location name (e.g. Downtown Cafe)'),
    ).toBeTruthy();
    expect(screen.getByText("Welcome! Let's set up your first location")).toBeTruthy();
  });

  it('creates the first location and advances to terminal registration', async () => {
    mockPosService.getMyLocations.mockResolvedValue({ locations: [] });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });
    mockPosService.createBranch.mockResolvedValue({ location_id: 7, message: 'ok' });
    mockPosService.getTerminals.mockResolvedValue({ terminals: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.change(await screen.findByPlaceholderText('Location name (e.g. Downtown Cafe)'), {
      target: { value: 'Downtown Cafe' },
    });
    fireEvent.click(screen.getByText('Create location'));

    expect(await screen.findByText('Register this terminal')).toBeTruthy();
    expect(mockPosService.createBranch).toHaveBeenCalledWith({
      name: 'Downtown Cafe',
      address: '',
      phone: '',
      is_active: true,
    });
    // Terminal name pre-filled for first-run flow
    const terminalInput = screen.getByPlaceholderText(
      'Terminal name (e.g. Register 1)',
    ) as HTMLInputElement;
    expect(terminalInput.value).toBe('Register 1');
  });

  it('reports a failed location create without leaving the form', async () => {
    mockPosService.getMyLocations.mockResolvedValue({ locations: [] });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });
    mockPosService.createBranch.mockRejectedValue(new Error('Failed to create location'));

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.change(await screen.findByPlaceholderText('Location name (e.g. Downtown Cafe)'), {
      target: { value: 'Downtown Cafe' },
    });
    fireEvent.click(screen.getByText('Create location'));

    await waitFor(() =>
      expect(mockToast.error).toHaveBeenCalledWith('Failed to create location'),
    );
    expect(screen.getByPlaceholderText('Location name (e.g. Downtown Cafe)')).toBeTruthy();
  });

  it('offers an Add location button alongside existing locations', async () => {
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.click(await screen.findByText('Add location'));

    expect(screen.getByText('Add a new location')).toBeTruthy();
    fireEvent.click(screen.getByText('← Back to locations'));
    expect(screen.getByText('Main')).toBeTruthy();
  });

  it('does not offer location creation to a store manager', async () => {
    mockUser.role = 'store_manager';
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('Main')).toBeTruthy();
    expect(screen.queryByText('Add location')).toBeNull();
  });

  it('tells an unassigned store manager to ask their owner instead of showing a create form', async () => {
    mockUser.role = 'store_manager';
    mockPosService.getMyLocations.mockResolvedValue({ locations: [] });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('No locations assigned yet')).toBeTruthy();
    expect(screen.queryByPlaceholderText('Location name (e.g. Downtown Cafe)')).toBeNull();
  });

  it('shows a terminal load error when terminals cannot be loaded for a location', async () => {
    mockPosService.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    mockPosService.getBranches.mockResolvedValue({ locations: [] });
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
    mockPosService.getBranches.mockResolvedValue({ locations: [] });
    mockPosService.getTerminals.mockResolvedValue({ terminals: [] });
    mockPosService.createTerminal.mockResolvedValue({ id: 'draft' });

    render(<LocationSelect onComplete={onComplete} />);

    fireEvent.click(await screen.findByText('Main'));
    const input = await screen.findByPlaceholderText('Terminal name (e.g. Register 1)');
    fireEvent.change(input, { target: { value: 'Register 1' } });
    fireEvent.click(screen.getByText('Add'));

    await waitFor(() =>
      expect(mockToast.error).toHaveBeenCalledWith(
        'Terminal registration did not return a valid terminal ID.',
      ),
    );
    expect(onComplete).not.toHaveBeenCalled();
  });
});
