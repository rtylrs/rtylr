import React from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

interface Crumb {
  label: string;
  to?: string;
  /** For a crumb that is not a route, such as a task opened in the drawer. */
  onClick?: () => void;
  /** A dot of this colour before the label, such as an epic's colour. */
  color?: string;
}

interface BreadcrumbsProps {
  items: Crumb[];
}

const linkClass =
  'min-w-0 max-w-[42vw] truncate hover:text-[var(--flow-text)] transition-colors sm:max-w-[12rem]';

export const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ items }) => {
  return (
    <nav className="flex min-w-0 items-center gap-1 text-flow-meta sm:text-flow-meta">
      {items.map((c, i) => (
        <React.Fragment key={i}>
          {i > 0 && (
            <ChevronRight
              size={11}
              className="shrink-0"
              style={{ color: 'var(--flow-text-dim)' }}
            />
          )}
          {c.color && (
            <span
              aria-hidden="true"
              className="h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ background: c.color }}
            />
          )}
          {c.to ? (
            <Link
              to={c.to as string}
              className={linkClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {c.label}
            </Link>
          ) : c.onClick ? (
            <button
              type="button"
              onClick={c.onClick}
              className={`flow-focus ${linkClass} text-start`}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {c.label}
            </button>
          ) : (
            // The current page keeps its width and the parents give way: on a
            // phone the task key was cut to "A…" while the project name stayed.
            <span
              className="max-w-full shrink-0 truncate font-medium"
              style={{ color: 'var(--flow-text)' }}
            >
              {c.label}
            </span>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
};

export default Breadcrumbs;
