import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useEmployeeSessionStore } from './employeeSessionStore';

// Mock the api module
vi.mock('../services/api/api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
  },
}));

import api from '../services/api/api';

const { getState, setState } = useEmployeeSessionStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({
    employee: null,
    isActive: false,
    loading: false,
    error: null,
  });
  vi.clearAllMocks();
});

describe('employeeSessionStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with no employee and inactive', () => {
    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
  });

  // ── authenticatePin ────────────────────────────────────────────

  it('authenticates with a valid pin', async () => {
    const mockEmployee = { id: '1', name: 'Alice', role: 'cashier' };
    vi.mocked(api.post).mockResolvedValueOnce({ data: { employee: mockEmployee } });

    await getState().authenticatePin('1234');

    expect(getState().employee).toEqual(mockEmployee);
    expect(getState().isActive).toBe(true);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
  });

  it('sets loading true during authentication', async () => {
    let loadingDuringCall = false;
    vi.mocked(api.post).mockImplementation(async () => {
      loadingDuringCall = getState().loading;
      return { data: { employee: { id: '1', name: 'Alice', role: 'cashier' } } };
    });

    await getState().authenticatePin('1234');
    expect(loadingDuringCall).toBe(true);
  });

  it('sets error on failed authentication', async () => {
    vi.mocked(api.post).mockRejectedValueOnce({
      response: { data: { message: 'Invalid PIN' } },
    });

    await expect(getState().authenticatePin('0000')).rejects.toBeTruthy();

    expect(getState().error).toBe('Invalid PIN');
    expect(getState().loading).toBe(false);
    expect(getState().isActive).toBe(false);
    expect(getState().employee).toBeNull();
  });

  it('uses fallback error message when server provides none', async () => {
    vi.mocked(api.post).mockRejectedValueOnce(new Error('Network error'));

    await expect(getState().authenticatePin('0000')).rejects.toBeTruthy();

    expect(getState().error).toBe('Incorrect PIN');
  });

  it('clears previous error before new auth attempt', async () => {
    setState({ error: 'old error' });
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { employee: { id: '1', name: 'Bob', role: 'manager' } },
    });

    await getState().authenticatePin('1111');
    expect(getState().error).toBeNull();
  });

  it('rejects blank PINs without calling the API', async () => {
    await expect(getState().authenticatePin('   ')).rejects.toThrow('Enter your employee PIN');

    expect(api.post).not.toHaveBeenCalled();
    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBe('Enter your employee PIN');
  });

  it('rejects malformed employee auth responses without activating a session', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({
      data: { employee: { id: 'employee-uuid', name: 'Alice', role: 'cashier', pin: '1234' } },
    });

    await expect(getState().authenticatePin('1234')).rejects.toThrow(
      'Employee login returned invalid employee data',
    );

    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBe('Employee login returned invalid employee data');
  });

  it('normalizes employee auth response and strips sensitive fields', async () => {
    vi.mocked(api.post).mockResolvedValueOnce({
      data: {
        employee: {
          id: 9.7,
          name: ' Alice ',
          role: ' cashier ',
          pin: '1234',
          pin_hash: 'secret',
        },
      },
    });

    await getState().authenticatePin(' 1234 ');

    expect(api.post).toHaveBeenCalledWith('/employees/pin', { pin: '1234' });
    expect(getState().employee).toEqual({ id: '9', name: 'Alice', role: 'cashier' });
  });

  // ── endSession ─────────────────────────────────────────────────

  it('ends the session and clears employee', () => {
    setState({
      employee: { id: '1', name: 'Alice', role: 'cashier' },
      isActive: true,
      error: 'some error',
    });

    act(() => getState().endSession());

    expect(getState().employee).toBeNull();
    expect(getState().isActive).toBe(false);
    expect(getState().error).toBeNull();
  });

  // ── clearError ─────────────────────────────────────────────────

  it('clears the error', () => {
    setState({ error: 'Something went wrong' });

    act(() => getState().clearError());

    expect(getState().error).toBeNull();
  });

  it('clearError is a no-op when error is already null', () => {
    act(() => getState().clearError());
    expect(getState().error).toBeNull();
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('endSession is safe to call when already inactive', () => {
    act(() => getState().endSession());
    expect(getState().isActive).toBe(false);
    expect(getState().employee).toBeNull();
  });

  it('multiple authentications replace the previous employee', async () => {
    const emp1 = { id: '1', name: 'Alice', role: 'cashier' };
    const emp2 = { id: '2', name: 'Bob', role: 'manager' };

    vi.mocked(api.post).mockResolvedValueOnce({ data: { employee: emp1 } });
    await getState().authenticatePin('1111');
    expect(getState().employee).toEqual(emp1);

    vi.mocked(api.post).mockResolvedValueOnce({ data: { employee: emp2 } });
    await getState().authenticatePin('2222');
    expect(getState().employee).toEqual(emp2);
  });
});
