import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { Button } from '@voxire/shared';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faArrowRight, faStar } from '@fortawesome/free-solid-svg-icons';
import useColors from '../../../hooks/useColors';
import { useUserStore } from '../../../stores/userStore';
import { QRCode } from 'react-qrcode-logo';
import { useWindowDimensions } from '@voxire/shared/utils';
import styles from '../Home.module.scss';

const Hero: React.FC<{}> = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const { width } = useWindowDimensions();
  const { primary, secondary } = useColors();
  const { isAuthenticated } = useUserStore();

  const memoizedQRCode = useMemo(
    () => (
      <QRCode
        value="https://rtylr.voxire.com/"
        bgColor="#ffffff"
        fgColor={primary}
        size={width < 768 ? 300 : 400}
        qrStyle="fluid"
      />
    ),
    [width, primary]
  );

  return (
    <div className={styles['Hero']}>
      <div className={styles['hero-text']}>
        <h1 className="rtylr-title">{t('Hero.welcome')}</h1>
        <h2>{t('Hero.discover')}</h2>
        <div className={styles['hero-buttons']}>
          <Button
            onClick={() => {
              if (isAuthenticated) {
                router.push('/app/business');
              } else {
                router.push('/auth/register');
              }
            }}
          >
            {isAuthenticated ? t('Hero.My Account') : t('Hero.Create Account')}
          </Button>
          <Button
            onClick={() => router.push('/uses')}
            variant="outline"
            suffix={<FontAwesomeIcon icon={faArrowRight} />}
          >
            {t('Hero.Learn More')}
          </Button>
        </div>
        <h5
          style={{
            padding: 0,
            marginTop: 10,
          }}
        >
          <FontAwesomeIcon icon={faStar} color={secondary} /> Rated 4.7/5 - from
          over 500+ reviews
        </h5>
      </div>
      <div className={styles['hero-image']}>{memoizedQRCode}</div>
    </div>
  );
};

export default Hero;
