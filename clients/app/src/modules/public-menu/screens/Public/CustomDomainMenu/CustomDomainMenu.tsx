import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { extractTenantSubdomain } from '@voxire/config';
import { EmptyState } from '@voxire/ui';
import { SearchX } from 'lucide-react';
import { resolvePublicCustomDomain } from '../../../shared/hooks/usePublicBrandBundle';
import { menuByBusinessIdPath, menuHomePath, menuLocationPath } from '../../../menuPaths';
import SlugLanding from '../SlugLanding/SlugLanding';
import LocationMenu from '../SlugLanding/LocationMenu';

function useResolvedTenantHost(): {
  loading: boolean;
  slug: string | null;
  businessId: number;
} {
  const host = typeof window !== 'undefined' ? window.location.hostname : '';
  const directSubdomain = extractTenantSubdomain(host);

  const [state, setState] = useState<{
    loading: boolean;
    slug: string | null;
    businessId: number;
  }>(() =>
    directSubdomain
      ? { loading: false, slug: directSubdomain, businessId: 0 }
      : { loading: Boolean(host), slug: null, businessId: 0 }
  );

  useEffect(() => {
    if (directSubdomain || !host) return;
    let cancelled = false;
    void resolvePublicCustomDomain(host)
      .then((res) => {
        if (cancelled) return;
        if (!res.found) {
          setState({ loading: false, slug: null, businessId: 0 });
          return;
        }
        setState({
          loading: false,
          slug: res.businessSlug ? res.businessSlug.toLowerCase() : null,
          businessId: res.businessId ?? 0,
        });
      })
      .catch(() => {
        if (!cancelled) setState({ loading: false, slug: null, businessId: 0 });
      });
    return () => {
      cancelled = true;
    };
  }, [directSubdomain, host]);

  return state;
}

const TenantHostNotFound: React.FC = () => {
  const { t } = useTranslation('locationmenu');
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '0 20px',
      }}
    >
      <EmptyState
        icon={SearchX}
        title={t('locationmenu.businessNotFoundTitle')}
        description={t('locationmenu.businessNotFoundDescription')}
      />
    </div>
  );
};

/**
 * Mounted at `/` on `<slug>.rtylr.com` and verified custom domains.
 * Renders the business storefront directly while keeping the branded origin
 * in the browser address bar.
 */
export const TenantHostLanding: React.FC = () => {
  const navigate = useNavigate();
  const { loading, slug, businessId } = useResolvedTenantHost();

  useEffect(() => {
    if (!loading && !slug && businessId > 0) {
      const search = typeof window !== 'undefined' ? window.location.search : '';
      void navigate({ to: `${menuByBusinessIdPath(businessId)}${search}`, replace: true });
    }
  }, [loading, slug, businessId, navigate]);

  if (loading || (!slug && businessId > 0)) return <FullScreenLoader />;
  if (!slug) return <TenantHostNotFound />;
  return <SlugLanding slugOverride={slug} />;
};

/**
 * Mounted at `/$location` on `<slug>.rtylr.com` and verified custom domains.
 */
export const TenantHostLocationMenu: React.FC = () => {
  const { loading, slug } = useResolvedTenantHost();
  if (loading) return <FullScreenLoader />;
  if (!slug) return <TenantHostNotFound />;
  return <LocationMenu slugOverride={slug} />;
};

/**
 * Resolves verified custom-domain menu routes (`/menu` and `/menu/$location`)
 * to the tenant's canonical public menu while preserving QR query parameters
 * such as `?table=...&tableLabel=...`.
 */
const CustomDomainMenu: React.FC = () => {
  const { location } = useParams({ strict: false });
  const navigate = useNavigate();
  const { loading, slug, businessId } = useResolvedTenantHost();

  useEffect(() => {
    if (loading) return;
    const search = typeof window !== 'undefined' ? window.location.search : '';
    if (slug) {
      const base = location ? menuLocationPath(slug, location) : menuHomePath(slug);
      void navigate({ to: `${base}${search}`, replace: true });
      return;
    }
    if (businessId > 0) {
      void navigate({ to: `${menuByBusinessIdPath(businessId)}${search}`, replace: true });
    }
  }, [loading, slug, businessId, location, navigate]);

  if (loading || slug || businessId > 0) return <FullScreenLoader />;
  return <TenantHostNotFound />;
};

export default CustomDomainMenu;
