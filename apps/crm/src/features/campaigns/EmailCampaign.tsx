import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Send, Save, Eye } from 'lucide-react';
import { useCreateCampaign } from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import { useTemplates } from '../../shared/hooks/queries/useTemplates';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../shared/components/ui';
import { EmailEditor } from '../../shared/components/crm';

const EmailCampaign: React.FC = () => {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const create = useCreateCampaign();
  const { data: fetchedSegments } = useListSegments();
  const { data: fetchedTemplates } = useTemplates();
  const segments = fetchedSegments ?? [];
  const templates = (fetchedTemplates ?? []).filter((t) => t.channel === 'email');

  const [name, setName] = useState(new URLSearchParams(searchStr).get('name') || '');
  const [segmentId, setSegmentId] = useState<number>(segments[0]?.id || 0);
  const [subject, setSubject] = useState(new URLSearchParams(searchStr).get('subject') || '');
  const [body, setBody] = useState(new URLSearchParams(searchStr).get('body') || '');
  const [previewOpen, setPreviewOpen] = useState(false);

  const applyTemplate = (templateId: string) => {
    const t = templates.find((x) => x.id === Number(templateId));
    if (!t) return;
    setSubject(t.subject || '');
    setBody(t.body);
    toast.success(`Applied template: ${t.name}`);
  };

  const handleSave = (send: boolean) => {
    create.mutate(
      {
        name,
        type: 'email',
        status: send ? 'active' : 'draft',
        subject,
        body,
        audienceSegmentId: segmentId,
      },
      {
        onSuccess: () => {
          toast.success(send ? 'Campaign sent' : 'Draft saved');
          navigate({ to: '/campaigns' });
        },
        onError: () => {
          toast.error(send ? 'Failed to send campaign' : 'Failed to save draft');
        },
      }
    );
  };

  const segment = segments.find((s) => s.id === segmentId);

  return (
    <>
      <HelmetContainer title="New email campaign · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs items={[{ label: 'Campaigns', to: '/campaigns' }, { label: 'New email' }]} />
        <PageHeader
          title="New email campaign"
          actions={
            <>
              <Button variant="outline" onClick={() => setPreviewOpen(true)}>
                <Eye size={13} />
                Preview
              </Button>
              <Button
                variant="outline"
                onClick={() => handleSave(false)}
                disabled={create.isPending || !name.trim()}
              >
                <Save size={13} />
                {create.isPending ? 'Saving…' : 'Save draft'}
              </Button>
              <Button
                variant="primary"
                onClick={() => handleSave(true)}
                disabled={create.isPending || !name.trim() || !subject.trim()}
              >
                <Send size={13} />
                {create.isPending ? 'Sending…' : 'Send now'}
              </Button>
            </>
          }
        />

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '20px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: '12px',
                  marginBottom: '14px',
                }}
              >
                <Input
                  label="Campaign name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Spring promo"
                />
                <Select label="Start from template" onChange={(e) => applyTemplate(e.target.value)}>
                  <option value="">Blank</option>
                  {templates.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name}
                    </option>
                  ))}
                </Select>
              </div>
              <EmailEditor
                subject={subject}
                body={body}
                onChange={(s, b) => {
                  setSubject(s);
                  setBody(b);
                }}
              />
            </Card>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 10px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Audience
              </h4>
              <Select value={segmentId} onChange={(e) => setSegmentId(Number(e.target.value))}>
                {segments.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.customerCount})
                  </option>
                ))}
              </Select>
              {segment && (
                <div
                  style={{
                    marginTop: '12px',
                    padding: '12px',
                    background: 'rgba(59,130,246,0.08)',
                    borderRadius: '8px',
                  }}
                >
                  <div
                    style={{
                      fontSize: '11px',
                      color: 'var(--hub-text-dim)',
                      textTransform: 'uppercase',
                      fontWeight: 700,
                    }}
                  >
                    Recipients
                  </div>
                  <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--hub-text)' }}>
                    {segment.customerCount.toLocaleString()}
                  </div>
                  <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                    {segment.description}
                  </div>
                </div>
              )}
            </Card>
            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 8px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                Tips
              </h4>
              <ul
                style={{
                  margin: 0,
                  paddingLeft: '16px',
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                  lineHeight: 1.7,
                }}
              >
                <li>Keep subject lines under 60 characters</li>
                <li>Use first_name for personalization</li>
                <li>Include a clear call to action</li>
              </ul>
            </Card>
          </div>
        </div>

        {previewOpen && (
          <div
            onClick={() => setPreviewOpen(false)}
            style={{
              position: 'fixed',
              inset: 0,
              background: 'rgba(0,0,0,0.6)',
              zIndex: 200,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '20px',
            }}
          >
            <div
              onClick={(e) => e.stopPropagation()}
              style={{
                width: '100%',
                maxWidth: '560px',
                background: '#fff',
                color: '#111',
                borderRadius: '14px',
                padding: '24px',
                fontFamily: 'system-ui',
              }}
            >
              <div style={{ fontSize: '11px', color: '#666', marginBottom: '8px' }}>
                From: hello@rtylr.com
              </div>
              <h2 style={{ fontSize: '18px', margin: '0 0 16px' }}>{subject || '(no subject)'}</h2>
              <div style={{ fontSize: '14px', lineHeight: 1.7, whiteSpace: 'pre-wrap' }}>
                {body || '(empty)'}
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default EmailCampaign;
