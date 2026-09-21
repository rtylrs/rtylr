import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const userState = vi.hoisted(() => ({
  user: { id: 7, onboarded: false } as { id?: number; onboarded?: boolean } | null,
}));

vi.mock('../stores/userStore', () => ({
  useUserStore: () => ({ user: userState.user }),
}));

const markOnboardedMock = vi.hoisted(() => vi.fn(() => Promise.resolve({})));

vi.mock('../lib/clients', () => ({
  onboardingClient: { markOnboarded: markOnboardedMock },
}));

const shepherdState = vi.hoisted(() => ({
  lastTour: null as {
    complete: () => void;
    cancel: () => void;
  } | null,
}));

vi.mock('shepherd.js', () => {
  class FakeTour {
    private handlers = new Map<string, () => void>();

    constructor() {
      shepherdState.lastTour = this as never;
    }

    addStep(): void {}

    on(event: string, handler: () => void): this {
      this.handlers.set(event, handler);
      return this;
    }

    off(event: string, handler?: () => void): this {
      if (!handler || this.handlers.get(event) === handler) {
        this.handlers.delete(event);
      }
      return this;
    }

    start(): void {}

    cancel(): void {
      this.handlers.get('cancel')?.();
    }

    complete(): void {
      this.handlers.get('complete')?.();
    }
  }

  return { default: { Tour: FakeTour } };
});

import { useOnboardingTour } from './useOnboardingTour';

const STORAGE_KEY = 'rtylr-tour-done:7';

describe('useOnboardingTour persistence', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.clear();
    vi.clearAllMocks();
    shepherdState.lastTour = null;
    userState.user = { id: 7, onboarded: false };
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('does not mark onboarding done when the layout unmounts mid-tour', () => {
    const { unmount } = renderHook(() => useOnboardingTour());

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(shepherdState.lastTour).not.toBeNull();

    unmount();

    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    expect(markOnboardedMock).not.toHaveBeenCalled();
  });

  it('still marks onboarding done when the user completes the tour', () => {
    const { unmount } = renderHook(() => useOnboardingTour());

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    act(() => {
      shepherdState.lastTour?.complete();
    });

    expect(localStorage.getItem(STORAGE_KEY)).toBe('1');
    expect(markOnboardedMock).toHaveBeenCalledTimes(1);

    // A later unmount must not fire a second completion call.
    unmount();
    expect(markOnboardedMock).toHaveBeenCalledTimes(1);
  });
});
