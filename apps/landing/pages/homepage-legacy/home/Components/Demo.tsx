import { useTranslation } from 'react-i18next';
import { Button } from '@voxire/shared';
import React from 'react';
import useColors from '../../../../hooks/useColors';
import styles from '../Home.module.scss';

const Demo: React.FC<{}> = () => {
  const { t } = useTranslation();
  const { primary } = useColors();

  return (
    <div className={styles['demo-section']}>
      <div className={styles['demo-menu-container']}>
        <h2>{t('Demo.title')}</h2>
        <p>{t('Demo.description')}</p>
        <div className="qr-code">
          <img
            src="/images/VoxireWebsiteQR.png"
            alt="QR code for rtylr.voxire.com"
            style={{
              width: 300,
              maxWidth: '100%',
              height: 'auto',
            }}
          />
        </div>
        <Button
          variant="primary"
          styles={{
            borderColor: primary,
          }}
        >
          {t('Demo.button')}
        </Button>
      </div>
    </div>
  );
};

export default Demo;
