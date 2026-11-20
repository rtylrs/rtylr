import { describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ listSessions: vi.fn() }),
}));

import {
  openRegisterErrorMessage,
  openRegisterSession,
  parseOpeningFloat,
} from './openSessionWorkflow';

const noSession = async (): Promise<number> => {
  throw new Error('No open register session found.');
};

describe('openRegisterSession', () => {
  it('opens a session for the signed-in staff member and sets the drawer float', async () => {
    const openSession = vi.fn(async () => ({ sessionId: '42' }));
    const openDrawer = vi.fn();

    const result = await openRegisterSession({
      floatInput: '150.50',
      locationId: 3,
      staffId: '7',
      openSession,
      resolveSessionId: noSession,
      openDrawer,
    });

    expect(result).toEqual({ status: 'opened', openingFloat: 150.5 });
    expect(openSession).toHaveBeenCalledWith({
      locationId: '3',
      staffId: '7',
      openingFloat: 150.5,
    });
    expect(openDrawer).toHaveBeenCalledWith(150.5);
  });

  it('treats a blank float as an empty drawer', async () => {
    const openSession = vi.fn(async () => ({}));
    const openDrawer = vi.fn();

    const result = await openRegisterSession({
      floatInput: '  ',
      locationId: 3,
      staffId: '7',
      openSession,
      resolveSessionId: noSession,
      openDrawer,
    });

    expect(result).toEqual({ status: 'opened', openingFloat: 0 });
    expect(openDrawer).toHaveBeenCalledWith(0);
  });

  it('rejects a malformed float before calling the server', async () => {
    const openSession = vi.fn();
    const openDrawer = vi.fn();

    await expect(
      openRegisterSession({
        floatInput: '-5',
        locationId: 3,
        staffId: '7',
        openSession,
        resolveSessionId: noSession,
        openDrawer,
      })
    ).rejects.toThrow('Enter a valid opening float.');

    expect(openSession).not.toHaveBeenCalled();
    expect(openDrawer).not.toHaveBeenCalled();
  });

  it('refuses to open without a location or a signed-in staff member', async () => {
    const openSession = vi.fn();

    await expect(
      openRegisterSession({
        floatInput: '10',
        locationId: null,
        staffId: null,
        openSession,
        resolveSessionId: noSession,
        openDrawer: vi.fn(),
      })
    ).rejects.toThrow('Select a location and sign in before opening the register.');

    expect(openSession).not.toHaveBeenCalled();
  });

  it('continues without resetting the drawer when a session is already open', async () => {
    const openDrawer = vi.fn();

    const result = await openRegisterSession({
      floatInput: '20',
      locationId: 3,
      staffId: '7',
      openSession: vi.fn(async () => {
        throw Object.assign(new Error('Failed to open session'), { statusCode: 409 });
      }),
      resolveSessionId: async () => 42,
      openDrawer,
    });

    expect(result).toEqual({ status: 'already_open' });
    expect(openDrawer).not.toHaveBeenCalled();
  });

  it('surfaces the open failure when no session exists afterwards', async () => {
    const rejection = Object.assign(new Error('Forbidden'), { statusCode: 403 });
    const openDrawer = vi.fn();

    await expect(
      openRegisterSession({
        floatInput: '20',
        locationId: 3,
        staffId: '7',
        openSession: vi.fn(async () => {
          throw rejection;
        }),
        resolveSessionId: noSession,
        openDrawer,
      })
    ).rejects.toBe(rejection);

    expect(openDrawer).not.toHaveBeenCalled();
  });
});

describe('parseOpeningFloat', () => {
  it('rejects text that is not an amount', () => {
    expect(() => parseOpeningFloat('abc')).toThrow('Enter a valid opening float.');
  });
});

describe('openRegisterErrorMessage', () => {
  it('tells a cashier without permission to ask a manager', () => {
    expect(openRegisterErrorMessage({ statusCode: 403 })).toBe(
      'You do not have permission to open a register. Ask a manager.'
    );
  });

  it('points at the connection for transport failures', () => {
    expect(openRegisterErrorMessage(new TypeError('Failed to fetch'))).toBe(
      'Could not open the register. Check your connection and try again.'
    );
  });
});
