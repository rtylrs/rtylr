import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import type { CreatePublicOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import { EmptyState } from '@voxire/ui';
import { ShoppingCart } from 'lucide-react';
import { orderClient } from '../../shared/lib/clients';
import { useCartStore } from '../../shared/stores/cartStore';

const Checkout = (): React.JSX.Element => {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { items, businessId, total, setItemNote, clearCart } = useCartStore();

  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [orderNotes, setOrderNotes] = useState('');
  const [expandedItemId, setExpandedItemId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const currency = 'USD';

  const formatPrice = (price: number): string =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(price);

  const onlineOrderingEnabled = import.meta.env.VITE_ONLINE_ORDERING_ENABLED !== 'false';

  const { t } = useTranslation('checkout');

  if (!onlineOrderingEnabled) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center">
        <h2 className="text-xl font-semibold">{t('checkout.orderingDisabledTitle')}</h2>
        <p className="mt-2 text-sm text-gray-600">{t('checkout.orderingDisabledDescription')}</p>
        <button
          type="button"
          onClick={() => navigate({ to: '/' })}
          className="mt-6 rounded-lg bg-black px-4 py-2 text-sm text-white"
        >
          {t('checkout.backToMenu')}
        </button>
      </div>
    );
  }

  const validate = (): boolean => {
    if (!name.trim()) {
      setErrorMessage(t('checkout.nameRequired'));
      return false;
    }
    if (!phone.trim()) {
      setErrorMessage(t('checkout.phoneRequired'));
      return false;
    }
    if (email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setErrorMessage(t('checkout.emailInvalid'));
      return false;
    }
    if (items.length === 0) {
      setErrorMessage(t('checkout.emptyCart'));
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    setErrorMessage(null);

    if (!validate()) return;

    setSubmitting(true);

    const request: CreatePublicOrderRequest = {
      ownerId: String(businessId),
      customerName: name.trim(),
      customerEmail: email.trim(),
      customerPhone: phone.trim(),
      notes: orderNotes.trim(),
      items: items.map((item) => ({
        menuItemId: String(item.menuItemId),
        name: item.name,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        notes: item.notes,
      })),
    };

    try {
      const response = await orderClient.createPublicOrder(request);
      const { orderId, orderNumber, publicToken } = response;
      if (!orderId || !orderNumber) throw new Error('Incomplete order response');

      clearCart();
      const businessQuery = businessId
        ? `&business_id=${encodeURIComponent(String(businessId))}`
        : '';
      const slug = new URLSearchParams(searchStr).get('slug');
      const slugQuery = slug ? `&slug=${encodeURIComponent(slug)}` : '';
      // The status endpoint takes an opaque per-order token now, so the
      // sequential order id alone can no longer be used to read anyone's
      // order. This is the only place it is handed out, so it has to travel
      // to the status page or that page cannot poll at all.
      const tokenQuery = publicToken ? `&public_token=${encodeURIComponent(publicToken)}` : '';
      void navigate({
        to: `/order-status/${orderId}?number=${encodeURIComponent(orderNumber)}${businessQuery}${slugQuery}${tokenQuery}`,
      });
    } catch {
      setErrorMessage(t('checkout.orderFailed'));
    } finally {
      setSubmitting(false);
    }
  };

  if (items.length === 0) {
    return (
      <>
        <HelmetContainer title={t('checkout.pageTitle')} />
        <div
          style={{
            minHeight: '100vh',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 16,
            padding: '0 20px',
          }}
        >
          <EmptyState
            icon={ShoppingCart}
            title={t('checkout.emptyCart')}
            action={{ label: t('checkout.backToMenu'), onClick: () => window.history.back() }}
          />
        </div>
      </>
    );
  }

  return (
    <>
      <HelmetContainer title={t('checkout.pageTitle')} />
      <div
        style={{
          maxWidth: 520,
          margin: '0 auto',
          padding: '24px 20px 40px',
          background: 'var(--menu-bg, #ffffff)',
          minHeight: '100vh',
        }}
      >
        {/* Back */}
        <button
          type="button"
          onClick={() => window.history.back()}
          aria-label={t('checkout.goBack')}
          style={{
            background: 'transparent',
            border: 'none',
            cursor: 'pointer',
            fontSize: '14px',
            color: 'rgba(0,0,0,0.5)',
            padding: '0 0 16px 0',
            display: 'flex',
            alignItems: 'center',
            gap: 4,
          }}
        >
          &larr; {t('checkout.back')}
        </button>

        <h1 style={{ fontSize: '22px', fontWeight: 800, margin: '0 0 24px' }}>
          {t('checkout.title')}
        </h1>

        <form onSubmit={(e) => void handleSubmit(e)} noValidate>
          {/* Cart summary */}
          <section
            style={{
              border: '1px solid rgba(0,0,0,0.1)',
              borderRadius: 16,
              padding: '16px',
              marginBottom: 24,
            }}
          >
            <h2 style={{ fontSize: '16px', fontWeight: 700, margin: '0 0 12px' }}>
              {t('checkout.orderSummary')}
            </h2>
            <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
              {items.map((item) => (
                <li
                  key={item.menuItemId}
                  style={{
                    padding: '8px 0',
                    borderBottom: '1px solid rgba(0,0,0,0.05)',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                    }}
                  >
                    <span style={{ fontSize: '15px' }}>
                      {item.name}
                      <span
                        style={{
                          fontSize: '13px',
                          color: 'rgba(0,0,0,0.45)',
                          marginLeft: 6,
                        }}
                      >
                        x{item.quantity}
                      </span>
                    </span>
                    <span style={{ fontSize: '15px', fontWeight: 600 }}>
                      {formatPrice(item.unitPrice * item.quantity)}
                    </span>
                  </div>

                  {/* Expandable item note */}
                  <button
                    type="button"
                    onClick={() =>
                      setExpandedItemId((prev) =>
                        prev === item.menuItemId ? null : item.menuItemId
                      )
                    }
                    style={{
                      background: 'transparent',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '13px',
                      color: 'var(--menu-accent, #03aed2)',
                      padding: '4px 0 0',
                    }}
                  >
                    {item.notes
                      ? t('checkout.editNote')
                      : expandedItemId === item.menuItemId
                        ? t('checkout.cancelNote')
                        : t('checkout.addNote')}
                  </button>

                  {expandedItemId === item.menuItemId && (
                    <textarea
                      value={item.notes}
                      onChange={(e) => setItemNote(item.menuItemId, e.target.value)}
                      placeholder={t('checkout.itemNotePlaceholder')}
                      rows={2}
                      style={{
                        width: '100%',
                        marginTop: 6,
                        padding: '8px 10px',
                        fontSize: '14px',
                        border: '1px solid rgba(0,0,0,0.15)',
                        borderRadius: 8,
                        resize: 'none',
                        boxSizing: 'border-box',
                      }}
                    />
                  )}
                </li>
              ))}
            </ul>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                paddingTop: 12,
                fontWeight: 700,
                fontSize: '16px',
              }}
            >
              <span>{t('checkout.total')}</span>
              <span>{formatPrice(total())}</span>
            </div>
          </section>

          {/* Customer info */}
          <section style={{ marginBottom: 24 }}>
            <h2 style={{ fontSize: '16px', fontWeight: 700, margin: '0 0 12px' }}>
              {t('checkout.yourInfo')}
            </h2>

            <label
              htmlFor="checkout-name"
              style={{ display: 'block', fontSize: '14px', fontWeight: 600, marginBottom: 4 }}
            >
              {t('checkout.name')} *
            </label>
            <input
              id="checkout-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('checkout.namePlaceholder')}
              required
              style={{
                width: '100%',
                padding: '10px 12px',
                fontSize: '16px',
                border: '1px solid rgba(0,0,0,0.15)',
                borderRadius: 10,
                marginBottom: 14,
                boxSizing: 'border-box',
              }}
            />

            <label
              htmlFor="checkout-phone"
              style={{ display: 'block', fontSize: '14px', fontWeight: 600, marginBottom: 4 }}
            >
              {t('checkout.phone')} *
            </label>
            <input
              id="checkout-phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder={t('checkout.phonePlaceholder')}
              required
              style={{
                width: '100%',
                padding: '10px 12px',
                fontSize: '16px',
                border: '1px solid rgba(0,0,0,0.15)',
                borderRadius: 10,
                marginBottom: 14,
                boxSizing: 'border-box',
              }}
            />

            <label
              htmlFor="checkout-email"
              style={{ display: 'block', fontSize: '14px', fontWeight: 600, marginBottom: 4 }}
            >
              {t('checkout.email')}
            </label>
            <input
              id="checkout-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t('checkout.emailPlaceholder')}
              style={{
                width: '100%',
                padding: '10px 12px',
                fontSize: '16px',
                border: '1px solid rgba(0,0,0,0.15)',
                borderRadius: 10,
                marginBottom: 14,
                boxSizing: 'border-box',
              }}
            />
          </section>

          {/* Order notes */}
          <section style={{ marginBottom: 28 }}>
            <label
              htmlFor="checkout-notes"
              style={{ display: 'block', fontSize: '14px', fontWeight: 600, marginBottom: 4 }}
            >
              {t('checkout.notes')}
            </label>
            <textarea
              id="checkout-notes"
              value={orderNotes}
              onChange={(e) => setOrderNotes(e.target.value)}
              placeholder={t('checkout.notesPlaceholder')}
              rows={3}
              style={{
                width: '100%',
                padding: '10px 12px',
                fontSize: '16px',
                border: '1px solid rgba(0,0,0,0.15)',
                borderRadius: 10,
                resize: 'none',
                boxSizing: 'border-box',
              }}
            />
          </section>

          {/* Error */}
          {errorMessage && (
            <p
              role="alert"
              style={{
                color: '#dc2626',
                fontSize: '14px',
                marginBottom: 16,
                padding: '10px 14px',
                background: '#fef2f2',
                borderRadius: 8,
              }}
            >
              {errorMessage}
            </p>
          )}

          {/* Submit */}
          <button
            type="submit"
            disabled={submitting}
            style={{
              width: '100%',
              background: submitting ? 'rgba(0,0,0,0.2)' : 'var(--menu-accent, #03aed2)',
              color: 'var(--menu-accent-text, #ffffff)',
              border: 'none',
              borderRadius: 14,
              padding: '16px',
              fontSize: '17px',
              fontWeight: 700,
              cursor: submitting ? 'not-allowed' : 'pointer',
              minHeight: 52,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
            }}
          >
            {submitting
              ? t('checkout.submitting')
              : `${t('checkout.placeOrder')} · ${formatPrice(total())}`}
          </button>
        </form>
      </div>
    </>
  );
};

export default Checkout;
