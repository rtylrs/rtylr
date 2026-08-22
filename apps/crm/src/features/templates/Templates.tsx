import React, { useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { FileText, Mail, MessageSquare, Plus, Copy } from 'lucide-react';
import { useTemplates, useCreateTemplate } from '../../shared/hooks/queries/useTemplates';
import { PageHeader, Card, Button, EmptyState, SkeletonList } from '../../shared/components/ui';
import { useDialogFocus } from '../../shared/lib/keyboard';

const Templates: React.FC = () => {
  const navigate = useNavigate();
  const { data: fetched, isLoading, isError } = useTemplates();
  const createMut = useCreateTemplate();
  const templates = fetched ?? [];

  const [filter, setFilter] = useState<'all' | 'email' | 'sms'>('all');
  const filtered = filter === 'all' ? templates : templates.filter((t) => t.channel === filter);

  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newChannel, setNewChannel] = useState<'email' | 'sms'>('email');
  const [newSubject, setNewSubject] = useState('');
  const [newBody, setNewBody] = useState('');
  const [newCategory, setNewCategory] = useState('general');
  const createDialogRef = useRef<HTMLDivElement>(null);
  useDialogFocus(showCreate, createDialogRef, () => setShowCreate(false));

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim() || !newBody.trim()) {
      toast.error('Template name and body are required');
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        channel: newChannel,
        category: newCategory.trim() || 'general',
        subject: newChannel === 'email' ? newSubject.trim() : undefined,
        body: newBody.trim(),
      });
      setShowCreate(false);
      setNewName('');
      setNewSubject('');
      setNewBody('');
      setNewCategory('general');
    } catch {
      // The mutation hook reports the localized error.
      return;
    }
  };

  const applyTemplate = (t: (typeof templates)[number]) => {
    if (t.channel === 'email') {
      navigate({
        to: `/campaigns/email?name=${encodeURIComponent(t.name)}&subject=${encodeURIComponent(
          t.subject || ''
        )}&body=${encodeURIComponent(t.body)}`,
      });
    } else {
      navigate({
        to: `/campaigns/sms?name=${encodeURIComponent(t.name)}&body=${encodeURIComponent(t.body)}`,
      });
    }
  };

  const duplicate = (t: (typeof templates)[number]): void => {
    setNewName(`Copy of ${t.name}`);
    setNewChannel(t.channel as 'email' | 'sms');
    setNewSubject(t.subject ?? '');
    setNewBody(t.body);
    setNewCategory(t.category);
    setShowCreate(true);
  };

  return (
    <>
      <HelmetContainer title="Templates · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title="Templates library"
          subtitle="Reusable email and SMS templates"
          actions={
            <Button variant="primary" onClick={() => setShowCreate(true)}>
              <Plus size={13} />
              New template
            </Button>
          }
        />

        <div style={{ display: 'flex', gap: '6px', marginBottom: '16px' }}>
          {(['all', 'email', 'sms'] as const).map((f) => (
            <button
              type="button"
              key={f}
              onClick={() => setFilter(f)}
              style={{
                padding: '6px 14px',
                borderRadius: '8px',
                fontSize: '12px',
                fontWeight: 600,
                cursor: 'pointer',
                border: '1px solid var(--hub-border)',
                background: filter === f ? '#3b82f6' : 'transparent',
                color: filter === f ? '#fff' : 'var(--hub-text-muted)',
                textTransform: 'capitalize',
              }}
            >
              {f}
            </button>
          ))}
        </div>

        {isLoading ? (
          <Card>
            <SkeletonList count={4} />
          </Card>
        ) : isError ? (
          <Card>
            <EmptyState
              icon={FileText}
              title="Unable to load templates"
              description="Check your connection and try again."
            />
          </Card>
        ) : filtered.length === 0 ? (
          <Card>
            <EmptyState
              icon={FileText}
              title="No templates"
              description="Create one to get started."
              action={
                <Button variant="primary" onClick={() => setShowCreate(true)}>
                  <Plus size={13} />
                  New template
                </Button>
              }
            />
          </Card>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
              gap: '14px',
            }}
          >
            {filtered.map((t) => {
              const Icon = t.channel === 'email' ? Mail : MessageSquare;
              const color = t.channel === 'email' ? '#3b82f6' : '#f59e0b';
              return (
                <div
                  key={t.id}
                  style={{
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    borderRadius: '14px',
                    padding: '16px',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '10px',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                    }}
                  >
                    <div
                      style={{
                        width: '36px',
                        height: '36px',
                        borderRadius: '10px',
                        background: `${color}14`,
                        color,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <Icon size={18} />
                    </div>
                    <span
                      style={{
                        padding: '2px 8px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        fontWeight: 600,
                        background: 'var(--hub-hover)',
                        color: 'var(--hub-text-muted)',
                        textTransform: 'uppercase',
                      }}
                    >
                      {t.category}
                    </span>
                  </div>
                  <div>
                    <h3
                      style={{
                        fontSize: '14px',
                        fontWeight: 700,
                        color: 'var(--hub-text)',
                        margin: '0 0 4px',
                      }}
                    >
                      {t.name}
                    </h3>
                    {t.subject && (
                      <div
                        style={{ fontSize: '11px', color: 'var(--hub-text-dim)', fontWeight: 500 }}
                      >
                        {t.subject}
                      </div>
                    )}
                    <p
                      style={{
                        fontSize: '11px',
                        color: 'var(--hub-text-muted)',
                        margin: '6px 0 0',
                        lineHeight: 1.5,
                        display: '-webkit-box',
                        WebkitLineClamp: 2,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden',
                      }}
                    >
                      {t.body}
                    </p>
                  </div>
                  <div style={{ display: 'flex', gap: '6px', marginTop: 'auto' }}>
                    <Button
                      variant="outline"
                      style={{ flex: 1, justifyContent: 'center', padding: '6px' }}
                      onClick={() => duplicate(t)}
                    >
                      <Copy size={11} /> Duplicate
                    </Button>
                    <Button
                      variant="primary"
                      style={{ flex: 1, justifyContent: 'center', padding: '6px' }}
                      onClick={() => applyTemplate(t)}
                    >
                      Use
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {showCreate && (
          <div
            style={{
              position: 'fixed',
              inset: 0,
              background: 'rgba(0,0,0,0.5)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              zIndex: 9999,
            }}
            onClick={(e) => {
              if (e.target === e.currentTarget) setShowCreate(false);
            }}
          >
            <div
              ref={createDialogRef}
              role="dialog"
              aria-modal="true"
              aria-label="New template"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
                borderRadius: '16px',
                padding: '24px',
                width: '420px',
                maxWidth: 'calc(100vw - 32px)',
                maxHeight: 'calc(100vh - 32px)',
                overflowY: 'auto',
                display: 'flex',
                flexDirection: 'column',
                gap: '12px',
              }}
            >
              <h3
                style={{ fontSize: '15px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}
              >
                New Template
              </h3>
              <input
                aria-label="Template name"
                placeholder="Template name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                style={{
                  padding: '8px 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  outline: 'none',
                }}
              />
              <select
                aria-label="Template channel"
                value={newChannel}
                onChange={(e) => setNewChannel(e.target.value as 'email' | 'sms')}
                style={{
                  padding: '8px 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  outline: 'none',
                }}
              >
                <option value="email">Email</option>
                <option value="sms">SMS</option>
              </select>
              <input
                aria-label="Template category"
                placeholder="Category (e.g. promo, welcome)"
                value={newCategory}
                onChange={(e) => setNewCategory(e.target.value)}
                style={{
                  padding: '8px 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  outline: 'none',
                }}
              />
              {newChannel === 'email' && (
                <input
                  aria-label="Template subject"
                  placeholder="Subject line"
                  value={newSubject}
                  onChange={(e) => setNewSubject(e.target.value)}
                  style={{
                    padding: '8px 12px',
                    borderRadius: '8px',
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-input-bg)',
                    color: 'var(--hub-text)',
                    fontSize: '13px',
                    outline: 'none',
                  }}
                />
              )}
              <textarea
                aria-label="Template body"
                placeholder="Body content"
                value={newBody}
                onChange={(e) => setNewBody(e.target.value)}
                rows={4}
                style={{
                  padding: '8px 12px',
                  borderRadius: '8px',
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  outline: 'none',
                  resize: 'vertical',
                  fontFamily: 'inherit',
                }}
              />
              <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                <Button variant="outline" onClick={() => setShowCreate(false)}>
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  onClick={handleCreate}
                  disabled={createMut.isPending || !newName.trim() || !newBody.trim()}
                >
                  {createMut.isPending ? 'Creating…' : 'Create'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default Templates;
