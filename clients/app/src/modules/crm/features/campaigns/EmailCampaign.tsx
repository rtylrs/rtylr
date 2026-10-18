import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useUserStore } from '../../../../shared/stores/userStore';
import { Eye, Save, Send } from 'lucide-react';
import {
  useCreateCampaign,
  useGetCampaign,
  useUpdateCampaign,
  useCampaignAudience,
} from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import { useTemplates } from '../../shared/hooks/queries/useTemplates';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../shared/components/ui';
import { EmailEditor } from '../../shared/components/crm';
import { canCrmPermission } from '../../shared/lib/crmPermissions';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import SendCampaignModal from './SendCampaignModal';

const EmailCampaign: React.FC = () => {
  const navigate = useNavigate();
  const search = useSearch({ from: '/business/crm/campaigns/email' });
  const campaignId = search.id;
  const isEditing = !!campaignId;
  const canCreateCampaign = canCrmPermission('crm:campaigns:create');
  const canManageCampaign = canCrmPermission('crm:campaigns:manage');
  const canWriteCampaign = isEditing ? canManageCampaign : canCreateCampaign;
  const canSendCampaigns = canCrmPermission('crm:campaigns:send');
  // Set after a create that could not be activated, so clicking "Save & activate"
  // again updates that campaign instead of creating a duplicate.
  const [createdId, setCreatedId] = useState<number | undefined>(undefined);
  const savedCampaignId = campaignId ?? createdId;

  const create = useCreateCampaign();
  const update = useUpdateCampaign();
  const { data: existingCampaign } = useGetCampaign(campaignId);
  const segmentsQuery = useListSegments();
  const templatesQuery = useTemplates();
  const segmentsFailed = hasQueryFailed(segmentsQuery);
  const templatesFailed = hasQueryFailed(templatesQuery);
  const { data: fetchedSegments } = segmentsQuery;
  const { data: fetchedTemplates } = templatesQuery;
  const segments = useMemo(() => fetchedSegments ?? [], [fetchedSegments]);
  const templates = (fetchedTemplates ?? []).filter((template) => template.channel === 'email');

  const { t } = useTranslation();
  const [name, setName] = useState(search.name || '');
  const [segmentId, setSegmentId] = useState<string>(search.segment || '');
  const [subject, setSubject] = useState(search.subject || '');
  const [description, setDescription] = useState(search.description || '');
  const [body, setBody] = useState(search.body || '');
  const [previewOpen, setPreviewOpen] = useState(false);
  const pending = create.isPending || update.isPending;

  useEffect(() => {
    if (existingCampaign) {
      setName(existingCampaign.name);
      setSegmentId(
        existingCampaign.audienceSegmentId != null ? String(existingCampaign.audienceSegmentId) : ''
      );
      setSubject(existingCampaign.subject ?? '');
      setBody(existingCampaign.body ?? '');
      return;
    }
    if (!isEditing && !segmentId && search.segment) {
      // ?segment= preselects the audience the user navigated here from;
      // otherwise the campaign goes to every customer who opted in.
      const preselect = segments.find((item) => item.id === search.segment);
      if (preselect) setSegmentId(preselect.id);
    }
  }, [existingCampaign, isEditing, segmentId, segments, search.segment]);

  const applyTemplate = (templateId: string): void => {
    const template = templates.find((item) => item.id === templateId);
    if (!template) return;
    setSubject(template.subject || '');
    setBody(template.body);
    toast.success(t('crm.campaignEditor.templateApplied', { name: template.name }));
  };

  const [sendOpen, setSendOpen] = useState(false);
  const audience = useCampaignAudience(segmentId ? Number(segmentId) : undefined);

  const handleSave = (send: boolean): void => {
    const payload = {
      name,
      type: 'email',
      subject,
      description,
      body,
      audienceSegmentId: segmentId ? Number(segmentId) : null,
    };
    const callbacks = {
      onSuccess: (result: { campaignId?: number; campaign_id?: number }) => {
        if (!send) {
          toast.success(t('crm.campaigns.draftSaved'));
          navigate({ to: '/crm/campaigns' });
          return;
        }
        const id = savedCampaignId ?? result.campaignId ?? result.campaign_id;
        if (!id) {
          toast.error(t('crm.campaignSend.failed'));
          return;
        }
        if (!savedCampaignId) setCreatedId(id);
        setSendOpen(true);
      },
      onError: () => {
        toast.error(t('crm.campaignEditor.saveFailed'));
      },
    };

    if (savedCampaignId) {
      update.mutate({ id: savedCampaignId, data: payload }, callbacks);
      return;
    }

    create.mutate(payload, callbacks);
  };

  const segment = segments.find((item) => item.id === segmentId);

  return (
    <>
      <HelmetContainer
        title={`${isEditing ? t('crm.campaignEditor.editEmailTitle') : t('crm.campaignEditor.newEmailTitle')} · rtylr`}
      />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs
          items={[
            { label: t('crmNav.campaigns'), to: '/crm/campaigns' },
            {
              label: isEditing
                ? t('crm.campaignEditor.breadcrumbEditEmail')
                : t('crm.campaignEditor.breadcrumbNewEmail'),
            },
          ]}
        />
        <PageHeader
          title={
            isEditing
              ? t('crm.campaignEditor.editEmailTitle')
              : t('crm.campaignEditor.newEmailTitle')
          }
          actions={
            <>
              <Button variant="outline" onClick={() => setPreviewOpen(true)}>
                <Eye size={13} />
                {t('crm.campaignEditor.preview')}
              </Button>
              {canWriteCampaign && (
                <Button
                  variant="outline"
                  onClick={() => handleSave(false)}
                  disabled={pending || !name.trim()}
                >
                  <Save size={13} />
                  {pending ? t('crm.states.saving') : t('crm.campaignEditor.saveDraft')}
                </Button>
              )}
              {canWriteCampaign && canSendCampaigns && (
                <Button
                  variant="primary"
                  onClick={() => handleSave(true)}
                  disabled={pending || !name.trim() || !subject.trim() || !body.trim()}
                >
                  <Send size={13} />
                  {t('crm.campaignEditor.saveAndSend')}
                </Button>
              )}
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
                  label={t('crm.campaignEditor.nameLabel')}
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder={t('crm.campaignEditor.emailNamePlaceholder')}
                />
                {/* The rest of this screen predates i18n and is still hardcoded
                    English. New strings go through t() per the repo rule;
                    converting the other ~40 is its own change. */}
                <Input
                  label={t('crm.campaigns.descriptionLabel')}
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  placeholder={t('crm.campaigns.descriptionPlaceholder')}
                />
                <Select
                  label={t('crm.campaignEditor.startFromTemplate')}
                  onChange={(event) => applyTemplate(event.target.value)}
                >
                  <option value="">{t('crm.campaignEditor.blank')}</option>
                  {templates.map((template) => (
                    <option key={template.id} value={template.id}>
                      {template.name}
                    </option>
                  ))}
                </Select>
                {templatesFailed && (
                  <div role="alert" style={{ marginTop: '8px', fontSize: '12px' }}>
                    <span style={{ color: 'var(--hub-text-muted)' }}>
                      {t('crmErrors.generic')}{' '}
                    </span>
                    <Button variant="outline" onClick={() => void templatesQuery.refetch()}>
                      {t('crm.actions.refresh')}
                    </Button>
                  </div>
                )}
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
                {t('crm.campaignEditor.audience')}
              </h4>
              <Select value={segmentId} onChange={(event) => setSegmentId(event.target.value)}>
                <option value="">{t('crm.campaignEditor.allCustomers')}</option>
                {segments.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name} ({item.customerCount})
                  </option>
                ))}
              </Select>
              {audience.data && (
                <p style={{ margin: '8px 0 0', fontSize: '12px', color: 'var(--hub-text-muted)' }}>
                  {t('crm.campaignSend.reach', {
                    count: audience.data.eligible,
                    total: audience.data.total,
                  })}
                </p>
              )}
              {segmentsFailed && (
                <div role="alert" style={{ marginTop: '8px', fontSize: '12px' }}>
                  <span style={{ color: 'var(--hub-text-muted)' }}>{t('crmErrors.generic')} </span>
                  <Button variant="outline" onClick={() => void segmentsQuery.refetch()}>
                    {t('crm.actions.refresh')}
                  </Button>
                </div>
              )}
              {segment && (
                <div
                  style={{
                    marginTop: '12px',
                    padding: '12px',
                    background: 'var(--dash-hover)',
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
                    {t('crm.campaignEditor.recipients')}
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
                {t('crm.campaignEditor.tips')}
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
                <li>{t('crm.campaignEditor.tipSubjectLength')}</li>
                <li>{t('crm.campaignEditor.tipPersonalization')}</li>
                <li>{t('crm.campaignEditor.tipCallToAction')}</li>
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
              onClick={(event) => event.stopPropagation()}
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
      <SendCampaignModal
        open={sendOpen}
        onClose={() => setSendOpen(false)}
        campaignId={savedCampaignId}
        segmentId={segmentId ? Number(segmentId) : undefined}
        onSent={() => navigate({ to: '/crm/campaigns' })}
      />
    </>
  );
};

export default EmailCampaign;
