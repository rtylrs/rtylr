import React from 'react';
import { useParams, Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import {
  Mail,
  Phone,
  MapPin,
  Calendar,
  Briefcase,
  FileText,
  Heart,
  Shield,
  ChevronLeft,
} from 'lucide-react';
import {
  useStaffById,
  useEmployeeProfile,
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
  EmptyState,
  useIsMobile,
} from '../../shared/components/ui';
import TrainingProgress from '../../shared/components/TrainingProgress';
import { initials, paletteFor } from '../../shared/components/EmployeeCard';

const EmployeeProfile: React.FC = () => {
  const isMobile = useIsMobile(900);
  const { id } = useParams({ strict: false });
  const staffId = id ? parseInt(id, 10) : undefined;
  const { data: staff, isLoading } = useStaffById(staffId);
  const { data: profile } = useEmployeeProfile(staffId);
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
        icon={<Briefcase size={20} />}
        title="Employee not found"
        description="The employee you're looking for may have been removed."
        action={
          <Link to="/staff" style={{ color: ACCENT, fontSize: 12, textDecoration: 'none' }}>
            Back to staff
          </Link>
        }
      />
    );
  }

  const { bg, fg } = paletteFor(staff.id);
  const myDocs = docs.filter((d) => d.staffId === staff.id);
  const myTraining = training.filter((t) => t.staffId === staff.id);
  const myBenefits = enrollments.filter((e) => e.staffId === staff.id);

  return (
    <>
      <HelmetContainer title="Employee profile · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Breadcrumbs items={[{ label: 'Staff', href: '/staff' }, { label: staff.name }]} />

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
          <Link to="/staff" style={{ color: 'var(--hr-text-muted)' }}>
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
            <SectionHeader title="Contact" />
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
              {profile?.phone && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Phone size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>{profile.phone}</span>
                </div>
              )}
              {profile?.address && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <MapPin size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>{profile.address}</span>
                </div>
              )}
            </div>
          </Card>

          <Card>
            <SectionHeader title="Employment" />
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 10,
                marginTop: 12,
                fontSize: 12,
              }}
            >
              {profile?.hiredAt && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Calendar size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>Hired {profile.hiredAt}</span>
                </div>
              )}
              {profile?.birthday && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Heart size={13} style={{ color: 'var(--hr-text-muted)' }} />
                  <span style={{ color: 'var(--hr-text)' }}>Birthday {profile.birthday}</span>
                </div>
              )}
            </div>
          </Card>
        </div>

        {/* Emergency contact + bio */}
        <div
          style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
        >
          <Card>
            <SectionHeader title="Emergency contact" />
            {profile?.emergency ? (
              <div style={{ marginTop: 12, fontSize: 12, lineHeight: 1.7 }}>
                <div style={{ color: 'var(--hr-text)', fontWeight: 600 }}>
                  {profile.emergency.name}
                </div>
                <div style={{ color: 'var(--hr-text-muted)' }}>{profile.emergency.relation}</div>
                <div style={{ color: 'var(--hr-text-muted)' }}>{profile.emergency.phone}</div>
              </div>
            ) : (
              <p style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 12 }}>
                Not provided
              </p>
            )}
          </Card>

          <Card>
            <SectionHeader title="About" />
            <p
              style={{
                margin: '10px 0 0',
                fontSize: 12,
                color: 'var(--hr-text-muted)',
                lineHeight: 1.6,
              }}
            >
              {profile?.bio ?? 'No bio.'}
            </p>
            {profile?.skills && profile.skills.length > 0 && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 12 }}>
                {profile.skills.map((s) => (
                  <Pill key={s}>{s}</Pill>
                ))}
              </div>
            )}
          </Card>
        </div>

        {/* Documents */}
        <Card>
          <SectionHeader title="Documents" subtitle={`${myDocs.length} files`} />
          <div style={{ marginTop: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
            {myDocs.length === 0 && (
              <p style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>No documents on file.</p>
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
          <SectionHeader title="Training" />
          <div style={{ marginTop: 12, display: 'flex', flexDirection: 'column', gap: 12 }}>
            {myTraining.length === 0 && (
              <p style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>No courses yet.</p>
            )}
            {myTraining.map((t) => (
              <TrainingProgress key={t.id} progress={t.progress} label={t.courseName} />
            ))}
          </div>
        </Card>

        {/* Benefits */}
        <Card>
          <SectionHeader title="Benefits" />
          <div style={{ marginTop: 12, display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {myBenefits.length === 0 && (
              <p style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                No active benefit enrollments.
              </p>
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
