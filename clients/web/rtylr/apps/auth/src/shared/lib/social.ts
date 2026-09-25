import { createOneKitClient } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import { sessionStorageSet } from './storage';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

export type SocialProvider = 'google' | 'github' | 'microsoft' | 'apple';

export const SOCIAL_PROVIDERS: SocialProvider[] = ['google', 'github', 'microsoft', 'apple'];
export const SOCIAL_AUTH_REDIRECT_KEY = 'social_auth_redirect';
export const SOCIAL_AUTH_BUSINESS_NAME_KEY = 'social_auth_business_name';

// startSocialLogin fetches the WorkOS-hosted consent URL for the provider and
// redirects the browser to it. The post-auth redirect target is stashed in sessionStorage
// so the OAuth callback can route the user back where they started without persisting in localStorage.
export async function startSocialLogin(
  provider: SocialProvider,
  redirectParam: string | null,
  businessName?: string
): Promise<void> {
  const { url } = await authClient.socialAuthorize({ provider });
  // Optional on the response. Navigating to undefined would put the literal
  // string in the address bar, so the caller's catch gets a real failure.
  if (url === undefined || url === '') throw new Error('Social sign-in returned no redirect URL');
  if (redirectParam) sessionStorageSet(SOCIAL_AUTH_REDIRECT_KEY, redirectParam);
  if (businessName?.trim()) sessionStorageSet(SOCIAL_AUTH_BUSINESS_NAME_KEY, businessName.trim());
  window.location.href = url;
}
