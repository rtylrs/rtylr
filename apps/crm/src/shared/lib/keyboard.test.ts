import { describe, expect, it, vi } from 'vitest';
import { activateOnKeyboard } from './keyboard';

function keyboardEvent(key: string, target: EventTarget, currentTarget: EventTarget) {
  return {
    key,
    target,
    currentTarget,
    preventDefault: vi.fn(),
  } as never;
}

describe('activateOnKeyboard', () => {
  it.each(['Enter', ' '])('activates on %j', (key) => {
    const onActivate = vi.fn();
    const owner = {};
    const event = keyboardEvent(key, owner, owner);

    activateOnKeyboard(event, onActivate);

    expect(onActivate).toHaveBeenCalledOnce();
    expect(event.preventDefault).toHaveBeenCalledOnce();
  });

  it('ignores other keys and nested controls', () => {
    const onActivate = vi.fn();
    const owner = {};

    activateOnKeyboard(keyboardEvent('Escape', owner, owner), onActivate);
    activateOnKeyboard(keyboardEvent('Enter', {}, owner), onActivate);

    expect(onActivate).not.toHaveBeenCalled();
  });
});
