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
} from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useUpdateProject, useDeleteProject, useMoveProject } from './useProjects';
import {
  useProjectMembers,
  useRemoveProjectMember,
  useUpdateProjectMemberRole,
  useTransferProjectOwnership,
  useInviteProjectMember,
} from './useProjectMembers';
import { HelmetContainer } from '@voxire/components';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@voxire/ui';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import type { ApiProjectMember } from '../../shared/services/flowApi';
import { displayDescription } from '../../shared/lib/stringUtils';

const ROLE_META: Record<string, { icon: React.ReactNode; color: string }> = {
  admin: { icon: <Crown size={10} />, color: '#f59e0b' },
  member: { icon: <Shield size={10} />, color: 'var(--accent)' },
  viewer: { icon: <Eye size={10} />, color: '#8b5cf6' },
};

const ProjectSettings: React.FC = () => {
  const { t } = useTranslation();
  const { projectId, workspaceId } = useParams({ strict: false });
  const navigate = useNavigate();
  const { projects, workspaces } = useFlowStore();
  const project = projects.find((p) => p.id === projectId);

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
  const deleteProject = useDeleteProject();
  const moveProject = useMoveProject();
  const { data: members = [], isLoading: membersLoading } = useProjectMembers(projectId);
  const { data: orgMembers = [] } = useFlowTeam();
  const inviteMember = useInviteProjectMember(projectId ?? '');
  const removeMember = useRemoveProjectMember(projectId ?? '');
  const updateRole = useUpdateProjectMemberRole(projectId ?? '');
  const transferOwnership = useTransferProjectOwnership();

  const transferRef = useRef<HTMLSelectElement>(null);
  const [targetWorkspaceId, setTargetWorkspaceId] = useState('');
  const [confirmMove, setConfirmMove] = useState(false);
  const [confirmTransfer, setConfirmTransfer] = useState(false);

  const [showAddMember, setShowAddMember] = useState(false);
  const [newEmail, setNewEmail] = useState('');
  const [newRole, setNewRole] = useState<'admin' | 'member' | 'viewer'>('member');

  // Type-to-confirm modal state for project deletion. Native confirm()
  // is too easy to click through; the user has to retype the project
  // name to enable the destructive button.
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleteTypedName, setDeleteTypedName] = useState('');

  if (!projectId) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Loading…
      </div>
    );
  }

  const otherWorkspaces = workspaces.filter((w) => w.id !== (project?.workspaceId ?? workspaceId));

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Project not found.
      </div>
    );
  }

  const tabs = [
    { id: 'general', label: 'General' },
    { id: 'members', label: 'Members' },
    { id: 'integrations', label: 'Integrations' },
    { id: 'danger', label: 'Danger Zone' },
  ] as const;

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={t('flow.projectSettings.pageTitle')} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <Breadcrumbs
          items={[
            { label: project.name, to: `/w/${project.workspaceId}/p/${project.id}/overview` },
            { label: 'Settings' },
          ]}
        />
        <h1 className="text-[17px] font-bold mt-1" style={{ color: 'var(--flow-text)' }}>
          Settings
        </h1>
        <div className="flex gap-4 mt-3">
          {tabs.map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className="text-[12px] pb-1 border-b-2 transition-colors"
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

      <div className="p-6 max-w-2xl space-y-6">
        {/* ── General ── */}
        {tab === 'general' && (
          <>
            <Field label="Name">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full rounded-md px-3 py-2 text-[12px] outline-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              />
            </Field>
            <Field label="Description">
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t('flow.projectSettings.descriptionPlaceholder')}
                rows={3}
                className="w-full rounded-md px-3 py-2 text-[12px] outline-none resize-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              />
            </Field>
            <button
              onClick={() =>
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
                      icon: project.icon ?? 'folder',
                      color: project.color,
                      status: project.status ?? 'active',
                      startDate: project.startDate,
                      dueDate: project.targetDate,
                    },
                  },
                  {
                    onSuccess: () => toast.success(t('toasts.project.updated')),
                    onError: () => toast.error(t('toasts.project.updateFailed')),
                  }
                )
              }
              disabled={updateProject.isPending}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              <Save size={11} /> {updateProject.isPending ? 'Saving...' : 'Save changes'}
            </button>
          </>
        )}

        {/* ── Members ── */}
        {tab === 'members' && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <p
                className="text-[11px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {members.length} member{members.length !== 1 ? 's' : ''}
              </p>
              <button
                onClick={() => setShowAddMember((v) => !v)}
                className="flex items-center gap-1 text-[11px] px-2 py-1 rounded"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
              >
                <Plus size={10} /> Add
              </button>
            </div>

            {showAddMember && (
              <div
                className="flex gap-2 p-3 rounded-lg"
                style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
              >
                <input
                  type="email"
                  placeholder={t('flow.projectSettings.emailPlaceholder')}
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  className="flex-1 rounded px-2 py-1.5 text-[11px] outline-none"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border)',
                    color: 'var(--flow-text)',
                  }}
                />
                <select
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as 'admin' | 'member' | 'viewer')}
                  className="rounded px-2 py-1.5 text-[11px] outline-none"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border)',
                    color: 'var(--flow-text)',
                  }}
                >
                  <option value="admin">Admin</option>
                  <option value="member">Member</option>
                  <option value="viewer">Viewer</option>
                </select>
                <button
                  onClick={() => {
                    if (!newEmail.trim()) return toast.error(t('toasts.projectMember.enterEmail'));
                    inviteMember.mutate(
                      { email: newEmail.trim(), role: newRole },
                      {
                        onSuccess: () => {
                          setNewEmail('');
                          setShowAddMember(false);
                        },
                      }
                    );
                  }}
                  disabled={inviteMember.isPending}
                  className="px-3 py-1.5 rounded text-[11px] font-semibold"
                  style={{ background: 'var(--accent)', color: '#0a0a0a' }}
                >
                  Invite
                </button>
              </div>
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
              <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
                No members yet.
              </p>
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
                        border: '1px solid var(--flow-border)',
                      }}
                    >
                      <span
                        className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                        style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
                      >
                        {initials}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p
                          className="text-[12px] font-medium truncate"
                          style={{ color: 'var(--flow-text)' }}
                        >
                          {m.name || m.email}
                        </p>
                        {m.name && (
                          <p
                            className="text-[10px] truncate"
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
                          onChange={(e) =>
                            updateRole.mutate(
                              {
                                project_id: Number(projectId),
                                user_id: Number(m.userId),
                                role: e.target.value,
                              },
                              {
                                onSuccess: () =>
                                  toast.success(t('toasts.projectMember.roleUpdated')),
                                onError: () =>
                                  toast.error(t('toasts.projectMember.roleUpdateFailed')),
                              }
                            )
                          }
                          className="text-[11px] outline-none bg-transparent cursor-pointer"
                          style={{ color: roleMeta.color }}
                        >
                          <option value="admin">Admin</option>
                          <option value="member">Member</option>
                          <option value="viewer">Viewer</option>
                        </select>
                      </div>
                      <button
                        onClick={() =>
                          removeMember.mutate(
                            { project_id: Number(projectId), user_id: Number(m.userId) },
                            {
                              onSuccess: () => toast.success(t('toasts.projectMember.removed')),
                              onError: () => toast.error(t('toasts.projectMember.removeFailed')),
                            }
                          )
                        }
                        className="h-6 w-6 rounded flex items-center justify-center shrink-0"
                        style={{ color: 'var(--flow-text-dim)' }}
                        onMouseEnter={(e) => (e.currentTarget.style.color = '#ef4444')}
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
              {/* Move to workspace */}
              {otherWorkspaces.length > 0 && (
                <div>
                  <p
                    className="text-[11px] font-semibold mb-1"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    Move to workspace
                  </p>
                  <p className="text-[11px] mb-2" style={{ color: 'var(--flow-text-dim)' }}>
                    Move this project (and all its tasks) to a different workspace.
                  </p>
                  <div className="flex gap-2">
                    <select
                      value={targetWorkspaceId}
                      onChange={(e) => {
                        setTargetWorkspaceId(e.target.value);
                        setConfirmMove(false);
                      }}
                      className="flex-1 rounded px-2 py-1.5 text-[11px] outline-none"
                      style={{
                        background: 'var(--flow-input-bg)',
                        border: '1px solid var(--flow-border)',
                        color: 'var(--flow-text)',
                      }}
                    >
                      <option value="">Select workspace…</option>
                      {otherWorkspaces.map((w) => (
                        <option key={w.id} value={w.id}>
                          {w.name}
                        </option>
                      ))}
                    </select>
                    {confirmMove ? (
                      <>
                        <button
                          onClick={() => {
                            moveProject.mutate(
                              {
                                project_id: Number(projectId),
                                workspace_id: Number(targetWorkspaceId),
                              },
                              {
                                onSuccess: () => {
                                  toast.success(t('toasts.project.moved'));
                                  navigate({ to: `/w/${targetWorkspaceId}` as string });
                                },
                                onError: () => {
                                  toast.error(t('toasts.project.moveFailed'));
                                  setConfirmMove(false);
                                },
                              }
                            );
                          }}
                          disabled={moveProject.isPending}
                          className="px-3 py-1.5 rounded text-[11px] font-semibold"
                          style={{ background: '#ef4444', color: '#fff' }}
                        >
                          {moveProject.isPending ? 'Moving…' : 'Confirm'}
                        </button>
                        <button
                          onClick={() => setConfirmMove(false)}
                          className="px-3 py-1.5 rounded text-[11px]"
                          style={{ color: 'var(--flow-text-muted)' }}
                        >
                          Cancel
                        </button>
                      </>
                    ) : (
                      <button
                        onClick={() => {
                          if (!targetWorkspaceId)
                            return toast.error(t('toasts.project.selectWorkspace'));
                          setConfirmMove(true);
                        }}
                        className="flex items-center gap-1 px-3 py-1.5 rounded text-[11px] font-semibold"
                        style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
                      >
                        <ArrowRightLeft size={10} /> Move
                      </button>
                    )}
                  </div>
                  {confirmMove && (
                    <p className="text-[11px] mt-1.5" style={{ color: '#f97316' }}>
                      This will move the project and all its tasks. Click Confirm to proceed.
                    </p>
                  )}
                </div>
              )}

              {/* Transfer ownership */}
              <div>
                <p
                  className="text-[11px] font-semibold mb-1"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  Transfer Ownership
                </p>
                <p className="text-[11px] mb-2" style={{ color: 'var(--flow-text-dim)' }}>
                  Transfer ownership to any org member. You'll remain a member.
                </p>
                <div className="flex gap-2">
                  <select
                    ref={transferRef}
                    id="transfer-select"
                    onChange={() => setConfirmTransfer(false)}
                    className="flex-1 rounded px-2 py-1.5 text-[11px] outline-none"
                    style={{
                      background: 'var(--flow-input-bg)',
                      border: '1px solid var(--flow-border)',
                      color: 'var(--flow-text)',
                    }}
                    defaultValue=""
                  >
                    <option value="" disabled>
                      Select new owner…
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
                        onClick={() => {
                          const newOwnerId = Number(transferRef.current?.value);
                          transferOwnership.mutate(
                            { project_id: Number(projectId), new_owner_id: newOwnerId },
                            {
                              onSuccess: () => {
                                toast.success(t('toasts.projectMember.transferred'));
                                setConfirmTransfer(false);
                              },
                              onError: () => {
                                toast.error(t('toasts.projectMember.transferFailed'));
                                setConfirmTransfer(false);
                              },
                            }
                          );
                        }}
                        disabled={transferOwnership.isPending}
                        className="px-3 py-1.5 rounded text-[11px] font-semibold"
                        style={{ background: '#ef4444', color: '#fff' }}
                      >
                        {transferOwnership.isPending ? 'Transferring…' : 'Confirm'}
                      </button>
                      <button
                        onClick={() => setConfirmTransfer(false)}
                        className="px-3 py-1.5 rounded text-[11px]"
                        style={{ color: 'var(--flow-text-muted)' }}
                      >
                        Cancel
                      </button>
                    </>
                  ) : (
                    <button
                      onClick={() => {
                        const newOwnerId = Number(transferRef.current?.value);
                        if (!newOwnerId)
                          return toast.error(t('toasts.projectMember.selectNewOwner'));
                        setConfirmTransfer(true);
                      }}
                      className="flex items-center gap-1 px-3 py-1.5 rounded text-[11px] font-semibold"
                      style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
                    >
                      <ArrowRightLeft size={10} /> Transfer
                    </button>
                  )}
                </div>
                {confirmTransfer && (
                  <p className="text-[11px] mt-1.5" style={{ color: '#f97316' }}>
                    Ownership will transfer immediately. Click Confirm to proceed.
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
                  border: '1px solid var(--flow-border)',
                }}
              >
                <Icon size={14} style={{ color: 'var(--flow-text-muted)' }} />
                <div className="flex-1">
                  <p className="text-[12px] font-medium" style={{ color: 'var(--flow-text)' }}>
                    {name}
                  </p>
                  <p className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                    {description}
                  </p>
                </div>
                <button
                  className="text-[11px] px-2 py-1 rounded hover:bg-[var(--flow-hover)]"
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
          <div
            className="rounded-md p-4 flex items-start gap-3"
            style={{ background: 'rgba(239,68,68,0.06)', border: '1px solid rgba(239,68,68,0.25)' }}
          >
            <Trash2 size={14} style={{ color: '#ef4444' }} className="mt-0.5" />
            <div className="flex-1">
              <p className="text-[12px] font-semibold" style={{ color: '#ef4444' }}>
                Delete this project
              </p>
              <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                Once deleted, tasks are archived and cannot be restored.
              </p>
            </div>
            <button
              className="px-3 py-1 rounded text-[11px] font-semibold"
              style={{ background: '#ef4444', color: '#fff' }}
              disabled={deleteProject.isPending}
              onClick={() => {
                setDeleteTypedName('');
                setConfirmDelete(true);
              }}
            >
              {deleteProject.isPending ? 'Deleting...' : 'Delete'}
            </button>
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
              <DialogTitle className="text-[14px] font-semibold flex items-center gap-2">
                <Trash2 size={14} style={{ color: '#ef4444' }} />
                Delete project
              </DialogTitle>
              <DialogDescription className="text-[12px]">
                This permanently archives <strong>{project?.name}</strong> and all its tasks. To
                confirm, retype the project name below.
              </DialogDescription>
            </DialogHeader>
            <input
              autoFocus
              type="text"
              value={deleteTypedName}
              onChange={(e) => setDeleteTypedName(e.target.value)}
              placeholder={project?.name ?? ''}
              className="w-full rounded-lg px-3 py-2 text-[13px] outline-none mt-2"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
            <div className="flex justify-end gap-2 mt-4">
              <button
                type="button"
                onClick={() => setConfirmDelete(false)}
                className="px-3 py-1.5 rounded-lg text-[12px] font-medium"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={
                  deleteTypedName.trim() !== (project?.name ?? '') || deleteProject.isPending
                }
                onClick={() => {
                  deleteProject.mutate(projectId, {
                    onSuccess: () => {
                      toast.success(t('toasts.project.deleted'));
                      setConfirmDelete(false);
                      navigate({ to: `/w/${project!.workspaceId}` as string });
                    },
                    onError: () => toast.error(t('toasts.project.deleteFailed')),
                  });
                }}
                className="px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
                style={{ background: '#ef4444', color: '#fff' }}
              >
                {deleteProject.isPending ? 'Deleting…' : 'Delete project'}
              </button>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div>
    <label className="text-[11px] mb-1 block" style={{ color: 'var(--flow-text-muted)' }}>
      {label}
    </label>
    {children}
  </div>
);

export default ProjectSettings;
