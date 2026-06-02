import React from 'react';
import { APP_URLS } from '@voxire/config';
import { ERPDemo } from '../home/FeaturesSection/demos/ERPDemo';
import { ModuleFooter } from './ModuleFooter';

export default function ERPWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <ERPDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.erp} label="Open ERP →" />
    </div>
  );
}
