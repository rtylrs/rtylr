import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { UserX } from 'lucide-react';
import { profileClient } from '../../shared/lib/clients';

interface ProfileProject {
  id?: string;
  name?: string;
  icon?: string;
  color?: string;
  description?: string;
  status?: string;
  taskCount?: number;
  completedCount?: number;
}

export const Profile: React.FC<{ userId: string }> = ({ userId }) => {
  const { t } = useTranslation();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['profile', userId],
    queryFn: async () => {
      const res = await profileClient.getProfile({ userId });
      return (res as { profile?: import('@voxire/voxie/flow/profile/v1/types').Profile }).profile;
    },
    enabled: !!userId,
  });

  if (isLoading) {
    return (
      <div className="h-full overflow-y-auto">
        <div className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
          <div
            className="h-20 rounded-lg animate-pulse mb-4"
            style={{ background: 'var(--flow-hover)' }}
          />
          <div className="grid gap-3 sm:grid-cols-2">
            {[...Array(4)].map((_, i) => (
              <div
                key={i}
                className="h-[88px] rounded-lg animate-pulse"
                style={{ background: 'var(--flow-hover)' }}
              />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      // The S-24 shell rather than a bare sentence: a 44px glyph, a heading
      // and one line, sized to its content.
      <div className="flex flex-col items-center justify-center px-6 py-16 text-center">
        <div
          className="mb-4 flex h-11 w-11 items-center justify-center rounded-full"
          style={{ background: 'var(--flow-hover)' }}
        >
          <UserX size={20} style={{ color: 'var(--flow-text-muted)' }} aria-hidden />
        </div>
        <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.profile.notFound')}
        </h1>
        <p
          className="mt-1.5 max-w-[44ch] text-flow-body"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.profile.notFoundHint')}
        </p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
        {/* Identity header */}
        {/* S-23: avatar 64px, name 20px/600, role 13px muted. The contract
            carries no department field, so only the role can be shown. */}
        <div className="mb-6 flex items-center gap-4">
          <div
            className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full text-flow-title font-bold"
            style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
            aria-hidden
          >
            {data.initials || '?'}
          </div>
          <div className="min-w-0">
            <h1 className="text-flow-screen font-semibold" style={{ color: 'var(--flow-text)' }}>
              {data.name}
            </h1>
            {data.roleTitle ? (
              <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
                {data.roleTitle}
              </p>
            ) : null}
          </div>
        </div>

        {/* Public projects */}
        <section aria-label={t('flow.profile.publicProjects')}>
          <h2
            className="mb-3 text-flow-section font-semibold"
            style={{ color: 'var(--flow-text)' }}
          >
            {t('flow.profile.publicProjects')}
          </h2>

          {(data.publicProjects ?? []).length === 0 ? (
            <p className="py-4 text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.profile.noPublicProjects')}
            </p>
          ) : (
            /* Chips, the same shape a project takes on a task row, a board
               card and a breadcrumb, so the colour a person recognises the
               project by finally appears here too. G-C5.
               No icon: it would carry the raw name through as text, which is
               what #1085 is about. The colour dot identifies the project. */
            <ul className="flex flex-wrap gap-2">
              {(data.publicProjects ?? []).map((pp: ProfileProject) => (
                <li key={pp.id}>
                  <span
                    className="inline-flex min-h-8 items-center gap-2 rounded-full px-3 text-flow-meta font-medium"
                    style={{ background: 'var(--flow-active)', color: 'var(--flow-text)' }}
                  >
                    <span
                      className="h-2 w-2 shrink-0 rounded-full"
                      style={{ background: pp.color || 'var(--accent)' }}
                      aria-hidden
                    />
                    {pp.name}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
};

export default Profile;
