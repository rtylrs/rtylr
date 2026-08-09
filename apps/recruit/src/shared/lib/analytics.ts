// Analytics event helpers for the public recruit pages.
// These wrap GA4 and Meta Pixel calls so component code never touches
// the window.gtag or window.fbq globals directly. Each function is a
// no-op when the relevant tracker is not loaded.

declare global {
  interface Window {
    gtag?: (...args: unknown[]) => void;
    fbq?: (...args: unknown[]) => void;
  }
}

/**
 * trackPageView fires a GA4 page_view event and a Meta Pixel PageView.
 * Called on each route change in the public careers SPA.
 */
export function trackPageView(path: string, title: string): void {
  if (typeof window === 'undefined') return;
  if (window.gtag) {
    window.gtag('event', 'page_view', {
      page_path: path,
      page_title: title,
    });
  }
  if (window.fbq) {
    window.fbq('track', 'PageView');
  }
}

/**
 * trackJobView fires an event when a candidate views a job listing.
 */
export function trackJobView(careerTitle: string, careerSlug: string): void {
  if (typeof window === 'undefined') return;
  if (window.gtag) {
    window.gtag('event', 'view_item', {
      item_name: careerTitle,
      item_id: careerSlug,
    });
  }
  if (window.fbq) {
    window.fbq('track', 'ViewContent', { content_name: careerTitle });
  }
}

/**
 * trackApplicationStart fires when the candidate opens the apply form.
 */
export function trackApplicationStart(careerTitle: string): void {
  if (typeof window === 'undefined') return;
  if (window.gtag) {
    window.gtag('event', 'begin_checkout', { item_name: careerTitle });
  }
  if (window.fbq) {
    window.fbq('track', 'InitiateCheckout', { content_name: careerTitle });
  }
}

/**
 * trackApplicationSubmit fires when the apply form is successfully
 * submitted.
 */
export function trackApplicationSubmit(careerTitle: string): void {
  if (typeof window === 'undefined') return;
  if (window.gtag) {
    window.gtag('event', 'purchase', { item_name: careerTitle });
  }
  if (window.fbq) {
    window.fbq('track', 'Lead', { content_name: careerTitle });
  }
}

/**
 * isGaLoaded returns true when the GA4 global is present and callable.
 */
export function isGaLoaded(): boolean {
  return typeof window !== 'undefined' && typeof window.gtag === 'function';
}

/**
 * isPixelLoaded returns true when the Meta Pixel global is present and callable.
 */
export function isPixelLoaded(): boolean {
  return typeof window !== 'undefined' && typeof window.fbq === 'function';
}
