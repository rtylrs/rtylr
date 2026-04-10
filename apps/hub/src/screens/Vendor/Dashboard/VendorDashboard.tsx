import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  DollarSign, ShoppingBag, Users, TrendingUp,
  UtensilsCrossed, QrCode, BarChart3, ArrowUpRight,
  Briefcase, Package, Activity as ActivityIcon, Gauge,
} from 'lucide-react';
import { useUserStore } from '../../../stores/userStore';
import {
  AppCard,
  OnboardingChecklist,
  StatCard,
  WelcomeBanner,
  ActivityEntryRow,
  EmptyState,
} from '../../../components/hub';
import { useApps } from '../../../hooks/queries/useApps';
import { useKpis } from '../../../hooks/queries/useKpis';
import { useActivity } from '../../../hooks/queries/useActivity';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  show:   { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

// ─── Staff view: app launcher ─────────────────────────────────────────────────

const StaffDashboard: React.FC<{ userApps: string[] }> = ({ userApps }) => {
  const { data: apps = [] } = useApps();
  const accessibleApps = apps.filter((a) => userApps.includes(a.key));
  const today = new Date().toLocaleDateString('en-US', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
  });

  return (
    <div className="space-y-6">
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.2 }}>
        <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Your apps</h1>
        <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
      </motion.div>

      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-3 gap-4">
        {accessibleApps.map((app) => (
          <motion.div key={app.key} variants={itemVariants}>
            <AppCard app={app} />
          </motion.div>
        ))}
      </motion.div>

      {accessibleApps.length === 0 && (
        <EmptyState
          icon={Briefcase}
          title="No apps assigned"
          description="Contact your admin to get access to apps."
        />
      )}
    </div>
  );
};

// ─── Admin view ───────────────────────────────────────────────────────────────

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
});

function formatMoney(n: number | undefined | null) {
  if (n == null) return '$0';
  return `$${n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
}

const quickActions = [
  { label: 'Add Menu Item', path: '/menus',     icon: UtensilsCrossed },
  { label: 'View QR Codes', path: '/qr-codes',  icon: QrCode          },
  { label: 'Analytics',     path: '/analytics', icon: BarChart3       },
];

const AdminDashboard: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useUserStore();
  const { data: kpis } = useKpis();
  const { data: apps = [] } = useApps();
  const { data: activityRaw = [] } = useActivity(5);
  const activity = Array.isArray(activityRaw) ? activityRaw : [];
  const businessName = (user as any)?.businessName || '';

  return (
    <div className="space-y-6 max-w-6xl">
      {/* Header */}
      <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.2 }}>
        <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>{businessName || 'Dashboard'}</h1>
        <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
      </motion.div>

      {/* KPI Stats */}
      <motion.div variants={containerVariants} initial="hidden" animate="show" className="grid grid-cols-4 gap-3">
        <motion.div variants={itemVariants}>
          <StatCard label="Revenue today" value={kpis ? formatMoney(kpis.revenueToday) : '$0'} icon={DollarSign} accent="#4583C4" trendPct={kpis?.revenueTrendPct} sub="vs yesterday" />
        </motion.div>
        <motion.div variants={itemVariants}>
          <StatCard label="Orders today" value={kpis?.ordersToday ?? 0} icon={ShoppingBag} accent="#8b5cf6" trendPct={kpis?.ordersTrendPct} sub="vs yesterday" />
        </motion.div>
        <motion.div variants={itemVariants}>
          <StatCard label="Team" value={`${kpis?.activeStaff ?? 0}/${kpis?.headcount ?? 0}`} icon={Users} accent="#f59e0b" sub="on shift" />
        </motion.div>
        <motion.div variants={itemVariants}>
          <StatCard label="Open invoices" value={kpis?.openInvoices ?? 0} icon={TrendingUp} accent="#10b981" sub="unpaid" />
        </motion.div>
      </motion.div>

      {/* Content grid */}
      <div className="grid grid-cols-3 gap-4">
        {/* Recent activity */}
        <motion.div
          initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.22, delay: 0.15 }}
          className="col-span-2 rounded-xl"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between px-5 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
            <div className="flex items-center gap-2">
              <ActivityIcon size={14} style={{ color: 'var(--hub-text-muted)' }} />
              <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Recent activity</p>
            </div>
            <button
              onClick={() => navigate('/activity')}
              className="flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-70"
              style={{ color: '#4583C4' }}
            >
              View all <ArrowUpRight size={12} />
            </button>
          </div>
          <div className="p-2">
            {activity.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 text-center">
                <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>No activity yet</p>
                <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>Actions will appear here as your team uses rtylr</p>
              </div>
            ) : (
              activity.map((e) => <ActivityEntryRow key={e.id} entry={e} />)
            )}
          </div>
        </motion.div>

        {/* Onboarding + quick actions */}
        <motion.div
          initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.22, delay: 0.2 }}
          className="space-y-4"
        >
          <div className="rounded-xl" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
            <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
              <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>What's new</p>
            </div>
            <div className="p-4 space-y-2.5">
              {[
                { version: '0.2.0', date: 'Apr 10', text: 'Addon system, team invites with email, request access flow' },
                { version: '0.1.5', date: 'Apr 9', text: 'All 75 backend modules migrated to sqlc, business/org model' },
                { version: '0.1.0', date: 'Apr 9', text: 'Full platform launch: POS, ERP, HR, CRM, Finance, Flow' },
              ].map((entry) => (
                <div key={entry.version} className="flex items-start gap-3">
                  <span className="text-[10px] font-mono font-semibold px-1.5 py-0.5 rounded shrink-0 mt-0.5" style={{ background: 'rgba(69,131,196,0.1)', color: '#4583C4' }}>{entry.version}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs leading-relaxed" style={{ color: 'var(--hub-text)' }}>{entry.text}</p>
                    <p className="text-[10px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{entry.date}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-xl" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
            <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
              <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Quick actions</p>
            </div>
            <div className="p-3 space-y-1">
              {quickActions.map((action) => {
                const Icon = action.icon;
                return (
                  <button
                    key={action.label}
                    onClick={() => navigate(action.path)}
                    className="flex items-center gap-2.5 w-full px-3 py-2.5 rounded-lg text-sm font-medium transition-colors duration-150 text-left hover:bg-[var(--hub-hover)]"
                    style={{ color: 'var(--hub-text-muted)' }}
                  >
                    <Icon size={15} className="shrink-0" />
                    {action.label}
                  </button>
                );
              })}
            </div>
          </div>
        </motion.div>
      </div>

    </div>
  );
};

// ─── Root dashboard — role-aware ──────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user, role } = useUserStore();

  const isAdmin = role === 'admin' || role === 'owner' || !role;
  const userApps: string[] = (user as any)?.apps ?? [];

  if (!isAdmin) {
    return <StaffDashboard userApps={userApps} />;
  }

  return <AdminDashboard />;
};

export default VendorDashboard;
