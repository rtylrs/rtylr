import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { navigateExternal, openInNewTab } from './navigation';

describe('navigateExternal', () => {
  let originalHref: string;

  beforeEach(() => {
    originalHref = window.location.href;
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: '' },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: originalHref },
    });
  });

  it('sets window.location.href to the given URL', () => {
    navigateExternal('https://rtylr.com');
    expect(window.location.href).toBe('https://rtylr.com');
  });

  it('sets window.location.href to any provided URL', () => {
    navigateExternal('https://rtylr.com/pricing');
    expect(window.location.href).toBe('https://rtylr.com/pricing');
  });
});

describe('openInNewTab', () => {
  it('calls window.open with _blank target and noopener noreferrer', () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);
    openInNewTab('https://rtylr.com');
    expect(openSpy).toHaveBeenCalledWith('https://rtylr.com', '_blank', 'noopener,noreferrer');
    openSpy.mockRestore();
  });

  it('passes the given URL to window.open', () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);
    openInNewTab('https://app.rtylr.com/login');
    expect(openSpy).toHaveBeenCalledWith(
      'https://app.rtylr.com/login',
      '_blank',
      'noopener,noreferrer'
    );
    openSpy.mockRestore();
  });
});
