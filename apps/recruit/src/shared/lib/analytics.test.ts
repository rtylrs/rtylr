import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  trackPageView,
  trackJobView,
  trackApplicationStart,
  trackApplicationSubmit,
  isGaLoaded,
  isPixelLoaded,
} from './analytics';

beforeEach(() => {
  vi.unstubAllGlobals();
});

describe('isGaLoaded', () => {
  it('returns false when gtag is not present', () => {
    vi.stubGlobal('window', { gtag: undefined });
    expect(isGaLoaded()).toBe(false);
    vi.unstubAllGlobals();
  });

  it('returns true when gtag is a function', () => {
    vi.stubGlobal('window', { gtag: vi.fn() });
    expect(isGaLoaded()).toBe(true);
    vi.unstubAllGlobals();
  });
});

describe('isPixelLoaded', () => {
  it('returns false when fbq is not present', () => {
    vi.stubGlobal('window', { fbq: undefined });
    expect(isPixelLoaded()).toBe(false);
    vi.unstubAllGlobals();
  });

  it('returns true when fbq is a function', () => {
    vi.stubGlobal('window', { fbq: vi.fn() });
    expect(isPixelLoaded()).toBe(true);
    vi.unstubAllGlobals();
  });
});

describe('trackPageView', () => {
  it('calls gtag with page_view event when available', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag, fbq: undefined });
    trackPageView('/careers', 'Careers');
    expect(gtag).toHaveBeenCalledWith(
      'event',
      'page_view',
      expect.objectContaining({ page_path: '/careers' })
    );
    vi.unstubAllGlobals();
  });

  it('calls fbq with PageView when available', () => {
    const fbq = vi.fn();
    vi.stubGlobal('window', { fbq, gtag: undefined });
    trackPageView('/careers', 'Careers');
    expect(fbq).toHaveBeenCalledWith('track', 'PageView');
    vi.unstubAllGlobals();
  });

  it('does not throw when neither tracker is available', () => {
    vi.stubGlobal('window', { gtag: undefined, fbq: undefined });
    expect(() => trackPageView('/careers', 'Careers')).not.toThrow();
    vi.unstubAllGlobals();
  });
});

describe('trackJobView', () => {
  it('calls gtag with view_item event', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag, fbq: undefined });
    trackJobView('Senior Engineer', 'senior-engineer');
    expect(gtag).toHaveBeenCalledWith(
      'event',
      'view_item',
      expect.objectContaining({ item_name: 'Senior Engineer' })
    );
    vi.unstubAllGlobals();
  });
});

describe('trackApplicationStart', () => {
  it('calls gtag with begin_checkout event', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag, fbq: undefined });
    trackApplicationStart('Frontend Developer');
    expect(gtag).toHaveBeenCalledWith(
      'event',
      'begin_checkout',
      expect.objectContaining({ item_name: 'Frontend Developer' })
    );
    vi.unstubAllGlobals();
  });
});

describe('trackApplicationSubmit', () => {
  it('calls gtag with purchase event', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag, fbq: undefined });
    trackApplicationSubmit('Backend Developer');
    expect(gtag).toHaveBeenCalledWith(
      'event',
      'purchase',
      expect.objectContaining({ item_name: 'Backend Developer' })
    );
    vi.unstubAllGlobals();
  });

  it('calls fbq with Lead event', () => {
    const fbq = vi.fn();
    vi.stubGlobal('window', { fbq, gtag: undefined });
    trackApplicationSubmit('Backend Developer');
    expect(fbq).toHaveBeenCalledWith(
      'track',
      'Lead',
      expect.objectContaining({ content_name: 'Backend Developer' })
    );
    vi.unstubAllGlobals();
  });
});
