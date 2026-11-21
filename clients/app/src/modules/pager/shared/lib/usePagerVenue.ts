import { useQuery } from '@tanstack/react-query';
import { configClient } from '../../../../shared/lib/clients';
import { useUserStore } from '../../../../shared/stores/userStore';
import { pagerJoinUrl } from '../../pagerPaths';
import { usePagerSandbox } from './pagerClients';

interface PagerVenue {
  name: string;
  /** Where the counter QR points; empty until the business slug is known. */
  joinUrl: string;
  /** True while the slug or name is still being fetched. */
  loading: boolean;
}

/**
 * The signed-in business as its customers meet it on the pager. The user store
 * usually knows the slug; the business config fills in whatever it lacks.
 * The public demo has no business behind it, so no venue and no lookup.
 */
export function usePagerVenue(): PagerVenue {
  const sandbox = usePagerSandbox();
  const user = useUserStore((state) => state.user);
  const knownSlug = user?.businessSlug?.trim() || user?.business_slug?.trim() || '';
  const knownName = user?.businessName?.trim() || '';
  const config = useQuery({
    queryKey: ['pager', 'business-config'],
    enabled: !sandbox && (knownSlug === '' || knownName === ''),
    queryFn: async () => {
      const { business } = await configClient.getConfig({});
      return { slug: business?.slug?.trim() ?? '', name: business?.name?.trim() ?? '' };
    },
  });
  if (sandbox) return { name: '', joinUrl: '', loading: false };
  const slug = knownSlug || config.data?.slug || '';
  return {
    name: knownName || config.data?.name || '',
    joinUrl: slug ? pagerJoinUrl(slug) : '',
    loading: config.isLoading,
  };
}
