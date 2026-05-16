import React from 'react';
import { SettingsPanel } from '@voxire/caret';
import { PageHeader } from './components';

/** Full-page wrapper around the in-panel SettingsPanel. */
const CaretSettingsPage: React.FC = () => (
  <div className="space-y-4 max-w-2xl">
    <PageHeader title="Caret settings" subtitle="Control your AI assistant." />
    <div
      style={{
        background: 'var(--dash-surface)',
        border: '1px solid var(--dash-border)',
        borderRadius: 10,
        overflow: 'hidden',
      }}
    >
      <SettingsPanel />
    </div>
  </div>
);

export default CaretSettingsPage;
