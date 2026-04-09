import React from 'react';
import { Link } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';

interface Crumb {
  label: string;
  to?: string;
}

interface BreadcrumbsProps {
  items: Crumb[];
}

export const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ items }) => {
  return (
    <nav className="flex items-center gap-1 text-[12px]">
      {items.map((c, i) => (
        <React.Fragment key={i}>
          {i > 0 && (
            <ChevronRight
              size={11}
              style={{ color: 'var(--flow-text-dim)' }}
            />
          )}
          {c.to ? (
            <Link
              to={c.to}
              className="hover:text-[var(--flow-text)] transition-colors"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {c.label}
            </Link>
          ) : (
            <span style={{ color: 'var(--flow-text)', fontWeight: 500 }}>
              {c.label}
            </span>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
};

export default Breadcrumbs;
