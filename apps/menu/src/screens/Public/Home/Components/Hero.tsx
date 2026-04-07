import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@voxire/components';
import { FaArrowRight, FaStar } from 'react-icons/fa';
import useColors from '../../../../hooks/useColors';
import { useSelector } from 'react-redux';
import { selectIsAuthenticated } from '../../../../redux/reducers/userSlice';
import { QRCode } from 'react-qrcode-logo';
import { useWindowDimensions } from '@voxire/utils';

const Hero: React.FC<{}> = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { width } = useWindowDimensions();
  const { primary, secondary } = useColors();
  const isAuthenticated = useSelector(selectIsAuthenticated);

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
    <div className="Hero">
      <div className="hero-text">
        <h1 className="rtylr-title">{t('Hero.welcome')}</h1>
        <h2>{t('Hero.discover')}</h2>
        <div className="hero-buttons">
          <Button
            onClick={() => {
              if (isAuthenticated) {
                navigate('/app/business');
              } else {
                navigate('/auth/register');
              }
            }}
          >
            {isAuthenticated ? t('Hero.My Account') : t('Hero.Create Account')}
          </Button>
          <Button
            onClick={() => navigate('/uses')}
            variant="outline"
            suffix={<FaArrowRight />}
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
          <FaStar color={secondary} /> Rated 4.7/5 - from
          over 500+ reviews
        </h5>
      </div>
      <div className="hero-image">{memoizedQRCode}</div>
    </div>
  );
};

export default Hero;
