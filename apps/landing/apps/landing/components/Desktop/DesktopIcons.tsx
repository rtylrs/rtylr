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

const ICONS: DesktopIconDef[] = [
  { key: 'pos', label: 'POS', svgIcon: POSIcon, href: '/pos', defaultPos: { x: 0, y: 0 } },
  { key: 'erp', label: 'ERP', svgIcon: ERPIcon, href: '/erp', defaultPos: { x: 0, y: 0 } },
  { key: 'crm', label: 'CRM', svgIcon: CRMIcon, href: '/crm', defaultPos: { x: 0, y: 0 } },
  { key: 'hr', label: 'HR', svgIcon: HRIcon, href: '/hr', defaultPos: { x: 0, y: 0 } },
  {
    key: 'finance',
    label: 'Finance',
    svgIcon: FinanceIcon,
    href: '/finance',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'flow',
    label: 'Flow',
    svgIcon: FlowIcon,
    href: '/flow',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'pricing',
    label: 'Pricing',
    svgIcon: PricingIcon,
    windowType: 'pricing',
    windowUrl: '/pricing?embed=1',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'docs',
    label: 'Docs',
    svgIcon: DocsIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'blog',
    label: 'Blog',
    svgIcon: BlogIcon,
    windowType: 'docs',
    windowUrl: 'https://blog.rtylr.com',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'changelog',
    label: 'Changelog',
    svgIcon: ChangelogIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com/changelog',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'terminal',
    label: 'Terminal',
    svgIcon: TerminalIcon,
    windowType: 'terminal',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'snake',
    label: 'Snake',
    svgIcon: SnakeIcon,
    windowType: 'snake',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'doom',
    label: 'Doom',
    svgIcon: DoomIcon,
    windowType: 'doom',
    defaultPos: { x: 0, y: 0 },
  },
  {
    key: 'status',
    label: 'Status',
    svgIcon: StatusIcon,
    windowType: 'status',
    defaultPos: { x: 0, y: 0 },
  },
];

// Flex column that wraps to a second column when icons overflow the desktop height.
// Icons fill top-to-bottom in col 1, overflow goes to col 2.
export const DesktopIcons: React.FC = () => (
  <div
    style={{
      position: 'absolute',
      top: 16,
      left: 16,
      bottom: 16,
      display: 'flex',
      flexDirection: 'column',
      flexWrap: 'wrap',
      alignContent: 'flex-start',
      gap: '4px 8px',
      pointerEvents: 'none',
    }}
  >
    {ICONS.map((def) => (
      <DraggableIcon key={def.key} def={def} />
    ))}
  </div>
);
