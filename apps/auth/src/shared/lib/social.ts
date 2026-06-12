import { unwrap } from '@voxire/api';
import { safe } from './api';
import { storageSet } from './storage';

export type SocialProvider = 'google' | 'github' | 'microsoft' | 'apple';

export const SOCIAL_PROVIDERS: SocialProvider[] = ['google', 'github', 'microsoft', 'apple'];
export const SOCIAL_AUTH_REDIRECT_KEY = 'social_auth_redirect';
export const SOCIAL_AUTH_BUSINESS_NAME_KEY = 'social_auth_business_name';

// startSocialLogin fetches the WorkOS-hosted consent URL for the provider and
// redirects the browser to it. The post-auth redirect target is stashed so the
// OAuth callback can route the user back where they started.
export async function startSocialLogin(
  provider: SocialProvider,
  redirectParam: string | null,
  businessName?: string
): Promise<void> {
  const { url } = await unwrap(
    safe.get<{ url: string }>('/auth/social/authorize', { params: { provider } })
  );
  if (redirectParam) storageSet(SOCIAL_AUTH_REDIRECT_KEY, redirectParam);
  if (businessName?.trim()) storageSet(SOCIAL_AUTH_BUSINESS_NAME_KEY, businessName.trim());
  window.location.href = url;
}
