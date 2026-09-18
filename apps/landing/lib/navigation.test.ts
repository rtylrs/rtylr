import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { navigateExternal, openInNewTab } from './navigation';

describe('navigateExternal', () => {
  let mockWindow: { location: { href: string }; open: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    mockWindow = {
      location: { href: 'https://rtylr.com' },
      open: vi.fn(),
    };
    vi.stubGlobal('window', mockWindow);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sets window.location.href to the given URL', () => {
    navigateExternal('https://rtylr.com');
    expect(mockWindow.location.href).toBe('https://rtylr.com');
  });

  it('sets window.location.href to any provided URL', () => {
    navigateExternal('https://rtylr.com/pricing');
    expect(mockWindow.location.href).toBe('https://rtylr.com/pricing');
  });
});

describe('openInNewTab', () => {
  let mockWindow: { location: { href: string }; open: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    mockWindow = {
      location: { href: 'https://rtylr.com' },
      open: vi.fn(),
    };
    vi.stubGlobal('window', mockWindow);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('calls window.open with _blank target and noopener noreferrer', () => {
    openInNewTab('https://rtylr.com');
    expect(mockWindow.open).toHaveBeenCalledWith('https://rtylr.com', '_blank', 'noopener,noreferrer');
  });

  it('passes the given URL to window.open', () => {
    openInNewTab('https://app.rtylr.com/login');
    expect(mockWindow.open).toHaveBeenCalledWith(
      'https://app.rtylr.com/login',
      '_blank',
      'noopener,noreferrer'
    );
  });
});
