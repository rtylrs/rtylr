import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Globe, ChevronLeft } from 'lucide-react';
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
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="text-sm font-medium" style={{ color: 'var(--flow-text)' }}>
          {t('flow.profile.notFound')}
        </p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
        {/* Back */}
        <button
          type="button"
          onClick={() => window.history.back()}
          className="inline-flex items-center gap-1 text-xs mb-4 hover:opacity-80 transition-opacity"
          style={{ color: 'var(--flow-text-muted)' }}
          aria-label={t('flow.profile.back')}
        >
          <ChevronLeft size={14} /> {t('flow.profile.back')}
        </button>

        {/* Identity header */}
        <div className="flex items-center gap-3 mb-6">
          <div
            className="h-14 w-14 rounded-full flex items-center justify-center text-lg font-bold shrink-0"
            style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
            aria-hidden
          >
            {data.initials || '?'}
          </div>
          <div>
            <h1 className="text-title font-bold" style={{ color: 'var(--flow-text)' }}>
              {data.name}
            </h1>
            {data.roleTitle ? (
              <p className="text-xs" style={{ color: 'var(--flow-text-muted)' }}>
                {data.roleTitle}
              </p>
            ) : null}
          </div>
        </div>

        {/* Public projects */}
        <section aria-label={t('flow.profile.publicProjects')}>
          <h2 className="text-sm font-semibold mb-3" style={{ color: 'var(--flow-text)' }}>
            {t('flow.profile.publicProjects')}
          </h2>

          {(data.publicProjects ?? []).length === 0 ? (
            <p className="text-xs py-4" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.profile.noPublicProjects')}
            </p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {(data.publicProjects ?? []).map((pp: ProfileProject) => (
                <div
                  key={pp.id}
                  className="rounded-lg border p-3"
                  style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-surface)' }}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <span
                        className="h-7 w-7 rounded-md flex items-center justify-center text-xs shrink-0"
                        style={{
                          background: `${pp.color ?? '#64C5B9'}20`,
                          color: pp.color ?? '#64C5B9',
                        }}
                        aria-hidden
                      >
                        {pp.icon || '📋'}
                      </span>
                      <p
                        className="font-medium text-sm truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {pp.name}
                      </p>
                    </div>
                    <Globe size={12} style={{ color: 'var(--flow-text-muted)' }} aria-hidden />
                  </div>
                  {pp.description ? (
                    <p
                      className="mt-1.5 text-xs line-clamp-2"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {pp.description}
                    </p>
                  ) : null}
                  <div
                    className="mt-2 flex items-center gap-2 text-xs"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <span>{pp.taskCount ?? 0} tasks</span>
                    <span>{pp.completedCount ?? 0} done</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

export default Profile;
