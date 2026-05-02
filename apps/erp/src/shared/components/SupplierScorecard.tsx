import React from 'react';
import { Supplier } from '../services/erpApi';
import { Truck, Target, Clock } from 'lucide-react';

interface SupplierScorecardProps {
  supplier: Supplier;
}

const Metric: React.FC<{
  label: string;
  value: number | null;
  suffix?: string;
  color: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}> = ({ label, value, suffix = '%', color, icon: Icon }) => (
  <div className="flex-1">
    <div className="flex items-center gap-1.5 mb-1">
      <Icon size={11} style={{ color }} />
      <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
        {label}
      </p>
    </div>
    <p className="text-lg font-bold" style={{ color }}>
      {value !== null ? (
        <>
          {value}
          <span className="text-xs ml-0.5">{suffix}</span>
        </>
      ) : (
        <span className="text-sm" style={{ color: 'var(--hub-text-dim)' }}>
          —
        </span>
      )}
    </p>
    <div className="h-1 rounded-full mt-1" style={{ background: 'var(--hub-hover)' }}>
      <div
        className="h-full rounded-full"
        style={{ width: value !== null ? `${Math.min(100, value)}%` : '0%', background: color }}
      />
    </div>
  </div>
);

const SupplierScorecard: React.FC<SupplierScorecardProps> = ({ supplier }) => {
  const onTime = supplier.onTimeDeliveryPct ?? null;
  const accuracy = supplier.fulfillmentAccuracy ?? null;
  const leadTime = supplier.leadTimeDays ?? null;

  return (
    <div
      className="rounded-xl p-4"
      style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
    >
      <div className="flex items-center justify-between mb-3">
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          {supplier.name}
        </p>
        <span
          className="text-[10px] font-medium px-2 py-0.5 rounded-full"
          style={{
            background: supplier.status === 'Active' ? 'rgba(16,185,129,0.12)' : 'var(--hub-hover)',
            color: supplier.status === 'Active' ? '#10b981' : 'var(--hub-text-muted)',
          }}
        >
          {supplier.status}
        </span>
      </div>
      <div className="flex gap-4">
        <Metric
          label="On-Time"
          value={onTime}
          color={onTime === null ? 'var(--hub-text-dim)' : onTime >= 90 ? '#10b981' : '#f59e0b'}
          icon={Truck}
        />
        <Metric
          label="Accuracy"
          value={accuracy}
          color={accuracy >= 95 ? '#10b981' : '#f59e0b'}
          icon={Target}
        />
        <Metric label="Lead Days" value={leadTime} suffix="d" color="#60a5fa" icon={Clock} />
      </div>
    </div>
  );
};

export default SupplierScorecard;
