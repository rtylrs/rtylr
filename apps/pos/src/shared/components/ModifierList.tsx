import React from 'react';
import type { Modifier, ModifierGroup } from '../types/pos';
import { formatCurrency } from '../utils/format';

interface ModifierListProps {
  groups: ModifierGroup[];
  selected: Record<string, string[]>; // group id -> modifier ids
  onToggle: (groupId: string, modifierId: string) => void;
}

const ModifierList: React.FC<ModifierListProps> = ({ groups, selected, onToggle }) => {
  return (
    <div className="space-y-4">
      {groups.map((g) => {
        const chosen = new Set(selected[g.id] ?? []);
        return (
          <div key={g.id} className="space-y-2">
            <div className="flex items-baseline justify-between">
              <h3 className="text-sm font-semibold text-pos-text">{g.name}</h3>
              <span className="text-[10px] text-pos-muted">
                {g.min > 0 ? 'Required' : 'Optional'} · max {g.max}
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2">
              {g.modifiers.map((m: Modifier) => {
                const active = chosen.has(m.id);
                return (
                  <button
                    key={m.id}
                    type="button"
                    onClick={() => onToggle(g.id, m.id)}
                    className={`flex items-center justify-between h-11 px-3 rounded-lg text-sm font-medium border transition ${
                      active
                        ? 'bg-pos-text text-pos-bg border-pos-text'
                        : 'bg-white/8 text-pos-text border-pos-border hover:bg-white/14'
                    }`}
                  >
                    <span>{m.name}</span>
                    {m.priceDelta !== 0 && (
                      <span className="text-[11px]">
                        {m.priceDelta > 0 ? '+' : ''}
                        {formatCurrency(m.priceDelta)}
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default ModifierList;
