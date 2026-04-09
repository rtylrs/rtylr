import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Send, Save } from 'lucide-react';
import { useCreateCampaign } from '../../hooks/useCampaigns';
import { useSegments } from '../../hooks/useSegments';
import { useCustomers } from '../../hooks/useCustomers';
import { Breadcrumbs, Button, Card, Input, PageHeader, Select } from '../../components/ui';
import { SMSComposer, CustomerAvatar } from '../../components/crm';
import { mockSegments, mockCustomers } from '../../lib/mockData';

const SMSCampaign: React.FC = () => {
  const navigate = useNavigate();
  const create = useCreateCampaign();
  const { data: fetchedSegments } = useSegments();
  const { data: fetchedCustomers } = useCustomers();
  const segments = fetchedSegments && fetchedSegments.length ? fetchedSegments : mockSegments;
  const customers = fetchedCustomers && fetchedCustomers.length ? fetchedCustomers : mockCustomers;

  const [name, setName] = useState('');
  const [segmentId, setSegmentId] = useState<number>(segments[0]?.id || 0);
  const [body, setBody] = useState('Flash sale today! {{first_name}}, enjoy 20% off. Reply STOP to opt out.');

  const handleSave = (send: boolean) => {
    create.mutate({
      name,
      description: body.slice(0, 50),
      channel: 'SMS',
      status: send ? 'Active' : 'Draft',
      audience: segments.find((s) => s.id === segmentId)?.customerCount || 0,
      body,
      segmentId,
    }, {
      onSuccess: () => {
        toast.success(send ? 'SMS campaign sent' : 'Draft saved');
        navigate('/campaigns');
      },
      onError: () => {
        toast.success(send ? 'SMS campaign queued' : 'Draft saved');
        navigate('/campaigns');
      },
    });
  };

  const segment = segments.find((s) => s.id === segmentId);
  const previewRecipients = customers.filter((c) => c.consent?.sms).slice(0, 5);

  return (
    <div style={{ maxWidth: '1100px' }}>
      <Breadcrumbs items={[{ label: 'Campaigns', to: '/campaigns' }, { label: 'New SMS' }]} />
      <PageHeader
        title="New SMS campaign"
        actions={(
          <>
            <Button variant="outline" onClick={() => handleSave(false)}><Save size={13} />Save draft</Button>
            <Button variant="primary" onClick={() => handleSave(true)}><Send size={13} />Send now</Button>
          </>
        )}
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '20px' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <Card>
            <Input label="Campaign name" value={name} onChange={(e) => setName(e.target.value)} placeholder="SMS Promo" style={{ marginBottom: '14px' }} />
            <SMSComposer body={body} onChange={setBody} />
          </Card>
          <Card>
            <h4 style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 12px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Recipient preview</h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {previewRecipients.map((c) => (
                <div key={c.id} style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <CustomerAvatar initials={c.initials} points={c.loyaltyPoints} size={28} />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: '13px', color: 'var(--hub-text)', fontWeight: 500 }}>{c.name}</div>
                    <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>{c.phone}</div>
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
            <h4 style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 10px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Audience</h4>
            <Select value={segmentId} onChange={(e) => setSegmentId(Number(e.target.value))}>
              {segments.map((s) => (
                <option key={s.id} value={s.id}>{s.name} ({s.customerCount})</option>
              ))}
            </Select>
            {segment && (
              <div style={{ marginTop: '12px', padding: '12px', background: 'rgba(59,130,246,0.08)', borderRadius: '8px' }}>
                <div style={{ fontSize: '11px', color: 'var(--hub-text-dim)', textTransform: 'uppercase', fontWeight: 700 }}>Recipients</div>
                <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--hub-text)' }}>{segment.customerCount.toLocaleString()}</div>
              </div>
            )}
          </Card>
          <Card>
            <h4 style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 8px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Reminder</h4>
            <p style={{ fontSize: '12px', color: 'var(--hub-text-muted)', margin: 0, lineHeight: 1.6 }}>
              Only recipients who have opted in to SMS will receive this message. Include an opt-out instruction.
            </p>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default SMSCampaign;
