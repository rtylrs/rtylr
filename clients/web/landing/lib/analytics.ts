/**
 * GA4 event helper.
 *
 * Consent is not read here on purpose: `lib/consentMode.ts` defines `gtag`
 * before consent exists and Consent Mode itself drops the hit when analytics
 * consent is denied, so the gate stays in one place.
 */
declare global {
  interface Window {
    gtag?: (...args: unknown[]) => void;
  }
}

export const EV_GENERATE_LEAD = 'generate_lead';
export const EV_FORM_START = 'form_start';
export const EV_CTA_CLICK = 'cta_click';
export const EV_SIGNUP_CLICK = 'sign_up_click';
// Homepage easter egg: the visitor caught the hiding Caret mascot.
export const EV_MASCOT_CATCH = 'mascot_catch';

export function track(name: string, params: Record<string, string | number | boolean> = {}): void {
  if (typeof window === 'undefined' || typeof window.gtag !== 'function') return;
  window.gtag('event', name, params);
}
