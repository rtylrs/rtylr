import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  DollarSign, ShoppingBag, Users, TrendingUp,
  UtensilsCrossed, QrCode, BarChart3, ArrowUpRight,
  MonitorSpeaker, Package, UserCheck, CreditCard, Briefcase, Layers,
  ExternalLink,
} from 'lucide-react';
import { useUserStore } from '../../../stores/userStore';

// ─── App launcher (for non-admin staff) ───────────────────────────────────────

const APPS = [
  {
    key: 'pos',
    label: 'Point of Sale',
    description: 'Floor & counter sales, table management, kitchen display.',
    icon: MonitorSpeaker,
    color: '#4583C4',
    bg: 'rgba(69,131,196,0.1)',
    href: 'https://pos.rtylr.com',
  },
  {
    key: 'erp',
    label: 'ERP',
    description: 'Inventory, purchasing, suppliers, cost of goods.',
    icon: Package,
    color: '#10b981',
    bg: 'rgba(16,185,129,0.1)',
    href: 'https://erp.rtylr.com',
  },
  {
    key: 'hr',
    label: 'HR',
    description: 'Staff, shifts, payroll, timesheets.',
    icon: UserCheck,
    color: '#f59e0b',
    bg: 'rgba(245,158,11,0.1)',
    href: 'https://hr.rtylr.com',
  },
  {
    key: 'crm',
    label: 'CRM',
    description: 'Customer profiles, loyalty programs, campaigns.',
    icon: Users,
    color: '#8b5cf6',
    bg: 'rgba(139,92,246,0.1)',
    href: 'https://crm.rtylr.com',
  },
  {
    key: 'finance',
    label: 'Finance',
    description: 'Invoicing, expenses, P&L, tax reporting.',
    icon: CreditCard,
    color: '#ef4444',
    bg: 'rgba(239,68,68,0.1)',
    href: 'https://finance.rtylr.com',
  },
  {
    key: 'tracker',
    label: 'Tracker',
    description: 'Projects, tasks, subtasks, team workspaces.',
    icon: Layers,
    color: '#64C5B9',
    bg: 'rgba(100,197,185,0.1)',
    href: 'https://tracker.rtylr.com',
  },
];

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  show:   { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

// ─── Staff view: app launcher ─────────────────────────────────────────────────

const StaffDashboard: React.FC<{ userApps: string[] }> = ({ userApps }) => {
  const accessibleApps = APPS.filter((a) => userApps.includes(a.key));
  const today = new Date().toLocaleDateString('en-US', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
  });

  return (
    <div className="space-y-6">
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.2 }}>
        <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Your apps</h1>
        <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
      </motion.div>

      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="show"
        className="grid grid-cols-3 gap-4"
      >
        {accessibleApps.map((app) => {
          const Icon = app.icon;
          return (
            <motion.a
              key={app.key}
              variants={itemVariants}
              href={app.href}
              target="_blank"
              rel="noopener noreferrer"
              className="group rounded-xl p-5 flex flex-col gap-4 cursor-pointer transition-all duration-150 no-underline"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
              }}
              whileHover={{ scale: 1.01 }}
            >
              <div className="flex items-start justify-between">
                <div
                  className="h-11 w-11 rounded-xl flex items-center justify-center"
                  style={{ background: app.bg }}
                >
                  <Icon size={22} style={{ color: app.color }} />
                </div>
                <ExternalLink
                  size={13}
                  className="opacity-0 group-hover:opacity-100 transition-opacity"
                  style={{ color: 'var(--hub-text-dim)' }}
                />
              </div>
              <div>
                <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{app.label}</p>
                <p className="text-xs mt-1 leading-relaxed" style={{ color: 'var(--hub-text-muted)' }}>
                  {app.description}
                </p>
              </div>
            </motion.a>
          );
        })}
      </motion.div>

      {accessibleApps.length === 0 && (
        <div
          className="rounded-xl p-10 flex flex-col items-center justify-center text-center"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <Briefcase size={28} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>No apps assigned</p>
          <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
            Contact your admin to get access to apps.
          </p>
        </div>
      )}
    </div>
  );
};

// ─── Admin view: analytics + app overview ─────────────────────────────────────

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
});

const stats = [
  { label: "Today's Revenue",  value: '$0.00', icon: DollarSign,  sub: 'vs yesterday',  accent: '#4583C4' },
  { label: 'Orders Today',     value: '0',     icon: ShoppingBag, sub: 'no orders yet', accent: '#8b5cf6' },
  { label: 'Active Tables',    value: '0',     icon: Users,       sub: '0 occupied',    accent: '#f59e0b' },
  { label: 'Avg. Order Value', value: '$0.00', icon: TrendingUp,  sub: 'no data',       accent: '#10b981' },
];

const quickActions = [
  { label: 'Add Menu Item', path: '/menus',     icon: UtensilsCrossed },
  { label: 'View QR Codes', path: '/qr-codes',  icon: QrCode          },
  { label: 'Analytics',     path: '/analytics', icon: BarChart3       },
];

const orderStatus = [
  { label: 'Pending',     count: 0, color: '#f59e0b' },
  { label: 'In Progress', count: 0, color: '#4583C4' },
  { label: 'Completed',   count: 0, color: '#10b981' },
];

const AdminDashboard: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
      >
        <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Good morning</h1>
        <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
      </motion.div>

      {/* Stats */}
      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-4 gap-3">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <motion.div
              key={stat.label}
              variants={itemVariants}
              className="rounded-xl p-4"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <div className="flex items-center justify-between mb-3">
                <p className="text-[11px] font-medium uppercase tracking-wider" style={{ color: 'var(--hub-text-dim)' }}>
                  {stat.label}
                </p>
                <div className="h-7 w-7 rounded-lg flex items-center justify-center" style={{ background: `${stat.accent}18` }}>
                  <Icon size={14} style={{ color: stat.accent }} />
                </div>
              </div>
              <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>{stat.value}</p>
              <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>{stat.sub}</p>
            </motion.div>
          );
        })}
      </motion.div>

      {/* Content grid */}
      <div className="grid grid-cols-3 gap-4">
        {/* Recent Orders */}
        <motion.div
          initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.22, delay: 0.15 }}
          className="col-span-2 rounded-xl"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between px-5 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
            <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Recent Orders</p>
            <button
              onClick={() => navigate('/orders')}
              className="flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-70"
              style={{ color: '#4583C4' }}
            >
              View all <ArrowUpRight size={12} />
            </button>
          </div>
          <div className="flex flex-col items-center justify-center py-14 text-center">
            <div className="h-10 w-10 rounded-xl flex items-center justify-center mb-3" style={{ background: 'var(--hub-hover)' }}>
              <ShoppingBag size={18} style={{ color: 'var(--hub-text-dim)' }} />
            </div>
            <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>No orders yet</p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>Orders appear once placed by customers</p>
          </div>
        </motion.div>

        {/* Quick Actions */}
        <motion.div
          initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.22, delay: 0.2 }}
          className="rounded-xl"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
            <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Quick Actions</p>
          </div>
          <div className="p-3 space-y-1">
            {quickActions.map((action) => {
              const Icon = action.icon;
              return (
                <button
                  key={action.label}
                  onClick={() => navigate(action.path)}
                  className="flex items-center gap-2.5 w-full px-3 py-2.5 rounded-lg text-sm font-medium transition-colors duration-150 text-left"
                  style={{ color: 'var(--hub-text-muted)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    (e.currentTarget as HTMLElement).style.color = 'var(--hub-text)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                    (e.currentTarget as HTMLElement).style.color = 'var(--hub-text-muted)';
                  }}
                >
                  <Icon size={15} className="shrink-0" />
                  {action.label}
                </button>
              );
            })}
          </div>

          {/* Apps quick-jump */}
          <div className="px-5 pt-1 pb-4" style={{ borderTop: '1px solid var(--hub-border)' }}>
            <p className="text-[10px] font-semibold uppercase tracking-widest mb-2 pt-3" style={{ color: 'var(--hub-text-dim)' }}>
              Your Apps
            </p>
            <div className="flex flex-wrap gap-1.5">
              {APPS.map((app) => {
                const Icon = app.icon;
                return (
                  <a
                    key={app.key}
                    href={app.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1.5 px-2 py-1 rounded-md text-[11px] font-medium transition-all no-underline hover:opacity-80"
                    style={{ background: app.bg, color: app.color }}
                  >
                    <Icon size={11} />
                    {app.label}
                  </a>
                );
              })}
            </div>
          </div>
        </motion.div>
      </div>

      {/* Order status */}
      <motion.div
        initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.22, delay: 0.25 }}
        className="grid grid-cols-3 gap-3"
      >
        {orderStatus.map((status) => (
          <div
            key={status.label}
            className="rounded-xl p-4 flex items-center justify-between"
            style={{ background: 'var(--hub-card)', border: `1px solid ${status.color}25` }}
          >
            <div className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full shrink-0" style={{ background: status.color }} />
              <span className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>{status.label}</span>
            </div>
            <span className="text-2xl font-bold" style={{ color: status.color }}>{status.count}</span>
          </div>
        ))}
      </motion.div>
    </div>
  );
};

// ─── Root dashboard — role-aware ──────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user, role } = useUserStore();

  const isAdmin = role === 'admin' || role === 'owner' || !role;
  // user.apps = ['pos', 'finance'] — set by backend on login
  const userApps: string[] = user?.apps ?? [];

  if (!isAdmin) {
    return <StaffDashboard userApps={userApps} />;
  }

  return <AdminDashboard />;
};

export default VendorDashboard;
