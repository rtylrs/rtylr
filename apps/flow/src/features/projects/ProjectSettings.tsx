import React, { useRef, useState, useEffect } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  MessageSquare,
  GitBranch,
  Mail,
  Trash2,
  Save,
  Plus,
  Crown,
  Shield,
  Eye,
  ArrowRightLeft,
  Users,
} from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useUpdateProject, useArchiveProject, useProjects } from './useProjects';
import { DEFAULT_PROJECT_ICON } from '../../shared/services/flowApi';
import {
  useProjectMembers,
  useRemoveProjectMember,
  useUpdateProjectMemberRole,
  useTransferProjectOwnership,
} from './useProjectMembers';
import { HelmetContainer } from '@voxire/components';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  EmptyState,
} from '@voxire/ui';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import GateNotice from '../../shared/components/GateNotice';
import AddProjectMemberDialog from './components/AddProjectMemberDialog';
import { displayDescription } from '../../shared/lib/stringUtils';
import { useCanAdministerProject } from '../../shared/lib/currentUser';
import ProjectVisibilityField from './components/ProjectVisibilityField';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const ROLE_META: Record<string, { icon: React.ReactNode; color: string }> = {
  admin: { icon: <Crown size={10} />, color: 'var(--flow-warn-fg)' },
  member: { icon: <Shield size={10} />, color: 'var(--accent)' },
  viewer: { icon: <Eye size={10} />, color: 'var(--flow-neutral-fg)' },
};

const ProjectSettings: React.FC = () => {
  const { t } = useTranslation();
  const { projectId } = useParams({ strict: false });
  const navigate = useNavigate();
  const { projects } = useFlowStore();
  const project = projects.find((p) => p.id === projectId);
  const projectsQuery = useProjects();

  const [name, setName] = useState(project?.name || '');
  const [description, setDescription] = useState(displayDescription(project?.description));
  const [tab, setTab] = useState<'general' | 'members' | 'integrations' | 'danger'>('general');

  // Projects load async into the store and the same component instance is
  // reused when switching projects, so sync the form when the project changes.
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(displayDescription(project.description));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [project?.id]);

  const updateProject = useUpdateProject();
  const archiveProject = useArchiveProject();
  const {
    data: members = [],
    isLoading: membersLoading,
    refetch: refetchMembers,
  } = useProjectMembers(projectId);
  const { data: orgMembers = [] } = useFlowTeam();
  const removeMember = useRemoveProjectMember(projectId ?? '');
  const updateRole = useUpdateProjectMemberRole(projectId ?? '');
  const transferOwnership = useTransferProjectOwnership();

  // Every write on this screen (rename, delete, transfer, member management)
  // is project-structural, so all of it is gated the same way ProjectBoard
  // gates board management: project admin+ only. A project viewer or member
  // reaching this screen sees it, but every control on it is disabled with
  // a GateNotice rather than actually usable - this is the fix for #793,
  // where none of these actions had any check at all.
  const canManageProject = useCanAdministerProject(projectId);

  const transferRef = useRef<HTMLSelectElement>(null);
  const [confirmTransfer, setConfirmTransfer] = useState(false);

  const [showAddMember, setShowAddMember] = useState(false);

  // Type-to-confirm modal state for project deletion. Native confirm()
  // is too easy to click through; the user has to retype the project
  // name to enable the destructive button.
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleteTypedName, setDeleteTypedName] = useState('');

  if (!projectId) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.loading')}
      </div>
    );
  }

  // The store is filled by the projects query, so a failed fetch empties it
  // exactly as a deleted project would, and settings reported the project gone.
  if (hasQueryFailed(projectsQuery)) {
    return <ListLoadFailed onRetry={() => void projectsQuery.refetch()} />;
  }

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.notFound')}
      </div>
    );
  }

  // Translated, and sentence case. All four labels were hardcoded English on
  // a screen that ships in three languages, and "Danger Zone" was Title Case
  // among single-word siblings.
  const tabs = [
    { id: 'general', label: t('flow.projectSettings.tabs.general') },
    { id: 'members', label: t('flow.projectSettings.tabs.members') },
    { id: 'integrations', label: t('flow.projectSettings.tabs.integrations') },
    { id: 'danger', label: t('flow.projectSettings.tabs.danger') },
  ] as const;

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={t('flow.projectSettings.pageTitle')} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <Breadcrumbs
          // The trail stops at the parent. The h1 directly below already says
          // Settings, and repeating it in the crumb put the same word twice in
          // two lines (#1097, D-029).
          items={[{ label: project.name, to: `/p/${project.id}/overview` }]}
        />
        <h1 className="text-title font-bold mt-1" style={{ color: 'var(--flow-text)' }}>
          {t('nav.settings')}
        </h1>
        <div className="flex gap-4 mt-3">
          {tabs.map((t) => (
            <button
              type="button"
              key={t.id}
              onClick={() => setTab(t.id)}
              className="text-meta pb-1 border-b-2 transition-colors"
              style={{
                color: tab === t.id ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                borderColor: tab === t.id ? 'var(--accent)' : 'transparent',
              }}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      {/* 640px, per S-11. max-w-2xl is 672, and there is no Tailwind step at
          640, so the value is written out. The empty space to the right of a
          capped form is the point of capping it: a 1440px-wide text input is
          harder to read, not easier. */}
      <div className="p-6 max-w-[640px] space-y-6">
        {/* ── General ── */}
        {tab === 'general' && (
          <>
            {!canManageProject && (
              <GateNotice
                id="flow-project-general-gate"
                reason={t('flow.projectSettings.generalGateReason')}
              />
            )}
            <SectionHeading>{t('flow.projectSettings.detailsSection')}</SectionHeading>
            <Field label={t('common.name')}>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={!canManageProject}
                aria-describedby={canManageProject ? undefined : 'flow-project-general-gate'}
                className="w-full rounded-md px-3 py-2 text-meta disabled:opacity-60"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              />
            </Field>
            <Field label={t('common.description')}>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t('flow.projectSettings.descriptionPlaceholder')}
                rows={3}
                disabled={!canManageProject}
                aria-describedby={canManageProject ? undefined : 'flow-project-general-gate'}
                className="w-full rounded-md px-3 py-2 text-meta resize-none disabled:opacity-60"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              />
            </Field>

            <ProjectVisibilityField projectId={projectId} canAdminister={canManageProject} />

            <button
              type="button"
              onClick={() => {
                if (!canManageProject) return;
                updateProject.mutate(
                  {
                    id: projectId!,
                    data: {
                      // The project endpoint is a full PUT. Include the
                      // untouched snapshot fields so a name/description edit
                      // cannot erase dates or other persisted settings.
                      name,
                      // UpdateProjectBody.description is @required, so a
                      // cleared/never-entered field must still resolve to the
                      // '-' sentinel on the wire - the same convention
                      // createProject uses. The field only ever displays as
                      // empty; it must not save as an empty string.
                      description: description.trim() || '-',
                      // Not 'folder'. The update body is a replacement, so a
                      // project with no icon had a lucide component name
                      // written into a column every reader renders as a glyph,
                      // and the settings page then showed that word.
                      icon: project.icon ?? DEFAULT_PROJECT_ICON,
                      color: project.color,
                      status: project.status ?? 'active',
                      startDate: project.startDate,
                      dueDate: project.targetDate,
                    },
                  },
                  {
                    onSuccess: () => toast.success(t('toasts.project.updated')),
                  }
                );
              }}
              disabled={updateProject.isPending || !canManageProject}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-40"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              <Save size={11} />{' '}
              {updateProject.isPending ? t('common.saving') : t('common.saveChanges')}
            </button>
          </>
        )}

        {/* ── Members ── */}
        {tab === 'members' && (
          <div className="space-y-3">
            {!canManageProject && (
              <GateNotice reason={t('flow.projectSettings.membersGateReason')} />
            )}
            <div className="flex items-center justify-between">
              <p
                className="text-meta font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {members.length} member{members.length !== 1 ? 's' : ''}
              </p>
              {canManageProject && (
                <button
                  type="button"
                  onClick={() => setShowAddMember(true)}
                  className="flex items-center gap-1 text-meta px-2 py-1 rounded-sm"
                  style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
                >
                  <Plus size={10} /> Add
                </button>
              )}
            </div>

            {canManageProject && (
              <AddProjectMemberDialog
                open={showAddMember}
                onOpenChange={setShowAddMember}
                projectId={projectId}
              />
            )}

            {membersLoading ? (
              <div className="space-y-2">
                {[1, 2].map((i) => (
                  <div
                    key={i}
                    className="h-12 rounded-lg animate-pulse"
                    style={{ background: 'var(--flow-hover)' }}
                  />
                ))}
              </div>
            ) : members.length === 0 ? (
              <EmptyState
                icon={Users}
                title={t('flow.projectSettings.noMembersYet')}
                action={
                  canManageProject
                    ? {
                        label: t('flow.projectSettings.addMemberCta'),
                        onClick: () => setShowAddMember(true),
                      }
                    : {
                        label: t('flow.projectSettings.refreshMembersCta'),
                        onClick: () => void refetchMembers(),
                      }
                }
                size="compact"
              />
            ) : (
              <div className="space-y-2">
                {members.map((m) => {
                  const roleMeta = ROLE_META[m.role] ?? ROLE_META.member;
                  const initials = m.name
                    ? m.name
                        .split(' ')
                        .map((n: string) => n[0])
                        .join('')
                        .toUpperCase()
                        .slice(0, 2)
                    : m.email.slice(0, 2).toUpperCase();
                  return (
                    <div
                      key={m.id}
                      className="flex items-center gap-3 rounded-md px-3 py-2"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border-control)',
                      }}
                    >
                      <span
                        className="h-7 w-7 rounded-full flex items-center justify-center text-meta font-bold shrink-0"
                        style={{
                          background: 'var(--accent-muted)',
                          color: 'var(--accent-on-muted)',
                        }}
                      >
                        {initials}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p
                          className="text-meta font-medium truncate"
                          style={{ color: 'var(--flow-text)' }}
                        >
                          {m.name || m.email}
                        </p>
                        {m.name && (
                          <p
                            className="text-meta truncate"
                            style={{ color: 'var(--flow-text-dim)' }}
                          >
                            {m.email}
                          </p>
                        )}
                      </div>
                      <div className="flex items-center gap-1" style={{ color: roleMeta.color }}>
                        {roleMeta.icon}
                        <select
                          value={m.role}
                          disabled={!canManageProject}
                          aria-label={t('flow.projectSettings.roleSelectLabel', {
                            name: m.name || m.email,
                          })}
                          onChange={(e) => {
                            if (!canManageProject) return;
                            updateRole.mutate(
                              {
                                project_id: Number(projectId),
                                user_id: Number(m.userId),
                                role: e.target.value,
                              },
                              {
                                onSuccess: () =>
                                  toast.success(t('toasts.projectMember.roleUpdated')),
                              }
                            );
                          }}
                          className="text-meta bg-transparent cursor-pointer disabled:cursor-not-allowed disabled:opacity-60"
                          style={{ color: roleMeta.color }}
                        >
                          <option value="admin">{t('common.roles.admin')}</option>
                          <option value="member">{t('common.roles.member')}</option>
                          <option value="viewer">{t('common.roles.viewer')}</option>
                        </select>
                      </div>
                      <button
                        type="button"
                        onClick={() => {
                          if (!canManageProject) return;
                          removeMember.mutate(
                            { project_id: Number(projectId), user_id: Number(m.userId) },
                            {
                              onSuccess: () => toast.success(t('toasts.projectMember.removed')),
                            }
                          );
                        }}
                        disabled={!canManageProject}
                        className="h-6 w-6 rounded-sm flex items-center justify-center shrink-0 disabled:opacity-40 disabled:cursor-not-allowed"
                        style={{ color: 'var(--flow-text-dim)' }}
                        onMouseEnter={(e) => {
                          if (canManageProject)
                            e.currentTarget.style.color = 'var(--flow-danger-fg)';
                        }}
                        onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--flow-text-dim)')}
                        title={t('flow.projectSettings.removeMemberLabel')}
                        aria-label={t('flow.projectSettings.removeMemberLabel')}
                      >
                        <Trash2 size={10} />
                      </button>
                    </div>
                  );
                })}
              </div>
            )}

            <div className="pt-4 space-y-4" style={{ borderTop: '1px solid var(--flow-border)' }}>
              {/* Transfer ownership */}
              <div>
                <p
                  className="text-meta font-semibold mb-1"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.projectSettings.transferHeading')}
                </p>
                <p className="text-meta mb-2" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.projectSettings.transferDescription')}
                </p>
                {!canManageProject && (
                  <GateNotice
                    id="flow-project-transfer-gate"
                    className="mb-2"
                    reason={t('flow.projectSettings.transferGateReason')}
                  />
                )}
                <div className="flex gap-2">
                  <select
                    ref={transferRef}
                    id="transfer-select"
                    disabled={!canManageProject}
                    aria-describedby={canManageProject ? undefined : 'flow-project-transfer-gate'}
                    onChange={() => setConfirmTransfer(false)}
                    className="flex-1 rounded-sm px-2 py-1.5 text-meta disabled:opacity-60"
                    style={{
                      background: 'var(--flow-input-bg)',
                      border: '1px solid var(--flow-border-control)',
                      color: 'var(--flow-text)',
                    }}
                    defaultValue=""
                  >
                    <option value="" disabled>
                      {t('flow.projectSettings.selectNewOwner')}
                    </option>
                    {orgMembers.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name || m.email}
                      </option>
                    ))}
                  </select>
                  {confirmTransfer ? (
                    <>
                      <button
                        type="button"
                        onClick={() => {
                          if (!canManageProject) return;
                          const newOwnerId = Number(transferRef.current?.value);
                          transferOwnership.mutate(
                            { project_id: Number(projectId), new_owner_id: newOwnerId },
                            {
                              onSuccess: () => {
                                toast.success(t('toasts.projectMember.transferred'));
                                setConfirmTransfer(false);
                              },
                              onError: () => setConfirmTransfer(false),
                            }
                          );
                        }}
                        disabled={transferOwnership.isPending || !canManageProject}
                        className="px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-40"
                        style={{
                          background: 'var(--flow-danger-fill)',
                          color: 'var(--flow-danger-on-fill)',
                        }}
                      >
                        {transferOwnership.isPending
                          ? t('flow.projectSettings.transferring')
                          : t('common.confirm')}
                      </button>
                      <button
                        type="button"
                        onClick={() => setConfirmTransfer(false)}
                        className="px-3 py-1.5 rounded-sm text-meta"
                        style={{ color: 'var(--flow-text-muted)' }}
                      >
                        {t('common.cancel')}
                      </button>
                    </>
                  ) : (
                    <button
                      type="button"
                      onClick={() => {
                        if (!canManageProject) return;
                        const newOwnerId = Number(transferRef.current?.value);
                        if (!newOwnerId)
                          return toast.error(t('toasts.projectMember.selectNewOwner'));
                        setConfirmTransfer(true);
                      }}
                      disabled={!canManageProject}
                      className="flex items-center gap-1 px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-40"
                      style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
                    >
                      <ArrowRightLeft size={10} /> {t('flow.projectSettings.transferCta')}
                    </button>
                  )}
                </div>
                {confirmTransfer && (
                  <p className="text-meta mt-1.5" style={{ color: 'var(--flow-warn-fg)' }}>
                    {t('flow.projectSettings.transferConfirmHint')}
                  </p>
                )}
              </div>
            </div>
          </div>
        )}

        {/* ── Integrations ── */}
        {tab === 'integrations' && (
          <div className="space-y-2">
            {[
              {
                icon: MessageSquare,
                name: 'Slack',
                description: 'Post project updates to #flow-updates',
              },
              { icon: GitBranch, name: 'GitHub', description: 'Link PRs to tasks automatically' },
              { icon: Mail, name: 'Email', description: 'Send daily digests to members' },
            ].map(({ icon: Icon, name, description }) => (
              <div
                key={name}
                className="flex items-center gap-3 rounded-md px-3 py-2.5"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                }}
              >
                <Icon size={14} style={{ color: 'var(--flow-text-muted)' }} />
                <div className="flex-1">
                  <p className="text-meta font-medium" style={{ color: 'var(--flow-text)' }}>
                    {name}
                  </p>
                  <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                    {description}
                  </p>
                </div>
                <button
                  type="button"
                  className="text-meta px-2 py-1 rounded-sm hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-muted)' }}
                  disabled
                  title={t('common.comingSoon')}
                >
                  {t('common.comingSoon')}
                </button>
              </div>
            ))}
          </div>
        )}

        {/* ── Danger Zone ── */}
        {tab === 'danger' && (
          <div className="space-y-2">
            {!canManageProject && (
              <GateNotice
                id="flow-project-delete-gate"
                reason={t('flow.projectSettings.deleteGateReason')}
              />
            )}
            <div
              className="rounded-md p-4 flex items-start gap-3"
              style={{
                background: 'rgba(239,68,68,0.06)',
                border: '1px solid rgba(239,68,68,0.25)',
              }}
            >
              <Trash2 size={14} style={{ color: 'var(--flow-danger-fg)' }} className="mt-0.5" />
              <div className="flex-1">
                <p className="text-meta font-semibold" style={{ color: 'var(--flow-danger-fg)' }}>
                  {t('flow.projectSettings.deleteHeading')}
                </p>
                <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.projectSettings.archiveNotice')}
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-1 rounded-sm text-meta font-semibold disabled:opacity-40"
                style={{
                  background: 'var(--flow-danger-fill)',
                  color: 'var(--flow-danger-on-fill)',
                }}
                disabled={archiveProject.isPending || !canManageProject}
                aria-describedby={canManageProject ? undefined : 'flow-project-delete-gate'}
                onClick={() => {
                  if (!canManageProject) return;
                  setDeleteTypedName('');
                  setConfirmDelete(true);
                }}
              >
                {archiveProject.isPending
                  ? t('common.saving')
                  : t('flow.projectSettings.archiveAction')}
              </button>
            </div>
          </div>
        )}

        {/* Type-to-confirm delete modal */}
        <Dialog
          open={confirmDelete}
          onOpenChange={(o) => {
            setConfirmDelete(o);
            if (!o) setDeleteTypedName('');
          }}
        >
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle className="text-section font-semibold flex items-center gap-2">
                <Trash2 size={14} style={{ color: 'var(--flow-danger-fg)' }} />
                {t('flow.projectSettings.deleteDialogTitle')}
              </DialogTitle>
              <DialogDescription className="text-meta">
                {t('flow.projectSettings.deleteDialogPrefix')} <strong>{project?.name}</strong>
                {t('flow.projectSettings.deleteDialogSuffix')}
              </DialogDescription>
            </DialogHeader>
            <input
              autoFocus
              type="text"
              value={deleteTypedName}
              onChange={(e) => setDeleteTypedName(e.target.value)}
              placeholder={project?.name ?? ''}
              className="w-full rounded-lg px-3 py-2 text-body mt-2"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
            <div className="flex justify-end gap-2 mt-4">
              <button
                type="button"
                onClick={() => setConfirmDelete(false)}
                className="px-3 py-1.5 rounded-lg text-meta font-medium"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                disabled={
                  deleteTypedName.trim() !== (project?.name ?? '') ||
                  archiveProject.isPending ||
                  !canManageProject
                }
                onClick={() => {
                  if (!canManageProject) return;
                  archiveProject.mutate(projectId, {
                    onSuccess: () => {
                      // No toast here. useArchiveProject already reports the
                      // archive, with the scheduled permanent-deletion date.
                      // This fired a second one reading "Project deleted",
                      // which contradicted the Danger Zone copy above telling
                      // the user archiving is reversible.
                      setConfirmDelete(false);
                      navigate({ to: '/projects' });
                    },
                  });
                }}
                className="px-3 py-1.5 rounded-lg text-meta font-semibold disabled:opacity-40"
                style={{
                  background: 'var(--flow-danger-fill)',
                  color: 'var(--flow-danger-on-fill)',
                }}
              >
                {archiveProject.isPending
                  ? t('common.saving')
                  : t('flow.projectSettings.archiveConfirm')}
              </button>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

/**
 * A section heading inside a settings tab.
 *
 * Name, Description and Visibility were three bare fields in a row with
 * nothing saying where one concern ended and the next began (#1088). Visibility
 * already carries its own `legend`, so it is its own section; this gives the
 * two text fields above it one too.
 */
const SectionHeading: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <h2
    className="text-meta font-semibold uppercase tracking-wider"
    style={{ color: 'var(--flow-text-muted)' }}
  >
    {children}
  </h2>
);

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div>
    <label className="text-meta mb-1 block" style={{ color: 'var(--flow-text-muted)' }}>
      {label}
    </label>
    {children}
  </div>
);

export default ProjectSettings;
