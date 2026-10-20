import { act, render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useLoadOnScroll } from './useLoadOnScroll';

type Callback = (entries: Array<{ isIntersecting: boolean }>) => void;

const observers: Array<{ callback: Callback; disconnect: ReturnType<typeof vi.fn> }> = [];

beforeEach(() => {
  observers.length = 0;
  vi.stubGlobal(
    'IntersectionObserver',
    class {
      disconnect = vi.fn();
      constructor(callback: Callback) {
        observers.push({ callback, disconnect: this.disconnect });
      }
      observe() {}
    }
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const List: React.FC<{ enabled: boolean; onLoad: () => void }> = ({ enabled, onLoad }) => {
  const ref = useLoadOnScroll(enabled, onLoad);
  return <div ref={ref} />;
};

describe('useLoadOnScroll', () => {
  it('loads when the sentinel comes into view', () => {
    const onLoad = vi.fn();
    render(<List enabled onLoad={onLoad} />);
    expect(observers).toHaveLength(1);

    act(() => observers[0].callback([{ isIntersecting: false }]));
    expect(onLoad).not.toHaveBeenCalled();
    act(() => observers[0].callback([{ isIntersecting: true }]));
    expect(onLoad).toHaveBeenCalledTimes(1);
  });

  it('stops watching while disabled and re-checks when enabled again', () => {
    const onLoad = vi.fn();
    const { rerender } = render(<List enabled onLoad={onLoad} />);
    rerender(<List enabled={false} onLoad={onLoad} />);
    expect(observers[0].disconnect).toHaveBeenCalled();
    expect(observers).toHaveLength(1);

    rerender(<List enabled onLoad={onLoad} />);
    expect(observers).toHaveLength(2);
    act(() => observers[1].callback([{ isIntersecting: true }]));
    expect(onLoad).toHaveBeenCalledTimes(1);
  });

  it('does nothing when there are no more pages', () => {
    render(<List enabled={false} onLoad={vi.fn()} />);
    expect(observers).toHaveLength(0);
  });
});
