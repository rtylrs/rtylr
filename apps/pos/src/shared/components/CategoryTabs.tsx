import React from 'react';

interface CategoryTabsProps {
  categories: string[];
  active: string;
  onChange: (c: string) => void;
  showAll?: boolean;
}

const CategoryTabs: React.FC<CategoryTabsProps> = ({
  categories,
  active,
  onChange,
  showAll = true,
}) => {
  const all = showAll ? ['All', ...categories] : categories;
  return (
    <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none" role="tablist">
      {all.map((cat) => {
        const isActive = active === cat;
        return (
          <button
            key={cat}
            role="tab"
            aria-selected={isActive}
            onClick={() => onChange(cat)}
            className={`shrink-0 h-11 px-4 rounded-full text-sm font-semibold transition-colors ${
              isActive ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-muted hover:bg-white/14'
            }`}
          >
            {cat}
          </button>
        );
      })}
    </div>
  );
};

export default CategoryTabs;
