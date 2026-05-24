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
import { TerminalIcon } from './icons/TerminalIcon';
import { DoomIcon } from './icons/DoomIcon';
import { StatusIcon } from './icons/StatusIcon';
import { SnakeIcon } from './icons/SnakeIcon';
import { WindowType } from '../../stores/useDesktopStore';

// 3-column grid: col1=x:16, col2=x:88, col3=x:160, row-height=80px -> max ~412px total
const ICONS: DesktopIconDef[] = [
  { key: 'pos', label: 'POS', svgIcon: POSIcon, windowType: 'pos', defaultPos: { x: 16, y: 16 } },
  { key: 'erp', label: 'ERP', svgIcon: ERPIcon, windowType: 'erp', defaultPos: { x: 88, y: 16 } },
  { key: 'crm', label: 'CRM', svgIcon: CRMIcon, windowType: 'crm', defaultPos: { x: 160, y: 16 } },
  { key: 'hr', label: 'HR', svgIcon: HRIcon, windowType: 'hr', defaultPos: { x: 16, y: 96 } },
  {
    key: 'finance',
    label: 'Finance',
    svgIcon: FinanceIcon,
    windowType: 'finance',
    defaultPos: { x: 88, y: 96 },
  },
  {
    key: 'flow',
    label: 'Flow',
    svgIcon: FlowIcon,
    windowType: 'flow',
    defaultPos: { x: 160, y: 96 },
  },
  {
    key: 'pricing',
    label: 'Pricing',
    svgIcon: PricingIcon,
    windowType: 'pricing',
    windowUrl: '/pricing?embed=1',
    defaultPos: { x: 16, y: 176 },
  },
  {
    key: 'docs',
    label: 'Docs',
    svgIcon: DocsIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com',
    defaultPos: { x: 88, y: 176 },
  },
  {
    key: 'blog',
    label: 'Blog',
    svgIcon: BlogIcon,
    windowType: 'docs',
    windowUrl: 'https://blog.rtylr.com',
    defaultPos: { x: 160, y: 176 },
  },
  {
    key: 'changelog',
    label: 'Changelog',
    svgIcon: ChangelogIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com/changelog',
    defaultPos: { x: 16, y: 256 },
  },
  {
    key: 'terminal',
    label: 'Terminal',
    svgIcon: TerminalIcon,
    windowType: 'terminal',
    defaultPos: { x: 88, y: 256 },
  },
  {
    key: 'snake',
    label: 'Snake',
    svgIcon: SnakeIcon,
    windowType: 'snake',
    defaultPos: { x: 160, y: 256 },
  },
  {
    key: 'doom',
    label: 'Doom',
    svgIcon: DoomIcon,
    windowType: 'doom',
    windowUrl: 'https://archive.org/embed/doom-1993',
    defaultPos: { x: 16, y: 336 },
  },
  {
    key: 'status',
    label: 'Status',
    svgIcon: StatusIcon,
    windowType: 'status',
    windowUrl: 'https://status.rtylr.com',
    defaultPos: { x: 88, y: 336 },
  },
];

export const DesktopIcons: React.FC = () => (
  <>
    {ICONS.map((def) => (
      <DraggableIcon key={def.key} def={def} />
    ))}
  </>
);
