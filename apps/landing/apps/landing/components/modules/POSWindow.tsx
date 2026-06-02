import React from 'react';
import { APP_URLS } from '@voxire/config';
import { POSDemo } from '../home/FeaturesSection/demos/POSDemo';
import { ModuleFooter } from './ModuleFooter';

export default function POSWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <POSDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.pos} label="Open POS →" />
    </div>
  );
}
