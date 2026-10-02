import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useParams, useNavigate, Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import {
  Mail,
  Phone,
  MapPin,
  Calendar,
  Briefcase,
  FileText,
  Shield,
  ChevronLeft,
  GraduationCap,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useStaffById,
  useDocuments,
  useTrainingProgress,
  useBenefitEnrollments,
} from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Pill,
  Skeleton,
  Breadcrumbs,
  ACCENT,
  useIsMobile,
} from '../../shared/components/ui';
import TrainingProgress from '../../shared/components/TrainingProgress';
import { initials, paletteFor } from '../../shared/components/EmployeeCard';

const EmployeeProfile: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const isMobile = useIsMobile(900);
  const { id } = useParams({ strict: false });
  const staffId = id != null ? Number(id) : undefined;
  // Everything this screen shows comes from getStaff. It used to also call
  // /hr/staff/detail, which no onion module serves, for phone, address,
  // start date and the emergency contact - all of which getStaff already
  // returns. See #875.
  const { data: staff, isLoading } = useStaffById(staffId);
  const { data: docs = [] } = useDocuments();
  const { data: training = [] } = useTrainingProgress();
  const { data: enrollments = [] } = useBenefitEnrollments();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Skeleton height={80} />
        <Skeleton height={200} />
      </div>
    );
  }

  if (!staff) {
    return (
      <EmptyState
        icon={Briefcase}
        title={t('hrStaff.profile.notFoundTitle')}
        description={t('hrStaff.profile.notFoundDescription')}
        action={{
          label: t('hrStaff.profile.backToStaff'),
          onClick: () => navigate({ to: '/hr/staff' }),
        }}
      />
    );
  }

  const { bg, fg } = paletteFor(staff.id);
  const myDocs = docs.filter((d) => d.staffId === staff.id);
  const myTraining = training.filter((t) => t.staffId === staff.id);
  const myBenefits = enrollments.filter((e) => e.staffId === staff.id);

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.employeeProfile'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Breadcrumbs
          items={[{ label: t('hrNav.staff'), href: '/hr/staff' }, { label: staff.name }]}
        />

        {/* Hero */}
        <Card
          style={{
            display: 'flex',
            alignItems: isMobile ? 'flex-start' : 'center',
            flexDirection: isMobile ? 'column' : 'row',
            gap: 20,
            padding: 24,
          }}
        >
          <Link to="/hr/staff" style={{ color: 'var(--hr-text-muted)' }}>
            <ChevronLeft size={16} />
          </Link>
          <div
            style={{
              width: 80,
              height: 80,
              borderRadius: 20,
              background: bg,
              color: fg,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 24,
              fontWeight: 700,
              flexShrink: 0,
            }}
          >
            {initials(staff.name)}
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <h1 style={{ margin: 0, fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
              {staff.name}
            </h1>
            <p style={{ margin: '4px 0', fontSize: 13, color: 'var(--hr-text-muted)' }}>
              {staff.role} • {staff.department}
            </p>
            <div style={{ display: 'flex', gap: 6, marginTop: 8, flexWrap: 'wrap' }}>
              <Pill color="#8b5cf6">{staff.employmentType}</Pill>
              <Pill color={staff.status === 'Active' ? '#10b981' : '#ef4444'}>{staff.status}</Pill>
            </div>
          </div>
        </Card>

        {/* Contact + Employment */}
        <div
          style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
        >
          <Card>
            <SectionHeader title={t('hrStaff.profile.contact')} />
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 10,
                marginTop: 12,
                fontSize: 12,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Mail size={13} style={{ color: 'var(--hr-text-muted)' }} />
                <span style={{ color: 'var(--hr-text)' }}>{staff.email}</span>
              </div>
              {staff.phone && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Phone size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>{staff.phone}</span>
                </div>
              )}
              {staff.address && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <MapPin size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>{staff.address}</span>
                </div>
              )}
            </div>
          </Card>

          <Card>
            <SectionHeader title={t('hrStaff.profile.employment')} />
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 10,
                marginTop: 12,
                fontSize: 12,
              }}
            >
              {staff.startDate && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Calendar size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>
                    {t('hrStaff.profile.hiredOn', { date: staff.startDate })}
                  </span>
                </div>
              )}
            </div>
          </Card>
        </div>

        {/* Emergency contact and About. hr.staff.v1 now carries bio (#875), so the
            About card is back. Skills is still absent: it has no contract field
            and no column, and it was a separate feature that was never built. */}
        <div
          style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
        >
          <Card>
            <SectionHeader title={t('hrStaff.profile.emergencyHeading')} />
            {staff.emergencyContactName || staff.emergencyContactPhone ? (
              <div style={{ marginTop: 12, fontSize: 12, lineHeight: 1.7 }}>
                {staff.emergencyContactName && (
                  <div style={{ color: 'var(--hr-text)', fontWeight: 600 }}>
                    {staff.emergencyContactName}
                  </div>
                )}
                {staff.emergencyContactPhone && (
                  <div style={{ color: 'var(--hr-text-muted)' }}>{staff.emergencyContactPhone}</div>
                )}
              </div>
            ) : (
              <p style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 12 }}>
                {t('hrStaff.profile.emergencyNotProvided')}
              </p>
            )}
          </Card>

          <Card>
            <SectionHeader title={t('hrStaff.profile.aboutHeading')} />
            <p
              style={{
                margin: '10px 0 0',
                fontSize: 12,
                color: 'var(--hr-text-muted)',
                lineHeight: 1.6,
              }}
            >
              {staff.bio || t('hrStaff.profile.noBio')}
            </p>
          </Card>
        </div>

        {/* Documents */}
        <Card>
          <SectionHeader
            title={t('hrStaff.profile.documents')}
            subtitle={`${myDocs.length} files`}
          />
          <div style={{ marginTop: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
            {myDocs.length === 0 && (
              <EmptyState
                icon={FileText}
                title={t('hrStaff.profile.noDocuments')}
                action={{
                  label: t('hrStaff.profile.goToDocuments'),
                  onClick: () => navigate({ to: '/hr/documents' }),
                }}
                size="compact"
              />
            )}
            {myDocs.map((d) => (
              <div
                key={d.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: 8,
                  borderRadius: 8,
                  background: 'var(--hr-hover)',
                }}
              >
                <FileText size={14} style={{ color: 'var(--hr-text-muted)' }} />
                <div style={{ flex: 1, fontSize: 12 }}>
                  <div style={{ color: 'var(--hr-text)', fontWeight: 600 }}>{d.name}</div>
                  <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>
                    {d.type} • {d.sizeKb}KB • {d.uploadedAt}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Training */}
        <Card>
          <SectionHeader title={t('hrStaff.profile.training')} />
          <div style={{ marginTop: 12, display: 'flex', flexDirection: 'column', gap: 12 }}>
            {myTraining.length === 0 && (
              <EmptyState
                icon={GraduationCap}
                title={t('hrStaff.profile.noCourses')}
                action={{
                  label: t('hrStaff.profile.goToTraining'),
                  onClick: () => navigate({ to: '/hr/training' }),
                }}
                size="compact"
              />
            )}
            {myTraining.map((t) => (
              <TrainingProgress key={t.id} progress={t.progress} label={t.courseName} />
            ))}
          </div>
        </Card>

        {/* Benefits */}
        <Card>
          <SectionHeader title={t('hrStaff.profile.benefits')} />
          <div style={{ marginTop: 12, display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {myBenefits.length === 0 && (
              <EmptyState
                icon={Shield}
                title={t('hrStaff.profile.noBenefits')}
                action={{
                  label: t('hrStaff.profile.goToBenefits'),
                  onClick: () => navigate({ to: '/hr/benefits' }),
                }}
                size="compact"
              />
            )}
            {myBenefits.map((b) => (
              <div
                key={b.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 6,
                  padding: '6px 10px',
                  background: 'var(--hr-hover)',
                  borderRadius: 8,
                  fontSize: 11,
                }}
              >
                <Shield size={12} style={{ color: ACCENT }} /> {b.planName}
              </div>
            ))}
          </div>
        </Card>
      </div>
    </>
  );
};

export default EmployeeProfile;
