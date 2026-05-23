import React from 'react';
import { APP_URLS } from '@voxire/config';
import { FinanceDemo } from '../home/FeaturesSection/demos/FinanceDemo';
import { ModuleFooter } from './ModuleFooter';

export default function FinanceWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <FinanceDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.finance} label="Open Finance →" />
    </div>
  );
}
