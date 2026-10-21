import { createOneKitClient } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import { sessionStorageSet, sessionStorageRemove } from './storage';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

export const SOCIAL_AUTH_RETURN_KEY = 'social_auth_return';
export const SOCIAL_AUTH_REDIRECT_KEY = 'social_auth_redirect';
export const SOCIAL_AUTH_BUSINESS_NAME_KEY = 'social_auth_business_name';
export const GOOGLE_AUTH_STATE_KEY = 'google_auth_state';

// Google sign-in is switched off for now. Set this to true to bring back the
// "Continue with Google" button on the sign-in and sign-up pages.
export const GOOGLE_SIGN_IN_ENABLED = false;

// Keep the redirect target and OAuth state in this tab only for callback validation.
export async function startGoogleLogin(
  redirectParam: string | null,
  businessName?: string,
  returnTo?: string | null
): Promise<void> {
  const { url } = await authClient.googleAuthorize({});
  // Optional on the response. Navigating to undefined would put the literal
  // string in the address bar, so the caller's catch gets a real failure.
  if (url === undefined || url === '') throw new Error('Social sign-in returned no redirect URL');
  const authorizeURL = new URL(url);
  const state = authorizeURL.searchParams.get('state');
  if (authorizeURL.origin !== 'https://accounts.google.com' || !state) {
    throw new Error('Google sign-in returned an invalid authorization URL');
  }
  sessionStorageRemove(SOCIAL_AUTH_RETURN_KEY);
  sessionStorageRemove(SOCIAL_AUTH_REDIRECT_KEY);
  sessionStorageRemove(SOCIAL_AUTH_BUSINESS_NAME_KEY);
  if (returnTo) sessionStorageSet(SOCIAL_AUTH_RETURN_KEY, returnTo);
  if (redirectParam) sessionStorageSet(SOCIAL_AUTH_REDIRECT_KEY, redirectParam);
  if (businessName?.trim()) sessionStorageSet(SOCIAL_AUTH_BUSINESS_NAME_KEY, businessName.trim());
  sessionStorageSet(GOOGLE_AUTH_STATE_KEY, state);
  window.location.href = authorizeURL.toString();
}
