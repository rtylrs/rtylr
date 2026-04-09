import React from 'react';
import { ChevronRight, Folder, FolderOpen } from 'lucide-react';
import { useCategories } from '../hooks/useCategories';
import { Category } from '../services/erpApi';

interface CategoryTreeProps {
  selectedId?: number;
  onSelect: (category: Category) => void;
}

const CategoryTree: React.FC<CategoryTreeProps> = ({ selectedId, onSelect }) => {
  const { data: categories = [], isLoading } = useCategories();
  const [expanded, setExpanded] = React.useState<Set<number>>(new Set());

  const toggle = (id: number) => {
    const next = new Set(expanded);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setExpanded(next);
  };

  if (isLoading) {
    return <div className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>Loading categories…</div>;
  }

  const rootCategories = categories.filter((c) => !c.parentId);

  return (
    <div className="space-y-0.5">
      {rootCategories.map((cat) => {
        const isExpanded = expanded.has(cat.id);
        const isSelected = selectedId === cat.id;
        return (
          <div key={cat.id}>
            <button
              onClick={() => {
                toggle(cat.id);
                onSelect(cat);
              }}
              className="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-sm text-left transition-colors"
              style={{
                color: isSelected ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                background: isSelected ? 'var(--hub-active)' : 'transparent',
              }}
              onMouseEnter={(e) => { if (!isSelected) (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
              onMouseLeave={(e) => { if (!isSelected) (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
            >
              <ChevronRight
                size={12}
                style={{
                  transform: isExpanded ? 'rotate(90deg)' : 'rotate(0)',
                  transition: 'transform 0.15s',
                }}
              />
              {isExpanded ? <FolderOpen size={14} /> : <Folder size={14} />}
              <span className="flex-1 truncate">{cat.name}</span>
              <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                {cat.productCount}
              </span>
            </button>
          </div>
        );
      })}
    </div>
  );
};

export default CategoryTree;
