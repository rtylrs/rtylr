import React, { useState, useEffect, useMemo, useRef } from 'react';
import { toast } from 'sonner';
import TrialBanner from '../shared/components/TrialBanner';
import { Link, useLocation, useNavigate, useParams, useRouterState } from '@tanstack/react-router';
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
  ArrowUp,
  ArrowDown,
  Menu,
  X,
  Shield,
  Clock,
} from 'lucide-react';
import { useFeatureFlag, FLAG_KEYS, getAvatarUrl } from '@voxire/config';
import { Avatar } from '@voxire/ui';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import { useForesightLinks } from '@voxire/utils';
import { signOut, useUserStore as useAuthUserStore } from '@voxire/auth';
import { useUserStore } from '../shared/stores/userStore';
import { useTranslation } from 'react-i18next';
import { useThemeStore } from '../shared/stores/themeStore';
import { accentTints, ACCENT_DEFAULT, type ResolvedTheme } from '../shared/lib/colorUtils';
import { useFlowStore } from '../shared/stores/flowStore';
import { cn } from '../shared/lib/utils';
import CommandBar from '../shared/components/CommandBar';
import QuickCreate from '../shared/components/QuickCreate';
import BulkEditBar from '../shared/components/BulkEditBar';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useWorkspaces } from '../features/workspace-dashboard/useWorkspaces';
import { useProjects, useCreateProject } from '../features/projects/useProjects';
import { useMyTasksProject } from '../features/tasks/useMyTasksProject';
import { PERSONAL_WORKSPACE_ID } from '../shared/lib/personalWorkspace';
import type { Member, Workspace } from '../shared/stores/flowStore';
import { isTaskMine } from '../shared/lib/currentUser';
import { displayUserEmail, displayUserName } from '../shared/lib/userDisplay';
import { useSidebarSections } from '../shared/hooks/useSidebarSections';
import {
  useSidebarSectionOrder,
  type SidebarSectionId,
} from '../shared/hooks/useSidebarSectionOrder';
import { useLifePreferences } from '../shared/lib/lifePreferences';
import { isDueToday, isQuietNow } from '../shared/lib/lifeRhythm';
import { useNotifications } from '../features/inbox/useNotifications';
import { apiProjectToProject, apiWorkspaceToWorkspace } from '../shared/lib/projectMapper';
import { LicenseBanner } from '@voxire/components';
import { safe } from '../shared/lib/api';
import { useLayoutRealtime } from '../shared/lib/useLayoutRealtime';
import { useFlowTeam } from '../shared/hooks/queries/useFlowTeam';
import { initialsFromName } from '../shared/lib/memberUtils';

/** Horizontal inset for sidebar content (single column). */
const sx = 'px-3';

/** Collapsible group headers, compact, same row height as nav links. */
const sidebarSectionHeaderClass =
  'w-full flex items-center justify-between gap-2 px-3 py-1.5 rounded-md text-start transition-colors hover:bg-[var(--flow-hover)]';

interface SidebarLinkProps {
  to: string;
  /** Search params for links that target a section of a screen, e.g. Settings. */
  search?: Record<string, string>;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  label: string;
  exact?: boolean;
  badge?: number;
  // When set, fully controls the active state (used for items like "My tasks"
  // that redirect to a project board, so path-matching can't detect them).
  activeOverride?: boolean;
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

/** Zero layout width until row hover; then expands so the title shifts, no permanent gutter.
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

const SidebarLink: React.FC<SidebarLinkProps> = ({
  to,
  search,
  icon: Icon,
  label,
  exact,
  badge,
  activeOverride,
}) => {
  const location = useLocation();
  const isActive =
    activeOverride !== undefined
      ? activeOverride
      : exact
        ? location.pathname === to
        : location.pathname.startsWith(to);

  return (
    <Link
      to={to}
      search={search}
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
          style={{ background: 'var(--accent)' }}
        />
      )}
    </Link>
  );
};

interface FlowLayoutProps {
  children: React.ReactNode;
}

const FlowLayout: React.FC<FlowLayoutProps> = ({ children }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const navigate = useNavigate();
  const { workspaceId, projectId } = useParams({ strict: false });
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);

  const { toggle: toggleSidebarSection, isCollapsed: isSidebarSectionCollapsed } =
    useSidebarSections();
  const { order: sidebarSectionOrder, moveSection: moveSidebarSection } = useSidebarSectionOrder();
  const { prefs: lifePrefs } = useLifePreferences();

  const authUser = useAuthUserStore((s) => s.user);
  const { user, setUser } = useUserStore();

  useEffect((): void => {
    if (!authUser) return;
    setUser({
      id: authUser.id ?? authUser.userId ?? '',
      name: authUser.name ?? authUser.firstName ?? '',
      email: authUser.email ?? authUser.emailAddress ?? '',
    });
  }, [authUser, setUser]);
  const { theme, setTheme } = useThemeStore();
  const {
    workspaces,
    projects,
    activeWorkspaceId,
    setActiveWorkspace,
    setWorkspaces,
    setProjects,
    setMembers,
    tasks,
    setCommandBarOpen,
    setQuickCreateOpen,
  } = useFlowStore();

  const { data: apiNotifications = [] } = useNotifications();
  const { isOpen: caretOpen } = useCaretStore();
  const unreadNotifications = apiNotifications.filter((n) => !n.read).length;

  // Fetch workspaces from API.
  // IMPORTANT: use React Query data directly for display, never wait for the
  // Zustand store to hydrate, because the store is populated via useEffect which
  // fires *after* render, causing a flash of "No workspace".
  const { data: apiWorkspaces, isLoading: workspacesLoading } = useWorkspaces();

  const PERSONAL_FALLBACK: Workspace = {
    id: PERSONAL_WORKSPACE_ID,
    name: 'Personal',
    color: ACCENT_DEFAULT,
  };

  // Effective workspace list: prefer live API data, fall back to store, guarantee Personal.
  // Always map through apiWorkspaceToWorkspace so IDs are strings - URL params are strings
  // and === comparisons would silently fail against numeric API IDs.
  const liveWorkspaces = useMemo<Workspace[]>(() => {
    const list =
      apiWorkspaces && apiWorkspaces.length > 0
        ? apiWorkspaces.map(apiWorkspaceToWorkspace)
        : workspaces;
    if (list.length === 0) return workspacesLoading ? [] : [PERSONAL_FALLBACK];
    return list;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiWorkspaces, workspaces, workspacesLoading]);

  const currentWorkspaceId = workspaceId || activeWorkspaceId || liveWorkspaces[0]?.id || '';

  // Resolves the caller's auto-provisioned personal project so it can be hidden
  // from the Projects list (its tasks surface in the global "My tasks" view).
  const { data: myTasksProject } = useMyTasksProject();
  const { data: apiProjects } = useProjects(currentWorkspaceId || undefined);

  // Once real workspaces load, redirect away from the "personal" sentinel to the
  // actual first workspace ID so all downstream API calls use a real numeric ID.
  useEffect(() => {
    if (workspaceId === PERSONAL_WORKSPACE_ID && apiWorkspaces && apiWorkspaces.length > 0) {
      navigate({ to: `/w/${apiWorkspaces[0].id}` as string, replace: true });
    }
  }, [workspaceId, apiWorkspaces, navigate]);

  // Hydrate the store so other consumers (tasks, etc.) see fresh data.
  // IDs from the wire are numeric; the Zustand store keys them as strings
  // (URL params and active-id comparisons are strings), so always route
  // through the mapper - never `setWorkspaces(apiWorkspaces as any)`.
  useEffect(() => {
    if (apiWorkspaces && apiWorkspaces.length > 0) {
      setWorkspaces(apiWorkspaces.map(apiWorkspaceToWorkspace));
    }
  }, [apiWorkspaces, setWorkspaces]);

  useEffect(() => {
    if (apiProjects && apiProjects.length > 0) {
      setProjects(apiProjects.map(apiProjectToProject));
    }
  }, [apiProjects, setProjects]);

  // Seed the store's members from the business team so the assignee picker (and
  // any other member-driven UI) has people to choose from. Nothing else calls
  // setMembers, so without this the store stays empty and assignment is impossible.
  const { data: teamMembers } = useFlowTeam();
  useEffect(() => {
    if (!teamMembers || teamMembers.length === 0) return;
    const mapped: Member[] = teamMembers
      .filter((m) => m.isActive)
      .map((m) => {
        const role: Member['role'] =
          m.role === 'owner' || m.role === 'admin' || m.role === 'viewer' ? m.role : 'member';
        return {
          id: String(m.id),
          name: m.name,
          email: m.email,
          initials: initialsFromName(m.name),
          role,
          capacity: 40,
        };
      });
    setMembers(mapped);
  }, [teamMembers, setMembers]);

  // ── Realtime wiring ─────────────────────────────────────────────────
  // Topics, invalidation fanout and the self-event skip all live in
  // useLayoutRealtime, alongside useTaskRealtime and useProjectRealtime.
  const userId = user?.id;
  const caretUserId = Number(userId) || 0;
  const shouldUseRealtime = !location.pathname.startsWith('/settings');
  useLayoutRealtime(currentWorkspaceId, userId, shouldUseRealtime);

  const sidebarRef = useRef<HTMLDivElement | null>(null);
  useForesightLinks(sidebarRef, {
    '/inbox': () => import('../features/inbox/Inbox'),
    '/my-issues': () => import('../features/tasks/MyIssues'),
    '/today': () => import('../features/focus-today/FocusToday'),
    '/triage': () => import('../features/triage/Triage'),
    '/cycles': () => import('../features/cycles/Cycles'),
    '/milestones': () => import('../features/milestones/Milestones'),
    '/calendar': () => import('../features/calendar/Calendar'),
    '/views': () => import('../features/views/Views'),
    '/archive': () => import('../features/archive/Archive'),
    '/settings': () => import('../features/settings/FlowSettings'),
  });

  const createProject = useCreateProject();
  const [workspaceSwitcherOpen, setWorkspaceSwitcherOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [newProjectOpen, setNewProjectOpen] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');
  const [newProjectDescription, setNewProjectDescription] = useState('');
  const [newProjectColor, setNewProjectColor] = useState('#64C5B9');

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

  const activeWorkspace =
    liveWorkspaces.find((w) => w.id === currentWorkspaceId) ||
    liveWorkspaces[0] ||
    (workspacesLoading ? { id: '', name: '', color: '#6b7280' } : PERSONAL_FALLBACK);

  // Sync --accent CSS vars to the active workspace color so every component
  // in the tree can reference var(--accent) without prop-drilling.
  //
  // These land as inline styles on documentElement, which beats every stylesheet
  // rule - so the per-mode alphas have to be applied here. A light-mode value for
  // --accent-muted written in globals.css is silently overridden the moment this
  // layout mounts, which is on every authenticated screen.
  useEffect(() => {
    const resolved: ResolvedTheme =
      document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark';
    const root = document.documentElement;
    for (const [name, value] of Object.entries(
      accentTints(activeWorkspace.color || ACCENT_DEFAULT, resolved)
    )) {
      root.style.setProperty(name, value);
    }
  }, [activeWorkspace.color, theme]);
  // Hide the auto-provisioned personal project - it's the backend's bucket for
  // the dedicated "My Tasks" page (top nav) and, like Asana, must not surface in
  // the Projects list. The authoritative match is the id returned by
  // useMyTasksProject (the icon/name can vary, so the heuristic alone misses it);
  // the name+icon checks remain as a fallback for before that query resolves.
  // "Inbox"/📥 is the legacy name (migrated to My Tasks).
  const myTasksProjectId =
    myTasksProject?.projectId !== undefined ? String(myTasksProject.projectId) : undefined;
  const workspaceProjects = projects.filter(
    (p) =>
      p.workspaceId === currentWorkspaceId &&
      !p.archived &&
      p.id !== myTasksProjectId &&
      !(p.name === 'My Tasks' && p.icon === '✅') &&
      !(p.name === 'Inbox' && p.icon === '📥')
  );
  const activeProject = workspaceProjects.find((project) => project.id === projectId);

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
  const flowPageTitle = activeProject
    ? activeProject.name
    : location.pathname.startsWith('/inbox')
      ? 'Inbox'
      : location.pathname.startsWith('/my-issues')
        ? 'My tasks'
        : location.pathname.startsWith('/today')
          ? 'Today'
          : location.pathname.startsWith('/triage')
            ? 'Triage'
            : location.pathname.startsWith('/calendar')
              ? 'Calendar'
              : location.pathname.startsWith('/cycles')
                ? 'Cycles'
                : location.pathname.startsWith('/milestones')
                  ? 'Milestones'
                  : location.pathname.startsWith('/settings')
                    ? 'Settings'
                    : 'Workspace';
  const caretPageContext = useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'flow',
        appName: 'Flow',
        pathname: location.pathname,
        search,
        page: flowPageTitle,
        businessName: activeWorkspace.name,
        selected: {
          workspace_id: currentWorkspaceId,
          ...(activeProject
            ? { project_id: activeProject.id, project_name: activeProject.name }
            : {}),
        },
        visibleRows: workspaceProjects.slice(0, 12).map((project) => ({
          id: project.id,
          name: project.name,
          open_tasks: Math.max(0, project.taskCount - project.completedCount),
        })),
        hints: [
          `${myIssuesCount} assigned open tasks`,
          `${todayFocusCount} due or pinned focus items`,
          inQuietHours ? 'quiet hours are active' : undefined,
        ].filter((hint): hint is string => Boolean(hint)),
      }),
    [
      activeProject,
      activeWorkspace.name,
      currentWorkspaceId,
      flowPageTitle,
      inQuietHours,
      location.pathname,
      search,
      myIssuesCount,
      todayFocusCount,
      workspaceProjects,
    ]
  );

  const userEmail = displayUserEmail(user);
  const userName = displayUserName(user);

  const handleLogout = (): void => {
    signOut('flow');
  };

  const isMac = typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform);
  const modKey = isMac ? '⌘' : 'Ctrl';

  const logoSrc =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)
      ? '/logo.png'
      : '/logo-dark.png';

  return (
    <>
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
          <img src={logoSrc} alt="" className="h-7 w-auto shrink-0" />
          <span
            className="min-w-0 flex-1 truncate text-start text-[13px] font-semibold"
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
            style={{ color: 'var(--accent)' }}
          >
            <Plus size={20} strokeWidth={2} />
          </button>
          <CaretButton />
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
            'fixed inset-y-0 start-0 z-50 flex h-full w-[min(18rem,88vw)] flex-col overflow-hidden border-e border-[var(--flow-border)] transition-transform duration-300 ease-out lg:static lg:inset-auto lg:z-auto lg:h-full lg:min-h-0 lg:w-60 lg:max-w-none lg:shrink-0 lg:translate-x-0',
            mobileNavOpen
              ? 'translate-x-0'
              : '-translate-x-full rtl:translate-x-full lg:translate-x-0 lg:rtl:translate-x-0'
          )}
          style={{
            background: 'var(--flow-surface)',
            paddingTop: 'env(safe-area-inset-top)',
          }}
        >
          {/* Theme-specific rtylr wordmark. */}
          <div
            className={cn('flex items-center gap-2.5 py-2.5', sx)}
            style={{ borderBottom: '1px solid var(--flow-border)' }}
          >
            <img src={logoSrc} alt="rtylr" className="h-9 w-auto" />
            <button
              type="button"
              aria-label="Close menu"
              onClick={() => setMobileNavOpen(false)}
              className="ms-auto flex h-9 w-9 items-center justify-center rounded-lg lg:hidden hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <X size={18} strokeWidth={2} />
            </button>
          </div>

          {/* Workspace + search, shared inset, even vertical rhythm */}
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
                className="flex-1 text-start text-[12px] font-medium truncate"
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
                  className="absolute inset-x-0 z-50 mt-1 rounded-lg py-1 shadow-lg"
                  style={{
                    background: 'var(--flow-card)',
                    border: '1px solid var(--flow-border)',
                    boxShadow: 'var(--flow-shadow)',
                  }}
                >
                  {liveWorkspaces.map((ws) => (
                    <button
                      key={ws.id}
                      onClick={() => {
                        setActiveWorkspace(ws.id);
                        navigate({ to: `/w/${ws.id}` as string });
                        setWorkspaceSwitcherOpen(false);
                      }}
                      className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-start transition-colors hover:bg-[var(--flow-hover)]"
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
                          className="ms-auto h-1.5 w-1.5 rounded-full"
                          style={{ background: 'var(--accent)' }}
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
              <span className="flex-1 text-start">Search…</span>
              <span
                className="hidden text-[9px] font-mono px-1.5 py-0.5 rounded sm:inline"
                style={{ background: 'var(--flow-active)' }}
              >
                {modKey}K
              </span>
            </button>
          </div>

          <div ref={sidebarRef} className={cn('flex-1 min-h-0 overflow-y-auto py-2', sx)}>
            <div className="flex flex-col gap-2">
              <nav className="flex flex-col gap-px">
                <SidebarLink to="/inbox" icon={Inbox} label="Inbox" badge={unreadNotifications} />
                <SidebarLink
                  to="/my-issues"
                  icon={ListTodo}
                  label="My tasks"
                  badge={myIssuesCount}
                  exact
                />
                <SidebarLink to="/today" icon={Sun} label="Today" badge={todayFocusCount} />
                <SidebarLink
                  to={`/w/${currentWorkspaceId}`}
                  icon={LayoutGrid}
                  label="Dashboard"
                  exact
                />
                <SidebarLink to="/triage" icon={Eye} label="Triage" />
                {/* Backlog is now a column inside /my-issues - no
                    standalone sidebar entry. */}
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
                            {showPreviewRoutes && (
                              <SidebarLink to="/roadmap" icon={Map} label="Roadmap" />
                            )}
                            <SidebarLink to="/cycles" icon={Timer} label="Cycles" />
                            <SidebarLink to="/milestones" icon={Milestone} label="Milestones" />
                            {showPreviewRoutes && (
                              <SidebarLink to="/timeline" icon={Timer} label="Timeline" />
                            )}
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
                            {showPreviewRoutes && (
                              <SidebarLink to="/team" icon={Users} label="Workload" />
                            )}
                            <SidebarLink
                              to={`/w/${currentWorkspaceId}/members`}
                              icon={Shield}
                              label="Members"
                            />
                            {showPreviewRoutes && (
                              <SidebarLink to="/insights" icon={BarChart3} label="Insights" />
                            )}
                            <SidebarLink to="/archive" icon={Archive} label="Archive" />
                          </nav>
                        </SidebarCollapsible>
                      </div>
                    )}

                    {sectionId === 'time' && (
                      <div className="flex flex-col gap-0">
                        <div className="group flex items-center gap-0 min-h-0">
                          <div className={sidebarReorderWrapClass}>
                            <SidebarSectionReorder
                              id="time"
                              order={sidebarSectionOrder}
                              moveSection={moveSidebarSection}
                            />
                          </div>
                          <button
                            type="button"
                            onClick={() => toggleSidebarSection('time')}
                            className={cn(sidebarSectionHeaderClass, 'flex-1')}
                          >
                            <span
                              className="text-[12px] font-semibold tracking-wide"
                              style={{ color: 'var(--flow-text-muted)' }}
                            >
                              {t('flow.time.heading')}
                            </span>
                            <ChevronDown
                              size={14}
                              strokeWidth={2}
                              className={cn(
                                'shrink-0 opacity-70 transition-transform duration-200 ease-out',
                                isSidebarSectionCollapsed('time') && '-rotate-90'
                              )}
                              style={{ color: 'var(--flow-text-muted)' }}
                            />
                          </button>
                        </div>
                        <SidebarCollapsible open={!isSidebarSectionCollapsed('time')}>
                          <nav className="flex flex-col gap-px pt-px">
                            {/* Timesheet, Review and Workload join these here
                             * as their backends land. Team and Reports have
                             * no client-side gate: a caller without workspace
                             * admin still opens them and sees why, the same
                             * disabled-with-a-reason rule as everywhere else,
                             * just enforced by a real 403 instead of a
                             * client-only guess. */}
                            <SidebarLink to="/time" icon={Clock} label={t('flow.time.myDay')} />
                            <SidebarLink to="/time/team" icon={Users} label={t('flow.time.team')} />
                            <SidebarLink
                              to="/time/reports"
                              icon={BarChart3}
                              label={t('flow.time.reports')}
                            />
                            <SidebarLink
                              to="/settings"
                              search={{ section: 'time' }}
                              icon={Settings}
                              label={t('flow.settings.time.section')}
                            />
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
                            onClick={() => {
                              setNewProjectName('');
                              setNewProjectDescription('');
                              setNewProjectColor('#64C5B9');
                              setNewProjectOpen(true);
                            }}
                          >
                            <Plus size={15} strokeWidth={2} />
                          </button>
                        </div>
                        <SidebarCollapsible open={!isSidebarSectionCollapsed('projects')}>
                          <div className="flex flex-col gap-px pt-px">
                            {workspaceProjects.map((project) => {
                              const boardPath = `/w/${currentWorkspaceId}/p/${project.id}`;
                              const isActive = projectId === project.id;
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
            <CaretButton className="w-full justify-start" />

            <div className="relative">
              <button
                onClick={() => setUserMenuOpen((o) => !o)}
                className="flex items-center gap-2.5 w-full px-3 py-2 rounded-md transition-colors hover:bg-[var(--flow-hover)]"
              >
                <Avatar
                  src={getAvatarUrl(userEmail || userName)}
                  alt={userName}
                  size="sm"
                  className="h-6 w-6"
                  style={{ background: 'var(--accent-muted)' }}
                />
                <div className="flex flex-col items-start min-w-0 flex-1">
                  <span
                    className="text-[12px] font-medium truncate w-full text-start"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {userName}
                  </span>
                  {userEmail && (
                    <span
                      className="text-[10px] truncate w-full text-start"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {userEmail}
                    </span>
                  )}
                </div>
                <ChevronRight
                  size={12}
                  className="flow-directional-icon shrink-0"
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
                    className="absolute bottom-full inset-x-0 mb-1 rounded-lg py-1 z-50"
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
                                theme === value ? 'var(--accent-muted)' : 'var(--flow-hover)',
                              color: theme === value ? 'var(--accent)' : 'var(--flow-text-muted)',
                              border:
                                theme === value
                                  ? '1px solid var(--accent-border)'
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
                      className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-start transition-colors hover:bg-[var(--flow-hover)]"
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

        {/* Create project modal */}
        <AnimatePresence>
          {newProjectOpen && (
            <motion.div
              className="fixed inset-0 z-[200] flex items-center justify-center p-4"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.15 }}
            >
              <motion.button
                type="button"
                aria-label="Close"
                className="absolute inset-0 cursor-default"
                style={{ background: 'rgba(0,0,0,0.45)' }}
                onClick={() => setNewProjectOpen(false)}
              />
              <motion.div
                role="dialog"
                aria-modal="true"
                className="relative w-full max-w-sm rounded-xl p-5 shadow-2xl"
                style={{
                  background: 'var(--flow-surface)',
                  border: '1px solid var(--flow-border)',
                }}
                initial={{ opacity: 0, scale: 0.96, y: 8 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.96, y: 8 }}
                transition={{ duration: 0.15 }}
              >
                <p className="text-[14px] font-semibold mb-4" style={{ color: 'var(--flow-text)' }}>
                  New project
                </p>
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-2">
                    <input
                      type="color"
                      value={newProjectColor}
                      onChange={(e) => setNewProjectColor(e.target.value)}
                      title="Pick a color"
                      className="h-8 w-8 shrink-0 rounded-md cursor-pointer border-0 p-0.5"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border)',
                      }}
                    />
                    <input
                      autoFocus
                      type="text"
                      placeholder="Project name"
                      value={newProjectName}
                      onChange={(e) => setNewProjectName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          if (!newProjectName.trim()) {
                            toast.error('Project name is required');
                            return;
                          }
                          if (!currentWorkspaceId) return;
                          createProject.mutate(
                            {
                              workspaceId: currentWorkspaceId,
                              name: newProjectName.trim(),
                              color: newProjectColor,
                            },
                            {
                              onSuccess: () => setNewProjectOpen(false),
                              onError: () => toast.error('Something went wrong. Please try again.'),
                            }
                          );
                        }
                        if (e.key === 'Escape') setNewProjectOpen(false);
                      }}
                      className="flex-1 rounded-lg px-3 py-2 text-[13px] outline-none"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border)',
                        color: 'var(--flow-text)',
                      }}
                    />
                  </div>
                  <textarea
                    placeholder="Description (optional)"
                    value={newProjectDescription}
                    onChange={(e) => setNewProjectDescription(e.target.value)}
                    rows={3}
                    className="w-full rounded-lg px-3 py-2.5 text-[13px] outline-none resize-none placeholder:opacity-45"
                    style={{
                      background: 'var(--flow-input-bg)',
                      border: '1px solid var(--flow-border)',
                      color: 'var(--flow-text)',
                    }}
                  />
                </div>
                <div className="flex justify-end gap-2 mt-5">
                  <button
                    type="button"
                    onClick={() => setNewProjectOpen(false)}
                    className="px-3 py-1.5 rounded-lg text-[12px] font-medium transition-colors hover:bg-[var(--flow-hover)]"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    disabled={!newProjectName.trim() || createProject.isPending}
                    onClick={() => {
                      if (!newProjectName.trim() || !currentWorkspaceId) return;
                      createProject.mutate(
                        {
                          workspaceId: currentWorkspaceId,
                          name: newProjectName.trim(),
                          color: newProjectColor,
                        },
                        {
                          onSuccess: () => setNewProjectOpen(false),
                          onError: () => toast.error('Something went wrong. Please try again.'),
                        }
                      );
                    }}
                    className="px-4 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40 disabled:cursor-not-allowed"
                    style={{ background: 'var(--accent)', color: '#0a0a0a' }}
                  >
                    {createProject.isPending ? 'Creating…' : 'Create'}
                  </button>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="flex min-h-0 flex-1 overflow-hidden">
          {/* Main content column */}
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
            <TrialBanner />
            {inQuietHours && (
              <div
                className="shrink-0 px-3 py-2 text-center text-[10px] sm:px-4 sm:text-[11px] leading-snug"
                style={{
                  background: 'var(--flow-hover)',
                  color: 'var(--flow-text-muted)',
                  borderBottom: '1px solid var(--flow-border)',
                }}
              >
                Quiet hours ({lifePrefs.quietStart}–{lifePrefs.quietEnd}), non-urgent reminders stay
                soft. Digest preference: {lifePrefs.digestHour}:00. Change in Settings → Life &
                rhythm.
              </div>
            )}
            <LicenseBanner safe={safe} />
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

          {/* Caret AI - inline third column, pushes content left */}
          <div
            style={{
              width: caretOpen ? '360px' : '0px',
              flexShrink: 0,
              overflow: 'hidden',
              transition: 'width 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            <CaretPanel
              variant="inline"
              appId="flow"
              userId={caretUserId}
              apiBaseUrl={import.meta.env.VITE_API_URL || 'https://api.voxire.com/'}
              systemContext="User is in the Flow app (projects, tasks, team workspaces)"
              pageContext={caretPageContext}
            />
          </div>
        </div>

        <CommandBar />
        <QuickCreate />
        <BulkEditBar />
      </div>
    </>
  );
};

export default FlowLayout;
