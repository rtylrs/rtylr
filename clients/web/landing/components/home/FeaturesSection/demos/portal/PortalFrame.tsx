import React from 'react';
import {
  BarChart2,
  Bell,
  Briefcase,
  Heart,
  Home,
  Package,
  Search,
  ShoppingCart,
  Users,
  Wallet,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../../lib/cn';
import { CaretMark } from '../../../../CaretMark/CaretMark';
import styles from './PortalDemo.module.scss';

export type AppId = 'flow' | 'finance' | 'crm' | 'hr' | 'erp' | 'pos' | 'insights';
export type PortalView = AppId | 'home';

// Rail order and accent colours mirror the portal (clients/app: --app-*).
const RAIL: { id: PortalView; Icon: React.ComponentType<{ size?: number }>; color?: string }[] =
  [
    { id: 'home', Icon: Home },
    { id: 'insights', Icon: BarChart2, color: '#6fa3db' },
    { id: 'flow', Icon: Briefcase, color: '#ef4444' },
    { id: 'crm', Icon: Heart, color: '#8b5cf6' },
    { id: 'erp', Icon: Package, color: '#64c5b9' },
    { id: 'finance', Icon: Wallet, color: '#f59e0b' },
    { id: 'hr', Icon: Users, color: '#10b981' },
    { id: 'pos', Icon: ShoppingCart, color: '#3b82f6' },
  ];

interface PortalFrameProps {
  app: PortalView;
  crumb: string[];
  /** Right side of the header, before the shared icons (view toggles etc). */
  actions?: React.ReactNode;
  children: React.ReactNode;
  /** Makes the rail clickable (the hero's live portal). Omit for a static demo. */
  onNavigate?: (view: PortalView) => void;
}

/**
 * A scaled-down rtylr portal: icon rail with the active app lit in its
 * colour, breadcrumb header, and a content area on the portal's light
 * surface. Every feature demo renders inside one of these.
 */
export const PortalFrame: React.FC<PortalFrameProps> = ({
  app,
  crumb,
  actions,
  children,
  onNavigate,
}) => {
  const { t } = useTranslation();
  const accent = RAIL.find((r) => r.id === app)?.color ?? '#09090b';

  return (
    <div className={styles.frame} style={{ '--p-accent': accent } as React.CSSProperties}>
      <nav className={styles.rail} aria-hidden={onNavigate ? undefined : true}>
        {RAIL.map(({ id, Icon }) =>
          onNavigate ? (
            <button
              key={id}
              type="button"
              className={cn(styles.railItem, styles.railButton, id === app && styles.railActive)}
              aria-label={t(`home.hero.portal.apps.${id}`)}
              aria-current={id === app ? 'page' : undefined}
              onClick={() => onNavigate(id)}
            >
              <Icon size={15} />
            </button>
          ) : (
            <span key={id} className={cn(styles.railItem, id === app && styles.railActive)}>
              <Icon size={15} />
            </span>
          ),
        )}
        <span className={cn(styles.railItem, styles.railCaret)}>
          <CaretMark size={15} />
        </span>
      </nav>

      <div className={styles.main}>
        <header className={styles.header}>
          <div className={styles.crumb}>
            {crumb.map((c, i) => (
              <React.Fragment key={c}>
                {i > 0 && <span className={styles.crumbSep}>/</span>}
                <span className={i === crumb.length - 1 ? styles.crumbCurrent : undefined}>
                  {c}
                </span>
              </React.Fragment>
            ))}
          </div>
          <div className={styles.headerActions}>
            {actions}
            <span className={styles.iconBtn} title={t('home.features.demos.frame.caret')}>
              <CaretMark size={13} />
            </span>
            <span className={styles.iconBtn}>
              <Search size={13} />
            </span>
            <span className={styles.iconBtn}>
              <Bell size={13} />
            </span>
          </div>
        </header>
        <div className={styles.body}>{children}</div>
      </div>
    </div>
  );
};

export default PortalFrame;
