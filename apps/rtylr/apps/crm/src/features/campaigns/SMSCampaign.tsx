import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useUserStore } from '@voxire/auth';
import { Save, Send } from 'lucide-react';
import {
  useCreateCampaign,
  useGetCampaign,
  useUpdateCampaign,
  useUpdateCampaignStatus,
} from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import { useListCustomers } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../shared/components/ui';
import { CustomerAvatar, SMSComposer } from '../../shared/components/crm';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

const DEFAULT_SMS_BODY = 'Flash sale today! {{first_name}}, enjoy 20% off. Reply STOP to opt out.';

const SMSCampaign: React.FC = () => {
  const navigate = useNavigate();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canManageCampaigns = canCrmPermission(role, 'crm:campaigns');
  const canSendCampaigns = canCrmPermission(role, 'crm:campaigns:send');
  const search = useSearch({ from: '/layout/campaigns/sms' });
  const campaignId = search.id;
  const isEditing = !!campaignId;

  const create = useCreateCampaign();
  const update = useUpdateCampaign();
  const updateStatus = useUpdateCampaignStatus();
  const { data: existingCampaign } = useGetCampaign(campaignId);
  const { data: fetchedSegments } = useListSegments();
  const { data: fetchedCustomers } = useListCustomers({ limit: 100, offset: 0 });
  const segments = useMemo(() => fetchedSegments ?? [], [fetchedSegments]);
  const customers = fetchedCustomers ?? [];

  const { t } = useTranslation();
  const [name, setName] = useState(search.name || '');
  const [segmentId, setSegmentId] = useState<string>(search.segment || '');
  const [description, setDescription] = useState(search.description || '');
  const [body, setBody] = useState(search.body || DEFAULT_SMS_BODY);
  const pending = create.isPending || update.isPending || updateStatus.isPending;

  useEffect(() => {
    if (existingCampaign) {
      setName(existingCampaign.name);
      setSegmentId(
        existingCampaign.audienceSegmentId != null ? String(existingCampaign.audienceSegmentId) : ''
      );
      setBody(existingCampaign.body ?? '');
      return;
    }
    if (!isEditing && !segmentId && segments.length > 0) {
      // ?segment= preselects the audience the user navigated here from.
      const preselect =
        (search.segment && segments.find((item) => item.id === search.segment)) || segments[0];
      setSegmentId(preselect.id);
    }
  }, [existingCampaign, isEditing, segmentId, segments, search.segment]);

  const handleSave = (send: boolean): void => {
    const payload = {
      name,
      type: 'sms',
      description,
      body,
      audienceSegmentId: segmentId ? Number(segmentId) : null,
    };
    const callbacks = {
      onSuccess: (result: { campaignId?: number; campaign_id?: number }) => {
        if (!send) {
          toast.success(t('crm.campaigns.draftSaved'));
          navigate({ to: '/campaigns' });
          return;
        }
        const createdId = result.campaignId ?? result.campaign_id;
        if (!createdId && !campaignId) {
          toast.error(t('crm.campaigns.savedNotActivated'));
          return;
        }
        updateStatus.mutate(
          { id: campaignId ?? createdId!, status: 'active' },
          {
            onSuccess: () => {
              toast.success(t('crm.campaigns.smsActivated'));
              navigate({ to: '/campaigns' });
            },
            onError: () => toast.error(t('crm.campaigns.savedNotActivated')),
          }
        );
      },
      onError: () => {
        toast.error(send ? 'Failed to save or activate SMS campaign' : 'Failed to save draft');
      },
    };

    if (campaignId) {
      update.mutate({ id: campaignId, data: payload }, callbacks);
      return;
    }

    create.mutate(payload, callbacks);
  };

  const segment = segments.find((item) => item.id === segmentId);
  const previewRecipients = customers.filter((customer) => customer.consent?.sms).slice(0, 5);

  return (
    <>
      <HelmetContainer
        title={`${isEditing ? t('crm.campaignEditor.editSmsTitle') : t('crm.campaignEditor.newSmsTitle')} · rtylr`}
      />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs
          items={[
            { label: t('nav.campaigns'), to: '/campaigns' },
            {
              label: isEditing
                ? t('crm.campaignEditor.breadcrumbEditSms')
                : t('crm.campaignEditor.breadcrumbNewSms'),
            },
          ]}
        />
        <PageHeader
          title={
            isEditing ? t('crm.campaignEditor.editSmsTitle') : t('crm.campaignEditor.newSmsTitle')
          }
          actions={
            <>
              {canManageCampaigns && (
                <Button
                  variant="outline"
                  onClick={() => handleSave(false)}
                  disabled={pending || !name.trim()}
                >
                  <Save size={13} />
                  {pending ? 'Saving…' : 'Save draft'}
                </Button>
              )}
              {canSendCampaigns && (
                <Button
                  variant="primary"
                  onClick={() => handleSave(true)}
                  disabled={pending || !name.trim() || !body.trim()}
                >
                  <Send size={13} />
                  {pending ? 'Activating…' : 'Save & activate'}
                </Button>
              )}
            </>
          }
        />

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '20px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <Input
                label={t('crm.campaignEditor.nameLabel')}
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder={t('crm.campaignEditor.smsNamePlaceholder')}
                style={{ marginBottom: '14px' }}
              />
              {/* The rest of this screen predates i18n and is still hardcoded
                  English. New strings go through t() per the repo rule;
                  converting the other ~40 is its own change. */}
              <Input
                label={t('crm.campaigns.descriptionLabel')}
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                placeholder={t('crm.campaigns.descriptionPlaceholder')}
                style={{ marginBottom: '14px' }}
              />
              <SMSComposer body={body} onChange={setBody} />
            </Card>
            <Card>
              <h4
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                  margin: '0 0 12px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {t('crm.campaignEditor.recipientPreview')}
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {previewRecipients.map((customer) => (
                  <div
                    key={customer.id}
                    style={{ display: 'flex', alignItems: 'center', gap: '10px' }}
                  >
                    <CustomerAvatar
                      initials={customer.initials}
                      points={customer.loyaltyPoints}
                      size={28}
                    />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: '13px', color: 'var(--hub-text)', fontWeight: 500 }}>
                        {customer.name}
                      </div>
                      <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                        {customer.phone}
                      </div>
                    </div>
                    <span
                      style={{
                        padding: '2px 8px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        background: 'rgba(16,185,129,0.12)',
                        color: '#10b981',
                        fontWeight: 600,
                      }}
                    >
                      opted in
                    </span>
                  </div>
                ))}
              </div>
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
                {segments.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name} ({item.customerCount})
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
                    {t('crm.campaignEditor.recipients')}
                  </div>
                  <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--hub-text)' }}>
                    {segment.customerCount.toLocaleString()}
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
                {t('crm.campaignEditor.reminder')}
              </h4>
              <p
                style={{
                  fontSize: '12px',
                  color: 'var(--hub-text-muted)',
                  margin: 0,
                  lineHeight: 1.6,
                }}
              >
                Only recipients who have opted in to SMS will receive this message. Include an
                opt-out instruction.
              </p>
            </Card>
          </div>
        </div>
      </div>
    </>
  );
};

export default SMSCampaign;
