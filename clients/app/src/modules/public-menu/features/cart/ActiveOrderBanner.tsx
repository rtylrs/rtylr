import React, { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { ArrowRight, Clock, X } from 'lucide-react';
import { orderClient } from '../../shared/lib/clients';
import {
  buildOrderStatusUrl,
  dismissActiveOrder,
  updateActiveOrderStatus,
  useActiveOrders,
} from '../../shared/stores/activeOrderStore';

const STEP_KEY_BY_STATUS: Record<string, 'received' | 'preparing' | 'ready'> = {
  pending: 'received',
  confirmed: 'received',
  processing: 'preparing',
  preparing: 'preparing',
  held: 'preparing',
  ready: 'ready',
  shipped: 'ready',
};

interface ActiveOrderBannerProps {
  businessId?: string | number | null;
}

const ActiveOrderBanner: React.FC<ActiveOrderBannerProps> = ({ businessId }) => {
  const { t, i18n } = useTranslation('orderstatus');
  const isRtl = i18n.dir() === 'rtl';
  const navigate = useNavigate();
  const orders = useActiveOrders(businessId);
  const orderIdsKey = orders.map((o) => o.orderId).join(',');

  useEffect(() => {
    if (orders.length === 0) return;
    let cancelled = false;

    for (const order of orders) {
      void orderClient
        .getPublicOrderStatus({
          orderId: order.orderId,
          businessId: order.businessId,
          publicToken: order.publicToken,
        })
        .then((res) => {
          if (cancelled || !res?.status) return;
          updateActiveOrderStatus(order.orderId, res.status);
        })
        .catch((err: unknown) => {
          if (cancelled) return;
          const statusCode =
            (err as { statusCode?: number })?.statusCode ??
            (err as { response?: { status?: number } })?.response?.status;
          if (statusCode === 404) {
            dismissActiveOrder(order.orderId);
          }
        });
    }

    return () => {
      cancelled = true;
    };
    // Re-sync once whenever the tracked order set changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orderIdsKey]);

  if (orders.length === 0) return null;

  const visibleOrders = orders.slice(0, 2);

  return (
    <div
      data-testid="active-order-banner"
      role="region"
      aria-label={t('orderstatus.trackOrder')}
      style={{
        position: 'fixed',
        bottom: 72,
        insetInline: 14,
        maxWidth: 520,
        margin: '0 auto',
        zIndex: 70,
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        pointerEvents: 'none',
      }}
    >
      {visibleOrders.map((order) => {
        const stepKey = STEP_KEY_BY_STATUS[order.status.toLowerCase()] ?? 'received';
        const statusLabel =
          stepKey === 'ready'
            ? t('orderstatus.steps.ready')
            : stepKey === 'preparing'
              ? t('orderstatus.steps.preparing')
              : t('orderstatus.steps.received');
        const isReady = stepKey === 'ready';

        return (
          <div
            key={order.orderId}
            style={{
              pointerEvents: 'auto',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '8px 10px 8px 14px',
              borderRadius: 14,
              background: isReady ? '#15803d' : '#111827',
              color: '#ffffff',
              boxShadow: '0 8px 24px rgba(0,0,0,0.22)',
              border: '1px solid rgba(255,255,255,0.12)',
            }}
          >
            <button
              type="button"
              onClick={() => void navigate({ to: buildOrderStatusUrl(order) })}
              style={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 10,
                background: 'transparent',
                border: 'none',
                color: 'inherit',
                padding: 0,
                cursor: 'pointer',
                textAlign: 'start',
                fontFamily: 'inherit',
              }}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: 8, minWidth: 0 }}>
                <Clock size={16} aria-hidden="true" style={{ flexShrink: 0, opacity: 0.9 }} />
                <span
                  style={{
                    fontSize: '13px',
                    fontWeight: 700,
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {t('orderstatus.activeOrderTitle', { number: order.orderNumber })} · {statusLabel}
                </span>
              </span>
              <span
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                  fontSize: '12px',
                  fontWeight: 600,
                  opacity: 0.92,
                  flexShrink: 0,
                }}
              >
                {t('orderstatus.trackOrder')}
                <ArrowRight
                  size={14}
                  aria-hidden="true"
                  style={isRtl ? { transform: 'scaleX(-1)' } : undefined}
                />
              </span>
            </button>
            <button
              type="button"
              onClick={() => dismissActiveOrder(order.orderId)}
              aria-label={t('orderstatus.dismissActiveOrder')}
              style={{
                width: 26,
                height: 26,
                borderRadius: '50%',
                border: 'none',
                background: 'rgba(255,255,255,0.14)',
                color: '#ffffff',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                flexShrink: 0,
              }}
            >
              <X size={14} aria-hidden="true" />
            </button>
          </div>
        );
      })}
    </div>
  );
};

export default ActiveOrderBanner;
