/**
 * The manager's "Reset PIN" field sanitised with /\D/g, which deletes
 * Arabic-Indic digits outright, so a manager on an Arabic keyboard could not
 * type a PIN at all. The value must normalise to ASCII instead.
 */
import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import StaffScreen from './index';
import { PIN_MAX_LENGTH } from '../login/pinPolicy';

const listPOSUsers = vi.hoisted(() => vi.fn());
const updatePOSUser = vi.hoisted(() => vi.fn(async () => undefined));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ listPOSUsers, updatePOSUser, deactivatePOSUser: vi.fn() }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../../shared/components/ManagerRequired', () => ({ default: () => <div /> }));

vi.mock('../../shared/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ role: 'manager' }),
}));

vi.mock('../../shared/stores/terminalContextStore', () => ({
  useTerminalContextStore: () => ({ locationId: 3 }),
}));

async function openPinField() {
  render(<StaffScreen />);
  fireEvent.click(await screen.findByText('Amira'));
  const input = await screen.findByPlaceholderText('Leave blank to keep current');
  return input as HTMLInputElement;
}

describe('Staff PIN reset field', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    listPOSUsers.mockResolvedValue({
      users: [
        {
          id: 7,
          name: 'Amira',
          role: 'cashier',
          isActive: true,
          maxDiscountPct: 10,
          canVoid: false,
          canRefund: false,
          canApplyDiscount: false,
          canOpenDrawer: false,
        },
      ],
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('normalises Arabic-Indic digits instead of deleting them', async () => {
    const input = await openPinField();

    fireEvent.change(input, { target: { value: '٤٥٦' } });

    expect(input.value).toBe('456');
  });

  it('normalises a full-length Arabic-Indic PIN', async () => {
    const input = await openPinField();

    fireEvent.change(input, { target: { value: '٤٥٦٧٨٩' } });

    expect(input.value).toBe('456789');
  });

  it('normalises Eastern Arabic-Indic digits', async () => {
    const input = await openPinField();

    fireEvent.change(input, { target: { value: '۱۲۳۴۵۶' } });

    expect(input.value).toBe('123456');
  });

  it('caps the field at the server maximum, which is also its maxLength', async () => {
    const input = await openPinField();

    expect(input.maxLength).toBe(PIN_MAX_LENGTH);

    fireEvent.change(input, { target: { value: '1a2b3c4d5e6f7g8h9' } });

    expect(input.value).toBe('123456789');
  });

  it('sends the normalised PIN to the server', async () => {
    const input = await openPinField();
    fireEvent.change(input, { target: { value: '٤٥٦٧٨٩' } });

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(updatePOSUser).toHaveBeenCalledTimes(1));
    expect(JSON.stringify(updatePOSUser.mock.calls)).toContain('456789');
  });
});
