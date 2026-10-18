import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useParams, useNavigate } from '@tanstack/react-router';
import {
  Phone,
  Mail,
  Users,
  Plus,
  Trash2,
  Edit2,
  CheckCircle,
  XCircle,
  MinusCircle,
  ArrowLeft,
  Clock,
  UserPlus,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { motion, AnimatePresence } from 'motion/react';
import {
  useLead,
  useUpdateLead,
  useDeleteLead,
  useLeadTouchpoints,
  useCreateTouchpoint,
  useDeleteTouchpoint,
  useContacts,
  useCreateContact,
  useDeleteContact,
  useContracts,
  useCreateContract,
  useSalesReps,
} from '../../shared/hooks/queries/useSales';
import type {
  LeadStage,
  TouchpointType,
  TouchpointOutcome,
  ContractPlan,
} from '../../shared/hooks/queries/useSales';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import { followUpStatus, formatDay } from '../../shared/lib/dateUtils';
import LeadFormModal from './components/LeadFormModal';
import {
  Button,
  Card,
  Input,
  Select,
  Textarea,
  SkeletonList,
  ConfirmModal,
} from '../../shared/components/ui';

// Ids only: this is evaluated at module load, before i18next has a language.
// The labels come from crm.leadStages.<id> at render.
const STAGE_VALUES: LeadStage[] = [
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];

const TOUCHPOINT_TYPE_ICONS: Record<TouchpointType, React.ReactNode> = {
  phone: <Phone size={14} />,
  email: <Mail size={14} />,
  meeting: <Users size={14} />,
};

const OUTCOME_COLORS: Record<TouchpointOutcome, { bg: string; color: string }> = {
  positive: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
  neutral: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  negative: { bg: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  no_answer: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
};

const OUTCOME_ICONS: Record<TouchpointOutcome, React.ReactNode> = {
  positive: <CheckCircle size={12} />,
  neutral: <MinusCircle size={12} />,
  negative: <XCircle size={12} />,
  no_answer: <MinusCircle size={12} />,
};

const InlineLoadError: React.FC<{ onRetry: () => void }> = ({ onRetry }) => {
  const { t } = useTranslation();
  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: '8px',
        color: 'var(--hub-text-muted)',
        fontSize: '12px',
      }}
    >
      <span>{t('crmErrors.generic')}</span>
      <Button variant="outline" onClick={onRetry}>
        {t('crm.actions.retry')}
      </Button>
    </div>
  );
};

function formatDate(iso: string): string {
  if (!iso) return '-';
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

const OUTCOME_LABEL_KEYS: Record<TouchpointOutcome, string> = {
  positive: 'crm.leadDetail.outcomePositive',
  neutral: 'crm.leadDetail.outcomeNeutral',
  negative: 'crm.leadDetail.outcomeNegative',
  no_answer: 'crm.leadDetail.outcomeNoAnswer',
};

const LeadDetail: React.FC = () => {
  const { id } = useParams({ strict: false });
  const leadId = id ? Number(id) : undefined;
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();

  const leadQuery = useLead(leadId);
  const touchpointsQuery = useLeadTouchpoints(leadId);
  const contactsQuery = useContacts(leadId);
  const contractsQuery = useContracts();
  const { data: lead, isLoading: leadLoading } = leadQuery;
  const { data: touchpoints } = touchpointsQuery;
  const { data: contacts } = contactsQuery;
  const { data: allContracts } = contractsQuery;
  const leadFailed = hasQueryFailed(leadQuery);
  const touchpointsFailed = hasQueryFailed(touchpointsQuery);
  const contactsFailed = hasQueryFailed(contactsQuery);
  const contractsFailed = hasQueryFailed(contractsQuery);
  const updateLead = useUpdateLead();
  const [editOpen, setEditOpen] = useState(false);
  const { data: reps = [] } = useSalesReps();
  const repName = (id: number | null | undefined): string | undefined =>
    id != null ? reps.find((rep) => rep.id === id)?.name : undefined;
  const deleteLead = useDeleteLead();
  const createTouchpoint = useCreateTouchpoint();
  const deleteTouchpoint = useDeleteTouchpoint();
  const createContact = useCreateContact();
  const deleteContact = useDeleteContact();
  const createContract = useCreateContract();

  const contract = (allContracts ?? []).find((c) => c.leadId === leadId);
  const sortedTouchpoints = [...(touchpoints ?? [])].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  // Stage change
  const handleStageChange = (stage: LeadStage): void => {
    if (!leadId) return;
    updateLead.mutate(
      { id: leadId, data: { stage } },
      { onError: () => toast.error(t('crm.leadDetail.stageUpdateFailed')) }
    );
  };

  // Delete lead
  const [confirmDeleteLead, setConfirmDeleteLead] = useState(false);
  const handleDeleteLead = (): void => {
    if (!leadId) return;
    deleteLead.mutate(leadId, {
      onSuccess: () => navigate({ to: '/crm/sales/leads' }),
      onError: () => toast.error(t('crm.leadDetail.deleteFailed')),
    });
  };

  // Touchpoint form
  const [tpOpen, setTpOpen] = useState(false);
  const [tpForm, setTpForm] = useState({
    type: 'phone' as TouchpointType,
    outcome: 'neutral' as TouchpointOutcome,
    notes: '',
    nextFollowUp: '',
    contactId: '',
  });

  const submitTouchpoint = (): void => {
    if (!tpForm.notes.trim()) {
      toast.error(t('crm.leadDetail.notesRequired'));
      return;
    }
    if (!leadId) return;
    createTouchpoint.mutate(
      {
        leadId,
        type: tpForm.type,
        outcome: tpForm.outcome,
        notes: tpForm.notes,
        repName: 'Me',
        nextFollowUp: tpForm.nextFollowUp || null,
        contactId: tpForm.contactId ? Number(tpForm.contactId) : null,
      },
      {
        onSuccess: () => {
          setTpForm({
            type: 'phone',
            outcome: 'neutral',
            notes: '',
            nextFollowUp: '',
            contactId: '',
          });
          setTpOpen(false);
        },
        onError: () => toast.error(t('crm.leadDetail.touchpointFailed')),
      }
    );
  };

  // Delete touchpoint
  const [deleteTpId, setDeleteTpId] = useState<number | null>(null);

  // Contact form
  const [contactOpen, setContactOpen] = useState(false);
  const [contactForm, setContactForm] = useState({ name: '', role: '', phone: '', email: '' });

  const submitContact = (): void => {
    if (!contactForm.name.trim()) {
      toast.error(t('crm.leadDetail.contactNameRequired'));
      return;
    }
    if (!leadId) return;
    createContact.mutate(
      { leadId, ...contactForm },
      {
        onSuccess: () => {
          setContactForm({ name: '', role: '', phone: '', email: '' });
          setContactOpen(false);
        },
        onError: () => toast.error(t('crm.leadDetail.addContactFailed')),
      }
    );
  };

  // Delete contact
  const [deleteContactId, setDeleteContactId] = useState<number | null>(null);

  // Contract form
  const [contractOpen, setContractOpen] = useState(false);
  const [contractForm, setContractForm] = useState({
    plan: 'pro' as ContractPlan,
    mrr: '',
    seats: '',
    signedAt: '',
    signedBy: '',
    closedBy: '',
    notes: '',
  });

  const submitContract = (): void => {
    if (!leadId || !contractForm.signedAt || !contractForm.signedBy) return;
    createContract.mutate(
      {
        leadId,
        plan: contractForm.plan,
        mrr: Number(contractForm.mrr),
        seats: Number(contractForm.seats),
        signedAt: contractForm.signedAt,
        signedBy: contractForm.signedBy,
        closedBy: contractForm.closedBy,
        notes: contractForm.notes,
      },
      {
        onSuccess: () => {
          setContractOpen(false);
        },
        onError: () => toast.error(t('crm.leadDetail.saveContractFailed')),
      }
    );
  };

  if (leadLoading) {
    return (
      <div style={{ padding: '24px' }}>
        <SkeletonList count={6} />
      </div>
    );
  }

  if (leadFailed) {
    return (
      <div style={{ padding: '24px', maxWidth: '520px' }}>
        <InlineLoadError onRetry={() => void leadQuery.refetch()} />
      </div>
    );
  }

  if (!lead) {
    return (
      <div style={{ padding: '24px', color: 'var(--hub-text-muted)', fontSize: '14px' }}>
        {t('crm.leadDetail.notFound')}
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={`${t('crm.leadDetail.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1200px' }}>
        {/* Back + header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <button
            type="button"
            aria-label={t('crm.leadDetail.backToLeads')}
            onClick={() => navigate({ to: '/crm/sales/leads' })}
            style={{
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
              display: 'flex',
              alignItems: 'center',
              padding: '4px',
            }}
          >
            <ArrowLeft size={16} />
          </button>
          <div style={{ flex: 1 }}>
            <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              {lead.company}
            </h2>
            <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
              {lead.industry} · {lead.source}
            </p>
          </div>
          <div style={{ display: 'flex', gap: '8px' }}>
            <Button variant="outline" onClick={() => setConfirmDeleteLead(true)}>
              <Trash2 size={13} />
              {t('crmCommon.delete')}
            </Button>
          </div>
        </div>

        {/* Two-column layout */}
        <div
          style={{ display: 'grid', gridTemplateColumns: '1fr minmax(280px, 360px)', gap: '20px' }}
        >
          {/* Left: Touchpoint timeline */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <h3
                style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}
              >
                {t('crm.leadDetail.timelineTitle')}
              </h3>
              <Button variant="primary" onClick={() => setTpOpen((v) => !v)}>
                <Plus size={13} />
                {t('crm.leadDetail.logTouchpointTitle')}
              </Button>
            </div>

            {/* Touchpoint form */}
            <AnimatePresence>
              {tpOpen && (
                <motion.div
                  initial={{ opacity: 0, y: -6 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -6 }}
                >
                  <Card>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                      <p
                        style={{
                          fontSize: '13px',
                          fontWeight: 600,
                          color: 'var(--hub-text)',
                          margin: 0,
                        }}
                      >
                        {t('crm.leadDetail.logTouchpointTitle')}
                      </p>
                      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                        <Select
                          label={t('crm.leadDetail.typeLabel')}
                          value={tpForm.type}
                          onChange={(e) =>
                            setTpForm((f) => ({ ...f, type: e.target.value as TouchpointType }))
                          }
                        >
                          <option value="phone">{t('crm.leadDetail.typePhone')}</option>
                          <option value="email">{t('crm.leadDetail.typeEmail')}</option>
                          <option value="meeting">{t('crm.leadDetail.typeMeeting')}</option>
                        </Select>
                        <Select
                          label={t('crm.leadDetail.outcomeLabel')}
                          value={tpForm.outcome}
                          onChange={(e) =>
                            setTpForm((f) => ({
                              ...f,
                              outcome: e.target.value as TouchpointOutcome,
                            }))
                          }
                        >
                          <option value="positive">{t('crm.leadDetail.outcomePositive')}</option>
                          <option value="neutral">{t('crm.leadDetail.outcomeNeutral')}</option>
                          <option value="negative">{t('crm.leadDetail.outcomeNegative')}</option>
                          <option value="no_answer">{t('crm.leadDetail.outcomeNoAnswer')}</option>
                        </Select>
                      </div>
                      <Textarea
                        label={t('crm.leadDetail.notesLabel')}
                        value={tpForm.notes}
                        onChange={(e) => setTpForm((f) => ({ ...f, notes: e.target.value }))}
                        placeholder={t('crm.leadDetail.notesPlaceholder')}
                      />
                      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                        <Input
                          label={t('crm.leadDetail.nextFollowUp')}
                          type="date"
                          value={tpForm.nextFollowUp}
                          onChange={(e) =>
                            setTpForm((f) => ({ ...f, nextFollowUp: e.target.value }))
                          }
                        />
                        <Select
                          label={t('crm.leadDetail.contactOptional')}
                          value={tpForm.contactId}
                          onChange={(e) => setTpForm((f) => ({ ...f, contactId: e.target.value }))}
                        >
                          <option value="">{t('crm.leadDetail.none')}</option>
                          {(contacts ?? []).map((c) => (
                            <option key={c.id} value={String(c.id)}>
                              {c.name}
                            </option>
                          ))}
                        </Select>
                      </div>
                      <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                        <Button variant="ghost" onClick={() => setTpOpen(false)}>
                          {t('crmCommon.cancel')}
                        </Button>
                        <Button
                          variant="primary"
                          onClick={submitTouchpoint}
                          disabled={createTouchpoint.isPending || !tpForm.notes.trim()}
                        >
                          {createTouchpoint.isPending ? 'Saving...' : 'Log'}
                        </Button>
                      </div>
                    </div>
                  </Card>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Timeline */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {touchpointsFailed && (
                <InlineLoadError onRetry={() => void touchpointsQuery.refetch()} />
              )}
              {!touchpointsFailed && sortedTouchpoints.length === 0 && (
                <div
                  style={{
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    borderRadius: '10px',
                  }}
                >
                  <EmptyState
                    size="compact"
                    icon={Clock}
                    title={t('crm.leadDetail.noTouchpointsTitle')}
                    description={t('crm.leadDetail.noTouchpointsDescription')}
                    action={{
                      label: t('crm.leadDetail.logTouchpoint'),
                      onClick: () => setTpOpen(true),
                    }}
                  />
                </div>
              )}
              {sortedTouchpoints.map((tp) => {
                const outcomeStyle = OUTCOME_COLORS[tp.outcome];
                return (
                  <Card key={tp.id}>
                    <div style={{ display: 'flex', gap: '12px' }}>
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '8px',
                          background: 'var(--hub-hover)',
                          color: 'var(--hub-text-muted)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          flexShrink: 0,
                        }}
                      >
                        {TOUCHPOINT_TYPE_ICONS[tp.type]}
                      </div>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            flexWrap: 'wrap',
                          }}
                        >
                          <span
                            style={{
                              fontSize: '11px',
                              fontWeight: 600,
                              padding: '2px 7px',
                              borderRadius: '5px',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '3px',
                              background: outcomeStyle.bg,
                              color: outcomeStyle.color,
                            }}
                          >
                            {OUTCOME_ICONS[tp.outcome]}
                            {t(OUTCOME_LABEL_KEYS[tp.outcome])}
                          </span>
                          <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>
                            {formatDate(tp.createdAt)}
                          </span>
                          {tp.repName && (
                            <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>
                              · {tp.repName}
                            </span>
                          )}
                        </div>
                        <p
                          style={{
                            fontSize: '13px',
                            color: 'var(--hub-text)',
                            margin: '6px 0 0',
                            lineHeight: 1.5,
                          }}
                        >
                          {tp.notes}
                        </p>
                        {tp.nextFollowUp && (
                          <p
                            style={{
                              fontSize: '11px',
                              color: 'var(--hub-text-muted)',
                              margin: '4px 0 0',
                            }}
                          >
                            {t('crm.leadDetail.followUpOn', {
                              date: formatDay(tp.nextFollowUp, i18n.language),
                            })}
                          </p>
                        )}
                      </div>
                      <button
                        type="button"
                        aria-label={t('crm.leadDetail.deleteTouchpoint')}
                        onClick={() => setDeleteTpId(tp.id)}
                        style={{
                          background: 'transparent',
                          border: 'none',
                          cursor: 'pointer',
                          color: 'var(--hub-text-dim)',
                          padding: '4px',
                          display: 'flex',
                          alignItems: 'center',
                        }}
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  </Card>
                );
              })}
            </div>
          </div>

          {/* Right: info panels */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
            {/* Lead info */}
            <Card>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '14px',
                }}
              >
                <p
                  style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}
                >
                  {t('crm.leadDetail.infoTitle')}
                </p>
                <button
                  type="button"
                  aria-label={t('crm.createLead.editTitle')}
                  onClick={() => setEditOpen(true)}
                  style={{
                    background: 'transparent',
                    border: 'none',
                    padding: '4px',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                    display: 'flex',
                  }}
                >
                  <Edit2 size={13} />
                </button>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                <InfoRow
                  label={t('crm.leadDetail.company')}
                  value={lead.companyName ?? lead.company ?? '—'}
                />
                <InfoRow label={t('crm.leadDetail.industry')} value={lead.industry ?? '—'} />
                <InfoRow label={t('crm.leadDetail.size')} value={lead.size ?? '—'} />
                <InfoRow label={t('crm.leadDetail.source')} value={lead.source ?? '—'} />
                <InfoRow
                  label={t('crm.leadDetail.mrrPotential')}
                  value={`$${(lead.mrrPotential ?? lead.dealValue ?? 0).toLocaleString()}/mo`}
                />
                <InfoRow
                  label={t('crm.leadDetail.assignedRep')}
                  value={
                    lead.assignedTo != null
                      ? (repName(lead.assignedTo) ?? t('crm.leads.formerRep'))
                      : t('crm.createLead.unassigned')
                  }
                />
                <InfoRow
                  label={t('crm.leadDetail.nextFollowUp')}
                  value={
                    lead.nextFollowUp
                      ? followUpStatus(lead.nextFollowUp) === 'upcoming'
                        ? formatDay(lead.nextFollowUp, i18n.language)
                        : `${formatDay(lead.nextFollowUp, i18n.language)} · ${t(
                            `crm.leads.followUp.${followUpStatus(lead.nextFollowUp)}`
                          )}`
                      : '—'
                  }
                />
                {lead.notes && (
                  <InfoRow label={t('crm.leadDetail.notesLabel')} value={lead.notes} />
                )}
                <div>
                  <p
                    style={{
                      fontSize: '11px',
                      color: 'var(--hub-text-dim)',
                      marginBottom: '4px',
                      textTransform: 'uppercase',
                      letterSpacing: '0.05em',
                      fontWeight: 600,
                    }}
                  >
                    {t('crm.leadDetail.stage')}
                  </p>
                  <select
                    value={lead.stage}
                    onChange={(e) => handleStageChange(e.target.value as LeadStage)}
                    style={{
                      width: '100%',
                      height: '34px',
                      padding: '0 10px',
                      borderRadius: '8px',
                      background: 'var(--hub-input-bg)',
                      border: '1px solid var(--hub-border)',
                      color: 'var(--hub-text)',
                      fontSize: '13px',
                      outline: 'none',
                      cursor: 'pointer',
                    }}
                  >
                    {STAGE_VALUES.map((stage) => (
                      <option key={stage} value={stage}>
                        {t(`crm.leadStages.${stage}`)}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </Card>

            {/* Contacts */}
            <Card>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '12px',
                }}
              >
                <p
                  style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}
                >
                  {t('crm.leadDetail.contactsTitle')}
                </p>
                <button
                  type="button"
                  aria-label={t('crm.leadDetail.addContactAria')}
                  onClick={() => setContactOpen((v) => !v)}
                  style={{
                    background: 'transparent',
                    border: '1px solid var(--hub-border)',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                    borderRadius: '6px',
                    padding: '3px 8px',
                    fontSize: '11px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                  }}
                >
                  <Plus size={11} />
                  {t('crm.leadDetail.add')}
                </button>
              </div>

              <AnimatePresence>
                {contactOpen && (
                  <motion.div
                    initial={{ opacity: 0, y: -4 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -4 }}
                    style={{
                      marginBottom: '12px',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '8px',
                    }}
                  >
                    <Input
                      label={t('crmCommon.name')}
                      value={contactForm.name}
                      onChange={(e) => setContactForm((f) => ({ ...f, name: e.target.value }))}
                      placeholder={t('crm.leadDetail.namePlaceholder')}
                    />
                    <Input
                      label={t('crm.leadDetail.roleLabel')}
                      value={contactForm.role}
                      onChange={(e) => setContactForm((f) => ({ ...f, role: e.target.value }))}
                      placeholder={t('crm.leadDetail.rolePlaceholder')}
                    />
                    <Input
                      label={t('crmCommon.phone')}
                      value={contactForm.phone}
                      onChange={(e) => setContactForm((f) => ({ ...f, phone: e.target.value }))}
                      placeholder={t('crm.leadDetail.phonePlaceholder')}
                    />
                    <Input
                      label={t('crmCommon.email')}
                      type="email"
                      value={contactForm.email}
                      onChange={(e) => setContactForm((f) => ({ ...f, email: e.target.value }))}
                      placeholder="jane@company.com"
                    />
                    <div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end' }}>
                      <Button variant="ghost" onClick={() => setContactOpen(false)}>
                        {t('crmCommon.cancel')}
                      </Button>
                      <Button
                        variant="primary"
                        onClick={submitContact}
                        disabled={createContact.isPending || !contactForm.name.trim()}
                        style={{ padding: '6px 12px', fontSize: '12px' }}
                      >
                        {createContact.isPending
                          ? t('crm.leadDetail.adding')
                          : t('crm.leadDetail.add')}
                      </Button>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>

              {contactsFailed && <InlineLoadError onRetry={() => void contactsQuery.refetch()} />}
              {!contactsFailed && (contacts ?? []).length === 0 && !contactOpen && (
                <EmptyState
                  size="compact"
                  icon={UserPlus}
                  title={t('crm.leadDetail.noContactsTitle')}
                  description={t('crm.leadDetail.noContactsDescription')}
                  action={{
                    label: t('crm.leadDetail.addContact'),
                    onClick: () => setContactOpen(true),
                  }}
                />
              )}
              {(contacts ?? []).map((c) => (
                <div
                  key={c.id}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '10px',
                    paddingBottom: '10px',
                    borderBottom: '1px solid var(--hub-border)',
                    marginBottom: '10px',
                  }}
                >
                  <div
                    style={{
                      width: '28px',
                      height: '28px',
                      borderRadius: '6px',
                      background: 'var(--dash-hover)',
                      color: 'var(--accent-fg)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      fontWeight: 700,
                      flexShrink: 0,
                    }}
                  >
                    {c.name[0]?.toUpperCase() ?? '?'}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p
                      style={{
                        fontSize: '12px',
                        fontWeight: 600,
                        color: 'var(--hub-text)',
                        margin: 0,
                      }}
                    >
                      {c.name}
                    </p>
                    <p
                      style={{
                        fontSize: '11px',
                        color: 'var(--hub-text-muted)',
                        margin: '1px 0 0',
                      }}
                    >
                      {c.role}
                    </p>
                    {c.phone && (
                      <p
                        style={{
                          fontSize: '11px',
                          color: 'var(--hub-text-dim)',
                          margin: '2px 0 0',
                        }}
                      >
                        {c.phone}
                      </p>
                    )}
                    {c.email && (
                      <p
                        style={{
                          fontSize: '11px',
                          color: 'var(--hub-text-dim)',
                          margin: '1px 0 0',
                        }}
                      >
                        {c.email}
                      </p>
                    )}
                  </div>
                  <button
                    type="button"
                    aria-label={t('crm.leadDetail.deleteContact')}
                    onClick={() => setDeleteContactId(c.id)}
                    style={{
                      background: 'transparent',
                      border: 'none',
                      cursor: 'pointer',
                      color: 'var(--hub-text-dim)',
                      padding: '4px',
                      display: 'flex',
                      alignItems: 'center',
                    }}
                  >
                    <Trash2 size={11} />
                  </button>
                </div>
              ))}
            </Card>

            {/* Contract */}
            <Card>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '12px',
                }}
              >
                <p
                  style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}
                >
                  {t('crm.leadDetail.contractTitle')}
                </p>
                {!contract && contractsQuery.isSuccess && !contractsQuery.isFetching && (
                  <button
                    type="button"
                    aria-label={t('crm.leadDetail.addContractAria')}
                    onClick={() => setContractOpen((v) => !v)}
                    style={{
                      background: 'transparent',
                      border: '1px solid var(--hub-border)',
                      cursor: 'pointer',
                      color: 'var(--hub-text-muted)',
                      borderRadius: '6px',
                      padding: '3px 8px',
                      fontSize: '11px',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                  >
                    <Plus size={11} />
                    {t('crm.leadDetail.add')}
                  </button>
                )}
              </div>

              {!contract && contractsFailed ? (
                <InlineLoadError onRetry={() => void contractsQuery.refetch()} />
              ) : !contract && contractsQuery.isLoading ? (
                <p role="status" style={{ fontSize: '12px', color: 'var(--hub-text-muted)' }}>
                  {t('crmCommon.loading')}
                </p>
              ) : contract ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span
                      style={{
                        fontSize: '11px',
                        fontWeight: 600,
                        padding: '2px 8px',
                        borderRadius: '6px',
                        background: 'rgba(16,185,129,0.12)',
                        color: '#10b981',
                        textTransform: 'capitalize',
                      }}
                    >
                      {contract.plan}
                    </span>
                  </div>
                  <InfoRow
                    label={t('crm.leadDetail.mrr')}
                    value={`$${contract.mrr.toLocaleString()}/mo`}
                  />
                  <InfoRow label={t('crm.leadDetail.seats')} value={String(contract.seats)} />
                  <InfoRow
                    label={t('crm.leadDetail.signed')}
                    value={formatDate(contract.signedAt)}
                  />
                  <InfoRow label={t('crm.leadDetail.signedBy')} value={contract.signedBy} />
                  <InfoRow label={t('crm.leadDetail.closedBy')} value={contract.closedBy} />
                  {contract.notes && (
                    <p
                      style={{
                        fontSize: '12px',
                        color: 'var(--hub-text-muted)',
                        margin: '4px 0 0',
                        lineHeight: 1.5,
                      }}
                    >
                      {contract.notes}
                    </p>
                  )}
                </div>
              ) : (
                <>
                  {!contractOpen && (
                    <p style={{ fontSize: '12px', color: 'var(--hub-text-dim)' }}>
                      {t('crm.leadDetail.noContract')}
                    </p>
                  )}
                  <AnimatePresence>
                    {contractOpen && (
                      <motion.div
                        initial={{ opacity: 0, y: -4 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -4 }}
                        style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}
                      >
                        <Select
                          label={t('crm.leadDetail.planLabel')}
                          value={contractForm.plan}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, plan: e.target.value as ContractPlan }))
                          }
                        >
                          <option value="free">{t('crm.leadDetail.planFree')}</option>
                          <option value="pro">{t('crm.leadDetail.planPro')}</option>
                          <option value="scale">{t('crm.leadDetail.planScale')}</option>
                        </Select>
                        <Input
                          label={t('crm.leadDetail.mrrUsd')}
                          type="number"
                          value={contractForm.mrr}
                          onChange={(e) => setContractForm((f) => ({ ...f, mrr: e.target.value }))}
                          placeholder="500"
                        />
                        <Input
                          label={t('crm.leadDetail.seats')}
                          type="number"
                          value={contractForm.seats}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, seats: e.target.value }))
                          }
                          placeholder="5"
                        />
                        <Input
                          label={t('crm.leadDetail.signedAt')}
                          type="date"
                          value={contractForm.signedAt}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, signedAt: e.target.value }))
                          }
                        />
                        <Input
                          label={t('crm.leadDetail.signedBy')}
                          value={contractForm.signedBy}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, signedBy: e.target.value }))
                          }
                          placeholder={t('crm.leadDetail.signedByPlaceholder')}
                        />
                        <Input
                          label={t('crm.leadDetail.closedByLabel')}
                          value={contractForm.closedBy}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, closedBy: e.target.value }))
                          }
                          placeholder={t('crm.leadDetail.closedByPlaceholder')}
                        />
                        <Textarea
                          label={t('crm.leadDetail.notesLabel')}
                          value={contractForm.notes}
                          onChange={(e) =>
                            setContractForm((f) => ({ ...f, notes: e.target.value }))
                          }
                          placeholder={t('crm.leadDetail.contractNotesPlaceholder')}
                        />
                        <div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end' }}>
                          <Button variant="ghost" onClick={() => setContractOpen(false)}>
                            {t('crmCommon.cancel')}
                          </Button>
                          <Button
                            variant="primary"
                            onClick={submitContract}
                            disabled={createContract.isPending}
                            style={{ padding: '6px 12px', fontSize: '12px' }}
                          >
                            {createContract.isPending ? 'Saving...' : 'Save'}
                          </Button>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </>
              )}
            </Card>
          </div>
        </div>

        {/* Delete lead confirm */}
        <LeadFormModal open={editOpen} onClose={() => setEditOpen(false)} lead={lead} />

        <ConfirmModal
          open={confirmDeleteLead}
          onClose={() => setConfirmDeleteLead(false)}
          onConfirm={handleDeleteLead}
          title={t('crm.leadDetail.deleteLead')}
          description={`Permanently delete "${lead.company}" and all associated data?`}
          confirmLabel={t('crmCommon.delete')}
          variant="danger"
        />

        {/* Delete touchpoint confirm */}
        <ConfirmModal
          open={deleteTpId !== null}
          onClose={() => setDeleteTpId(null)}
          onConfirm={() => {
            if (deleteTpId !== null) deleteTouchpoint.mutate(deleteTpId);
            setDeleteTpId(null);
          }}
          title={t('crm.leadDetail.deleteTouchpoint')}
          description={t('crm.leadDetail.deleteTouchpointMessage')}
          confirmLabel={t('crmCommon.delete')}
          variant="danger"
        />

        {/* Delete contact confirm */}
        <ConfirmModal
          open={deleteContactId !== null}
          onClose={() => setDeleteContactId(null)}
          onConfirm={() => {
            if (deleteContactId !== null && leadId !== undefined) {
              deleteContact.mutate({ id: deleteContactId, leadId });
            }
            setDeleteContactId(null);
          }}
          title={t('crm.leadDetail.removeContact')}
          description={t('crm.leadDetail.removeContactMessage')}
          confirmLabel={t('crm.leadDetail.remove')}
          variant="danger"
        />
      </div>
    </>
  );
};

const InfoRow: React.FC<{ label: string; value: string }> = ({ label, value }) => (
  <div>
    <p
      style={{
        fontSize: '10px',
        color: 'var(--hub-text-dim)',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        fontWeight: 600,
        margin: '0 0 2px',
      }}
    >
      {label}
    </p>
    <p style={{ fontSize: '13px', color: 'var(--hub-text)', margin: 0 }}>{value || '-'}</p>
  </div>
);

export default LeadDetail;
