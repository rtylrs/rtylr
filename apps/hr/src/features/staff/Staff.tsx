import React, { useState, useEffect, useRef } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import type { Variants } from 'motion/react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { UserPlus, Mail, Search, Users } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { StaffMember } from '../../shared/services/hrApi';
import { useStaff } from './useStaff';
import { useIsMobile } from '../../shared/components/ui';

/** The themed accent, so a workspace's own colour is used (rtylr#1242). */
const ACCENT = 'var(--accent)';

type EmploymentType = 'Full-time' | 'Part-time' | 'Contractor';
type FilterTab = 'All' | EmploymentType;
const filterTabs: FilterTab[] = ['All', 'Full-time', 'Part-time', 'Contractor'];

function getInitials(name: string) {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const avatarColors = [
  { bg: 'rgba(139,92,246,0.15)', color: '#8b5cf6' },
  { bg: 'rgba(16,185,129,0.15)', color: '#10b981' },
  { bg: 'rgba(59,130,246,0.15)', color: '#3b82f6' },
  { bg: 'rgba(245,158,11,0.15)', color: '#f59e0b' },
  { bg: 'rgba(239,68,68,0.15)', color: '#ef4444' },
  { bg: 'rgba(100,197,185,0.15)', color: '#64C5B9' },
];

const container: Variants = { hidden: {}, show: { transition: { staggerChildren: 0.06 } } };
const cardAnim: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const Staff: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile(900);
  const [activeTab, setActiveTab] = useState<FilterTab>('All');
  const [search, setSearch] = useState('');
  const searchRef = useRef<HTMLInputElement>(null);
  const { data: fetchedStaff, isLoading: loading } = useStaff();
  const staffData: StaffMember[] = fetchedStaff || [];

  useEffect(() => {
    const handler = () => searchRef.current?.focus();
    window.addEventListener('hr:focus-search', handler);
    return () => window.removeEventListener('hr:focus-search', handler);
  }, []);

  const filtered = staffData.filter((s) => {
    const matchesTab = activeTab === 'All' || s.employmentType === activeTab;
    const matchesSearch =
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.role.toLowerCase().includes(search.toLowerCase());
    return matchesTab && matchesSearch;
  });

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.staff'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'space-between',
            gap: 12,
            flexWrap: 'wrap',
          }}
        >
          <div>
            <h2 style={{ margin: 0, fontSize: '18px', fontWeight: 700, color: 'var(--hr-text)' }}>
              {t('staff.directory')}
            </h2>
            <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--hr-text-muted)' }}>
              {loading ? '…' : t('staff.memberCount', { count: staffData.length })}
            </p>
          </div>
          <button
            type="button"
            onClick={() => navigate({ to: '/employees/new' })}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '9px 16px',
              borderRadius: '8px',
              background: ACCENT,
              border: 'none',
              color: '#fff',
              fontSize: '13px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <UserPlus size={15} />
            {t('staff.inviteStaff')}
          </button>
        </div>

        {/* Filters + search */}
        <div
          style={{
            display: 'flex',
            alignItems: isMobile ? 'stretch' : 'center',
            flexDirection: isMobile ? 'column' : 'row',
            gap: '12px',
          }}
        >
          {/* Tabs */}
          <div
            style={{
              display: 'flex',
              gap: '4px',
              padding: '4px',
              background: 'var(--hr-surface)',
              borderRadius: '10px',
              border: '1px solid var(--hr-border)',
              overflowX: 'auto',
              width: isMobile ? '100%' : 'auto',
            }}
          >
            {filterTabs.map((tab) => (
              <button
                type="button"
                key={tab}
                onClick={() => setActiveTab(tab)}
                style={{
                  padding: '6px 14px',
                  borderRadius: '7px',
                  border: 'none',
                  cursor: 'pointer',
                  fontSize: '12px',
                  fontWeight: 600,
                  background: activeTab === tab ? ACCENT : 'transparent',
                  color: activeTab === tab ? '#fff' : 'var(--hr-text-muted)',
                  transition: 'all 0.15s',
                }}
              >
                {tab}
              </button>
            ))}
          </div>

          {/* Search */}
          <div
            style={{
              position: 'relative',
              flex: 1,
              width: isMobile ? '100%' : undefined,
              maxWidth: isMobile ? '100%' : '280px',
            }}
          >
            <Search
              size={13}
              style={{
                position: 'absolute',
                left: '10px',
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'var(--hr-text-dim)',
              }}
            />
            <input
              ref={searchRef}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('staff.searchPlaceholder')}
              style={{
                width: '100%',
                padding: '8px 12px 8px 30px',
                borderRadius: '8px',
                border: '1px solid var(--hr-border)',
                background: 'var(--hr-input-bg)',
                color: 'var(--hr-text)',
                fontSize: '13px',
                outline: 'none',
              }}
            />
          </div>
        </div>

        {/* Staff cards grid */}
        <motion.div
          variants={container}
          initial="hidden"
          animate="show"
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
            gap: '14px',
          }}
        >
          {filtered.map((s, i) => {
            const av = avatarColors[i % avatarColors.length];
            return (
              <motion.div key={s.id} variants={cardAnim} whileHover={{ scale: 1.015 }}>
                <Link
                  to="/staff/$id"
                  params={{ id: String(s.id) }}
                  style={{ textDecoration: 'none', display: 'block' }}
                >
                  <div
                    style={{
                      background: 'var(--hr-card)',
                      border: '1px solid var(--hr-border)',
                      borderRadius: '12px',
                      padding: '18px',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '14px',
                      cursor: 'pointer',
                      transition: 'border-color 0.15s',
                    }}
                  >
                    {/* Top row */}
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'flex-start',
                        justifyContent: 'space-between',
                      }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            height: 42,
                            width: 42,
                            borderRadius: '50%',
                            background: av.bg,
                            color: av.color,
                            fontSize: '14px',
                            fontWeight: 700,
                            flexShrink: 0,
                          }}
                        >
                          {getInitials(s.name)}
                        </span>
                        <div>
                          <p
                            style={{
                              margin: 0,
                              fontSize: '14px',
                              fontWeight: 700,
                              color: 'var(--hr-text)',
                            }}
                          >
                            {s.name}
                          </p>
                          <p
                            style={{
                              margin: '2px 0 0',
                              fontSize: '12px',
                              color: 'var(--hr-text-muted)',
                            }}
                          >
                            {s.role}
                          </p>
                        </div>
                      </div>
                      {/* Status badge */}
                      <span
                        style={{
                          padding: '3px 9px',
                          borderRadius: '20px',
                          fontSize: '11px',
                          fontWeight: 600,
                          background:
                            s.status === 'Active'
                              ? 'rgba(16,185,129,0.12)'
                              : 'rgba(239,68,68,0.12)',
                          color: s.status === 'Active' ? '#10b981' : '#ef4444',
                        }}
                      >
                        {s.status}
                      </span>
                    </div>

                    {/* Meta */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <Mail size={12} style={{ color: 'var(--hr-text-dim)', flexShrink: 0 }} />
                        <span style={{ fontSize: '12px', color: 'var(--hr-text-muted)' }}>
                          {s.email}
                        </span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <span style={{ fontSize: '11px', color: 'var(--hr-text-dim)' }}>
                          {t('staff.deptPrefix')}
                        </span>
                        <span
                          style={{
                            fontSize: '12px',
                            color: 'var(--hr-text-muted)',
                            fontWeight: 500,
                          }}
                        >
                          {s.department}
                        </span>
                        <span
                          style={{
                            marginLeft: 'auto',
                            padding: '2px 8px',
                            borderRadius: '20px',
                            fontSize: '10px',
                            fontWeight: 600,
                            background: 'var(--hr-hover)',
                            color: 'var(--hr-text-muted)',
                            border: '1px solid var(--hr-border)',
                          }}
                        >
                          {s.employmentType}
                        </span>
                      </div>
                    </div>
                  </div>
                </Link>
              </motion.div>
            );
          })}
        </motion.div>

        {filtered.length === 0 && !loading && (
          <EmptyState
            icon={Users}
            title={t('staff.emptyState.title')}
            description={t('staff.emptyState.description')}
            action={
              search || activeTab !== 'All'
                ? {
                    label: t('common.clearFilters'),
                    onClick: () => {
                      setSearch('');
                      setActiveTab('All');
                    },
                  }
                : {
                    label: t('staff.inviteStaff'),
                    onClick: () => navigate({ to: '/employees/new' }),
                  }
            }
          />
        )}
      </div>
    </>
  );
};

export default Staff;
