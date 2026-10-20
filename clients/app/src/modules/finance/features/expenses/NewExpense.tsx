import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Save, ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import ReceiptUpload from '../../shared/components/ReceiptUpload';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { minorToMajor } from '../../shared/lib/money';
import { useSettings } from '../../shared/lib/settings';
import { useExpenseCategories } from '../../shared/hooks/useExpenseCategories';
import { useCreateExpense } from './useExpenses';
import { cardStyle, ghostBtn, primaryBtn, inputStyle, labelStyle } from '../../shared/lib/tokens';
import { todayIso } from '../../shared/lib/date';
import { PAYMENT_METHODS, lastPaymentMethod, rememberPaymentMethod } from './expenseDisplay';

const NewExpense: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const navigate = useNavigate();
  const currency = useSettings((state) => state.currency);
  const { data: categories = [] } = useExpenseCategories();
  const [receipt, setReceipt] = useState<string | null>(null);
  const [description, setDescription] = useState('');
  const [amount, setAmount] = useState(0);
  const [categoryId, setCategoryId] = useState<number | undefined>(undefined);
  const [date, setDate] = useState(todayIso());
  const [notes, setNotes] = useState('');
  const [vendor, setVendor] = useState('');
  const [paymentMethod, setPaymentMethod] = useState(lastPaymentMethod);
  const createExpense = useCreateExpense();

  const handleSave = (): void => {
    if (!description.trim() || categoryId === undefined) {
      toast.error(t('financeExpenses.descriptionAndCategoryRequired'));
      return;
    }
    const amountInMajorUnits = minorToMajor(amount, currency);
    if (!Number.isFinite(amountInMajorUnits) || amountInMajorUnits <= 0) {
      toast.error(t('financeExpenses.amountPositive'));
      return;
    }
    createExpense.mutate(
      {
        description,
        amount: amountInMajorUnits,
        categoryId,
        currency,
        expenseDate: date,
        receiptUrl: receipt ?? undefined,
        notes,
        vendor: vendor.trim(),
        paymentMethod,
      },
      {
        onSuccess: () => {
          rememberPaymentMethod(paymentMethod);
          toast.success(t('financeExpenses.submitted'));
          navigate({ to: '/finance/expenses' });
        },
      }
    );
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.submitExpense'))} />
      <PageHeader
        title={t('financeExpenses.submitTitle')}
        breadcrumbs={[
          { label: t('financeNav.expenses'), to: '/finance/expenses' },
          { label: t('financeCommon.new') },
        ]}
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <div
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 14 }}
        >
          <div>
            <label style={labelStyle}>{t('financeExpenses.description')}</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              style={{ ...inputStyle, marginTop: 6 }}
              placeholder={t('financeExpenses.descriptionPlaceholder')}
            />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div>
              <label style={labelStyle}>{t('financeExpenses.amount')}</label>
              <div style={{ marginTop: 6 }}>
                <CurrencyInput value={amount} onChange={setAmount} currency={currency} />
              </div>
            </div>
            <div>
              <label style={labelStyle}>{t('financeExpenses.date')}</label>
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                style={{ ...inputStyle, marginTop: 6 }}
              />
            </div>
          </div>
          <div>
            <label style={labelStyle} htmlFor="expense-category">
              {t('financeExpenses.category')}
            </label>
            <select
              id="expense-category"
              value={categoryId ?? ''}
              onChange={(e) =>
                setCategoryId(e.target.value === '' ? undefined : Number(e.target.value))
              }
              style={{ ...inputStyle, marginTop: 6 }}
            >
              <option value="">{t('financeExpenses.selectCategory')}</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div>
              <label style={labelStyle} htmlFor="expense-vendor">
                {t('financeExpenses.vendor')}
              </label>
              <input
                id="expense-vendor"
                value={vendor}
                onChange={(e) => setVendor(e.target.value)}
                style={{ ...inputStyle, marginTop: 6 }}
                placeholder={t('financeExpenses.vendorPlaceholder')}
              />
            </div>
            <div>
              <label style={labelStyle} htmlFor="expense-payment-method">
                {t('financeExpenses.paymentMethod')}
              </label>
              <select
                id="expense-payment-method"
                value={paymentMethod}
                onChange={(e) => setPaymentMethod(e.target.value)}
                style={{ ...inputStyle, marginTop: 6 }}
              >
                {PAYMENT_METHODS.map((method) => (
                  <option key={method} value={method}>
                    {t(`financeExpenses.paymentMethods.${method}`)}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div>
            <label style={labelStyle}>{t('financeExpenses.notes')}</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              style={{ ...inputStyle, marginTop: 6, resize: 'vertical', fontFamily: 'inherit' }}
            />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div
            style={{
              fontSize: 11,
              color: 'var(--hub-text-dim)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              marginBottom: 12,
            }}
          >
            {t('financeExpenses.receipt')}
          </div>
          <ReceiptUpload value={receipt} onChange={setReceipt} />
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <button
          type="button"
          style={ghostBtn}
          onClick={() => navigate({ to: '/finance/expenses' })}
          disabled={createExpense.isPending}
        >
          <ArrowLeft size={13} /> {t('financeCommon.cancel')}
        </button>
        <button
          type="button"
          style={{ ...primaryBtn, opacity: createExpense.isPending ? 0.6 : 1 }}
          onClick={handleSave}
          disabled={createExpense.isPending}
        >
          <Save size={13} />{' '}
          {createExpense.isPending ? t('financeExpenses.submitting') : t('financeExpenses.submit')}
        </button>
      </div>
    </motion.div>
  );
};

export default NewExpense;
