import React from 'react';
import Breadcrumbs, { Crumb } from './Breadcrumbs';

interface PageHeaderProps {
  title: string;
  subtitle?: React.ReactNode;
  breadcrumbs?: Crumb[];
  actions?: React.ReactNode;
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, subtitle, breadcrumbs, actions }) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginBottom: 20,
    }}
  >
    <div>
      {breadcrumbs && breadcrumbs.length > 0 && (
        <div style={{ marginBottom: 6 }}>
          <Breadcrumbs items={breadcrumbs} />
        </div>
      )}
      <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
        {title}
      </h2>
      {subtitle && (
        <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
          {subtitle}
        </p>
      )}
    </div>
    {actions && <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>{actions}</div>}
  </div>
);

export default PageHeader;
