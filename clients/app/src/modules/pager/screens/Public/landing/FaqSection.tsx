import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AnimatePresence, motion } from 'motion/react';

interface FaqItem {
  q: string;
  a: string;
}

export const FaqSection = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawItems = t('pager.landing.faq.items', { returnObjects: true });
  const items = Array.isArray(rawItems) ? (rawItems as FaqItem[]) : [];
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  return (
    <section id="faq" className="rl-section" aria-labelledby="rl-faq-heading">
      <div className="rl-wrap rl-faq-layout">
        <div className="rl-faq-head">
          <p className="rl-kicker">{t('pager.landing.faq.kicker')}</p>
          <h2 id="rl-faq-heading" className="rl-section-title">
            {t('pager.landing.faq.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.faq.intro')}</p>
        </div>

        <div className="rl-faq-list">
          {items.map((item, index) => {
            const isOpen = openIndex === index;
            const panelId = `rl-faq-panel-${index}`;
            return (
              <div key={item.q} className={`rl-faq-item ${isOpen ? 'rl-faq-item--open' : ''}`}>
                <h3 className="rl-faq-h">
                  <button
                    type="button"
                    className="rl-faq-q"
                    aria-expanded={isOpen}
                    aria-controls={panelId}
                    onClick={() => setOpenIndex(isOpen ? null : index)}
                  >
                    <span>{item.q}</span>
                    <span className="rl-faq-toggle" aria-hidden="true" />
                  </button>
                </h3>
                <AnimatePresence initial={false}>
                  {isOpen && (
                    <motion.div
                      id={panelId}
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
                      style={{ overflow: 'hidden' }}
                    >
                      <p className="rl-faq-a">{item.a}</p>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};
