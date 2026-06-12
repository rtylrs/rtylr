import React from 'react';
import { DesktopIcon, DesktopIconDef } from './DesktopIcon';
import { LandingIcon } from './icons/LandingIcon';
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
  { key: 'landing', label: 'rtylr', svgIcon: LandingIcon, windowType: 'landing' },
  { key: 'pos', label: 'POS', svgIcon: POSIcon, href: '/pos' },
  { key: 'erp', label: 'ERP', svgIcon: ERPIcon, href: '/erp' },
  { key: 'crm', label: 'CRM', svgIcon: CRMIcon, href: '/crm' },
  { key: 'hr', label: 'HR', svgIcon: HRIcon, href: '/hr' },
  {
    key: 'finance',
    label: 'Finance',
    svgIcon: FinanceIcon,
    href: '/finance',
  },
  {
    key: 'flow',
    label: 'Flow',
    svgIcon: FlowIcon,
    href: '/flow',
  },
  {
    key: 'pricing',
    label: 'Pricing',
    svgIcon: PricingIcon,
    windowType: 'pricing',
    windowUrl: '/pricing?embed=1',
  },
  {
    key: 'docs',
    label: 'Docs',
    svgIcon: DocsIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com',
  },
  {
    key: 'blog',
    label: 'Blog',
    svgIcon: BlogIcon,
    href: '/blog',
  },
  {
    key: 'changelog',
    label: 'Changelog',
    svgIcon: ChangelogIcon,
    windowType: 'docs',
    windowUrl: 'https://docs.rtylr.com/changelog',
  },
  {
    key: 'terminal',
    label: 'Terminal',
    svgIcon: TerminalIcon,
    windowType: 'terminal',
  },
  {
    key: 'snake',
    label: 'Snake',
    svgIcon: SnakeIcon,
    windowType: 'snake',
  },
  {
    key: 'doom',
    label: 'Doom',
    svgIcon: DoomIcon,
    windowType: 'doom',
  },
  {
    key: 'status',
    label: 'Status',
    svgIcon: StatusIcon,
    windowType: 'status',
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
      <DesktopIcon key={def.key} def={def} />
    ))}
  </div>
);
