import React from 'react';
import { BusinessOffer } from '../../types';
import styles from './OffersPage.module.scss';

interface OffersPageProps {
  offers: BusinessOffer[];
}

const OffersPage: React.FC<OffersPageProps> = ({ offers }) => {
  if (!offers || offers.length === 0) {
    return (
      <div className={styles.offers__empty}>
        <span className={styles.offers__emptyIcon}>🏷️</span>
        <p>No active offers right now. Check back soon!</p>
      </div>
    );
  }

  return (
    <div className={styles.offers}>
      {offers.map((offer) => (
        <div key={offer.id} className={styles.offers__card}>
          {offer.image && (
            <div className={styles.offers__image}>
              <img src={offer.image} alt={offer.title} loading="lazy" />
              {offer.badge && (
                <span className={styles.offers__badge}>{offer.badge}</span>
              )}
            </div>
          )}
          <div className={styles.offers__body}>
            {!offer.image && offer.badge && (
              <span className={styles.offers__badgeInline}>{offer.badge}</span>
            )}
            <h3 className={styles.offers__title}>{offer.title}</h3>
            <p className={styles.offers__description}>{offer.description}</p>
            {offer.validUntil && (
              <span className={styles.offers__validity}>Valid until {offer.validUntil}</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
};

export default OffersPage;
