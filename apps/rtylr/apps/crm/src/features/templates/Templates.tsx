import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { FileText, Mail, MessageSquare, Plus, Copy } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useTemplates, useCreateTemplate } from '../../shared/hooks/queries/useTemplates';
import { PageHeader, Card, Button, SkeletonList } from '../../shared/components/ui';

const Templates: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { data: fetched, isLoading } = useTemplates();
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

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim() || !newBody.trim()) {
      toast.error(t('crm.templates.nameAndBodyRequired'));
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
      toast.success(t('crm.templates.created'));
      setShowCreate(false);
      setNewName('');
      setNewSubject('');
      setNewBody('');
      setNewCategory('general');
    } catch {
      toast.error(t('crm.templates.createFailed'));
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
      navigate({ to: '/campaigns/sms' });
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
      <HelmetContainer title={`${t('crm.templates.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title={t('crm.templates.libraryTitle')}
          subtitle={t('crm.templates.librarySubtitle')}
          actions={
            <Button variant="primary" onClick={() => setShowCreate(true)}>
              <Plus size={13} />
              {t('crm.templates.new')}
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
        ) : filtered.length === 0 ? (
          <Card>
            <EmptyState
              icon={FileText}
              title={t('crm.templates.empty')}
              description={t('crm.templates.emptyDescription')}
              action={{ label: t('crm.templates.new'), onClick: () => setShowCreate(true) }}
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
            {filtered.map((template) => {
              const Icon = template.channel === 'email' ? Mail : MessageSquare;
              const color = template.channel === 'email' ? '#3b82f6' : '#f59e0b';
              return (
                <div
                  key={template.id}
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
                      {template.category}
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
                      {template.name}
                    </h3>
                    {template.subject && (
                      <div
                        style={{ fontSize: '11px', color: 'var(--hub-text-dim)', fontWeight: 500 }}
                      >
                        {template.subject}
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
                      {template.body}
                    </p>
                  </div>
                  <div style={{ display: 'flex', gap: '6px', marginTop: 'auto' }}>
                    <Button
                      variant="outline"
                      style={{ flex: 1, justifyContent: 'center', padding: '6px' }}
                      onClick={() => duplicate(template)}
                    >
                      <Copy size={11} /> Duplicate
                    </Button>
                    <Button
                      variant="primary"
                      style={{ flex: 1, justifyContent: 'center', padding: '6px' }}
                      onClick={() => applyTemplate(template)}
                    >
                      {t('crm.templates.use')}
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
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
                borderRadius: '16px',
                padding: '24px',
                width: '420px',
                display: 'flex',
                flexDirection: 'column',
                gap: '12px',
              }}
            >
              <h3
                style={{ fontSize: '15px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}
              >
                {t('crm.templates.new')}
              </h3>
              <input
                placeholder={t('crm.templates.nameLabel')}
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
                <option value="email">{t('common.email')}</option>
                <option value="sms">{t('crm.components.sms')}</option>
              </select>
              <input
                placeholder={t('crm.templates.categoryLabel')}
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
                  placeholder={t('crm.templates.subjectLine')}
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
                placeholder={t('crm.templates.bodyContent')}
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
                  {t('common.cancel')}
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
