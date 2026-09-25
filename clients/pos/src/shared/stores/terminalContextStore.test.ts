import { beforeEach, describe, expect, it } from 'vitest';
import { useTerminalContextStore } from './terminalContextStore';

const { getState, setState } = useTerminalContextStore;

beforeEach(() => {
  setState({ locationId: null, terminalId: null });
});

describe('terminalContextStore', () => {
  it('stores selected location and terminal together', () => {
    getState().setTerminalContext(3, 9);

    expect(getState().locationId).toBe(3);
    expect(getState().terminalId).toBe(9);
  });

  it('clears context when either id is invalid', () => {
    getState().setTerminalContext(0, 9);
    expect(getState().locationId).toBeNull();
    expect(getState().terminalId).toBeNull();

    getState().setTerminalContext(3, Number.NaN);
    expect(getState().locationId).toBeNull();
    expect(getState().terminalId).toBeNull();
  });

  it('rounds valid persisted ids to integers', () => {
    getState().setTerminalContext(3.7, 9.2);

    expect(getState().locationId).toBe(3);
    expect(getState().terminalId).toBe(9);
  });

  it('clears selected location and terminal together', () => {
    getState().setTerminalContext(3, 9);

    getState().clearTerminalContext();

    expect(getState().locationId).toBeNull();
    expect(getState().terminalId).toBeNull();
  });
});
