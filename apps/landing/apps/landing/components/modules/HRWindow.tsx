import React from 'react';
import { APP_URLS } from '@voxire/config';
import { HRDemo } from '../home/FeaturesSection/demos/HRDemo';
import { ModuleFooter } from './ModuleFooter';

export default function HRWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <HRDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.hr} label="Open HR →" />
    </div>
  );
}
