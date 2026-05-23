import React from 'react';
import { motion, useMotionValue, useTransform, animate } from 'motion/react';
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
  /** Index in the parent grid - drives stagger-in delay. */
  index?: number;
}

const CountUp: React.FC<{ to: number; format?: (n: number) => string }> = ({ to, format }) => {
  const mv = useMotionValue(0);
  const rounded = useTransform(mv, (v) => (format ? format(v) : Math.round(v).toLocaleString()));
  React.useEffect(() => {
    const controls = animate(mv, to, { duration: 0.9, ease: [0.22, 1, 0.36, 1] });
    return controls.stop;
  }, [mv, to]);
  return <motion.span>{rounded}</motion.span>;
};

export const StatCard: React.FC<StatCardProps> = ({
  label,
  value,
  trendPct,
  sub,
  icon: Icon,
  accent = '#4583C4',
  className,
  onClick,
  index = 0,
}) => {
  const hasTrend = typeof trendPct === 'number';
  const trendPositive = hasTrend && (trendPct ?? 0) > 0;
  const trendNegative = hasTrend && (trendPct ?? 0) < 0;

  const numericValue = typeof value === 'number' ? value : null;
  const displayValue = numericValue !== null ? <CountUp to={numericValue} /> : value;

  return (
    <motion.div
      className={cn('rounded-xl p-4', onClick && 'cursor-pointer', className)}
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      onClick={onClick}
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.32, ease: [0.22, 1, 0.36, 1], delay: index * 0.05 }}
      whileHover={{ y: -3, boxShadow: '0 12px 28px -12px rgba(0,0,0,0.18)' }}
    >
      <div className="flex items-center justify-between mb-3">
        <p
          className="text-[11px] font-medium uppercase tracking-wider"
          style={{ color: 'var(--dash-text-dim)' }}
        >
          {label}
        </p>
        {Icon && (
          <div
            className="h-7 w-7 rounded-lg flex items-center justify-center"
            style={{ background: `${accent}18` }}
          >
            <Icon size={14} style={{ color: accent }} />
          </div>
        )}
      </div>
      <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
        {displayValue}
      </p>
      <div className="flex items-center gap-1.5 mt-1">
        {hasTrend && (
          <span
            className="inline-flex items-center gap-0.5 text-[11px] font-semibold"
            style={{
              color: trendPositive ? '#10b981' : trendNegative ? '#ef4444' : 'var(--dash-text-dim)',
            }}
          >
            {trendPositive ? (
              <ArrowUpRight size={11} />
            ) : trendNegative ? (
              <ArrowDownRight size={11} />
            ) : (
              <Minus size={11} />
            )}
            {Math.abs(trendPct ?? 0).toFixed(1)}%
          </span>
        )}
        {sub && (
          <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
            {sub}
          </span>
        )}
      </div>
    </motion.div>
  );
};

export default StatCard;
