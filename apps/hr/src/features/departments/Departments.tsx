import React, { useState } from 'react';
import { motion, type Variants } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import {
  Users,
  UserCheck,
  DollarSign,
  ChefHat,
  Coffee,
  Briefcase,
  Truck,
  Plus,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Department } from '../../shared/services/hrApi';
import { useDepartments } from './useDepartments';
import CreateDepartmentForm from './CreateDepartmentForm';
import { useIsMobile } from '../../shared/components/ui';
import { useUserStore } from '@voxire/auth';
import { canViewCompensation } from '../../shared/lib/hrPermissions';

const ACCENT = '#8b5cf6';

const DEPT_ICONS: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  Kitchen: ChefHat,
  'Front-of-House': Coffee,
  Management: Briefcase,
  Delivery: Truck,
};

const DEPT_COLORS = ['#f59e0b', '#10b981', ACCENT, '#3b82f6', '#ef4444', '#3b82f6'];

function getDeptIcon(name: string) {
  return DEPT_ICONS[name] || Briefcase;
}

function getDeptColor(index: number) {
  return DEPT_COLORS[index % DEPT_COLORS.length];
}

function fmt(n: number): string {
  return `$${n.toLocaleString('en-US')}`;
}

const container = { hidden: {}, show: { transition: { staggerChildren: 0.08 } } };
const cardAnim: Variants = {
  hidden: { opacity: 0, y: 14 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Departments: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const showCompensation = canViewCompensation(role);
  const { data: fetchedDepartments, isLoading: loading } = useDepartments(showCompensation);
  const departments: Department[] = fetchedDepartments || [];
  const [createOpen, setCreateOpen] = useState(false);

  return (
    <>
      <HelmetContainer title="Departments · rtylr" />
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
              Departments
            </h2>
            <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--hr-text-muted)' }}>
              {loading
                ? '…'
                : `${departments.length} departments · ${departments.reduce(
                    (s, d) => s + d.staffCount,
                    0
                  )} total staff`}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setCreateOpen(true)}
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
            <Plus size={15} />
            {t('departments.addNew')}
          </button>
        </div>

        {/* Department cards */}
        <motion.div
          variants={container}
          initial="hidden"
          animate="show"
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '16px',
          }}
        >
          {departments.map((dept, idx) => {
            const color = dept.color ?? getDeptColor(idx);
            const bg = `${color}18`;
            const DeptIcon = getDeptIcon(dept.name);
            return (
              <motion.div
                key={dept.id}
                variants={cardAnim}
                whileHover={{ scale: 1.012 }}
                style={{
                  background: 'var(--hr-card)',
                  border: '1px solid var(--hr-border)',
                  borderRadius: '14px',
                  overflow: 'hidden',
                  cursor: 'pointer',
                }}
              >
                {/* Card top band */}
                <div
                  style={{
                    padding: '20px 22px 16px',
                    background: bg,
                    borderBottom: '1px solid var(--hr-border)',
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
                        height: 44,
                        width: 44,
                        borderRadius: '12px',
                        background: `${color}22`,
                        color,
                        flexShrink: 0,
                      }}
                    >
                      <DeptIcon size={22} />
                    </span>
                    <div>
                      <h3
                        style={{
                          margin: 0,
                          fontSize: '16px',
                          fontWeight: 700,
                          color: 'var(--hr-text)',
                        }}
                      >
                        {dept.name}
                      </h3>
                      <p
                        style={{
                          margin: '3px 0 0',
                          fontSize: '12px',
                          color: 'var(--hr-text-muted)',
                          maxWidth: '240px',
                        }}
                      >
                        {dept.description}
                      </p>
                    </div>
                  </div>
                  {/* Staff count badge */}
                  <span
                    style={{
                      padding: '4px 12px',
                      borderRadius: '20px',
                      fontSize: '12px',
                      fontWeight: 700,
                      background: `${color}18`,
                      color,
                      flexShrink: 0,
                    }}
                  >
                    {dept.staffCount} staff
                  </span>
                </div>

                {/* Card body */}
                <div
                  style={{
                    padding: '16px 22px',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '12px',
                  }}
                >
                  {/* Department head */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <UserCheck size={14} style={{ color: 'var(--hr-text-dim)', flexShrink: 0 }} />
                    <span style={{ fontSize: '12px', color: 'var(--hr-text-dim)' }}>
                      Department Head
                    </span>
                    <span
                      style={{
                        marginLeft: 'auto',
                        fontSize: '13px',
                        fontWeight: 600,
                        color: 'var(--hr-text)',
                      }}
                    >
                      {dept.head}
                    </span>
                  </div>

                  <div style={{ height: '1px', background: 'var(--hr-border)' }} />

                  {/* Stats row */}
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr',
                      gap: '12px',
                    }}
                  >
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <Users size={13} style={{ color: 'var(--hr-text-dim)' }} />
                        <span style={{ fontSize: '11px', color: 'var(--hr-text-dim)' }}>
                          Team Size
                        </span>
                      </div>
                      <span style={{ fontSize: '18px', fontWeight: 700, color }}>
                        {dept.staffCount}
                      </span>
                    </div>
                    {showCompensation && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <DollarSign size={13} style={{ color: 'var(--hr-text-dim)' }} />
                          <span style={{ fontSize: '11px', color: 'var(--hr-text-dim)' }}>
                            Avg Salary
                          </span>
                        </div>
                        <span
                          style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hr-text)' }}
                        >
                          {fmt(dept.avgSalary ?? 0)}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Headcount bar */}
                  <div>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        marginBottom: '6px',
                      }}
                    >
                      <span style={{ fontSize: '11px', color: 'var(--hr-text-dim)' }}>
                        Share of total workforce
                      </span>
                      <span style={{ fontSize: '11px', fontWeight: 600, color }}>
                        {departments.reduce((s, d) => s + d.staffCount, 0) > 0
                          ? `${Math.round(
                              (dept.staffCount /
                                departments.reduce((s, d) => s + d.staffCount, 0)) *
                                100
                            )}%`
                          : '0%'}
                      </span>
                    </div>
                    <div
                      style={{
                        height: 6,
                        borderRadius: 9999,
                        background: 'var(--hr-hover)',
                        overflow: 'hidden',
                      }}
                    >
                      <motion.div
                        initial={{ width: 0 }}
                        animate={{
                          width:
                            departments.reduce((s, d) => s + d.staffCount, 0) > 0
                              ? `${
                                  (dept.staffCount /
                                    departments.reduce((s, d) => s + d.staffCount, 0)) *
                                  100
                                }%`
                              : '0%',
                        }}
                        transition={{ duration: 0.6, ease: 'easeOut', delay: 0.2 }}
                        style={{ height: '100%', background: color, borderRadius: 9999 }}
                      />
                    </div>
                  </div>
                </div>
              </motion.div>
            );
          })}
        </motion.div>
      </div>

      <CreateDepartmentForm open={createOpen} onClose={() => setCreateOpen(false)} />
    </>
  );
};

export default Departments;
