import React from 'react';
import { useCaretTools } from '../../shared/hooks/queries/useCaretTools';
import { PageHeader, ToolsList } from './components';

const CaretToolsPage: React.FC = () => {
  const { data: tools = [], isLoading } = useCaretTools();
  return (
    <div className="space-y-4 max-w-3xl">
      <PageHeader
        title="Caret tools"
        subtitle="Everything Caret can call on your behalf, grouped by module."
      />
      {isLoading && <p style={{ color: 'var(--dash-text-muted)' }}>Loading…</p>}
      {!isLoading && tools.length === 0 && (
        <p style={{ color: 'var(--dash-text-muted)' }}>No tools available yet.</p>
      )}
      <ToolsList tools={tools} />
    </div>
  );
};

export default CaretToolsPage;
