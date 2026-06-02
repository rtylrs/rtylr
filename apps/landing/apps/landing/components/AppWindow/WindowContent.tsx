import React, { Suspense, lazy, Component } from 'react';
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
const DoomWindow = lazy(() => import('../modules/DoomWindow'));

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
  doom: DoomWindow,
};

const DARK_FILL = {
  width: '100%',
  height: '100%',
  background: 'rgba(18,18,20,0.98)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
} as const;

const Spinner = () => (
  <div style={DARK_FILL}>
    <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 13 }}>Loading...</span>
  </div>
);

const ChunkError = () => (
  <div
    style={{
      ...DARK_FILL,
      flexDirection: 'column',
      gap: 12,
      color: 'rgba(255,255,255,0.4)',
      fontSize: 13,
    }}
  >
    <span>Failed to load window.</span>
    <button
      onClick={() => window.location.reload()}
      style={{
        fontSize: 12,
        color: 'rgba(255,255,255,0.5)',
        background: 'none',
        border: '1px solid rgba(255,255,255,0.15)',
        borderRadius: 6,
        padding: '6px 14px',
        cursor: 'pointer',
      }}
    >
      Reload page
    </button>
  </div>
);

interface BoundaryState {
  error: boolean;
}

class ChunkErrorBoundary extends Component<React.PropsWithChildren, BoundaryState> {
  state: BoundaryState = { error: false };

  static getDerivedStateFromError(): BoundaryState {
    return { error: true };
  }

  render() {
    if (this.state.error) return <ChunkError />;
    return this.props.children;
  }
}

interface WindowContentProps {
  type: WindowType;
  url?: string;
}

export const WindowContent: React.FC<WindowContentProps> = ({ type, url }) => {
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

  const WindowComponent = LOADERS[type];
  if (!WindowComponent) return <Spinner />;

  return (
    <ChunkErrorBoundary>
      <Suspense fallback={<Spinner />}>
        <WindowComponent />
      </Suspense>
    </ChunkErrorBoundary>
  );
};
