import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { SearchX } from 'lucide-react';
import { resolvePublicCustomDomain } from '../../../shared/hooks/usePublicBrandBundle';
import { menuByBusinessIdPath, menuHomePath, menuLocationPath } from '../../../menuPaths';

/**
 * Resolves verified custom-domain menu routes (`/menu` and `/menu/$location`)
 * to the tenant's canonical public menu while preserving QR query parameters
 * such as `?table=...&tableLabel=...`.
 */
const CustomDomainMenu: React.FC = () => {
  const { t } = useTranslation('locationmenu');
  const { location } = useParams({ strict: false });
  const navigate = useNavigate();
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const host = typeof window !== 'undefined' ? window.location.hostname : '';
    const search = typeof window !== 'undefined' ? window.location.search : '';
    if (!host) {
      setNotFound(true);
      return;
    }

    void resolvePublicCustomDomain(host)
      .then((res) => {
        if (cancelled) return;
        if (!res.found) {
          setNotFound(true);
          return;
        }
        if (res.businessSlug) {
          const base = location
            ? menuLocationPath(res.businessSlug, location)
            : menuHomePath(res.businessSlug);
          void navigate({ to: `${base}${search}`, replace: true });
          return;
        }
        if (res.businessId > 0) {
          void navigate({ to: `${menuByBusinessIdPath(res.businessId)}${search}`, replace: true });
          return;
        }
        setNotFound(true);
      })
      .catch(() => {
        if (!cancelled) setNotFound(true);
      });

    return () => {
      cancelled = true;
    };
  }, [location, navigate]);

  if (!notFound) return <FullScreenLoader />;

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

export default CustomDomainMenu;
