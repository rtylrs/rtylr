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

// 2-column grid anchored to bottom-left.
// col1=x:8, col2=x:80, row-height=80px, rows read top-to-bottom within the container.
// Container sits at bottom of desktop so these positions are container-relative.
const ICONS: DesktopIconDef[] = [
  { key: 'pos', label: 'POS', svgIcon: POSIcon, windowType: 'pos', defaultPos: { x: 8, y: 0 } },
  { key: 'erp', label: 'ERP', svgIcon: ERPIcon, windowType: 'erp', defaultPos: { x: 80, y: 0 } },
  { key: 'crm', label: 'CRM', svgIcon: CRMIcon, windowType: 'crm', defaultPos: { x: 8, y: 80 } },
  { key: 'hr', label: 'HR', svgIcon: HRIcon, windowType: 'hr', defaultPos: { x: 80, y: 80 } },
  {
    key: 'finance',
    label: 'Finance',
    svgIcon: FinanceIcon,
    windowType: 'finance',
    defaultPos: { x: 8, y: 160 },
  },
  {
    key: 'flow',
    label: 'Flow',
    svgIcon: FlowIcon,
    windowType: 'flow',
    defaultPos: { x: 80, y: 160 },
  },
  {
    key: 'pricing',
    label: 'Pricing',
    svgIcon: PricingIcon,
    windowType: 'pricing',
    windowUrl: '/pricing?embed=1',
    defaultPos: { x: 8, y: 240 },
  },
  {
    key: 'docs',
    label: 'Docs',
    svgIcon: DocsIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com',
    defaultPos: { x: 80, y: 240 },
  },
  {
    key: 'blog',
    label: 'Blog',
    svgIcon: BlogIcon,
    windowType: 'docs',
    windowUrl: 'https://blog.rtylr.com',
    defaultPos: { x: 8, y: 320 },
  },
  {
    key: 'changelog',
    label: 'Changelog',
    svgIcon: ChangelogIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com/changelog',
    defaultPos: { x: 80, y: 320 },
  },
  {
    key: 'terminal',
    label: 'Terminal',
    svgIcon: TerminalIcon,
    windowType: 'terminal',
    defaultPos: { x: 8, y: 400 },
  },
  {
    key: 'snake',
    label: 'Snake',
    svgIcon: SnakeIcon,
    windowType: 'snake',
    defaultPos: { x: 80, y: 400 },
  },
  {
    key: 'doom',
    label: 'Doom',
    svgIcon: DoomIcon,
    windowType: 'doom',
    defaultPos: { x: 8, y: 480 },
  },
  {
    key: 'status',
    label: 'Status',
    svgIcon: StatusIcon,
    windowType: 'status',
    defaultPos: { x: 80, y: 480 },
  },
];

// Container anchored to bottom-left. Icons position themselves absolutely within it.
// Height = 7 rows × 80px + icon height (~60px) = 620px so all icons fit.
export const DesktopIcons: React.FC = () => (
  <div
    style={{
      position: 'absolute',
      bottom: 24,
      left: 16,
      width: 152,
      height: 560,
      pointerEvents: 'none',
    }}
  >
    {ICONS.map((def) => (
      <DraggableIcon key={def.key} def={def} />
    ))}
  </div>
);
