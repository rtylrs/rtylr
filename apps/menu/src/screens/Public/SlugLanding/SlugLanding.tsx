import React from 'react';
import { useParams, useNavigate, Link } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { usePublicBusinessBySlug } from '../../../shared/hooks/usePublicMenu';
import Menu from '../Menu/Menu';

/**
 * Single screen mounted at `/:slugOrId`. Tries the new slug lookup
 * first. If that 404s and the param looks like a legacy mongo id,
 * falls back to the old menu renderer so existing QR codes keep
 * resolving.
 *
 * On hit, renders a location picker (or, if there's exactly one,
 * deep-links straight into it). `/:slugOrId/:location` is handled
 * by a sibling screen that re-resolves the slug + picks the
 * matching location by slug.
 */

const isMongoId = (s: string): boolean => /^[a-f0-9]{24}$/i.test(s);

const SlugLanding: React.FC = (): React.JSX.Element => {
  const { slug: raw } = useParams({ strict: false });
  const navigate = useNavigate();
  const slug = (raw ?? '').toLowerCase();

  const { data, isLoading, isError } = usePublicBusinessBySlug(slug);

  React.useEffect(() => {
    if (data && data.locations.length === 1) {
      const loc = data.locations[0];
      if (loc.slug) navigate({ to: `/${slug}/${loc.slug}`, replace: true });
    }
  }, [data, slug, navigate]);

  if (isLoading) return <FullScreenLoader />;

  if (isError) {
    if (isMongoId(slug)) return <Menu />;
    return (
      <div
        style={{
          padding: 32,
          textAlign: 'center',
          color: 'var(--menu-text, #333)',
        }}
      >
        <h1 style={{ fontSize: 18, fontWeight: 600 }}>Business not found</h1>
        <p style={{ fontSize: 14, opacity: 0.7 }}>
          The URL <code>/{slug}</code> doesn't match any active business.
        </p>
      </div>
    );
  }

  if (!data) return <FullScreenLoader />;

  return (
    <>
      <HelmetContainer title="Locations · rtylr" />
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
            Pick a location to see its menu
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
              No active locations yet.
            </li>
          )}
          {data.locations.map((loc) => (
            <li key={loc.id}>
              <Link
                to={`/${slug}/${loc.slug}`}
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
