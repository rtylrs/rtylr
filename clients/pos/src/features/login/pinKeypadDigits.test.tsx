/**
 * The PIN keypad is the one input an Arabic-locale cashier cannot avoid, and
 * both its entry paths used to drop non-ASCII digits: the physical keyboard
 * compared the raw key against '0'-'9', and the on-screen dots were fed by
 * whatever survived that check.
 *
 * The length policy here is the server's: 6-12 digits to set a PIN, and a
 * legacy 4-digit PIN still verifies, so neither step may assume one length.
 */
import React from 'react';
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import PinLogin from './index';
import { PIN_MAX_LENGTH } from './pinPolicy';

const authenticatePin = vi.hoisted(() => vi.fn(async () => undefined));
const setPIN = vi.hoisted(() => vi.fn(async (_input: { userId: string; pin: string }) => undefined));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (client: { name?: string }) => ({
    getLocationRoster: vi.fn(async () => ({
      roster: [
        { userId: 7, name: 'Amira', hasPin: true, isManager: false, posRole: 'cashier' },
        { userId: 8, name: 'Yusuf', hasPin: false, isManager: false, posRole: 'server' },
      ],
    })),
    setPIN,
  }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../../shared/stores/employeeSessionStore', () => ({
  useEmployeeSessionStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({
      authenticatePin,
      loading: false,
      error: null,
      clearError: vi.fn(),
    }),
}));

vi.mock('../../shared/stores/terminalContextStore', () => ({
  useTerminalContextStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ locationId: 3 }),
}));

vi.mock('../../shared/components/RtylrWordmark', () => ({ RtylrWordmark: () => <span>rtylr</span> }));

async function pickMember(name: string, heading: string) {
  render(<PinLogin />);
  fireEvent.click(await screen.findByText(name));
  await screen.findByText(heading);
}

function pressKeys(keys: string[]) {
  act(() => {
    for (const key of keys) {
      window.dispatchEvent(new KeyboardEvent('keydown', { key }));
    }
  });
}

const signIn = () => fireEvent.click(screen.getByText('Sign in'));
const button = (label: string) => screen.getByText(label) as HTMLButtonElement;

describe('PinLogin keypad', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authenticatePin.mockImplementation(async () => undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it('accepts Arabic-Indic digits from the physical keyboard', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['١', '٢', '٣', '٤']);
    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '1234'));
  });

  it('accepts Eastern Arabic-Indic digits from the physical keyboard', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['۱', '۲', '۳', '۴']);
    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '1234'));
  });

  it('still accepts ASCII digits', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['1', '2', '3', '4']);
    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '1234'));
  });

  it('ignores non-digit keys', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['a', 'Enter', ' ']);

    expect(authenticatePin).not.toHaveBeenCalled();
  });

  it('backspace still deletes after an Arabic-Indic digit', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['١', '٢']);
    pressKeys(['Backspace']);
    pressKeys(['٣', '٤', '٥']);

    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '1345'));
  });
});

describe('PinLogin PIN length', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authenticatePin.mockImplementation(async () => undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it('verifies a 6-digit PIN, the new server minimum', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['1', '2', '3', '4', '5', '6']);
    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '123456'));
  });

  it('still lets an operator with a legacy 4-digit PIN clock in', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys(['1', '2', '3']);
    expect(button('Sign in').disabled).toBe(true);

    pressKeys(['4']);
    signIn();

    await waitFor(() => expect(authenticatePin).toHaveBeenCalledWith(3, 7, '1234'));
  });

  it('renders one input per digit the server can hold and ignores the rest', async () => {
    await pickMember('Amira', 'Enter your PIN');

    pressKeys([...'1234567890123'.split('')]);

    expect(screen.getAllByText('•')).toHaveLength(PIN_MAX_LENGTH);
  });

  it('creates a 6-digit PIN for a member who has none', async () => {
    await pickMember('Yusuf', 'Create your PIN');

    pressKeys(['1', '2', '3', '4', '5', '6']);
    await screen.findByText('Confirm your PIN');
    pressKeys(['1', '2', '3', '4', '5', '6']);
    fireEvent.click(screen.getByText('Set PIN'));

    await waitFor(() => expect(setPIN).toHaveBeenCalledTimes(1));
    expect(setPIN.mock.calls[0][0]).toMatchObject({ userId: '8', pin: '123456' });
  });

  it('creates a 6-digit PIN typed as six Arabic-Indic digits', async () => {
    await pickMember('Yusuf', 'Create your PIN');

    pressKeys(['٤', '٥', '٦', '٧', '٨', '٩']);
    await screen.findByText('Confirm your PIN');
    pressKeys(['٤', '٥', '٦', '٧', '٨', '٩']);
    fireEvent.click(screen.getByText('Set PIN'));

    await waitFor(() => expect(setPIN).toHaveBeenCalledTimes(1));
    expect(setPIN.mock.calls[0][0]).toMatchObject({ userId: '8', pin: '456789' });
  });

  it('does not offer a 4-digit PIN to create, which the server now rejects', async () => {
    await pickMember('Yusuf', 'Create your PIN');

    pressKeys(['1', '2', '3', '4']);
    expect(button('Set PIN').disabled).toBe(true);

    pressKeys(['5', '6']);
    await screen.findByText('Confirm your PIN');
    pressKeys(['1', '2', '3', '4', '5', '6']);
    fireEvent.click(screen.getByText('Set PIN'));

    await waitFor(() => expect(setPIN).toHaveBeenCalledTimes(1));
    expect(setPIN.mock.calls[0][0]).toMatchObject({ pin: '123456' });
  });
});
