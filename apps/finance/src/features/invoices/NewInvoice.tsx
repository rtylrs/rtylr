import React, { useEffect, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { getRouteApi, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { ChevronRight, Plus, ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import PageHeader from '../../shared/components/PageHeader';
import ClientSelect, { Client } from '../../shared/components/ClientSelect';
import LineItemRow, { LineItem } from '../../shared/components/LineItemRow';
import {
  useSignedContracts,
  type SignedContract,
} from '../../shared/hooks/queries/useSignedContracts';
import TaxBreakdown from '../../shared/components/TaxBreakdown';
import InvoicePreview from '../../shared/components/InvoicePreview';
import { useTranslation } from 'react-i18next';
import { useCreateInvoice, useInvoiceDetail, useUpdateInvoiceStatus } from './useInvoices';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  inputStyle,
  labelStyle,
  ACCENT,
} from '../../shared/lib/tokens';
import { useSettings } from '../../shared/lib/settings';
import { unwrap } from '@voxire/api';
import { listCrmCustomers } from '../../shared/lib/clients';
import { parseClientsResponse } from '../../shared/schemas/clients';
import { toLocalIso, todayIso } from '../../shared/lib/date';

const newInvoiceRoute = getRouteApi('/layout/invoices/new');

const STEPS = ['Client', 'Items', 'Tax & Notes', 'Preview'] as const;

const NewInvoice: React.FC = () => {
  const { t } = useTranslation('finance');
  const navigate = useNavigate();
  const { clone, client: clientId } = newInvoiceRoute.useSearch();
  const settings = useSettings();
  const [step, setStep] = useState(0);
  const [client, setClient] = useState<Client | null>(null);
  const today = todayIso();
  const due = toLocalIso(new Date(Date.now() + 30 * 86_400_000));
  const [issuedDate, setIssuedDate] = useState(today);
  const [dueDate, setDueDate] = useState(due);
  const [invoiceNumber, setInvoiceNumber] = useState(
    `${settings.invoicePrefix}-${String(Date.now()).slice(-6)}`
  );
  const [items, setItems] = useState<LineItem[]>([
    { id: 'l1', description: '', quantity: 1, unitPrice: 0, taxRate: settings.defaultTaxRate },
  ]);
  const [notes, setNotes] = useState(settings.invoiceNotes);
  const [discount, setDiscount] = useState(0);
  // CRM->FIN-01 (#839): signed contracts offered as one-click prefills.
  const { data: signedContracts } = useSignedContracts();
  const [prefilledContractId, setPrefilledContractId] = useState('');
  const createInvoice = useCreateInvoice();
  const updateInvoiceStatus = useUpdateInvoiceStatus();
  const { data: sourceInvoice } = useInvoiceDetail(clone);

  useEffect(() => {
    if (sourceInvoice) {
      setClient({
        crmCustomerId: undefined,
        id: sourceInvoice.id,
        name: sourceInvoice.client,
        email: sourceInvoice.clientEmail,
      });
      setIssuedDate(sourceInvoice.issuedDate ?? today);
      setDueDate(sourceInvoice.dueDate ?? due);
      setInvoiceNumber(sourceInvoice.number ?? sourceInvoice.id);
      setItems(
        (sourceInvoice.items ?? []).map((item, index) => ({
          id: `clone-${index}`,
          description: item.description ?? '',
          quantity: item.quantity ?? 1,
          unitPrice: item.unitPrice ?? 0,
          taxRate: settings.defaultTaxRate,
        }))
      );
      setNotes(sourceInvoice.notes ?? settings.invoiceNotes);
      setDiscount(sourceInvoice.discount ?? 0);
    }
    // Clone prefill only: re-running on the derived defaults (today/due)
    // or settings would wipe in-progress edits, so this intentionally
    // syncs from the clone source alone.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sourceInvoice]);

  useEffect(() => {
    if (!clientId || clone) return;
    void unwrap(listCrmCustomers())
      .then((data) => {
        const clients = parseClientsResponse(data);
        const selected = (clients ?? []).find((entry) => entry.id === clientId);
        if (selected) setClient(selected);
      })
      .catch(() => {});
  }, [clientId, clone]);

  const addItem = () =>
    setItems((prev) => [
      ...prev,
      {
        id: `l${prev.length + 1}-${Date.now()}`,
        description: '',
        quantity: 1,
        unitPrice: 0,
        taxRate: settings.defaultTaxRate,
      },
    ]);

  const prefillFromContract = (contract: SignedContract): void => {
    const seats = contract.seats ?? 0;
    const mrr = Number(contract.mrr ?? 0);
    setItems([
      {
        id: 'c1',
        description: `${contract.plan ?? 'Subscription'} — ${seats} seat${seats === 1 ? '' : 's'}`,
        quantity: 1,
        unitPrice: mrr,
        taxRate: settings.defaultTaxRate,
      },
    ]);
    setNotes(
      `Prefilled from signed CRM contract #${contract.id}` +
        (contract.signedAt ? ` (signed ${contract.signedAt})` : '') +
        `\n${settings.invoiceNotes}`
    );
  };

  const updateItem = (i: number, item: LineItem) =>
    setItems((prev) => prev.map((it, idx) => (idx === i ? item : it)));

  const removeItem = (i: number) => setItems((prev) => prev.filter((_, idx) => idx !== i));

  const canProceed = () => {
    if (step === 0) return !!client;
    if (step === 1)
      return (
        items.length > 0 &&
        items.every((it) => it.description && it.unitPrice > 0 && (it.quantity ?? 1) > 0)
      );
    return true;
  };

  const handleSave = (status: 'Draft' | 'Sent'): void => {
    if (!client) {
      toast.error(t('invoices.new.selectClientFirst'));
      return;
    }
    const mappedItems = items.map((it, idx) => ({
      description: it.description,
      quantity: it.quantity,
      unit: 'item',
      unitPrice: it.unitPrice,
      taxRate: it.taxRate,
      discountPct: 0,
      sortOrder: String(idx),
      total: it.quantity * it.unitPrice,
    }));
    const payload: CreateInvoiceRequest = {
      customerId: client.crmCustomerId,
      clientName: client.name,
      clientEmail: client.email ?? '',
      clientAddress: '',
      clientTaxId: '',
      issueDate: issuedDate,
      dueDate,
      currency: settings.currency,
      discountType: discount > 0 ? 'fixed' : 'none',
      discountValue: discount,
      taxRate: settings.defaultTaxRate,
      notes,
      terms: '',
      items: mappedItems,
    };
    createInvoice.mutate(payload, {
      onSuccess: (created) => {
        // Create always persists a Draft. For "Save & Send", transition it to
        // Sent so the toast isn't lying about what happened.
        if (status === 'Sent' && created?.id) {
          updateInvoiceStatus.mutate(
            { id: created.id, status: 'Sent' },
            {
              onSuccess: () => {
                toast.success(`Invoice ${invoiceNumber} sent`);
                navigate({ to: '/invoices' });
              },
              onError: (error) => {
                toast.error(
                  getOneKitErrorMessage(
                    error,
                    `Invoice ${invoiceNumber} saved as draft — couldn't mark it sent`
                  )
                );
                navigate({ to: '/invoices' });
              },
            }
          );
          return;
        }
        toast.success(`Invoice ${invoiceNumber} saved as Draft`);
        navigate({ to: '/invoices' });
      },
      onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to save invoice')),
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('common.pageTitles.newInvoice'))} />
      <PageHeader
        title={t('invoices.new.title')}
        breadcrumbs={[{ label: t('nav.invoices'), to: '/invoices' }, { label: t('common.new') }]}
      />

      {/* Stepper */}
      <div style={{ ...cardStyle, padding: 12, display: 'flex', alignItems: 'center', gap: 8 }}>
        {STEPS.map((s, i) => (
          <React.Fragment key={s}>
            <button
              type="button"
              onClick={() => setStep(i)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 12px',
                borderRadius: 7,
                border: 'none',
                cursor: 'pointer',
                background: step === i ? ACCENT : 'transparent',
                color:
                  step === i ? 'var(--finance-white)' : step > i ? ACCENT : 'var(--hub-text-muted)',
                fontSize: 12,
                fontWeight: 600,
              }}
            >
              <span
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 18,
                  height: 18,
                  borderRadius: 9,
                  background: step === i ? 'var(--finance-white-overlay)' : 'var(--hub-hover)',
                  fontSize: 10,
                }}
              >
                {i + 1}
              </span>
              {s}
            </button>
            {i < STEPS.length - 1 && (
              <ChevronRight size={12} style={{ color: 'var(--hub-text-dim)' }} />
            )}
          </React.Fragment>
        ))}
      </div>

      <div
        style={{ display: 'grid', gridTemplateColumns: step === 3 ? '1fr 1fr' : '1fr', gap: 20 }}
      >
        <div style={{ ...cardStyle, padding: 20 }}>
          {step === 0 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div>
                <label style={labelStyle}>{t('invoices.new.client')}</label>
                <div style={{ marginTop: 6 }}>
                  <ClientSelect value={client} onChange={setClient} />
                </div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
                <div>
                  <label style={labelStyle}>{t('invoices.new.number')}</label>
                  <input
                    value={invoiceNumber}
                    onChange={(e) => setInvoiceNumber(e.target.value)}
                    style={{ ...inputStyle, marginTop: 6 }}
                  />
                </div>
                <div>
                  <label style={labelStyle}>{t('invoices.new.issued')}</label>
                  <input
                    type="date"
                    value={issuedDate}
                    onChange={(e) => setIssuedDate(e.target.value)}
                    style={{ ...inputStyle, marginTop: 6 }}
                  />
                </div>
                <div>
                  <label style={labelStyle}>{t('invoices.new.due')}</label>
                  <input
                    type="date"
                    value={dueDate}
                    onChange={(e) => setDueDate(e.target.value)}
                    style={{ ...inputStyle, marginTop: 6 }}
                  />
                </div>
              </div>
            </div>
          )}

          {step === 1 && (
            <div>
              {signedContracts && signedContracts.length > 0 && (
                <div
                  style={{
                    ...cardStyle,
                    padding: 12,
                    marginBottom: 14,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    flexWrap: 'wrap',
                  }}
                >
                  <label
                    htmlFor="contract-prefill"
                    style={{ fontSize: 12, color: 'var(--hub-text-muted)', fontWeight: 500 }}
                  >
                    {t('finance.invoices.prefillFromContract')}
                  </label>
                  <select
                    id="contract-prefill"
                    value={prefilledContractId}
                    onChange={(e) => {
                      const id = e.target.value;
                      setPrefilledContractId(id);
                      const c = signedContracts.find((x) => x.id === id);
                      if (c) prefillFromContract(c);
                    }}
                    style={{
                      padding: '7px 10px',
                      borderRadius: 8,
                      border: '1px solid var(--hub-border)',
                      background: 'var(--hub-surface)',
                      color: 'var(--hub-text)',
                      fontSize: 16,
                      minWidth: 220,
                    }}
                  >
                    <option value="">{t('finance.invoices.chooseContract')}</option>
                    {signedContracts.map((c) => (
                      <option key={c.id} value={c.id}>
                        {`${c.plan ?? 'Contract'} — ${c.mrr ?? '0'}/mo${c.signedAt ? ` (${c.signedAt})` : ''}`}
                      </option>
                    ))}
                  </select>
                </div>
              )}
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    {['Description', 'Qty', 'Price', 'Tax %', 'Total', ''].map((h) => (
                      <th
                        key={h}
                        style={{
                          padding: '8px',
                          textAlign:
                            h === 'Total' || h === 'Qty' || h === 'Tax %' ? 'end' : 'start',
                          fontSize: 10,
                          textTransform: 'uppercase',
                          letterSpacing: '0.06em',
                          color: 'var(--hub-text-dim)',
                        }}
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {items.map((it, i) => (
                    <LineItemRow
                      key={it.id}
                      item={it}
                      onChange={(v) => updateItem(i, v)}
                      onRemove={() => removeItem(i)}
                    />
                  ))}
                </tbody>
              </table>
              <button type="button" style={{ ...ghostBtn, marginTop: 12 }} onClick={addItem}>
                <Plus size={13} /> Add line item
              </button>
            </div>
          )}

          {step === 2 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <TaxBreakdown items={items} discount={discount} />
              <div>
                <label style={labelStyle}>{t('invoices.new.discount')}</label>
                <input
                  type="number"
                  value={discount}
                  onChange={(e) => {
                    const v = parseFloat(e.target.value);
                    if (!isNaN(v) && v >= 0) setDiscount(v);
                  }}
                  style={{ ...inputStyle, marginTop: 6 }}
                />
              </div>
              <div>
                <label style={labelStyle}>{t('invoices.new.notes')}</label>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  rows={4}
                  style={{ ...inputStyle, marginTop: 6, resize: 'vertical', fontFamily: 'inherit' }}
                />
              </div>
            </div>
          )}

          {step === 3 && (
            <div style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
              Review the invoice on the right. Click <strong>{t('invoices.new.saveDraft')}</strong>{' '}
              or <strong>{t('invoices.new.saveAndSend')}</strong> to finalize.
            </div>
          )}
        </div>

        {step === 3 && (
          <InvoicePreview
            invoiceNumber={invoiceNumber}
            client={client}
            issuedDate={issuedDate}
            dueDate={dueDate}
            items={items}
            notes={notes}
            discount={discount}
          />
        )}
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <button
          type="button"
          style={ghostBtn}
          onClick={() => (step === 0 ? navigate({ to: '/invoices' }) : setStep(step - 1))}
        >
          <ArrowLeft size={13} /> {step === 0 ? 'Cancel' : 'Back'}
        </button>
        <div style={{ display: 'flex', gap: 8 }}>
          {step < STEPS.length - 1 ? (
            <button
              type="button"
              style={primaryBtn}
              disabled={!canProceed()}
              onClick={() => setStep(step + 1)}
            >
              Next <ChevronRight size={13} />
            </button>
          ) : (
            <>
              <button
                type="button"
                style={{ ...ghostBtn, opacity: createInvoice.isPending ? 0.6 : 1 }}
                onClick={() => handleSave('Draft')}
                disabled={createInvoice.isPending}
              >
                <Save size={13} /> {createInvoice.isPending ? 'Saving...' : 'Save as Draft'}
              </button>
              <button
                type="button"
                style={{ ...primaryBtn, opacity: createInvoice.isPending ? 0.6 : 1 }}
                onClick={() => handleSave('Sent')}
                disabled={createInvoice.isPending}
              >
                {createInvoice.isPending ? 'Saving...' : 'Save & Send'}
              </button>
            </>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default NewInvoice;
