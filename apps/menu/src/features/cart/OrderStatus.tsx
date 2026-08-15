import React, { useCallback, useEffect, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useParams, useRouterState } from '@tanstack/react-router';
import api from '../../shared/lib/api';
import { useCartStore } from '../../shared/stores/cartStore';

// The backend returns the internal pos_orders enum, not the FE's display steps.
type BackendOrderStatus =
  | 'pending'
  | 'confirmed'
  | 'preparing'
  | 'ready'
  | 'served'
  | 'paid'
  | 'held'
  | 'cancelled'
  | 'refunded'
  | 'voided';

interface OrderStatusResponse {
  order_number: string;
  status: BackendOrderStatus | string;
}

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
  const { id } = useParams({ strict: false });
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { clearCart } = useCartStore();
  const businessId = new URLSearchParams(searchStr).get('business_id');

  const orderNumber = new URLSearchParams(searchStr).get('number') ?? '';
  const businessId =
    new URLSearchParams(searchStr).get('business_id') ??
    new URLSearchParams(searchStr).get('businessId') ??
    '';

  const [status, setStatus] = useState<string | null>(null);
  const [fetchError, setFetchError] = useState(false);
  const [notFound, setNotFound] = useState(false);

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchStatus = useCallback(async (): Promise<void> => {
    if (!id) return;
    try {
      const params = businessId ? { business_id: businessId } : undefined;
      const res = await api.get<OrderStatusResponse>(`/pos/orders/v1/orders/${id}/public-status`, {
        params,
      });
      setStatus(res.data.status);
      setFetchError(false);
      setNotFound(false);

      // Stop polling once the order reaches a terminal state (done/cancelled).
      if (isTerminal(res.data.status) && intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 404) {
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
  }, [id, businessId]);

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
      void navigate({ to: `/${slug}` });
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
          <div style={{ textAlign: 'center' }}>
            <p style={{ color: '#dc2626', fontSize: '15px', fontWeight: 600, marginBottom: 8 }}>
              Order not found
            </p>
            <p style={{ color: 'rgba(0,0,0,0.5)', fontSize: '13px', marginBottom: 16 }}>
              We could not find an order matching this ID. Please check your order link.
            </p>
          </div>
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
        {!isTerminal(status) && !fetchError && (
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
        <button
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
      </div>
    </>
  );
};

export default OrderStatus;
