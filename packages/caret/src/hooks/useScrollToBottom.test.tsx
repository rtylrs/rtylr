import { fireEvent, render } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { useScrollToBottom } from './useScrollToBottom';

interface HarnessProps {
  scrollKey: string;
  contentVersion: number;
  scrollHeight: number;
}

function Harness({ scrollKey, contentVersion, scrollHeight }: HarnessProps): React.ReactNode {
  const ref = useScrollToBottom<HTMLDivElement>([contentVersion], scrollKey);
  return (
    <div
      ref={ref}
      data-testid="scroll-container"
      data-client-height="200"
      data-scroll-height={scrollHeight}
    />
  );
}

let originalClientHeight: PropertyDescriptor | undefined;
let originalScrollHeight: PropertyDescriptor | undefined;

beforeAll(() => {
  originalClientHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'clientHeight');
  originalScrollHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'scrollHeight');
  Object.defineProperties(HTMLElement.prototype, {
    clientHeight: {
      configurable: true,
      get() {
        return Number((this as HTMLElement).dataset.clientHeight ?? 0);
      },
    },
    scrollHeight: {
      configurable: true,
      get() {
        return Number((this as HTMLElement).dataset.scrollHeight ?? 0);
      },
    },
  });
});

afterAll(() => {
  if (originalClientHeight) {
    Object.defineProperty(HTMLElement.prototype, 'clientHeight', originalClientHeight);
  } else {
    delete (HTMLElement.prototype as unknown as Record<string, unknown>).clientHeight;
  }
  if (originalScrollHeight) {
    Object.defineProperty(HTMLElement.prototype, 'scrollHeight', originalScrollHeight);
  } else {
    delete (HTMLElement.prototype as unknown as Record<string, unknown>).scrollHeight;
  }
});

describe('useScrollToBottom', () => {
  it('restores an independent scroll position for each chat tab', () => {
    const { getByTestId, rerender } = render(
      <Harness scrollKey="scroll-test-first" contentVersion={1} scrollHeight={1000} />
    );
    const container = getByTestId('scroll-container');

    container.scrollTop = 180;
    fireEvent.scroll(container);

    rerender(<Harness scrollKey="scroll-test-second" contentVersion={1} scrollHeight={1600} />);
    expect(container.scrollTop).toBe(1600);

    container.scrollTop = 260;
    fireEvent.scroll(container);

    rerender(<Harness scrollKey="scroll-test-first" contentVersion={1} scrollHeight={1000} />);
    expect(container.scrollTop).toBe(180);
  });

  it('restores a scrolled-up position when remounting with the same key', () => {
    const scrollKey = 'scroll-test-remount';
    const { getByTestId, unmount } = render(
      <Harness scrollKey={scrollKey} contentVersion={1} scrollHeight={1000} />
    );
    const container = getByTestId('scroll-container');

    container.scrollTop = 180;
    fireEvent.scroll(container);

    unmount();

    const { getByTestId: getByTestIdAfterRemount } = render(
      <Harness scrollKey={scrollKey} contentVersion={1} scrollHeight={1000} />
    );
    expect(getByTestIdAfterRemount('scroll-container').scrollTop).toBe(180);
  });
});
