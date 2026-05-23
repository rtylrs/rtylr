import React from 'react';
import { DraggableIcon, DesktopIconDef } from './DraggableIcon';
import { POSIcon } from './icons/POSIcon';
import { ERPIcon } from './icons/ERPIcon';
import { CRMIcon } from './icons/CRMIcon';
import { HRIcon } from './icons/HRIcon';
import { FinanceIcon } from './icons/FinanceIcon';
import { FlowIcon } from './icons/FlowIcon';
import { PricingIcon } from './icons/PricingIcon';
import { DocsIcon } from './icons/DocsIcon';
import { BlogIcon } from './icons/BlogIcon';
import { ChangelogIcon } from './icons/ChangelogIcon';
import { WindowType } from '../../stores/useDesktopStore';

const ICONS: DesktopIconDef[] = [
  {
    key: 'pos',
    label: 'POS · vbitez',
    svgIcon: POSIcon,
    windowType: 'pos',
    defaultPos: { x: 20, y: 16 },
  },
  { key: 'erp', label: 'ERP', svgIcon: ERPIcon, windowType: 'erp', defaultPos: { x: 110, y: 16 } },
  { key: 'crm', label: 'CRM', svgIcon: CRMIcon, windowType: 'crm', defaultPos: { x: 20, y: 106 } },
  { key: 'hr', label: 'HR', svgIcon: HRIcon, windowType: 'hr', defaultPos: { x: 110, y: 106 } },
  {
    key: 'finance',
    label: 'Finance',
    svgIcon: FinanceIcon,
    windowType: 'finance',
    defaultPos: { x: 20, y: 196 },
  },
  {
    key: 'flow',
    label: 'Flow',
    svgIcon: FlowIcon,
    windowType: 'flow',
    defaultPos: { x: 110, y: 196 },
  },
  {
    key: 'pricing',
    label: 'Pricing',
    svgIcon: PricingIcon,
    windowType: 'pricing',
    defaultPos: { x: 20, y: 286 },
  },
  {
    key: 'docs',
    label: 'Docs',
    svgIcon: DocsIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com',
    defaultPos: { x: 110, y: 286 },
  },
  {
    key: 'blog',
    label: 'Blog',
    svgIcon: BlogIcon,
    windowType: 'docs',
    windowUrl: 'https://blog.rtylr.com',
    defaultPos: { x: 20, y: 376 },
  },
  {
    key: 'changelog',
    label: 'Changelog',
    svgIcon: ChangelogIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com/changelog',
    defaultPos: { x: 110, y: 376 },
  },
];

export const DesktopIcons: React.FC = () => (
  <>
    {ICONS.map((def) => (
      <DraggableIcon key={def.key} def={def} />
    ))}
  </>
);
