import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useClickOutside } from './useClickOutside';
import { createRef } from 'react';

describe('useClickOutside', () => {
  // Vitest 4 tightened Mock typing - generic now defaults to
  // `Procedure | Constructable`, which doesn't match the single
  // `(event: ...) => void` signature `useClickOutside` expects.
  // Pin the signature explicitly to keep this assignable.
  type OutsideHandler = (event: MouseEvent | TouchEvent) => void;
  let handler: ReturnType<typeof vi.fn<OutsideHandler>>;

  beforeEach(() => {
    handler = vi.fn<OutsideHandler>();
  });

  it('calls handler when clicking outside the element', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;

    renderHook(() => useClickOutside(ref, handler));

    const outside = new MouseEvent('mousedown', { bubbles: true });
    document.dispatchEvent(outside);

    expect(handler).toHaveBeenCalledTimes(1);
    document.body.removeChild(div);
  });

  it('does not call handler when clicking inside the element', () => {
    const div = document.createElement('div');
    const child = document.createElement('span');
    div.appendChild(child);
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;

    renderHook(() => useClickOutside(ref, handler));

    const inside = new MouseEvent('mousedown', { bubbles: true });
    child.dispatchEvent(inside);

    expect(handler).not.toHaveBeenCalled();
    document.body.removeChild(div);
  });

  it('does not call handler when clicking the element itself', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;

    renderHook(() => useClickOutside(ref, handler));

    const self = new MouseEvent('mousedown', { bubbles: true });
    div.dispatchEvent(self);

    expect(handler).not.toHaveBeenCalled();
    document.body.removeChild(div);
  });

  it('handles touchstart events', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;

    renderHook(() => useClickOutside(ref, handler));

    const touch = new TouchEvent('touchstart', { bubbles: true });
    document.dispatchEvent(touch);

    expect(handler).toHaveBeenCalledTimes(1);
    document.body.removeChild(div);
  });

  it('does not call handler when ref.current is null', () => {
    const ref = { current: null } as unknown as React.RefObject<HTMLDivElement>;

    renderHook(() => useClickOutside(ref, handler));

    const outside = new MouseEvent('mousedown', { bubbles: true });
    document.dispatchEvent(outside);

    expect(handler).not.toHaveBeenCalled();
  });

  it('cleans up event listeners on unmount', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;
    const removeSpy = vi.spyOn(document, 'removeEventListener');

    const { unmount } = renderHook(() => useClickOutside(ref, handler));
    unmount();

    expect(removeSpy).toHaveBeenCalledWith('mousedown', expect.any(Function));
    expect(removeSpy).toHaveBeenCalledWith('touchstart', expect.any(Function));
    removeSpy.mockRestore();
    document.body.removeChild(div);
  });

  it('does not fire after unmount', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;

    const { unmount } = renderHook(() => useClickOutside(ref, handler));
    unmount();

    document.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
    expect(handler).not.toHaveBeenCalled();
    document.body.removeChild(div);
  });

  // The assertion this hook used to carry, `event.target as Node`, was a lie:
  // EventTarget is not necessarily a Node. Dispatching on a non-Node target
  // reached `contains()` with something it cannot accept. A target that is not
  // a Node is not inside the element, so it counts as an outside interaction.
  it('treats a non-Node event target as outside rather than passing it to contains', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;
    const containsSpy = vi.spyOn(div, 'contains');

    renderHook(() => useClickOutside(ref, handler));

    // window is an EventTarget and not a Node.
    const event = new MouseEvent('mousedown', { bubbles: true });
    Object.defineProperty(event, 'target', { value: window, configurable: true });
    document.dispatchEvent(event);

    expect(containsSpy).not.toHaveBeenCalled();
    expect(handler).toHaveBeenCalledTimes(1);

    containsSpy.mockRestore();
    document.body.removeChild(div);
  });

  it('updates handler reference without re-registering listeners', () => {
    const div = document.createElement('div');
    document.body.appendChild(div);
    const ref = { current: div } as React.RefObject<HTMLDivElement>;
    const handler2 = vi.fn<OutsideHandler>();

    const { rerender } = renderHook(({ h }) => useClickOutside(ref, h), {
      initialProps: { h: handler },
    });

    rerender({ h: handler2 });

    document.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
    expect(handler).not.toHaveBeenCalled();
    expect(handler2).toHaveBeenCalledTimes(1);
    document.body.removeChild(div);
  });
});
