import React from 'react';
import { ShoppingBag, Sparkles } from 'lucide-react';
import { Breadcrumbs, EmptyState } from '../../shared/components/dash';

const Marketplace: React.FC = () => {
  return (
    <div className="space-y-5 max-w-5xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(168,85,247,0.15)' }}
        >
          <ShoppingBag size={16} style={{ color: '#a855f7' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            App marketplace
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            Install third-party apps and extensions.
          </p>
        </div>
      </div>
      <EmptyState
        icon={Sparkles}
        title="Marketplace coming soon"
        description="We're lining up great partners. Developers can start building today using the API and webhook system."
      />
    </div>
  );
};

export default Marketplace;
