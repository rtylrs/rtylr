import React from 'react';
import { Folder } from 'lucide-react';
import { useCategories } from '../../features/categories/useCategories';
import type { Category } from '../../features/categories/useCategories';

interface CategoryTreeProps {
  selectedId?: string;
  onSelect: (category: Category) => void;
}

const CategoryTree: React.FC<CategoryTreeProps> = ({ selectedId, onSelect }) => {
  const { data: categories = [], isLoading } = useCategories();

  if (isLoading) {
    return (
      <div className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
        Loading categories…
      </div>
    );
  }

  return (
    <div className="space-y-0.5">
      {categories.map((cat) => {
        const isSelected = selectedId === cat.id;
        return (
          <div key={cat.id}>
            <button
              type="button"
              onClick={() => {
                onSelect(cat);
              }}
              className="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-sm text-left transition-colors"
              style={{
                color: isSelected ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                background: isSelected ? 'var(--hub-active)' : 'transparent',
              }}
              onMouseEnter={(e) => {
                if (!isSelected)
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                if (!isSelected) (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <Folder size={14} />
              <span className="flex-1 truncate">{cat.name}</span>
            </button>
          </div>
        );
      })}
    </div>
  );
};

export default CategoryTree;
