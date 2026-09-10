import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useParams, useRouterState, useNavigate } from '@tanstack/react-router';
import { EmptyState } from '@voxire/ui';
import { PackageX } from 'lucide-react';
import { orderClient } from '../../shared/lib/clients';
import { useCartStore } from '../../shared/stores/cartStore';

// The backend returns the internal pos_orders enum, not the FE's display steps.
// STEP_FOR_STATUS below maps each known status onto a display step.
const POLL_INTERVAL_MS = 30_000;

const STATUS_STEPS: { label: string }[] = [
  { label: 'Received' },
  { label: 'Preparing' },
  { label: 'Ready' },
  { label: 'Done' },
];

// Map each backend status onto one of the 4 display steps.
const STEP_FOR_STATUS: Record<string, number> = {
  pending: 0,
  confirmed: 0,
  preparing: 1,
  held: 1,
  ready: 2,
  served: 3,
  paid: 3,
};

const CANCELLED_STATUSES = new Set(['cancelled', 'refunded', 'voided']);

const stepForStatus = (s: string): number => {
  const norm = s.toLowerCase();
  return STEP_FOR_STATUS[norm] ?? -1;
};
const isCancelled = (s: string | null): boolean =>
  s != null && CANCELLED_STATUSES.has(s.toLowerCase());
const isTerminal = (s: string | null): boolean =>
  s != null && (isCancelled(s) || stepForStatus(s) === STATUS_STEPS.length - 1);

const OrderStatus = (): React.JSX.Element => {
  const { t } = useTranslation('orderstatus');
  const { id } = useParams({ strict: false });
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { clearCart } = useCartStore();

  const orderNumber = new URLSearchParams(searchStr).get('number') ?? '';
  const businessId =
    new URLSearchParams(searchStr).get('business_id') ??
    new URLSearchParams(searchStr).get('businessId') ??
    '';

  // Handed out once by createPublicOrder and carried here in the URL. The
  // status endpoint requires it, so a link that predates it (an old bookmark,
  // a shared URL) can no longer poll. That is the point of the change: the
  // sequential order id on its own used to be enough to read any order.
  const publicToken = new URLSearchParams(searchStr).get('public_token') ?? '';

  const [status, setStatus] = useState<string | null>(null);
  const [fetchError, setFetchError] = useState(false);
  const [notFound, setNotFound] = useState(false);

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchStatus = useCallback(async (): Promise<void> => {
    if (!id || !businessId) return;
    // Without the token the request would be rejected as malformed rather than
    // as not-found, so surface the same dead-end state instead of polling a
    // call that cannot succeed.
    if (!publicToken) {
      setNotFound(true);
      return;
    }
    try {
      const res = await orderClient.getPublicOrderStatus({
        orderId: id,
        businessId,
        publicToken,
      });
      const nextStatus = res.status ?? '';
      setStatus(nextStatus);
      setFetchError(false);
      setNotFound(false);

      // Stop polling once the order reaches a terminal state (done/cancelled).
      if (isTerminal(nextStatus) && intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    } catch (err: unknown) {
      // Generated onekit clients throw an ApiError with `statusCode`;
      // legacy axios-style errors carry `response.status`.
      const statusCode =
        (err as { statusCode?: number })?.statusCode ??
        (err as { response?: { status?: number } })?.response?.status;
      if (statusCode === 404) {
        setNotFound(true);
        setFetchError(false);
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      } else {
        setFetchError(true);
      }
    }
  }, [id, businessId, publicToken]);

  useEffect(() => {
    void fetchStatus();
    intervalRef.current = setInterval(() => void fetchStatus(), POLL_INTERVAL_MS);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [fetchStatus]);

  const navigate = useNavigate();

  const handleOrderAnother = (): void => {
    clearCart();
    const returnUrl = new URLSearchParams(searchStr).get('return_url');
    if (returnUrl) {
      window.location.href = returnUrl;
      return;
    }
    const slug = new URLSearchParams(searchStr).get('slug');
    if (slug) {
      void navigate({ to: '/$slug', params: { slug } });
      return;
    }
    void navigate({ to: '/' });
  };

  const cancelled = isCancelled(status);
  const currentIndex = status != null && !cancelled ? stepForStatus(status) : -1;

  return (
    <>
      <HelmetContainer title="Order status · rtylr" />
      <div
        style={{
          maxWidth: 480,
          margin: '0 auto',
          padding: '40px 20px',
          minHeight: '100vh',
          background: 'var(--menu-bg, #ffffff)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 32,
        }}
      >
        {/* Order number */}
        <div style={{ textAlign: 'center' }}>
          <p
            style={{
              fontSize: '14px',
              color: 'rgba(0,0,0,0.5)',
              margin: '0 0 6px',
              letterSpacing: '0.05em',
              textTransform: 'uppercase',
            }}
          >
            Order
          </p>
          <h1
            style={{
              fontSize: '40px',
              fontWeight: 900,
              margin: 0,
              color: 'var(--menu-accent, #03aed2)',
              letterSpacing: '-0.02em',
            }}
          >
            {orderNumber || `#${id ?? ''}`}
          </h1>
        </div>

        {/* Cancelled / refunded / voided terminal state */}
        {cancelled && !fetchError && (
          <p style={{ textAlign: 'center', fontSize: '18px', fontWeight: 700, color: '#dc2626' }}>
            {status === 'refunded' ? 'Order refunded' : 'Order cancelled'}
          </p>
        )}

        {/* Progress bar */}
        {status != null && !fetchError && !cancelled && (
          <div style={{ width: '100%' }}>
            {/* Step labels */}
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              {STATUS_STEPS.map((step, i) => (
                <span
                  key={step.label}
                  style={{
                    fontSize: '12px',
                    fontWeight: i <= currentIndex ? 700 : 400,
                    color: i <= currentIndex ? 'var(--menu-accent, #03aed2)' : 'rgba(0,0,0,0.35)',
                    flex: 1,
                    textAlign: 'center',
                  }}
                >
                  {step.label}
                </span>
              ))}
            </div>
            {/* Track */}
            <div
              style={{
                height: 8,
                background: 'rgba(0,0,0,0.08)',
                borderRadius: 999,
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  borderRadius: 999,
                  background: 'var(--menu-accent, #03aed2)',
                  width: `${((currentIndex + 1) / STATUS_STEPS.length) * 100}%`,
                  transition: 'width 0.5s ease',
                }}
              />
            </div>
            {/* Current status label */}
            <p
              style={{
                textAlign: 'center',
                marginTop: 16,
                fontSize: '18px',
                fontWeight: 700,
                color: currentIndex === STATUS_STEPS.length - 1 ? '#16a34a' : 'rgba(0,0,0,0.8)',
              }}
            >
              {currentIndex === 0 && 'Order received!'}
              {currentIndex === 1 && 'Being prepared...'}
              {currentIndex === 2 && 'Ready for pickup!'}
              {currentIndex === 3 && 'Order complete!'}
              {currentIndex === -1 && `Status: ${status}`}
            </p>
          </div>
        )}

        {/* Not Found state */}
        {notFound && (
          <EmptyState
            icon={PackageX}
            title={t('orderstatus.notFoundTitle')}
            description={t('orderstatus.notFoundDescription')}
            action={{ label: t('orderstatus.backToMenu'), onClick: handleOrderAnother }}
          />
        )}

        {/* Loading state */}
        {status == null && !fetchError && !notFound && (
          <p style={{ color: 'rgba(0,0,0,0.4)', fontSize: '15px' }}>Loading order status...</p>
        )}

        {/* Error state */}
        {fetchError && (
          <div style={{ textAlign: 'center' }}>
            <p style={{ color: '#dc2626', fontSize: '15px', marginBottom: 12 }}>
              Could not load order status.
            </p>
            <button
              type="button"
              onClick={() => void fetchStatus()}
              style={{
                background: 'transparent',
                border: '1px solid rgba(0,0,0,0.2)',
                borderRadius: 10,
                padding: '10px 20px',
                fontSize: '14px',
                cursor: 'pointer',
                minHeight: 44,
              }}
            >
              Retry
            </button>
          </div>
        )}

        {/* Poll notice */}
        {!isTerminal(status) && !fetchError && !notFound && (
          <p
            style={{
              fontSize: '12px',
              color: 'rgba(0,0,0,0.3)',
              textAlign: 'center',
              margin: 0,
            }}
          >
            Status refreshes automatically every 30 seconds.
          </p>
        )}

        {/* Order another */}
        {!notFound && (
          <button
            type="button"
            onClick={handleOrderAnother}
            style={{
              marginTop: 'auto',
              background: 'transparent',
              border: '2px solid var(--menu-accent, #03aed2)',
              color: 'var(--menu-accent, #03aed2)',
              borderRadius: 14,
              padding: '14px 32px',
              fontSize: '16px',
              fontWeight: 700,
              cursor: 'pointer',
              minHeight: 48,
              width: '100%',
            }}
          >
            Order Another
          </button>
        )}
      </div>
    </>
  );
};

export default OrderStatus;
