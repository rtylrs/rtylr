import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useCaretInstall } from './useCaretInstall';

afterEach(() => {
  document.head.innerHTML = '';
});

describe('useCaretInstall', () => {
  it('attaches the Caret manifest only while mounted', () => {
    const { unmount } = renderHook(() => useCaretInstall());
    expect(document.querySelector('link[rel="manifest"]')?.getAttribute('href')).toBe(
      '/caret/manifest.webmanifest'
    );
    unmount();
    expect(document.querySelector('link[rel="manifest"]')).toBeNull();
  });

  it('swaps the touch icon to Caret and restores the portal one', () => {
    const icon = document.createElement('link');
    icon.rel = 'apple-touch-icon';
    icon.href = '/favicon/apple-touch-icon.png';
    document.head.appendChild(icon);

    const { unmount } = renderHook(() => useCaretInstall());
    expect(icon.getAttribute('href')).toBe('/caret/apple-touch-icon.png');
    unmount();
    expect(icon.getAttribute('href')).toBe('/favicon/apple-touch-icon.png');
  });

  it('offers install once the browser allows it', async () => {
    const { result } = renderHook(() => useCaretInstall());
    expect(result.current.canInstall).toBe(false);

    const event = Object.assign(new Event('beforeinstallprompt', { cancelable: true }), {
      prompt: vi.fn().mockResolvedValue(undefined),
      userChoice: Promise.resolve({ outcome: 'accepted' as const }),
    });
    act(() => {
      window.dispatchEvent(event);
    });
    expect(result.current.canInstall).toBe(true);

    await act(async () => {
      await result.current.install();
    });
    expect(event.prompt).toHaveBeenCalled();
    expect(result.current.canInstall).toBe(false);
    expect(result.current.installed).toBe(true);
  });
});
