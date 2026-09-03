import React, { useState, useEffect, useMemo, useRef, useCallback } from 'react';
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
  Clock,
  UserCircle,
  CalendarOff,
  Contact,
  Radar,
} from 'lucide-react';
import { useFeatureFlag, FLAG_KEYS, getAvatarUrl } from '@voxire/config';
import { AppearanceProvider, Avatar } from '@voxire/ui';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import { useClickOutside, useForesightLinks } from '@voxire/utils';
import { useMemberCapacityHours } from '../shared/hooks/queries/useMemberCapacity';
import { signOut, useUserStore as useAuthUserStore } from '@voxire/auth';
import { useUserStore } from '../shared/stores/userStore';
import { useTranslation } from 'react-i18next';
import { useAppearanceStore } from '../shared/stores/appearanceStore';
import { ACCENT_DEFAULT } from '../shared/lib/colorUtils';
import { useFlowStore } from '../shared/stores/flowStore';
import { cn } from '../shared/lib/utils';
import CommandBar from '../shared/components/CommandBar';
import QuickCreate from '../shared/components/QuickCreate';
import BulkEditBar from '../shared/components/BulkEditBar';
import GateNotice from '../shared/components/GateNotice';
import ModeToggle from '../shared/components/ModeToggle';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useProjects, useCreateProject } from '../features/projects/useProjects';
import { useMyTasksProject } from '../features/tasks/useMyTasksProject';
import { RecycleBinSection } from '../features/projects-v2/RecycleBinSection';
import { PERSONAL_WORKSPACE_ID } from '../shared/lib/personalWorkspace';
import type { Member } from '../shared/stores/flowStore';
import { isTaskMine } from '../shared/lib/currentUser';
import { hasFlowPermission, toFlowMemberRole } from '../shared/lib/roleUtils';
import { displayUserEmail, displayUserName } from '../shared/lib/userDisplay';
import { useSidebarSections } from '../shared/hooks/useSidebarSections';
import {
  useSidebarSectionOrder,
  type SidebarSectionId,
} from '../shared/hooks/useSidebarSectionOrder';
import { useLifePreferences } from '../shared/lib/lifePreferences';
import { isDueToday, isQuietNow } from '../shared/lib/lifeRhythm';
import { useNotifications } from '../features/inbox/useNotifications';
import { apiProjectToProject } from '../shared/lib/projectMapper';
import { LicenseBanner } from '@voxire/components';
import { licenseClient } from '../shared/lib/clients';
import { useFlowBusinessName, useFlowBusinessId } from '../shared/hooks/useFlowBusinessName';
import { useLayoutRealtime } from '../shared/lib/useLayoutRealtime';
import { useRealtimeReconnectRefetch, useRealtimeStatus } from '../shared/lib/realtime';
import { routeNeedsRealtime } from '../shared/lib/realtimeRouteGate';
import { useFlowTeam } from '../shared/hooks/queries/useFlowTeam';
import { initialsFromName } from '../shared/lib/memberUtils';
import RealtimeConnectionIndicator from '../shared/components/RealtimeConnectionIndicator';
import { SidebarNavigation } from '../shared/components/SidebarNavigation';
import { recordRecentSidebarPage, useSidebarPreferences } from '../shared/lib/sidebarPreferences';

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
  const { t } = useTranslation();
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
        aria-label={t('flow.shell.moveSectionUp')}
        title={t('flow.shell.moveUp')}
        className={cn(
          'rounded-sm p-px transition-colors flex items-center justify-center',
          canUp ? 'hover:bg-[var(--flow-hover)]' : 'opacity-25 cursor-not-allowed'
        )}
        onClick={() => canUp && moveSection(id, 'up')}
      >
        <ArrowUp size={11} strokeWidth={2} style={{ color: 'var(--flow-text-muted)' }} />
      </button>
      <button
        type="button"
        disabled={!canDown}
        aria-label={t('flow.shell.moveSectionDown')}
        title={t('flow.shell.moveDown')}
        className={cn(
          'rounded-sm p-px transition-colors flex items-center justify-center',
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
  const navigate = useNavigate();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const { workspaceId, projectId } = useParams({ strict: false });
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);

  const { toggle: toggleSidebarSection, isCollapsed: isSidebarSectionCollapsed } =
    useSidebarSections();
  const { order: sidebarSectionOrder, moveSection: moveSidebarSection } = useSidebarSectionOrder();
  const {
    preferences: sidebarPreferences,
    update: updateSidebarPreferences,
    isLoading: sidebarPreferencesLoading,
  } = useSidebarPreferences();
  const { prefs: lifePrefs } = useLifePreferences();

  const authUser = useAuthUserStore((s) => s.user);
  const effectiveAccess = useAuthUserStore((s) => s.effectiveAccess);
  const { user, setUser } = useUserStore();

  useEffect((): void => {
    if (!authUser) return;
    setUser({
      id: authUser.id ?? authUser.userId ?? '',
      name: authUser.name ?? authUser.firstName ?? '',
      email: authUser.email ?? authUser.emailAddress ?? '',
      businessName: authUser.businessName,
    });
  }, [authUser, setUser]);
  // The provider takes the whole preference set, so this reads the store
  // unselected. Zustand hands back the same object until something is set, and
  // the provider re-derives from the individual fields, so a render that changes
  // nothing does not re-write a token.
  const appearance = useAppearanceStore();
  const { mode } = appearance;
  const {
    projects,
    activeWorkspaceId,
    setActiveWorkspace,
    setProjects,
    members,
    setMembers,
    tasks,
    setCommandBarOpen,
    setQuickCreateOpen,
  } = useFlowStore();

  const { data: apiNotifications = [] } = useNotifications();
  const { isOpen: caretOpen } = useCaretStore();
  const unreadNotifications = apiNotifications.filter((n) => !n.read).length;

  const { data: businessName } = useFlowBusinessName();
  const { data: realtimeBusinessId } = useFlowBusinessId();

  // The workspace layer is gone. Realtime and the Caret route context still
  // take a workspace argument, so this stays as one empty constant rather than
  // threading the removal through those signatures here.
  const currentWorkspaceId = '';

  const canCreateProject = hasFlowPermission(effectiveAccess, 'flow:projects:create');
  const createProjectGateReason = canCreateProject ? '' : t('flow.shell.createNeedsRole');

  // Workspace-scoped routes are authoritative even when the workspace list
  // request is unavailable. Keep the shared store aligned so screens without
  // a workspace segment, such as Cycles and Settings, continue using the same
  // workspace after client-side navigation.
  useEffect(() => {
    if (workspaceId && workspaceId !== PERSONAL_WORKSPACE_ID && activeWorkspaceId !== workspaceId) {
      setActiveWorkspace(workspaceId);
    }
  }, [activeWorkspaceId, setActiveWorkspace, workspaceId]);

  // Resolves the caller's auto-provisioned personal project so it can be hidden
  // from the Projects list (its tasks surface in the global "My tasks" view).
  const { data: myTasksProject } = useMyTasksProject();
  const { data: apiProjects } = useProjects();

  useEffect(() => {
    if (apiProjects && apiProjects.length > 0) {
      setProjects(apiProjects.map(apiProjectToProject));
    }
  }, [apiProjects, setProjects]);

  // Seed the store's members from the business team so the assignee picker (and
  // any other member-driven UI) has people to choose from. Nothing else calls
  // setMembers, so without this the store stays empty and assignment is impossible.
  const { data: teamMembers } = useFlowTeam();
  // Real contracted hours per member, replacing the hardcoded 40 (#868). Absent
  // for a member with no work schedule, which downstream renders as unknown.
  const { data: capacityHours = EMPTY_CAPACITY } = useMemberCapacityHours();
  useEffect(() => {
    if (!teamMembers || teamMembers.length === 0) return;
    const mapped: Member[] = teamMembers
      .filter((m) => m.isActive)
      .map((m) => {
        const role = toFlowMemberRole(m.role);
        return {
          id: String(m.id),
          name: m.name,
          email: m.email,
          initials: initialsFromName(m.name),
          role,
          capacity: capacityHours[String(m.id)],
        };
      });
    setMembers(mapped);
  }, [teamMembers, capacityHours, setMembers]);

  // ── Realtime wiring ─────────────────────────────────────────────────
  // Topics, invalidation fanout and the self-event skip all live in
  // useLayoutRealtime, alongside useTaskRealtime and useProjectRealtime.
  const userId = user?.id;
  const caretUserId = Number(userId) || 0;
  const shouldUseRealtime = routeNeedsRealtime(location.pathname);
  useLayoutRealtime(realtimeBusinessId ?? undefined, userId, shouldUseRealtime);
  // Reconnect can happen regardless of which topics are currently active, so
  // this runs unconditionally rather than gated by shouldUseRealtime.
  useRealtimeReconnectRefetch();
  // Same connection this drives - surfaced here so a dropped socket is
  // visible from anywhere in the shell, not just inside TimePanel. Gated by
  // shouldUseRealtime below at both render sites: on a route that never asks
  // the socket to connect (Settings), the status is "disconnected" in the
  // narrow sense but nothing on screen is stale, so showing the indicator
  // there is a false alarm, not a real one (issue #761).
  const realtimeStatus = useRealtimeStatus();

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
  const [newProjectVisibility, setNewProjectVisibility] = useState<'private' | 'public'>('private');

  // Account menu (avatar/name trigger in the sidebar footer): closes on an
  // outside click and on Escape, scoped to this menu only - it must not
  // interfere with the unrelated "New Project" modal's own Escape handling
  // further down in this file.
  const userMenuRef = useRef<HTMLDivElement | null>(null);
  const closeUserMenu = useCallback((): void => setUserMenuOpen(false), []);
  useClickOutside(userMenuRef, closeUserMenu);
  useEffect(() => {
    if (!userMenuOpen) return;
    const handleKeyDown = (e: KeyboardEvent): void => {
      if (e.key === 'Escape') setUserMenuOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [userMenuOpen]);

  // Workspace switcher (sidebar top): same behaviour as the account menu above
  // and kept independent - each menu closes only itself.
  const workspaceSwitcherRef = useRef<HTMLDivElement | null>(null);
  const closeWorkspaceSwitcher = useCallback((): void => setWorkspaceSwitcherOpen(false), []);
  useClickOutside(workspaceSwitcherRef, closeWorkspaceSwitcher);
  useEffect(() => {
    if (!workspaceSwitcherOpen) return;
    const handleKeyDown = (e: KeyboardEvent): void => {
      if (e.key === 'Escape') setWorkspaceSwitcherOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [workspaceSwitcherOpen]);

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

  const activeWorkspace = {
    id: '',
    name: businessName || user?.businessName || authUser?.businessName || t('flow.shell.workspace'),
    color: ACCENT_DEFAULT,
  };

  // Hide the auto-provisioned personal project - it's the backend's bucket for
  // the dedicated "My Tasks" page (top nav) and, like Asana, must not surface in
  // the Projects list. The authoritative match is the id returned by
  // useMyTasksProject (the icon/name can vary, so the heuristic alone misses it);
  // the name+icon checks remain as a fallback for before that query resolves.
  // "Inbox"/📥 is the legacy name (migrated to My Tasks).
  const myTasksProjectId =
    myTasksProject?.projectId !== undefined ? String(myTasksProject.projectId) : undefined;
  // Every project in the business, not one workspace's: the sidebar is the
  // curated business list now (P8).
  const sidebarProjects = projects.filter(
    (p) =>
      !p.archived &&
      p.id !== myTasksProjectId &&
      !(p.name === 'My Tasks' && p.icon === '✅') &&
      !(p.name === 'Inbox' && p.icon === '📥')
  );
  const activeProject = sidebarProjects.find((project) => project.id === projectId);

  // P9: a project visible only because the caller manages someone on it belongs
  // under its own heading. Listed together with the caller's own it would read as
  // their work, which is what visible_via_report exists to distinguish. With no
  // reports every project falls in the first group and the sidebar is unchanged.
  const ownSidebarProjects = sidebarProjects.filter((p) => !p.visibleViaReport);
  const reportSidebarProjects = sidebarProjects.filter((p) => p.visibleViaReport);
  const orderedSidebarProjects = [...ownSidebarProjects, ...reportSidebarProjects];

  const myIssuesCount = tasks.filter(
    (t) =>
      isTaskMine(t.assigneeId) && t.status !== 'done' && t.status !== 'cancelled' && !t.archived
  ).length;
  const todayFocusCount = tasks.filter(
    (t) =>
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
      ? t('nav.inbox')
      : location.pathname.startsWith('/my-issues')
        ? t('flow.shell.myTasks')
        : location.pathname.startsWith('/today')
          ? t('nav.focusToday')
          : location.pathname.startsWith('/triage')
            ? t('flow.shell.triage')
            : location.pathname.startsWith('/calendar')
              ? t('flow.shell.calendar')
              : location.pathname.startsWith('/cycles')
                ? t('flow.shell.cycles')
                : location.pathname.startsWith('/milestones')
                  ? t('flow.shell.milestones')
                  : location.pathname.startsWith('/settings')
                    ? t('nav.settings')
                    : t('flow.shell.workspace');
  const lastRecentRoute = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (sidebarPreferencesLoading) return;
    if (lastRecentRoute.current === location.pathname) return;
    lastRecentRoute.current = location.pathname;
    updateSidebarPreferences(
      (draft) => recordRecentSidebarPage(draft, location.pathname, flowPageTitle),
      { showToast: false }
    );
  }, [flowPageTitle, location.pathname, sidebarPreferencesLoading, updateSidebarPreferences]);
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
          ...(activeProject
            ? { project_id: activeProject.id, project_name: activeProject.name }
            : {}),
        },
        visibleRows: sidebarProjects.slice(0, 12).map((project) => ({
          id: project.id,
          name: project.name,
          open_tasks: Math.max(0, (project.taskCount ?? 0) - (project.completedCount ?? 0)),
        })),
        hints: [
          t('flow.shell.assignedOpenTasks', { count: myIssuesCount }),
          t('flow.shell.focusItems', { count: todayFocusCount }),
          inQuietHours ? t('flow.shell.quietHoursActive') : undefined,
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
      sidebarProjects,
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
    mode === 'dark' ||
    (mode === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)
      ? '/logo.png'
      : '/logo-dark.png';

  return (
    // The single writer of the appearance tokens, mounted here rather than in
    // main.tsx because the workspace accent it falls back to is resolved here,
    // from the route params and the workspaces query. Its effect targets
    // documentElement, so the position in the tree governs when it writes and
    // never what it reaches; initAppearance in main.tsx covers the window
    // before this mounts.
    <AppearanceProvider
      prefs={appearance}
      workspaceAccent={activeWorkspace.color || ACCENT_DEFAULT}
    >
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
            aria-label={t('flow.shell.openMenu')}
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
            aria-label={t('common.search')}
            onClick={() => setCommandBarOpen(true)}
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <Search size={18} strokeWidth={2} />
          </button>
          <button
            type="button"
            aria-label={t('flow.shell.newTask')}
            onClick={() => setQuickCreateOpen(true)}
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--accent)' }}
          >
            <Plus size={20} strokeWidth={2} />
          </button>
          {shouldUseRealtime && <RealtimeConnectionIndicator status={realtimeStatus} compact />}
          <CaretButton />
        </header>

        {mobileNavOpen && (
          <button
            type="button"
            aria-label={t('flow.shell.closeMenu')}
            className="fixed inset-0 z-40 bg-black/50 lg:hidden"
            onClick={() => setMobileNavOpen(false)}
          />
        )}

        <aside
          className={cn(
            'fixed inset-y-0 start-0 z-50 flex h-full w-[min(18rem,88vw)] flex-col overflow-hidden border-e border-[var(--flow-border)] transition-transform duration-300 ease-out lg:static lg:inset-auto lg:z-auto lg:h-full lg:min-h-0 lg:w-[240px] lg:max-w-none lg:shrink-0 lg:translate-x-0',
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
              aria-label={t('flow.shell.closeMenu')}
              onClick={() => setMobileNavOpen(false)}
              className="ms-auto flex h-9 w-9 items-center justify-center rounded-lg lg:hidden hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <X size={18} strokeWidth={2} />
            </button>
          </div>

          {/* Workspace + search, shared inset, even vertical rhythm */}
          <div ref={workspaceSwitcherRef} className={cn('space-y-1.5 pt-1.5 pb-1.5 relative', sx)}>
            {/* The business name, not a switcher. Workspaces are gone, so there
                is nothing to switch between. */}
            <div
              className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <span
                className="h-4 w-4 rounded-sm shrink-0 flex items-center justify-center text-[9px] font-bold"
                style={{
                  background: 'var(--accent-muted)',
                  color: 'var(--accent-fg)',
                }}
                aria-hidden
              >
                {activeWorkspace.name.charAt(0)}
              </span>
              <span
                className="flex-1 text-start text-[12px] font-medium truncate"
                style={{ color: 'var(--flow-text)' }}
              >
                {activeWorkspace.name}
              </span>
            </div>

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
              <span className="flex-1 text-start">{t('flow.shell.searchPlaceholder')}</span>
              <span
                className="hidden text-[9px] font-mono px-1.5 py-0.5 rounded-sm sm:inline"
                style={{ background: 'var(--flow-active)' }}
              >
                {modKey}K
              </span>
            </button>
          </div>

          <SidebarNavigation
            preferences={sidebarPreferences}
            update={updateSidebarPreferences}
            showPreviewRoutes={showPreviewRoutes}
            unreadNotifications={unreadNotifications}
            myIssuesCount={myIssuesCount}
            todayFocusCount={todayFocusCount}
            projects={apiProjects ?? []}
            projectId={projectId}
            myTasksProjectId={myTasksProjectId}
            onNewProject={() => {
              setNewProjectName('');
              setNewProjectDescription('');
              setNewProjectColor('#64C5B9');
              setNewProjectOpen(true);
            }}
            sidebarRef={sidebarRef}
          />
          {false && (
            <div ref={sidebarRef} className={cn('flex-1 min-h-0 overflow-y-auto py-2', sx)}>
              <div className="flex flex-col gap-2">
                <nav className="flex flex-col gap-px">
                  <SidebarLink
                    to="/inbox"
                    icon={Inbox}
                    label={t('nav.inbox')}
                    badge={unreadNotifications}
                  />
                  <SidebarLink
                    to="/my-issues"
                    icon={ListTodo}
                    label={t('flow.shell.myTasks')}
                    badge={myIssuesCount}
                    exact
                  />
                  <SidebarLink
                    to="/today"
                    icon={Sun}
                    label={t('nav.focusToday')}
                    badge={todayFocusCount}
                  />
                  {/* Profile and Leave moved into the account menu at the
                    bottom (with the theme switcher and sign-out) - they are
                    account-scoped destinations, not work-item triage like
                    the entries around them. */}
                  <SidebarLink to="/projects" icon={LayoutGrid} label={t('nav.projects')} exact />
                  <SidebarLink to="/triage" icon={Eye} label={t('flow.shell.triage')} />
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
                                {t('flow.shell.plan')}
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
                                <SidebarLink
                                  to="/roadmap"
                                  icon={Map}
                                  label={t('flow.shell.roadmap')}
                                />
                              )}
                              <SidebarLink
                                to="/cycles"
                                icon={Timer}
                                label={t('flow.shell.cycles')}
                              />
                              <SidebarLink
                                to="/milestones"
                                icon={Milestone}
                                label={t('flow.shell.milestones')}
                              />
                              {showPreviewRoutes && (
                                <SidebarLink
                                  to="/timeline"
                                  icon={Timer}
                                  label={t('flow.shell.timeline')}
                                />
                              )}
                              <SidebarLink
                                to="/calendar"
                                icon={Calendar}
                                label={t('flow.shell.calendar')}
                              />
                              <SidebarLink to="/views" icon={Eye} label={t('flow.shell.views')} />
                              <SidebarLink
                                to="/archive"
                                icon={Archive}
                                label={t('flow.shell.archive')}
                              />
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
                                {t('flow.shell.team')}
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
                                <SidebarLink
                                  to="/team"
                                  icon={Users}
                                  label={t('flow.shell.workload')}
                                />
                              )}

                              {showPreviewRoutes && (
                                <SidebarLink
                                  to="/insights"
                                  icon={BarChart3}
                                  label={t('flow.shell.insights')}
                                />
                              )}
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
                              <SidebarLink
                                to="/time/team"
                                icon={Users}
                                label={t('flow.time.team')}
                              />
                              <SidebarLink
                                to="/time/reports"
                                icon={BarChart3}
                                label={t('flow.time.reports')}
                              />
                              <SidebarLink
                                to="/directory"
                                icon={Contact}
                                label={t('flow.directory.heading')}
                              />
                              <SidebarLink
                                to="/presence"
                                icon={Radar}
                                label={t('flow.presence.heading')}
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
                                {t('nav.projects')}
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
                              title={t('flow.shell.newProject')}
                              aria-label={t('flow.shell.newProject')}
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
                              {orderedSidebarProjects.map((project, index) => {
                                const boardPath = `/p/${project.id}`;
                                const isActive = projectId === project.id;
                                // The heading is emitted before the first proxied
                                // project rather than wrapping a second map, so the
                                // row markup below stays in one place.
                                const startsReports =
                                  reportSidebarProjects.length > 0 &&
                                  index === ownSidebarProjects.length;
                                return (
                                  <React.Fragment key={project.id}>
                                    {startsReports && (
                                      <div
                                        className="px-2 pb-px pt-2 text-[10px] font-semibold uppercase tracking-wider"
                                        style={{ color: 'var(--flow-text-muted)' }}
                                      >
                                        {t('flow.sidebar.yourReports')}
                                      </div>
                                    )}
                                    <Link
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
                                          transition={{
                                            type: 'spring',
                                            stiffness: 400,
                                            damping: 35,
                                          }}
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
                                        {(project.taskCount ?? 0) - (project.completedCount ?? 0)}
                                      </span>
                                    </Link>
                                  </React.Fragment>
                                );
                              })}
                            </div>
                          </SidebarCollapsible>
                        </div>
                      )}
                    </React.Fragment>
                  ))}
                </div>

                {/* The bin renders itself away when empty, so it costs nothing
                  here. It was imported but never mounted, which left the whole
                  recycle bin unreachable from the UI. */}
                <RecycleBinSection />
              </div>
            </div>
          )}

          {/* Footer */}
          <div
            className={cn('flex flex-col gap-1.5 py-2', sx)}
            style={{ borderTop: '1px solid var(--flow-border)' }}
          >
            {shouldUseRealtime && <RealtimeConnectionIndicator status={realtimeStatus} />}
            <SidebarLink to="/settings" icon={Settings} label={t('nav.settings')} />
            <CaretButton className="w-full justify-start" />

            <div className="relative" ref={userMenuRef}>
              <button
                type="button"
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
                        {t('flow.shell.appearance')}
                      </p>
                      <ModeToggle />
                    </div>
                    <div
                      style={{
                        height: '1px',
                        background: 'var(--flow-border)',
                        margin: '4px 0',
                      }}
                    />
                    <Link
                      to="/u/$userId"
                      params={{ userId: user?.id ?? '' }}
                      onClick={() => setUserMenuOpen(false)}
                      className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-start transition-colors hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text)' }}
                    >
                      <UserCircle size={13} />
                      {t('flow.profile.heading')}
                    </Link>
                    <Link
                      to="/leave"
                      onClick={() => setUserMenuOpen(false)}
                      className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-start transition-colors hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text)' }}
                    >
                      <CalendarOff size={13} />
                      {t('flow.leave.heading')}
                    </Link>
                    <div
                      style={{
                        height: '1px',
                        background: 'var(--flow-border)',
                        margin: '4px 0',
                      }}
                    />
                    <button
                      type="button"
                      onClick={handleLogout}
                      className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-start transition-colors hover:bg-[var(--flow-hover)]"
                      style={{ color: '#ef4444' }}
                    >
                      <LogOut size={13} />
                      {t('flow.shell.signOut')}
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
                aria-label={t('common.close')}
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
                  {t('flow.shell.newProject')}
                </p>
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-2">
                    <input
                      type="color"
                      value={newProjectColor}
                      onChange={(e) => setNewProjectColor(e.target.value)}
                      title={t('flow.shell.pickColor')}
                      className="h-8 w-8 shrink-0 rounded-md cursor-pointer border-0 p-0.5"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border)',
                      }}
                    />
                    <input
                      autoFocus
                      type="text"
                      placeholder={t('flow.shell.projectNamePlaceholder')}
                      value={newProjectName}
                      onChange={(e) => setNewProjectName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          if (!newProjectName.trim()) {
                            toast.error(t('flow.shell.projectNameRequired'));
                            return;
                          }
                          if (!canCreateProject) return;
                          createProject.mutate(
                            {
                              name: newProjectName.trim(),
                              color: newProjectColor,
                            },
                            {
                              onSuccess: () => setNewProjectOpen(false),
                              onError: () => toast.error(t('errors.generic')),
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
                    placeholder={t('flow.shell.projectDescriptionPlaceholder')}
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
                {/* Visibility toggle */}
                <div className="flex items-center justify-between mt-3">
                  <label className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.visibility.label')}
                  </label>
                  <div
                    className="flex gap-1"
                    role="radiogroup"
                    aria-label={t('flow.visibility.label')}
                  >
                    {(['private', 'public'] as const).map((vis) => (
                      <button
                        key={vis}
                        type="button"
                        role="radio"
                        aria-checked={newProjectVisibility === vis}
                        onClick={() => setNewProjectVisibility(vis)}
                        className="px-3 h-8 rounded-lg text-[11px] font-semibold transition-colors"
                        style={{
                          background:
                            newProjectVisibility === vis ? 'var(--accent)' : 'var(--flow-input-bg)',
                          color:
                            newProjectVisibility === vis ? '#0a0a0a' : 'var(--flow-text-muted)',
                          border: '1px solid var(--flow-border)',
                        }}
                      >
                        {t(`flow.visibility.${vis}`)}
                      </button>
                    ))}
                  </div>
                </div>
                {newProjectVisibility === 'public' && (
                  <p className="mt-1.5 text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.visibility.publicHint')}
                  </p>
                )}
                <div className="mt-4 flex flex-col gap-2">
                  <GateNotice
                    id="flow-new-project-gate"
                    reason={createProjectGateReason}
                    className="justify-end text-end"
                  />
                  <div className="flex justify-end gap-2">
                    <button
                      type="button"
                      onClick={() => setNewProjectOpen(false)}
                      className="px-3 py-1.5 rounded-lg text-[12px] font-medium transition-colors hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {t('common.cancel')}
                    </button>
                    <button
                      type="button"
                      disabled={
                        !newProjectName.trim() || createProject.isPending || !canCreateProject
                      }
                      aria-describedby={canCreateProject ? undefined : 'flow-new-project-gate'}
                      onClick={() => {
                        if (!newProjectName.trim() || !canCreateProject) return;
                        createProject.mutate(
                          {
                            name: newProjectName.trim(),
                            color: newProjectColor,
                          },
                          {
                            onSuccess: () => setNewProjectOpen(false),
                            onError: () => toast.error(t('errors.generic')),
                          }
                        );
                      }}
                      className="px-4 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40 disabled:cursor-not-allowed"
                      style={{ background: 'var(--accent)', color: '#0a0a0a' }}
                    >
                      {createProject.isPending ? t('flow.shell.creating') : t('common.create')}
                    </button>
                  </div>
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
                {t('flow.shell.quietHoursNotice', {
                  start: lifePrefs.quietStart,
                  end: lifePrefs.quietEnd,
                  digestHour: lifePrefs.digestHour,
                })}
              </div>
            )}
            <LicenseBanner client={licenseClient} />
            {/* mode="wait" (the default AnimatePresence sequencing) would unmount
                the outgoing <main> before mounting the incoming one - since this
                is keyed on the full pathname, that gap opens on every navigation
                between two projects, not just section changes. During it there
                is no <main> in the DOM at all, so the layout's own background
                (var(--flow-bg), near-black in dark mode) shows through with
                nothing on top - a solid black flash until the incoming main
                mounts. Overlapping the two (no mode prop) keeps the outgoing
                page visible while the incoming one fades in underneath it. */}
            <AnimatePresence>
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
    </AppearanceProvider>
  );
};

export default FlowLayout;

// Same reason: the capacity query has no data while /hr/work-schedule 403s,
// so a `= {}` destructuring default would mint a new object every render.
const EMPTY_CAPACITY: Record<string, number> = {};
