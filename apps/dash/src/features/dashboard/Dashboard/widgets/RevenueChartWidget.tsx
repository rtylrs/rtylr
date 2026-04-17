import React from 'react';
import { BarChart3 } from 'lucide-react';

// Mock data for the chart since we don't have historical revenue yet
const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const values = [0, 0, 0, 0, 0, 0, 0]; // Will be real once POS data flows
const max = Math.max(...values, 1);

const RevenueChartWidget: React.FC = () => {
  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <BarChart3 size={14} style={{ color: '#4583C4' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Revenue (7 days)
        </p>
      </div>
      <div className="flex-1 flex items-end gap-2 px-4 pb-4 pt-2">
        {days.map((day, i) => (
          <div key={day} className="flex-1 flex flex-col items-center gap-1">
            <div className="w-full relative" style={{ height: '100%', minHeight: 60 }}>
              <div
                className="absolute bottom-0 w-full rounded-t-md transition-all"
                style={{
                  height: values[i] > 0 ? `${(values[i] / max) * 100}%` : '4px',
                  background: values[i] > 0 ? '#4583C4' : 'var(--dash-border)',
                  minHeight: 4,
                }}
              />
            </div>
            <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
              {day}
            </span>
          </div>
        ))}
      </div>
      {values.every((v) => v === 0) && (
        <p className="text-center text-[11px] pb-3" style={{ color: 'var(--dash-text-dim)' }}>
          Revenue data will appear once orders start flowing
        </p>
      )}
    </div>
  );
};

export default RevenueChartWidget;
