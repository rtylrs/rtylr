import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { UserCircle } from 'lucide-react';
import { Avatar, EmptyState } from '@voxire/ui';
import { CardSkeleton } from '../../shared/components/Skeleton';
import { useUserStore } from '../../shared/stores/userStore';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { todayLocalDate, useWorkDaysRange } from '../time/useWorkDay';
import { addDays, startOfWeek } from '../../shared/lib/dateUtils';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { useMyProfileDetail, useMyStaffRecord } from './useProfile';

const cardStyle: React.CSSProperties = {
  background: 'var(--flow-card)',
  border: '1px solid var(--flow-border)',
};
const labelClass = 'mb-1 block text-[10px] font-semibold uppercase tracking-wider';

function initials(name: string): string {
  return name
    .split(' ')
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
}

interface FieldProps {
  label: string;
  value?: string;
}

const Field = ({ label, value }: FieldProps): React.JSX.Element | null => {
  if (!value) return null;
  return (
    <div>
      <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
        {label}
      </span>
      <p className="text-[13px]" style={{ color: 'var(--flow-text)' }}>
        {value}
      </p>
    </div>
  );
};

/**
 * Read-only: editing an HR record stays in HR, full stop. This screen shows
 * the caller their own record and their own Flow activity - it never writes
 * either back.
 *
 * Basic account info (name, email, avatar) comes from the auth/user session
 * and renders unconditionally - it has nothing to do with HR linkage. Only
 * the HR-specific section below it is gated behind the staff lookup, so a
 * user with no HR record (or a token lacking HR read access, e.g. a 403)
 * still sees their own profile instead of a blank page.
 */
const ProfileScreen = (): React.JSX.Element => {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const {
    data: staff,
    isLoading: isStaffLoading,
    refetch: refetchStaff,
  } = useMyStaffRecord(user?.id);
  const { data: detail } = useMyProfileDetail(staff?.id);

  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  const weekRange = useMemo(() => {
    const start = startOfWeek(todayLocalDate(), weekStart);
    return [start, addDays(start, 6)] as const;
  }, [weekStart]);
  const { data: weekData } = useWorkDaysRange(weekRange[0], weekRange[1]);
  const minutesThisWeek = useMemo(
    () => (weekData?.workDays ?? []).reduce((sum, day) => sum + (day.loggedMinutes ?? 0), 0),
    [weekData]
  );

  const displayName = user?.name ?? '';
  const subtitle = staff?.role || user?.email;
  const emergencyContact =
    staff?.emergencyContactName && staff?.emergencyContactPhone
      ? `${staff.emergencyContactName} · ${staff.emergencyContactPhone}`
      : undefined;

  return (
    <>
      <HelmetContainer title={`${t('flow.profile.heading')} · Flow`} />
      <div className="mx-auto max-w-2xl p-4 sm:p-6">
        <div className="mb-6 flex items-center gap-3">
          <Avatar
            src={detail?.photoUrl ?? user?.avatar}
            initials={initials(displayName)}
            size="lg"
          />
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--flow-text)' }}>
              {displayName}
            </h1>
            {subtitle && (
              <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
                {subtitle}
              </p>
            )}
          </div>
        </div>

        <section className="mb-4 rounded-lg p-4" style={cardStyle}>
          <h2
            className="mb-2 text-[13px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.profile.hrDetails')}
          </h2>
          {isStaffLoading ? (
            <CardSkeleton />
          ) : staff ? (
            <>
              <div className="grid gap-3 sm:grid-cols-2">
                <Field label={t('flow.profile.department')} value={staff.departmentName} />
                <Field label={t('flow.profile.employmentType')} value={staff.employmentType} />
                <Field label={t('flow.profile.status')} value={staff.status} />
                <Field label={t('flow.profile.startDate')} value={staff.startDate} />
                <Field label={t('flow.profile.email')} value={staff.email} />
                <Field label={t('flow.profile.phone')} value={staff.phone} />
                <Field label={t('flow.profile.address')} value={staff.address} />
                <Field label={t('flow.profile.emergencyContact')} value={emergencyContact} />
              </div>
              {detail?.bio && (
                <div className="mt-3">
                  <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.profile.bio')}
                  </span>
                  <p className="text-[13px]" style={{ color: 'var(--flow-text)' }}>
                    {detail.bio}
                  </p>
                </div>
              )}
            </>
          ) : (
            <EmptyState
              icon={UserCircle}
              title={t('flow.profile.noRecordTitle')}
              description={t('flow.profile.noRecordDescription')}
              action={{
                label: t('flow.profile.noRecordCta'),
                onClick: () => void refetchStaff(),
              }}
            />
          )}
        </section>

        <section className="rounded-lg p-4" style={cardStyle}>
          <h2
            className="mb-2 text-[13px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.profile.activity')}
          </h2>
          <p className="mb-2 text-[13px]" style={{ color: 'var(--flow-text)' }}>
            {t('flow.profile.hoursLoggedThisWeek', {
              hours: formatDurationClock(minutesThisWeek),
            })}
          </p>
          <Link to="/time" className="flow-focus text-[13px] font-medium underline">
            {t('flow.profile.viewMyDay')}
          </Link>
        </section>
      </div>
    </>
  );
};

export default ProfileScreen;
