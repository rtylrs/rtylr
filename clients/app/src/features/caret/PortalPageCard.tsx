import type { ComponentType, CSSProperties } from 'react';
import { useTranslation } from 'react-i18next';
import { useRouter } from '@tanstack/react-router';
import { ArrowUpRight, Lock } from 'lucide-react';
import type { CaretPageSpec } from '@voxire/caret';
import { useUserStore } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { describePlace, isPortalRoute, type PortalApp } from './portalPlaces';
import './page-card.css';

export interface RailPreviewItem {
  id: string;
  icon: ComponentType<{ size?: number }>;
}

interface PortalPageCardProps {
  spec: CaretPageSpec;
  /** The person's app bar as arranged, so the sketch looks like their Portal. */
  rail: { top: RailPreviewItem[]; bottom: RailPreviewItem[] };
  onOpen: (spec: CaretPageSpec) => void;
}

// The app bar icon that leads to each app.
const RAIL_ID: Partial<Record<PortalApp, string>> = {
  workspace: 'home',
  insights: 'insights',
  flow: 'flow',
  crm: 'crm',
  erp: 'erp',
  finance: 'finance',
  hr: 'hr',
  automations: 'automations',
  caret: 'caret',
};

const SKETCH_ROWS = ['92%', '78%', '85%', '64%'];

/** A page Caret points at: a sketch of where it is in the Portal and the
 *  button to press, opening it with Caret still beside it. */
export default function PortalPageCard({ spec, rail, onOpen }: PortalPageCardProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const user = useUserStore((state) => state.user);
  const place = describePlace(spec.path, t);
  const title = place.title ?? spec.title ?? spec.path;
  const exists = isPortalRoute(router, spec.path);
  const allowed = exists && canAccessPortalPath(user, spec.path.split(/[?#]/)[0]);
  const activeRail = RAIL_ID[place.app];
  const activeIndex = Math.max(
    0,
    place.items.findIndex((item) => item.active)
  );
  // A long sidebar is cut to the few items around the page.
  const from = Math.max(0, Math.min(activeIndex - 2, place.items.length - 6));
  const items = place.items.slice(from, from + 6);
  const status = !exists
    ? t('common.caret.pageCard.missing')
    : !allowed
      ? t('common.caret.pageCard.noAccess')
      : t('common.caret.pageCard.open');
  const railIcon = (item: RailPreviewItem) => (
    <span key={item.id} className={`ppc-rail-icon${item.id === activeRail ? ' is-active' : ''}`}>
      <item.icon size={9} />
    </span>
  );

  return (
    <button
      type="button"
      className="portal-page-card"
      data-app={place.app}
      style={{ '--ppc-accent': `var(--app-${place.app}, var(--dash-text))` } as CSSProperties}
      disabled={!allowed}
      onClick={() => onOpen(spec)}
      aria-label={
        allowed
          ? t('common.caret.pageCard.openLabel', { page: title, app: place.appLabel })
          : `${title}: ${status}`
      }
    >
      <span className="ppc-sketch" aria-hidden="true">
        <span className="ppc-rail">
          {rail.top.map(railIcon)}
          <span className="ppc-rail-spacer" />
          {rail.bottom.map(railIcon)}
        </span>
        {items.length > 0 && (
          <span className="ppc-sidebar">
            <span className="ppc-sidebar-app">{place.appLabel}</span>
            {items.map((item) => (
              <span key={item.key} className={`ppc-nav${item.active ? ' is-active' : ''}`}>
                {item.label}
              </span>
            ))}
          </span>
        )}
        <span className="ppc-page">
          <span className="ppc-page-head">
            <span className="ppc-page-title">{title}</span>
            {spec.action && <span className="ppc-action">{spec.action}</span>}
          </span>
          {SKETCH_ROWS.map((width) => (
            <span key={width} className="ppc-row" style={{ width }} />
          ))}
        </span>
      </span>
      <span className="ppc-footer">
        <span className="ppc-icon">
          <place.icon size={16} />
        </span>
        <span className="ppc-text">
          <span className="ppc-title">{title}</span>
          <span className="ppc-meta">
            {spec.action
              ? t('common.caret.pageCard.press', { app: place.appLabel, action: spec.action })
              : place.appLabel}
          </span>
        </span>
        <span className="ppc-open">
          {status}
          {allowed ? <ArrowUpRight size={14} /> : <Lock size={13} />}
        </span>
      </span>
    </button>
  );
}
