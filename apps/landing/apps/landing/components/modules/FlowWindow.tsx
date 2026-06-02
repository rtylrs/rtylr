import React from 'react';
import { APP_URLS } from '@voxire/config';
import { FlowDemo } from '../home/FeaturesSection/demos/FlowDemo';
import { ModuleFooter } from './ModuleFooter';

export default function FlowWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <FlowDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.flow} label="Open Flow →" />
    </div>
  );
}
