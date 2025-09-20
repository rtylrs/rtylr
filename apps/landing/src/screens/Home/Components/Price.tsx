import { Button } from '@voxire/shared';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faStar } from '@fortawesome/free-solid-svg-icons';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import React from 'react';
import { useUserStore } from '../../../stores/userStore';

const Price: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const { isAuthenticated } = useUserStore();

  return (
    <div className="Prices">
      <h1 className="vbitez-title price-text">{t('PricingSection.title')}</h1>
      <p className="vbitez-text-price-text">
        {t('PricingSection.description')}
      </p>
      <div className="price-container">
        <div className="price-card">
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
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                {t('PricingSection.Section1_star1')}
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                {t('PricingSection.Section1_star2')}
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
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
        <div className="price-card popular">
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
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                60-day conversations history
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
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
        <div className="price-card">
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
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                Unlimited chats & websites
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                60-day conversations history
              </li>
              <li>
                <FontAwesomeIcon
                  icon={faStar}
                  style={{ color: 'black', marginRight: '8px' }}
                />
                Basic widget customization
              </li>
            </ul>
            <div className="price-button">
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
