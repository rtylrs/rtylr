import {
  FaComments,
  FaPhone,
  FaUtensils,
  FaHome,
  FaInfoCircle,
} from 'react-icons/fa';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import useAsync from '../../../hooks/useAsync';
import { menuApi } from '../../../api';
import LandingSection from './Sections/LandingSection';
import MenuSection from './Sections/MenuSection';
import AboutSection from './Sections/AboutSection';
import ContactSection from './Sections/ContactSection';
import FeedbackSection from './Sections/FeedbackSection';
import { FullScreenLoader } from '@voxire/shared';
import Alert from '@mui/material/Alert';

function Menu() {
  const [menuArray, setmenuArray] = useState([]);
  const [restaurant, setRestaurant] = useState<any>();
  const [theme, setTheme] = useState<any>();
  const [status, setStatus] = useState('idle');
  const [retryCount, setRetryCount] = useState<number>(0);
  const [activeSection, setActiveSection] = useState<string>('menu');
  const { id } = useParams();

  const getMenu = useAsync({
    fn: menuApi.getMenu,
    onSuccess: () => {
      setRetryCount(0);
      const { details } = getMenu.result;
      setmenuArray(details.menu.items);
      setRestaurant(details.business);
      setTheme(details.theme);
      setStatus('succeeded');

      document.title = details?.business?.name || 'Menu';

      const url = new URL(window.location.href);

      if (!url.searchParams.has('title')) {
        url.searchParams.append('title', details?.business?.name || '');
      }

      window.history.replaceState({}, '', url.toString());
    },
    onError: () => {
      setStatus('failed');

      if (retryCount <= 3) {
        setTimeout(() => {
          getMenu.main({ id });
          setRetryCount(retryCount + 1);
        }, 10000);
      } else {
        console.error('Failed to load menu');
      }
    },
    callOnMount: false,
  });

  useEffect(() => {
    setStatus('loading');
    getMenu.main({ id });
  }, []);

  if (status === 'loading') {
    return <FullScreenLoader />;
  }

  if (status === 'failed') {
    return (
      <div
        className="menu-container"
        style={{ justifyContent: 'center', alignItems: 'center' }}
      >
        <div className="menu">
          <Alert severity="error" style={{ width: '90%', margin: 'auto' }}>
            Failed to load menu
          </Alert>
        </div>
      </div>
    );
  }

  const renderActiveSection = () => {
    switch (activeSection) {
      case 'landing':
        return <LandingSection />;
      case 'menu':
        return <MenuSection menuArray={menuArray} />;
      case 'about':
        return <AboutSection />;
      case 'contact':
        return <ContactSection />;
      case 'feedback':
        return <FeedbackSection />;
      default:
        return null;
    }
  };

  return (
    <div
      className="menu-container"
      style={{ backgroundColor: theme?.menuBackgroundColor }}
    >
      <div className="menu">
        <h1 style={{ color: theme?.primaryColor }}>
          {restaurant?.name}'s Menu
        </h1>
        {renderActiveSection()}
      </div>
      <div
        className="bottom-navigation"
        style={{
          backgroundColor: theme?.bottomNavigationBackgroundColor,
        }}
      >
        <div onClick={() => setActiveSection('about')}>
          <FaInfoCircle />
          <p>About</p>
        </div>
        {restaurant?.addons?.landing && (
          <div onClick={() => setActiveSection('menu')}>
            <FaUtensils />
            <p>Menu</p>
          </div>
        )}
        <div
          onClick={() => {
            if (restaurant?.addons?.landing) {
              setActiveSection('landing');
            } else {
              setActiveSection('menu');
            }
          }}
        >
          {restaurant?.addons.landing ? <FaHome /> : <FaUtensils />}
          <p>{restaurant?.addons?.landing ? 'Home' : 'Menu'}</p>
        </div>
        <div onClick={() => setActiveSection('contact')}>
          <FaPhone />
          <p>Contact</p>
        </div>
        <div onClick={() => setActiveSection('feedback')}>
          <FaComments />
          <p>Feedback</p>
        </div>
      </div>
    </div>
  );
}

export default Menu;
