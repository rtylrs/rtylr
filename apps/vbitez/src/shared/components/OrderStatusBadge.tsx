import React from 'react';
import type { OrderStatus } from '../types/pos';

const STYLES: Record<OrderStatus, string> = {
  open: 'bg-blue-500/15 text-blue-300 border-blue-500/30',
  sent: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
  ready: 'bg-teal-500/15 text-teal-300 border-teal-500/30',
  paid: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  voided: 'bg-red-500/15 text-red-300 border-red-500/30',
  refunded: 'bg-purple-500/15 text-purple-300 border-purple-500/30',
};

interface OrderStatusBadgeProps {
  status: OrderStatus;
}

const OrderStatusBadge: React.FC<OrderStatusBadgeProps> = ({ status }) => (
  <span
    className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wider border ${STYLES[status]}`}
  >
    {status}
  </span>
);

export default OrderStatusBadge;
