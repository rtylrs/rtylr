import React, { useEffect, useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Save, Send } from 'lucide-react';
import {
  useCreateCampaign,
  useGetCampaign,
  useUpdateCampaign,
  useUpdateCampaignStatus,
} from '../../shared/hooks/queries/useCampaigns';
import { useListSegments } from '../../shared/hooks/queries/useSegments';
import { useListAllCustomers } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../shared/components/ui';
import { CustomerAvatar, SMSComposer, SMS_MAX_LENGTH } from '../../shared/components/crm';

const DEFAULT_SMS_BODY = 'Flash sale today! {{first_name}}, enjoy 20% off. Reply STOP to opt out.';

const SMSCampaign: React.FC = () => {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(searchStr);
  const campaignId = Number(searchParams.get('id') || 0) || undefined;
  const requestedSegmentId = searchParams.get('segment') || '';
  const isEditing = !!campaignId;

  const create = useCreateCampaign();
  const update = useUpdateCampaign();
  const updateStatus = useUpdateCampaignStatus();
  const { data: existingCampaign } = useGetCampaign(campaignId);
  const {
    data: fetchedSegments,
    isLoading: segmentsLoading,
    isError: segmentsError,
  } = useListSegments();
  const {
    data: fetchedCustomers,
    isLoading: customersLoading,
    isError: customersError,
  } = useListAllCustomers();
  const segments = useMemo(() => fetchedSegments ?? [], [fetchedSegments]);
  const customers = fetchedCustomers ?? [];

  const [name, setName] = useState(searchParams.get('name') || '');
  const [segmentId, setSegmentId] = useState<string>(requestedSegmentId);
  const [body, setBody] = useState(searchParams.get('body') || DEFAULT_SMS_BODY);
  const pending = create.isPending || update.isPending || updateStatus.isPending;

  useEffect(() => {
    if (isEditing) return;
    const params = new URLSearchParams(searchStr);
    setName(params.get('name') ?? '');
    setSegmentId(params.get('segment') ?? '');
    setBody(params.get('body') ?? DEFAULT_SMS_BODY);
  }, [isEditing, searchStr]);

  useEffect(() => {
    if (existingCampaign) {
      setName(existingCampaign.name);
      setSegmentId(
        existingCampaign.audienceSegmentId != null ? String(existingCampaign.audienceSegmentId) : ''
      );
      setBody(existingCampaign.body ?? '');
      return;
    }
    if (!isEditing && !segmentId && segments[0]) {
      setSegmentId(segments[0].id);
    }
  }, [existingCampaign, isEditing, segmentId, segments]);

  const segment = segments.find((item) => String(item.id) === segmentId);
  const audienceUnavailable =
    segmentsLoading || segmentsError || !segment || segment.customerCount <= 0;
  const bodyTooLong = body.length > SMS_MAX_LENGTH;

  const handleSave = (send: boolean): void => {
    if (pending) return;
    if (bodyTooLong) {
      toast.error(`SMS messages must be ${SMS_MAX_LENGTH} characters or fewer`);
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
      type: 'sms',
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

  const previewRecipients = customers.filter((customer) => customer.consent?.sms).slice(0, 5);

  return (
    <>
      <HelmetContainer
        title={isEditing ? 'Edit SMS campaign · rtylr' : 'New SMS campaign · rtylr'}
      />
      <div style={{ maxWidth: '1100px' }}>
        <Breadcrumbs
          items={[
            { label: 'Campaigns', to: '/campaigns' },
            { label: isEditing ? 'Edit SMS' : 'New SMS' },
          ]}
        />
        <PageHeader
          title={isEditing ? 'Edit SMS campaign' : 'New SMS campaign'}
          actions={
            <>
              <Button
                variant="outline"
                onClick={() => handleSave(false)}
                disabled={pending || !name.trim() || !body.trim() || bodyTooLong}
              >
                <Save size={13} />
                {pending ? 'Saving…' : 'Save draft'}
              </Button>
              <Button
                variant="primary"
                onClick={() => handleSave(true)}
                disabled={
                  pending || !name.trim() || !body.trim() || bodyTooLong || audienceUnavailable
                }
              >
                <Send size={13} />
                {pending ? 'Activating…' : 'Save & activate'}
              </Button>
            </>
          }
        />

        <div className="crm-two-column" style={{ display: 'grid', gap: '20px' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <Card>
              <Input
                label="Campaign name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="SMS Promo"
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
                Recipient preview
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {customersLoading ? (
                  <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: 0 }}>
                    Loading recipient preview…
                  </p>
                ) : customersError ? (
                  <p style={{ fontSize: '12px', color: '#ef4444', margin: 0 }}>
                    Recipient data is unavailable. Refresh to preview opted-in customers.
                  </p>
                ) : previewRecipients.length === 0 ? (
                  <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: 0 }}>
                    No opted-in SMS recipients are available in the loaded customer data.
                  </p>
                ) : (
                  previewRecipients.map((customer) => (
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
                        <div
                          style={{ fontSize: '13px', color: 'var(--hub-text)', fontWeight: 500 }}
                        >
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
                  ))
                )}
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
                Reminder
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
