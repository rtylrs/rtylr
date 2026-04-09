import React from 'react';
import Breadcrumbs, { Crumb } from './Breadcrumbs';

interface PageHeaderProps {
  title: string;
  description?: string;
  breadcrumbs?: Crumb[];
  actions?: React.ReactNode;
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, description, breadcrumbs, actions }) => {
  return (
    <div className="flex items-start justify-between">
      <div>
        {breadcrumbs && breadcrumbs.length > 0 && (
          <div className="mb-2">
            <Breadcrumbs items={breadcrumbs} />
          </div>
        )}
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {title}
        </h2>
        {description && (
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {description}
          </p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
};

export default PageHeader;
