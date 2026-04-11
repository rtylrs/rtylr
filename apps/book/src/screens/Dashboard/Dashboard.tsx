import React from 'react';
import { motion } from 'framer-motion';
import { Calendar, Clock, XCircle, BarChart3, Plus } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useDashboardStats, useTodayAppointments } from '../../hooks/useDashboard';

const INDIGO = '#6366f1';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

function getStatusColor(status: string): string {
  if (status === 'confirmed') return '#10b981';
  if (status === 'pending') return '#f59e0b';
  if (status === 'cancelled') return '#ef4444';
  if (status === 'completed') return '#6366f1';
  return 'var(--hub-text-dim)';
}

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const Dashboard: React.FC = () => {
  const { data: stats, isLoading: statsLoading, isError: statsError } = useDashboardStats();
  const { data: todayAppts, isLoading: todayLoading, isError: todayError } = useTodayAppointments();

  const isLoading = statsLoading || todayLoading;

  const statCards = [
    {
      label: "Today's Appointments",
      value: stats?.todayCount ?? 0,
      icon: Calendar,
      change: stats?.todayChange ?? '',
      to: '/appointments',
    },
    {
      label: 'Upcoming',
      value: stats?.upcomingCount ?? 0,
      icon: Clock,
      change: stats?.upcomingChange ?? '',
      to: '/calendar',
    },
    {
      label: 'Cancellation Rate',
      value: stats?.cancellationRate != null ? `${stats.cancellationRate}%` : '0%',
      icon: XCircle,
      change: stats?.cancellationChange ?? '',
      to: '/appointments',
    },
    {
      label: 'Utilization',
      value: stats?.utilization != null ? `${stats.utilization}%` : '0%',
      icon: BarChart3,
      change: stats?.utilizationChange ?? '',
      to: '/calendar',
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants} className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Booking Overview
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {today}
          </p>
        </div>
        <Link
          to="/appointments"
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: INDIGO, color: '#fff', textDecoration: 'none' }}
        >
          <Plus size={14} />
          New Appointment
        </Link>
      </motion.div>

      {statsError && (
        <div
          className="rounded-xl p-4 text-sm"
          style={{ background: 'rgba(239,68,68,0.08)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
        >
          Failed to load dashboard stats. Please try again later.
        </div>
      )}

      <motion.div variants={containerVariants} className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {statCards.map((stat) => {
          const Icon = stat.icon;
          return (
            <motion.div key={stat.label} variants={itemVariants}>
              <Link
                to={stat.to}
                className="block rounded-xl p-4 transition-all"
                style={{
                  background: 'var(--hub-card)',
                  border: '1px solid var(--hub-border)',
                  textDecoration: 'none',
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'rgba(99,102,241,0.3)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)'; }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                    {stat.label}
                  </span>
                  <span
                    className="flex items-center justify-center h-7 w-7 rounded-lg"
                    style={{ background: 'rgba(99,102,241,0.1)' }}
                  >
                    <Icon size={14} style={{ color: INDIGO }} />
                  </span>
                </div>
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                  {isLoading ? '...' : stat.value}
                </p>
                {stat.change && (
                  <p className="text-xs mt-1" style={{ color: 'var(--hub-text-muted)' }}>
                    {stat.change}
                  </p>
                )}
              </Link>
            </motion.div>
          );
        })}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <Clock size={15} style={{ color: INDIGO }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Today's Schedule
          </h3>
        </div>

        {todayError && (
          <p className="text-sm py-4 text-center" style={{ color: '#ef4444' }}>
            Failed to load today's appointments.
          </p>
        )}

        {!todayError && isLoading && (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        )}

        {!todayError && !isLoading && (!todayAppts || todayAppts.length === 0) && (
          <div className="py-8 text-center">
            <Calendar size={32} className="mx-auto mb-2" style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No appointments scheduled for today
            </p>
          </div>
        )}

        {!todayError && !isLoading && todayAppts && todayAppts.length > 0 && (
          <div className="space-y-1">
            {todayAppts.map((appt: any) => {
              const statusColor = getStatusColor(appt.status);
              return (
                <div
                  key={appt.id}
                  className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {appt.time}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {appt.duration} min
                    </p>
                  </div>
                  <p className="text-xs truncate pr-2" style={{ color: 'var(--hub-text)' }}>
                    {appt.customerName}
                  </p>
                  <p className="text-xs truncate pr-2" style={{ color: 'var(--hub-text-muted)' }}>
                    {appt.serviceName}
                  </p>
                  <div className="flex justify-end">
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize"
                      style={{ background: `${statusColor}18`, color: statusColor }}
                    >
                      {appt.status}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Dashboard;
