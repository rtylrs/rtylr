import React, { useEffect, useMemo, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Eye, Save, Send, X } from 'lucide-react';
import {
  useCreateCampaign,
  useGetCampaign,
  useUpdateCampaign,
  useUpdateCampaignStatus,
} from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import { useTemplates } from '../../shared/hooks/queries/useTemplates';
import { useCrmSettings } from '../../shared/hooks/queries/useCrmSettings';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../shared/components/ui';
import { EmailEditor } from '../../shared/components/crm';
import { useDialogFocus } from '../../shared/lib/keyboard';

const EmailCampaign: React.FC = () => {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(searchStr);
  const campaignId = Number(searchParams.get('id') || 0) || undefined;
  const requestedSegmentId = searchParams.get('segment') || '';
  const isEditing = !!campaignId;

  const create = useCreateCampaign();
  const update = useUpdateCampaign();
  const updateStatus = useUpdateCampaignStatus();
  const {
    data: existingCampaign,
    isLoading: campaignLoading,
    isError: campaignError,
  } = useGetCampaign(campaignId);
  const {
    data: fetchedSegments,
    isLoading: segmentsLoading,
    isError: segmentsError,
  } = useListSegments();
  const { data: fetchedTemplates } = useTemplates();
  const { data: settings, isLoading: settingsLoading, isError: settingsError } = useCrmSettings();
  const segments = useMemo(() => fetchedSegments ?? [], [fetchedSegments]);
  const templates = (fetchedTemplates ?? []).filter((template) => template.channel === 'email');

  const [name, setName] = useState(searchParams.get('name') || '');
  const [segmentId, setSegmentId] = useState<string>(requestedSegmentId);
  const [subject, setSubject] = useState(searchParams.get('subject') || '');
  const [body, setBody] = useState(searchParams.get('body') || '');
  const [previewOpen, setPreviewOpen] = useState(false);
  const pending = create.isPending || update.isPending || updateStatus.isPending;
  const previewDialogRef = useRef<HTMLDivElement>(null);
  useDialogFocus(previewOpen, previewDialogRef, () => setPreviewOpen(false));

  useEffect(() => {
    if (isEditing) return;
    const params = new URLSearchParams(searchStr);
    setName(params.get('name') ?? '');
    setSegmentId(params.get('segment') ?? '');
    setSubject(params.get('subject') ?? '');
    setBody(params.get('body') ?? '');
  }, [isEditing, searchStr]);

  useEffect(() => {
    if (existingCampaign) {
      setName(existingCampaign.name);
      setSegmentId(
        existingCampaign.audienceSegmentId != null ? String(existingCampaign.audienceSegmentId) : ''
      );
      setSubject(existingCampaign.subject ?? '');
      setBody(existingCampaign.body ?? '');
    }
  }, [existingCampaign]);

  useEffect(() => {
    if (!isEditing && !segmentId && segments[0]) {
      setSegmentId(segments[0].id);
    }
  }, [isEditing, segmentId, segments]);

  const segment = segments.find((item) => String(item.id) === segmentId);
  const audienceUnavailable =
    segmentsLoading || segmentsError || !segment || segment.customerCount <= 0;
  const campaignReady =
    !isEditing || (!campaignLoading && !campaignError && existingCampaign?.id === campaignId);
  const previewSender =
    settings?.fromName && settings.fromEmail
      ? `${settings.fromName} <${settings.fromEmail}>`
      : settings?.fromEmail ||
        settings?.fromName ||
        (settingsLoading
          ? 'Loading sender…'
          : settingsError
            ? 'Sender unavailable'
            : 'Sender not configured');

  const applyTemplate = (templateId: string): void => {
    const template = templates.find((item) => item.id === templateId);
    if (!template) return;
    setSubject(template.subject || '');
    setBody(template.body);
    toast.success(`Applied template: ${template.name}`);
  };

  const handleSave = (send: boolean): void => {
    if (pending) return;
    if (isEditing && !campaignReady) {
      toast.error(
        campaignError ? 'Campaign could not be loaded' : 'Campaign details are still loading'
      );
      return;
    }
    if (send) {
      if (segmentsLoading) {
        toast.error('Audience data is still loading');
        return;
      }
      if (segmentsError) {
        toast.error('Cannot verify the campaign audience right now');
        return;
      }
      if (!segment) {
        toast.error('Select a valid audience segment');
        return;
      }
      if (segment.customerCount <= 0) {
        toast.error('The selected audience has no eligible recipients');
        return;
      }
    }
    const payload = {
      name,
      type: 'email',
      subject,
      body,
      audienceSegmentId: segmentId ? Number(segmentId) : null,
    };
    const callbacks = {
      onSuccess: (result: { campaignId?: number; campaign_id?: number }) => {
        if (!send) {
          navigate({ to: '/campaigns' });
          return;
        }
        const createdId = result.campaignId ?? result.campaign_id;
        if (!createdId && !campaignId) {
          toast.error('Campaign was saved but could not be activated');
          return;
        }
        updateStatus.mutate(
          { id: campaignId ?? createdId!, status: 'active' },
          {
            onSuccess: () => {
              navigate({ to: '/campaigns' });
            },
          }
        );
      },
    };

    if (campaignId) {
      update.mutate({ id: campaignId, data: payload }, callbacks);
      return;
    }

    create.mutate(payload, callbacks);
  };

  return (
    <>
      <HelmetContainer
        title={isEditing ? 'Edit email campaign · rtylr' : 'New email campaign · rtylr'}
      />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs
          items={[
            { label: 'Campaigns', to: '/campaigns' },
            { label: isEditing ? 'Edit email' : 'New email' },
          ]}
        />
        <PageHeader
          title={isEditing ? 'Edit email campaign' : 'New email campaign'}
          actions={
            <>
              <Button variant="outline" onClick={() => setPreviewOpen(true)}>
                <Eye size={13} />
                Preview
              </Button>
              <Button
                variant="outline"
                onClick={() => handleSave(false)}
                disabled={pending || !campaignReady || !name.trim() || !body.trim()}
              >
                <Save size={13} />
                {pending ? 'Saving…' : 'Save draft'}
              </Button>
              <Button
                variant="primary"
                onClick={() => handleSave(true)}
                disabled={
                  pending ||
                  !campaignReady ||
                  !name.trim() ||
                  !subject.trim() ||
                  !body.trim() ||
                  audienceUnavailable
                }
              >
                <Send size={13} />
                {pending ? 'Activating…' : 'Save & activate'}
              </Button>
            </>
          }
        />
        {isEditing && !campaignReady && (
          <p
            role={campaignError ? 'alert' : 'status'}
            style={{ color: campaignError ? '#ef4444' : 'var(--hub-text-muted)', fontSize: '12px' }}
          >
            {campaignError
              ? 'Unable to load this campaign. Refresh and try again.'
              : 'Loading campaign details…'}
          </p>
        )}

        <div className="crm-two-column" style={{ display: 'grid', gap: '20px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <div
                className="crm-form-grid"
                style={{ display: 'grid', gap: '12px', marginBottom: '14px' }}
              >
                <Input
                  label="Campaign name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Spring promo"
                />
                <Select
                  label="Start from template"
                  onChange={(event) => applyTemplate(event.target.value)}
                >
                  <option value="">Blank</option>
                  {templates.map((template) => (
                    <option key={template.id} value={template.id}>
                      {template.name}
                    </option>
                  ))}
                </Select>
              </div>
              <EmailEditor
                subject={subject}
                body={body}
                onChange={(nextSubject, nextBody) => {
                  setSubject(nextSubject);
                  setBody(nextBody);
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
              <Select
                aria-label="Campaign audience segment"
                value={segmentId}
                onChange={(event) => setSegmentId(event.target.value)}
                disabled={pending || segmentsLoading}
              >
                {segments.length === 0 && <option value="">No segments available</option>}
                {segments.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name} ({item.customerCount})
                  </option>
                ))}
              </Select>
              {segmentsLoading && (
                <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: '8px 0 0' }}>
                  Loading audience data…
                </p>
              )}
              {segmentsError && (
                <p style={{ fontSize: '11px', color: '#ef4444', margin: '8px 0 0' }}>
                  Audience data is unavailable. Activation is disabled.
                </p>
              )}
              {!segmentsLoading && !segmentsError && !segment && (
                <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: '8px 0 0' }}>
                  Select an audience segment before activating.
                </p>
              )}
              {!segmentsLoading && !segmentsError && segment && segment.customerCount <= 0 && (
                <p style={{ fontSize: '11px', color: '#f59e0b', margin: '8px 0 0' }}>
                  This segment has no eligible recipients.
                </p>
              )}
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
              ref={previewDialogRef}
              onClick={(event) => event.stopPropagation()}
              role="dialog"
              aria-modal="true"
              aria-label="Email preview"
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
                From: {previewSender}
              </div>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  justifyContent: 'space-between',
                  gap: '12px',
                }}
              >
                <h2 style={{ fontSize: '18px', margin: '0 0 16px' }}>
                  {subject || '(no subject)'}
                </h2>
                <button
                  type="button"
                  aria-label="Close email preview"
                  onClick={() => setPreviewOpen(false)}
                  style={{
                    border: 'none',
                    background: 'transparent',
                    color: '#666',
                    cursor: 'pointer',
                    padding: '2px',
                    lineHeight: 1,
                  }}
                >
                  <X size={18} />
                </button>
              </div>
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
