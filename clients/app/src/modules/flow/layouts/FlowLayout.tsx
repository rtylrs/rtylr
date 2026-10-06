import { useEffect, useRef, useState, type ReactNode } from 'react';
import { Link, useLocation, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Menu, PanelLeft, Plus, UserCircle, X } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@voxire/ui';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { toast } from 'sonner';
import { useUserStore } from '../../../shared/stores/userStore';
import { Button } from '../../../shared/components/ui/button';
import { usePortalHeaderActions } from '../../../shell/PortalHeaderActions';
import { useSidebarCollapse } from '../../../shell/SidebarCollapseContext';
import { canAccessPortalPath } from '../../../shared/utils/canAccessPortalPath';
import { useFlowStore, type Member } from '../shared/stores/flowStore';
import { useProjects, useCreateProject } from '../features/projects/useProjects';
import { useMyTasksProject } from '../features/tasks/useMyTasksProject';
import { useMyIssues } from '../features/tasks/useMyIssues';
import { useFlowTeam } from '../shared/hooks/queries/useFlowTeam';
import { useMemberCapacityHours } from '../shared/hooks/queries/useMemberCapacity';
import { useNotifications } from '../features/inbox/useNotifications';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useLayoutRealtime } from '../shared/lib/useLayoutRealtime';
import { useRealtimeReconnectRefetch, useRealtimeStatus } from '../shared/lib/realtime';
import { apiProjectToProject } from '../shared/lib/projectMapper';
import { apiTaskToTask } from '../shared/lib/taskMapper';
import { initialsFromName } from '../shared/lib/memberUtils';
import { canFlowPermission } from '../shared/lib/roleUtils';
import { isTaskMine } from '../shared/lib/currentUser';
import { isDueToday, isQuietNow } from '../shared/lib/lifeRhythm';
import { useLifePreferences } from '../shared/lib/lifePreferences';
import { useSidebarPreferences } from '../shared/lib/sidebarPreferences';
import { ACCENT_DEFAULT } from '../shared/lib/colorUtils';
import { SidebarNavigation } from '../shared/components/SidebarNavigation';
import RealtimeConnectionIndicator from '../shared/components/RealtimeConnectionIndicator';
import TrialBanner from '../shared/components/TrialBanner';
import GateNotice from '../shared/components/GateNotice';
import BulkEditBar from '../shared/components/BulkEditBar';
import TaskDrawer from '../shared/components/TaskDrawer';

export default function FlowLayout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const headerActions = usePortalHeaderActions();
  const { collapsed, setCollapsed } = useSidebarCollapse();
  const user = useUserStore((state) => state.user);
  const location = useLocation();
  const { projectId } = useParams({ strict: false });
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const { preferences, update } = useSidebarPreferences();
  const { prefs: lifePrefs } = useLifePreferences();
  const {
    setProjects,
    setMembers,
    setQuickCreateOpen,
    projectCreateOpen: projectOpen,
    setProjectCreateOpen: setProjectOpen,
  } = useFlowStore();
  const { data: projects } = useProjects({
    enabled: canFlowPermission('flow:projects:view', user?.effectiveAccess),
  });
  const { data: personalProject } = useMyTasksProject({
    enabled: canFlowPermission('flow:tasks:view', user?.effectiveAccess),
  });
  const { data: myTasks = [] } = useMyIssues(
    canFlowPermission('flow:tasks:view', user?.effectiveAccess)
  );
  const { data: members } = useFlowTeam();
  const { data: capacity } = useMemberCapacityHours();
  const { data: notifications = [] } = useNotifications();
  const createProject = useCreateProject();
  const sidebarRef = useRef<HTMLDivElement>(null);
  const mobileRef = useRef<HTMLDialogElement>(null);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState(ACCENT_DEFAULT);
  const [visibility, setVisibility] = useState<'private' | 'public'>('private');
  const canCreateProject = canFlowPermission('flow:projects:create', user?.effectiveAccess);
  const canCreateTask = canFlowPermission('flow:tasks:create', user?.effectiveAccess);
  const visiblePreferences = {
    ...preferences,
    items: preferences.items.map((item) => ({
      ...item,
      visible: item.visible && (!item.route || canAccessPortalPath(user, item.route.split('?')[0])),
    })),
    groups: preferences.groups.map((group) => ({
      ...group,
      visible:
        group.visible && (group.id !== 'projects' || canAccessPortalPath(user, '/flow/projects')),
    })),
  };
  const gateReason = canCreateProject ? '' : t('flow.shell.createNeedsRole');
  const userId = user?.id === undefined ? undefined : String(user.id);
  const realtimeEnabled = true;
  useLayoutRealtime(
    user?.businessId === undefined ? undefined : String(user.businessId),
    userId,
    realtimeEnabled
  );
  useRealtimeReconnectRefetch();
  const realtimeStatus = useRealtimeStatus();
  useKeyboardShortcuts();

  useEffect(() => {
    if (projects) setProjects(projects.map(apiProjectToProject));
  }, [projects, setProjects]);
  useEffect(() => {
    if (!members) return;
    const mapped: Member[] = members
      .filter((member) => member.isActive)
      .map((member) => ({
        id: String(member.id),
        name: member.name,
        email: member.email,
        initials: initialsFromName(member.name),
        capacity: capacity?.[String(member.id)],
      }));
    setMembers(mapped);
  }, [members, capacity, setMembers]);
  useEffect(() => {
    if (projectOpen) {
      setName('');
      setDescription('');
      setColor(ACCENT_DEFAULT);
      setVisibility('private');
    }
  }, [projectOpen]);
  useEffect(() => setMobileOpen(false), [location.pathname]);
  useEffect(() => {
    if (mobileOpen && !mobileRef.current?.open) mobileRef.current?.showModal();
    if (!mobileOpen && mobileRef.current?.open) mobileRef.current.close();
  }, [mobileOpen]);

  const activeTasks = myTasks
    .map(apiTaskToTask)
    .filter(
      (task) =>
        !task.archived &&
        task.status !== 'done' &&
        task.status !== 'cancelled' &&
        isTaskMine(task.assigneeId)
    );
  const navigation = (mobile = false) => (
    <>
      <div className="portal-flow-sidebar-header">
        <Link to="/" className="portal-wordmark" aria-label="RTYLR">
          <img className="portal-logo-light" src="/logo-dark.png" alt="RTYLR" />
          <img className="portal-logo-dark" src="/logo.png" alt="RTYLR" />
        </Link>
        <div className="flex items-center gap-1">
          {!mobile && (
            <button
              type="button"
              className="portal-icon-button"
              onClick={() => setCollapsed(true)}
              aria-label={t('common.portal.collapseSidebar')}
              aria-expanded={!collapsed}
              aria-controls="portal-flow-sidebar"
              title={t('common.portal.collapseSidebar')}
            >
              <PanelLeft size={18} />
            </button>
          )}
          {canCreateTask && (
            <button
              type="button"
              className="portal-icon-button"
              onClick={() => setQuickCreateOpen(true)}
              aria-label={t('flow.shell.newTask')}
              title={t('flow.shell.newTask')}
            >
              <Plus size={18} />
            </button>
          )}
          {mobile && (
            <button
              type="button"
              className="portal-icon-button"
              onClick={() => setMobileOpen(false)}
              aria-label={t('flow.shell.closeMenu')}
            >
              <X size={18} />
            </button>
          )}
        </div>
      </div>
      {realtimeEnabled && (
        <div className="px-4">
          <RealtimeConnectionIndicator status={realtimeStatus} compact />
        </div>
      )}
      <SidebarNavigation
        preferences={visiblePreferences}
        update={update}
        showPreviewRoutes={showPreviewRoutes}
        unreadNotifications={notifications.filter((notification) => !notification.read).length}
        myIssuesCount={activeTasks.length}
        todayFocusCount={
          activeTasks.filter((task) => task.focusPin || isDueToday(task.dueDate)).length
        }
        projects={projects ?? []}
        projectId={projectId}
        myTasksProjectId={
          personalProject?.projectId === undefined ? undefined : String(personalProject.projectId)
        }
        onNewProject={() => setProjectOpen(true)}
        canCreateProject={canCreateProject}
        sidebarRef={sidebarRef}
      />
      <nav className="portal-flow-footer" aria-label={t('flowNav.settings')}>
        <Link to="/flow/u/$userId" params={{ userId: userId ?? '' }} className="portal-nav-link">
          <UserCircle size={15} />
          {t('flow.profile.heading')}
        </Link>
      </nav>
    </>
  );
  const submitProject = () => {
    if (!name.trim() || !canCreateProject || createProject.isPending) return;
    createProject.mutate(
      { name: name.trim(), description: description.trim() || undefined, color, visibility },
      {
        onSuccess: () => setProjectOpen(false),
        onError: () => toast.error(t('flowErrors.generic')),
      }
    );
  };
  return (
    <div className="portal-flow">
      {!collapsed && (
        <aside className="portal-flow-sidebar" id="portal-flow-sidebar">
          {navigation()}
        </aside>
      )}
      <div className="portal-flow-content">
        <header className="portal-topbar portal-flow-header">
          <div className="portal-topbar-start">
            {collapsed && (
              <button
                type="button"
                className="portal-icon-button portal-flow-show-sidebar"
                onClick={() => setCollapsed(false)}
                aria-label={t('common.toggleNavigation')}
                aria-expanded={!collapsed}
                aria-controls="portal-flow-sidebar"
                title={t('common.toggleNavigation')}
              >
                <PanelLeft size={19} />
              </button>
            )}
            <span>{t('appNames.short.flow')}</span>
          </div>
          {headerActions}
        </header>
        <div className="portal-flow-mobile-trigger px-3 py-2">
          <button
            type="button"
            className="portal-icon-button"
            onClick={() => setMobileOpen(true)}
            aria-label={t('flow.shell.openMenu')}
          >
            <Menu size={19} />
          </button>
          {canCreateTask && (
            <button
              type="button"
              className="portal-icon-button"
              onClick={() => setQuickCreateOpen(true)}
              aria-label={t('flow.shell.newTask')}
            >
              <Plus size={19} />
            </button>
          )}
        </div>
        <TrialBanner />
        {isQuietNow(lifePrefs.quietStart, lifePrefs.quietEnd) && (
          <p
            className="px-4 py-2 text-center text-flow-meta"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
          >
            {t('flow.shell.quietHoursNotice', {
              start: lifePrefs.quietStart,
              end: lifePrefs.quietEnd,
              digestHour: lifePrefs.digestHour,
            })}
          </p>
        )}
        <div className="portal-flow-page">{children}</div>
      </div>
      <dialog
        ref={mobileRef}
        className="portal-flow-mobile-nav"
        onCancel={() => setMobileOpen(false)}
        onClose={() => setMobileOpen(false)}
      >
        {mobileOpen && navigation(true)}
      </dialog>
      <Dialog open={projectOpen} onOpenChange={setProjectOpen}>
        <DialogContent className="portal-flow-dialog max-w-sm" closeLabel={t('flowCommon.close')}>
          <DialogTitle>{t('flow.shell.newProject')}</DialogTitle>
          <DialogDescription className="sr-only">
            {t('flow.shell.projectNamePlaceholder')}
          </DialogDescription>
          <form
            className="space-y-3"
            onSubmit={(event) => {
              event.preventDefault();
              submitProject();
            }}
          >
            <div className="flex gap-2">
              <input
                type="color"
                value={color}
                onChange={(event) => setColor(event.target.value)}
                aria-label={t('flow.shell.pickColor')}
                className="h-10 w-10 rounded-md"
              />
              <input
                autoFocus
                value={name}
                onChange={(event) => setName(event.target.value)}
                aria-label={t('flow.shell.projectNamePlaceholder')}
                placeholder={t('flow.shell.projectNamePlaceholder')}
                className="min-w-0 flex-1 rounded-lg border border-[var(--flow-border-control)] bg-[var(--flow-input-bg)] px-3 py-2 text-flow-body"
              />
            </div>
            <textarea
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              aria-label={t('flow.shell.projectDescriptionPlaceholder')}
              placeholder={t('flow.shell.projectDescriptionPlaceholder')}
              rows={3}
              className="w-full rounded-lg border border-[var(--flow-border-control)] bg-[var(--flow-input-bg)] px-3 py-2 text-flow-body"
            />
            <fieldset>
              <legend className="mb-2 text-flow-meta">{t('flow.visibility.label')}</legend>
              <div className="flex gap-3">
                {(['private', 'public'] as const).map((value) => (
                  <label key={value} className="flex items-center gap-2 text-flow-body">
                    <input
                      type="radio"
                      name="visibility"
                      value={value}
                      checked={visibility === value}
                      onChange={() => setVisibility(value)}
                    />
                    {t(`flow.visibility.${value}`)}
                  </label>
                ))}
              </div>
              {visibility === 'public' && (
                <p className="mt-2 text-flow-meta text-[var(--flow-text-muted)]">
                  {t('flow.visibility.publicHint')}
                </p>
              )}
            </fieldset>
            <GateNotice id="flow-new-project-gate" reason={gateReason} />
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => setProjectOpen(false)}>
                {t('flowCommon.cancel')}
              </Button>
              <Button
                type="submit"
                disabled={!name.trim() || !canCreateProject || createProject.isPending}
                aria-describedby={canCreateProject ? undefined : 'flow-new-project-gate'}
              >
                {createProject.isPending ? t('flow.shell.creating') : t('flowCommon.create')}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
      <BulkEditBar />
      <TaskDrawer />
    </div>
  );
}
