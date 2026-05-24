import React, { Suspense, lazy } from 'react';
import { WindowType } from '../../stores/useDesktopStore';

const POSWindow = lazy(() => import('../modules/POSWindow'));
const ShortcutsWindow = lazy(() => import('../modules/ShortcutsWindow'));
const ERPWindow = lazy(() => import('../modules/ERPWindow'));
const CRMWindow = lazy(() => import('../modules/CRMWindow'));
const HRWindow = lazy(() => import('../modules/HRWindow'));
const FinanceWindow = lazy(() => import('../modules/FinanceWindow'));
const FlowWindow = lazy(() => import('../modules/FlowWindow'));
const SettingsWindow = lazy(() => import('../modules/SettingsWindow'));
const AboutWindow = lazy(() => import('../modules/AboutWindow'));
const LandingWindow = lazy(() => import('../modules/LandingWindow'));
const PricingWindow = lazy(() => import('../modules/PricingWindow'));
const TerminalWindow = lazy(() => import('../modules/TerminalWindow'));
const SnakeWindow = lazy(() => import('../modules/SnakeWindow'));
const StatusWindow = lazy(() => import('../modules/StatusWindow'));

const LOADERS: Partial<Record<WindowType, React.LazyExoticComponent<() => React.JSX.Element>>> = {
  pos: POSWindow,
  erp: ERPWindow,
  crm: CRMWindow,
  hr: HRWindow,
  finance: FinanceWindow,
  flow: FlowWindow,
  settings: SettingsWindow,
  about: AboutWindow,
  landing: LandingWindow,
  pricing: PricingWindow,
  shortcuts: ShortcutsWindow,
  terminal: TerminalWindow,
  snake: SnakeWindow,
  status: StatusWindow,
};

const Spinner = () => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100%',
      color: 'rgba(255,255,255,0.3)',
      fontSize: 13,
    }}
  >
    Loading…
  </div>
);

interface WindowContentProps {
  type: WindowType;
  url?: string;
}

export const WindowContent: React.FC<WindowContentProps> = ({ type, url }) => {
  // If a URL is provided, render an iframe (used for docs, about, changelog, blog, etc.)
  if (url) {
    return (
      <iframe
        src={url}
        style={{ width: '100%', height: '100%', border: 'none', display: 'block' }}
        sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-pointer-lock"
        title={type}
      />
    );
  }

  const Component = LOADERS[type];
  if (!Component) return <Spinner />;

  return (
    <Suspense fallback={<Spinner />}>
      <Component />
    </Suspense>
  );
};
