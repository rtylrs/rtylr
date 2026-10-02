import React from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate, useRouter, Link } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { SearchX } from 'lucide-react';
import { usePublicBusinessBySlug } from '../../../shared/hooks/usePublicMenu';

/**
 * Single screen mounted at `/:slugOrId`. Tries the new slug lookup
 * first. Legacy Mongo menu URLs are no longer resolved because their
 * transport has no OneKit contract.
 *
 * On hit, renders a location picker (or, if there's exactly one,
 * deep-links straight into it). `/:slugOrId/:location` is handled
 * by a sibling screen that re-resolves the slug + picks the
 * matching location by slug.
 */

const SlugLanding: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('sluglanding');
  const { slug: raw } = useParams({ strict: false });
  const navigate = useNavigate();
  const router = useRouter();
  const slug = (raw ?? '').toLowerCase();

  const { data, isLoading, isError } = usePublicBusinessBySlug(slug);

  React.useEffect(() => {
    if (data && data.locations.length === 1) {
      const loc = data.locations[0];
      if (loc.slug) {
        navigate({
          to: '/$slug/$location',
          params: { slug, location: loc.slug },
          replace: true,
        });
      }
    }
  }, [data, slug, navigate]);

  if (isLoading) return <FullScreenLoader />;

  if (isError) {
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
          title={t('sluglanding.businessNotFoundTitle')}
          description={t('sluglanding.businessNotFoundDescription', { slug })}
          action={{ label: t('sluglanding.goBack'), onClick: () => router.history.back() }}
        />
      </div>
    );
  }

  if (!data) return <FullScreenLoader />;

  return (
    <>
      <HelmetContainer title={t('sluglanding.pageTitle')} />
      <div style={{ minHeight: '100vh', background: 'var(--menu-bg, #fafafa)' }}>
        <header style={{ padding: '24px 16px', textAlign: 'center' }}>
          {data.business.logoUrl && (
            <img
              src={data.business.logoUrl}
              alt={data.business.name}
              style={{ height: 56, margin: '0 auto 12px', display: 'block' }}
            />
          )}
          <h1 style={{ fontSize: 22, fontWeight: 700, margin: 0 }}>{data.business.name}</h1>
          <p style={{ fontSize: 13, opacity: 0.65, marginTop: 4 }}>
            {t('sluglanding.pickLocation')}
          </p>
        </header>

        <ul
          style={{
            listStyle: 'none',
            padding: '0 16px',
            margin: '0 auto',
            maxWidth: 480,
            display: 'flex',
            flexDirection: 'column',
            gap: 10,
          }}
        >
          {data.locations.length === 0 && (
            <li
              style={{
                padding: 16,
                textAlign: 'center',
                border: '1px dashed rgba(0,0,0,0.15)',
                borderRadius: 12,
                opacity: 0.7,
              }}
            >
              {t('sluglanding.noLocations')}
            </li>
          )}
          {data.locations.map((loc) => (
            <li key={loc.id}>
              <Link
                to="/$slug/$location"
                params={{ slug, location: loc.slug }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '14px 16px',
                  background: '#fff',
                  border: '1px solid rgba(0,0,0,0.08)',
                  borderRadius: 12,
                  textDecoration: 'none',
                  color: 'inherit',
                  boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                }}
              >
                <div>
                  <p style={{ fontSize: 15, fontWeight: 600, margin: 0 }}>{loc.name}</p>
                  {loc.address && (
                    <p style={{ fontSize: 12, opacity: 0.6, margin: '2px 0 0' }}>{loc.address}</p>
                  )}
                </div>
                <span style={{ fontSize: 18, opacity: 0.4 }}>›</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
};

export default SlugLanding;
