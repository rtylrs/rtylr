import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Milestone as MilestoneIcon, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker, EmptyState } from '@voxire/ui';
import { useUserStore as useAuthUserStore } from '@voxire/auth';
import { useMilestones, useCreateMilestone } from './useMilestones';
import { canFlowPermission } from '../../shared/lib/roleUtils';
import { ACCENT_DEFAULT } from '../../shared/lib/colorUtils';
import { formatDueDate } from '../../shared/lib/utils';
import { CardSkeleton } from '../../shared/components/Skeleton';
import { useProjects, useCreateProject } from '../projects/useProjects';
import GateNotice from '../../shared/components/GateNotice';
import { withAlpha } from '../../shared/lib/colorUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const STATUS_STYLES: Record<string, { bg: string; color: string }> = {
  planning: { bg: withAlpha('var(--flow-neutral-fg)', 12), color: 'var(--flow-neutral-fg)' },
  active: { bg: 'var(--flow-status-approved-bg)', color: 'var(--flow-status-approved)' },
  at_risk: { bg: withAlpha('var(--flow-warn-fg)', 12), color: 'var(--flow-warn-fg)' },
  completed: { bg: withAlpha('var(--flow-info-fg)', 12), color: 'var(--flow-info-fg)' },
};

const STATUS_LABEL_KEYS: Record<string, string> = {
  planning: 'planning',
  active: 'active',
  at_risk: 'atRisk',
  completed: 'completed',
};

// Sentinel option value. Never stored as the selected project id, so it can
// not leak into the milestone payload the way a real id would.
const NEW_PROJECT_OPTION = '__new__';

const NO_PROJECTS_GATE_ID = 'milestone-project-gate';

const Milestones: React.FC = () => {
  const { t, i18n } = useTranslation();
  // Business-wide list; v2 takes project_id only as an optional narrowing
  // filter, which this screen does not use.
  const milestonesQuery = useMilestones();
  const { data: milestones = [], isLoading } = milestonesQuery;
  const { data: projects = [] } = useProjects();
  const createMilestone = useCreateMilestone();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDate, setNewDate] = useState('');
  // Every milestone belongs to a project on v2 (project_id is @required), so
  // the form has to ask. Defaults to the first project rather than leaving an
  // empty select that fails on submit.
  const [newProjectId, setNewProjectId] = useState('');
  const selectedProjectId = newProjectId || projects[0]?.id || '';

  // Inline project creation. Every milestone needs a project, so with an empty
  // project list this form had no way to be completed: the only option carried
  // value="" and Create stayed disabled. Gated on the same permission the
  // sidebar's create uses, and hidden rather than disabled - a disabled
  // <option> in a native select has nowhere to explain itself.
  const createProject = useCreateProject();
  const effectiveAccess = useAuthUserStore((s) => s.effectiveAccess);
  const canCreateProject = canFlowPermission('flow:projects:create', effectiveAccess);
  const [creatingProject, setCreatingProject] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');
  // The list refetch after a create is asynchronous, so for a moment there is
  // no <option> matching the id just selected. A native select falls back to
  // rendering its first option, which would show the wrong project while the
  // payload carried the right one. Keeping the created project here means the
  // select always has something truthful to display.
  const [justCreated, setJustCreated] = useState<{ id: string; name: string } | null>(null);

  const submitNewProject = (): void => {
    const name = newProjectName.trim();
    if (!name || createProject.isPending) return;
    createProject.mutate(
      { name, color: ACCENT_DEFAULT },
      {
        onSuccess: (created) => {
          // The mutation resolves with the { project } envelope the contract
          // returns, not a bare project. Select what was just created so the
          // milestone can be submitted without reopening the dropdown.
          const project = created?.project;
          if (project?.id) {
            setNewProjectId(String(project.id));
            setJustCreated({ id: String(project.id), name: String(project.name ?? name) });
          }
          setCreatingProject(false);
          setNewProjectName('');
          toast.success(t('toasts.projectCreated'));
        },
        // Leave the input open with the text intact; retyping a name because
        // the request failed is the wrong thing to ask of anyone.
        onError: () => toast.error(t('errors.generic')),
      }
    );
  };

  if (isLoading) {
    return (
      <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  }

  // Without this a 500 fell through to the empty state, which reads as "no
  // milestones yet" and invites creating ones that already exist.
  if (hasQueryFailed(milestonesQuery) && !showCreate) {
    return <ListLoadFailed onRetry={() => void milestonesQuery.refetch()} />;
  }

  if (milestones.length === 0 && !showCreate) {
    return (
      <div className="flex flex-col h-full">
        <HelmetContainer title={`${t('flow.milestones.heading')} · Flow`} />
        <div
          className="flex items-center justify-between px-6 py-4 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div>
            <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
              {t('flow.milestones.heading')}
            </h1>
            <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.milestones.trackedCount', { count: 0 })}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={11} /> {t('flow.milestones.newMilestone')}
          </button>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center p-6">
          <EmptyState
            icon={MilestoneIcon}
            title={t('flow.milestones.emptyTitle')}
            description={t('flow.milestones.emptyDescription')}
            action={{
              label: t('flow.milestones.createFirst'),
              onClick: () => setShowCreate(true),
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.milestones.heading')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.milestones.heading')}
          </h1>
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.milestones.trackedCount', { count: milestones.length })}
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          <Plus size={11} /> {t('flow.milestones.newMilestone')}
        </button>
      </div>
      {showCreate && (
        <div
          className="mx-6 mt-4 rounded-xl p-5 space-y-3"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder={t('flow.milestones.namePlaceholder')}
            className="w-full rounded-md px-3 py-2 text-meta"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
            autoFocus
          />
          <label className="block">
            <span className="mb-1 block text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.milestones.project')}
            </span>
            {creatingProject ? (
              <div className="flex flex-col gap-2 sm:flex-row">
                <input
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      submitNewProject();
                    }
                    if (e.key === 'Escape') {
                      setCreatingProject(false);
                      setNewProjectName('');
                    }
                  }}
                  placeholder={t('flow.milestones.newProjectPlaceholder')}
                  className="w-full rounded-md px-3 py-2 text-meta"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border-control)',
                    color: 'var(--flow-text)',
                  }}
                  autoFocus
                />
                <div className="flex shrink-0 gap-2">
                  <button
                    type="button"
                    onClick={submitNewProject}
                    disabled={!newProjectName.trim() || createProject.isPending}
                    className="min-h-11 rounded-sm px-3 py-1.5 text-meta font-semibold disabled:opacity-50"
                    style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
                  >
                    {createProject.isPending
                      ? t('flow.milestones.creating')
                      : t('flow.milestones.createProject')}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setCreatingProject(false);
                      setNewProjectName('');
                    }}
                    className="min-h-11 rounded-sm px-3 py-1.5 text-meta"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {t('flow.milestones.cancelProject')}
                  </button>
                </div>
              </div>
            ) : (
              <select
                value={selectedProjectId}
                aria-describedby={
                  projects.length === 0 && !canCreateProject ? NO_PROJECTS_GATE_ID : undefined
                }
                onChange={(e) => {
                  if (e.target.value === NEW_PROJECT_OPTION) {
                    setCreatingProject(true);
                    return;
                  }
                  setNewProjectId(e.target.value);
                }}
                className="w-full rounded-md px-3 py-2 text-title sm:text-meta"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                {projects.length === 0 ? (
                  <option value="">{t('flow.milestones.noProjects')}</option>
                ) : (
                  projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))
                )}
                {justCreated && !projects.some((p) => String(p.id) === justCreated.id) && (
                  <option value={justCreated.id}>{justCreated.name}</option>
                )}
                {canCreateProject && (
                  <option value={NEW_PROJECT_OPTION}>
                    {t('flow.milestones.newProjectOption')}
                  </option>
                )}
              </select>
            )}
            {/* Without this a member who has no projects and cannot create
                one still reaches a dead end: the only option carries
                value="", so Create stays disabled with nothing saying why.
                Root CLAUDE.md rule 12, and the same GateNotice the settings
                screen uses for its eleven gated controls. */}
            {projects.length === 0 && !canCreateProject && !justCreated && (
              <GateNotice
                id={NO_PROJECTS_GATE_ID}
                className="mt-2"
                reason={t('flow.milestones.noProjectsGate')}
              />
            )}
          </label>
          <DatePicker
            value={newDate}
            onChange={setNewDate}
            locale={i18n.language}
            aria-label={t('flow.milestones.targetDate')}
            labels={{
              placeholder: t('flow.milestones.targetDate'),
              previousMonth: t('common.datePicker.previousMonth'),
              nextMonth: t('common.datePicker.nextMonth'),
            }}
            className="text-meta"
          />
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => {
                if (!newName.trim() || !newDate || !selectedProjectId) {
                  toast.error(t('errors.generic'));
                  return;
                }
                createMilestone.mutate(
                  {
                    projectId: selectedProjectId,
                    name: newName,
                    targetDate: newDate,
                  },
                  {
                    onSuccess: () => {
                      toast.success(t('toasts.milestoneCreated'));
                      setShowCreate(false);
                      setNewName('');
                      setNewDate('');
                    },
                    onError: () => toast.error(t('errors.generic')),
                  }
                );
              }}
              disabled={
                createMilestone.isPending || !newName.trim() || !newDate || !selectedProjectId
              }
              className="px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {createMilestone.isPending
                ? t('flow.milestones.creating')
                : t('flow.milestones.create')}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowCreate(false);
                setNewName('');
                setNewDate('');
              }}
              className="px-3 py-1.5 rounded-sm text-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.milestones.cancel')}
            </button>
          </div>
        </div>
      )}
      <div className="flex-1 overflow-auto p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
        {milestones.map((m, i) => {
          const style = STATUS_STYLES[m.status];
          return (
            <motion.div
              key={m.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.16, delay: i * 0.03 }}
              className="rounded-xl p-5"
              style={{
                background: 'var(--flow-card)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span
                    className="h-8 w-8 rounded-md flex items-center justify-center"
                    style={{ background: withAlpha(m.project?.color || 'var(--accent)', 13) }}
                  >
                    <MilestoneIcon
                      size={14}
                      style={{ color: m.project?.color || 'var(--accent)' }}
                    />
                  </span>
                  <div>
                    <p className="text-body font-bold" style={{ color: 'var(--flow-text)' }}>
                      {m.name}
                    </p>
                    <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                      {m.project?.name}
                    </p>
                  </div>
                </div>
                <span
                  className="text-meta px-2 py-0.5 rounded-full font-semibold capitalize"
                  style={{ background: style.bg, color: style.color }}
                >
                  {t(`flow.milestones.status.${STATUS_LABEL_KEYS[m.status]}`)}
                </span>
              </div>
              {m.description && (
                <p className="text-meta mb-3" style={{ color: 'var(--flow-text-muted)' }}>
                  {m.description}
                </p>
              )}
              {/* No progress bar, and no task count, until a task can be put
                  into a milestone. There is no attach operation in the UI or
                  the API, so every milestone's progress is 0 of 0 forever, and
                  a bar that can only ever read empty tells a customer their
                  work has not started when nothing could have started it.

                  Restore both when the attach operation lands: rtylr#1064 and
                  REQ-R15. The percentage and the fill already read
                  milestoneProgress, so it is the surrounding markup that comes
                  back, not the calculation. */}
              <div
                className="flex items-center justify-between text-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span>{t('flow.milestones.noTasksHint')}</span>
                <span>
                  {t('flow.milestones.target')} {formatDueDate(m.targetDate, t)}
                </span>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
};

export default Milestones;
