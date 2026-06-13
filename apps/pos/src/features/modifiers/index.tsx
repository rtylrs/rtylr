import React from 'react';
import { AlertTriangle, SlidersHorizontal } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useModifiers } from './useModifiers';
import { formatCurrency } from '../../shared/utils/format';

const ModifiersScreen: React.FC = () => {
  const { groups, loading, error } = useModifiers();

  return (
    <ScreenLayout title="Modifiers" subtitle="Ingredient modifiers with price deltas">
      <div className="p-5 space-y-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : groups.length === 0 ? (
          <EmptyState
            icon={SlidersHorizontal}
            title="No modifier groups"
            message="Create groups in your back office to see them here."
          />
        ) : (
          groups.map((g) => (
            <section
              key={g.id}
              className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-3"
            >
              <div className="flex items-baseline justify-between">
                <h2 className="text-base font-bold text-pos-text">{g.name}</h2>
                <span className="text-xs text-pos-muted">
                  min {g.min} · max {g.max}
                </span>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
                {g.modifiers.map((m) => (
                  <div
                    key={m.id}
                    className="flex items-center justify-between px-3 py-2 rounded-lg bg-white/6 border border-pos-border"
                  >
                    <span className="text-sm text-pos-text">{m.name}</span>
                    <span className="text-xs text-pos-muted tabular-nums">
                      {m.priceDelta >= 0 ? '+' : ''}
                      {formatCurrency(m.priceDelta)}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ))
        )}
      </div>
    </ScreenLayout>
  );
};

export default ModifiersScreen;
