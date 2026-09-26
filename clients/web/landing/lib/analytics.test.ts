import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  EV_CTA_CLICK,
  EV_FORM_START,
  EV_GENERATE_LEAD,
  EV_SIGNUP_CLICK,
  track,
} from './analytics';

describe('track', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('does nothing without a gtag function', () => {
    expect(() => track(EV_GENERATE_LEAD, { form: 'demo' })).not.toThrow();
  });

  it('forwards the event name and params to gtag', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag });

    track(EV_GENERATE_LEAD, { form: 'contact' });

    expect(gtag).toHaveBeenCalledWith('event', 'generate_lead', { form: 'contact' });
  });

  it('passes an empty params object when none are given', () => {
    const gtag = vi.fn();
    vi.stubGlobal('window', { gtag });

    track(EV_CTA_CLICK);

    expect(gtag).toHaveBeenCalledWith('event', 'cta_click', {});
  });

  it('exports the exact event names the GA4 key events use', () => {
    expect(EV_GENERATE_LEAD).toBe('generate_lead');
    expect(EV_FORM_START).toBe('form_start');
    expect(EV_CTA_CLICK).toBe('cta_click');
    expect(EV_SIGNUP_CLICK).toBe('sign_up_click');
  });
});
