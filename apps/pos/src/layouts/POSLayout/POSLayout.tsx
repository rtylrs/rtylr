import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Monitor,
  ClipboardList,
  LayoutGrid,
  LogOut,
  Clock,
  ChevronDown,
  ExternalLink,
  Utensils,
} from 'lucide-react';
import { usePOSStore, TABLE_OPTIONS } from '../../stores/posStore';
import { cn } from '../../lib/cn';

interface NavLink {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const NAV_LINKS: NavLink[] = [
  { label: 'Terminal', path: '/',       icon: Monitor      },
  { label: 'Orders',   path: '/orders', icon: ClipboardList },
  { label: 'Tables',   path: '/tables', icon: LayoutGrid   },
];

function useSessionClock(start: Date): string {
  const [, rerender] = React.useState(0);
  React.useEffect(() => {
    const id = setInterval(() => rerender((n) => n + 1), 1000);
    return () => clearInterval(id);
  }, []);
  const elapsed = Math.floor((Date.now() - start.getTime()) / 1000);
  const h = Math.floor(elapsed / 3600);
  const m = Math.floor((elapsed % 3600) / 60);
  const s = elapsed % 60;
  const pad = (n: number) => String(n).padStart(2, '0');
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}

interface POSLayoutProps {
  children?: React.ReactNode;
}

const POSLayout: React.FC<POSLayoutProps> = ({ children }) => {
  const location = useLocation();
  const { activeTableId, setActiveTableId, staffName, sessionStart } = usePOSStore();
  const sessionTime = useSessionClock(sessionStart);
  const [tableOpen, setTableOpen] = useState(false);

  const activeTable = TABLE_OPTIONS.find((t) => t.id === activeTableId);

  return (
    <div
      className="flex flex-col h-screen overflow-hidden"
      style={{ background: 'var(--hub-bg)' }}
    >
      {/* ── Top Bar ──────────────────────────────────────────── */}
      <header
        className="h-12 shrink-0 flex items-center justify-between px-4 gap-4"
        style={{
          background: 'var(--hub-surface)',
          borderBottom: '1px solid var(--hub-border)',
        }}
      >
        {/* Left: logo + nav */}
        <div className="flex items-center gap-1 shrink-0">
          {/* Logo mark */}
          <div className="flex items-center gap-2 pr-3 mr-1" style={{ borderRight: '1px solid var(--hub-border)' }}>
            <span
              className="flex items-center justify-center h-7 w-7 rounded-lg"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <Utensils size={14} style={{ color: '#4583C4' }} />
            </span>
            <span className="text-sm font-bold tracking-tight" style={{ color: 'var(--hub-text)' }}>
              rtylr
              <span
                className="ml-1 text-[10px] font-semibold px-1.5 py-0.5 rounded"
                style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
              >
                POS
              </span>
            </span>
          </div>

          {/* Nav links */}
          {NAV_LINKS.map(({ label, path, icon: Icon }) => {
            const isActive =
              path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);
            return (
              <Link
                key={path}
                to={path}
                className={cn(
                  'relative flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors duration-150',
                  isActive ? '' : 'hover:bg-[var(--hub-hover)]'
                )}
                style={{ color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)' }}
              >
                {isActive && (
                  <motion.span
                    layoutId="topnav-active"
                    className="absolute inset-0 rounded-lg"
                    style={{ background: 'var(--hub-active)' }}
                    transition={{ type: 'spring', stiffness: 400, damping: 35 }}
                  />
                )}
                <Icon size={13} className="relative shrink-0" />
                <span className="relative">{label}</span>
                {isActive && (
                  <span
                    className="relative ml-0.5 h-1 w-1 rounded-full"
                    style={{ background: '#4583C4' }}
                  />
                )}
              </Link>
            );
          })}
        </div>

        {/* Center: Table selector */}
        <div className="flex-1 flex justify-center">
          <div className="relative">
            <button
              onClick={() => setTableOpen((v) => !v)}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-semibold transition-colors hover:bg-[var(--hub-hover)]"
              style={{
                color: 'var(--hub-text)',
                border: '1px solid var(--hub-border)',
                background: 'var(--hub-card)',
                minWidth: '160px',
                justifyContent: 'space-between',
              }}
            >
              <span style={{ color: 'var(--hub-text-muted)', fontSize: '11px', fontWeight: 500 }}>
                Active:
              </span>
              <span>{activeTable?.label ?? 'Takeaway'}</span>
              <ChevronDown
                size={13}
                style={{
                  color: 'var(--hub-text-muted)',
                  transform: tableOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                  transition: 'transform 0.15s',
                }}
              />
            </button>

            <AnimatePresence>
              {tableOpen && (
                <motion.div
                  initial={{ opacity: 0, y: -4, scale: 0.97 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: -4, scale: 0.97 }}
                  transition={{ duration: 0.12 }}
                  className="absolute top-full mt-1 left-0 right-0 rounded-xl overflow-hidden z-50"
                  style={{
                    background: 'var(--hub-surface)',
                    border: '1px solid var(--hub-border)',
                    boxShadow: 'var(--hub-shadow)',
                  }}
                >
                  {TABLE_OPTIONS.map((table) => (
                    <button
                      key={table.id}
                      onClick={() => {
                        setActiveTableId(table.id);
                        setTableOpen(false);
                      }}
                      className="w-full text-left px-3 py-2 text-sm transition-colors hover:bg-[var(--hub-hover)]"
                      style={{
                        color: table.id === activeTableId ? '#4583C4' : 'var(--hub-text)',
                        background: table.id === activeTableId ? 'rgba(69,131,196,0.08)' : 'transparent',
                        fontWeight: table.id === activeTableId ? 600 : 400,
                      }}
                    >
                      {table.label}
                    </button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>

        {/* Right: staff + session + actions */}
        <div className="flex items-center gap-2 shrink-0">
          {/* Session clock */}
          <div
            className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <Clock size={11} style={{ color: 'var(--hub-text-muted)' }} />
            <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
              {sessionTime}
            </span>
          </div>

          {/* Staff name */}
          <div
            className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <span
              className="flex items-center justify-center h-4 w-4 rounded-full text-[9px] font-bold shrink-0"
              style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
            >
              {staffName.charAt(0)}
            </span>
            <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {staffName}
            </span>
          </div>

          {/* Hub link */}
          <a
            href="http://localhost:5175"
            className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium transition-colors hover:bg-[var(--hub-hover)]"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <ExternalLink size={11} />
            Hub
          </a>

          {/* End session */}
          <Link
            to="/close"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors hover:opacity-90"
            style={{ background: 'rgba(239,68,68,0.12)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
          >
            <LogOut size={12} />
            End Session
          </Link>
        </div>
      </header>

      {/* ── Content area ─────────────────────────────────────── */}
      <AnimatePresence mode="wait">
        <motion.div
          key={location.pathname}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="flex-1 overflow-hidden"
        >
          {children}
        </motion.div>
      </AnimatePresence>
    </div>
  );
};

export default POSLayout;
