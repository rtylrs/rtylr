import { useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faStar,
  faInfoCircle,
  faPhone,
  faGlobe,
} from '@fortawesome/free-solid-svg-icons';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../../redux/store';
import useAsync from '../../hooks/useAsync';
import { userApi } from '../../api';
import { Button } from '@voxire/shared';
import { useNavigate } from 'react-router-dom';
import { getUserInfo } from '../../redux/reducers/userSlice';
import notify from '../../lib/notify/notify';
import { CircularProgress } from '@mui/material';

const ADDONS = [
  {
    key: 'feedback',
    name: 'Feedback',
    description: 'Collect feedback from your customers.',
    icon: faStar,
  },
  {
    key: 'about',
    name: 'About',
    description: 'Information about your business.',
    icon: faInfoCircle,
  },
  {
    key: 'contact',
    name: 'Contact',
    description: 'Enable customers to contact you easily.',
    icon: faPhone,
  },
  {
    key: 'landing',
    name: 'Landing',
    description: 'Create a landing page for your business.',
    icon: faGlobe,
  },
];

const Addons = () => {
  const planType = useSelector((state: RootState) => state.user.user.planType);
  const userAddons = useSelector((state: RootState) => state.user.addons);
  const userId = useSelector((state: RootState) => state.user.userId);
  const [searchTerm, setSearchTerm] = useState('');
  const [showInstalled, setShowInstalled] = useState(false);
  const navigate = useNavigate();
  const dispatch = useDispatch<AppDispatch>();
  const [keyLoading, setKeyLoading] = useState<number | null>();

  const editAddonsFunc = useAsync({
    fn: userApi.editAddons,
    onSuccess: () => {
      dispatch(getUserInfo()).then(() => {
        notify.success('Updated Successfully');
        setKeyLoading(null);
      });
    },
    onError: () => {
      alert(editAddonsFunc.error);
    },
    callOnMount: false,
  });

  const handleToggleAddon = (addonKey) => {
    const newAddons = {
      ...userAddons,
      [addonKey]: !userAddons[addonKey],
    };

    if (planType !== 'free') {
      editAddonsFunc.main({
        id: userId,
        addons: newAddons,
      });

      setKeyLoading(addonKey);
    }
  };

  const filteredAddons = ADDONS.filter(
    (addon) =>
      addon.name.toLowerCase().includes(searchTerm.toLowerCase()) &&
      (!showInstalled || userAddons[addon.key])
  );

  return (
    <div className="addons">
      {planType === 'free' && (
        <div className="overlay">
          <div className="overlay-message">
            Addons are not available in the free plan
          </div>
          <Button onClick={() => navigate('/app/business/billing')}>
            Upgrade
          </Button>
        </div>
      )}
      <div className="addons-header">
        <h1>Add-Ons</h1>
        <div>
          <Button
            onClick={() => setShowInstalled(!showInstalled)}
            styles={{
              height: '35px',
            }}
          >
            {showInstalled ? 'Show All' : 'Installed'}
          </Button>
          <input
            type="text"
            placeholder="Search for an add-on"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>
      <div className="addons-content">
        <div className="addon-category">
          <h2>Power-Ups</h2>
          <div className="addon-grid">
            {filteredAddons.map((addon) => (
              <div key={addon.key} className="addon-card">
                <FontAwesomeIcon icon={addon.icon} className="addon-icon" />
                <h3>{addon.name}</h3>
                <p>{addon.description}</p>
                {keyLoading === (addon.key as string | null) ? (
                  <CircularProgress />
                ) : (
                  <button
                    className={userAddons?.[addon.key] ? 'installed' : ''}
                    onClick={() => handleToggleAddon(addon.key)}
                  >
                    {userAddons?.[addon.key] ? 'Remove' : 'Install (Free)'}
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>
        <div className="more-addons">
          <h3>More Addons Coming Soon!</h3>
        </div>
      </div>
    </div>
  );
};

export default Addons;
