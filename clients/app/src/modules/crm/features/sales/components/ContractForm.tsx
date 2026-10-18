import React from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { Link } from '@tanstack/react-router';
import { Input, Select, Textarea } from '../../../shared/components/ui';
import { useSalesReps } from '../../../shared/hooks/queries/useSales';
import type { Contract, ContractBilling } from '../../../shared/hooks/queries/useSales';
import type { CreateContractRequest } from '../../../shared/services/crmApi';
import { formatDay, todayIso } from '../../../shared/lib/dateUtils';
import { formatMoney } from '../../../../../shared/lib/money';

/** A contract being written, as the inputs hold it. */
export interface ContractDraft {
  plan: string;
  value: string;
  billing: ContractBilling;
  startDate: string;
  endDate: string;
  signedAt: string;
  signedBy: string;
  closedBy: string;
  notes: string;
}

export const EMPTY_CONTRACT: ContractDraft = {
  plan: '',
  value: '',
  billing: 'monthly',
  startDate: '',
  endDate: '',
  signedAt: '',
  signedBy: '',
  closedBy: '',
  notes: '',
};

const BILLINGS: ContractBilling[] = ['one_time', 'monthly', 'yearly'];

/** The translation key of the first thing stopping the draft from saving. */
export function contractDraftIssue(draft: ContractDraft): string | null {
  if (!draft.plan.trim()) return 'crm.contractForm.planRequired';
  const value = Number(draft.value);
  if (draft.value.trim() === '' || !Number.isFinite(value) || value < 0) {
    return 'crm.contractForm.valueRequired';
  }
  if (draft.startDate && draft.endDate && draft.endDate < draft.startDate) {
    return 'crm.contractForm.endBeforeStart';
  }
  return null;
}

export function toContractRequest(draft: ContractDraft): Omit<CreateContractRequest, 'leadId'> {
  return {
    plan: draft.plan.trim(),
    value: Number(draft.value),
    billing: draft.billing,
    startDate: draft.startDate || undefined,
    endDate: draft.endDate || undefined,
    signedAt: draft.signedAt || undefined,
    signedBy: draft.signedBy.trim(),
    closedBy: draft.closedBy ? Number(draft.closedBy) : null,
    notes: draft.notes,
  };
}

/** "$1,200.00 / month", "$300.00 one-time". */
export function formatContractValue(
  contract: Pick<Contract, 'value' | 'billing'>,
  currency: string,
  t: TFunction
): string {
  return t(`crm.contractForm.valueBy.${contract.billing}`, {
    amount: formatMoney(contract.value, currency),
  });
}

export type ContractStatus = 'draft' | 'signed' | 'ended';

/** Unsigned contracts are drafts; signed ones end once their end date passes. */
export function contractStatus(contract: Pick<Contract, 'signedAt' | 'endDate'>): ContractStatus {
  if (!contract.signedAt) return 'draft';
  if (contract.endDate && contract.endDate < todayIso()) return 'ended';
  return 'signed';
}

const STATUS_COLORS: Record<ContractStatus, { bg: string; color: string }> = {
  draft: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  signed: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
  ended: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
};

export const ContractStatusBadge: React.FC<{
  contract: Pick<Contract, 'signedAt' | 'endDate'>;
}> = ({ contract }) => {
  const { t } = useTranslation();
  const status = contractStatus(contract);
  return (
    <span
      style={{
        fontSize: '11px',
        fontWeight: 600,
        padding: '2px 8px',
        borderRadius: '6px',
        background: STATUS_COLORS[status].bg,
        color: STATUS_COLORS[status].color,
        whiteSpace: 'nowrap',
      }}
    >
      {t(`crm.contractForm.status.${status}`)}
    </span>
  );
};

/** The inputs for a contract's terms, shared by the lead page and Contracts. */
export const ContractFields: React.FC<{
  draft: ContractDraft;
  onChange: (patch: Partial<ContractDraft>) => void;
  currency: string;
}> = ({ draft, onChange, currency }) => {
  const { t } = useTranslation();
  const { data: reps = [] } = useSalesReps();
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
      <Input
        label={t('crm.contractForm.plan')}
        value={draft.plan}
        onChange={(e) => onChange({ plan: e.target.value })}
        placeholder={t('crm.contractForm.planPlaceholder')}
      />
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
        <Input
          label={t('crm.contractForm.value', { currency })}
          type="number"
          min={0}
          step="any"
          value={draft.value}
          onChange={(e) => onChange({ value: e.target.value })}
        />
        <Select
          label={t('crm.contractForm.billing')}
          value={draft.billing}
          onChange={(e) => onChange({ billing: e.target.value as ContractBilling })}
        >
          {BILLINGS.map((billing) => (
            <option key={billing} value={billing}>
              {t(`crm.contractForm.billingOption.${billing}`)}
            </option>
          ))}
        </Select>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
        <Input
          label={t('crm.contractForm.startDate')}
          type="date"
          value={draft.startDate}
          onChange={(e) => onChange({ startDate: e.target.value })}
        />
        <Input
          label={t('crm.contractForm.endDate')}
          type="date"
          value={draft.endDate}
          min={draft.startDate || undefined}
          onChange={(e) => onChange({ endDate: e.target.value })}
        />
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
        <Input
          label={t('crm.contractForm.signedAt')}
          type="date"
          value={draft.signedAt}
          onChange={(e) => onChange({ signedAt: e.target.value })}
        />
        <Input
          label={t('crm.leadDetail.signedBy')}
          value={draft.signedBy}
          onChange={(e) => onChange({ signedBy: e.target.value })}
          placeholder={t('crm.leadDetail.signedByPlaceholder')}
        />
      </div>
      <Select
        label={t('crm.leadDetail.closedByLabel')}
        value={draft.closedBy}
        onChange={(e) => onChange({ closedBy: e.target.value })}
      >
        <option value="">{t('crm.createLead.unassigned')}</option>
        {reps.map((rep) => (
          <option key={rep.id} value={String(rep.id)}>
            {rep.name}
          </option>
        ))}
      </Select>
      <Textarea
        label={t('crm.leadDetail.notesLabel')}
        value={draft.notes}
        onChange={(e) => onChange({ notes: e.target.value })}
        placeholder={t('crm.leadDetail.contractNotesPlaceholder')}
      />
      <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: 0 }}>
        {t('crm.contractForm.draftHint')}
      </p>
    </div>
  );
};

const Detail: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
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
    <div style={{ fontSize: '13px', color: 'var(--hub-text)' }}>{children}</div>
  </div>
);

/** A contract's terms, read-only. */
export const ContractSummary: React.FC<{
  contract: Contract;
  currency: string;
  columns?: number;
}> = ({ contract, currency, columns = 2 }) => {
  const { t, i18n } = useTranslation();
  const { data: reps = [] } = useSalesReps();
  const closedBy = reps.find((rep) => rep.id === contract.closedBy)?.name;
  const day = (value?: string): string => (value ? formatDay(value, i18n.language) : '');
  const term =
    contract.startDate || contract.endDate
      ? `${day(contract.startDate) || '…'} – ${day(contract.endDate) || t('crm.contractForm.openEnded')}`
      : '—';
  return (
    <div style={{ display: 'grid', gridTemplateColumns: `repeat(${columns}, 1fr)`, gap: '12px' }}>
      <Detail label={t('crm.contractForm.plan')}>{contract.plan || '—'}</Detail>
      <Detail label={t('crm.contractForm.valueLabel')}>
        {formatContractValue(contract, currency, t)}
      </Detail>
      <Detail label={t('crm.contractForm.term')}>{term}</Detail>
      <Detail label={t('crm.contractForm.statusLabel')}>
        <ContractStatusBadge contract={contract} />
      </Detail>
      <Detail label={t('crm.contractForm.signedAt')}>
        {contract.signedAt
          ? [day(contract.signedAt), contract.signedBy].filter(Boolean).join(' · ')
          : '—'}
      </Detail>
      <Detail label={t('crm.leadDetail.closedBy')}>{closedBy ?? '—'}</Detail>
      {contract.customerId ? (
        <Detail label={t('crm.contractForm.customer')}>
          <Link
            to="/crm/customers/$id"
            params={{ id: String(contract.customerId) }}
            style={{ color: 'var(--crm-accent)' }}
          >
            {t('crm.contractForm.viewCustomer')}
          </Link>
        </Detail>
      ) : null}
      {contract.notes ? (
        <div style={{ gridColumn: '1 / -1' }}>
          <Detail label={t('crm.leadDetail.notesLabel')}>{contract.notes}</Detail>
        </div>
      ) : null}
    </div>
  );
};
