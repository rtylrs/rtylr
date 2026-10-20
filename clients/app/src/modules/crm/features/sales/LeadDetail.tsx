import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { ArrowRight, ExternalLink, Globe, Mail, MapPin, Pencil, Phone, Trash2 } from 'lucide-react';
import { Breadcrumbs, Button, ConfirmModal } from '../../shared/components/ui';
import { type Lead, leadTitle, personName, telHref } from './data/model';
import { useDeleteLead, useLead, useUpdateLead } from './data/queries';
import { LeadStatusBadge, LoadFailed, Money, RepPicker } from './components/common';
import { LogActivityPanel, SubjectTasks, Timeline } from './components/Activity';
import LeadFormModal, { leadFieldsOf } from './components/LeadFormModal';
import { ConvertLeadModal, DisqualifyModal } from './components/LeadModals';

function externalUrl(value: string): string {
  return /^https?:\/\//i.test(value) ? value : `https://${value}`;
}

const ContactCard: React.FC<{ lead: Lead }> = ({ lead }) => {
  const { t } = useTranslation();
  const phone = telHref(lead.phone);
  const location = [lead.city, lead.country].filter(Boolean).join(', ');
  return (
    <section className="sales-panel" aria-label={t('crmSales.leadDetail.contact')}>
      <h2>{t('crmSales.leadDetail.contact')}</h2>
      {phone && (
        <a className="sales-contact-line" href={phone}>
          <Phone size={14} aria-hidden="true" />
          <bdi dir="ltr">{lead.phone}</bdi>
        </a>
      )}
      {lead.email && (
        <a className="sales-contact-line" href={`mailto:${lead.email}`}>
          <Mail size={14} aria-hidden="true" />
          <bdi dir="ltr">{lead.email}</bdi>
        </a>
      )}
      {lead.website && (
        <a
          className="sales-contact-line"
          href={externalUrl(lead.website)}
          target="_blank"
          rel="noreferrer"
        >
          <Globe size={14} aria-hidden="true" />
          {lead.website}
        </a>
      )}
      {lead.linkedin && (
        <a
          className="sales-contact-line"
          href={externalUrl(lead.linkedin)}
          target="_blank"
          rel="noreferrer"
        >
          <ExternalLink size={14} aria-hidden="true" />
          {t('crmSales.fields.linkedin')}
        </a>
      )}
      {location && (
        <span className="sales-contact-line">
          <MapPin size={14} aria-hidden="true" />
          {location}
        </span>
      )}
      {!phone && !lead.email && !lead.website && !lead.linkedin && !location && (
        <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
          {t('crmSales.leadDetail.noContact')}
        </p>
      )}
    </section>
  );
};

const LeadDetail: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams({ strict: false }) as { id?: string };
  const lead = useLead(id);
  const update = useUpdateLead();
  const remove = useDeleteLead();
  const [editing, setEditing] = useState(false);
  const [converting, setConverting] = useState(false);
  const [disqualifying, setDisqualifying] = useState(false);
  const [deleting, setDeleting] = useState(false);

  if (lead.isError) {
    return (
      <LoadFailed title={t('crmSales.leadDetail.loadFailed')} onRetry={() => void lead.refetch()} />
    );
  }
  if (!lead.data) {
    return (
      <p className="sales-empty" role="status">
        {t('crmSales.common.loading')}
      </p>
    );
  }

  const data = lead.data;
  const title = leadTitle(data);
  const person = personName(data);
  const converted = data.status === 'qualified';
  const disqualified = data.status === 'disqualified';
  const setStatus = (status: string) =>
    update.mutate({
      id: data.id ?? '',
      lead: { ...leadFieldsOf(data), value: data.value },
      status,
    });

  return (
    <div className="sales-page">
      <HelmetContainer title={`${title} · rtylr`} />
      <Breadcrumbs
        items={[{ label: t('crmSales.leads.title'), to: '/crm/sales/leads' }, { label: title }]}
      />
      <header className="sales-header">
        <div>
          <h1 style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
            {title}
            <LeadStatusBadge status={data.status} />
          </h1>
          <p>
            {[data.companyName ? person : '', data.jobTitle].filter(Boolean).join(' · ') ||
              t('crmSales.leadDetail.noPerson')}
          </p>
        </div>
        <div className="sales-actions">
          {!converted && (
            <Button onClick={() => setConverting(true)}>
              {t('crmSales.convert.action')}
              <ArrowRight size={13} />
            </Button>
          )}
          {!converted && !disqualified && (
            <Button variant="outline" onClick={() => setDisqualifying(true)}>
              {t('crmSales.disqualify.open')}
            </Button>
          )}
          <Button
            variant="outline"
            onClick={() => setEditing(true)}
            aria-label={t('crmSales.common.edit')}
          >
            <Pencil size={13} />
            {t('crmSales.common.edit')}
          </Button>
          {!converted && (
            <Button
              variant="ghost"
              onClick={() => setDeleting(true)}
              aria-label={t('crmSales.common.delete')}
            >
              <Trash2 size={13} />
            </Button>
          )}
        </div>
      </header>

      {converted && data.dealId && (
        <div className="sales-banner" role="status">
          {t('crmSales.leadDetail.convertedBanner')}
          <Link
            className="sales-link"
            to={`/crm/sales/deals/${data.dealId}`}
            style={{ fontWeight: 600 }}
          >
            {t('crmSales.leadDetail.openDeal')}
          </Link>
        </div>
      )}
      {disqualified && (
        <div className="sales-banner" role="status">
          <span style={{ flex: 1 }}>
            {t('crmSales.leadDetail.disqualifiedBanner', { reason: data.disqualifyReason || '—' })}
          </span>
          <Button variant="outline" onClick={() => setStatus('new')} disabled={update.isPending}>
            {t('crmSales.leadDetail.reopen')}
          </Button>
        </div>
      )}

      <div className="sales-detail">
        <aside style={{ display: 'flex', flexDirection: 'column', gap: 16, minWidth: 0 }}>
          <ContactCard lead={data} />
          <section className="sales-panel" aria-label={t('crmSales.leadDetail.details')}>
            <h2>{t('crmSales.leadDetail.details')}</h2>
            <dl className="sales-fields">
              <dt>{t('crmSales.fields.rep')}</dt>
              <dd>
                <RepPicker
                  label={t('crmSales.fields.rep')}
                  value={data.assignedTo}
                  onChange={(assignedTo) =>
                    update.mutate({
                      id: data.id ?? '',
                      lead: { ...leadFieldsOf(data), value: data.value, assignedTo },
                      status: data.status,
                      disqualifyReason: data.disqualifyReason,
                    })
                  }
                />
              </dd>
              <dt>{t('crmSales.fields.company')}</dt>
              <dd>{data.companyName || '—'}</dd>
              <dt>{t('crmSales.fields.industry')}</dt>
              <dd>{data.industry || '—'}</dd>
              <dt>{t('crmSales.fields.size')}</dt>
              <dd>{data.size || '—'}</dd>
              <dt>{t('crmSales.fields.source')}</dt>
              <dd>{data.source || '—'}</dd>
              <dt>{t('crmSales.fields.estimatedValue')}</dt>
              <dd>
                <Money value={data.value} billing={data.billing} />
              </dd>
              <dt>{t('crmSales.fields.attempts')}</dt>
              <dd>{data.attemptCount ?? 0}</dd>
            </dl>
            {['new', 'attempting', 'connected'].includes(data.status ?? '') && (
              <Button
                variant="ghost"
                onClick={() => setStatus('nurture')}
                disabled={update.isPending}
              >
                {t('crmSales.leadDetail.markNurture')}
              </Button>
            )}
          </section>
          {data.notes && (
            <section className="sales-panel" aria-label={t('crmSales.fields.notes')}>
              <h2>{t('crmSales.fields.notes')}</h2>
              <p
                style={{
                  margin: 0,
                  fontSize: 13,
                  whiteSpace: 'pre-wrap',
                  overflowWrap: 'anywhere',
                }}
              >
                {data.notes}
              </p>
            </section>
          )}
        </aside>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16, minWidth: 0 }}>
          <LogActivityPanel subject={{ leadId: data.id }} phone={data.phone} />
          <SubjectTasks subject={{ leadId: data.id }} />
          <Timeline subject={{ leadId: data.id }} />
        </div>
      </div>

      <LeadFormModal open={editing} onClose={() => setEditing(false)} lead={data} />
      <ConvertLeadModal
        lead={data}
        open={converting}
        onClose={() => setConverting(false)}
        onConverted={(deal) => deal?.id && navigate({ to: `/crm/sales/deals/${deal.id}` })}
      />
      <DisqualifyModal lead={data} open={disqualifying} onClose={() => setDisqualifying(false)} />
      <ConfirmModal
        open={deleting}
        onClose={() => setDeleting(false)}
        onConfirm={() =>
          remove.mutate(data.id ?? '', { onSuccess: () => navigate({ to: '/crm/sales/leads' }) })
        }
        title={t('crmSales.leadDetail.deleteTitle')}
        description={t('crmSales.leadDetail.deleteHint')}
        confirmLabel={t('crmSales.common.delete')}
        variant="danger"
      />
    </div>
  );
};

export default LeadDetail;
