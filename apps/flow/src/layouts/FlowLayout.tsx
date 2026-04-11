import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
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
  ArrowUp,
  ArrowDown,
  Menu,
  X,
} from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { useUserStore } from '../shared/stores/userStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { useFlowStore } from '../shared/stores/flowStore';
import { cn, getInitials } from '../shared/lib/utils';
import CommandBar from '../shared/components/CommandBar';
import QuickCreate from '../shared/components/QuickCreate';
import BulkEditBar from '../shared/components/BulkEditBar';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useWorkspaces } from '../features/workspace-dashboard/useWorkspaces';
import { useProjects } from '../features/projects/useProjects';
import { isTaskMine } from '../shared/lib/currentUser';
import { isFlowLocal } from '../shared/lib/flowLocalMode';
import { displayUserEmail, displayUserName } from '../shared/lib/userDisplay';
import { useSidebarSections } from '../shared/hooks/useSidebarSections';
import {
  useSidebarSectionOrder,
  type SidebarSectionId,
} from '../shared/hooks/useSidebarSectionOrder';
import { useLifePreferences } from '../shared/lib/lifePreferences';
import { isDueToday, isQuietNow } from '../shared/lib/lifeRhythm';

const ACCENT = '#64C5B9';

/** Horizontal inset for sidebar content (single column). */
const sx = 'px-3';

/** Collapsible group headers — compact, same row height as nav links. */
const sidebarSectionHeaderClass =
  'w-full flex items-center justify-between gap-2 px-3 py-1.5 rounded-md text-left transition-colors hover:bg-[var(--flow-hover)]';

interface SidebarLinkProps {
  to: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  label: string;
  exact?: boolean;
  badge?: number;
}

const SidebarSectionReorder: React.FC<{
  id: SidebarSectionId;
  order: SidebarSectionId[];
  moveSection: (id: SidebarSectionId, direction: 'up' | 'down') => void;
}> = ({ id, order, moveSection }) => {
  const i = order.indexOf(id);
  const canUp = i > 0;
  const canDown = i >= 0 && i < order.length - 1;
  return (
    <div
      className="flex h-full flex-col justify-center gap-px overflow-hidden"
      onClick={(e) => e.stopPropagation()}
      onKeyDown={(e) => e.stopPropagation()}
    >
      <button
        type="button"
        disabled={!canUp}
        aria-label="Move section up"
        title="Move up"
        className={cn(
          'rounded p-px transition-colors flex items-center justify-center',
          canUp ? 'hover:bg-[var(--flow-hover)]' : 'opacity-25 cursor-not-allowed'
        )}
        onClick={() => canUp && moveSection(id, 'up')}
      >
        <ArrowUp size={11} strokeWidth={2} style={{ color: 'var(--flow-text-muted)' }} />
      </button>
      <button
        type="button"
        disabled={!canDown}
        aria-label="Move section down"
        title="Move down"
        className={cn(
          'rounded p-px transition-colors flex items-center justify-center',
          canDown ? 'hover:bg-[var(--flow-hover)]' : 'opacity-25 cursor-not-allowed'
        )}
        onClick={() => canDown && moveSection(id, 'down')}
      >
        <ArrowDown size={11} strokeWidth={2} style={{ color: 'var(--flow-text-muted)' }} />
      </button>
    </div>
  );
};

/** Zero layout width until row hover; then expands so the title shifts — no permanent gutter.
 *  (Do not use focus-within: clicking the section toggle keeps focus on the button and would
 *  leave the strip expanded after the mouse leaves.) */
const sidebarReorderWrapClass =
  'shrink-0 w-0 overflow-hidden opacity-0 transition-[width,opacity] duration-150 ease-out pointer-events-none group-hover:w-4 group-hover:opacity-100 group-hover:pointer-events-auto';

const SIDEBAR_SECTION_EASE = [0.4, 0, 0.2, 1] as const;

/** Height + fade when expanding/collapsing Plan / Team / Projects bodies. */
const SidebarCollapsible: React.FC<{ open: boolean; children: React.ReactNode }> = ({
  open,
  children,
}) => (
  <AnimatePresence initial={false}>
    {open && (
      <motion.div
        initial={{ height: 0, opacity: 0 }}
        animate={{ height: 'auto', opacity: 1 }}
        exit={{ height: 0, opacity: 0 }}
        transition={{ duration: 0.22, ease: SIDEBAR_SECTION_EASE }}
        className="overflow-hidden"
      >
        {children}
      </motion.div>
    )}
  </AnimatePresence>
);

const SidebarLink: React.FC<SidebarLinkProps> = ({ to, icon: Icon, label, exact, badge }) => {
  const location = useLocation();
  const isActive = exact ? location.pathname === to : location.pathname.startsWith(to);

  return (
    <Link
      to={to}
      className={cn(
        'relative flex min-h-[40px] items-center gap-2.5 px-3 py-2 rounded-md text-[12px] font-medium transition-colors duration-150 lg:min-h-0 lg:py-1.5',
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

  const { toggle: toggleSidebarSection, isCollapsed: isSidebarSectionCollapsed } =
    useSidebarSections();
  const { order: sidebarSectionOrder, moveSection: moveSidebarSection } = useSidebarSectionOrder();
  const { prefs: lifePrefs } = useLifePreferences();

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
    setQuickCreateOpen,
  } = useFlowStore();

  // Fetch workspaces and projects from API and populate store
  const { data: apiWorkspaces } = useWorkspaces();
  const currentWorkspaceId = workspaceId || activeWorkspaceId || workspaces[0]?.id || '';
  const { data: apiProjects } = useProjects(currentWorkspaceId || undefined);

  // Remote API: hydrate the store from React Query. Local mode uses the store as SSOT — syncing
  // back would loop (setProjects → new filtered array from useProjects → effect runs again).
  useEffect(() => {
    if (isFlowLocal()) return;
    if (apiWorkspaces && apiWorkspaces.length > 0) {
      setWorkspaces(apiWorkspaces as any);
    }
  }, [apiWorkspaces, setWorkspaces]);

  useEffect(() => {
    if (isFlowLocal()) return;
    if (apiProjects && apiProjects.length > 0) {
      setProjects(apiProjects as any);
    }
  }, [apiProjects, setProjects]);

  const [workspaceSwitcherOpen, setWorkspaceSwitcherOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  useEffect(() => {
    setMobileNavOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!mobileNavOpen) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = prev;
    };
  }, [mobileNavOpen]);

  useKeyboardShortcuts();

  const activeWorkspace = workspaces.find((w) => w.id === currentWorkspaceId) ||
    workspaces[0] || { id: '', name: 'Loading...', color: '#6b7280' };
  const workspaceProjects = projects.filter(
    (p) => p.workspaceId === currentWorkspaceId && !p.archived
  );

  const myIssuesCount = tasks.filter(
    (t) =>
      isTaskMine(t.assigneeId) && t.status !== 'done' && t.status !== 'cancelled' && !t.archived
  ).length;
  const todayFocusCount = tasks.filter(
    (t) =>
      t.workspaceId === currentWorkspaceId &&
      !t.archived &&
      t.status !== 'done' &&
      t.status !== 'cancelled' &&
      isTaskMine(t.assigneeId) &&
      (t.focusPin === true || isDueToday(t.dueDate))
  ).length;
  const inQuietHours = isQuietNow(lifePrefs.quietStart, lifePrefs.quietEnd);
  const unreadNotifications = notifications.filter((n) => !n.read).length;

  const userEmail = displayUserEmail(user);
  const userName = displayUserName(user);

  const handleLogout = () => {
    logout();
    window.location.href = APP_URLS.auth;
  };

  const isMac = typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform);
  const modKey = isMac ? '⌘' : 'Ctrl';

  const logoFilter =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)
      ? 'brightness(0) invert(1)'
      : 'brightness(0)';

  return (
    <div
      className="flex h-[100dvh] min-h-0 flex-col overflow-hidden lg:flex-row"
      style={{ background: 'var(--flow-bg)' }}
    >
      {/* Mobile: top bar */}
      <header
        className="flex shrink-0 items-center gap-2 border-b px-3 py-2.5 lg:hidden"
        style={{
          borderColor: 'var(--flow-border)',
          background: 'var(--flow-surface)',
          paddingTop: 'max(0.5rem, env(safe-area-inset-top))',
        }}
      >
        <button
          type="button"
          aria-label="Open menu"
          onClick={() => setMobileNavOpen(true)}
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text)' }}
        >
          <Menu size={20} strokeWidth={2} />
        </button>
        <img
          src="/logo.png"
          alt=""
          className="h-5 w-auto shrink-0"
          style={{ filter: logoFilter }}
        />
        <span
          className="min-w-0 flex-1 truncate text-left text-[13px] font-semibold"
          style={{ color: 'var(--flow-text)' }}
        >
          {activeWorkspace.name}
        </span>
        <button
          type="button"
          aria-label="Search"
          onClick={() => setCommandBarOpen(true)}
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Search size={18} strokeWidth={2} />
        </button>
        <button
          type="button"
          aria-label="New task"
          onClick={() => setQuickCreateOpen(true)}
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
          style={{ color: ACCENT }}
        >
          <Plus size={20} strokeWidth={2} />
        </button>
      </header>

      {mobileNavOpen && (
        <button
          type="button"
          aria-label="Close menu"
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setMobileNavOpen(false)}
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 flex h-full w-[min(18rem,88vw)] flex-col overflow-hidden border-r border-[var(--flow-border)] transition-transform duration-300 ease-out lg:static lg:inset-auto lg:z-auto lg:h-full lg:min-h-0 lg:w-60 lg:max-w-none lg:shrink-0 lg:translate-x-0',
          mobileNavOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
        style={{
          background: 'var(--flow-surface)',
          paddingTop: 'env(safe-area-inset-top)',
        }}
      >
        {/* Logo — same markup as Hub `DefaultLayout` (img src + filter + padding) */}
        <div
          className={cn('flex items-center gap-2.5 py-2.5', sx)}
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <img src="/logo.png" alt="rtylr" className="h-6 w-auto" style={{ filter: logoFilter }} />
          <button
            type="button"
            aria-label="Close menu"
            onClick={() => setMobileNavOpen(false)}
            className="ml-auto flex h-9 w-9 items-center justify-center rounded-lg lg:hidden hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <X size={18} strokeWidth={2} />
          </button>
        </div>

        {/* Workspace + search — shared inset, even vertical rhythm */}
        <div className={cn('space-y-1.5 pt-1.5 pb-1.5 relative', sx)}>
          <button
            type="button"
            onClick={() => setWorkspaceSwitcherOpen((o) => !o)}
            className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md transition-colors hover:bg-[var(--flow-hover)]"
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
                className="absolute left-0 right-0 z-50 mt-1 rounded-lg py-1 shadow-lg"
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

          <button
            type="button"
            onClick={() => setCommandBarOpen(true)}
            className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md text-[11px] transition-colors hover:opacity-90"
            style={{
              background: 'var(--flow-hover)',
              color: 'var(--flow-text-dim)',
            }}
          >
            <Search size={12} />
            <span className="flex-1 text-left">Search…</span>
            <span
              className="hidden text-[9px] font-mono px-1.5 py-0.5 rounded sm:inline"
              style={{ background: 'var(--flow-active)' }}
            >
              {modKey}K
            </span>
          </button>
        </div>

        <div className={cn('flex-1 min-h-0 overflow-y-auto py-2', sx)}>
          <div className="flex flex-col gap-2">
            <nav className="flex flex-col gap-px">
              <SidebarLink to="/inbox" icon={Inbox} label="Inbox" badge={unreadNotifications} />
              <SidebarLink to="/my-issues" icon={ListTodo} label="My tasks" badge={myIssuesCount} />
              <SidebarLink to="/today" icon={Sun} label="Today" badge={todayFocusCount} />
              <SidebarLink
                to={`/w/${currentWorkspaceId}`}
                icon={LayoutGrid}
                label="Dashboard"
                exact
              />
              <SidebarLink to="/triage" icon={Eye} label="Triage" />
            </nav>

            <div className="flex flex-col gap-2">
              {sidebarSectionOrder.map((sectionId) => (
                <React.Fragment key={sectionId}>
                  {sectionId === 'plan' && (
                    <div className="flex flex-col gap-0">
                      <div className="group flex items-center gap-0 min-h-0">
                        <div className={sidebarReorderWrapClass}>
                          <SidebarSectionReorder
                            id="plan"
                            order={sidebarSectionOrder}
                            moveSection={moveSidebarSection}
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => toggleSidebarSection('plan')}
                          className={cn(sidebarSectionHeaderClass, 'flex-1')}
                        >
                          <span
                            className="text-[12px] font-semibold tracking-wide"
                            style={{ color: 'var(--flow-text-muted)' }}
                          >
                            Plan
                          </span>
                          <ChevronDown
                            size={14}
                            strokeWidth={2}
                            className={cn(
                              'shrink-0 opacity-70 transition-transform duration-200 ease-out',
                              isSidebarSectionCollapsed('plan') && '-rotate-90'
                            )}
                            style={{ color: 'var(--flow-text-muted)' }}
                          />
                        </button>
                      </div>
                      <SidebarCollapsible open={!isSidebarSectionCollapsed('plan')}>
                        <nav className="flex flex-col gap-px pt-px">
                          <SidebarLink to="/roadmap" icon={Map} label="Roadmap" />
                          <SidebarLink to="/cycles" icon={Timer} label="Cycles" />
                          <SidebarLink to="/milestones" icon={Milestone} label="Milestones" />
                          <SidebarLink to="/timeline" icon={Timer} label="Timeline" />
                          <SidebarLink to="/calendar" icon={Calendar} label="Calendar" />
                          <SidebarLink to="/views" icon={Eye} label="Views" />
                        </nav>
                      </SidebarCollapsible>
                    </div>
                  )}

                  {sectionId === 'team' && (
                    <div className="flex flex-col gap-0">
                      <div className="group flex items-center gap-0 min-h-0">
                        <div className={sidebarReorderWrapClass}>
                          <SidebarSectionReorder
                            id="team"
                            order={sidebarSectionOrder}
                            moveSection={moveSidebarSection}
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => toggleSidebarSection('team')}
                          className={cn(sidebarSectionHeaderClass, 'flex-1')}
                        >
                          <span
                            className="text-[12px] font-semibold tracking-wide"
                            style={{ color: 'var(--flow-text-muted)' }}
                          >
                            Team
                          </span>
                          <ChevronDown
                            size={14}
                            strokeWidth={2}
                            className={cn(
                              'shrink-0 opacity-70 transition-transform duration-200 ease-out',
                              isSidebarSectionCollapsed('team') && '-rotate-90'
                            )}
                            style={{ color: 'var(--flow-text-muted)' }}
                          />
                        </button>
                      </div>
                      <SidebarCollapsible open={!isSidebarSectionCollapsed('team')}>
                        <nav className="flex flex-col gap-px pt-px">
                          <SidebarLink to="/team" icon={Users} label="Workload" />
                          <SidebarLink to="/insights" icon={BarChart3} label="Insights" />
                          <SidebarLink to="/archive" icon={Archive} label="Archive" />
                        </nav>
                      </SidebarCollapsible>
                    </div>
                  )}

                  {sectionId === 'projects' && (
                    <div className="flex flex-col gap-0">
                      <div className="group flex items-center gap-0 min-h-0">
                        <div className={sidebarReorderWrapClass}>
                          <SidebarSectionReorder
                            id="projects"
                            order={sidebarSectionOrder}
                            moveSection={moveSidebarSection}
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => toggleSidebarSection('projects')}
                          className={cn(sidebarSectionHeaderClass, 'flex-1 min-w-0')}
                        >
                          <span
                            className="text-[12px] font-semibold tracking-wide truncate"
                            style={{ color: 'var(--flow-text-muted)' }}
                          >
                            Projects
                          </span>
                          <ChevronDown
                            size={14}
                            strokeWidth={2}
                            className={cn(
                              'shrink-0 opacity-70 transition-transform duration-200 ease-out',
                              isSidebarSectionCollapsed('projects') && '-rotate-90'
                            )}
                            style={{ color: 'var(--flow-text-muted)' }}
                          />
                        </button>
                        <button
                          type="button"
                          className="h-8 w-8 shrink-0 rounded-md flex items-center justify-center transition-colors hover:bg-[var(--flow-hover)]"
                          style={{ color: 'var(--flow-text-muted)' }}
                          title="New Project"
                        >
                          <Plus size={15} strokeWidth={2} />
                        </button>
                      </div>
                      <SidebarCollapsible open={!isSidebarSectionCollapsed('projects')}>
                        <div className="flex flex-col gap-px pt-px">
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
                                  color: isActive ? 'var(--flow-text)' : 'var(--flow-text-muted)',
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
                      </SidebarCollapsible>
                    </div>
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div
          className={cn('flex flex-col gap-1.5 py-2', sx)}
          style={{ borderTop: '1px solid var(--flow-border)' }}
        >
          <SidebarLink to="/settings" icon={Settings} label="Settings" />

          <div className="relative">
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
                            background: theme === value ? `${ACCENT}22` : 'var(--flow-hover)',
                            color: theme === value ? ACCENT : 'var(--flow-text-muted)',
                            border:
                              theme === value ? `1px solid ${ACCENT}44` : '1px solid transparent',
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

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
        {inQuietHours && (
          <div
            className="shrink-0 px-3 py-2 text-center text-[10px] sm:px-4 sm:text-[11px] leading-snug"
            style={{
              background: 'var(--flow-hover)',
              color: 'var(--flow-text-muted)',
              borderBottom: '1px solid var(--flow-border)',
            }}
          >
            Quiet hours ({lifePrefs.quietStart}–{lifePrefs.quietEnd}) — non-urgent reminders stay
            soft. Digest preference: {lifePrefs.digestHour}:00. Change in Settings → Life & rhythm.
          </div>
        )}
        <AnimatePresence mode="wait">
          <motion.main
            key={location.pathname}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -4 }}
            transition={{ duration: 0.18, ease: 'easeOut' }}
            className="flex-1 overflow-y-auto pb-[max(0.5rem,env(safe-area-inset-bottom))]"
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
