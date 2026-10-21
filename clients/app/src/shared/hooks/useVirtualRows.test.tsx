import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { VirtualList } from '../components/VirtualList';

const ROW = 40;
const VIEWPORT = 400;

const List: React.FC<{ count: number; scrolls?: boolean }> = ({ count, scrolls }) => (
  <div data-testid="scroller" style={scrolls ? { overflowY: 'auto' } : undefined}>
    <VirtualList
      count={count}
      estimateSize={ROW}
      renderRow={(index, rowProps) => (
        <div key={index} {...rowProps}>
          row {index}
        </div>
      )}
    />
  </div>
);

const rendered = () => screen.queryAllByText(/^row \d+$/).length;

/** jsdom lays nothing out: give the scroller a viewport and each row a height. */
function stubLayout() {
  const isScroller = (el: Element) => (el as HTMLElement).dataset?.testid === 'scroller';
  vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockImplementation(function (
    this: HTMLElement
  ) {
    return isScroller(this) ? VIEWPORT : 0;
  });
  // The virtualizer sizes its viewport from offsetHeight.
  vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockImplementation(function (
    this: HTMLElement
  ) {
    return isScroller(this) ? VIEWPORT : this.dataset.index !== undefined ? ROW : 0;
  });
  vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockImplementation(function (
    this: HTMLElement
  ) {
    return isScroller(this) ? 100_000 : 0;
  });
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (
    this: HTMLElement
  ) {
    const height = isScroller(this) ? VIEWPORT : this.dataset.index !== undefined ? ROW : 0;
    return {
      top: 0,
      left: 0,
      right: 300,
      bottom: height,
      width: 300,
      height,
      x: 0,
      y: 0,
    } as DOMRect;
  });
  vi.stubGlobal(
    'ResizeObserver',
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
  );
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe('useVirtualRows', () => {
  it('renders every row of a short list', () => {
    stubLayout();
    render(<List count={40} scrolls />);
    expect(rendered()).toBe(40);
  });

  it('renders every row when nothing scrolls, so tests and odd layouts lose nothing', () => {
    render(<List count={300} />);
    expect(rendered()).toBe(300);
    expect(document.querySelector('.virtual-spacer')).toBeNull();
  });

  it('renders only the rows near the viewport of a long list', () => {
    stubLayout();
    render(<List count={300} scrolls />);

    const first = rendered();
    expect(first).toBeGreaterThan(VIEWPORT / ROW);
    expect(first).toBeLessThan(30);
    expect(screen.getByText('row 0')).toBeTruthy();
    expect(screen.queryByText('row 299')).toBeNull();
    // The unrendered rows still hold their space below.
    const after = document.querySelector<HTMLElement>('.virtual-spacer');
    expect(parseFloat(after!.style.height)).toBeGreaterThan(250 * ROW);

    const scroller = screen.getByTestId('scroller');
    act(() => {
      scroller.scrollTop = 200 * ROW;
      fireEvent.scroll(scroller);
    });
    expect(screen.getByText('row 205')).toBeTruthy();
    expect(screen.queryByText('row 0')).toBeNull();
    expect(rendered()).toBeLessThan(30);
  });
});
