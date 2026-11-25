import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { APP_URLS } from '@voxire/config';
import {
  ArrowRight,
  BarChart3,
  Boxes,
  Calculator,
  HeartHandshake,
  Kanban,
  MonitorSmartphone,
  QrCode,
  Users,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { PAGER_RTYLR_PATH } from '../../../pagerPaths';
import { RtylrBrandText } from '../../../shared/components/RenniliLogo';

export interface RtylrServiceMeta {
  id: 'pos' | 'menu' | 'erp' | 'finance' | 'hr' | 'crm' | 'flow' | 'insights';
  icon: LucideIcon;
  href: string;
}

export const RTYLR_SERVICES: readonly RtylrServiceMeta[] = [
  { id: 'pos', icon: MonitorSmartphone, href: `${APP_URLS.landing}/workspace/#sales` },
  { id: 'menu', icon: QrCode, href: `${APP_URLS.landing}/workspace/#sales` },
  { id: 'erp', icon: Boxes, href: `${APP_URLS.landing}/workspace/#operations` },
  { id: 'finance', icon: Calculator, href: `${APP_URLS.landing}/workspace/#finance` },
  { id: 'hr', icon: Users, href: `${APP_URLS.landing}/workspace/#team` },
  { id: 'crm', icon: HeartHandshake, href: `${APP_URLS.landing}/workspace/#customers` },
  { id: 'flow', icon: Kanban, href: `${APP_URLS.landing}/workspace/#work` },
  { id: 'insights', icon: BarChart3, href: `${APP_URLS.landing}/workspace/#insights` },
] as const;

/** Compact homepage teaser driving venue owners to the dedicated `/rtylr` upsell page. */
export const RtylrUpsellBand = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const featured = RTYLR_SERVICES.slice(0, 4);

  return (
    <section className="rl-section rl-upsell-band" aria-labelledby="rl-upsell-heading">
      <div className="rl-wrap">
        <div className="rl-upsell-card">
          <div className="rl-upsell-copy">
            <p className="rl-kicker">
              <RtylrBrandText text={t('pager.landing.rtylr.eyebrow')} />
            </p>
            <h2 id="rl-upsell-heading" className="rl-upsell-title">
              {t('pager.landing.rtylr.teaserTitle')}
            </h2>
            <p className="rl-upsell-body">{t('pager.landing.rtylr.teaserBody')}</p>
            <div className="rl-upsell-actions">
              <Link to={PAGER_RTYLR_PATH} className="rl-btn rl-btn--primary">
                {t('pager.landing.rtylr.teaserCta')}
                <ArrowRight
                  size={16}
                  strokeWidth={2.4}
                  aria-hidden="true"
                  className="rtl:rotate-180"
                />
              </Link>
              <a
                href={`${APP_URLS.landing}/workspace/`}
                target="_blank"
                rel="noreferrer"
                className="rl-btn rl-btn--secondary"
              >
                {t('pager.landing.rtylr.visitWorkspace')}
              </a>
            </div>
          </div>

          <div className="rl-upsell-pills" aria-hidden="true">
            {featured.map((service) => {
              const Icon = service.icon;
              return (
                <div key={service.id} className="rl-upsell-mini">
                  <span className="rl-upsell-mini-icon">
                    <Icon size={18} strokeWidth={2} />
                  </span>
                  <div>
                    <strong>{t(`pager.landing.rtylr.services.${service.id}.title`)}</strong>
                    <span>{t(`pager.landing.rtylr.services.${service.id}.short`)}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
};
