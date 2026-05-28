import React from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

export interface Crumb {
  label: string;
  to?: string;
}

const Breadcrumbs: React.FC<{ items: Crumb[] }> = ({ items }) => (
  <nav
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      fontSize: 12,
      color: 'var(--hub-text-muted)',
    }}
  >
    {items.map((item, i) => (
      <React.Fragment key={i}>
        {item.to ? (
          <Link to={item.to} style={{ color: 'var(--hub-text-muted)', textDecoration: 'none' }}>
            {item.label}
          </Link>
        ) : (
          <span style={{ color: 'var(--hub-text)', fontWeight: 600 }}>{item.label}</span>
        )}
        {i < items.length - 1 && (
          <ChevronRight size={12} style={{ color: 'var(--hub-text-dim)' }} />
        )}
      </React.Fragment>
    ))}
  </nav>
);

export default Breadcrumbs;
