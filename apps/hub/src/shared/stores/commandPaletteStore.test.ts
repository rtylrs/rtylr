import { describe, it, expect, beforeEach } from 'vitest';
import { useCommandPalette } from './commandPaletteStore';

const { getState } = useCommandPalette;

beforeEach(() => {
  useCommandPalette.setState({ open: false });
});

describe('commandPaletteStore', () => {
  it('starts closed', () => {
    expect(getState().open).toBe(false);
  });

  it('setOpen(true) opens palette', () => {
    getState().setOpen(true);
    expect(getState().open).toBe(true);
  });

  it('setOpen(false) closes palette', () => {
    getState().setOpen(true);
    getState().setOpen(false);
    expect(getState().open).toBe(false);
  });

  it('toggle opens when closed', () => {
    getState().toggle();
    expect(getState().open).toBe(true);
  });

  it('toggle closes when open', () => {
    getState().setOpen(true);
    getState().toggle();
    expect(getState().open).toBe(false);
  });

  it('toggle twice returns to original state', () => {
    getState().toggle();
    getState().toggle();
    expect(getState().open).toBe(false);
  });

  it('multiple setOpen calls work correctly', () => {
    getState().setOpen(true);
    getState().setOpen(true);
    expect(getState().open).toBe(true);

    getState().setOpen(false);
    expect(getState().open).toBe(false);
  });

  it('toggle after setOpen works correctly', () => {
    getState().setOpen(true);
    getState().toggle();
    expect(getState().open).toBe(false);
  });
});
