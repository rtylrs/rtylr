import React from 'react';
import { useTranslation } from 'react-i18next';
import { HiChartBar, HiPencil, HiStar, HiCamera, HiGlobeAlt, HiFilter } from 'react-icons/hi';
import styles from '../Home.module.scss';

const Features: React.FC<{}> = () => {
  const { t } = useTranslation();

  return (
    <div className={styles['Features']}>
      <h1 className={`rtylr-title ${styles['features-text']}`}>{t('Features.Features-title')}</h1>

      <div className={styles['features-container']}>
        <div className={styles['feature']}>
          <div className="icon">
            <HiCamera style={{ color: '#655496' }} />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature1-title')}</b>
            </h4>
            <p>{t('Features.Feature1-description')}</p>
          </div>
        </div>

        <div className={styles['feature']}>
          <div className="icon">
            <HiChartBar />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature2-title')}</b>
            </h4>
            <p>{t('Features.Feature2-description')}</p>
          </div>
        </div>
        <div className={styles['feature']}>
          <div className="icon">
            <HiGlobeAlt />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature3-title')}</b>
            </h4>
            <p>{t('Features.Feature3-description')}</p>
          </div>
        </div>
        <div className={styles['feature']}>
          <div className="icon">
            <HiPencil style={{ color: '#000000' }} />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature4-title')}</b>
            </h4>
            <p>{t('Features.Feature4-description')}</p>
          </div>
        </div>
        <div className={styles['feature']}>
          <div className="icon">
            <HiFilter />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature5-title')}</b>
            </h4>
            <p>{t('Features.Feature5-description')}</p>
          </div>
        </div>
        <div className={styles['feature']}>
          <div className="icon">
            <HiStar style={{ color: '#FFD43B' }} />
          </div>
          <div className={styles['feature-container']}>
            <h4>
              <b>{t('Features.Feature6-title')}</b>
            </h4>
            <p>{t('Features.Feature6-description')}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default React.memo(Features);
