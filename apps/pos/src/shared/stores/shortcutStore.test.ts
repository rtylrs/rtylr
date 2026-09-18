import { describe, it, expect, beforeEach } from 'vitest';
import { useShortcutStore } from './shortcutStore';

const { getState, setState } = useShortcutStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ commandOpen: false, helpOpen: false });
});

describe('shortcutStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with both panels closed', () => {
    expect(getState().commandOpen).toBe(false);
    expect(getState().helpOpen).toBe(false);
  });

  // ── toggleCommand ──────────────────────────────────────────────

  it('toggles command open', () => {
    act(() => getState().toggleCommand());
    expect(getState().commandOpen).toBe(true);
  });

  it('toggles command closed', () => {
    setState({ commandOpen: true });
    act(() => getState().toggleCommand());
    expect(getState().commandOpen).toBe(false);
  });

  it('toggles command back and forth', () => {
    act(() => getState().toggleCommand());
    expect(getState().commandOpen).toBe(true);
    act(() => getState().toggleCommand());
    expect(getState().commandOpen).toBe(false);
  });

  // ── setCommand ─────────────────────────────────────────────────

  it('sets command to true', () => {
    act(() => getState().setCommand(true));
    expect(getState().commandOpen).toBe(true);
  });

  it('sets command to false', () => {
    setState({ commandOpen: true });
    act(() => getState().setCommand(false));
    expect(getState().commandOpen).toBe(false);
  });

  // ── toggleHelp ─────────────────────────────────────────────────

  it('toggles help open', () => {
    act(() => getState().toggleHelp());
    expect(getState().helpOpen).toBe(true);
  });

  it('toggles help closed', () => {
    setState({ helpOpen: true });
    act(() => getState().toggleHelp());
    expect(getState().helpOpen).toBe(false);
  });

  // ── setHelp ────────────────────────────────────────────────────

  it('sets help to true', () => {
    act(() => getState().setHelp(true));
    expect(getState().helpOpen).toBe(true);
  });

  it('sets help to false', () => {
    setState({ helpOpen: true });
    act(() => getState().setHelp(false));
    expect(getState().helpOpen).toBe(false);
  });

  // ── Independence ───────────────────────────────────────────────

  it('toggling command does not affect help', () => {
    act(() => getState().toggleCommand());
    expect(getState().commandOpen).toBe(true);
    expect(getState().helpOpen).toBe(false);
  });

  it('toggling help does not affect command', () => {
    act(() => getState().toggleHelp());
    expect(getState().helpOpen).toBe(true);
    expect(getState().commandOpen).toBe(false);
  });

  it('both can be open simultaneously', () => {
    act(() => {
      getState().setCommand(true);
      getState().setHelp(true);
    });
    expect(getState().commandOpen).toBe(true);
    expect(getState().helpOpen).toBe(true);
  });
});
