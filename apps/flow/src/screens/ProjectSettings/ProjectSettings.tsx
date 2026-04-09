import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Slack, Github, Mail, Trash2, Save } from 'lucide-react';
import { useFlowStore } from '../../stores/flowStore';
import Breadcrumbs from '../../components/Breadcrumbs';

const ProjectSettings: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const { projects, members } = useFlowStore();
  const project = projects.find((p) => p.id === projectId);
  const [name, setName] = useState(project?.name || '');
  const [description, setDescription] = useState(project?.description || '');

  if (!project) {
    return <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>Project not found.</div>;
  }

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div
        className="px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <Breadcrumbs
          items={[
            { label: project.name, to: `/w/${project.workspaceId}/p/${project.id}/overview` },
            { label: 'Settings' },
          ]}
        />
        <h1 className="text-[17px] font-bold mt-1" style={{ color: 'var(--flow-text)' }}>
          Settings
        </h1>
      </div>

      <div className="p-6 max-w-2xl space-y-6">
        <Section title="General">
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
            onClick={() => toast.success('Project updated')}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
            style={{ background: '#64C5B9', color: '#0a0a0a' }}
          >
            <Save size={11} /> Save changes
          </button>
        </Section>

        <Section title="Members">
          <div className="space-y-2">
            {members.map((m) => (
              <div
                key={m.id}
                className="flex items-center gap-3 rounded-md px-3 py-2"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                }}
              >
                <span
                  className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold"
                  style={{ background: 'rgba(100,197,185,0.22)', color: '#64C5B9' }}
                >
                  {m.initials}
                </span>
                <div className="flex-1">
                  <p className="text-[12px] font-medium" style={{ color: 'var(--flow-text)' }}>
                    {m.name}
                  </p>
                  <p className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                    {m.email}
                  </p>
                </div>
                <select
                  defaultValue={m.role}
                  className="bg-transparent text-[11px] outline-none"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  <option style={{ background: 'var(--flow-card)' }}>owner</option>
                  <option style={{ background: 'var(--flow-card)' }}>admin</option>
                  <option style={{ background: 'var(--flow-card)' }}>member</option>
                  <option style={{ background: 'var(--flow-card)' }}>viewer</option>
                </select>
              </div>
            ))}
          </div>
        </Section>

        <Section title="Integrations">
          <div className="space-y-2">
            <Integration icon={Slack} name="Slack" description="Post project updates to #flow-updates" />
            <Integration icon={Github} name="GitHub" description="Link PRs to tasks automatically" />
            <Integration icon={Mail} name="Email" description="Send daily digests to members" />
          </div>
        </Section>

        <Section title="Danger Zone">
          <div
            className="rounded-md p-4 flex items-start gap-3"
            style={{
              background: 'rgba(239,68,68,0.06)',
              border: '1px solid rgba(239,68,68,0.25)',
            }}
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
              onClick={() => toast.error('Not implemented in mock data')}
            >
              Delete
            </button>
          </div>
        </Section>
      </div>
    </div>
  );
};

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <section>
    <h3
      className="text-[11px] font-semibold uppercase tracking-wider mb-2"
      style={{ color: 'var(--flow-text-muted)' }}
    >
      {title}
    </h3>
    <div className="space-y-3">{children}</div>
  </section>
);

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div>
    <label className="text-[11px] mb-1 block" style={{ color: 'var(--flow-text-muted)' }}>
      {label}
    </label>
    {children}
  </div>
);

const Integration: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  name: string;
  description: string;
}> = ({ icon: Icon, name, description }) => (
  <div
    className="flex items-center gap-3 rounded-md px-3 py-2.5"
    style={{
      background: 'var(--flow-input-bg)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <Icon size={14} />
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
    >
      Connect
    </button>
  </div>
);

export default ProjectSettings;
