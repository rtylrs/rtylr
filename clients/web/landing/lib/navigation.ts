/**
 * Navigation utilities for external and internal routing
 */

/**
 * Navigate to an external URL with full page reload
 * Use this for navigating to different apps (auth, menu, etc.)
 *
 * @param url - The full URL to navigate to
 */
export const navigateExternal = (url: string): void => {
  window.location.href = url;
};

/**
 * Open URL in a new tab
 *
 * @param url - The URL to open
 */
export const openInNewTab = (url: string): void => {
  window.open(url, '_blank', 'noopener,noreferrer');
};
