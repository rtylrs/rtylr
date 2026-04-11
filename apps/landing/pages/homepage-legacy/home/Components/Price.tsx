import { Button } from '@voxire/components';
import { HiStar } from 'react-icons/hi';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import React from 'react';
import { useUserStore } from '../../../../stores/userStore';
import styles from '../Home.module.scss';

const Price: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const { isAuthenticated } = useUserStore();

  return (
    <div className={styles['Prices']}>
      <h1 className={`rtylr-title ${styles['price-text']}`}>{t('PricingSection.title')}</h1>
      <p className={styles['rtylr-text-price-text']}>{t('PricingSection.description')}</p>
      <div className={styles['price-container']}>
        <div className={styles['price-card']}>
          <div className="prices-container">
            <h4>
              <b>{t('PricingSection.Section1_title')}</b>
            </h4>
            <h3>
              <span style={{ fontSize: '70px', fontWeight: 'bold' }}>
                {t('PricingSection.Section1_Price')}
              </span>
            </h3>
            <p>{t('PricingSection.Section1_Requirements')}</p>
            <hr style={{ border: '1px solid grey', margin: '20px 0' }} />
            <ul>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                {t('PricingSection.Section1_star1')}
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                {t('PricingSection.Section1_star2')}
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                Basic widget customization
              </li>
            </ul>
            <div className={styles['price-button']}>
              <span
                onClick={() => {
                  router.push('/pricing');
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
        <div className={`${styles['price-card']} popular`}>
          <div className="prices-container">
            <h4>
              <b>Advanced</b>
            </h4>
            <h3>
              $ <span style={{ fontSize: '70px', fontWeight: 'bold' }}>49</span>
              /year
            </h3>
            <p>For small restaurant or bar.</p>

            <hr style={{ border: '1px solid grey', margin: '20px 0' }} />
            <ul>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                Unlimited chats & websites
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                60-day conversations history
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                Basic widget customization
              </li>
            </ul>
            <div className={styles['price-button']}>
              {/* <Button>Get Started</Button> */}
              <span
                onClick={() => {
                  router.push('/pricing');
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
        <div className={styles['price-card']}>
          <div className="prices-container">
            <h4>
              <b>Premium</b>
            </h4>
            <h3>
              $ <span style={{ fontSize: '70px', fontWeight: 'bold' }}>59</span>
              /year
            </h3>
            <p>For high-demand restaurants and bars.</p>

            <hr style={{ border: '1px solid grey', margin: '20px 0' }} />
            <ul>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                Unlimited chats & websites
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                60-day conversations history
              </li>
              <li>
                <HiStar style={{ color: 'black', marginRight: '8px' }} />
                Basic widget customization
              </li>
            </ul>
            <div className={styles['price-button']}>
              {/* <Button>Get Started</Button> */}
              <span
                onClick={() => {
                  router.push('/pricing');
                }}
              >
                Learn More
              </span>
            </div>
          </div>
        </div>
      </div>
      <div
        style={{
          marginTop: '50px',
          marginBottom: '0px',
        }}
      >
        <Button
          onClick={() => {
            if (isAuthenticated) {
              router.push('/app/business');
            } else {
              router.push('/auth/register');
            }
          }}
        >
          {isAuthenticated ? 'Dashboard' : 'Get Started'}
        </Button>
      </div>
    </div>
  );
};

export default Price;
