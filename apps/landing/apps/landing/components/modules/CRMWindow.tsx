import React from 'react';
import { APP_URLS } from '@voxire/config';
import { CRMDemo } from '../home/FeaturesSection/demos/CRMDemo';
import { ModuleFooter } from './ModuleFooter';

export default function CRMWindow(): React.JSX.Element {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <CRMDemo isActive={true} />
      </div>
      <ModuleFooter href={APP_URLS.crm} label="Open CRM →" />
    </div>
  );
}
