import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Inbox,
  LayoutGrid,
  Plus,
  ChevronDown,
  ChevronRight,
  Settings,
  LogOut,
  Sun,
  Moon,
  Monitor,
  ArrowLeft,
  Circle,
  ListTodo,
  Calendar,
  Timer,
  Milestone,
  Map,
  BarChart3,
  Users,
  Eye,
  Archive,
  Search,
  Command,
} from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { useUserStore } from '../../stores/userStore';
import { useThemeStore } from '../../stores/themeStore';
import { useFlowStore } from '../../stores/flowStore';
import { cn, getInitials } from '../../lib/utils';
import CommandBar from '../../components/CommandBar';
import QuickCreate from '../../components/QuickCreate';
import BulkEditBar from '../../components/BulkEditBar';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { useWorkspaces } from '../../hooks/useWorkspaces';
import { useProjects } from '../../hooks/useProjects';

const ACCENT = '#64C5B9';

interface SidebarLinkProps {
  to: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  label: string;
  exact?: boolean;
  badge?: number;
}

const SidebarLink: React.FC<SidebarLinkProps> = ({
  to,
  icon: Icon,
  label,
  exact,
  badge,
}) => {
  const location = useLocation();
  const isActive = exact
    ? location.pathname === to
    : location.pathname.startsWith(to);

  return (
    <Link
      to={to}
      className={cn(
        'relative flex items-center gap-2.5 px-3 py-1.5 rounded-md text-[12px] font-medium transition-colors duration-150',
        isActive ? '' : 'hover:bg-[var(--flow-hover)]'
      )}
      style={{ color: isActive ? 'var(--flow-text)' : 'var(--flow-text-muted)' }}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active-flow"
          className="absolute inset-0 rounded-md"
          style={{ background: 'var(--flow-active)' }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={14} className="relative shrink-0" />
      <span className="relative flex-1 truncate">{label}</span>
      {badge !== undefined && badge > 0 && (
        <span
          className="relative text-[10px] px-1.5 rounded-full font-semibold"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {badge}
        </span>
      )}
      {isActive && (
        <span
          className="relative h-1.5 w-1.5 rounded-full shrink-0"
          style={{ background: ACCENT }}
        />
      )}
    </Link>
  );
};

interface FlowLayoutProps {
  children: React.ReactNode;
}

const FlowLayout: React.FC<FlowLayoutProps> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { workspaceId } = useParams<{ workspaceId: string }>();

  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const {
    workspaces,
    projects,
    activeWorkspaceId,
    setActiveWorkspace,
    setWorkspaces,
    setProjects,
    notifications,
    tasks,
    setCommandBarOpen,
  } = useFlowStore();

  // Fetch workspaces and projects from API and populate store
  const { data: apiWorkspaces } = useWorkspaces();
  const currentWorkspaceId = workspaceId || activeWorkspaceId || workspaces[0]?.id || '';
  const { data: apiProjects } = useProjects(currentWorkspaceId || undefined);

  useEffect(() => {
    if (apiWorkspaces && apiWorkspaces.length > 0) {
      setWorkspaces(apiWorkspaces as any);
    }
  }, [apiWorkspaces, setWorkspaces]);

  useEffect(() => {
    if (apiProjects && apiProjects.length > 0) {
      setProjects(apiProjects as any);
    }
  }, [apiProjects, setProjects]);

  const [workspaceSwitcherOpen, setWorkspaceSwitcherOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  useKeyboardShortcuts();

  const activeWorkspace =
    workspaces.find((w) => w.id === currentWorkspaceId) || workspaces[0] || { id: '', name: 'Loading...', color: '#6b7280' };
  const workspaceProjects = projects.filter(
    (p) => p.workspaceId === currentWorkspaceId && !p.archived
  );

  const myIssuesCount = tasks.filter(
    (t) => t.assigneeId === 'u1' && t.status !== 'done' && t.status !== 'cancelled' && !t.archived
  ).length;
  const unreadNotifications = notifications.filter((n) => !n.read).length;

  const userName = user?.name || 'User';
  const userEmail = user?.email || '';

  const handleLogout = () => {
    logout();
    window.location.href = APP_URLS.auth;
  };

  const isMac = typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform);
  const modKey = isMac ? '⌘' : 'Ctrl';

  return (
    <div
      className="flex h-screen overflow-hidden"
      style={{ background: 'var(--flow-bg)' }}
    >
      <aside
        className="w-60 flex flex-col shrink-0 overflow-hidden"
        style={{
          background: 'var(--flow-surface)',
          borderRight: '1px solid var(--flow-border)',
        }}
      >
        {/* Top logo */}
        <div
          className="flex items-center justify-between px-4 py-3.5"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="flex items-center gap-2">
            <img src="/logo.png" alt="rtylr" className="h-6 w-auto" />
          </div>
          <a
            href={APP_URLS.hub}
            className="flex items-center gap-1 text-[11px] px-1.5 py-1 rounded transition-colors hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title="Back to Hub"
          >
            <ArrowLeft size={11} />
            Hub
          </a>
        </div>

        {/* Workspace Switcher */}
        <div className="px-2 pt-3 pb-1 relative">
          <button
            onClick={() => setWorkspaceSwitcherOpen((o) => !o)}
            className="w-full flex items-center gap-2 px-3 py-2 rounded-md transition-colors hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <span
              className="h-4 w-4 rounded shrink-0 flex items-center justify-center text-[9px] font-bold"
              style={{
                background: `${activeWorkspace.color}22`,
                color: activeWorkspace.color,
              }}
            >
              {activeWorkspace.name.charAt(0)}
            </span>
            <span
              className="flex-1 text-left text-[12px] font-medium truncate"
              style={{ color: 'var(--flow-text)' }}
            >
              {activeWorkspace.name}
            </span>
            <ChevronDown
              size={13}
              className={cn(
                'shrink-0 transition-transform',
                workspaceSwitcherOpen ? 'rotate-180' : ''
              )}
            />
          </button>

          <AnimatePresence>
            {workspaceSwitcherOpen && (
              <motion.div
                initial={{ opacity: 0, y: -4, scale: 0.97 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -4, scale: 0.97 }}
                transition={{ duration: 0.12 }}
                className="absolute left-2 right-2 z-50 mt-1 rounded-lg py-1 shadow-lg"
                style={{
                  background: 'var(--flow-card)',
                  border: '1px solid var(--flow-border)',
                  boxShadow: 'var(--flow-shadow)',
                }}
              >
                {workspaces.map((ws) => (
                  <button
                    key={ws.id}
                    onClick={() => {
                      setActiveWorkspace(ws.id);
                      navigate(`/w/${ws.id}`);
                      setWorkspaceSwitcherOpen(false);
                    }}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-left transition-colors hover:bg-[var(--flow-hover)]"
                    style={{
                      color:
                        ws.id === currentWorkspaceId
                          ? 'var(--flow-text)'
                          : 'var(--flow-text-muted)',
                    }}
                  >
                    <span
                      className="h-5 w-5 rounded shrink-0 flex items-center justify-center text-[10px] font-bold"
                      style={{
                        background: `${ws.color}22`,
                        color: ws.color,
                      }}
                    >
                      {ws.name.charAt(0)}
                    </span>
                    {ws.name}
                    {ws.id === currentWorkspaceId && (
                      <span
                        className="ml-auto h-1.5 w-1.5 rounded-full"
                        style={{ background: ACCENT }}
                      />
                    )}
                  </button>
                ))}
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Search trigger */}
        <div className="px-2 pt-1 pb-2">
          <button
            onClick={() => setCommandBarOpen(true)}
            className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md text-[11px] transition-colors hover:bg-[var(--flow-hover)]"
            style={{
              background: 'var(--flow-hover)',
              color: 'var(--flow-text-dim)',
            }}
          >
            <Search size={11} />
            <span className="flex-1 text-left">Search…</span>
            <span
              className="text-[9px] font-mono px-1.5 py-0.5 rounded"
              style={{ background: 'var(--flow-active)' }}
            >
              {modKey}K
            </span>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-2 pb-2">
          <nav className="space-y-0.5">
            <SidebarLink to="/inbox" icon={Inbox} label="Inbox" badge={unreadNotifications} />
            <SidebarLink to="/my-issues" icon={ListTodo} label="My Issues" badge={myIssuesCount} />
            <SidebarLink to={`/w/${currentWorkspaceId}`} icon={LayoutGrid} label="Dashboard" exact />
            <SidebarLink to="/triage" icon={Eye} label="Triage" />
          </nav>

          <div className="flex items-center justify-between px-3 mt-4 mb-1.5">
            <span
              className="text-[10px] font-semibold uppercase tracking-widest"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              Plan
            </span>
          </div>
          <nav className="space-y-0.5">
            <SidebarLink to="/roadmap" icon={Map} label="Roadmap" />
            <SidebarLink to="/cycles" icon={Timer} label="Cycles" />
            <SidebarLink to="/milestones" icon={Milestone} label="Milestones" />
            <SidebarLink to="/timeline" icon={Timer} label="Timeline" />
            <SidebarLink to="/calendar" icon={Calendar} label="Calendar" />
            <SidebarLink to="/views" icon={Eye} label="Views" />
          </nav>

          <div className="flex items-center justify-between px-3 mt-4 mb-1.5">
            <span
              className="text-[10px] font-semibold uppercase tracking-widest"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              Team
            </span>
          </div>
          <nav className="space-y-0.5">
            <SidebarLink to="/team" icon={Users} label="Workload" />
            <SidebarLink to="/insights" icon={BarChart3} label="Insights" />
            <SidebarLink to="/archive" icon={Archive} label="Archive" />
          </nav>

          <div className="flex items-center justify-between px-3 mt-4 mb-1.5">
            <span
              className="text-[10px] font-semibold uppercase tracking-widest"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              Projects
            </span>
            <button
              className="h-4 w-4 rounded flex items-center justify-center transition-colors hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-dim)' }}
              title="New Project"
            >
              <Plus size={12} />
            </button>
          </div>
          <div className="space-y-0.5">
            {workspaceProjects.map((project) => {
              const boardPath = `/w/${currentWorkspaceId}/p/${project.id}`;
              const isActive = location.pathname.startsWith(boardPath);
              return (
                <Link
                  key={project.id}
                  to={boardPath}
                  className={cn(
                    'relative flex items-center gap-2.5 px-3 py-1.5 rounded-md text-[12px] transition-colors duration-150',
                    isActive ? '' : 'hover:bg-[var(--flow-hover)]'
                  )}
                  style={{
                    color: isActive
                      ? 'var(--flow-text)'
                      : 'var(--flow-text-muted)',
                  }}
                >
                  {isActive && (
                    <motion.span
                      layoutId={`project-active-${project.id}`}
                      className="absolute inset-0 rounded-md"
                      style={{ background: 'var(--flow-active)' }}
                      transition={{ type: 'spring', stiffness: 400, damping: 35 }}
                    />
                  )}
                  <Circle
                    size={8}
                    className="relative shrink-0 fill-current"
                    style={{ color: project.color }}
                  />
                  <span className="relative flex-1 truncate font-medium">
                    {project.name}
                  </span>
                  <span
                    className="relative text-[10px]"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {project.taskCount - project.completedCount}
                  </span>
                </Link>
              );
            })}
          </div>
        </div>

        {/* Footer */}
        <div className="p-2" style={{ borderTop: '1px solid var(--flow-border)' }}>
          <SidebarLink to="/settings" icon={Settings} label="Settings" />

          <div className="relative mt-0.5">
            <button
              onClick={() => setUserMenuOpen((o) => !o)}
              className="flex items-center gap-2.5 w-full px-3 py-2 rounded-md transition-colors hover:bg-[var(--flow-hover)]"
            >
              <span
                className="h-6 w-6 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                style={{
                  background: `${ACCENT}22`,
                  color: ACCENT,
                }}
              >
                {getInitials(userName)}
              </span>
              <div className="flex flex-col items-start min-w-0 flex-1">
                <span
                  className="text-[12px] font-medium truncate w-full text-left"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {userName}
                </span>
                {userEmail && (
                  <span
                    className="text-[10px] truncate w-full text-left"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {userEmail}
                  </span>
                )}
              </div>
              <ChevronRight
                size={12}
                className="shrink-0"
                style={{ color: 'var(--flow-text-dim)' }}
              />
            </button>

            <AnimatePresence>
              {userMenuOpen && (
                <motion.div
                  initial={{ opacity: 0, y: 4, scale: 0.97 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 4, scale: 0.97 }}
                  transition={{ duration: 0.12 }}
                  className="absolute bottom-full left-0 right-0 mb-1 rounded-lg py-1 z-50"
                  style={{
                    background: 'var(--flow-card)',
                    border: '1px solid var(--flow-border)',
                    boxShadow: 'var(--flow-shadow)',
                  }}
                >
                  <div className="px-3 py-2">
                    <p
                      className="text-[10px] font-semibold uppercase tracking-wider mb-1.5"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      Appearance
                    </p>
                    <div className="flex gap-1">
                      {(
                        [
                          { value: 'light', icon: Sun, label: 'Light' },
                          { value: 'dark', icon: Moon, label: 'Dark' },
                          { value: 'system', icon: Monitor, label: 'System' },
                        ] as const
                      ).map(({ value, icon: Icon, label }) => (
                        <button
                          key={value}
                          onClick={() => setTheme(value)}
                          className="flex-1 flex flex-col items-center gap-1 py-2 rounded-md text-[10px] font-medium transition-colors"
                          style={{
                            background:
                              theme === value
                                ? `${ACCENT}22`
                                : 'var(--flow-hover)',
                            color:
                              theme === value ? ACCENT : 'var(--flow-text-muted)',
                            border:
                              theme === value
                                ? `1px solid ${ACCENT}44`
                                : '1px solid transparent',
                          }}
                        >
                          <Icon size={12} />
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>
                  <div
                    style={{
                      height: '1px',
                      background: 'var(--flow-border)',
                      margin: '4px 0',
                    }}
                  />
                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-left transition-colors hover:bg-[var(--flow-hover)]"
                    style={{ color: '#ef4444' }}
                  >
                    <LogOut size={13} />
                    Sign out
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </aside>

      <div className="flex flex-col flex-1 overflow-hidden">
        <AnimatePresence mode="wait">
          <motion.main
            key={location.pathname}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -4 }}
            transition={{ duration: 0.18, ease: 'easeOut' }}
            className="flex-1 overflow-y-auto"
          >
            {children}
          </motion.main>
        </AnimatePresence>
      </div>

      <CommandBar />
      <QuickCreate />
      <BulkEditBar />
    </div>
  );
};

export default FlowLayout;
