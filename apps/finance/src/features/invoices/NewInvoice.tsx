import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { getRouteApi, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { ChevronRight, Plus, ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import PageHeader from '../../shared/components/PageHeader';
import ClientSelect, { Client } from '../../shared/components/ClientSelect';
import LineItemRow, { LineItem } from '../../shared/components/LineItemRow';
import TaxBreakdown from '../../shared/components/TaxBreakdown';
import InvoicePreview from '../../shared/components/InvoicePreview';
import { useCreateInvoice, useInvoiceDetail, useUpdateInvoiceStatus } from './useInvoices';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  inputStyle,
  labelStyle,
  ACCENT,
} from '../../shared/lib/tokens';
import { computeTaxBreakdown, roundMoney } from '../../shared/lib/tax';
import { useSettings } from '../../shared/lib/settings';
import { unwrap } from '@voxire/api';
import { listCrmCustomers } from '../../shared/lib/clients';
import { parseClientsResponse } from '../../shared/schemas/clients';

const newInvoiceRoute = getRouteApi('/layout/invoices/new');

const STEPS = ['Client', 'Items', 'Tax & Notes', 'Preview'] as const;

// Mirrors finance.invoices.v1 validateCreateInvoiceRequest for client_email.
const EMAIL_PATTERN = /^[^@\s]+@[^@\s]+\.[^@\s]+$/;

const NewInvoice: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { clone, client: clientId } = newInvoiceRoute.useSearch();
  const settings = useSettings();
  const [step, setStep] = useState(0);
  // Stepper clicks may only reach steps already unlocked by a validated Next.
  const [maxStepReached, setMaxStepReached] = useState(0);
  const [client, setClient] = useState<Client | null>(null);
  const today = new Date().toISOString().slice(0, 10);
  const due = new Date(Date.now() + 30 * 86_400_000).toISOString().slice(0, 10);
  const [issuedDate, setIssuedDate] = useState(today);
  const [dueDate, setDueDate] = useState(due);
  const [items, setItems] = useState<LineItem[]>([
    { id: 'l1', description: '', quantity: 1, unitPrice: 0, taxRate: settings.defaultTaxRate },
  ]);
  const [notes, setNotes] = useState(settings.invoiceNotes);
  const [discount, setDiscount] = useState(0);
  const createInvoice = useCreateInvoice();
  const updateInvoiceStatus = useUpdateInvoiceStatus();
  const { data: sourceInvoice } = useInvoiceDetail(clone);

  useEffect(() => {
    if (sourceInvoice) {
      setClient({
        id: sourceInvoice.id,
        name: sourceInvoice.client,
        email: sourceInvoice.clientEmail,
      });
      setIssuedDate(sourceInvoice.issuedDate ?? today);
      setDueDate(sourceInvoice.dueDate ?? due);
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

  const updateItem = (i: number, item: LineItem) =>
    setItems((prev) => prev.map((it, idx) => (idx === i ? item : it)));

  const removeItem = (i: number) => setItems((prev) => prev.filter((_, idx) => idx !== i));

  const clientEmailValid = !!client?.email && EMAIL_PATTERN.test(client.email);

  const itemsValid =
    items.length > 0 &&
    items.every((it) => it.description && it.unitPrice > 0 && (it.quantity ?? 1) > 0);

  const canProceed = () => {
    if (step === 0) return !!client && clientEmailValid;
    if (step === 1) return itemsValid;
    return true;
  };

  const goToStep = (target: number): void => {
    if (target <= maxStepReached) setStep(target);
  };

  const advance = (): void => {
    if (!canProceed()) return;
    setMaxStepReached((prev) => Math.max(prev, step + 1));
    setStep(step + 1);
  };

  // Full pre-flight re-check used by both save handlers - reaching the final
  // step through stepper clicks alone must never enable a guaranteed-throw save.
  const handleSave = (status: 'Draft' | 'Sent'): void => {
    if (!client || !clientEmailValid) {
      setStep(0);
      toast.error(client ? t('invoices.clientEmailInvalid') : t('invoices.clientRequired'));
      return;
    }
    if (!itemsValid) {
      setStep(1);
      toast.error(t('invoices.itemsRequired'));
      return;
    }
    // The payload line total must match what LineItemRow displays:
    // quantity * unitPrice plus per-line tax via computeTaxBreakdown.
    const mappedItems = items.map((it, idx) => {
      const breakdown = computeTaxBreakdown({
        subtotal: roundMoney(it.quantity * it.unitPrice),
        taxRate: it.taxRate,
      });
      return {
        description: it.description,
        quantity: it.quantity,
        unit: 'item',
        unitPrice: it.unitPrice,
        taxRate: it.taxRate,
        discountPct: 0,
        sortOrder: String(idx),
        total: roundMoney(breakdown.total),
      };
    });
    const payload: CreateInvoiceRequest = {
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
        // The invoice number is assigned by the backend; cite the returned
        // value, not any locally invented one.
        const numberLabel = created?.number ?? created?.id;
        // Create always persists a Draft. For "Save & Send", transition it to
        // Sent so the toast isn't lying about what happened.
        if (status === 'Sent' && created?.id) {
          updateInvoiceStatus.mutate(
            { id: created.id, status: 'Sent' },
            {
              onSuccess: () => {
                toast.success(t('invoices.sentSuccess', { number: numberLabel }));
                navigate({ to: '/invoices' });
              },
              onError: (error) => {
                toast.error(
                  getOneKitErrorMessage(
                    error,
                    t('invoices.sendAfterDraftFailed', { number: numberLabel })
                  )
                );
                navigate({ to: '/invoices' });
              },
            }
          );
          return;
        }
        toast.success(t('invoices.draftSaved', { number: numberLabel }));
        navigate({ to: '/invoices' });
      },
      onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="New invoice · rtylr" />
      <PageHeader
        title="New Invoice"
        breadcrumbs={[{ label: 'Invoices', to: '/invoices' }, { label: 'New' }]}
      />

      {/* Stepper */}
      <div style={{ ...cardStyle, padding: 12, display: 'flex', alignItems: 'center', gap: 8 }}>
        {STEPS.map((s, i) => (
          <React.Fragment key={s}>
            <button
              type="button"
              onClick={() => goToStep(i)}
              disabled={i > maxStepReached}
              aria-label={`${t('invoices.goToStep')} ${i + 1}: ${s}`}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 12px',
                borderRadius: 7,
                border: 'none',
                cursor: i > maxStepReached ? 'not-allowed' : 'pointer',
                opacity: i > maxStepReached ? 0.5 : 1,
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
                <label style={labelStyle}>Client</label>
                <div style={{ marginTop: 6 }}>
                  <ClientSelect value={client} onChange={setClient} />
                </div>
                {!clientEmailValid && (
                  <span
                    role="alert"
                    style={{
                      display: 'block',
                      marginTop: 6,
                      fontSize: 11,
                      color: 'var(--finance-danger)',
                    }}
                  >
                    {client ? t('invoices.clientEmailMissing') : t('invoices.clientRequired')}
                  </span>
                )}
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
                <div>
                  <label style={labelStyle}>{t('invoices.numberLabel')}</label>
                  <span
                    style={{
                      display: 'block',
                      ...inputStyle,
                      marginTop: 6,
                      color: 'var(--hub-text-dim)',
                    }}
                  >
                    {t('invoices.autoAssigned')}
                  </span>
                </div>
                <div>
                  <label style={labelStyle}>Issued</label>
                  <input
                    type="date"
                    value={issuedDate}
                    onChange={(e) => setIssuedDate(e.target.value)}
                    style={{ ...inputStyle, marginTop: 6 }}
                  />
                </div>
                <div>
                  <label style={labelStyle}>Due</label>
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
                <label style={labelStyle}>Discount</label>
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
                <label style={labelStyle}>Notes</label>
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
              Review the invoice on the right. Click <strong>Save as Draft</strong> or{' '}
              <strong>Save & Send</strong> to finalize.
            </div>
          )}
        </div>

        {step === 3 && (
          <InvoicePreview
            invoiceNumber={t('invoices.numberPending')}
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
          onClick={() => (step === 0 ? navigate({ to: '/invoices' }) : goToStep(step - 1))}
        >
          <ArrowLeft size={13} /> {step === 0 ? 'Cancel' : 'Back'}
        </button>
        <div style={{ display: 'flex', gap: 8 }}>
          {step < STEPS.length - 1 ? (
            <button type="button" style={primaryBtn} disabled={!canProceed()} onClick={advance}>
              Next <ChevronRight size={13} />
            </button>
          ) : (
            <>
              <button
                type="button"
                style={{
                  ...ghostBtn,
                  opacity: createInvoice.isPending || !clientEmailValid || !itemsValid ? 0.6 : 1,
                }}
                onClick={() => handleSave('Draft')}
                disabled={createInvoice.isPending || !clientEmailValid || !itemsValid}
              >
                <Save size={13} />{' '}
                {createInvoice.isPending ? t('common.saving') : t('invoices.saveDraft')}
              </button>
              <button
                type="button"
                style={{
                  ...primaryBtn,
                  opacity: createInvoice.isPending || !clientEmailValid || !itemsValid ? 0.6 : 1,
                }}
                onClick={() => handleSave('Sent')}
                disabled={createInvoice.isPending || !clientEmailValid || !itemsValid}
              >
                {createInvoice.isPending ? t('common.saving') : t('invoices.saveAndSend')}
              </button>
            </>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default NewInvoice;
