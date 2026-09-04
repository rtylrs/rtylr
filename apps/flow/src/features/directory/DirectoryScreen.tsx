import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Search, Users, AlertCircle } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { Avatar, EmptyState } from '@voxire/ui';
import { CardSkeleton } from '../../shared/components/Skeleton';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { useDirectory } from './useDirectory';

const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border)',
  color: 'var(--flow-text)',
};

function initials(name: string): string {
  return name
    .split(' ')
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
}

/**
 * Business-wide roster, backed by hr.staff.v1's listStaff - gated
 * server-side by view_directory (onion PR #267), so a caller without it
 * gets a real 403 here, not just a hidden screen.
 */
const DirectoryScreen = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const directoryQuery = useDirectory(search);
  const { data: members, isLoading, refetch } = directoryQuery;
  const isError = hasQueryFailed(directoryQuery);

  return (
    <>
      <HelmetContainer title={`${t('flow.directory.heading')} · Flow`} />
      <div className="mx-auto max-w-2xl p-4 sm:p-6">
        <h1 className="mb-4 text-lg font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.directory.heading')}
        </h1>

        <div className="relative mb-4">
          <Search
            size={14}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2"
            style={{ color: 'var(--flow-text-muted)' }}
            aria-hidden="true"
          />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('flow.directory.searchPlaceholder')}
            aria-label={t('flow.directory.searchPlaceholder')}
            className="flow-focus min-h-[44px] w-full rounded-sm pl-9 pr-2 text-title md:min-h-0 md:py-1.5 md:text-body"
            style={inputStyle}
          />
        </div>

        {isLoading && <CardSkeleton />}

        {!isLoading && isError && (
          <EmptyState
            icon={AlertCircle}
            title={t('flow.directory.loadFailedTitle')}
            description={t('flow.directory.loadFailedDescription')}
            action={{
              label: t('flow.directory.retryCta'),
              onClick: () => void refetch(),
            }}
          />
        )}

        {!isLoading &&
          !isError &&
          (members?.length ?? 0) === 0 &&
          (search.trim().length > 0 ? (
            <EmptyState
              icon={Users}
              title={t('flow.directory.noResults')}
              action={{
                label: t('flow.directory.clearSearchCta'),
                onClick: () => setSearch(''),
              }}
            />
          ) : (
            <EmptyState
              icon={Users}
              title={t('flow.directory.noMembersYet')}
              action={{
                label: t('flow.directory.refreshCta'),
                onClick: () => void refetch(),
              }}
            />
          ))}

        {!isLoading && !isError && members && members.length > 0 && (
          <ul className="space-y-1.5">
            {members.map((member) => {
              const fullName = `${member.firstName ?? ''} ${member.lastName ?? ''}`.trim();
              return (
                <li
                  key={member.id}
                  className="flex items-center gap-3 rounded-md p-2"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <Avatar initials={initials(fullName)} size="sm" />
                  <div className="min-w-0 flex-1">
                    <div className="text-body font-medium" style={{ color: 'var(--flow-text)' }}>
                      {fullName}
                    </div>
                    <div className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                      {[member.role, member.departmentName].filter(Boolean).join(' · ')}
                    </div>
                  </div>
                  <div
                    className="hidden shrink-0 text-meta sm:block"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {member.email}
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </>
  );
};

export default DirectoryScreen;
