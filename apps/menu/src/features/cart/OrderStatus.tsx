import React, { useCallback, useEffect, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useParams, useRouterState } from '@tanstack/react-router';
import api from '../../shared/lib/api';
import { useCartStore } from '../../shared/stores/cartStore';

type OrderStatusValue = 'received' | 'preparing' | 'ready' | 'done';

interface OrderStatusResponse {
  order_number: string;
  status: OrderStatusValue;
}

const POLL_INTERVAL_MS = 30_000;

const STATUS_STEPS: { key: OrderStatusValue; label: string }[] = [
  { key: 'received', label: 'Received' },
  { key: 'preparing', label: 'Preparing' },
  { key: 'ready', label: 'Ready' },
  { key: 'done', label: 'Done' },
];

const statusIndex = (s: OrderStatusValue): number =>
  STATUS_STEPS.findIndex((step) => step.key === s);

const OrderStatus = (): React.JSX.Element => {
  const { id } = useParams({ strict: false });
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const navigate = useNavigate();
  const { clearCart } = useCartStore();

  const orderNumber = new URLSearchParams(searchStr).get('number') ?? '';

  const [status, setStatus] = useState<OrderStatusValue | null>(null);
  const [fetchError, setFetchError] = useState(false);

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchStatus = useCallback(async (): Promise<void> => {
    if (!id) return;
    try {
      const res = await api.get<OrderStatusResponse>(`/pos/public/orders/${id}/status`);
      setStatus(res.data.status);
      setFetchError(false);

      // Stop polling when order is done
      if (res.data.status === 'done' && intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    } catch {
      setFetchError(true);
    }
  }, [id]);

  useEffect(() => {
    void fetchStatus();
    intervalRef.current = setInterval(() => void fetchStatus(), POLL_INTERVAL_MS);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [fetchStatus]);

  const handleOrderAnother = (): void => {
    clearCart();
    window.history.back();
  };

  const currentIndex = status != null ? statusIndex(status) : -1;

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

        {/* Progress bar */}
        {status != null && !fetchError && (
          <div style={{ width: '100%' }}>
            {/* Step labels */}
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              {STATUS_STEPS.map((step, i) => (
                <span
                  key={step.key}
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
                color: status === 'done' ? '#16a34a' : 'rgba(0,0,0,0.8)',
              }}
            >
              {status === 'received' && 'Order received!'}
              {status === 'preparing' && 'Being prepared...'}
              {status === 'ready' && 'Ready for pickup!'}
              {status === 'done' && 'Order complete!'}
            </p>
          </div>
        )}

        {/* Loading state */}
        {status == null && !fetchError && (
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
        {status !== 'done' && (
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
