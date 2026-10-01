import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import LocationSelect from './LocationSelect';

const accessClient = vi.hoisted(() => ({
  getMyLocations: vi.fn(),
}));

const locationClient = vi.hoisted(() => ({
  listLocations: vi.fn(),
  createLocation: vi.fn(),
}));

const terminalClient = vi.hoisted(() => ({
  listTerminals: vi.fn(),
  createTerminal: vi.fn(),
}));

const mockToast = vi.hoisted(() => ({
  error: vi.fn(),
}));

const mockUser = vi.hoisted(() => ({
  effectiveAccess: { isOwner: true, isAdmin: false } as { isOwner: boolean; isAdmin: boolean } | null,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (s: { effectiveAccess: typeof mockUser.effectiveAccess }) => unknown) =>
    selector(mockUser),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: vi.fn((client: { name?: string }) => {
    if (client.name === 'AccessServiceClient') return accessClient;
    if (client.name === 'LocationServiceClient') return locationClient;
    if (client.name === 'TerminalServiceClient') return terminalClient;
    return {};
  }),
}));

vi.mock('sonner', () => ({
  toast: mockToast,
}));

describe('LocationSelect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUser.effectiveAccess = { isOwner: true, isAdmin: false };
  });

  afterEach(() => {
    cleanup();
  });

  it('shows a load error when assigned and owned locations both fail', async () => {
    accessClient.getMyLocations.mockRejectedValue(new Error('assigned failed'));
    locationClient.listLocations.mockRejectedValue(new Error('branches failed'));

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(
      await screen.findByText('Unable to load locations. Check your connection and try again.'),
    ).toBeTruthy();
  });

  it('shows owned locations when the assigned list is empty', async () => {
    accessClient.getMyLocations.mockResolvedValue({ locations: [] });
    locationClient.listLocations.mockResolvedValue({
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
    accessClient.getMyLocations.mockResolvedValue({
      locations: [{ id: 1, name: 'Main' }],
    });
    locationClient.listLocations.mockResolvedValue({
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
    accessClient.getMyLocations.mockResolvedValue({ locations: [] });
    locationClient.listLocations.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(
      await screen.findByPlaceholderText('Location name (e.g. Downtown Cafe)'),
    ).toBeTruthy();
    expect(screen.getByText("Welcome! Let's set up your first location")).toBeTruthy();
  });

  it('creates the first location and advances to terminal registration', async () => {
    accessClient.getMyLocations.mockResolvedValue({ locations: [] });
    locationClient.listLocations.mockResolvedValue({ locations: [] });
    locationClient.createLocation.mockResolvedValue({ locationId: 7, message: 'ok' });
    terminalClient.listTerminals.mockResolvedValue({ terminals: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.change(await screen.findByPlaceholderText('Location name (e.g. Downtown Cafe)'), {
      target: { value: 'Downtown Cafe' },
    });
    fireEvent.click(screen.getByText('Create location'));

    expect(await screen.findByText('Register this terminal')).toBeTruthy();
    expect(locationClient.createLocation).toHaveBeenCalledWith({
      name: 'Downtown Cafe',
      address: '',
      phone: '',
      isActive: true,
    });
    // Terminal name pre-filled for first-run flow
    const terminalInput = screen.getByPlaceholderText(
      'Terminal name (e.g. Register 1)',
    ) as HTMLInputElement;
    expect(terminalInput.value).toBe('Register 1');
  });

  it('reports a failed location create without leaving the form', async () => {
    accessClient.getMyLocations.mockResolvedValue({ locations: [] });
    locationClient.listLocations.mockResolvedValue({ locations: [] });
    locationClient.createLocation.mockRejectedValue(new Error('Failed to create location'));

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
    accessClient.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    locationClient.listLocations.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.click(await screen.findByText('Add location'));

    expect(screen.getByText('Add a new location')).toBeTruthy();
    fireEvent.click(screen.getByText('← Back to locations'));
    expect(screen.getByText('Main')).toBeTruthy();
  });

  it('does not offer location creation to a store manager', async () => {
    mockUser.effectiveAccess = { isOwner: false, isAdmin: false };
    accessClient.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    locationClient.listLocations.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('Main')).toBeTruthy();
    expect(screen.queryByText('Add location')).toBeNull();
  });

  it('tells an unassigned store manager to ask their owner instead of showing a create form', async () => {
    mockUser.effectiveAccess = { isOwner: false, isAdmin: false };
    accessClient.getMyLocations.mockResolvedValue({ locations: [] });
    locationClient.listLocations.mockResolvedValue({ locations: [] });

    render(<LocationSelect onComplete={vi.fn()} />);

    expect(await screen.findByText('No locations assigned yet')).toBeTruthy();
    expect(screen.queryByPlaceholderText('Location name (e.g. Downtown Cafe)')).toBeNull();
  });

  it('shows a terminal load error when terminals cannot be loaded for a location', async () => {
    accessClient.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    locationClient.listLocations.mockResolvedValue({ locations: [] });
    terminalClient.listTerminals.mockRejectedValue(new Error('terminals failed'));

    render(<LocationSelect onComplete={vi.fn()} />);

    fireEvent.click(await screen.findByText('Main'));

    expect(
      await screen.findByText('Unable to load terminals. You can retry or register a new terminal.'),
    ).toBeTruthy();
  });

  it('reports invalid terminal create responses and does not complete setup', async () => {
    const onComplete = vi.fn();
    accessClient.getMyLocations.mockResolvedValue({
      locations: [
        { id: 1, name: 'Main' },
        { id: 2, name: 'Uptown' },
      ],
    });
    locationClient.listLocations.mockResolvedValue({ locations: [] });
    terminalClient.listTerminals.mockResolvedValue({ terminals: [] });
    terminalClient.createTerminal.mockResolvedValue({ terminalId: 'draft' });

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
