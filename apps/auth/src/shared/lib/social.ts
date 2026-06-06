import { unwrap } from '@voxire/api';
import { safe } from './api';

export type SocialProvider = 'google' | 'github' | 'microsoft' | 'apple';

export const SOCIAL_PROVIDERS: SocialProvider[] = ['google', 'github', 'microsoft', 'apple'];

// startSocialLogin fetches the WorkOS-hosted consent URL for the provider and
// redirects the browser to it. The post-auth redirect target is stashed so the
// OAuth callback can route the user back where they started.
export async function startSocialLogin(
  provider: SocialProvider,
  redirectParam: string | null
): Promise<void> {
  const { url } = await unwrap(
    safe.get<{ url: string }>('/auth/social/authorize', { params: { provider } })
  );
  if (redirectParam) localStorage.setItem('social_auth_redirect', redirectParam);
  window.location.href = url;
}
