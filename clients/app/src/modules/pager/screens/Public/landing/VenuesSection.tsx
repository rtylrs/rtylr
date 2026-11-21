import React from 'react';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { Coffee, Croissant, IceCreamCone, Store, Truck, UtensilsCrossed } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface VenueItem {
  id: string;
  title: string;
  body: string;
}

const VENUE_ICONS: Record<string, LucideIcon> = {
  cafes: Coffee,
  restaurants: UtensilsCrossed,
  bakeries: Croissant,
  trucks: Truck,
  desserts: IceCreamCone,
  foodCourts: Store,
};

export const VenuesSection = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawItems = t('pager.landing.venues.items', { returnObjects: true });
  const items = Array.isArray(rawItems) ? (rawItems as VenueItem[]) : [];

  return (
    <section id="venues" className="rl-section rl-venues" aria-labelledby="rl-venues-heading">
      <div className="rl-wrap">
        <div className="rl-section-head">
          <p className="rl-kicker">{t('pager.landing.venues.kicker')}</p>
          <h2 id="rl-venues-heading" className="rl-section-title">
            {t('pager.landing.venues.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.venues.intro')}</p>
        </div>

        <div className="rl-venues-grid">
          {items.map((item, index) => {
            const Icon = VENUE_ICONS[item.id] ?? Store;
            return (
              <motion.article
                key={item.id}
                className="rl-venue-card"
                initial={{ opacity: 0, y: 14 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.2 }}
                transition={{
                  duration: 0.42,
                  delay: Math.min(index * 0.05, 0.25),
                  ease: [0.22, 1, 0.36, 1],
                }}
              >
                <span className="rl-venue-icon" aria-hidden="true">
                  <Icon size={20} strokeWidth={1.9} />
                </span>
                <h3 className="rl-venue-title">{item.title}</h3>
                <p className="rl-venue-body">{item.body}</p>
              </motion.article>
            );
          })}
        </div>
      </div>
    </section>
  );
};
