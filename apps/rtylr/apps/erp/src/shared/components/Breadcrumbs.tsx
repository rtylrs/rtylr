import React from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

export interface Crumb {
  label: string;
  to?: string;
}

interface BreadcrumbsProps {
  items: Crumb[];
}

const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ items }) => {
  return (
    <nav className="flex items-center gap-1.5 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
      {items.map((item, i) => {
        const isLast = i === items.length - 1;
        return (
          <React.Fragment key={i}>
            {item.to && !isLast ? (
              <Link
                to={item.to}
                className="hover:underline"
                style={{ color: 'var(--hub-text-muted)', textDecoration: 'none' }}
              >
                {item.label}
              </Link>
            ) : (
              <span style={{ color: isLast ? 'var(--hub-text)' : 'var(--hub-text-muted)' }}>
                {item.label}
              </span>
            )}
            {!isLast && (
              <ChevronRight
                size={11}
                className="erp-directional-icon"
                style={{ color: 'var(--hub-text-dim)' }}
              />
            )}
          </React.Fragment>
        );
      })}
    </nav>
  );
};

export default Breadcrumbs;
