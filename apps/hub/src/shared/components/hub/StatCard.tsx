import React from 'react';
import { ArrowDownRight, ArrowUpRight, Minus } from 'lucide-react';
import { cn } from '../../lib/utils';

export interface StatCardProps {
  label: string;
  value: React.ReactNode;
  trendPct?: number;
  sub?: string;
  icon?: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  accent?: string;
  className?: string;
  onClick?: () => void;
}

export const StatCard: React.FC<StatCardProps> = ({
  label,
  value,
  trendPct,
  sub,
  icon: Icon,
  accent = '#4583C4',
  className,
  onClick,
}) => {
  const hasTrend = typeof trendPct === 'number';
  const trendPositive = hasTrend && (trendPct ?? 0) > 0;
  const trendNegative = hasTrend && (trendPct ?? 0) < 0;

  return (
    <div
      className={cn(
        'rounded-xl p-4 transition-colors',
        onClick && 'cursor-pointer hover:bg-[var(--hub-hover)]',
        className,
      )}
      style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      onClick={onClick}
    >
      <div className="flex items-center justify-between mb-3">
        <p className="text-[11px] font-medium uppercase tracking-wider" style={{ color: 'var(--hub-text-dim)' }}>
          {label}
        </p>
        {Icon && (
          <div className="h-7 w-7 rounded-lg flex items-center justify-center" style={{ background: `${accent}18` }}>
            <Icon size={14} style={{ color: accent }} />
          </div>
        )}
      </div>
      <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
        {value}
      </p>
      <div className="flex items-center gap-1.5 mt-1">
        {hasTrend && (
          <span
            className="inline-flex items-center gap-0.5 text-[11px] font-semibold"
            style={{
              color: trendPositive ? '#10b981' : trendNegative ? '#ef4444' : 'var(--hub-text-dim)',
            }}
          >
            {trendPositive ? <ArrowUpRight size={11} /> : trendNegative ? <ArrowDownRight size={11} /> : <Minus size={11} />}
            {Math.abs(trendPct ?? 0).toFixed(1)}%
          </span>
        )}
        {sub && (
          <span className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
            {sub}
          </span>
        )}
      </div>
    </div>
  );
};

export default StatCard;
